package repo

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	apperrors "goadmin/core/errors"
	coretenant "goadmin/core/tenant"
	authmodel "goadmin/modules/auth/domain/model"

	"gorm.io/gorm"
)

type CredentialStore interface {
	Authenticate(ctx context.Context, username, password string) (authmodel.Identity, error)
}

type UserTableStore struct {
	db       *gorm.DB
	fallback CredentialStore
}

func NewUserTableStore(db *gorm.DB, fallback CredentialStore) *UserTableStore {
	return &UserTableStore{db: db, fallback: fallback}
}

type userTableRecord struct {
	ID           string    `gorm:"column:id;primaryKey;size:64"`
	TenantID     string    `gorm:"column:tenant_id;size:64"`
	Username     string    `gorm:"column:username;size:128"`
	DisplayName  string    `gorm:"column:display_name;size:128"`
	Language     string    `gorm:"column:language;size:32"`
	PasswordHash string    `gorm:"column:password_hash;type:text"`
	RoleCodesRaw string    `gorm:"column:role_codes;type:text;not null"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (userTableRecord) TableName() string {
	return "user"
}

type roleTableRecord struct {
	ID         string `gorm:"column:id;primaryKey;size:64"`
	TenantID   string `gorm:"column:tenant_id;size:64"`
	Code       string `gorm:"column:code;size:128"`
	MenuIDsRaw string `gorm:"column:menu_ids;type:text;not null"`
}

func (roleTableRecord) TableName() string {
	return "role"
}

type menuTableRecord struct {
	ID         string `gorm:"column:id;primaryKey;size:64"`
	Permission string `gorm:"column:permission;size:255"`
}

func (menuTableRecord) TableName() string {
	return "menu"
}

func (s *UserTableStore) Authenticate(ctx context.Context, username, password string) (authmodel.Identity, error) {
	if s == nil || s.db == nil {
		if s != nil && s.fallback != nil {
			return s.fallback.Authenticate(ctx, username, password)
		}
		return authmodel.Identity{}, apperrors.New(apperrors.CodeUnauthorized, "authentication store is not configured")
	}

	key := normalize(username)
	if key == "" {
		return authmodel.Identity{}, apperrors.New(apperrors.CodeUnauthorized, "invalid credentials")
	}

	user, found, err := s.findUserByUsername(ctx, key)
	if err != nil {
		return authmodel.Identity{}, err
	}
	if !found {
		if s.fallback != nil {
			return s.fallback.Authenticate(ctx, username, password)
		}
		return authmodel.Identity{}, apperrors.New(apperrors.CodeUnauthorized, "invalid credentials")
	}

	if err := verifyStoredPassword(user.PasswordHash, password); err != nil {
		return authmodel.Identity{}, err
	}

	roles := normalizeStrings(parseRoleCodes(user.RoleCodesRaw))
	if len(roles) == 0 {
		roles = []string{"user"}
	}

	permissions := s.resolvePermissions(ctx, user.TenantID, roles)
	identity := authmodel.Identity{
		UserID:      fallback(user.Username, key),
		TenantID:    strings.TrimSpace(user.TenantID),
		Username:    fallback(user.Username, key),
		DisplayName: fallback(user.DisplayName, fallback(user.Username, key)),
		Language:    strings.TrimSpace(user.Language),
		Roles:       append([]string(nil), roles...),
		Permissions: permissions,
	}
	if !coretenant.Enabled() {
		identity.TenantID = ""
	}
	if identity.UserID == "" {
		identity.UserID = identity.Username
	}
	if identity.DisplayName == "" {
		identity.DisplayName = identity.Username
	}
	if len(identity.Roles) == 0 {
		identity.Roles = []string{"user"}
	}
	if strings.TrimSpace(identity.Language) == "" {
		identity.Language = "zh-CN"
	}
	return identity, nil
}

func (s *UserTableStore) findUserByUsername(ctx context.Context, username string) (*userTableRecord, bool, error) {
	var row userTableRecord
	query := s.db.WithContext(ctx).Model(&userTableRecord{})
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		query = query.Where("tenant_id = ?", strings.TrimSpace(tenantID))
	}
	if err := query.Order("updated_at DESC, created_at DESC, id ASC").First(&row, "LOWER(username) = LOWER(?)", strings.TrimSpace(username)).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, false, nil
		default:
			return nil, false, apperrors.Wrap(err, apperrors.CodeInternal, "load user failed")
		}
	}
	return &row, true, nil
}

func (s *UserTableStore) resolvePermissions(ctx context.Context, tenantID string, roles []string) []string {
	if containsRole(roles, "admin") {
		return []string{"*"}
	}
	menuIDs := s.resolveMenuIDs(ctx, tenantID, roles)
	if len(menuIDs) == 0 {
		return nil
	}

	var menus []menuTableRecord
	query := s.db.WithContext(ctx).Model(&menuTableRecord{})
	if err := query.Where("id IN ?", menuIDs).Find(&menus).Error; err != nil {
		return nil
	}

	permissions := make([]string, 0, len(menus))
	seen := make(map[string]struct{}, len(menus))
	for _, menu := range menus {
		permission := strings.TrimSpace(menu.Permission)
		if permission == "" {
			continue
		}
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		permissions = append(permissions, permission)
	}
	sort.Strings(permissions)
	return permissions
}

func (s *UserTableStore) resolveMenuIDs(ctx context.Context, tenantID string, roles []string) []string {
	if len(roles) == 0 {
		return nil
	}

	var roleRows []roleTableRecord
	query := s.db.WithContext(ctx).Model(&roleTableRecord{})
	if trimmedTenant := strings.TrimSpace(tenantID); trimmedTenant != "" {
		query = query.Where("tenant_id = ?", trimmedTenant)
	}
	if err := query.Where("code IN ?", roles).Find(&roleRows).Error; err != nil {
		return nil
	}

	menuIDs := make([]string, 0, len(roleRows))
	seen := make(map[string]struct{}, len(roleRows))
	for _, role := range roleRows {
		var ids []string
		if strings.TrimSpace(role.MenuIDsRaw) != "" {
			if err := json.Unmarshal([]byte(role.MenuIDsRaw), &ids); err != nil {
				continue
			}
		}
		for _, id := range ids {
			menuID := strings.TrimSpace(id)
			if menuID == "" {
				continue
			}
			if _, ok := seen[menuID]; ok {
				continue
			}
			seen[menuID] = struct{}{}
			menuIDs = append(menuIDs, menuID)
		}
	}
	return menuIDs
}

func verifyStoredPassword(passwordHash, password string) error {
	if strings.TrimSpace(passwordHash) != "" {
		hash := strings.TrimSpace(passwordHash)
		if strings.HasPrefix(hash, "sha256:") {
			want := strings.TrimPrefix(hash, "sha256:")
			sum := sha256.Sum256([]byte(password))
			if subtle.ConstantTimeCompare([]byte(strings.ToLower(want)), []byte(hex.EncodeToString(sum[:]))) != 1 {
				return apperrors.New(apperrors.CodeUnauthorized, "invalid credentials")
			}
			return nil
		}
		if subtle.ConstantTimeCompare([]byte(hash), []byte(password)) != 1 {
			return apperrors.New(apperrors.CodeUnauthorized, "invalid credentials")
		}
		return nil
	}
	return apperrors.New(apperrors.CodeUnauthorized, "invalid credentials")
}

func containsRole(roles []string, role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		return false
	}
	for _, item := range roles {
		if strings.ToLower(strings.TrimSpace(item)) == role {
			return true
		}
	}
	return false
}

func parseRoleCodes(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var codes []string
	if err := json.Unmarshal([]byte(raw), &codes); err != nil {
		return nil
	}
	return codes
}

package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	coretenant "goadmin/core/tenant"
	"goadmin/modules/user/domain/model"
	userrepo "goadmin/modules/user/domain/repository"

	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

type userRecord struct {
	ID           string    `gorm:"column:id;primaryKey;size:64"`
	TenantID     string    `gorm:"column:tenant_id;size:64;index:idx_users_tenant_username,priority:1"`
	Username     string    `gorm:"column:username;size:128;not null;index:idx_users_tenant_username,priority:2"`
	DisplayName  string    `gorm:"column:display_name;size:128"`
	Mobile       string    `gorm:"column:mobile;size:32"`
	Email        string    `gorm:"column:email;size:128"`
	Status       string    `gorm:"column:status;size:32;not null;index"`
	RoleCodesRaw string    `gorm:"column:role_codes;type:text;not null"`
	PasswordHash string    `gorm:"column:password_hash;type:text"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (userRecord) TableName() string { return "users" }

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("user gorm repository requires db")
	}
	return &GormRepository{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("user migrate requires db")
	}
	return db.AutoMigrate(&userRecord{})
}

func (r *GormRepository) List(ctx context.Context, filter userrepo.ListFilter) ([]model.User, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("user gorm repository is not configured")
	}
	if strings.TrimSpace(filter.TenantID) == "" {
		if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok {
			filter.TenantID = tenantID
		}
	}
	base := r.applyFilters(r.db.WithContext(ctx).Model(&userRecord{}), filter)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(filter.Page, filter.PageSize)
	var rows []userRecord
	if err := base.Order("updated_at DESC, created_at DESC, id ASC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items, err := mapUserRecords(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GormRepository) Get(ctx context.Context, id string) (*model.User, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("user gorm repository is not configured")
	}
	var row userRecord
	query := r.db.WithContext(ctx).Model(&userRecord{})
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&row, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return nil, mapUserRepoError(err)
	}
	item, err := row.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("user gorm repository is not configured")
	}
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}
	record, err := toUserRecord(*user)
	if err != nil {
		return nil, err
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		if strings.TrimSpace(record.TenantID) == "" {
			record.TenantID = tenantID
		} else if strings.TrimSpace(record.TenantID) != tenantID {
			return nil, coretenant.ErrTenantMismatch
		}
	}
	if strings.TrimSpace(record.ID) == "" {
		record.ID = nextRecordID("user")
	}
	if exists, err := r.exists(ctx, record.TenantID, record.Username, ""); err != nil {
		return nil, err
	} else if exists {
		return nil, userrepo.ErrConflict
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, mapUserRepoError(err)
	}
	item, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Update(ctx context.Context, user *model.User) (*model.User, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("user gorm repository is not configured")
	}
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}
	record, err := toUserRecord(*user)
	if err != nil {
		return nil, err
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		if strings.TrimSpace(record.TenantID) == "" {
			record.TenantID = tenantID
		} else if strings.TrimSpace(record.TenantID) != tenantID {
			return nil, coretenant.ErrTenantMismatch
		}
	}
	var existing userRecord
	query := r.db.WithContext(ctx).Model(&userRecord{})
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing, "id = ?", strings.TrimSpace(record.ID)).Error; err != nil {
		return nil, mapUserRepoError(err)
	}
	if exists, err := r.exists(ctx, record.TenantID, record.Username, record.ID); err != nil {
		return nil, err
	} else if exists {
		return nil, userrepo.ErrConflict
	}
	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, mapUserRepoError(err)
	}
	item, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Delete(ctx context.Context, id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("user gorm repository is not configured")
	}
	query := r.db.WithContext(ctx).Model(&userRecord{})
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	result := query.Delete(&userRecord{}, "id = ?", strings.TrimSpace(id))
	if result.Error != nil {
		return mapUserRepoError(result.Error)
	}
	if result.RowsAffected == 0 {
		return userrepo.ErrNotFound
	}
	return nil
}

func (r *GormRepository) applyFilters(db *gorm.DB, filter userrepo.ListFilter) *gorm.DB {
	if tenant := strings.TrimSpace(filter.TenantID); tenant != "" {
		db = db.Where("tenant_id = ?", tenant)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		db = db.Where("status = ?", strings.ToLower(status))
	}
	if kw := strings.TrimSpace(strings.ToLower(filter.Keyword)); kw != "" {
		like := "%" + kw + "%"
		db = db.Where("LOWER(username) LIKE ? OR LOWER(display_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(mobile) LIKE ? OR LOWER(tenant_id) LIKE ?", like, like, like, like, like)
	}
	return db
}

func (r *GormRepository) exists(ctx context.Context, tenantID, username, excludeID string) (bool, error) {
	query := r.db.WithContext(ctx).Model(&userRecord{}).Where("tenant_id = ? AND username = ?", strings.TrimSpace(tenantID), strings.TrimSpace(username))
	if strings.TrimSpace(excludeID) != "" {
		query = query.Where("id <> ?", strings.TrimSpace(excludeID))
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func toUserRecord(user model.User) (userRecord, error) {
	codes, err := json.Marshal(append([]string(nil), user.RoleCodes...))
	if err != nil {
		return userRecord{}, fmt.Errorf("marshal user role codes: %w", err)
	}
	status := strings.TrimSpace(string(user.Status))
	if status == "" {
		status = string(model.StatusActive)
	}
	return userRecord{
		ID:           strings.TrimSpace(user.ID),
		TenantID:     strings.TrimSpace(user.TenantID),
		Username:     strings.TrimSpace(user.Username),
		DisplayName:  strings.TrimSpace(user.DisplayName),
		Mobile:       strings.TrimSpace(user.Mobile),
		Email:        strings.TrimSpace(user.Email),
		Status:       status,
		RoleCodesRaw: string(codes),
		PasswordHash: strings.TrimSpace(user.PasswordHash),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}

func (r userRecord) toModel() (model.User, error) {
	var codes []string
	if strings.TrimSpace(r.RoleCodesRaw) != "" {
		if err := json.Unmarshal([]byte(r.RoleCodesRaw), &codes); err != nil {
			return model.User{}, fmt.Errorf("unmarshal user role codes: %w", err)
		}
	}
	return model.User{ID: strings.TrimSpace(r.ID), TenantID: strings.TrimSpace(r.TenantID), Username: strings.TrimSpace(r.Username), DisplayName: strings.TrimSpace(r.DisplayName), Mobile: strings.TrimSpace(r.Mobile), Email: strings.TrimSpace(r.Email), Status: model.Status(strings.TrimSpace(r.Status)), RoleCodes: append([]string(nil), codes...), PasswordHash: strings.TrimSpace(r.PasswordHash), CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}, nil
}

func mapUserRecords(rows []userRecord) ([]model.User, error) {
	items := make([]model.User, 0, len(rows))
	for _, row := range rows {
		item, err := row.toModel()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func mapUserRepoError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return userrepo.ErrNotFound
	case strings.Contains(strings.ToLower(err.Error()), "unique constraint failed"):
		return userrepo.ErrConflict
	default:
		return err
	}
}

func nextRecordID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}

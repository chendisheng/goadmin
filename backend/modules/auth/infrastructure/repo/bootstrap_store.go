package repo

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
	"sync"

	"goadmin/core/config"
	apperrors "goadmin/core/errors"
	coretenant "goadmin/core/tenant"
	"goadmin/modules/auth/domain/model"
)

type BootstrapStore struct {
	mu    sync.RWMutex
	users map[string]config.BootstrapUser
}

func NewBootstrapStore(users []config.BootstrapUser) *BootstrapStore {
	store := &BootstrapStore{users: make(map[string]config.BootstrapUser, len(users))}
	for _, user := range users {
		username := normalize(user.Username)
		if username == "" {
			continue
		}
		store.users[username] = user
	}
	return store
}

func (s *BootstrapStore) Authenticate(_ context.Context, username, password string) (model.Identity, error) {
	if s == nil {
		return model.Identity{}, apperrors.New(apperrors.CodeUnauthorized, "authentication store is not configured")
	}
	key := normalize(username)
	if key == "" {
		return model.Identity{}, apperrors.New(apperrors.CodeUnauthorized, "invalid credentials")
	}

	s.mu.RLock()
	user, ok := s.users[key]
	s.mu.RUnlock()
	if !ok {
		return model.Identity{}, apperrors.New(apperrors.CodeUnauthorized, "invalid credentials")
	}

	if err := verifyPassword(user, password); err != nil {
		return model.Identity{}, err
	}

	identity := model.Identity{
		UserID:      fallback(user.Username, key),
		TenantID:    strings.TrimSpace(user.TenantID),
		Username:    fallback(user.Username, key),
		DisplayName: fallback(user.DisplayName, fallback(user.Username, key)),
		Roles:       append([]string(nil), user.Roles...),
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
	return identity, nil
}

func verifyPassword(user config.BootstrapUser, password string) error {
	if strings.TrimSpace(user.PasswordHash) != "" {
		hash := strings.TrimSpace(user.PasswordHash)
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
	if subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) != 1 {
		return apperrors.New(apperrors.CodeUnauthorized, "invalid credentials")
	}
	return nil
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func fallback(value, def string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return def
	}
	return value
}

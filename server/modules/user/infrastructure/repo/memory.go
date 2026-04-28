package repo

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	coretenant "goadmin/core/tenant"
	"goadmin/modules/user/domain/model"
	userrepo "goadmin/modules/user/domain/repository"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*model.User
	order []string
	seq   int64
}

func NewMemoryRepository(seed []model.User) *MemoryRepository {
	r := &MemoryRepository{
		items: make(map[string]*model.User),
		order: make([]string, 0),
	}
	if len(seed) == 0 {
		seed = defaultUsers()
	}
	for i := range seed {
		_ = r.mustInsert(seed[i])
	}
	return r
}

func (r *MemoryRepository) List(ctx context.Context, filter userrepo.ListFilter) ([]model.User, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if strings.TrimSpace(filter.TenantID) == "" {
		if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok {
			filter.TenantID = tenantID
		}
	}

	items := make([]model.User, 0, len(r.order))
	for _, id := range r.order {
		item := r.items[id]
		if item == nil {
			continue
		}
		if filter.TenantID != "" && item.TenantID != filter.TenantID {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(string(item.Status), filter.Status) {
			continue
		}
		if kw := strings.TrimSpace(strings.ToLower(filter.Keyword)); kw != "" {
			if !containsUserKeyword(item, kw) {
				continue
			}
		}
		items = append(items, item.Clone())
	}
	total := int64(len(items))
	page, size := normalizePage(filter.Page, filter.PageSize)
	start := (page - 1) * size
	if start >= len(items) {
		return []model.User{}, total, nil
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return append([]model.User(nil), items[start:end]...), total, nil
}

func (r *MemoryRepository) Get(ctx context.Context, id string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok || item == nil {
		return nil, userrepo.ErrNotFound
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" && !strings.EqualFold(strings.TrimSpace(item.TenantID), tenantID) {
		return nil, userrepo.ErrNotFound
	}
	clone := item.Clone()
	return &clone, nil
}

func (r *MemoryRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		if strings.TrimSpace(user.TenantID) == "" {
			user.TenantID = tenantID
		} else if !strings.EqualFold(strings.TrimSpace(user.TenantID), tenantID) {
			return nil, coretenant.ErrTenantMismatch
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if conflict := r.findByUsernameLocked(user.TenantID, user.Username, ""); conflict != nil {
		return nil, userrepo.ErrConflict
	}
	copy := user.Clone()
	if strings.TrimSpace(copy.ID) == "" {
		copy.ID = r.nextIDLocked("user")
	}
	now := time.Now().UTC()
	copy.CreatedAt = now
	copy.UpdatedAt = now
	r.items[copy.ID] = &copy
	r.order = append(r.order, copy.ID)
	clone := copy.Clone()
	return &clone, nil
}

func (r *MemoryRepository) Update(ctx context.Context, user *model.User) (*model.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		if strings.TrimSpace(user.TenantID) == "" {
			user.TenantID = tenantID
		} else if !strings.EqualFold(strings.TrimSpace(user.TenantID), tenantID) {
			return nil, coretenant.ErrTenantMismatch
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.items[user.ID]
	if !ok || existing == nil {
		return nil, userrepo.ErrNotFound
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" && !strings.EqualFold(strings.TrimSpace(existing.TenantID), tenantID) {
		return nil, userrepo.ErrNotFound
	}
	if conflict := r.findByUsernameLocked(user.TenantID, user.Username, user.ID); conflict != nil {
		return nil, userrepo.ErrConflict
	}
	updated := existing.Clone()
	if strings.TrimSpace(user.TenantID) != "" {
		updated.TenantID = user.TenantID
	}
	if strings.TrimSpace(user.Username) != "" {
		updated.Username = user.Username
	}
	if strings.TrimSpace(user.DisplayName) != "" {
		updated.DisplayName = user.DisplayName
	}
	if strings.TrimSpace(user.Language) != "" {
		updated.Language = user.Language
	}
	if strings.TrimSpace(user.Mobile) != "" {
		updated.Mobile = user.Mobile
	}
	if strings.TrimSpace(user.Email) != "" {
		updated.Email = user.Email
	}
	if strings.TrimSpace(string(user.Status)) != "" {
		updated.Status = user.Status
	}
	if user.RoleCodes != nil {
		updated.RoleCodes = append([]string(nil), user.RoleCodes...)
	}
	if strings.TrimSpace(user.PasswordHash) != "" {
		updated.PasswordHash = user.PasswordHash
	}
	updated.UpdatedAt = time.Now().UTC()
	r.items[user.ID] = &updated
	clone := updated.Clone()
	return &clone, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.items[id]
	if !ok || item == nil {
		return userrepo.ErrNotFound
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" && !strings.EqualFold(strings.TrimSpace(item.TenantID), tenantID) {
		return userrepo.ErrNotFound
	}
	delete(r.items, id)
	filtered := r.order[:0]
	for _, itemID := range r.order {
		if itemID == id {
			continue
		}
		filtered = append(filtered, itemID)
	}
	r.order = append([]string(nil), filtered...)
	return nil
}

func (r *MemoryRepository) mustInsert(user model.User) error {
	_, err := r.Create(context.Background(), &user)
	return err
}

func (r *MemoryRepository) nextIDLocked(prefix string) string {
	seq := atomic.AddInt64(&r.seq, 1)
	return fmt.Sprintf("%s-%d", prefix, seq)
}

func (r *MemoryRepository) findByUsernameLocked(tenantID, username, excludeID string) *model.User {
	for _, item := range r.items {
		if item == nil {
			continue
		}
		if excludeID != "" && item.ID == excludeID {
			continue
		}
		if strings.TrimSpace(tenantID) != "" && !strings.EqualFold(strings.TrimSpace(item.TenantID), strings.TrimSpace(tenantID)) {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(item.Username), strings.TrimSpace(username)) {
			clone := item.Clone()
			return &clone
		}
	}
	return nil
}

func normalizePage(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func containsUserKeyword(user *model.User, kw string) bool {
	return strings.Contains(strings.ToLower(user.Username), kw) ||
		strings.Contains(strings.ToLower(user.DisplayName), kw) ||
		strings.Contains(strings.ToLower(user.Email), kw) ||
		strings.Contains(strings.ToLower(user.Mobile), kw) ||
		strings.Contains(strings.ToLower(user.TenantID), kw)
}

func defaultUsers() []model.User {
	return []model.User{
		{ID: "user-admin", TenantID: "system", Username: "admin", DisplayName: "System Admin", Language: "zh-CN", Email: "admin@goadmin.local", Mobile: "", Status: model.StatusActive, RoleCodes: []string{"admin"}},
		{ID: "user-demo", TenantID: "system", Username: "demo", DisplayName: "Demo User", Language: "zh-CN", Email: "demo@goadmin.local", Mobile: "", Status: model.StatusActive, RoleCodes: []string{"user"}},
	}
}

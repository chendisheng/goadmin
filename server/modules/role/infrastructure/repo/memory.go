package repo

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	coretenant "goadmin/core/tenant"
	"goadmin/modules/role/domain/model"
	rolerepo "goadmin/modules/role/domain/repository"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*model.Role
	order []string
	seq   int64
}

func NewMemoryRepository(seed []model.Role) *MemoryRepository {
	r := &MemoryRepository{items: make(map[string]*model.Role), order: make([]string, 0)}
	if len(seed) == 0 {
		seed = defaultRoles()
	}
	for i := range seed {
		_ = r.mustInsert(seed[i])
	}
	return r
}

func (r *MemoryRepository) List(ctx context.Context, filter rolerepo.ListFilter) ([]model.Role, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if strings.TrimSpace(filter.TenantID) == "" {
		if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok {
			filter.TenantID = tenantID
		}
	}
	items := make([]model.Role, 0, len(r.order))
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
			if !containsRoleKeyword(item, kw) {
				continue
			}
		}
		items = append(items, item.Clone())
	}
	total := int64(len(items))
	page, size := normalizePage(filter.Page, filter.PageSize)
	start := (page - 1) * size
	if start >= len(items) {
		return []model.Role{}, total, nil
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return append([]model.Role(nil), items[start:end]...), total, nil
}

func (r *MemoryRepository) Get(ctx context.Context, id string) (*model.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok || item == nil {
		return nil, rolerepo.ErrNotFound
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" && !strings.EqualFold(strings.TrimSpace(item.TenantID), tenantID) {
		return nil, rolerepo.ErrNotFound
	}
	clone := item.Clone()
	return &clone, nil
}

func (r *MemoryRepository) Create(ctx context.Context, role *model.Role) (*model.Role, error) {
	if role == nil {
		return nil, fmt.Errorf("role is nil")
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		if strings.TrimSpace(role.TenantID) == "" {
			role.TenantID = tenantID
		} else if !strings.EqualFold(strings.TrimSpace(role.TenantID), tenantID) {
			return nil, coretenant.ErrTenantMismatch
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if conflict := r.findByCodeLocked(role.TenantID, role.Code, ""); conflict != nil {
		return nil, rolerepo.ErrConflict
	}
	copy := role.Clone()
	if strings.TrimSpace(copy.ID) == "" {
		copy.ID = r.nextIDLocked("role")
	}
	now := time.Now().UTC()
	copy.CreatedAt = now
	copy.UpdatedAt = now
	r.items[copy.ID] = &copy
	r.order = append(r.order, copy.ID)
	clone := copy.Clone()
	return &clone, nil
}

func (r *MemoryRepository) Update(ctx context.Context, role *model.Role) (*model.Role, error) {
	if role == nil {
		return nil, fmt.Errorf("role is nil")
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		if strings.TrimSpace(role.TenantID) == "" {
			role.TenantID = tenantID
		} else if !strings.EqualFold(strings.TrimSpace(role.TenantID), tenantID) {
			return nil, coretenant.ErrTenantMismatch
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.items[role.ID]
	if !ok || existing == nil {
		return nil, rolerepo.ErrNotFound
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" && !strings.EqualFold(strings.TrimSpace(existing.TenantID), tenantID) {
		return nil, rolerepo.ErrNotFound
	}
	if conflict := r.findByCodeLocked(role.TenantID, role.Code, role.ID); conflict != nil {
		return nil, rolerepo.ErrConflict
	}
	updated := existing.Clone()
	if strings.TrimSpace(role.TenantID) != "" {
		updated.TenantID = role.TenantID
	}
	if strings.TrimSpace(role.Name) != "" {
		updated.Name = role.Name
	}
	if strings.TrimSpace(role.Code) != "" {
		updated.Code = role.Code
	}
	if strings.TrimSpace(string(role.Status)) != "" {
		updated.Status = role.Status
	}
	if strings.TrimSpace(role.Remark) != "" {
		updated.Remark = role.Remark
	}
	if role.MenuIDs != nil {
		updated.MenuIDs = append([]string(nil), role.MenuIDs...)
	}
	updated.UpdatedAt = time.Now().UTC()
	r.items[role.ID] = &updated
	clone := updated.Clone()
	return &clone, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.items[id]
	if !ok || item == nil {
		return rolerepo.ErrNotFound
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" && !strings.EqualFold(strings.TrimSpace(item.TenantID), tenantID) {
		return rolerepo.ErrNotFound
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

func (r *MemoryRepository) mustInsert(role model.Role) error {
	_, err := r.Create(context.Background(), &role)
	return err
}

func (r *MemoryRepository) nextIDLocked(prefix string) string {
	seq := atomic.AddInt64(&r.seq, 1)
	return fmt.Sprintf("%s-%d", prefix, seq)
}

func (r *MemoryRepository) findByCodeLocked(tenantID, code, excludeID string) *model.Role {
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
		if strings.EqualFold(strings.TrimSpace(item.Code), strings.TrimSpace(code)) {
			clone := item.Clone()
			return &clone
		}
	}
	return nil
}

func defaultRoles() []model.Role {
	return []model.Role{
		{ID: "role-admin", TenantID: "system", Name: "Administrator", Code: "admin", Status: model.StatusActive, Remark: "Built-in admin", MenuIDs: []string{"menu-home", "menu-dashboard", "menu-profile", "menu-system", "menu-users", "menu-roles", "menu-menus", "menu-nested", "menu-nested-menu1", "menu-nested-menu2", "menu-nested-menu2-one", "menu-nested-menu2-two", "menu-nested-menu3", "menu-nested-menu3-one", "menu-nested-menu3-two"}},
		{ID: "role-user", TenantID: "system", Name: "User", Code: "user", Status: model.StatusActive, Remark: "Built-in user", MenuIDs: []string{"menu-home", "menu-dashboard", "menu-profile"}},
	}
}

func containsRoleKeyword(role *model.Role, kw string) bool {
	return strings.Contains(strings.ToLower(role.Name), kw) ||
		strings.Contains(strings.ToLower(role.Code), kw) ||
		strings.Contains(strings.ToLower(role.Remark), kw) ||
		strings.Contains(strings.ToLower(role.TenantID), kw)
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

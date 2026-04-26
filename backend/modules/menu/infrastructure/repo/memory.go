package repo

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"goadmin/modules/menu/domain/model"
	menurepo "goadmin/modules/menu/domain/repository"
)

type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]*model.Menu
	order []string
	seq   int64
}

func NewMemoryRepository(seed []model.Menu) *MemoryRepository {
	r := &MemoryRepository{items: make(map[string]*model.Menu), order: make([]string, 0)}
	if len(seed) == 0 {
		seed = defaultMenus()
	}
	for i := range seed {
		_ = r.mustInsert(seed[i])
	}
	return r
}

func (r *MemoryRepository) List(_ context.Context, filter menurepo.ListFilter) ([]model.Menu, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]model.Menu, 0, len(r.order))
	for _, id := range r.order {
		item := r.items[id]
		if item == nil {
			continue
		}
		if filter.ParentID != "" && item.ParentID != filter.ParentID {
			continue
		}
		if filter.Visible != nil && item.Visible != *filter.Visible {
			continue
		}
		if filter.Enabled != nil && item.Enabled != *filter.Enabled {
			continue
		}
		if kw := strings.TrimSpace(strings.ToLower(filter.Keyword)); kw != "" {
			if !containsMenuKeyword(item, kw) {
				continue
			}
		}
		items = append(items, item.Clone())
	}
	total := int64(len(items))
	page, size := normalizePage(filter.Page, filter.PageSize)
	start := (page - 1) * size
	if start >= len(items) {
		return []model.Menu{}, total, nil
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return append([]model.Menu(nil), items[start:end]...), total, nil
}

func (r *MemoryRepository) Get(_ context.Context, id string) (*model.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok || item == nil {
		return nil, menurepo.ErrNotFound
	}
	clone := item.Clone()
	return &clone, nil
}

func (r *MemoryRepository) Create(_ context.Context, menu *model.Menu) (*model.Menu, error) {
	if menu == nil {
		return nil, fmt.Errorf("menu is nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if conflict := r.findByPathLocked(menu.Path, ""); conflict != nil {
		return nil, menurepo.ErrConflict
	}
	copy := menu.Clone()
	if strings.TrimSpace(copy.ID) == "" {
		copy.ID = r.nextIDLocked("menu")
	}
	if copy.Type == "" {
		copy.Type = model.TypeMenu
	}
	now := time.Now().UTC()
	copy.CreatedAt = now
	copy.UpdatedAt = now
	r.items[copy.ID] = &copy
	r.order = append(r.order, copy.ID)
	clone := copy.Clone()
	return &clone, nil
}

func (r *MemoryRepository) Update(_ context.Context, menu *model.Menu) (*model.Menu, error) {
	if menu == nil {
		return nil, fmt.Errorf("menu is nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.items[menu.ID]
	if !ok || existing == nil {
		return nil, menurepo.ErrNotFound
	}
	if conflict := r.findByPathLocked(menu.Path, menu.ID); conflict != nil {
		return nil, menurepo.ErrConflict
	}
	updated := existing.Clone()
	if strings.TrimSpace(menu.ParentID) != "" {
		updated.ParentID = menu.ParentID
	}
	if strings.TrimSpace(menu.Name) != "" {
		updated.Name = menu.Name
	}
	if strings.TrimSpace(menu.Path) != "" {
		updated.Path = menu.Path
	}
	if strings.TrimSpace(menu.Component) != "" {
		updated.Component = menu.Component
	}
	if strings.TrimSpace(menu.Icon) != "" {
		updated.Icon = menu.Icon
	}
	if menu.Sort != 0 {
		updated.Sort = menu.Sort
	}
	if strings.TrimSpace(menu.Permission) != "" {
		updated.Permission = menu.Permission
	}
	if strings.TrimSpace(string(menu.Type)) != "" {
		updated.Type = menu.Type
	}
	updated.Visible = menu.Visible
	updated.Enabled = menu.Enabled
	if strings.TrimSpace(menu.Redirect) != "" {
		updated.Redirect = menu.Redirect
	}
	if strings.TrimSpace(menu.ExternalURL) != "" {
		updated.ExternalURL = menu.ExternalURL
	}
	updated.UpdatedAt = time.Now().UTC()
	r.items[menu.ID] = &updated
	clone := updated.Clone()
	return &clone, nil
}

func (r *MemoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return menurepo.ErrNotFound
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

func (r *MemoryRepository) Tree(_ context.Context) ([]model.Menu, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	flat := make([]model.Menu, 0, len(r.order))
	for _, id := range r.order {
		if item := r.items[id]; item != nil {
			flat = append(flat, item.Clone())
		}
	}
	return buildTree(flat), nil
}

func (r *MemoryRepository) mustInsert(menu model.Menu) error {
	_, err := r.Create(context.Background(), &menu)
	return err
}

func (r *MemoryRepository) nextIDLocked(prefix string) string {
	seq := atomic.AddInt64(&r.seq, 1)
	return fmt.Sprintf("%s-%d", prefix, seq)
}

func (r *MemoryRepository) findByPathLocked(pathValue, excludeID string) *model.Menu {
	for _, item := range r.items {
		if item == nil {
			continue
		}
		if excludeID != "" && item.ID == excludeID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(item.Path), strings.TrimSpace(pathValue)) {
			clone := item.Clone()
			return &clone
		}
	}
	return nil
}

func buildTree(items []model.Menu) []model.Menu {
	byParent := make(map[string][]model.Menu)
	for _, item := range items {
		byParent[item.ParentID] = append(byParent[item.ParentID], item)
	}
	for key := range byParent {
		sort.Slice(byParent[key], func(i, j int) bool {
			if byParent[key][i].Sort == byParent[key][j].Sort {
				return byParent[key][i].Name < byParent[key][j].Name
			}
			return byParent[key][i].Sort < byParent[key][j].Sort
		})
	}
	return buildChildren("", byParent)
}

func buildChildren(parentID string, byParent map[string][]model.Menu) []model.Menu {
	children := byParent[parentID]
	result := make([]model.Menu, 0, len(children))
	for _, child := range children {
		clone := child.Clone()
		clone.Children = buildChildren(child.ID, byParent)
		result = append(result, clone)
	}
	return result
}

func defaultMenus() []model.Menu {
	return []model.Menu{
		{ID: "menu-home", Name: "Home", Path: "/", Component: "Layout", Icon: "home", Sort: 1, Permission: "home:view", Type: model.TypeDirectory, Visible: true, Enabled: true, Redirect: "/dashboard"},
		{ID: "menu-dashboard", ParentID: "menu-home", Name: "Dashboard", Path: "/dashboard", Component: "view/dashboard/index", Icon: "dashboard", Sort: 1, Permission: "dashboard:view", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-profile", ParentID: "menu-home", Name: "Profile", Path: "/profile", Component: "view/profile/index", Icon: "user", Sort: 2, Permission: "profile:view", Type: model.TypeMenu, Visible: false, Enabled: true},
		{ID: "menu-system", Name: "System", Path: "/system", Component: "Layout", Icon: "setting", Sort: 2, Permission: "system:view", Type: model.TypeDirectory, Visible: true, Enabled: true, Redirect: "/system/users"},
		{ID: "menu-users", ParentID: "menu-system", Name: "Users", Path: "/system/users", Component: "view/system/user/index", Icon: "user", Sort: 1, Permission: "user:list", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-roles", ParentID: "menu-system", Name: "Roles", Path: "/system/roles", Component: "view/system/role/index", Icon: "role", Sort: 2, Permission: "role:list", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-menus", ParentID: "menu-system", Name: "Menus", Path: "/system/menus", Component: "view/system/menu/index", Icon: "menu", Sort: 3, Permission: "menu:list", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-dictionary", ParentID: "menu-system", Name: "Data Dictionary", Path: "/system/dictionary", Component: "Layout", Icon: "menu", Sort: 4, Permission: "dictionary:view", Type: model.TypeDirectory, Visible: true, Enabled: true, Redirect: "/system/dictionary/categories"},
		{ID: "menu-dictionary-categories", ParentID: "menu-dictionary", Name: "Categories", Path: "/system/dictionary/categories", Component: "view/system/dictionary/category/index", Icon: "menu", Sort: 1, Permission: "dictionary:category:list", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-dictionary-items", ParentID: "menu-dictionary", Name: "Items", Path: "/system/dictionary/items", Component: "view/system/dictionary/item/index", Icon: "menu", Sort: 2, Permission: "dictionary:item:list", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-plugins", ParentID: "menu-system", Name: "Plugins", Path: "/system/plugins", Component: "view/plugin/center/index", Icon: "box", Sort: 5, Permission: "plugin:list", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-codegen", ParentID: "menu-system", Name: "CodeGen", Path: "/system/codegen", Component: "view/system/codegen/index", Icon: "magic-stick", Sort: 6, Permission: "codegen:view", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-casbin", ParentID: "menu-system", Name: "Casbin", Path: "/system/casbin", Component: "Layout", Icon: "menu", Sort: 7, Permission: "system:view", Type: model.TypeDirectory, Visible: true, Enabled: true, Redirect: "/system/casbin/overview"},
		{ID: "menu-casbin-overview", ParentID: "menu-casbin", Name: "Overview", Path: "/system/casbin/overview", Component: "view/system/casbin/index", Icon: "menu", Sort: 1, Permission: "casbin:view", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-casbin-models", ParentID: "menu-casbin", Name: "Model Management", Path: "/system/casbin/models", Component: "view/casbin_model/index", Icon: "menu", Sort: 2, Permission: "casbin_model:list", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-casbin-rules", ParentID: "menu-casbin", Name: "Policy Management", Path: "/system/casbin/rules", Component: "view/casbin_rule/index", Icon: "menu", Sort: 3, Permission: "casbin_rule:list", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-nested", Name: "Nested", Path: "/nested", Component: "Layout", Icon: "menu", Sort: 3, Permission: "nested:view", Type: model.TypeDirectory, Visible: true, Enabled: true, Redirect: "/nested/menu1"},
		{ID: "menu-nested-menu1", ParentID: "menu-nested", Name: "Menu1", Path: "/nested/menu1", Component: "view/nested/menu1/index", Icon: "circle", Sort: 1, Permission: "nested:menu1:view", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-nested-menu2", ParentID: "menu-nested", Name: "Menu2", Path: "/nested/menu2", Component: "Layout", Icon: "menu", Sort: 2, Permission: "nested:menu2:view", Type: model.TypeDirectory, Visible: true, Enabled: true, Redirect: "/nested/menu2/one"},
		{ID: "menu-nested-menu2-one", ParentID: "menu-nested-menu2", Name: "Menu2-1", Path: "/nested/menu2/one", Component: "view/nested/menu2/one/index", Icon: "dot", Sort: 1, Permission: "nested:menu2:one:view", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-nested-menu2-two", ParentID: "menu-nested-menu2", Name: "Menu2-2", Path: "/nested/menu2/two", Component: "view/nested/menu2/two/index", Icon: "dot", Sort: 2, Permission: "nested:menu2:two:view", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-nested-menu3", ParentID: "menu-nested", Name: "Menu3", Path: "/nested/menu3", Component: "Layout", Icon: "menu", Sort: 3, Permission: "nested:menu3:view", Type: model.TypeDirectory, Visible: true, Enabled: true, Redirect: "/nested/menu3/one"},
		{ID: "menu-nested-menu3-one", ParentID: "menu-nested-menu3", Name: "Menu3-1", Path: "/nested/menu3/one", Component: "view/nested/menu3/one/index", Icon: "dot", Sort: 1, Permission: "nested:menu3:one:view", Type: model.TypeMenu, Visible: true, Enabled: true},
		{ID: "menu-nested-menu3-two", ParentID: "menu-nested-menu3", Name: "Menu3-2", Path: "/nested/menu3/two", Component: "view/nested/menu3/two/index", Icon: "dot", Sort: 2, Permission: "nested:menu3:two:view", Type: model.TypeMenu, Visible: true, Enabled: true},
	}
}

func containsMenuKeyword(menu *model.Menu, kw string) bool {
	return strings.Contains(strings.ToLower(menu.Name), kw) ||
		strings.Contains(strings.ToLower(menu.Path), kw) ||
		strings.Contains(strings.ToLower(menu.Component), kw) ||
		strings.Contains(strings.ToLower(menu.Permission), kw)
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

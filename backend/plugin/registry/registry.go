package registry

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	pluginiface "goadmin/plugin/interface"
)

type Registry struct {
	mu           sync.RWMutex
	plugins      map[string]string
	activePlugin string
	routes       map[string]pluginiface.Route
	menus        map[string]pluginiface.Menu
	permissions  map[string]pluginiface.Permission
}

func New() *Registry {
	return &Registry{
		plugins:     make(map[string]string),
		routes:      make(map[string]pluginiface.Route),
		menus:       make(map[string]pluginiface.Menu),
		permissions: make(map[string]pluginiface.Permission),
	}
}

func (r *Registry) RegisterPlugin(name string) error {
	if r == nil {
		return fmt.Errorf("plugin registry is not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("plugin name is required")
	}
	key := strings.ToLower(name)

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.plugins[key]; exists {
		return fmt.Errorf("plugin %q already registered", name)
	}
	r.plugins[key] = name
	r.activePlugin = name
	return nil
}

func (r *Registry) AddRoute(route pluginiface.Route) error {
	if r == nil {
		return fmt.Errorf("plugin registry is not configured")
	}
	method := strings.ToUpper(strings.TrimSpace(route.Method))
	path := strings.TrimSpace(route.Path)
	if method == "" {
		return errors.New("route method is required")
	}
	if path == "" {
		return errors.New("route path is required")
	}
	if route.Handler == nil {
		return errors.New("route handler is required")
	}
	if route.Access == "" {
		route.Access = pluginiface.AccessProtected
	}
	if strings.TrimSpace(route.Plugin) == "" {
		route.Plugin = r.activePlugin
	}
	route.Method = method
	route.Path = path
	if strings.TrimSpace(route.Name) == "" {
		route.Name = routeKey(route)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[routeKey(route)] = route
	return nil
}

func (r *Registry) AddMenu(menu pluginiface.Menu) error {
	if r == nil {
		return fmt.Errorf("plugin registry is not configured")
	}
	if strings.TrimSpace(menu.Path) == "" && strings.TrimSpace(menu.ID) == "" {
		return errors.New("menu id or path is required")
	}
	if strings.TrimSpace(menu.Name) == "" {
		return errors.New("menu name is required")
	}
	if menu.Type == "" {
		menu.Type = pluginiface.MenuTypeMenu
	}
	if strings.TrimSpace(menu.Plugin) == "" {
		menu.Plugin = r.activePlugin
	}
	key := menuKey(menu)
	if strings.TrimSpace(menu.ID) == "" {
		menu.ID = key
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.menus[key] = menu
	return nil
}

func (r *Registry) AddPermission(permission pluginiface.Permission) error {
	if r == nil {
		return fmt.Errorf("plugin registry is not configured")
	}
	object := strings.TrimSpace(permission.Object)
	action := strings.TrimSpace(permission.Action)
	if object == "" || action == "" {
		return errors.New("permission object and action are required")
	}
	if strings.TrimSpace(permission.Plugin) == "" {
		permission.Plugin = r.activePlugin
	}
	key := permissionKey(permission)
	if strings.TrimSpace(permission.Description) == "" {
		permission.Description = fmt.Sprintf("%s %s", object, action)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.permissions[key] = permission
	return nil
}

func (r *Registry) PluginNames() []string {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.plugins))
	for _, name := range r.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) Routes() []pluginiface.Route {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.routes))
	for key := range r.routes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	items := make([]pluginiface.Route, 0, len(keys))
	for _, key := range keys {
		items = append(items, r.routes[key])
	}
	return items
}

func (r *Registry) Menus() []pluginiface.Menu {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]pluginiface.Menu, 0, len(r.menus))
	for _, item := range r.menus {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Sort == items[j].Sort {
			if items[i].Name == items[j].Name {
				return items[i].Path < items[j].Path
			}
			return items[i].Name < items[j].Name
		}
		return items[i].Sort < items[j].Sort
	})
	return items
}

func (r *Registry) Permissions() []pluginiface.Permission {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.permissions))
	for key := range r.permissions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	items := make([]pluginiface.Permission, 0, len(keys))
	for _, key := range keys {
		items = append(items, r.permissions[key])
	}
	return items
}

func routeKey(route pluginiface.Route) string {
	return strings.ToLower(strings.TrimSpace(route.Plugin)) + ":" + strings.ToUpper(strings.TrimSpace(route.Method)) + ":" + strings.TrimSpace(route.Path)
}

func menuKey(menu pluginiface.Menu) string {
	plugin := strings.ToLower(strings.TrimSpace(menu.Plugin))
	if id := strings.TrimSpace(menu.ID); id != "" {
		return plugin + ":id:" + strings.ToLower(id)
	}
	return plugin + ":path:" + strings.ToLower(strings.TrimSpace(menu.Path))
}

func permissionKey(permission pluginiface.Permission) string {
	return strings.ToLower(strings.TrimSpace(permission.Plugin)) + ":" + strings.ToLower(strings.TrimSpace(permission.Object)) + ":" + strings.ToLower(strings.TrimSpace(permission.Action))
}

package deletion

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type moduleManifestReferenceIndex struct {
	menuOwners       map[string]map[string]struct{}
	routeOwners      map[string]map[string]struct{}
	permissionOwners map[string]map[string]struct{}
}

func loadModuleManifestReferenceIndex(backendRoot string) (*moduleManifestReferenceIndex, error) {
	backendRoot = filepath.Clean(strings.TrimSpace(backendRoot))
	index := &moduleManifestReferenceIndex{
		menuOwners:       make(map[string]map[string]struct{}),
		routeOwners:      make(map[string]map[string]struct{}),
		permissionOwners: make(map[string]map[string]struct{}),
	}
	if backendRoot == "" || backendRoot == "." {
		return index, nil
	}
	modulesDir := filepath.Join(backendRoot, "modules")
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return index, nil
		}
		return nil, fmt.Errorf("scan modules dir: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		moduleDir := filepath.Join(modulesDir, entry.Name())
		manifest, ok, err := loadModuleManifest(moduleDir)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		moduleName := NormalizeModuleName(manifest.Module)
		if moduleName == "" {
			moduleName = NormalizeModuleName(entry.Name())
		}
		if moduleName == "" {
			continue
		}
		for _, menu := range manifest.Menus {
			path := normalizeRuntimePath(menu.Path)
			if path == "" {
				continue
			}
			index.addOwner(index.menuOwners, path, moduleName)
		}
		for _, route := range manifest.Routes {
			key := routeReferenceKey(route.Method, route.Path)
			if key == "" {
				continue
			}
			index.addOwner(index.routeOwners, key, moduleName)
		}
		for _, permission := range manifest.Permissions {
			key := permissionReferenceKey(permission.Object, permission.Action)
			if key == "" {
				continue
			}
			index.addOwner(index.permissionOwners, key, moduleName)
		}
	}
	return index, nil
}

func (i *moduleManifestReferenceIndex) addOwner(bucket map[string]map[string]struct{}, key, module string) {
	if i == nil {
		return
	}
	key = strings.TrimSpace(key)
	module = NormalizeModuleName(module)
	if key == "" || module == "" {
		return
	}
	owners, ok := bucket[key]
	if !ok {
		owners = make(map[string]struct{})
		bucket[key] = owners
	}
	owners[module] = struct{}{}
}

func (i *moduleManifestReferenceIndex) menuOwnersFor(path string) []string {
	return ownersFromBucket(i.menuOwners, normalizeRuntimePath(path))
}

func (i *moduleManifestReferenceIndex) routeOwnersFor(method, path string) []string {
	return ownersFromBucket(i.routeOwners, routeReferenceKey(method, path))
}

func (i *moduleManifestReferenceIndex) permissionOwnersFor(object, action string) []string {
	return ownersFromBucket(i.permissionOwners, permissionReferenceKey(object, action))
}

func (i *moduleManifestReferenceIndex) menuOwnerCount(path string) int {
	return len(i.menuOwnersFor(path))
}

func (i *moduleManifestReferenceIndex) routeOwnerCount(method, path string) int {
	return len(i.routeOwnersFor(method, path))
}

func (i *moduleManifestReferenceIndex) permissionOwnerCount(object, action string) int {
	return len(i.permissionOwnersFor(object, action))
}

func ownersFromBucket(bucket map[string]map[string]struct{}, key string) []string {
	key = strings.TrimSpace(key)
	if key == "" || len(bucket) == 0 {
		return nil
	}
	ownersSet, ok := bucket[key]
	if !ok || len(ownersSet) == 0 {
		return nil
	}
	owners := make([]string, 0, len(ownersSet))
	for owner := range ownersSet {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	return owners
}

func routeReferenceKey(method, path string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	path = normalizeRuntimePath(path)
	if method == "" || path == "" {
		return ""
	}
	return method + " " + path
}

func permissionReferenceKey(object, action string) string {
	object = strings.TrimSpace(object)
	action = strings.TrimSpace(action)
	if object == "" && action == "" {
		return ""
	}
	return object + " " + action
}

func normalizeRuntimePath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	return filepath.ToSlash(filepath.Clean(value))
}

func loadModuleManifest(moduleDir string) (moduleManifest, bool, error) {
	for _, name := range []string{"manifest.yaml", "manifest.yml", "codegen.manifest.json"} {
		path := filepath.Join(moduleDir, name)
		if !fileExists(path) {
			continue
		}
		doc, err := loadManifestFromPath(path)
		if err != nil {
			return moduleManifest{}, false, err
		}
		return doc, true, nil
	}
	return moduleManifest{}, false, nil
}

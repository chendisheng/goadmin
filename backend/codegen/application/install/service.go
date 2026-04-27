package install

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	menucommand "goadmin/modules/menu/application/command"
	menuservice "goadmin/modules/menu/application/service"
	menumodel "goadmin/modules/menu/domain/model"

	"gopkg.in/yaml.v3"
)

type Dependencies struct {
	MenuService *menuservice.Service
}

type Service struct {
	menus *menuservice.Service
}

type ManifestDocument struct {
	Name        string           `yaml:"name,omitempty"`
	Version     string           `yaml:"version,omitempty"`
	Kind        string           `yaml:"kind,omitempty"`
	Module      string           `yaml:"module,omitempty"`
	Menus       []ManifestMenu   `yaml:"menus,omitempty"`
	Routes      []ManifestRoute  `yaml:"routes,omitempty"`
	Permissions []ManifestPolicy `yaml:"permissions,omitempty"`
}

type ManifestMenu struct {
	Name         string `yaml:"name"`
	TitleKey     string `yaml:"title_key,omitempty"`
	TitleDefault string `yaml:"title_default,omitempty"`
	Path         string `yaml:"path"`
	ParentPath   string `yaml:"parent_path,omitempty"`
	Component    string `yaml:"component,omitempty"`
	Icon         string `yaml:"icon,omitempty"`
	Permission   string `yaml:"permission,omitempty"`
	Type         string `yaml:"type,omitempty"`
	Redirect     string `yaml:"redirect,omitempty"`
	Visible      bool   `yaml:"visible,omitempty"`
	Enabled      bool   `yaml:"enabled,omitempty"`
	Sort         int    `yaml:"sort,omitempty"`
}

type ManifestRoute struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	Name   string `yaml:"name,omitempty"`
}

type ManifestPolicy struct {
	Object      string `yaml:"object,omitempty"`
	Action      string `yaml:"action,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type InstalledMenu struct {
	Path       string `json:"path"`
	ParentPath string `json:"parent_path,omitempty"`
	MenuID     string `json:"menu_id"`
	ParentID   string `json:"parent_id,omitempty"`
	Action     string `json:"action"`
}

type InstallResult struct {
	ManifestPath string          `json:"manifest_path"`
	Name         string          `json:"name,omitempty"`
	Module       string          `json:"module,omitempty"`
	Kind         string          `json:"kind,omitempty"`
	MenuTotal    int             `json:"menu_total"`
	CreatedCount int             `json:"created_count"`
	UpdatedCount int             `json:"updated_count"`
	SkippedCount int             `json:"skipped_count"`
	Menus        []InstalledMenu `json:"menus,omitempty"`
	Messages     []string        `json:"messages,omitempty"`
}

func NewService(deps Dependencies) *Service {
	return &Service{menus: deps.MenuService}
}

func (s *Service) InstallManifest(ctx context.Context, manifestPath string) (InstallResult, error) {
	if s == nil || s.menus == nil {
		return InstallResult{}, fmt.Errorf("manifest install service is not configured")
	}
	manifestPath = strings.TrimSpace(manifestPath)
	if manifestPath == "" {
		return InstallResult{}, fmt.Errorf("manifest path is required")
	}
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return InstallResult{}, fmt.Errorf("read manifest: %w", err)
	}
	var doc ManifestDocument
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return InstallResult{}, fmt.Errorf("parse manifest: %w", err)
	}
	result := InstallResult{
		ManifestPath: manifestPath,
		Name:         strings.TrimSpace(doc.Name),
		Module:       strings.TrimSpace(doc.Module),
		Kind:         strings.TrimSpace(doc.Kind),
		MenuTotal:    len(doc.Menus),
	}
	if len(doc.Menus) == 0 {
		result.Messages = append(result.Messages, "manifest contains no menus")
		return result, nil
	}

	existingTree, err := s.menus.Tree(ctx)
	if err != nil {
		return InstallResult{}, fmt.Errorf("load existing menus: %w", err)
	}
	existingByPath := flattenMenuTree(existingTree)
	manifestByPath := make(map[string]ManifestMenu, len(doc.Menus))
	for _, menu := range doc.Menus {
		normalized, err := normalizeManifestMenu(menu)
		if err != nil {
			return InstallResult{}, err
		}
		if _, exists := manifestByPath[normalized.Path]; exists {
			return InstallResult{}, fmt.Errorf("duplicate manifest menu path: %s", normalized.Path)
		}
		manifestByPath[normalized.Path] = normalized
	}

	installedByPath := make(map[string]*menumodel.Menu, len(doc.Menus))
	visiting := make(map[string]bool, len(doc.Menus))
	var installOne func(string) (*menumodel.Menu, error)
	installOne = func(menuPath string) (*menumodel.Menu, error) {
		menuPath = normalizeMenuPath(menuPath)
		if menuPath == "" {
			return nil, fmt.Errorf("menu path is required")
		}
		if installed, ok := installedByPath[menuPath]; ok {
			return installed, nil
		}
		if visiting[menuPath] {
			return nil, fmt.Errorf("circular manifest menu dependency detected for %s", menuPath)
		}
		spec, ok := manifestByPath[menuPath]
		if !ok {
			if existing, exists := existingByPath[menuPath]; exists {
				return existing, nil
			}
			return nil, fmt.Errorf("manifest menu %s not found", menuPath)
		}
		visiting[menuPath] = true
		defer delete(visiting, menuPath)

		parentID, err := s.resolveParentID(ctx, spec.ParentPath, manifestByPath, existingByPath, installedByPath, installOne)
		if err != nil {
			return nil, err
		}

		if current, ok := existingByPath[menuPath]; ok && current != nil {
			updated, err := s.menus.Update(ctx, current.ID, menucommand.UpdateMenu{
				ParentID:     parentID,
				Name:         spec.Name,
				TitleKey:     spec.TitleKey,
				TitleDefault: spec.TitleDefault,
				Path:         spec.Path,
				Component:    spec.Component,
				Icon:         spec.Icon,
				Sort:         spec.Sort,
				Permission:   spec.Permission,
				Type:         spec.Type,
				Visible:      spec.Visible,
				Enabled:      spec.Enabled,
				Redirect:     spec.Redirect,
				ExternalURL:  "",
			})
			if err != nil {
				return nil, err
			}
			installedByPath[menuPath] = updated
			existingByPath[menuPath] = updated
			result.UpdatedCount++
			result.Menus = append(result.Menus, InstalledMenu{
				Path:       spec.Path,
				ParentPath: spec.ParentPath,
				MenuID:     updated.ID,
				ParentID:   updated.ParentID,
				Action:     "updated",
			})
			return updated, nil
		}

		created, err := s.menus.Create(ctx, menucommand.CreateMenu{
			ParentID:     parentID,
			Name:         spec.Name,
			TitleKey:     spec.TitleKey,
			TitleDefault: spec.TitleDefault,
			Path:         spec.Path,
			Component:    spec.Component,
			Icon:         spec.Icon,
			Sort:         spec.Sort,
			Permission:   spec.Permission,
			Type:         spec.Type,
			Visible:      spec.Visible,
			Enabled:      spec.Enabled,
			Redirect:     spec.Redirect,
			ExternalURL:  "",
		})
		if err != nil {
			return nil, err
		}
		installedByPath[menuPath] = created
		existingByPath[menuPath] = created
		result.CreatedCount++
		result.Menus = append(result.Menus, InstalledMenu{
			Path:       spec.Path,
			ParentPath: spec.ParentPath,
			MenuID:     created.ID,
			ParentID:   created.ParentID,
			Action:     "created",
		})
		return created, nil
	}

	for _, menu := range doc.Menus {
		normalizedPath := normalizeMenuPath(menu.Path)
		if normalizedPath == "" {
			return InstallResult{}, fmt.Errorf("menu path is required")
		}
		if _, err := installOne(normalizedPath); err != nil {
			return InstallResult{}, err
		}
	}
	result.SkippedCount = result.MenuTotal - result.CreatedCount - result.UpdatedCount
	if result.SkippedCount < 0 {
		result.SkippedCount = 0
	}
	result.Messages = append(result.Messages, fmt.Sprintf("installed %d menu(s)", result.CreatedCount+result.UpdatedCount))
	return result, nil
}

func (s *Service) resolveParentID(
	ctx context.Context,
	parentPath string,
	manifestByPath map[string]ManifestMenu,
	existingByPath map[string]*menumodel.Menu,
	installedByPath map[string]*menumodel.Menu,
	installOne func(string) (*menumodel.Menu, error),
) (string, error) {
	parentPath = normalizeMenuPath(parentPath)
	if parentPath == "" {
		return "", nil
	}
	if installed, ok := installedByPath[parentPath]; ok && installed != nil {
		return installed.ID, nil
	}
	if existing, ok := existingByPath[parentPath]; ok && existing != nil {
		return existing.ID, nil
	}
	if _, ok := manifestByPath[parentPath]; ok {
		parentMenu, err := installOne(parentPath)
		if err != nil {
			return "", err
		}
		return parentMenu.ID, nil
	}
	return "", fmt.Errorf("parent menu %s not found", parentPath)
}

func flattenMenuTree(items []menumodel.Menu) map[string]*menumodel.Menu {
	result := make(map[string]*menumodel.Menu)
	var walk func([]menumodel.Menu)
	walk = func(list []menumodel.Menu) {
		for _, item := range list {
			clone := item.Clone()
			result[normalizeMenuPath(clone.Path)] = &clone
			if len(clone.Children) > 0 {
				walk(clone.Children)
			}
		}
	}
	walk(items)
	return result
}

func normalizeManifestMenu(menu ManifestMenu) (ManifestMenu, error) {
	menu.Name = strings.TrimSpace(menu.Name)
	menu.TitleKey = strings.TrimSpace(menu.TitleKey)
	menu.TitleDefault = strings.TrimSpace(menu.TitleDefault)
	menu.Path = normalizeMenuPath(menu.Path)
	menu.ParentPath = normalizeMenuPath(menu.ParentPath)
	menu.Component = strings.TrimSpace(menu.Component)
	menu.Icon = strings.TrimSpace(menu.Icon)
	menu.Permission = strings.TrimSpace(menu.Permission)
	menu.Type = strings.TrimSpace(menu.Type)
	menu.Redirect = normalizeMenuPath(menu.Redirect)
	if menu.Name == "" {
		return ManifestMenu{}, fmt.Errorf("manifest menu name is required")
	}
	if menu.Path == "" {
		return ManifestMenu{}, fmt.Errorf("manifest menu path is required for %s", menu.Name)
	}
	if menu.Type == "" {
		menu.Type = string(menumodel.TypeMenu)
	}
	return menu, nil
}

func normalizeMenuPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	cleaned := path.Clean(value)
	if cleaned == "." {
		return ""
	}
	return cleaned
}

package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	plugincommand "goadmin/plugin/application/command"
	pluginmodel "goadmin/plugin/domain/model"
	pluginrepo "goadmin/plugin/domain/repository"
	pluginiface "goadmin/plugin/interface"
	pluginregistry "goadmin/plugin/registry"
)

type Service struct {
	repo pluginrepo.Repository
}

func New(repo pluginrepo.Repository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("plugin repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context) ([]pluginmodel.Plugin, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("plugin service is not configured")
	}
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, name string) (*pluginmodel.Plugin, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("plugin service is not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	return s.repo.Get(ctx, name)
}

func (s *Service) Create(ctx context.Context, input plugincommand.CreatePlugin) (*pluginmodel.Plugin, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("plugin service is not configured")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	created, err := s.repo.Create(ctx, &pluginmodel.Plugin{
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		Enabled:     input.Enabled,
		Menus:       normalizeMenus(name, input.Menus),
		Permissions: normalizePermissions(name, input.Permissions),
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) Update(ctx context.Context, name string, input plugincommand.UpdatePlugin) (*pluginmodel.Plugin, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("plugin service is not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	current, err := s.repo.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	if input.Description != nil {
		current.Description = strings.TrimSpace(*input.Description)
	}
	if input.Enabled != nil {
		current.Enabled = *input.Enabled
	}
	if input.Menus != nil {
		current.Menus = normalizeMenus(name, input.Menus)
	}
	if input.Permissions != nil {
		current.Permissions = normalizePermissions(name, input.Permissions)
	}
	updated, err := s.repo.Update(ctx, current)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, name string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("plugin service is not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("plugin name is required")
	}
	return s.repo.Delete(ctx, name)
}

func (s *Service) Menus(ctx context.Context) ([]pluginiface.Menu, error) {
	items, _, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]pluginiface.Menu, 0)
	for _, item := range items {
		for _, menu := range item.Menus {
			result = append(result, menu)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Plugin == result[j].Plugin {
			if result[i].Sort == result[j].Sort {
				return result[i].Path < result[j].Path
			}
			return result[i].Sort < result[j].Sort
		}
		return result[i].Plugin < result[j].Plugin
	})
	return result, nil
}

func (s *Service) Permissions(ctx context.Context) ([]pluginiface.Permission, error) {
	items, _, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]pluginiface.Permission, 0)
	for _, item := range items {
		for _, permission := range item.Permissions {
			result = append(result, permission)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Plugin == result[j].Plugin {
			if result[i].Object == result[j].Object {
				return result[i].Action < result[j].Action
			}
			return result[i].Object < result[j].Object
		}
		return result[i].Plugin < result[j].Plugin
	})
	return result, nil
}

func (s *Service) SeedFromRegistry(ctx context.Context, registry *pluginregistry.Registry) error {
	if s == nil || s.repo == nil || registry == nil {
		return nil
	}
	for _, name := range registry.PluginNames() {
		pluginName := strings.TrimSpace(name)
		if pluginName == "" {
			continue
		}
		menus := filterMenusByPlugin(registry.Menus(), pluginName)
		permissions := filterPermissionsByPlugin(registry.Permissions(), pluginName)
		existing, err := s.repo.Get(ctx, pluginName)
		switch {
		case err == nil && existing != nil:
			existing.Menus = menus
			existing.Permissions = permissions
			if _, err := s.repo.Update(ctx, existing); err != nil {
				return err
			}
		case isNotFound(err):
			if _, err := s.repo.Create(ctx, &pluginmodel.Plugin{
				Name:        pluginName,
				Enabled:     true,
				Menus:       menus,
				Permissions: permissions,
			}); err != nil && !isConflict(err) {
				return err
			}
		case err != nil:
			return err
		}
	}
	return nil
}

func normalizeMenus(pluginName string, menus []pluginiface.Menu) []pluginiface.Menu {
	result := make([]pluginiface.Menu, 0, len(menus))
	for _, menu := range menus {
		menu.Plugin = pluginName
		result = append(result, menu)
	}
	return result
}

func normalizePermissions(pluginName string, permissions []pluginiface.Permission) []pluginiface.Permission {
	result := make([]pluginiface.Permission, 0, len(permissions))
	for _, permission := range permissions {
		permission.Plugin = pluginName
		result = append(result, permission)
	}
	return result
}

func filterMenusByPlugin(items []pluginiface.Menu, pluginName string) []pluginiface.Menu {
	result := make([]pluginiface.Menu, 0)
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.Plugin), pluginName) {
			result = append(result, item)
		}
	}
	return result
}

func filterPermissionsByPlugin(items []pluginiface.Permission, pluginName string) []pluginiface.Permission {
	result := make([]pluginiface.Permission, 0)
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.Plugin), pluginName) {
			result = append(result, item)
		}
	}
	return result
}

func isNotFound(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "not found")
}

func isConflict(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "already exists")
}

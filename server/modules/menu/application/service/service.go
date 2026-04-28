package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	apperrors "goadmin/core/errors"
	"goadmin/modules/menu/application/command"
	"goadmin/modules/menu/application/query"
	"goadmin/modules/menu/domain/model"
	menurepo "goadmin/modules/menu/domain/repository"
)

type Service struct {
	repo menurepo.Repository
}

func New(repo menurepo.Repository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("menu repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.ListMenus) ([]model.Menu, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("menu service is not configured")
	}
	return s.repo.List(ctx, menurepo.ListFilter{
		Keyword:  q.Keyword,
		ParentID: q.ParentID,
		Visible:  q.Visible,
		Enabled:  q.Enabled,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}

func (s *Service) Get(ctx context.Context, id string) (*model.Menu, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("menu service is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "menu.id_required", "menu id is required")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, input command.CreateMenu) (*model.Menu, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("menu service is not configured")
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "menu.name_required", "name is required")
	}
	if strings.TrimSpace(input.Path) == "" {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "menu.path_required", "path is required")
	}
	entity := &model.Menu{
		ParentID:     strings.TrimSpace(input.ParentID),
		Name:         strings.TrimSpace(input.Name),
		TitleKey:     strings.TrimSpace(input.TitleKey),
		TitleDefault: strings.TrimSpace(input.TitleDefault),
		Path:         strings.TrimSpace(input.Path),
		Component:    strings.TrimSpace(input.Component),
		Icon:         strings.TrimSpace(input.Icon),
		Sort:         input.Sort,
		Permission:   strings.TrimSpace(input.Permission),
		Type:         normalizeType(input.Type),
		Visible:      input.Visible,
		Enabled:      input.Enabled,
		Redirect:     strings.TrimSpace(input.Redirect),
		ExternalURL:  strings.TrimSpace(input.ExternalURL),
	}
	if entity.Type == "" {
		entity.Type = model.TypeMenu
	}
	created, err := s.repo.Create(ctx, entity)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return created, nil
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateMenu) (*model.Menu, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("menu service is not configured")
	}
	current, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.ParentID) != "" {
		current.ParentID = strings.TrimSpace(input.ParentID)
	}
	if strings.TrimSpace(input.Name) != "" {
		current.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.TitleKey) != "" {
		current.TitleKey = strings.TrimSpace(input.TitleKey)
	}
	if strings.TrimSpace(input.TitleDefault) != "" {
		current.TitleDefault = strings.TrimSpace(input.TitleDefault)
	}
	if strings.TrimSpace(input.Path) != "" {
		current.Path = strings.TrimSpace(input.Path)
	}
	if strings.TrimSpace(input.Component) != "" {
		current.Component = strings.TrimSpace(input.Component)
	}
	if strings.TrimSpace(input.Icon) != "" {
		current.Icon = strings.TrimSpace(input.Icon)
	}
	if input.Sort != 0 {
		current.Sort = input.Sort
	}
	if strings.TrimSpace(input.Permission) != "" {
		current.Permission = strings.TrimSpace(input.Permission)
	}
	if strings.TrimSpace(input.Type) != "" {
		current.Type = normalizeType(input.Type)
	}
	current.Visible = input.Visible
	current.Enabled = input.Enabled
	if strings.TrimSpace(input.Redirect) != "" {
		current.Redirect = strings.TrimSpace(input.Redirect)
	}
	if strings.TrimSpace(input.ExternalURL) != "" {
		current.ExternalURL = strings.TrimSpace(input.ExternalURL)
	}
	updated, err := s.repo.Update(ctx, current)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("menu service is not configured")
	}
	if strings.TrimSpace(id) == "" {
		return apperrors.NewWithKey(apperrors.CodeBadRequest, "menu.id_required", "menu id is required")
	}
	if err := s.repo.Delete(ctx, strings.TrimSpace(id)); err != nil {
		return mapRepositoryError(err)
	}
	return nil
}

func (s *Service) Tree(ctx context.Context) ([]model.Menu, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("menu service is not configured")
	}
	return s.repo.Tree(ctx)
}

func mapRepositoryError(err error) error {
	switch {
	case errors.Is(err, menurepo.ErrNotFound):
		return apperrors.NewWithKey(apperrors.CodeNotFound, "menu.not_found", err.Error())
	case errors.Is(err, menurepo.ErrConflict):
		return apperrors.NewWithKey(apperrors.CodeConflict, "menu.conflict", err.Error())
	default:
		return apperrors.WrapWithKey(err, apperrors.CodeInternal, "menu.operation_failed", "menu operation failed")
	}
}

func normalizeType(value string) model.Type {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(model.TypeDirectory):
		return model.TypeDirectory
	case string(model.TypeButton):
		return model.TypeButton
	default:
		return model.TypeMenu
	}
}

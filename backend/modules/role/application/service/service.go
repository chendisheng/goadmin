package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	apperrors "goadmin/core/errors"
	coretenant "goadmin/core/tenant"
	"goadmin/modules/role/application/command"
	"goadmin/modules/role/application/query"
	"goadmin/modules/role/domain/model"
	rolerepo "goadmin/modules/role/domain/repository"
)

type Service struct {
	repo rolerepo.Repository
}

func New(repo rolerepo.Repository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("role repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.ListRoles) ([]model.Role, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("role service is not configured")
	}
	tenantID, err := coretenant.ResolveTenantID(ctx, q.TenantID)
	if err != nil {
		return nil, 0, apperrors.New(apperrors.CodeForbidden, err.Error())
	}
	return s.repo.List(ctx, rolerepo.ListFilter{
		TenantID: tenantID,
		Keyword:  q.Keyword,
		Status:   q.Status,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}

func (s *Service) Get(ctx context.Context, id string) (*model.Role, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("role service is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "role id is required")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, input command.CreateRole) (*model.Role, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("role service is not configured")
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "name is required")
	}
	if strings.TrimSpace(input.Code) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "code is required")
	}
	tenantID, err := coretenant.ResolveTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeForbidden, err.Error())
	}
	entity := &model.Role{
		TenantID: tenantID,
		Name:     strings.TrimSpace(input.Name),
		Code:     strings.TrimSpace(input.Code),
		Status:   normalizeStatus(input.Status),
		Remark:   strings.TrimSpace(input.Remark),
		MenuIDs:  append([]string(nil), input.MenuIDs...),
	}
	if entity.Status == "" {
		entity.Status = model.StatusActive
	}
	created, err := s.repo.Create(ctx, entity)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return created, nil
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateRole) (*model.Role, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("role service is not configured")
	}
	current, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	tenantID, err := coretenant.ResolveTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeForbidden, err.Error())
	}
	current.TenantID = tenantID
	if strings.TrimSpace(input.Name) != "" {
		current.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Code) != "" {
		current.Code = strings.TrimSpace(input.Code)
	}
	if strings.TrimSpace(input.Status) != "" {
		current.Status = normalizeStatus(input.Status)
	}
	if strings.TrimSpace(input.Remark) != "" {
		current.Remark = strings.TrimSpace(input.Remark)
	}
	if input.MenuIDs != nil {
		current.MenuIDs = append([]string(nil), input.MenuIDs...)
	}
	updated, err := s.repo.Update(ctx, current)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("role service is not configured")
	}
	if strings.TrimSpace(id) == "" {
		return apperrors.New(apperrors.CodeBadRequest, "role id is required")
	}
	if err := s.repo.Delete(ctx, strings.TrimSpace(id)); err != nil {
		return mapRepositoryError(err)
	}
	return nil
}

func mapRepositoryError(err error) error {
	switch {
	case errors.Is(err, coretenant.ErrTenantMismatch):
		return apperrors.New(apperrors.CodeForbidden, err.Error())
	case errors.Is(err, rolerepo.ErrNotFound):
		return apperrors.New(apperrors.CodeNotFound, err.Error())
	case errors.Is(err, rolerepo.ErrConflict):
		return apperrors.New(apperrors.CodeConflict, err.Error())
	default:
		return apperrors.Wrap(err, apperrors.CodeInternal, "role operation failed")
	}
}

func normalizeStatus(value string) model.Status {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(model.StatusInactive):
		return model.StatusInactive
	default:
		return model.StatusActive
	}
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	coreevent "goadmin/core/event"

	apperrors "goadmin/core/errors"
	coretenant "goadmin/core/tenant"
	"goadmin/modules/user/application/command"
	userevent "goadmin/modules/user/application/event"
	"goadmin/modules/user/application/query"
	"goadmin/modules/user/domain/model"
	userrepo "goadmin/modules/user/domain/repository"
)

type Service struct {
	repo userrepo.Repository
	bus  coreevent.Bus
}

func New(repo userrepo.Repository, bus coreevent.Bus) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("user repository is required")
	}
	return &Service{repo: repo, bus: bus}, nil
}

func (s *Service) List(ctx context.Context, q query.ListUsers) ([]model.User, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("user service is not configured")
	}
	tenantID, err := coretenant.ResolveTenantID(ctx, q.TenantID)
	if err != nil {
		return nil, 0, apperrors.New(apperrors.CodeForbidden, err.Error())
	}
	return s.repo.List(ctx, userrepo.ListFilter{
		TenantID: tenantID,
		Keyword:  q.Keyword,
		Status:   q.Status,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}

func (s *Service) Get(ctx context.Context, id string) (*model.User, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("user service is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "user id is required")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, input command.CreateUser) (*model.User, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("user service is not configured")
	}
	if strings.TrimSpace(input.Username) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "username is required")
	}
	tenantID, err := coretenant.ResolveTenantID(ctx, input.TenantID)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeForbidden, err.Error())
	}
	entity := &model.User{
		TenantID:     tenantID,
		Username:     strings.TrimSpace(input.Username),
		DisplayName:  strings.TrimSpace(input.DisplayName),
		Mobile:       strings.TrimSpace(input.Mobile),
		Email:        strings.TrimSpace(input.Email),
		Status:       normalizeStatus(input.Status),
		RoleCodes:    append([]string(nil), input.RoleCodes...),
		PasswordHash: strings.TrimSpace(input.PasswordHash),
	}
	if entity.Status == "" {
		entity.Status = model.StatusActive
	}
	created, err := s.repo.Create(ctx, entity)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	if s.bus != nil {
		_ = s.bus.Publish(ctx, userevent.Created{
			UserID:      created.ID,
			TenantID:    created.TenantID,
			Username:    created.Username,
			DisplayName: created.DisplayName,
			RoleCodes:   append([]string(nil), created.RoleCodes...),
			CreatedAt:   created.CreatedAt,
		})
	}
	return created, nil
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateUser) (*model.User, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("user service is not configured")
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
	if strings.TrimSpace(input.Username) != "" {
		current.Username = strings.TrimSpace(input.Username)
	}
	if strings.TrimSpace(input.DisplayName) != "" {
		current.DisplayName = strings.TrimSpace(input.DisplayName)
	}
	if strings.TrimSpace(input.Mobile) != "" {
		current.Mobile = strings.TrimSpace(input.Mobile)
	}
	if strings.TrimSpace(input.Email) != "" {
		current.Email = strings.TrimSpace(input.Email)
	}
	if strings.TrimSpace(input.Status) != "" {
		current.Status = normalizeStatus(input.Status)
	}
	if input.RoleCodes != nil {
		current.RoleCodes = append([]string(nil), input.RoleCodes...)
	}
	if strings.TrimSpace(input.PasswordHash) != "" {
		current.PasswordHash = strings.TrimSpace(input.PasswordHash)
	}
	updated, err := s.repo.Update(ctx, current)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("user service is not configured")
	}
	if strings.TrimSpace(id) == "" {
		return apperrors.New(apperrors.CodeBadRequest, "user id is required")
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
	case errors.Is(err, userrepo.ErrNotFound):
		return apperrors.New(apperrors.CodeNotFound, err.Error())
	case errors.Is(err, userrepo.ErrConflict):
		return apperrors.New(apperrors.CodeConflict, err.Error())
	default:
		return apperrors.Wrap(err, apperrors.CodeInternal, "user operation failed")
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

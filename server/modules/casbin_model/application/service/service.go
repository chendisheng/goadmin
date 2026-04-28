package service

import (
	"context"

	apperrors "goadmin/core/errors"

	"goadmin/modules/casbin_model/application/command"
	"goadmin/modules/casbin_model/application/query"
	"goadmin/modules/casbin_model/domain/model"
	"goadmin/modules/casbin_model/domain/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) (*Service, error) {
	if repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.repository_required", "casbin_model repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.Listcasbin_models) ([]model.CasbinModel, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.service_not_configured", "casbin_model service is not configured")
	}
	return s.repo.List(ctx, q.Keyword, q.Page, q.PageSize)
}

func (s *Service) Get(ctx context.Context, id string) (*model.CasbinModel, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.service_not_configured", "casbin_model service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		if err.Error() == repository.ErrNotFound.Error() {
			return nil, apperrors.NewWithKey(apperrors.CodeNotFound, "casbin_model.not_found", err.Error())
		}
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "casbin_model.operation_failed", "casbin_model operation failed")
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, input command.CreateCasbinModel) (*model.CasbinModel, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.service_not_configured", "casbin_model service is not configured")
	}
	item := &model.CasbinModel{}
	item.Content = input.Content
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "casbin_model.operation_failed", "casbin_model operation failed")
	}
	return created, nil
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateCasbinModel) (*model.CasbinModel, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.service_not_configured", "casbin_model service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		if err.Error() == repository.ErrNotFound.Error() {
			return nil, apperrors.NewWithKey(apperrors.CodeNotFound, "casbin_model.not_found", err.Error())
		}
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "casbin_model.operation_failed", "casbin_model operation failed")
	}
	cloned := item.Clone()
	item = &cloned
	item.Content = input.Content
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "casbin_model.operation_failed", "casbin_model operation failed")
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.service_not_configured", "casbin_model service is not configured")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if err.Error() == repository.ErrNotFound.Error() {
			return apperrors.NewWithKey(apperrors.CodeNotFound, "casbin_model.not_found", err.Error())
		}
		return apperrors.WrapWithKey(err, apperrors.CodeInternal, "casbin_model.operation_failed", "casbin_model operation failed")
	}
	return nil
}

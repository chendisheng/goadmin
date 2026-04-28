package service

import (
	"context"

	apperrors "goadmin/core/errors"
	"goadmin/modules/casbin_rule/application/command"
	"goadmin/modules/casbin_rule/application/query"
	"goadmin/modules/casbin_rule/domain/model"
	"goadmin/modules/casbin_rule/domain/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) (*Service, error) {
	if repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.repository_required", "casbin_rule repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.Listcasbin_rules) ([]model.CasbinRule, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.service_not_configured", "casbin_rule service is not configured")
	}
	return s.repo.List(ctx, q.Keyword, q.Page, q.PageSize)
}

func (s *Service) Get(ctx context.Context, id string) (*model.CasbinRule, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.service_not_configured", "casbin_rule service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		if err.Error() == repository.ErrNotFound.Error() {
			return nil, apperrors.NewWithKey(apperrors.CodeNotFound, "casbin_rule.not_found", err.Error())
		}
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "casbin_rule.operation_failed", "casbin_rule operation failed")
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, input command.CreateCasbinRule) (*model.CasbinRule, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.service_not_configured", "casbin_rule service is not configured")
	}
	item := &model.CasbinRule{}
	item.Ptype = input.Ptype
	item.V0 = input.V0
	item.V1 = input.V1
	item.V2 = input.V2
	item.V3 = input.V3
	item.V4 = input.V4
	item.V5 = input.V5
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "casbin_rule.operation_failed", "casbin_rule operation failed")
	}
	return created, nil
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateCasbinRule) (*model.CasbinRule, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.service_not_configured", "casbin_rule service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		if err.Error() == repository.ErrNotFound.Error() {
			return nil, apperrors.NewWithKey(apperrors.CodeNotFound, "casbin_rule.not_found", err.Error())
		}
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "casbin_rule.operation_failed", "casbin_rule operation failed")
	}
	cloned := item.Clone()
	item = &cloned
	item.Ptype = input.Ptype
	item.V0 = input.V0
	item.V1 = input.V1
	item.V2 = input.V2
	item.V3 = input.V3
	item.V4 = input.V4
	item.V5 = input.V5
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "casbin_rule.operation_failed", "casbin_rule operation failed")
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.service_not_configured", "casbin_rule service is not configured")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if err.Error() == repository.ErrNotFound.Error() {
			return apperrors.NewWithKey(apperrors.CodeNotFound, "casbin_rule.not_found", err.Error())
		}
		return apperrors.WrapWithKey(err, apperrors.CodeInternal, "casbin_rule.operation_failed", "casbin_rule operation failed")
	}
	return nil
}

package service

import (
	"context"
	"fmt"

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
		return nil, fmt.Errorf("casbin_rule repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.Listcasbin_rules) ([]model.CasbinRule, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("casbin_rule service is not configured")
	}
	return s.repo.List(ctx, q.Keyword, q.Page, q.PageSize)
}

func (s *Service) Get(ctx context.Context, id string) (*model.CasbinRule, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("casbin_rule service is not configured")
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input command.CreateCasbinRule) (*model.CasbinRule, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("casbin_rule service is not configured")
	}
	item := &model.CasbinRule{}
	item.Ptype = input.Ptype
	item.V0 = input.V0
	item.V1 = input.V1
	item.V2 = input.V2
	item.V3 = input.V3
	item.V4 = input.V4
	item.V5 = input.V5
	return s.repo.Create(ctx, item)
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateCasbinRule) (*model.CasbinRule, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("casbin_rule service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
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
	return s.repo.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("casbin_rule service is not configured")
	}
	return s.repo.Delete(ctx, id)
}

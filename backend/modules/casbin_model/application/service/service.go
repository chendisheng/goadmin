package service

import (
	"context"
	"fmt"

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
		return nil, fmt.Errorf("casbin_model repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.Listcasbin_models) ([]model.CasbinModel, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("casbin_model service is not configured")
	}
	return s.repo.List(ctx, q.Keyword, q.Page, q.PageSize)
}

func (s *Service) Get(ctx context.Context, id string) (*model.CasbinModel, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("casbin_model service is not configured")
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input command.CreateCasbinModel) (*model.CasbinModel, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("casbin_model service is not configured")
	}
	item := &model.CasbinModel{}
	item.Content = input.Content
	return s.repo.Create(ctx, item)
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateCasbinModel) (*model.CasbinModel, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("casbin_model service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	cloned := item.Clone()
	item = &cloned
	item.Content = input.Content
	return s.repo.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("casbin_model service is not configured")
	}
	return s.repo.Delete(ctx, id)
}

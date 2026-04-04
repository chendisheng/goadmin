package service

import (
	"context"
	"fmt"

	"goadmin/modules/codegen_console/application/command"
	"goadmin/modules/codegen_console/application/query"
	"goadmin/modules/codegen_console/domain/model"
	"goadmin/modules/codegen_console/domain/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("codegen_console repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.Listcodegen_consoles) ([]model.CodegenConsole, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("codegen_console service is not configured")
	}
	return s.repo.List(ctx, q.Keyword, q.Page, q.PageSize)
}

func (s *Service) Get(ctx context.Context, id string) (*model.CodegenConsole, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("codegen_console service is not configured")
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input command.CreateCodegenConsole) (*model.CodegenConsole, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("codegen_console service is not configured")
	}
	item := &model.CodegenConsole{}
	item.Name = input.Name
	item.Enabled = input.Enabled
	return s.repo.Create(ctx, item)
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateCodegenConsole) (*model.CodegenConsole, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("codegen_console service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	cloned := item.Clone()
	item = &cloned
	item.Name = input.Name
	item.Enabled = input.Enabled
	return s.repo.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("codegen_console service is not configured")
	}
	return s.repo.Delete(ctx, id)
}

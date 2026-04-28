package service

import (
	"context"
	"fmt"

	"goadmin/modules/book/application/command"
	"goadmin/modules/book/application/query"
	"goadmin/modules/book/domain/model"
	"goadmin/modules/book/domain/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("book repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.Listbooks) ([]model.Book, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("book service is not configured")
	}
	return s.repo.List(ctx, q.Keyword, q.Page, q.PageSize)
}

func (s *Service) Get(ctx context.Context, id string) (*model.Book, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("book service is not configured")
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input command.CreateBook) (*model.Book, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("book service is not configured")
	}
	item := &model.Book{}
	item.TenantId = input.TenantId
	item.Title = input.Title
	item.Author = input.Author
	item.Isbn = input.Isbn
	item.Publisher = input.Publisher
	item.PublishDate = input.PublishDate
	item.Category = input.Category
	item.Description = input.Description
	item.Status = input.Status
	item.Price = input.Price
	item.StockQuantity = input.StockQuantity
	item.CoverImageUrl = input.CoverImageUrl
	item.Tags = input.Tags
	return s.repo.Create(ctx, item)
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateBook) (*model.Book, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("book service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	cloned := item.Clone()
	item = &cloned
	item.TenantId = input.TenantId
	item.Title = input.Title
	item.Author = input.Author
	item.Isbn = input.Isbn
	item.Publisher = input.Publisher
	item.PublishDate = input.PublishDate
	item.Category = input.Category
	item.Description = input.Description
	item.Status = input.Status
	item.Price = input.Price
	item.StockQuantity = input.StockQuantity
	item.CoverImageUrl = input.CoverImageUrl
	item.Tags = input.Tags
	return s.repo.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("book service is not configured")
	}
	return s.repo.Delete(ctx, id)
}

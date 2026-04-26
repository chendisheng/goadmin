package service

import (
	"context"
	"errors"

	apperrors "goadmin/core/errors"
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
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "book.repository_required", "book repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.Listbooks) ([]model.Book, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, apperrors.NewWithKey(apperrors.CodeInternal, "book.service_not_configured", "book service is not configured")
	}
	return s.repo.List(ctx, q.Keyword, q.Page, q.PageSize)
}

func (s *Service) Get(ctx context.Context, id string) (*model.Book, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "book.service_not_configured", "book service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, input command.CreateBook) (*model.Book, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "book.service_not_configured", "book service is not configured")
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
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return created, nil
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateBook) (*model.Book, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "book.service_not_configured", "book service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, mapRepositoryError(err)
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
	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "book.service_not_configured", "book service is not configured")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return mapRepositoryError(err)
	}
	return nil
}

func mapRepositoryError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrNotFound):
		return apperrors.NewWithKey(apperrors.CodeNotFound, "book.not_found", err.Error())
	default:
		return apperrors.WrapWithKey(err, apperrors.CodeInternal, "book.operation_failed", "book operation failed")
	}
}

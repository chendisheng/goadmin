package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	apperrors "goadmin/core/errors"
	"goadmin/modules/dictionary/application/command"
	"goadmin/modules/dictionary/application/query"
	dictmodel "goadmin/modules/dictionary/domain/model"
	dictrepo "goadmin/modules/dictionary/domain/repository"
)

type Service struct {
	categoryRepo dictrepo.CategoryRepository
	itemRepo     dictrepo.ItemRepository
}

func New(categoryRepo dictrepo.CategoryRepository, itemRepo dictrepo.ItemRepository) (*Service, error) {
	if categoryRepo == nil {
		return nil, fmt.Errorf("dictionary category repository is required")
	}
	if itemRepo == nil {
		return nil, fmt.Errorf("dictionary item repository is required")
	}
	return &Service{categoryRepo: categoryRepo, itemRepo: itemRepo}, nil
}

func (s *Service) ListCategories(ctx context.Context, q query.ListCategories) ([]dictmodel.Category, int64, error) {
	if s == nil || s.categoryRepo == nil {
		return nil, 0, fmt.Errorf("dictionary service is not configured")
	}
	return s.categoryRepo.List(ctx, dictrepo.CategoryListFilter{
		Keyword:  q.Keyword,
		Status:   q.Status,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
}

func (s *Service) GetCategory(ctx context.Context, id string) (*dictmodel.Category, error) {
	if s == nil || s.categoryRepo == nil {
		return nil, fmt.Errorf("dictionary service is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "dictionary category id is required")
	}
	item, err := s.categoryRepo.Get(ctx, id)
	if err != nil {
		return nil, mapCategoryError(err)
	}
	return item, nil
}

func (s *Service) CreateCategory(ctx context.Context, input command.CreateCategory) (*dictmodel.Category, error) {
	if s == nil || s.categoryRepo == nil {
		return nil, fmt.Errorf("dictionary service is not configured")
	}
	if strings.TrimSpace(input.Code) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "code is required")
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "name is required")
	}
	entity := &dictmodel.Category{
		ID:          strings.TrimSpace(input.ID),
		Code:        strings.TrimSpace(input.Code),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Status:      normalizeStatus(input.Status),
		Sort:        input.Sort,
		Remark:      strings.TrimSpace(input.Remark),
	}
	if entity.Status == "" {
		entity.Status = dictmodel.StatusEnabled
	}
	created, err := s.categoryRepo.Create(ctx, entity)
	if err != nil {
		return nil, mapCategoryError(err)
	}
	return created, nil
}

func (s *Service) UpdateCategory(ctx context.Context, id string, input command.UpdateCategory) (*dictmodel.Category, error) {
	if s == nil || s.categoryRepo == nil {
		return nil, fmt.Errorf("dictionary service is not configured")
	}
	current, err := s.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Code) != "" {
		current.Code = strings.TrimSpace(input.Code)
	}
	if strings.TrimSpace(input.Name) != "" {
		current.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Description) != "" {
		current.Description = strings.TrimSpace(input.Description)
	}
	if strings.TrimSpace(input.Status) != "" {
		current.Status = normalizeStatus(input.Status)
	}
	if input.Sort != 0 {
		current.Sort = input.Sort
	}
	if strings.TrimSpace(input.Remark) != "" {
		current.Remark = strings.TrimSpace(input.Remark)
	}
	updated, err := s.categoryRepo.Update(ctx, current)
	if err != nil {
		return nil, mapCategoryError(err)
	}
	return updated, nil
}

func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	if s == nil || s.categoryRepo == nil {
		return fmt.Errorf("dictionary service is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return apperrors.New(apperrors.CodeBadRequest, "dictionary category id is required")
	}
	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return mapCategoryError(err)
	}
	return nil
}

func (s *Service) ListItems(ctx context.Context, q query.ListItems) ([]dictmodel.Item, int64, error) {
	if s == nil || s.itemRepo == nil {
		return nil, 0, fmt.Errorf("dictionary service is not configured")
	}
	return s.itemRepo.List(ctx, dictrepo.ItemListFilter{
		CategoryID:   q.CategoryID,
		CategoryCode: q.CategoryCode,
		Keyword:      q.Keyword,
		Status:       q.Status,
		Page:         q.Page,
		PageSize:     q.PageSize,
	})
}

func (s *Service) GetItem(ctx context.Context, id string) (*dictmodel.Item, error) {
	if s == nil || s.itemRepo == nil {
		return nil, fmt.Errorf("dictionary service is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "dictionary item id is required")
	}
	item, err := s.itemRepo.Get(ctx, id)
	if err != nil {
		return nil, mapItemError(err)
	}
	return item, nil
}

func (s *Service) CreateItem(ctx context.Context, input command.CreateItem) (*dictmodel.Item, error) {
	if s == nil || s.itemRepo == nil {
		return nil, fmt.Errorf("dictionary service is not configured")
	}
	if strings.TrimSpace(input.CategoryID) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "category id is required")
	}
	if strings.TrimSpace(input.Value) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "value is required")
	}
	if strings.TrimSpace(input.Label) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "label is required")
	}
	entity := &dictmodel.Item{
		ID:         strings.TrimSpace(input.ID),
		CategoryID: strings.TrimSpace(input.CategoryID),
		Value:      strings.TrimSpace(input.Value),
		Label:      strings.TrimSpace(input.Label),
		TagType:    strings.TrimSpace(input.TagType),
		TagColor:   strings.TrimSpace(input.TagColor),
		Extra:      strings.TrimSpace(input.Extra),
		IsDefault:  input.IsDefault,
		Status:     normalizeStatus(input.Status),
		Sort:       input.Sort,
		Remark:     strings.TrimSpace(input.Remark),
	}
	if entity.Status == "" {
		entity.Status = dictmodel.StatusEnabled
	}
	created, err := s.itemRepo.Create(ctx, entity)
	if err != nil {
		return nil, mapItemError(err)
	}
	return created, nil
}

func (s *Service) UpdateItem(ctx context.Context, id string, input command.UpdateItem) (*dictmodel.Item, error) {
	if s == nil || s.itemRepo == nil {
		return nil, fmt.Errorf("dictionary service is not configured")
	}
	current, err := s.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.CategoryID) != "" {
		current.CategoryID = strings.TrimSpace(input.CategoryID)
	}
	if strings.TrimSpace(input.Value) != "" {
		current.Value = strings.TrimSpace(input.Value)
	}
	if strings.TrimSpace(input.Label) != "" {
		current.Label = strings.TrimSpace(input.Label)
	}
	if strings.TrimSpace(input.TagType) != "" {
		current.TagType = strings.TrimSpace(input.TagType)
	}
	if strings.TrimSpace(input.TagColor) != "" {
		current.TagColor = strings.TrimSpace(input.TagColor)
	}
	if strings.TrimSpace(input.Extra) != "" {
		current.Extra = strings.TrimSpace(input.Extra)
	}
	current.IsDefault = input.IsDefault
	if strings.TrimSpace(input.Status) != "" {
		current.Status = normalizeStatus(input.Status)
	}
	if input.Sort != 0 {
		current.Sort = input.Sort
	}
	if strings.TrimSpace(input.Remark) != "" {
		current.Remark = strings.TrimSpace(input.Remark)
	}
	updated, err := s.itemRepo.Update(ctx, current)
	if err != nil {
		return nil, mapItemError(err)
	}
	return updated, nil
}

func (s *Service) DeleteItem(ctx context.Context, id string) error {
	if s == nil || s.itemRepo == nil {
		return fmt.Errorf("dictionary service is not configured")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return apperrors.New(apperrors.CodeBadRequest, "dictionary item id is required")
	}
	if err := s.itemRepo.Delete(ctx, id); err != nil {
		return mapItemError(err)
	}
	return nil
}

func (s *Service) LookupItems(ctx context.Context, q query.LookupItems) ([]dictmodel.Item, error) {
	if s == nil || s.itemRepo == nil {
		return nil, fmt.Errorf("dictionary service is not configured")
	}
	code := strings.TrimSpace(q.CategoryCode)
	if code == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "dictionary category code is required")
	}
	items, err := s.itemRepo.ListByCategoryCode(ctx, code)
	if err != nil {
		return nil, mapItemError(err)
	}
	return items, nil
}

func (s *Service) LookupItem(ctx context.Context, q query.LookupItem) (*dictmodel.Item, error) {
	if s == nil || s.itemRepo == nil {
		return nil, fmt.Errorf("dictionary service is not configured")
	}
	code := strings.TrimSpace(q.CategoryCode)
	if code == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "dictionary category code is required")
	}
	value := strings.TrimSpace(q.Value)
	if value == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "dictionary item value is required")
	}
	item, err := s.itemRepo.GetByCategoryCodeAndValue(ctx, code, value)
	if err != nil {
		return nil, mapItemError(err)
	}
	return item, nil
}

func mapCategoryError(err error) error {
	switch {
	case errors.Is(err, dictrepo.ErrCategoryNotFound):
		return apperrors.New(apperrors.CodeNotFound, err.Error())
	case errors.Is(err, dictrepo.ErrCategoryConflict):
		return apperrors.New(apperrors.CodeConflict, err.Error())
	default:
		return apperrors.Wrap(err, apperrors.CodeInternal, "dictionary category operation failed")
	}
}

func mapItemError(err error) error {
	switch {
	case errors.Is(err, dictrepo.ErrItemNotFound):
		return apperrors.New(apperrors.CodeNotFound, err.Error())
	case errors.Is(err, dictrepo.ErrItemConflict):
		return apperrors.New(apperrors.CodeConflict, err.Error())
	default:
		return apperrors.Wrap(err, apperrors.CodeInternal, "dictionary item operation failed")
	}
}

func normalizeStatus(value string) dictmodel.Status {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(dictmodel.StatusDisabled):
		return dictmodel.StatusDisabled
	case string(dictmodel.StatusEnabled):
		return dictmodel.StatusEnabled
	default:
		return dictmodel.StatusEnabled
	}
}

package repository

import (
	"context"
	"errors"

	"goadmin/modules/dictionary/domain/model"
)

var (
	ErrCategoryNotFound = errors.New("dictionary category not found")
	ErrCategoryConflict = errors.New("dictionary category already exists")
	ErrItemNotFound     = errors.New("dictionary item not found")
	ErrItemConflict     = errors.New("dictionary item already exists")
)

type CategoryListFilter struct {
	Keyword  string
	Status   string
	Page     int
	PageSize int
}

type ItemListFilter struct {
	CategoryID   string
	CategoryCode string
	Keyword      string
	Status       string
	Page         int
	PageSize     int
}

type CategoryRepository interface {
	List(ctx context.Context, filter CategoryListFilter) ([]model.Category, int64, error)
	Get(ctx context.Context, id string) (*model.Category, error)
	Create(ctx context.Context, category *model.Category) (*model.Category, error)
	Update(ctx context.Context, category *model.Category) (*model.Category, error)
	Delete(ctx context.Context, id string) error
}

type ItemRepository interface {
	List(ctx context.Context, filter ItemListFilter) ([]model.Item, int64, error)
	Get(ctx context.Context, id string) (*model.Item, error)
	Create(ctx context.Context, item *model.Item) (*model.Item, error)
	Update(ctx context.Context, item *model.Item) (*model.Item, error)
	Delete(ctx context.Context, id string) error
	ListByCategoryCode(ctx context.Context, categoryCode string) ([]model.Item, error)
	GetByCategoryCodeAndValue(ctx context.Context, categoryCode, value string) (*model.Item, error)
}

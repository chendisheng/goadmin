package repo

import (
	"context"
	"fmt"

	dictmodel "goadmin/modules/dictionary/domain/model"
	dictrepo "goadmin/modules/dictionary/domain/repository"

	"gorm.io/gorm"
)

type CategoryGormRepository struct {
	base *GormRepository
}

type ItemGormRepository struct {
	base *GormRepository
}

func NewCategoryRepository(db *gorm.DB) (dictrepo.CategoryRepository, error) {
	base, err := NewGormRepository(db)
	if err != nil {
		return nil, err
	}
	return &CategoryGormRepository{base: base}, nil
}

func NewItemRepository(db *gorm.DB) (dictrepo.ItemRepository, error) {
	base, err := NewGormRepository(db)
	if err != nil {
		return nil, err
	}
	return &ItemGormRepository{base: base}, nil
}

func (r *CategoryGormRepository) List(ctx context.Context, filter dictrepo.CategoryListFilter) ([]dictmodel.Category, int64, error) {
	if r == nil || r.base == nil {
		return nil, 0, fmt.Errorf("dictionary category repository is not configured")
	}
	return r.base.ListCategories(ctx, filter)
}

func (r *CategoryGormRepository) Get(ctx context.Context, id string) (*dictmodel.Category, error) {
	if r == nil || r.base == nil {
		return nil, fmt.Errorf("dictionary category repository is not configured")
	}
	return r.base.GetCategory(ctx, id)
}

func (r *CategoryGormRepository) Create(ctx context.Context, category *dictmodel.Category) (*dictmodel.Category, error) {
	if r == nil || r.base == nil {
		return nil, fmt.Errorf("dictionary category repository is not configured")
	}
	return r.base.CreateCategory(ctx, category)
}

func (r *CategoryGormRepository) Update(ctx context.Context, category *dictmodel.Category) (*dictmodel.Category, error) {
	if r == nil || r.base == nil {
		return nil, fmt.Errorf("dictionary category repository is not configured")
	}
	return r.base.UpdateCategory(ctx, category)
}

func (r *CategoryGormRepository) Delete(ctx context.Context, id string) error {
	if r == nil || r.base == nil {
		return fmt.Errorf("dictionary category repository is not configured")
	}
	return r.base.DeleteCategory(ctx, id)
}

func (r *ItemGormRepository) List(ctx context.Context, filter dictrepo.ItemListFilter) ([]dictmodel.Item, int64, error) {
	if r == nil || r.base == nil {
		return nil, 0, fmt.Errorf("dictionary item repository is not configured")
	}
	return r.base.ListItems(ctx, filter)
}

func (r *ItemGormRepository) Get(ctx context.Context, id string) (*dictmodel.Item, error) {
	if r == nil || r.base == nil {
		return nil, fmt.Errorf("dictionary item repository is not configured")
	}
	return r.base.GetItem(ctx, id)
}

func (r *ItemGormRepository) Create(ctx context.Context, item *dictmodel.Item) (*dictmodel.Item, error) {
	if r == nil || r.base == nil {
		return nil, fmt.Errorf("dictionary item repository is not configured")
	}
	return r.base.CreateItem(ctx, item)
}

func (r *ItemGormRepository) Update(ctx context.Context, item *dictmodel.Item) (*dictmodel.Item, error) {
	if r == nil || r.base == nil {
		return nil, fmt.Errorf("dictionary item repository is not configured")
	}
	return r.base.UpdateItem(ctx, item)
}

func (r *ItemGormRepository) Delete(ctx context.Context, id string) error {
	if r == nil || r.base == nil {
		return fmt.Errorf("dictionary item repository is not configured")
	}
	return r.base.DeleteItem(ctx, id)
}

func (r *ItemGormRepository) ListByCategoryCode(ctx context.Context, categoryCode string) ([]dictmodel.Item, error) {
	if r == nil || r.base == nil {
		return nil, fmt.Errorf("dictionary item repository is not configured")
	}
	return r.base.ListByCategoryCode(ctx, categoryCode)
}

func (r *ItemGormRepository) GetByCategoryCodeAndValue(ctx context.Context, categoryCode, value string) (*dictmodel.Item, error) {
	if r == nil || r.base == nil {
		return nil, fmt.Errorf("dictionary item repository is not configured")
	}
	return r.base.GetByCategoryCodeAndValue(ctx, categoryCode, value)
}

package repo

import (
	"context"
	"fmt"

	"goadmin/modules/order/domain/model"

	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("order gorm repository requires db")
	}
	return &GormRepository{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("order migrate requires db")
	}
	return db.AutoMigrate(&model.Order{})
}

func (r *GormRepository) List(ctx context.Context, keyword string, page int, pageSize int) ([]model.Order, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("order gorm repository is not configured")
	}
	base := r.db.WithContext(ctx).Model(&model.Order{})
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Order
	if err := base.Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GormRepository) Get(ctx context.Context, id string) (*model.Order, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("order gorm repository is not configured")
	}
	var item model.Order
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Create(ctx context.Context, item *model.Order) (*model.Order, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("order gorm repository is not configured")
	}
	if item == nil {
		return nil, fmt.Errorf("order item is nil")
	}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (r *GormRepository) Update(ctx context.Context, item *model.Order) (*model.Order, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("order gorm repository is not configured")
	}
	if item == nil {
		return nil, fmt.Errorf("order item is nil")
	}
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (r *GormRepository) Delete(ctx context.Context, id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("order gorm repository is not configured")
	}
	if err := r.db.WithContext(ctx).Delete(&model.Order{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

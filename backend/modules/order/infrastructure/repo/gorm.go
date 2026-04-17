package repo

import (
	"context"
	"fmt"
	"strings"

	"time"

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
	if kw := strings.TrimSpace(strings.ToLower(keyword)); kw != "" {
		like := "%" + kw + "%"
		base = base.Where(
			"LOWER(tenant_id) LIKE ? OR LOWER(order_no) LIKE ? OR LOWER(user_id) LIKE ? OR LOWER(customer_name) LIKE ? OR LOWER(customer_email) LIKE ? OR LOWER(customer_phone) LIKE ? OR LOWER(shipping_address) LIKE ? OR LOWER(billing_address) LIKE ? OR LOWER(order_status) LIKE ? OR LOWER(payment_status) LIKE ? OR LOWER(payment_method) LIKE ? OR LOWER(currency) LIKE ? OR LOWER(notes) LIKE ? OR LOWER(internal_notes) LIKE ?",
			like,
			like,
			like,
			like,
			like,
			like,
			like,
			like,
			like,
			like,
			like,
			like,
			like,
			like,
		)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize = normalizePage(page, pageSize)
	var items []model.Order
	if err := base.Order("updated_at DESC, created_at DESC, id ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error; err != nil {
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

	if strings.TrimSpace(item.Id) == "" {
		item.Id = nextRecordID("order")
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

func nextRecordID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return page, pageSize
}

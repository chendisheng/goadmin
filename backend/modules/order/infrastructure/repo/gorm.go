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
	if db.Dialector.Name() == "mysql" && db.Migrator().HasTable(&model.Order{}) {
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN order_no VARCHAR(64)").Error; err != nil {
			return fmt.Errorf("ensure orders.order_no column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN tenant_id VARCHAR(64) NOT NULL").Error; err != nil {
			return fmt.Errorf("ensure orders.tenant_id column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN user_id VARCHAR(64)").Error; err != nil {
			return fmt.Errorf("ensure orders.user_id column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN customer_name VARCHAR(255)").Error; err != nil {
			return fmt.Errorf("ensure orders.customer_name column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN customer_email VARCHAR(255)").Error; err != nil {
			return fmt.Errorf("ensure orders.customer_email column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN customer_phone VARCHAR(32)").Error; err != nil {
			return fmt.Errorf("ensure orders.customer_phone column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN shipping_address VARCHAR(512)").Error; err != nil {
			return fmt.Errorf("ensure orders.shipping_address column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN billing_address VARCHAR(512)").Error; err != nil {
			return fmt.Errorf("ensure orders.billing_address column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN order_status VARCHAR(32)").Error; err != nil {
			return fmt.Errorf("ensure orders.order_status column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN payment_status VARCHAR(32)").Error; err != nil {
			return fmt.Errorf("ensure orders.payment_status column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN payment_method VARCHAR(64)").Error; err != nil {
			return fmt.Errorf("ensure orders.payment_method column: %w", err)
		}
		if err := db.Exec("ALTER TABLE orders MODIFY COLUMN currency VARCHAR(16)").Error; err != nil {
			return fmt.Errorf("ensure orders.currency column: %w", err)
		}
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

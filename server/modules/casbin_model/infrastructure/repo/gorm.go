package repo

import (
	"context"
	"fmt"
	"strings"

	"time"

	apperrors "goadmin/core/errors"
	"goadmin/modules/casbin_model/domain/model"

	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.gorm_repository_required", "casbin_model gorm repository requires db")
	}
	return &GormRepository{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.migrate_requires_db", "casbin_model migrate requires db")
	}
	return db.AutoMigrate(&model.CasbinModel{})
}

func (r *GormRepository) List(ctx context.Context, keyword string, page int, pageSize int) ([]model.CasbinModel, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.gorm_repository_not_configured", "casbin_model gorm repository is not configured")
	}
	base := r.db.WithContext(ctx).Model(&model.CasbinModel{})
	if kw := strings.TrimSpace(strings.ToLower(keyword)); kw != "" {
		like := "%" + kw + "%"
		base = base.Where(
			"LOWER(content) LIKE ?",
			like,
		)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize = normalizePage(page, pageSize)
	var items []model.CasbinModel
	if err := base.Order("updated_at DESC, created_at DESC, name ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GormRepository) Get(ctx context.Context, name string) (*model.CasbinModel, error) {
	if r == nil || r.db == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.gorm_repository_not_configured", "casbin_model gorm repository is not configured")
	}
	var item model.CasbinModel
	if err := r.db.WithContext(ctx).First(&item, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Create(ctx context.Context, item *model.CasbinModel) (*model.CasbinModel, error) {
	if r == nil || r.db == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.gorm_repository_not_configured", "casbin_model gorm repository is not configured")
	}
	if item == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.item_nil", "casbin_model item is nil")
	}

	if strings.TrimSpace(item.Name) == "" {
		item.Name = nextRecordID("casbin_model")
	}

	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (r *GormRepository) Update(ctx context.Context, item *model.CasbinModel) (*model.CasbinModel, error) {
	if r == nil || r.db == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.gorm_repository_not_configured", "casbin_model gorm repository is not configured")
	}
	if item == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.item_nil", "casbin_model item is nil")
	}
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (r *GormRepository) Delete(ctx context.Context, name string) error {
	if r == nil || r.db == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "casbin_model.gorm_repository_not_configured", "casbin_model gorm repository is not configured")
	}
	if err := r.db.WithContext(ctx).Delete(&model.CasbinModel{}, "name = ?", name).Error; err != nil {
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

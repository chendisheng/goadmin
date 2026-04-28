package repo

import (
	"context"
	"strings"

	apperrors "goadmin/core/errors"
	"goadmin/modules/casbin_rule/domain/model"

	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.gorm_repository_required", "casbin_rule gorm repository requires db")
	}
	return &GormRepository{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.migrate_requires_db", "casbin_rule migrate requires db")
	}
	return db.AutoMigrate(&model.CasbinRule{})
}

func (r *GormRepository) List(ctx context.Context, keyword string, page int, pageSize int) ([]model.CasbinRule, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.gorm_repository_not_configured", "casbin_rule gorm repository is not configured")
	}
	base := r.db.WithContext(ctx).Model(&model.CasbinRule{})
	if kw := strings.TrimSpace(strings.ToLower(keyword)); kw != "" {
		like := "%" + kw + "%"
		base = base.Where(
			"LOWER(ptype) LIKE ? OR LOWER(v0) LIKE ? OR LOWER(v1) LIKE ? OR LOWER(v2) LIKE ? OR LOWER(v3) LIKE ? OR LOWER(v4) LIKE ? OR LOWER(v5) LIKE ?",
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
	var items []model.CasbinRule
	if err := base.Order("updated_at DESC, created_at DESC, id ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GormRepository) Get(ctx context.Context, id string) (*model.CasbinRule, error) {
	if r == nil || r.db == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.gorm_repository_not_configured", "casbin_rule gorm repository is not configured")
	}
	var item model.CasbinRule
	if err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Create(ctx context.Context, item *model.CasbinRule) (*model.CasbinRule, error) {
	if r == nil || r.db == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.gorm_repository_not_configured", "casbin_rule gorm repository is not configured")
	}
	if item == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.item_nil", "casbin_rule item is nil")
	}

	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (r *GormRepository) Update(ctx context.Context, item *model.CasbinRule) (*model.CasbinRule, error) {
	if r == nil || r.db == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.gorm_repository_not_configured", "casbin_rule gorm repository is not configured")
	}
	if item == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.item_nil", "casbin_rule item is nil")
	}
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (r *GormRepository) Delete(ctx context.Context, id string) error {
	if r == nil || r.db == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "casbin_rule.gorm_repository_not_configured", "casbin_rule gorm repository is not configured")
	}
	if err := r.db.WithContext(ctx).Delete(&model.CasbinRule{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
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

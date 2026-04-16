package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	dictmodel "goadmin/modules/dictionary/domain/model"
	dictrepo "goadmin/modules/dictionary/domain/repository"

	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

type categoryRecord struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(64);size:64"`
	Code        string    `gorm:"column:code;size:64;not null;uniqueIndex"`
	Name        string    `gorm:"column:name;size:128;not null;index"`
	Description string    `gorm:"column:description;type:text"`
	Status      string    `gorm:"column:status;size:32;not null;default:enabled;index"`
	Sort        int       `gorm:"column:sort;index"`
	Remark      string    `gorm:"column:remark;size:512"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (categoryRecord) TableName() string {
	return "dictionary_categories"
}

type itemRecord struct {
	ID         string    `gorm:"column:id;primaryKey;type:varchar(64);size:64"`
	CategoryID string    `gorm:"column:category_id;type:varchar(64);size:64;index;uniqueIndex:ux_dictionary_item_category_value,priority:1"`
	Value      string    `gorm:"column:value;size:128;not null;index;uniqueIndex:ux_dictionary_item_category_value,priority:2"`
	Label      string    `gorm:"column:label;size:128;not null;index"`
	TagType    string    `gorm:"column:tag_type;size:32"`
	TagColor   string    `gorm:"column:tag_color;size:32"`
	Extra      string    `gorm:"column:extra;type:text"`
	IsDefault  bool      `gorm:"column:is_default;default:false;index"`
	Status     string    `gorm:"column:status;size:32;not null;default:enabled;index"`
	Sort       int       `gorm:"column:sort;index"`
	Remark     string    `gorm:"column:remark;size:512"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (itemRecord) TableName() string {
	return "dictionary_items"
}

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("dictionary gorm repository requires db")
	}
	return &GormRepository{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("dictionary migrate requires db")
	}
	if db.Dialector.Name() == "mysql" {
		if db.Migrator().HasTable(&categoryRecord{}) {
			stmts := []string{
				"ALTER TABLE dictionary_categories MODIFY COLUMN id VARCHAR(64) NOT NULL",
				"ALTER TABLE dictionary_categories MODIFY COLUMN code VARCHAR(64) NOT NULL",
				"ALTER TABLE dictionary_categories MODIFY COLUMN name VARCHAR(128) NOT NULL",
				"ALTER TABLE dictionary_categories MODIFY COLUMN status VARCHAR(32) NOT NULL",
				"ALTER TABLE dictionary_categories MODIFY COLUMN remark VARCHAR(512)",
			}
			for _, stmt := range stmts {
				if err := db.Exec(stmt).Error; err != nil {
					return fmt.Errorf("ensure dictionary_categories schema: %w", err)
				}
			}
		}
		if db.Migrator().HasTable(&itemRecord{}) {
			stmts := []string{
				"ALTER TABLE dictionary_items MODIFY COLUMN id VARCHAR(64) NOT NULL",
				"ALTER TABLE dictionary_items MODIFY COLUMN category_id VARCHAR(64) NOT NULL",
				"ALTER TABLE dictionary_items MODIFY COLUMN value VARCHAR(128) NOT NULL",
				"ALTER TABLE dictionary_items MODIFY COLUMN label VARCHAR(128) NOT NULL",
				"ALTER TABLE dictionary_items MODIFY COLUMN status VARCHAR(32) NOT NULL",
				"ALTER TABLE dictionary_items MODIFY COLUMN remark VARCHAR(512)",
			}
			for _, stmt := range stmts {
				if err := db.Exec(stmt).Error; err != nil {
					return fmt.Errorf("ensure dictionary_items schema: %w", err)
				}
			}
		}
	}
	return db.AutoMigrate(&categoryRecord{}, &itemRecord{})
}

func (r *GormRepository) ListCategories(ctx context.Context, filter dictrepo.CategoryListFilter) ([]dictmodel.Category, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("dictionary gorm repository is not configured")
	}
	base := r.db.WithContext(ctx).Model(&categoryRecord{})
	if kw := strings.TrimSpace(strings.ToLower(filter.Keyword)); kw != "" {
		like := "%" + kw + "%"
		base = base.Where("LOWER(code) LIKE ? OR LOWER(name) LIKE ? OR LOWER(COALESCE(description, '')) LIKE ? OR LOWER(COALESCE(remark, '')) LIKE ? OR LOWER(status) LIKE ?", like, like, like, like, like)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		base = base.Where("status = ?", status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	var rows []categoryRecord
	if err := base.Order("sort ASC, updated_at DESC, created_at DESC, id ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]dictmodel.Category, 0, len(rows))
	for _, row := range rows {
		item, err := row.toModel()
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *GormRepository) GetCategory(ctx context.Context, id string) (*dictmodel.Category, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("dictionary gorm repository is not configured")
	}
	var row categoryRecord
	if err := r.db.WithContext(ctx).First(&row, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return nil, mapCategoryRepoError(err)
	}
	item, err := row.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) CreateCategory(ctx context.Context, category *dictmodel.Category) (*dictmodel.Category, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("dictionary gorm repository is not configured")
	}
	if category == nil {
		return nil, fmt.Errorf("dictionary category is nil")
	}
	record, err := toCategoryRecord(*category)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(record.ID) == "" {
		record.ID = nextRecordID("dictionary-category")
	}
	if exists, err := r.categoryCodeExists(ctx, record.Code, ""); err != nil {
		return nil, err
	} else if exists {
		return nil, dictrepo.ErrCategoryConflict
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, mapCategoryRepoError(err)
	}
	item, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) UpdateCategory(ctx context.Context, category *dictmodel.Category) (*dictmodel.Category, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("dictionary gorm repository is not configured")
	}
	if category == nil {
		return nil, fmt.Errorf("dictionary category is nil")
	}
	record, err := toCategoryRecord(*category)
	if err != nil {
		return nil, err
	}
	var existing categoryRecord
	if err := r.db.WithContext(ctx).First(&existing, "id = ?", strings.TrimSpace(record.ID)).Error; err != nil {
		return nil, mapCategoryRepoError(err)
	}
	if exists, err := r.categoryCodeExists(ctx, record.Code, record.ID); err != nil {
		return nil, err
	} else if exists {
		return nil, dictrepo.ErrCategoryConflict
	}
	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, mapCategoryRepoError(err)
	}
	item, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) DeleteCategory(ctx context.Context, id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("dictionary gorm repository is not configured")
	}
	if err := r.db.WithContext(ctx).Delete(&categoryRecord{}, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return mapCategoryRepoError(err)
	}
	return nil
}

func (r *GormRepository) ListItems(ctx context.Context, filter dictrepo.ItemListFilter) ([]dictmodel.Item, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("dictionary gorm repository is not configured")
	}
	base := r.db.WithContext(ctx).Model(&itemRecord{}).Joins("JOIN dictionary_categories ON dictionary_categories.id = dictionary_items.category_id")
	if categoryID := strings.TrimSpace(filter.CategoryID); categoryID != "" {
		base = base.Where("dictionary_items.category_id = ?", categoryID)
	}
	if categoryCode := strings.TrimSpace(filter.CategoryCode); categoryCode != "" {
		base = base.Where("dictionary_categories.code = ?", categoryCode)
	}
	if kw := strings.TrimSpace(strings.ToLower(filter.Keyword)); kw != "" {
		like := "%" + kw + "%"
		base = base.Where("LOWER(dictionary_items.value) LIKE ? OR LOWER(dictionary_items.label) LIKE ? OR LOWER(COALESCE(dictionary_items.tag_type, '')) LIKE ? OR LOWER(COALESCE(dictionary_items.tag_color, '')) LIKE ? OR LOWER(COALESCE(dictionary_items.extra, '')) LIKE ? OR LOWER(COALESCE(dictionary_items.remark, '')) LIKE ? OR LOWER(dictionary_categories.code) LIKE ? OR LOWER(dictionary_categories.name) LIKE ?", like, like, like, like, like, like, like, like)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		base = base.Where("dictionary_items.status = ?", status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	var rows []itemRecord
	if err := base.Order("dictionary_items.sort ASC, dictionary_items.updated_at DESC, dictionary_items.created_at DESC, dictionary_items.id ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]dictmodel.Item, 0, len(rows))
	for _, row := range rows {
		item, err := row.toModel()
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *GormRepository) GetItem(ctx context.Context, id string) (*dictmodel.Item, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("dictionary gorm repository is not configured")
	}
	var row itemRecord
	if err := r.db.WithContext(ctx).First(&row, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return nil, mapItemRepoError(err)
	}
	item, err := row.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) CreateItem(ctx context.Context, item *dictmodel.Item) (*dictmodel.Item, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("dictionary gorm repository is not configured")
	}
	if item == nil {
		return nil, fmt.Errorf("dictionary item is nil")
	}
	record, err := toItemRecord(*item)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(record.ID) == "" {
		record.ID = nextRecordID("dictionary-item")
	}
	if exists, err := r.itemValueExists(ctx, record.CategoryID, record.Value, ""); err != nil {
		return nil, err
	} else if exists {
		return nil, dictrepo.ErrItemConflict
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, mapItemRepoError(err)
	}
	created, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *GormRepository) UpdateItem(ctx context.Context, item *dictmodel.Item) (*dictmodel.Item, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("dictionary gorm repository is not configured")
	}
	if item == nil {
		return nil, fmt.Errorf("dictionary item is nil")
	}
	record, err := toItemRecord(*item)
	if err != nil {
		return nil, err
	}
	var existing itemRecord
	if err := r.db.WithContext(ctx).First(&existing, "id = ?", strings.TrimSpace(record.ID)).Error; err != nil {
		return nil, mapItemRepoError(err)
	}
	if exists, err := r.itemValueExists(ctx, record.CategoryID, record.Value, record.ID); err != nil {
		return nil, err
	} else if exists {
		return nil, dictrepo.ErrItemConflict
	}
	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, mapItemRepoError(err)
	}
	updated, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *GormRepository) DeleteItem(ctx context.Context, id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("dictionary gorm repository is not configured")
	}
	if err := r.db.WithContext(ctx).Delete(&itemRecord{}, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return mapItemRepoError(err)
	}
	return nil
}

func (r *GormRepository) ListByCategoryCode(ctx context.Context, categoryCode string) ([]dictmodel.Item, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("dictionary gorm repository is not configured")
	}
	var rows []itemRecord
	if err := r.db.WithContext(ctx).
		Model(&itemRecord{}).
		Joins("JOIN dictionary_categories ON dictionary_categories.id = dictionary_items.category_id").
		Where("dictionary_categories.code = ?", strings.TrimSpace(categoryCode)).
		Order("dictionary_items.sort ASC, dictionary_items.value ASC, dictionary_items.id ASC").
		Find(&rows).Error; err != nil {
		return nil, mapItemRepoError(err)
	}
	items := make([]dictmodel.Item, 0, len(rows))
	for _, row := range rows {
		item, err := row.toModel()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *GormRepository) GetByCategoryCodeAndValue(ctx context.Context, categoryCode, value string) (*dictmodel.Item, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("dictionary gorm repository is not configured")
	}
	var row itemRecord
	if err := r.db.WithContext(ctx).
		Model(&itemRecord{}).
		Joins("JOIN dictionary_categories ON dictionary_categories.id = dictionary_items.category_id").
		Where("dictionary_categories.code = ? AND dictionary_items.value = ?", strings.TrimSpace(categoryCode), strings.TrimSpace(value)).
		First(&row).Error; err != nil {
		return nil, mapItemRepoError(err)
	}
	item, err := row.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) categoryCodeExists(ctx context.Context, code, excludeID string) (bool, error) {
	query := r.db.WithContext(ctx).Model(&categoryRecord{}).Where("code = ?", strings.TrimSpace(code))
	if strings.TrimSpace(excludeID) != "" {
		query = query.Where("id <> ?", strings.TrimSpace(excludeID))
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormRepository) itemValueExists(ctx context.Context, categoryID, value, excludeID string) (bool, error) {
	query := r.db.WithContext(ctx).Model(&itemRecord{}).Where("category_id = ? AND value = ?", strings.TrimSpace(categoryID), strings.TrimSpace(value))
	if strings.TrimSpace(excludeID) != "" {
		query = query.Where("id <> ?", strings.TrimSpace(excludeID))
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func toCategoryRecord(category dictmodel.Category) (categoryRecord, error) {
	status := strings.TrimSpace(string(category.Status))
	if status == "" || !dictmodel.Status(status).Valid() {
		status = string(dictmodel.StatusEnabled)
	}
	return categoryRecord{
		ID:          strings.TrimSpace(category.ID),
		Code:        strings.TrimSpace(category.Code),
		Name:        strings.TrimSpace(category.Name),
		Description: strings.TrimSpace(category.Description),
		Status:      status,
		Sort:        category.Sort,
		Remark:      strings.TrimSpace(category.Remark),
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}, nil
}

func (r categoryRecord) toModel() (dictmodel.Category, error) {
	return dictmodel.Category{
		ID:          strings.TrimSpace(r.ID),
		Code:        strings.TrimSpace(r.Code),
		Name:        strings.TrimSpace(r.Name),
		Description: strings.TrimSpace(r.Description),
		Status:      dictmodel.Status(strings.TrimSpace(r.Status)),
		Sort:        r.Sort,
		Remark:      strings.TrimSpace(r.Remark),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}, nil
}

func toItemRecord(item dictmodel.Item) (itemRecord, error) {
	status := strings.TrimSpace(string(item.Status))
	if status == "" || !dictmodel.Status(status).Valid() {
		status = string(dictmodel.StatusEnabled)
	}
	return itemRecord{
		ID:         strings.TrimSpace(item.ID),
		CategoryID: strings.TrimSpace(item.CategoryID),
		Value:      strings.TrimSpace(item.Value),
		Label:      strings.TrimSpace(item.Label),
		TagType:    strings.TrimSpace(item.TagType),
		TagColor:   strings.TrimSpace(item.TagColor),
		Extra:      strings.TrimSpace(item.Extra),
		IsDefault:  item.IsDefault,
		Status:     status,
		Sort:       item.Sort,
		Remark:     strings.TrimSpace(item.Remark),
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}, nil
}

func (r itemRecord) toModel() (dictmodel.Item, error) {
	return dictmodel.Item{
		ID:         strings.TrimSpace(r.ID),
		CategoryID: strings.TrimSpace(r.CategoryID),
		Value:      strings.TrimSpace(r.Value),
		Label:      strings.TrimSpace(r.Label),
		TagType:    strings.TrimSpace(r.TagType),
		TagColor:   strings.TrimSpace(r.TagColor),
		Extra:      strings.TrimSpace(r.Extra),
		IsDefault:  r.IsDefault,
		Status:     dictmodel.Status(strings.TrimSpace(r.Status)),
		Sort:       r.Sort,
		Remark:     strings.TrimSpace(r.Remark),
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}, nil
}

func mapCategoryRepoError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return dictrepo.ErrCategoryNotFound
	case isUniqueConstraintError(err):
		return dictrepo.ErrCategoryConflict
	default:
		return err
	}
}

func mapItemRepoError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return dictrepo.ErrItemNotFound
	case isUniqueConstraintError(err):
		return dictrepo.ErrItemConflict
	default:
		return err
	}
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique constraint failed") || strings.Contains(message, "duplicate entry") || strings.Contains(message, "duplicate key")
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

package repo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"goadmin/modules/menu/domain/model"
	menurepo "goadmin/modules/menu/domain/repository"

	"gorm.io/gorm"
)

type GormRepository struct{ db *gorm.DB }

type menuRecord struct {
	ID          string    `gorm:"column:id;primaryKey;size:64"`
	ParentID    string    `gorm:"column:parent_id;size:64;index"`
	Name        string    `gorm:"column:name;size:128;not null"`
	Path        string    `gorm:"column:path;size:255;not null;uniqueIndex"`
	Component   string    `gorm:"column:component;size:255"`
	Icon        string    `gorm:"column:icon;size:128"`
	Sort        int       `gorm:"column:sort;index"`
	Permission  string    `gorm:"column:permission;size:255;index"`
	Type        string    `gorm:"column:type;size:32;not null;index"`
	Visible     bool      `gorm:"column:visible;index"`
	Enabled     bool      `gorm:"column:enabled;index"`
	Redirect    string    `gorm:"column:redirect;size:255"`
	ExternalURL string    `gorm:"column:external_url;size:255"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (menuRecord) TableName() string { return "menu" }

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("menu gorm repository requires db")
	}
	return &GormRepository{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("menu migrate requires db")
	}
	return db.AutoMigrate(&menuRecord{})
}

func SeedDefaults(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("menu seed requires db")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := cleanupLegacyMenuPaths(tx, []string{
			"/casbin_models",
			"/casbin_models/list",
			"/casbin_rules",
			"/casbin_rules/list",
			"/system/casbin",
			"/system/casbin/overview",
			"/system/casbin/models",
			"/system/casbin/rules",
		}); err != nil {
			return err
		}
		now := time.Now().UTC()
		seedIDMap := make(map[string]string, len(defaultMenus()))
		for _, menu := range defaultMenus() {
			actualID, err := syncSeedMenu(tx, menu, now, seedIDMap)
			if err != nil {
				return err
			}
			if seedID := strings.TrimSpace(menu.ID); seedID != "" {
				seedIDMap[seedID] = actualID
			}
		}
		return nil
	})
}

func syncSeedMenu(tx *gorm.DB, menu model.Menu, now time.Time, seedIDMap map[string]string) (string, error) {
	record, err := toMenuRecord(menu)
	if err != nil {
		return "", err
	}
	if parentID := strings.TrimSpace(record.ParentID); parentID != "" {
		if mappedParentID, ok := seedIDMap[parentID]; ok && strings.TrimSpace(mappedParentID) != "" {
			record.ParentID = mappedParentID
		}
	}
	if strings.TrimSpace(record.Type) == "" {
		record.Type = string(model.TypeMenu)
	}

	if existing, found, err := findSeedMenuRecord(tx, record.ID, record.Path); err != nil {
		return "", err
	} else if found {
		actualID := strings.TrimSpace(existing.ID)
		if actualID == "" {
			actualID = strings.TrimSpace(record.ID)
		}
		updates := map[string]any{
			"parent_id":    record.ParentID,
			"name":         record.Name,
			"path":         record.Path,
			"component":    record.Component,
			"icon":         record.Icon,
			"sort":         record.Sort,
			"permission":   record.Permission,
			"type":         record.Type,
			"visible":      record.Visible,
			"enabled":      record.Enabled,
			"redirect":     record.Redirect,
			"external_url": record.ExternalURL,
			"updated_at":   now,
		}
		if err := tx.Model(&menuRecord{}).Where("id = ?", actualID).Updates(updates).Error; err != nil {
			return "", mapMenuRepoError(err)
		}
		return actualID, nil
	}

	if strings.TrimSpace(record.ID) == "" {
		record.ID = nextRecordID("menu")
	}
	record.CreatedAt = now
	record.UpdatedAt = now
	if err := tx.Create(&record).Error; err != nil {
		return "", mapMenuRepoError(err)
	}
	return record.ID, nil
}

func findSeedMenuRecord(tx *gorm.DB, id, path string) (menuRecord, bool, error) {
	if trimmedID := strings.TrimSpace(id); trimmedID != "" {
		var existing menuRecord
		err := tx.First(&existing, "id = ?", trimmedID).Error
		switch {
		case err == nil:
			return existing, true, nil
		case errors.Is(err, gorm.ErrRecordNotFound):
		default:
			return menuRecord{}, false, mapMenuRepoError(err)
		}
	}

	if trimmedPath := strings.TrimSpace(path); trimmedPath != "" {
		var existing menuRecord
		err := tx.First(&existing, "path = ?", trimmedPath).Error
		switch {
		case err == nil:
			return existing, true, nil
		case errors.Is(err, gorm.ErrRecordNotFound):
		default:
			return menuRecord{}, false, mapMenuRepoError(err)
		}
	}

	return menuRecord{}, false, nil
}

func (r *GormRepository) List(ctx context.Context, filter menurepo.ListFilter) ([]model.Menu, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("menu gorm repository is not configured")
	}
	base := r.applyFilters(r.db.WithContext(ctx).Model(&menuRecord{}), filter)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(filter.Page, filter.PageSize)
	var rows []menuRecord
	if err := base.Order("sort ASC, updated_at DESC, id ASC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items, err := mapMenuRecords(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GormRepository) Get(ctx context.Context, id string) (*model.Menu, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("menu gorm repository is not configured")
	}
	var row menuRecord
	if err := r.db.WithContext(ctx).First(&row, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return nil, mapMenuRepoError(err)
	}
	item, err := row.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Create(ctx context.Context, menu *model.Menu) (*model.Menu, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("menu gorm repository is not configured")
	}
	if menu == nil {
		return nil, fmt.Errorf("menu is nil")
	}
	record, err := toMenuRecord(*menu)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(record.ID) == "" {
		record.ID = nextRecordID("menu")
	}
	if exists, err := r.exists(ctx, record.Path, ""); err != nil {
		return nil, err
	} else if exists {
		return nil, menurepo.ErrConflict
	}
	if record.Type == "" {
		record.Type = string(model.TypeMenu)
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, mapMenuRepoError(err)
	}
	item, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Update(ctx context.Context, menu *model.Menu) (*model.Menu, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("menu gorm repository is not configured")
	}
	if menu == nil {
		return nil, fmt.Errorf("menu is nil")
	}
	record, err := toMenuRecord(*menu)
	if err != nil {
		return nil, err
	}
	var existing menuRecord
	if err := r.db.WithContext(ctx).First(&existing, "id = ?", strings.TrimSpace(record.ID)).Error; err != nil {
		return nil, mapMenuRepoError(err)
	}
	if exists, err := r.exists(ctx, record.Path, record.ID); err != nil {
		return nil, err
	} else if exists {
		return nil, menurepo.ErrConflict
	}
	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, mapMenuRepoError(err)
	}
	item, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Delete(ctx context.Context, id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("menu gorm repository is not configured")
	}
	result := r.db.WithContext(ctx).Delete(&menuRecord{}, "id = ?", strings.TrimSpace(id))
	if result.Error != nil {
		return mapMenuRepoError(result.Error)
	}
	if result.RowsAffected == 0 {
		return menurepo.ErrNotFound
	}
	return nil
}

func (r *GormRepository) Tree(ctx context.Context) ([]model.Menu, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("menu gorm repository is not configured")
	}
	var rows []menuRecord
	if err := r.db.WithContext(ctx).Order("sort ASC, updated_at DESC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	items, err := mapMenuRecords(rows)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(items), nil
}

func (r *GormRepository) applyFilters(db *gorm.DB, filter menurepo.ListFilter) *gorm.DB {
	if parentID := strings.TrimSpace(filter.ParentID); parentID != "" {
		db = db.Where("parent_id = ?", parentID)
	}
	if filter.Visible != nil {
		db = db.Where("visible = ?", *filter.Visible)
	}
	if filter.Enabled != nil {
		db = db.Where("enabled = ?", *filter.Enabled)
	}
	if kw := strings.TrimSpace(strings.ToLower(filter.Keyword)); kw != "" {
		like := "%" + kw + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(path) LIKE ? OR LOWER(component) LIKE ? OR LOWER(permission) LIKE ? OR LOWER(icon) LIKE ?", like, like, like, like, like)
	}
	return db
}

func toMenuRecord(menu model.Menu) (menuRecord, error) {
	return menuRecord{
		ID:          strings.TrimSpace(menu.ID),
		ParentID:    strings.TrimSpace(menu.ParentID),
		Name:        strings.TrimSpace(menu.Name),
		Path:        strings.TrimSpace(menu.Path),
		Component:   strings.TrimSpace(menu.Component),
		Icon:        strings.TrimSpace(menu.Icon),
		Sort:        menu.Sort,
		Permission:  strings.TrimSpace(menu.Permission),
		Type:        string(menu.Type),
		Visible:     menu.Visible,
		Enabled:     menu.Enabled,
		Redirect:    strings.TrimSpace(menu.Redirect),
		ExternalURL: strings.TrimSpace(menu.ExternalURL),
		CreatedAt:   menu.CreatedAt,
		UpdatedAt:   menu.UpdatedAt,
	}, nil
}

func (r menuRecord) toModel() (model.Menu, error) {
	return model.Menu{
		ID:          strings.TrimSpace(r.ID),
		ParentID:    strings.TrimSpace(r.ParentID),
		Name:        strings.TrimSpace(r.Name),
		Path:        strings.TrimSpace(r.Path),
		Component:   strings.TrimSpace(r.Component),
		Icon:        strings.TrimSpace(r.Icon),
		Sort:        r.Sort,
		Permission:  strings.TrimSpace(r.Permission),
		Type:        model.Type(strings.TrimSpace(r.Type)),
		Visible:     r.Visible,
		Enabled:     r.Enabled,
		Redirect:    strings.TrimSpace(r.Redirect),
		ExternalURL: strings.TrimSpace(r.ExternalURL),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}, nil
}

func mapMenuRecords(rows []menuRecord) ([]model.Menu, error) {
	result := make([]model.Menu, 0, len(rows))
	for _, row := range rows {
		item, err := row.toModel()
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *GormRepository) exists(ctx context.Context, pathValue, excludeID string) (bool, error) {
	if r == nil || r.db == nil {
		return false, fmt.Errorf("menu gorm repository is not configured")
	}
	query := r.db.WithContext(ctx).Model(&menuRecord{}).Where("path = ?", strings.TrimSpace(pathValue))
	if trimmedID := strings.TrimSpace(excludeID); trimmedID != "" {
		query = query.Where("id <> ?", trimmedID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, mapMenuRepoError(err)
	}
	return count > 0, nil
}

func buildMenuTree(items []model.Menu) []model.Menu {
	byParent := make(map[string][]model.Menu)
	for _, item := range items {
		byParent[item.ParentID] = append(byParent[item.ParentID], item)
	}
	for key := range byParent {
		sort.Slice(byParent[key], func(i, j int) bool {
			if byParent[key][i].Sort == byParent[key][j].Sort {
				return byParent[key][i].Name < byParent[key][j].Name
			}
			return byParent[key][i].Sort < byParent[key][j].Sort
		})
	}
	return buildMenuChildren("", byParent)
}

func buildMenuChildren(parentID string, byParent map[string][]model.Menu) []model.Menu {
	children := byParent[parentID]
	result := make([]model.Menu, 0, len(children))
	for _, child := range children {
		clone := child.Clone()
		clone.Children = buildMenuChildren(child.ID, byParent)
		result = append(result, clone)
	}
	return result
}

func mapMenuRepoError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return menurepo.ErrNotFound
	case strings.Contains(strings.ToLower(err.Error()), "unique constraint failed"):
		return menurepo.ErrConflict
	default:
		return err
	}
}

func cleanupLegacyMenuPaths(tx *gorm.DB, paths []string) error {
	if tx == nil || len(paths) == 0 {
		return nil
	}
	for _, pathValue := range paths {
		pathValue = strings.TrimSpace(pathValue)
		if pathValue == "" {
			continue
		}
		if err := tx.Where("path = ?", pathValue).Delete(&menuRecord{}).Error; err != nil {
			return mapMenuRepoError(err)
		}
	}
	return nil
}

func nextRecordID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}

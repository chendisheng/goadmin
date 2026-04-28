package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	pluginmodel "goadmin/plugin/domain/model"
	pluginrepo "goadmin/plugin/domain/repository"
	pluginiface "goadmin/plugin/interface"

	"gorm.io/gorm"
)

type GormRepository struct{ db *gorm.DB }

type pluginRecord struct {
	Name           string    `gorm:"column:name;primaryKey;size:128"`
	Description    string    `gorm:"column:description;type:text"`
	Enabled        bool      `gorm:"column:enabled;index"`
	MenusRaw       string    `gorm:"column:menus_json;type:text;not null"`
	PermissionsRaw string    `gorm:"column:permissions_json;type:text;not null"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (pluginRecord) TableName() string { return "plugin" }

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("plugin gorm repository requires db")
	}
	return &GormRepository{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("plugin migrate requires db")
	}
	return db.AutoMigrate(&pluginRecord{})
}

func (r *GormRepository) List(ctx context.Context) ([]pluginmodel.Plugin, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("plugin gorm repository is not configured")
	}
	base := r.db.WithContext(ctx).Model(&pluginRecord{})
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []pluginRecord
	if err := base.Order("updated_at DESC, created_at DESC, name ASC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]pluginmodel.Plugin, 0, len(rows))
	for _, row := range rows {
		item, err := row.toModel()
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *GormRepository) Get(ctx context.Context, name string) (*pluginmodel.Plugin, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("plugin gorm repository is not configured")
	}
	var row pluginRecord
	if err := r.db.WithContext(ctx).First(&row, "name = ?", strings.TrimSpace(name)).Error; err != nil {
		return nil, mapRepoError(err)
	}
	item, err := row.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Create(ctx context.Context, plugin *pluginmodel.Plugin) (*pluginmodel.Plugin, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("plugin gorm repository is not configured")
	}
	if plugin == nil {
		return nil, fmt.Errorf("plugin is nil")
	}
	record, err := toRecord(*plugin)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(record.Name) == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	if exists, err := r.exists(ctx, record.Name); err != nil {
		return nil, err
	} else if exists {
		return nil, pluginrepo.ErrConflict
	}
	record.CreatedAt = time.Now().UTC()
	record.UpdatedAt = record.CreatedAt
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, mapRepoError(err)
	}
	item, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Update(ctx context.Context, plugin *pluginmodel.Plugin) (*pluginmodel.Plugin, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("plugin gorm repository is not configured")
	}
	if plugin == nil {
		return nil, fmt.Errorf("plugin is nil")
	}
	record, err := toRecord(*plugin)
	if err != nil {
		return nil, err
	}
	var existing pluginRecord
	if err := r.db.WithContext(ctx).First(&existing, "name = ?", strings.TrimSpace(record.Name)).Error; err != nil {
		return nil, mapRepoError(err)
	}
	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, mapRepoError(err)
	}
	item, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Delete(ctx context.Context, name string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("plugin gorm repository is not configured")
	}
	result := r.db.WithContext(ctx).Delete(&pluginRecord{}, "name = ?", strings.TrimSpace(name))
	if result.Error != nil {
		return mapRepoError(result.Error)
	}
	if result.RowsAffected == 0 {
		return pluginrepo.ErrNotFound
	}
	return nil
}

func (r *GormRepository) exists(ctx context.Context, name string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&pluginRecord{}).Where("name = ?", strings.TrimSpace(name)).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func toRecord(plugin pluginmodel.Plugin) (pluginRecord, error) {
	menus, err := json.Marshal(plugin.Menus)
	if err != nil {
		return pluginRecord{}, fmt.Errorf("marshal plugin menus: %w", err)
	}
	permissions, err := json.Marshal(plugin.Permissions)
	if err != nil {
		return pluginRecord{}, fmt.Errorf("marshal plugin permissions: %w", err)
	}
	return pluginRecord{
		Name:           strings.TrimSpace(plugin.Name),
		Description:    strings.TrimSpace(plugin.Description),
		Enabled:        plugin.Enabled,
		MenusRaw:       string(menus),
		PermissionsRaw: string(permissions),
		CreatedAt:      plugin.CreatedAt,
		UpdatedAt:      plugin.UpdatedAt,
	}, nil
}

func (r pluginRecord) toModel() (pluginmodel.Plugin, error) {
	var menus []pluginiface.Menu
	if strings.TrimSpace(r.MenusRaw) != "" {
		if err := json.Unmarshal([]byte(r.MenusRaw), &menus); err != nil {
			return pluginmodel.Plugin{}, fmt.Errorf("unmarshal plugin menus: %w", err)
		}
	}
	var permissions []pluginiface.Permission
	if strings.TrimSpace(r.PermissionsRaw) != "" {
		if err := json.Unmarshal([]byte(r.PermissionsRaw), &permissions); err != nil {
			return pluginmodel.Plugin{}, fmt.Errorf("unmarshal plugin permissions: %w", err)
		}
	}
	return pluginmodel.Plugin{
		Name:        strings.TrimSpace(r.Name),
		Description: strings.TrimSpace(r.Description),
		Enabled:     r.Enabled,
		Menus:       menus,
		Permissions: permissions,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}, nil
}

func mapRepoError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return pluginrepo.ErrNotFound
	default:
		return err
	}
}

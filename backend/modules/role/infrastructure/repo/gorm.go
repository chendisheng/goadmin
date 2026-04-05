package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	coretenant "goadmin/core/tenant"
	"goadmin/modules/role/domain/model"
	rolerepo "goadmin/modules/role/domain/repository"

	"gorm.io/gorm"
)

type GormRepository struct{ db *gorm.DB }

type roleRecord struct {
	ID         string    `gorm:"column:id;primaryKey;size:64"`
	TenantID   string    `gorm:"column:tenant_id;size:64;index:idx_roles_tenant_code,priority:1"`
	Name       string    `gorm:"column:name;size:128;not null"`
	Code       string    `gorm:"column:code;size:128;not null;index:idx_roles_tenant_code,priority:2"`
	Status     string    `gorm:"column:status;size:32;not null;index"`
	Remark     string    `gorm:"column:remark;type:text"`
	MenuIDsRaw string    `gorm:"column:menu_ids;type:text;not null"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (roleRecord) TableName() string { return "roles" }

func NewGormRepository(db *gorm.DB) (*GormRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("role gorm repository requires db")
	}
	return &GormRepository{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("role migrate requires db")
	}
	if db.Dialector.Name() == "mysql" && db.Migrator().HasTable(&roleRecord{}) {
		if err := db.Exec("ALTER TABLE roles MODIFY COLUMN tenant_id VARCHAR(64) NOT NULL").Error; err != nil {
			return fmt.Errorf("ensure roles.tenant_id column: %w", err)
		}
	}
	return db.AutoMigrate(&roleRecord{})
}

func (r *GormRepository) List(ctx context.Context, filter rolerepo.ListFilter) ([]model.Role, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("role gorm repository is not configured")
	}
	if strings.TrimSpace(filter.TenantID) == "" {
		if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok {
			filter.TenantID = tenantID
		}
	}
	base := r.applyFilters(r.db.WithContext(ctx).Model(&roleRecord{}), filter)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(filter.Page, filter.PageSize)
	var rows []roleRecord
	if err := base.Order("updated_at DESC, created_at DESC, id ASC").Limit(size).Offset((page - 1) * size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items, err := mapRoleRecords(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GormRepository) Get(ctx context.Context, id string) (*model.Role, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("role gorm repository is not configured")
	}
	var row roleRecord
	query := r.db.WithContext(ctx).Model(&roleRecord{})
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&row, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return nil, mapRoleRepoError(err)
	}
	item, err := row.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Create(ctx context.Context, role *model.Role) (*model.Role, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("role gorm repository is not configured")
	}
	if role == nil {
		return nil, fmt.Errorf("role is nil")
	}
	record, err := toRoleRecord(*role)
	if err != nil {
		return nil, err
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		if strings.TrimSpace(record.TenantID) == "" {
			record.TenantID = tenantID
		} else if strings.TrimSpace(record.TenantID) != tenantID {
			return nil, coretenant.ErrTenantMismatch
		}
	}
	if strings.TrimSpace(record.ID) == "" {
		record.ID = nextRecordID("role")
	}
	if exists, err := r.exists(ctx, record.TenantID, record.Code, ""); err != nil {
		return nil, err
	} else if exists {
		return nil, rolerepo.ErrConflict
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return nil, mapRoleRepoError(err)
	}
	item, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Update(ctx context.Context, role *model.Role) (*model.Role, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("role gorm repository is not configured")
	}
	if role == nil {
		return nil, fmt.Errorf("role is nil")
	}
	record, err := toRoleRecord(*role)
	if err != nil {
		return nil, err
	}
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		if strings.TrimSpace(record.TenantID) == "" {
			record.TenantID = tenantID
		} else if strings.TrimSpace(record.TenantID) != tenantID {
			return nil, coretenant.ErrTenantMismatch
		}
	}
	var existing roleRecord
	query := r.db.WithContext(ctx).Model(&roleRecord{})
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing, "id = ?", strings.TrimSpace(record.ID)).Error; err != nil {
		return nil, mapRoleRepoError(err)
	}
	if exists, err := r.exists(ctx, record.TenantID, record.Code, record.ID); err != nil {
		return nil, err
	} else if exists {
		return nil, rolerepo.ErrConflict
	}
	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, mapRoleRepoError(err)
	}
	item, err := record.toModel()
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormRepository) Delete(ctx context.Context, id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("role gorm repository is not configured")
	}
	query := r.db.WithContext(ctx).Model(&roleRecord{})
	if tenantID, ok := coretenant.TenantIDFromContext(ctx); ok && strings.TrimSpace(tenantID) != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	result := query.Delete(&roleRecord{}, "id = ?", strings.TrimSpace(id))
	if result.Error != nil {
		return mapRoleRepoError(result.Error)
	}
	if result.RowsAffected == 0 {
		return rolerepo.ErrNotFound
	}
	return nil
}

func (r *GormRepository) applyFilters(db *gorm.DB, filter rolerepo.ListFilter) *gorm.DB {
	if tenant := strings.TrimSpace(filter.TenantID); tenant != "" {
		db = db.Where("tenant_id = ?", tenant)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		db = db.Where("status = ?", strings.ToLower(status))
	}
	if kw := strings.TrimSpace(strings.ToLower(filter.Keyword)); kw != "" {
		like := "%" + kw + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ? OR LOWER(remark) LIKE ? OR LOWER(tenant_id) LIKE ?", like, like, like, like)
	}
	return db
}

func (r *GormRepository) exists(ctx context.Context, tenantID, code, excludeID string) (bool, error) {
	query := r.db.WithContext(ctx).Model(&roleRecord{}).Where("tenant_id = ? AND code = ?", strings.TrimSpace(tenantID), strings.TrimSpace(code))
	if strings.TrimSpace(excludeID) != "" {
		query = query.Where("id <> ?", strings.TrimSpace(excludeID))
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func toRoleRecord(role model.Role) (roleRecord, error) {
	menuIDs, err := json.Marshal(append([]string(nil), role.MenuIDs...))
	if err != nil {
		return roleRecord{}, fmt.Errorf("marshal role menu ids: %w", err)
	}
	status := strings.TrimSpace(string(role.Status))
	if status == "" {
		status = string(model.StatusActive)
	}
	return roleRecord{ID: strings.TrimSpace(role.ID), TenantID: strings.TrimSpace(role.TenantID), Name: strings.TrimSpace(role.Name), Code: strings.TrimSpace(role.Code), Status: status, Remark: strings.TrimSpace(role.Remark), MenuIDsRaw: string(menuIDs), CreatedAt: role.CreatedAt, UpdatedAt: role.UpdatedAt}, nil
}

func (r roleRecord) toModel() (model.Role, error) {
	var menuIDs []string
	if strings.TrimSpace(r.MenuIDsRaw) != "" {
		if err := json.Unmarshal([]byte(r.MenuIDsRaw), &menuIDs); err != nil {
			return model.Role{}, fmt.Errorf("unmarshal role menu ids: %w", err)
		}
	}
	return model.Role{ID: strings.TrimSpace(r.ID), TenantID: strings.TrimSpace(r.TenantID), Name: strings.TrimSpace(r.Name), Code: strings.TrimSpace(r.Code), Status: model.Status(strings.TrimSpace(r.Status)), Remark: strings.TrimSpace(r.Remark), MenuIDs: append([]string(nil), menuIDs...), CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}, nil
}

func mapRoleRecords(rows []roleRecord) ([]model.Role, error) {
	items := make([]model.Role, 0, len(rows))
	for _, row := range rows {
		item, err := row.toModel()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func mapRoleRepoError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return rolerepo.ErrNotFound
	case strings.Contains(strings.ToLower(err.Error()), "unique constraint failed"):
		return rolerepo.ErrConflict
	default:
		return err
	}
}

func nextRecordID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}

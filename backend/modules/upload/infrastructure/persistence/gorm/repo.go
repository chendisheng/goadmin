package gorm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"goadmin/modules/upload/domain/model"
	uploadrepo "goadmin/modules/upload/domain/repository"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type record struct {
	Id             string     `gorm:"column:id;primaryKey;type:varchar(64);size:64"`
	TenantId       string     `gorm:"column:tenant_id;type:varchar(64);size:64"`
	OriginalName   string     `gorm:"column:original_name;type:varchar(255);size:255"`
	StorageName    string     `gorm:"column:storage_name;type:varchar(255);size:255"`
	StorageKey     string     `gorm:"column:storage_key;type:varchar(255);size:255;index"`
	StorageDriver  string     `gorm:"column:storage_driver;type:varchar(64);size:64;index"`
	StoragePath    string     `gorm:"column:storage_path;type:varchar(255);size:255"`
	PublicURL      string     `gorm:"column:public_url;type:varchar(512);size:512"`
	MimeType       string     `gorm:"column:mime_type;type:varchar(128);size:128"`
	Extension      string     `gorm:"column:extension;type:varchar(32);size:32"`
	SizeBytes      int64      `gorm:"column:size_bytes"`
	ChecksumSHA256 string     `gorm:"column:checksum_sha256;type:varchar(128);size:128"`
	Visibility     string     `gorm:"column:visibility;type:varchar(32);size:32;index"`
	BizModule      string     `gorm:"column:biz_module;type:varchar(64);size:64;index"`
	BizType        string     `gorm:"column:biz_type;type:varchar(64);size:64;index"`
	BizId          string     `gorm:"column:biz_id;type:varchar(64);size:64;index"`
	BizField       string     `gorm:"column:biz_field;type:varchar(64);size:64"`
	UploadedBy     string     `gorm:"column:uploaded_by;type:varchar(64);size:64;index"`
	Status         string     `gorm:"column:status;type:varchar(32);size:32;index"`
	Remark         string     `gorm:"column:remark;type:varchar(255);size:255"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index"`
}

func (record) TableName() string { return "upload_file" }

const storageSettingKeyDefaultDriver = "default_storage_driver"

func New(db *gorm.DB) (*Repository, error) {
	if db == nil {
		return nil, fmt.Errorf("upload repository requires db")
	}
	return &Repository{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("upload migrate requires db")
	}
	return db.AutoMigrate(&record{}, &model.StorageSetting{})
}

func (r *Repository) List(ctx context.Context, filter uploadrepo.ListFilter) ([]model.FileAsset, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("upload repository is not configured")
	}
	base := r.db.WithContext(ctx).Model(&record{})
	if kw := strings.TrimSpace(strings.ToLower(filter.Keyword)); kw != "" {
		like := "%" + kw + "%"
		base = base.Where("LOWER(original_name) LIKE ? OR LOWER(storage_name) LIKE ? OR LOWER(storage_key) LIKE ? OR LOWER(public_url) LIKE ? OR LOWER(mime_type) LIKE ? OR LOWER(extension) LIKE ? OR LOWER(biz_module) LIKE ? OR LOWER(biz_type) LIKE ? OR LOWER(biz_id) LIKE ? OR LOWER(uploaded_by) LIKE ? OR LOWER(remark) LIKE ?", like, like, like, like, like, like, like, like, like, like, like)
	}
	if v := strings.TrimSpace(filter.Visibility); v != "" {
		base = base.Where("visibility = ?", v)
	}
	if s := strings.TrimSpace(filter.Status); s != "" {
		base = base.Where("status = ?", s)
	} else {
		base = base.Where("status <> ?", string(model.FileStatusDeleted))
	}
	if s := strings.TrimSpace(filter.BizModule); s != "" {
		base = base.Where("biz_module = ?", s)
	}
	if s := strings.TrimSpace(filter.BizType); s != "" {
		base = base.Where("biz_type = ?", s)
	}
	if s := strings.TrimSpace(filter.BizId); s != "" {
		base = base.Where("biz_id = ?", s)
	}
	if s := strings.TrimSpace(filter.UploadedBy); s != "" {
		base = base.Where("uploaded_by = ?", s)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	var rows []record
	if err := base.Order("updated_at DESC, created_at DESC, id ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]model.FileAsset, 0, len(rows))
	for _, row := range rows {
		item := row.toModel()
		items = append(items, item)
	}
	return items, total, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*model.FileAsset, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("upload repository is not configured")
	}
	var row record
	if err := r.db.WithContext(ctx).First(&row, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return nil, mapErr(err)
	}
	item := row.toModel()
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item *model.FileAsset) (*model.FileAsset, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("upload repository is not configured")
	}
	if item == nil {
		return nil, fmt.Errorf("upload file asset is nil")
	}
	rec := toRecord(*item)
	if strings.TrimSpace(rec.Id) == "" {
		rec.Id = nextRecordID("upload-file")
	}
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return nil, err
	}
	result := rec.toModel()
	return &result, nil
}

func (r *Repository) Update(ctx context.Context, item *model.FileAsset) (*model.FileAsset, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("upload repository is not configured")
	}
	if item == nil {
		return nil, fmt.Errorf("upload file asset is nil")
	}
	rec := toRecord(*item)
	if err := r.db.WithContext(ctx).Save(&rec).Error; err != nil {
		return nil, err
	}
	result := rec.toModel()
	return &result, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("upload repository is not configured")
	}
	var row record
	if err := r.db.WithContext(ctx).First(&row, "id = ?", strings.TrimSpace(id)).Error; err != nil {
		return mapErr(err)
	}
	now := time.Now().UTC()
	row.Status = string(model.FileStatusDeleted)
	row.DeletedAt = &now
	row.UpdatedAt = now
	return r.db.WithContext(ctx).Save(&row).Error
}

func (r *Repository) Bind(ctx context.Context, id string, binding model.FileBinding) (*model.FileAsset, error) {
	item, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	item.BizModule = strings.TrimSpace(binding.BizModule)
	item.BizType = strings.TrimSpace(binding.BizType)
	item.BizId = strings.TrimSpace(binding.BizId)
	item.BizField = strings.TrimSpace(binding.BizField)
	return r.Update(ctx, item)
}

func (r *Repository) Unbind(ctx context.Context, id string) (*model.FileAsset, error) {
	item, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	item.BizModule = ""
	item.BizType = ""
	item.BizId = ""
	item.BizField = ""
	return r.Update(ctx, item)
}

func (r *Repository) DefaultStorageDriver(ctx context.Context, fallback string) (string, error) {
	if r == nil || r.db == nil {
		return strings.TrimSpace(fallback), fmt.Errorf("upload repository is not configured")
	}
	var row model.StorageSetting
	if err := r.db.WithContext(ctx).First(&row, "setting_key = ?", storageSettingKeyDefaultDriver).Error; err != nil {
		if mapErr(err) == uploadrepo.ErrNotFound {
			return strings.TrimSpace(fallback), nil
		}
		return strings.TrimSpace(fallback), err
	}
	driver := strings.TrimSpace(row.SettingValue)
	if driver == "" {
		driver = strings.TrimSpace(fallback)
	}
	if driver == "" {
		driver = "local"
	}
	return driver, nil
}

func (r *Repository) SetDefaultStorageDriver(ctx context.Context, driver string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("upload repository is not configured")
	}
	driver = strings.TrimSpace(driver)
	if driver == "" {
		return fmt.Errorf("upload default storage driver is required")
	}
	row := model.StorageSetting{}
	if err := r.db.WithContext(ctx).First(&row, "setting_key = ?", storageSettingKeyDefaultDriver).Error; err != nil {
		if mapErr(err) != uploadrepo.ErrNotFound {
			return err
		}
		row = model.StorageSetting{SettingKey: storageSettingKeyDefaultDriver, SettingValue: driver}
		return r.db.WithContext(ctx).Create(&row).Error
	}
	row.SettingValue = driver
	return r.db.WithContext(ctx).Save(&row).Error
}

func toRecord(item model.FileAsset) record {
	return record{
		Id:             strings.TrimSpace(item.Id),
		TenantId:       strings.TrimSpace(item.TenantId),
		OriginalName:   strings.TrimSpace(item.OriginalName),
		StorageName:    strings.TrimSpace(item.StorageName),
		StorageKey:     strings.TrimSpace(item.StorageKey),
		StorageDriver:  strings.TrimSpace(item.StorageDriver),
		StoragePath:    strings.TrimSpace(item.StoragePath),
		PublicURL:      strings.TrimSpace(item.PublicURL),
		MimeType:       strings.TrimSpace(item.MimeType),
		Extension:      strings.TrimSpace(item.Extension),
		SizeBytes:      item.SizeBytes,
		ChecksumSHA256: strings.TrimSpace(item.ChecksumSHA256),
		Visibility:     string(item.Visibility),
		BizModule:      strings.TrimSpace(item.BizModule),
		BizType:        strings.TrimSpace(item.BizType),
		BizId:          strings.TrimSpace(item.BizId),
		BizField:       strings.TrimSpace(item.BizField),
		UploadedBy:     strings.TrimSpace(item.UploadedBy),
		Status:         string(item.Status),
		Remark:         strings.TrimSpace(item.Remark),
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
		DeletedAt:      item.DeletedAt,
	}
}

func (r record) toModel() model.FileAsset {
	return model.FileAsset{
		Id:             strings.TrimSpace(r.Id),
		TenantId:       strings.TrimSpace(r.TenantId),
		OriginalName:   strings.TrimSpace(r.OriginalName),
		StorageName:    strings.TrimSpace(r.StorageName),
		StorageKey:     strings.TrimSpace(r.StorageKey),
		StorageDriver:  strings.TrimSpace(r.StorageDriver),
		StoragePath:    strings.TrimSpace(r.StoragePath),
		PublicURL:      strings.TrimSpace(r.PublicURL),
		MimeType:       strings.TrimSpace(r.MimeType),
		Extension:      strings.TrimSpace(r.Extension),
		SizeBytes:      r.SizeBytes,
		ChecksumSHA256: strings.TrimSpace(r.ChecksumSHA256),
		Visibility:     model.FileVisibility(strings.TrimSpace(r.Visibility)),
		BizModule:      strings.TrimSpace(r.BizModule),
		BizType:        strings.TrimSpace(r.BizType),
		BizId:          strings.TrimSpace(r.BizId),
		BizField:       strings.TrimSpace(r.BizField),
		UploadedBy:     strings.TrimSpace(r.UploadedBy),
		Status:         model.FileStatus(strings.TrimSpace(r.Status)),
		Remark:         strings.TrimSpace(r.Remark),
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		DeletedAt:      r.DeletedAt,
	}
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "record not found") {
		return uploadrepo.ErrNotFound
	}
	return err
}

func nextRecordID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return page, pageSize
}

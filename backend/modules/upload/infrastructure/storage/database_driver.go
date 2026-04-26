package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	apperrors "goadmin/core/errors"
	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"

	"gorm.io/gorm"
)

type DatabaseDriver struct {
	db *gorm.DB
}

type databaseObjectRecord struct {
	StorageKey     string    `gorm:"column:storage_key;primaryKey;type:varchar(255);size:255"`
	StorageName    string    `gorm:"column:storage_name;type:varchar(255);size:255"`
	ContentType    string    `gorm:"column:content_type;type:varchar(128);size:128"`
	SizeBytes      int64     `gorm:"column:size_bytes"`
	ChecksumSHA256 string    `gorm:"column:checksum_sha256;type:varchar(128);size:128"`
	Visibility     string    `gorm:"column:visibility;type:varchar(32);size:32"`
	MetadataJSON   string    `gorm:"column:metadata_json;type:longtext"`
	PublicURL      string    `gorm:"column:public_url;type:varchar(512);size:512"`
	Data           []byte    `gorm:"column:data;type:longblob"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (databaseObjectRecord) TableName() string {
	return "upload_storage_blob"
}

func NewDatabaseDriver() *DatabaseDriver {
	return &DatabaseDriver{}
}

func (d *DatabaseDriver) SetDB(db *gorm.DB) error {
	if db == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "upload.storage.db_driver_required", "upload storage db driver requires db")
	}
	d.db = db
	return d.db.AutoMigrate(&databaseObjectRecord{})
}

func (d *DatabaseDriver) Name() string { return "db" }

func (d *DatabaseDriver) Put(ctx context.Context, req storagecontract.PutObjectRequest) (*storagecontract.PutObjectResult, error) {
	if d == nil || d.db == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.storage.db_driver_not_configured", "database storage driver is not configured")
	}
	if req.Reader == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.reader_required", "upload reader is required")
	}
	key, err := normalizeDatabaseStorageKey(req.Key)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(req.Reader)
	if err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.read_stream_failed", "read upload stream")
	}
	if req.Size > 0 && int64(len(data)) != req.Size {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.file_size_mismatch", fmt.Sprintf("upload file size mismatch: got %d want %d", len(data), req.Size))
	}
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])
	if expected := strings.TrimSpace(req.ChecksumSHA256); expected != "" && !strings.EqualFold(expected, checksum) {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.checksum_mismatch", fmt.Sprintf("upload checksum mismatch: got %s want %s", checksum, expected))
	}
	contentType := strings.TrimSpace(req.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(req.Filename))
	}
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(key))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	metadataJSON, err := encodeDatabaseMetadata(req.Metadata)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	rec := databaseObjectRecord{
		StorageKey:     key,
		StorageName:    firstNonEmptyDatabaseValue(filepath.Base(strings.TrimSpace(req.Filename)), filepath.Base(key)),
		ContentType:    contentType,
		SizeBytes:      int64(len(data)),
		ChecksumSHA256: checksum,
		Visibility:     strings.TrimSpace(req.Visibility),
		MetadataJSON:   metadataJSON,
		PublicURL:      strings.TrimSpace(req.Metadata["public_url"]),
		Data:           append([]byte(nil), data...),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := d.db.WithContext(ctx).Save(&rec).Error; err != nil {
		return nil, err
	}
	return &storagecontract.PutObjectResult{
		Key:            rec.StorageKey,
		StorageName:    rec.StorageName,
		URL:            strings.TrimSpace(rec.PublicURL),
		ETag:           rec.ChecksumSHA256,
		Size:           rec.SizeBytes,
		ChecksumSHA256: rec.ChecksumSHA256,
	}, nil
}

func (d *DatabaseDriver) Get(ctx context.Context, key string) (io.ReadCloser, *storagecontract.ObjectInfo, error) {
	if d == nil || d.db == nil {
		return nil, nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.storage.db_driver_not_configured", "database storage driver is not configured")
	}
	rec, err := d.loadRecord(ctx, key)
	if err != nil {
		return nil, nil, err
	}
	meta := decodeDatabaseMetadata(rec.MetadataJSON)
	info := &storagecontract.ObjectInfo{
		Key:         rec.StorageKey,
		Size:        rec.SizeBytes,
		ContentType: firstNonEmptyDatabaseValue(rec.ContentType, mime.TypeByExtension(filepath.Ext(rec.StorageKey)), "application/octet-stream"),
		ETag:        rec.ChecksumSHA256,
		ModTime:     rec.UpdatedAt,
		Metadata:    meta,
		PublicURL:   strings.TrimSpace(rec.PublicURL),
	}
	if info.Metadata == nil {
		info.Metadata = map[string]string{}
	}
	return io.NopCloser(bytes.NewReader(append([]byte(nil), rec.Data...))), info, nil
}

func (d *DatabaseDriver) Delete(ctx context.Context, key string) error {
	if d == nil || d.db == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "upload.storage.db_driver_not_configured", "database storage driver is not configured")
	}
	key, err := normalizeDatabaseStorageKey(key)
	if err != nil {
		return err
	}
	if err := d.db.WithContext(ctx).Delete(&databaseObjectRecord{}, "storage_key = ?", key).Error; err != nil {
		return err
	}
	return nil
}

func (d *DatabaseDriver) Exists(ctx context.Context, key string) (bool, error) {
	if d == nil || d.db == nil {
		return false, apperrors.NewWithKey(apperrors.CodeInternal, "upload.storage.db_driver_not_configured", "database storage driver is not configured")
	}
	_, err := d.loadRecord(ctx, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *DatabaseDriver) PublicURL(ctx context.Context, key string) (string, error) {
	if d == nil || d.db == nil {
		return "", apperrors.NewWithKey(apperrors.CodeInternal, "upload.storage.db_driver_not_configured", "database storage driver is not configured")
	}
	rec, err := d.loadRecord(ctx, key)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(rec.PublicURL), nil
}

func (d *DatabaseDriver) SignedURL(ctx context.Context, key string, opts storagecontract.SignedURLOptions) (string, error) {
	return d.PublicURL(ctx, key)
}

func (d *DatabaseDriver) loadRecord(ctx context.Context, key string) (*databaseObjectRecord, error) {
	key, err := normalizeDatabaseStorageKey(key)
	if err != nil {
		return nil, err
	}
	var rec databaseObjectRecord
	if err := d.db.WithContext(ctx).First(&rec, "storage_key = ?", key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapWithKey(err, apperrors.CodeNotFound, "upload.file_not_found", fmt.Sprintf("upload file asset %s not found", key))
		}
		return nil, err
	}
	return &rec, nil
}

func normalizeDatabaseStorageKey(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_key_required", "upload storage key is required")
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = filepath.Clean(trimmed)
	if trimmed == "." || trimmed == "" {
		return "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_key_invalid", "upload storage key is invalid")
	}
	if strings.HasPrefix(trimmed, "../") || trimmed == ".." || strings.Contains(trimmed, "/../") {
		return "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_key_traversal", "upload storage key contains path traversal")
	}
	return strings.TrimPrefix(trimmed, "/"), nil
}

func encodeDatabaseMetadata(metadata map[string]string) (string, error) {
	if len(metadata) == 0 {
		return "{}", nil
	}
	filtered := make(map[string]string, len(metadata))
	for key, value := range metadata {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		filtered[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if len(filtered) == 0 {
		return "{}", nil
	}
	data, err := json.Marshal(filtered)
	if err != nil {
		return "", apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.encode_metadata_failed", "encode upload metadata")
	}
	return string(data), nil
}

func decodeDatabaseMetadata(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}
	}
	meta := make(map[string]string)
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return map[string]string{}
	}
	if meta == nil {
		return map[string]string{}
	}
	return meta
}

func firstNonEmptyDatabaseValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

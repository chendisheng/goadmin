package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"goadmin/core/config"
	apperrors "goadmin/core/errors"
	"goadmin/modules/upload/domain/model"
	uploadrepo "goadmin/modules/upload/domain/repository"
	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
)

type Service struct {
	repo   uploadrepo.Repository
	driver storagecontract.Driver
	policy config.StoragePolicyConfig
}

type UploadRequest struct {
	File        io.Reader
	Filename    string
	ContentType string
	Size        int64
	TenantId    string
	UploadedBy  string
	Visibility  string
	BizModule   string
	BizType     string
	BizId       string
	BizField    string
	Remark      string
}

func New(repo uploadrepo.Repository, driver storagecontract.Driver, policy config.StoragePolicyConfig) (*Service, error) {
	if repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.repository_required", "upload repository is required")
	}
	if driver == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.storage_driver_required", "upload storage driver is required")
	}
	policy = normalizePolicy(policy)
	return &Service{repo: repo, driver: driver, policy: policy}, nil
}

func (s *Service) Upload(ctx context.Context, req UploadRequest) (*model.FileAsset, error) {
	if s == nil || s.repo == nil || s.driver == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.service_not_configured", "upload service is not configured")
	}
	if req.File == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.file_required", "upload file stream is required")
	}
	originalName, ext, err := normalizeFilename(req.Filename)
	if err != nil {
		return nil, err
	}
	if err := validateSize(req.Size, s.policy.MaxUploadSize); err != nil {
		return nil, err
	}
	contentType := strings.TrimSpace(req.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(ext)
	}
	if err := validateExt(ext, s.policy.AllowedExtensions); err != nil {
		return nil, err
	}
	if err := validateMIME(contentType, s.policy.AllowedMIMETypes); err != nil {
		return nil, err
	}
	visibility := strings.TrimSpace(req.Visibility)
	if visibility == "" {
		visibility = strings.TrimSpace(s.policy.VisibilityDefault)
	}
	if visibility == "" {
		visibility = string(model.FileVisibilityPrivate)
	}
	if visibility != string(model.FileVisibilityPrivate) && visibility != string(model.FileVisibilityPublic) {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.visibility_invalid", fmt.Sprintf("invalid upload visibility %q", visibility))
	}

	key := buildStorageKey(s.policy.PathPrefix, ext)
	putResult, err := s.driver.Put(ctx, storagecontract.PutObjectRequest{
		Key:         key,
		Reader:      req.File,
		Size:        req.Size,
		ContentType: contentType,
		Filename:    originalName,
		Metadata: map[string]string{
			"original_name": originalName,
			"uploaded_by":   strings.TrimSpace(req.UploadedBy),
			"biz_module":    strings.TrimSpace(req.BizModule),
			"biz_type":      strings.TrimSpace(req.BizType),
			"biz_id":        strings.TrimSpace(req.BizId),
			"biz_field":     strings.TrimSpace(req.BizField),
		},
		Visibility:     visibility,
		ChecksumSHA256: "",
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	asset := &model.FileAsset{
		TenantId:       strings.TrimSpace(req.TenantId),
		OriginalName:   originalName,
		StorageName:    putResult.StorageName,
		StorageKey:     putResult.Key,
		StorageDriver:  s.driver.Name(),
		StoragePath:    putResult.Key,
		MimeType:       contentType,
		Extension:      ext,
		SizeBytes:      putResult.Size,
		ChecksumSHA256: putResult.ChecksumSHA256,
		Visibility:     model.FileVisibility(visibility),
		BizModule:      strings.TrimSpace(req.BizModule),
		BizType:        strings.TrimSpace(req.BizType),
		BizId:          strings.TrimSpace(req.BizId),
		BizField:       strings.TrimSpace(req.BizField),
		UploadedBy:     strings.TrimSpace(req.UploadedBy),
		Status:         model.FileStatusActive,
		Remark:         strings.TrimSpace(req.Remark),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if visibility == string(model.FileVisibilityPublic) {
		publicURL := strings.TrimSpace(putResult.URL)
		if publicURL == "" {
			publicURL, err = s.driver.PublicURL(ctx, putResult.Key)
			if err != nil {
				_ = s.driver.Delete(ctx, putResult.Key)
				return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.public_url_failed", "resolve upload public url")
			}
		}
		asset.PublicURL = publicURL
	}
	created, err := s.repo.Create(ctx, asset)
	if err != nil {
		_ = s.driver.Delete(ctx, putResult.Key)
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.persist_failed", "store upload metadata")
	}
	return created, nil
}

func (s *Service) Get(ctx context.Context, id string) (*model.FileAsset, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.service_not_configured", "upload service is not configured")
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Open(ctx context.Context, id string) (io.ReadCloser, *storagecontract.ObjectInfo, *model.FileAsset, error) {
	if s == nil || s.repo == nil || s.driver == nil {
		return nil, nil, nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.service_not_configured", "upload service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	if item == nil {
		return nil, nil, nil, apperrors.NewWithKey(apperrors.CodeNotFound, "upload.file_not_found", fmt.Sprintf("upload file asset %s not found", strings.TrimSpace(id)))
	}
	key := strings.TrimSpace(item.StorageKey)
	if key == "" {
		key = strings.TrimSpace(item.StoragePath)
	}
	if key == "" {
		return nil, nil, nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.storage_key_required", fmt.Sprintf("upload storage key is empty for %s", item.Id))
	}
	reader, info, err := s.driver.Get(ctx, key)
	if err != nil {
		return nil, nil, nil, err
	}
	return reader, info, item, nil
}

func (s *Service) List(ctx context.Context, filter uploadrepo.ListFilter) ([]model.FileAsset, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, apperrors.NewWithKey(apperrors.CodeInternal, "upload.service_not_configured", "upload service is not configured")
	}
	return s.repo.List(ctx, filter)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil || s.driver == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "upload.service_not_configured", "upload service is not configured")
	}
	asset, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.driver.Delete(ctx, asset.StorageKey); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) Bind(ctx context.Context, id string, binding model.FileBinding) (*model.FileAsset, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.service_not_configured", "upload service is not configured")
	}
	return s.repo.Bind(ctx, id, binding)
}

func (s *Service) Unbind(ctx context.Context, id string) (*model.FileAsset, error) {
	if s == nil || s.repo == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.service_not_configured", "upload service is not configured")
	}
	return s.repo.Unbind(ctx, id)
}

func (s *Service) DefaultStorageDriver(ctx context.Context) (string, error) {
	if s == nil || s.repo == nil || s.driver == nil {
		return "", apperrors.NewWithKey(apperrors.CodeInternal, "upload.service_not_configured", "upload service is not configured")
	}
	return s.repo.DefaultStorageDriver(ctx, s.driver.Name())
}

func (s *Service) SetDefaultStorageDriver(ctx context.Context, driver string) error {
	if s == nil || s.repo == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "upload.service_not_configured", "upload service is not configured")
	}
	driver = strings.ToLower(strings.TrimSpace(driver))
	if driver == "" {
		return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.default_storage_driver_required", "default storage driver is required")
	}
	switch driver {
	case "local", "db", "database", "s3-compatible", "oss", "cos", "qiniu", "minio":
		return s.repo.SetDefaultStorageDriver(ctx, driver)
	default:
		return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.unsupported_storage_driver", fmt.Sprintf("unsupported upload storage driver %q", driver))
	}
}

func normalizePolicy(policy config.StoragePolicyConfig) config.StoragePolicyConfig {
	if strings.TrimSpace(policy.MaxUploadSize) == "" {
		policy.MaxUploadSize = "20mb"
	}
	if len(policy.AllowedExtensions) == 0 {
		policy.AllowedExtensions = []string{".png", ".jpg", ".jpeg", ".webp", ".pdf", ".txt"}
	}
	if len(policy.AllowedMIMETypes) == 0 {
		policy.AllowedMIMETypes = []string{"image/png", "image/jpeg", "image/webp", "application/pdf", "text/plain"}
	}
	if strings.TrimSpace(policy.VisibilityDefault) == "" {
		policy.VisibilityDefault = string(model.FileVisibilityPrivate)
	}
	if strings.TrimSpace(policy.PathPrefix) == "" {
		policy.PathPrefix = "uploads"
	}
	return policy
}

func normalizeFilename(name string) (string, string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.filename_required", "upload filename is required")
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	if strings.Contains(trimmed, "/") {
		trimmed = filepath.Base(trimmed)
	}
	if trimmed == "." || trimmed == ".." || trimmed == "" {
		return "", "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.filename_invalid", "upload filename is invalid")
	}
	ext := strings.ToLower(filepath.Ext(trimmed))
	return trimmed, ext, nil
}

func validateSize(size int64, maxSize string) error {
	if strings.TrimSpace(maxSize) == "" || size <= 0 {
		return nil
	}
	limit, err := parseByteSize(maxSize)
	if err != nil {
		return err
	}
	if size > limit {
		return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.file_size_exceeded", fmt.Sprintf("upload file size %d exceeds limit %d", size, limit))
	}
	return nil
}

func validateExt(ext string, allowed []string) error {
	if len(allowed) == 0 || strings.TrimSpace(ext) == "" {
		return nil
	}
	ext = strings.ToLower(strings.TrimSpace(ext))
	for _, candidate := range allowed {
		if strings.ToLower(strings.TrimSpace(candidate)) == ext {
			return nil
		}
	}
	return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.extension_not_allowed", fmt.Sprintf("upload extension %q is not allowed", ext))
}

func validateMIME(contentType string, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.content_type_required", "upload content type is required")
	}
	for _, candidate := range allowed {
		if strings.EqualFold(strings.TrimSpace(candidate), contentType) {
			return nil
		}
	}
	return apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.content_type_not_allowed", fmt.Sprintf("upload content type %q is not allowed", contentType))
}

func buildStorageKey(prefix, ext string) string {
	prefix = strings.Trim(prefix, "/")
	datePart := time.Now().UTC().Format("2006/01/02")
	name := randomToken(16)
	if ext != "" {
		name += ext
	}
	parts := []string{prefix, datePart, name}
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, "/")
		if part != "" {
			clean = append(clean, part)
		}
	}
	return strings.Join(clean, "/")
}

func randomToken(n int) string {
	if n <= 0 {
		n = 16
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	}
	return hex.EncodeToString(buf)
}

func parseByteSize(raw string) (int64, error) {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return 0, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.byte_size_required", "empty byte size")
	}
	type unit struct {
		suffix string
		mul    int64
	}
	units := []unit{{"tb", 1024 * 1024 * 1024 * 1024}, {"gb", 1024 * 1024 * 1024}, {"mb", 1024 * 1024}, {"kb", 1024}, {"b", 1}}
	for _, u := range units {
		if strings.HasSuffix(trimmed, u.suffix) {
			value := strings.TrimSpace(strings.TrimSuffix(trimmed, u.suffix))
			if value == "" {
				return 0, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.byte_size_invalid", fmt.Sprintf("invalid byte size %q", raw))
			}
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return 0, apperrors.WrapWithKey(err, apperrors.CodeBadRequest, "upload.byte_size_invalid", fmt.Sprintf("invalid byte size %q", raw))
			}
			return int64(num * float64(u.mul)), nil
		}
	}
	num, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, apperrors.WrapWithKey(err, apperrors.CodeBadRequest, "upload.byte_size_invalid", fmt.Sprintf("invalid byte size %q", raw))
	}
	return int64(num), nil
}

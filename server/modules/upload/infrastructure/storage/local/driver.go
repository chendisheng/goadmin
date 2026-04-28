package local

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"goadmin/core/config"
	apperrors "goadmin/core/errors"
	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
)

type Driver struct {
	cfg              config.LocalStorageConfig
	baseDir          string
	publicBaseURL    string
	useProxyDownload bool
}

func NewDriver(cfg config.LocalStorageConfig) (*Driver, error) {
	baseDir := strings.TrimSpace(cfg.BaseDir)
	if baseDir == "" {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.local_base_dir_required", "upload storage local base_dir is required")
	}
	publicBaseURL := strings.TrimSpace(cfg.PublicBaseURL)
	if publicBaseURL == "" {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.local_public_base_url_required", "upload storage local public_base_url is required")
	}
	baseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.local_resolve_base_dir_failed", "resolve upload storage local base_dir")
	}
	return &Driver{
		cfg:              cfg,
		baseDir:          filepath.Clean(baseDir),
		publicBaseURL:    strings.TrimRight(publicBaseURL, "/"),
		useProxyDownload: cfg.UseProxyDownload,
	}, nil
}

func (d *Driver) Name() string { return "local" }

func (d *Driver) Put(ctx context.Context, req storagecontract.PutObjectRequest) (*storagecontract.PutObjectResult, error) {
	if d == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.local_driver_not_configured", "local storage driver is not configured")
	}
	if req.Reader == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.reader_required", "upload reader is required")
	}
	key, err := d.normalizeKey(req.Key)
	if err != nil {
		return nil, err
	}
	if err := d.ensureBaseDir(); err != nil {
		return nil, err
	}
	absPath, err := d.resolvePath(key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.local_create_dir_failed", "create upload directory")
	}

	tmp, err := os.CreateTemp(filepath.Dir(absPath), ".upload-*")
	if err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.local_create_temp_failed", "create upload temp file")
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(tmp, hasher), req.Reader)
	if err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.local_write_file_failed", "write upload file")
	}
	if req.Size > 0 && written != req.Size {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.file_size_mismatch", fmt.Sprintf("upload file size mismatch: got %d want %d", written, req.Size))
	}
	if err := tmp.Sync(); err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.local_sync_file_failed", "sync upload file")
	}
	if err := tmp.Close(); err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.local_close_temp_failed", "close upload temp file")
	}
	if err := os.Rename(tmpPath, absPath); err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.local_move_file_failed", "move upload file into place")
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	contentType := strings.TrimSpace(req.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(key))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return &storagecontract.PutObjectResult{
		Key:            key,
		StorageName:    filepath.Base(key),
		URL:            d.buildPublicURL(key),
		ETag:           checksum,
		Size:           written,
		ChecksumSHA256: checksum,
	}, nil
}

func (d *Driver) Get(ctx context.Context, key string) (io.ReadCloser, *storagecontract.ObjectInfo, error) {
	if d == nil {
		return nil, nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.local_driver_not_configured", "local storage driver is not configured")
	}
	key, err := d.normalizeKey(key)
	if err != nil {
		return nil, nil, err
	}
	absPath, err := d.resolvePath(key)
	if err != nil {
		return nil, nil, err
	}
	file, err := os.Open(absPath)
	if err != nil {
		return nil, nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}
	info := &storagecontract.ObjectInfo{
		Key:         key,
		Size:        stat.Size(),
		ContentType: mime.TypeByExtension(filepath.Ext(key)),
		ModTime:     stat.ModTime(),
		PublicURL:   d.buildPublicURL(key),
	}
	if info.ContentType == "" {
		info.ContentType = "application/octet-stream"
	}
	return file, info, nil
}

func (d *Driver) Delete(ctx context.Context, key string) error {
	if d == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "upload.local_driver_not_configured", "local storage driver is not configured")
	}
	key, err := d.normalizeKey(key)
	if err != nil {
		return err
	}
	absPath, err := d.resolvePath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (d *Driver) Exists(ctx context.Context, key string) (bool, error) {
	if d == nil {
		return false, apperrors.NewWithKey(apperrors.CodeInternal, "upload.local_driver_not_configured", "local storage driver is not configured")
	}
	key, err := d.normalizeKey(key)
	if err != nil {
		return false, err
	}
	absPath, err := d.resolvePath(key)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(absPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (d *Driver) PublicURL(ctx context.Context, key string) (string, error) {
	if d == nil {
		return "", apperrors.NewWithKey(apperrors.CodeInternal, "upload.local_driver_not_configured", "local storage driver is not configured")
	}
	key, err := d.normalizeKey(key)
	if err != nil {
		return "", err
	}
	return d.buildPublicURL(key), nil
}

func (d *Driver) SignedURL(ctx context.Context, key string, opts storagecontract.SignedURLOptions) (string, error) {
	return d.PublicURL(ctx, key)
}

func (d *Driver) normalizeKey(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_key_required", "upload storage key is required")
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = pathClean(trimmed)
	if trimmed == "." || trimmed == "" {
		return "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_key_invalid", "upload storage key is invalid")
	}
	if strings.HasPrefix(trimmed, "../") || trimmed == ".." || strings.Contains(trimmed, "/../") {
		return "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.storage_key_traversal", "upload storage key contains path traversal")
	}
	return strings.TrimPrefix(trimmed, "/"), nil
}

func (d *Driver) resolvePath(key string) (string, error) {
	cleaned, err := d.normalizeKey(key)
	if err != nil {
		return "", err
	}
	absPath := filepath.Clean(filepath.Join(d.baseDir, filepath.FromSlash(cleaned)))
	if absPath != d.baseDir && !strings.HasPrefix(absPath, d.baseDir+string(os.PathSeparator)) {
		return "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.local_key_escapes_base_dir", "upload storage key escapes base dir")
	}
	return absPath, nil
}

func (d *Driver) ensureBaseDir() error {
	if err := os.MkdirAll(d.baseDir, 0o755); err != nil {
		return apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.local_ensure_base_dir_failed", "ensure upload storage base dir")
	}
	return nil
}

func (d *Driver) buildPublicURL(key string) string {
	key = strings.TrimPrefix(strings.ReplaceAll(key, "\\", "/"), "/")
	if key == "" {
		return d.publicBaseURL
	}
	return strings.TrimRight(d.publicBaseURL, "/") + "/" + key
}

func pathClean(value string) string {
	cleaned := filepath.Clean(strings.ReplaceAll(value, "/", string(os.PathSeparator)))
	return strings.ReplaceAll(cleaned, string(os.PathSeparator), "/")
}

package adapter

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
	"time"

	apperrors "goadmin/core/errors"
	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
)

type Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	AccessKeySecret string
	UseSSL          bool
	PathStyle       bool
	PublicBaseURL   string
}

type minioServer interface {
	Name() string
	PutObject(ctx context.Context, req storagecontract.PutObjectRequest) (*storagecontract.PutObjectResult, error)
	GetObject(ctx context.Context, key string) (io.ReadCloser, *storagecontract.ObjectInfo, error)
	HeadObject(ctx context.Context, key string) (*storagecontract.ObjectInfo, error)
	DeleteObject(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	PublicURL(ctx context.Context, key string) (string, error)
	SignedURL(ctx context.Context, key string, opts storagecontract.SignedURLOptions) (string, error)
}

type Client struct {
	name          string
	baseDir       string
	publicBaseURL string
	minio         minioServer
}

func New(name string, cfg Config) (*Client, error) {
	if strings.TrimSpace(name) == "" {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.object_storage_driver_name_required", "object storage driver name is required")
	}
	if strings.TrimSpace(name) == "minio" {
		server, enabled, err := newMinIOServer(cfg)
		if err != nil {
			return nil, err
		}
		if enabled {
			return &Client{name: name, minio: server}, nil
		}
	}
	if err := validateConfig(name, cfg); err != nil {
		return nil, err
	}
	baseDir := filepath.Join(os.TempDir(), "goadmin", "uploads", "object", sanitizeSegment(name), stableConfigKey(cfg))
	return &Client{
		name:          name,
		baseDir:       filepath.Clean(baseDir),
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

func (c *Client) Name() string {
	if c == nil {
		return ""
	}
	return c.name
}

func (c *Client) PutObject(ctx context.Context, req storagecontract.PutObjectRequest) (*storagecontract.PutObjectResult, error) {
	if c == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.object_storage_client_not_configured", "object storage client is not configured")
	}
	if c.minio != nil {
		return c.minio.PutObject(ctx, req)
	}
	if req.Reader == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.reader_required", "upload reader is required")
	}
	key, err := normalizeKey(req.Key)
	if err != nil {
		return nil, err
	}
	if err := c.ensureBaseDir(); err != nil {
		return nil, err
	}
	absPath, metaPath, err := c.resolvePaths(key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.object_create_dir_failed", "create object directory")
	}

	tmp, err := os.CreateTemp(filepath.Dir(absPath), ".upload-*")
	if err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.object_create_temp_failed", "create object temp file")
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(tmp, hasher), req.Reader)
	if err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.object_write_file_failed", "write object file")
	}
	if req.Size > 0 && written != req.Size {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.file_size_mismatch", fmt.Sprintf("upload file size mismatch: got %d want %d", written, req.Size))
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))
	if expected := strings.TrimSpace(req.ChecksumSHA256); expected != "" && !strings.EqualFold(expected, checksum) {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.checksum_mismatch", fmt.Sprintf("upload checksum mismatch: got %s want %s", checksum, expected))
	}
	if err := tmp.Sync(); err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.object_sync_file_failed", "sync object file")
	}
	if err := tmp.Close(); err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.object_close_temp_failed", "close object temp file")
	}
	if err := os.Rename(tmpPath, absPath); err != nil {
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.object_move_file_failed", "move object file into place")
	}

	contentType := strings.TrimSpace(req.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(key))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	info := storedObjectMeta{
		Key:                key,
		StorageName:        filepath.Base(key),
		Size:               written,
		ContentType:        contentType,
		ETag:               checksum,
		ChecksumSHA256:     checksum,
		Metadata:           cloneStringMap(req.Metadata),
		Visibility:         strings.TrimSpace(req.Visibility),
		Filename:           strings.TrimSpace(req.Filename),
		ContentDisposition: strings.TrimSpace(req.ContentDisposition),
		PublicURL:          c.buildPublicURL(key),
		StoredAt:           time.Now().UTC(),
	}
	if err := writeJSON(metaPath, info); err != nil {
		_ = os.Remove(absPath)
		return nil, apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.object_write_metadata_failed", "write object metadata")
	}
	return &storagecontract.PutObjectResult{
		Key:            key,
		StorageName:    info.StorageName,
		URL:            info.PublicURL,
		ETag:           checksum,
		Size:           written,
		ChecksumSHA256: checksum,
	}, nil
}

func (c *Client) GetObject(ctx context.Context, key string) (io.ReadCloser, *storagecontract.ObjectInfo, error) {
	if c == nil {
		return nil, nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.object_storage_client_not_configured", "object storage client is not configured")
	}
	if c.minio != nil {
		return c.minio.GetObject(ctx, key)
	}
	absPath, info, err := c.inspectObject(key)
	if err != nil {
		return nil, nil, err
	}
	file, err := os.Open(absPath)
	if err != nil {
		return nil, nil, err
	}
	return file, info, nil
}

func (c *Client) HeadObject(ctx context.Context, key string) (*storagecontract.ObjectInfo, error) {
	if c == nil {
		return nil, apperrors.NewWithKey(apperrors.CodeInternal, "upload.object_storage_client_not_configured", "object storage client is not configured")
	}
	if c.minio != nil {
		return c.minio.HeadObject(ctx, key)
	}
	_, info, err := c.inspectObject(key)
	return info, err
}

func (c *Client) DeleteObject(ctx context.Context, key string) error {
	if c == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "upload.object_storage_client_not_configured", "object storage client is not configured")
	}
	if c.minio != nil {
		return c.minio.DeleteObject(ctx, key)
	}
	key, err := normalizeKey(key)
	if err != nil {
		return err
	}
	absPath, metaPath, err := c.resolvePaths(key)
	if err != nil {
		return err
	}
	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	if c == nil {
		return false, apperrors.NewWithKey(apperrors.CodeInternal, "upload.object_storage_client_not_configured", "object storage client is not configured")
	}
	if c.minio != nil {
		return c.minio.Exists(ctx, key)
	}
	key, err := normalizeKey(key)
	if err != nil {
		return false, err
	}
	absPath, _, err := c.resolvePaths(key)
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

func (c *Client) PublicURL(ctx context.Context, key string) (string, error) {
	if c == nil {
		return "", apperrors.NewWithKey(apperrors.CodeInternal, "upload.object_storage_client_not_configured", "object storage client is not configured")
	}
	if c.minio != nil {
		return c.minio.PublicURL(ctx, key)
	}
	if strings.TrimSpace(c.publicBaseURL) == "" {
		return "", apperrors.NewWithKey(apperrors.CodeInternal, "upload.object_storage_public_base_url_required", "object storage public_base_url is required for public urls")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return "", err
	}
	return c.buildPublicURL(key), nil
}

func (c *Client) SignedURL(ctx context.Context, key string, opts storagecontract.SignedURLOptions) (string, error) {
	if c == nil {
		return "", apperrors.NewWithKey(apperrors.CodeInternal, "upload.object_storage_client_not_configured", "object storage client is not configured")
	}
	if c.minio != nil {
		return c.minio.SignedURL(ctx, key, opts)
	}
	if strings.TrimSpace(c.publicBaseURL) == "" {
		return "", apperrors.NewWithKey(apperrors.CodeInternal, "upload.object_storage_public_base_url_required", "object storage public_base_url is required for signed urls")
	}
	base, err := c.PublicURL(ctx, key)
	if err != nil {
		return "", err
	}
	return buildSignedURL(base, opts), nil
}

func (c *Client) inspectObject(key string) (string, *storagecontract.ObjectInfo, error) {
	key, err := normalizeKey(key)
	if err != nil {
		return "", nil, err
	}
	absPath, metaPath, err := c.resolvePaths(key)
	if err != nil {
		return "", nil, err
	}
	stat, err := os.Stat(absPath)
	if err != nil {
		return "", nil, err
	}
	meta, _ := readMeta(metaPath)
	return absPath, objectInfoFromMeta(key, stat, meta, c.buildPublicURL(key)), nil
}

func (c *Client) ensureBaseDir() error {
	if err := os.MkdirAll(c.baseDir, 0o755); err != nil {
		return apperrors.WrapWithKey(err, apperrors.CodeInternal, "upload.object_storage_ensure_base_dir_failed", "ensure object storage base dir")
	}
	return nil
}

func (c *Client) resolvePaths(key string) (string, string, error) {
	cleaned, err := normalizeKey(key)
	if err != nil {
		return "", "", err
	}
	absPath := filepath.Clean(filepath.Join(c.baseDir, filepath.FromSlash(cleaned)))
	if absPath != c.baseDir && !strings.HasPrefix(absPath, c.baseDir+string(os.PathSeparator)) {
		return "", "", apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.object_storage_key_escapes_base_dir", "object storage key escapes base dir")
	}
	return absPath, absPath + ".meta.json", nil
}

func (c *Client) buildPublicURL(key string) string {
	if strings.TrimSpace(c.publicBaseURL) == "" {
		return ""
	}
	key = strings.TrimPrefix(strings.ReplaceAll(key, "\\", "/"), "/")
	if key == "" {
		return c.publicBaseURL
	}
	return strings.TrimRight(c.publicBaseURL, "/") + "/" + key
}

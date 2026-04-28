package objectstore

import (
	"context"
	"io"
	"strings"

	"goadmin/core/config"
	apperrors "goadmin/core/errors"
	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
)

type Driver struct {
	name   string
	cfg    storageConfig
	client objectStoreClient
}

type storageConfig struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	AccessKeySecret string
	UseSSL          bool
	PathStyle       bool
	PublicBaseURL   string
}

func NewS3CompatibleDriver(cfg config.S3CompatibleConfig) (*Driver, error) {
	return newDriver("s3-compatible", storageConfig{
		Endpoint:        cfg.Endpoint,
		Region:          cfg.Region,
		Bucket:          cfg.Bucket,
		AccessKeyID:     cfg.AccessKeyID,
		AccessKeySecret: cfg.AccessKeySecret,
		UseSSL:          cfg.UseSSL,
		PathStyle:       cfg.PathStyle,
		PublicBaseURL:   cfg.PublicBaseURL,
	})
}

func NewOSSDriver(cfg config.OSSStorageConfig) (*Driver, error) {
	return newDriver("oss", storageConfig{
		Endpoint:        cfg.Endpoint,
		Bucket:          cfg.Bucket,
		AccessKeyID:     cfg.AccessKeyID,
		AccessKeySecret: cfg.AccessKeySecret,
		PublicBaseURL:   cfg.PublicBaseURL,
	})
}

func NewCOSDriver(cfg config.COSStorageConfig) (*Driver, error) {
	return newDriver("cos", storageConfig{
		Region:          cfg.Region,
		Bucket:          cfg.Bucket,
		AccessKeyID:     cfg.SecretID,
		AccessKeySecret: cfg.SecretKey,
		PublicBaseURL:   cfg.PublicBaseURL,
	})
}

func NewMinIODriver(cfg config.MinIOStorageConfig) (*Driver, error) {
	return newDriver("minio", storageConfig{
		Endpoint:        cfg.Endpoint,
		Bucket:          cfg.Bucket,
		AccessKeyID:     cfg.AccessKeyID,
		AccessKeySecret: cfg.AccessKeySecret,
		UseSSL:          cfg.UseSSL,
		PathStyle:       cfg.PathStyle,
		PublicBaseURL:   cfg.PublicBaseURL,
	})
}

func newDriver(name string, cfg storageConfig) (*Driver, error) {
	if strings.TrimSpace(name) == "" {
		return nil, apperrors.NewWithKey(apperrors.CodeBadRequest, "upload.object_storage_driver_name_required", "object storage driver name is required")
	}
	client, err := newFileClient(name, cfg, cfg.PublicBaseURL)
	if err != nil {
		return nil, err
	}
	return &Driver{
		name:   name,
		cfg:    cfg,
		client: client,
	}, nil
}

func (d *Driver) Name() string {
	if d == nil {
		return ""
	}
	return d.name
}

func (d *Driver) Put(ctx context.Context, req storagecontract.PutObjectRequest) (*storagecontract.PutObjectResult, error) {
	return d.client.PutObject(ctx, req)
}

func (d *Driver) Get(ctx context.Context, key string) (io.ReadCloser, *storagecontract.ObjectInfo, error) {
	return d.client.GetObject(ctx, key)
}

func (d *Driver) Delete(ctx context.Context, key string) error {
	return d.client.DeleteObject(ctx, key)
}

func (d *Driver) Exists(ctx context.Context, key string) (bool, error) {
	return d.client.Exists(ctx, key)
}

func (d *Driver) PublicURL(ctx context.Context, key string) (string, error) {
	return d.client.PublicURL(ctx, key)
}

func (d *Driver) SignedURL(ctx context.Context, key string, opts storagecontract.SignedURLOptions) (string, error) {
	return d.client.SignedURL(ctx, key, opts)
}

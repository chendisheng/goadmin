package objectstore

import (
	"context"
	"io"

	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
	adapter "goadmin/modules/upload/infrastructure/storage/objectstore/adapter"
)

type objectStoreClient interface {
	Name() string
	PutObject(ctx context.Context, req storagecontract.PutObjectRequest) (*storagecontract.PutObjectResult, error)
	GetObject(ctx context.Context, key string) (io.ReadCloser, *storagecontract.ObjectInfo, error)
	HeadObject(ctx context.Context, key string) (*storagecontract.ObjectInfo, error)
	DeleteObject(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	PublicURL(ctx context.Context, key string) (string, error)
	SignedURL(ctx context.Context, key string, opts storagecontract.SignedURLOptions) (string, error)
}

type fileClient struct {
	adapter *adapter.Client
}

func newFileClient(name string, cfg storageConfig, publicBaseURL string) (*fileClient, error) {
	client, err := adapter.New(name, adapter.Config{
		Endpoint:        cfg.Endpoint,
		Region:          cfg.Region,
		Bucket:          cfg.Bucket,
		AccessKeyID:     cfg.AccessKeyID,
		AccessKeySecret: cfg.AccessKeySecret,
		UseSSL:          cfg.UseSSL,
		PathStyle:       cfg.PathStyle,
		PublicBaseURL:   publicBaseURL,
	})
	if err != nil {
		return nil, err
	}
	return &fileClient{adapter: client}, nil
}

func (c *fileClient) Name() string {
	if c == nil {
		return ""
	}
	return c.adapter.Name()
}

func (c *fileClient) PutObject(ctx context.Context, req storagecontract.PutObjectRequest) (*storagecontract.PutObjectResult, error) {
	return c.adapter.PutObject(ctx, req)
}

func (c *fileClient) GetObject(ctx context.Context, key string) (io.ReadCloser, *storagecontract.ObjectInfo, error) {
	return c.adapter.GetObject(ctx, key)
}

func (c *fileClient) HeadObject(ctx context.Context, key string) (*storagecontract.ObjectInfo, error) {
	return c.adapter.HeadObject(ctx, key)
}

func (c *fileClient) DeleteObject(ctx context.Context, key string) error {
	return c.adapter.DeleteObject(ctx, key)
}

func (c *fileClient) Exists(ctx context.Context, key string) (bool, error) {
	return c.adapter.Exists(ctx, key)
}

func (c *fileClient) PublicURL(ctx context.Context, key string) (string, error) {
	return c.adapter.PublicURL(ctx, key)
}

func (c *fileClient) SignedURL(ctx context.Context, key string, opts storagecontract.SignedURLOptions) (string, error) {
	return c.adapter.SignedURL(ctx, key, opts)
}

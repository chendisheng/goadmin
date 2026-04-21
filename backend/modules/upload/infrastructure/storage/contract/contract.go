package contract

import (
	"context"
	"io"
	"time"

	"gorm.io/gorm"
)

type Driver interface {
	Name() string
	Put(ctx context.Context, req PutObjectRequest) (*PutObjectResult, error)
	Get(ctx context.Context, key string) (io.ReadCloser, *ObjectInfo, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	PublicURL(ctx context.Context, key string) (string, error)
	SignedURL(ctx context.Context, key string, opts SignedURLOptions) (string, error)
}

type DatabaseAwareDriver interface {
	Driver
	SetDB(db *gorm.DB) error
}

type PutObjectRequest struct {
	Key                string
	Reader             io.Reader
	Size               int64
	ContentType        string
	Filename           string
	Metadata           map[string]string
	Visibility         string
	ChecksumSHA256     string
	ContentDisposition string
}

type PutObjectResult struct {
	Key            string
	StorageName    string
	URL            string
	ETag           string
	Size           int64
	ChecksumSHA256 string
}

type ObjectInfo struct {
	Key         string
	Size        int64
	ContentType string
	ETag        string
	ModTime     time.Time
	Metadata    map[string]string
	PublicURL   string
}

type SignedURLOptions struct {
	Method                     string
	Expires                    time.Duration
	ResponseContentType        string
	ResponseContentDisposition string
}

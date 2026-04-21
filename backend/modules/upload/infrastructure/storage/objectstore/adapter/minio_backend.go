package adapter

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioObjectBackend struct {
	client        *minio.Client
	bucket        string
	region        string
	endpoint      string
	publicBaseURL string
	mu            sync.RWMutex
	metadataCache map[string]map[string]string
}

func newMinIOBackend(cfg Config) (minioBackend, bool, error) {
	endpoint, secure, err := normalizeMinIOEndpoint(cfg.Endpoint, cfg.UseSSL)
	if err != nil {
		return nil, false, err
	}
	if strings.TrimSpace(endpoint) == "" || strings.TrimSpace(cfg.Bucket) == "" {
		return nil, false, nil
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(strings.TrimSpace(cfg.AccessKeyID), strings.TrimSpace(cfg.AccessKeySecret), ""),
		Secure:       secure,
		Region:       strings.TrimSpace(cfg.Region),
		BucketLookup: bucketLookupMode(cfg.PathStyle),
	})
	if err != nil {
		return nil, false, err
	}
	backend := &minioObjectBackend{
		client:        client,
		bucket:        strings.TrimSpace(cfg.Bucket),
		region:        strings.TrimSpace(cfg.Region),
		endpoint:      strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/"),
		publicBaseURL: resolveMinIOPublicBaseURL(cfg),
		metadataCache: make(map[string]map[string]string),
	}
	if err := backend.ensureBucket(context.Background()); err != nil {
		return nil, false, err
	}
	return backend, true, nil
}

func normalizeMinIOEndpoint(raw string, forceSSL bool) (string, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", forceSSL, nil
	}
	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", false, fmt.Errorf("parse upload.storage.minio.endpoint: %w", err)
		}
		if parsed.Host == "" {
			return "", false, fmt.Errorf("upload.storage.minio.endpoint must include a host")
		}
		switch strings.ToLower(parsed.Scheme) {
		case "https":
			return parsed.Host, true, nil
		case "http":
			return parsed.Host, false, nil
		default:
			return parsed.Host, forceSSL, nil
		}
	}
	return raw, forceSSL, nil
}

func bucketLookupMode(pathStyle bool) minio.BucketLookupType {
	if pathStyle {
		return minio.BucketLookupPath
	}
	return minio.BucketLookupAuto
}

func resolveMinIOPublicBaseURL(cfg Config) string {
	if strings.TrimSpace(cfg.PublicBaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	}
	return ""
}

func (c *minioObjectBackend) Name() string {
	if c == nil {
		return ""
	}
	return "minio"
}

func (c *minioObjectBackend) PutObject(ctx context.Context, req storagecontract.PutObjectRequest) (*storagecontract.PutObjectResult, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("minio backend is not configured")
	}
	if req.Reader == nil {
		return nil, fmt.Errorf("upload reader is required")
	}
	key, err := normalizeKey(req.Key)
	if err != nil {
		return nil, err
	}
	contentType := strings.TrimSpace(req.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	payload, err := io.ReadAll(req.Reader)
	if err != nil {
		return nil, err
	}
	if req.Size > 0 && int64(len(payload)) != req.Size {
		return nil, fmt.Errorf("upload file size mismatch: got %d want %d", len(payload), req.Size)
	}
	hasher := sha256.Sum256(payload)
	info, err := c.client.PutObject(ctx, c.bucket, key, bytes.NewReader(payload), int64(len(payload)), minio.PutObjectOptions{
		ContentType:          contentType,
		ContentDisposition:   strings.TrimSpace(req.ContentDisposition),
		SendContentMd5:       true,
		DisableContentSha256: true,
		UserMetadata:         buildMinIOUserMetadata(req),
	})
	if err != nil {
		return nil, err
	}
	c.storeMetadata(key, buildMinIOUserMetadata(req))
	checksum := hex.EncodeToString(hasher[:])
	if expected := strings.TrimSpace(req.ChecksumSHA256); expected != "" && !strings.EqualFold(expected, checksum) {
		return nil, fmt.Errorf("upload checksum mismatch: got %s want %s", checksum, expected)
	}
	return &storagecontract.PutObjectResult{
		Key:            key,
		StorageName:    filepath.Base(key),
		URL:            c.buildPublicURL(key),
		ETag:           strings.Trim(info.ETag, `"`),
		Size:           info.Size,
		ChecksumSHA256: checksum,
	}, nil
}

func (c *minioObjectBackend) GetObject(ctx context.Context, key string) (io.ReadCloser, *storagecontract.ObjectInfo, error) {
	if c == nil || c.client == nil {
		return nil, nil, fmt.Errorf("minio backend is not configured")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return nil, nil, err
	}
	object, err := c.client.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, err
	}
	info, err := object.Stat()
	if err != nil {
		_ = object.Close()
		return nil, nil, err
	}
	return object, c.objectInfoFromMinIO(info, key), nil
}

func (c *minioObjectBackend) HeadObject(ctx context.Context, key string) (*storagecontract.ObjectInfo, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("minio backend is not configured")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return nil, err
	}
	info, err := c.client.StatObject(ctx, c.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return c.objectInfoFromMinIO(info, key), nil
}

func (c *minioObjectBackend) DeleteObject(ctx context.Context, key string) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("minio backend is not configured")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return err
	}
	c.deleteMetadata(key)
	return c.client.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
}

func (c *minioObjectBackend) Exists(ctx context.Context, key string) (bool, error) {
	if c == nil || c.client == nil {
		return false, fmt.Errorf("minio backend is not configured")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return false, err
	}
	_, err = c.client.StatObject(ctx, c.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if isMinIONotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *minioObjectBackend) PublicURL(ctx context.Context, key string) (string, error) {
	if c == nil || c.client == nil {
		return "", fmt.Errorf("minio backend is not configured")
	}
	if strings.TrimSpace(c.publicBaseURL) == "" {
		return "", fmt.Errorf("minio public_base_url is required for public urls")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return "", err
	}
	publicURL := c.buildPublicURL(key)
	if strings.TrimSpace(publicURL) == "" {
		return "", fmt.Errorf("minio public_base_url is required for public urls")
	}
	return publicURL, nil
}

func (c *minioObjectBackend) SignedURL(ctx context.Context, key string, opts storagecontract.SignedURLOptions) (string, error) {
	if c == nil || c.client == nil {
		return "", fmt.Errorf("minio backend is not configured")
	}
	if strings.TrimSpace(c.publicBaseURL) == "" {
		return "", fmt.Errorf("minio public_base_url is required for signed urls")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return "", err
	}
	method := strings.ToUpper(strings.TrimSpace(opts.Method))
	switch method {
	case "", "GET":
		reqParams := url.Values{}
		if contentType := strings.TrimSpace(opts.ResponseContentType); contentType != "" {
			reqParams.Set("response-content-type", contentType)
		}
		if disposition := strings.TrimSpace(opts.ResponseContentDisposition); disposition != "" {
			reqParams.Set("response-content-disposition", disposition)
		}
		if opts.Expires <= 0 {
			return c.buildPublicURL(key), nil
		}
		presigned, err := c.client.PresignedGetObject(ctx, c.bucket, key, opts.Expires, reqParams)
		if err != nil {
			return "", err
		}
		return presigned.String(), nil
	case "HEAD":
		if opts.Expires <= 0 {
			return c.buildPublicURL(key), nil
		}
		presigned, err := c.client.PresignedHeadObject(ctx, c.bucket, key, opts.Expires, nil)
		if err != nil {
			return "", err
		}
		return presigned.String(), nil
	case "PUT":
		if opts.Expires <= 0 {
			return c.buildPublicURL(key), nil
		}
		presigned, err := c.client.PresignedPutObject(ctx, c.bucket, key, opts.Expires)
		if err != nil {
			return "", err
		}
		return presigned.String(), nil
	default:
		return "", fmt.Errorf("unsupported signed url method %q for minio backend", opts.Method)
	}
}

func (c *minioObjectBackend) ensureBucket(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("minio backend is not configured")
	}
	bucket := strings.TrimSpace(c.bucket)
	if bucket == "" {
		return fmt.Errorf("minio bucket is required")
	}
	exists, err := c.client.BucketExists(ctx, bucket)
	if err != nil {
		if isMinIONotFound(err) {
			exists = false
		} else {
			return err
		}
	}
	if exists {
		return nil
	}
	if err := c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: c.region}); err != nil {
		if isBucketAlreadyExists(err) {
			return nil
		}
		return err
	}
	return nil
}

func (c *minioObjectBackend) objectInfoFromMinIO(info minio.ObjectInfo, key string) *storagecontract.ObjectInfo {
	metadata := map[string]string{}
	for k, v := range info.UserMetadata {
		metadata[k] = v
	}
	if len(metadata) == 0 {
		for k, values := range info.Metadata {
			if len(values) == 0 {
				continue
			}
			metadata[strings.ToLower(k)] = values[0]
		}
	}
	if cached := c.cachedMetadata(key); len(cached) > 0 {
		if len(metadata) == 0 {
			metadata = make(map[string]string, len(cached))
		}
		for k, v := range cached {
			if _, exists := metadata[k]; !exists {
				metadata[k] = v
			}
		}
	}
	return &storagecontract.ObjectInfo{
		Key:         key,
		Size:        info.Size,
		ContentType: info.ContentType,
		ETag:        strings.Trim(info.ETag, `"`),
		ModTime:     info.LastModified,
		Metadata:    metadata,
		PublicURL:   c.buildPublicURL(key),
	}
}

func (c *minioObjectBackend) buildPublicURL(key string) string {
	base := strings.TrimRight(c.publicBaseURL, "/")
	if base == "" {
		return ""
	}
	return joinURL(base, key)
}

func (c *minioObjectBackend) storeMetadata(key string, metadata map[string]string) {
	if c == nil {
		return
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(metadata) == 0 {
		delete(c.metadataCache, key)
		return
	}
	clone := make(map[string]string, len(metadata))
	for k, v := range metadata {
		clone[k] = v
	}
	c.metadataCache[key] = clone
}

func (c *minioObjectBackend) cachedMetadata(key string) map[string]string {
	if c == nil {
		return map[string]string{}
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return map[string]string{}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	metadata := c.metadataCache[key]
	if len(metadata) == 0 {
		return map[string]string{}
	}
	clone := make(map[string]string, len(metadata))
	for k, v := range metadata {
		clone[k] = v
	}
	return clone
}

func (c *minioObjectBackend) deleteMetadata(key string) {
	if c == nil {
		return
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.metadataCache, key)
}

func buildMinIOUserMetadata(req storagecontract.PutObjectRequest) map[string]string {
	metadata := make(map[string]string, len(req.Metadata)+3)
	for key, value := range req.Metadata {
		metadata[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	if filename := strings.TrimSpace(req.Filename); filename != "" {
		metadata["filename"] = filename
	}
	if visibility := strings.TrimSpace(req.Visibility); visibility != "" {
		metadata["visibility"] = visibility
	}
	if checksum := strings.TrimSpace(req.ChecksumSHA256); checksum != "" {
		metadata["checksum_sha256"] = checksum
	}
	return metadata
}

func isMinIONotFound(err error) bool {
	if err == nil {
		return false
	}
	code := minio.ToErrorResponse(err).Code
	switch code {
	case "NoSuchBucket", "NoSuchKey", "NotFound", "XMinioNoSuchKey":
		return true
	default:
		return false
	}
}

func isBucketAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	code := minio.ToErrorResponse(err).Code
	switch code {
	case "BucketAlreadyExists", "BucketAlreadyOwnedByYou":
		return true
	default:
		return false
	}
}

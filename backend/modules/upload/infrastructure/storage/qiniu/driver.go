//go:build ignore
// +build ignore

package qiniu

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"goadmin/core/config"
	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"

	"github.com/qiniu/api.v7/v7/auth"
	"github.com/qiniu/api.v7/v7/storage"
)

const defaultQiniuUploadURL = "https://upload.qiniup.com"

type Driver struct {
	cfg           config.QiniuStorageConfig
	mac           *auth.Credentials
	uploader      *storage.FormUploader
	bucketManager *storage.BucketManager
	uploadURL     string
	zone          *storage.Zone
}

func NewDriver(cfg config.QiniuStorageConfig) (*Driver, error) {
	bucket := strings.TrimSpace(cfg.Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("upload storage qiniu bucket is required")
	}
	accessKey := strings.TrimSpace(cfg.AccessKeyID)
	if accessKey == "" {
		return nil, fmt.Errorf("upload storage qiniu access_key_id is required")
	}
	secretKey := strings.TrimSpace(cfg.AccessKeySecret)
	if secretKey == "" {
		return nil, fmt.Errorf("upload storage qiniu access_key_secret is required")
	}
	publicBaseURL := strings.TrimSpace(cfg.PublicBaseURL)
	if publicBaseURL == "" {
		return nil, fmt.Errorf("upload storage qiniu public_base_url is required")
	}
	zone := resolveZone(strings.TrimSpace(cfg.Region))
	qiniuCfg := &storage.Config{
		Zone:          zone,
		UseHTTPS:      strings.HasPrefix(strings.ToLower(strings.TrimSpace(cfg.UploadURL)), "https://"),
		UseCdnDomains: false,
	}
	mac := auth.New(accessKey, secretKey)
	uploader := storage.NewFormUploader(qiniuCfg)
	bucketManager := storage.NewBucketManager(mac, &storage.Config{Zone: zone})
	return &Driver{
		cfg:           cfg,
		mac:           mac,
		uploader:      uploader,
		bucketManager: bucketManager,
		uploadURL:     normalizeUploadURL(cfg.UploadURL),
		zone:          zone,
	}, nil
}

func (d *Driver) Name() string { return "qiniu" }

func (d *Driver) Put(ctx context.Context, req storagecontract.PutObjectRequest) (*storagecontract.PutObjectResult, error) {
	if d == nil {
		return nil, fmt.Errorf("qiniu storage driver is not configured")
	}
	if req.Reader == nil {
		return nil, fmt.Errorf("upload reader is required")
	}
	key, err := normalizeKey(req.Key)
	if err != nil {
		return nil, err
	}
	if err := ensureParentDir(key); err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp("", "goadmin-qiniu-upload-*")
	if err != nil {
		return nil, fmt.Errorf("create qiniu upload temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(tmp, hasher), req.Reader)
	if err != nil {
		return nil, fmt.Errorf("write qiniu upload file: %w", err)
	}
	if req.Size > 0 && written != req.Size {
		return nil, fmt.Errorf("upload file size mismatch: got %d want %d", written, req.Size)
	}
	if err := tmp.Sync(); err != nil {
		return nil, fmt.Errorf("sync qiniu upload file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("close qiniu upload temp file: %w", err)
	}

	file, err := os.Open(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("reopen qiniu upload temp file: %w", err)
	}
	defer func() { _ = file.Close() }()

	contentType := strings.TrimSpace(req.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(key))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	putPolicy := storage.PutPolicy{
		Scope:        d.cfg.Bucket + ":" + key,
		InsertOnly:   0,
		Expires:      3600,
		ForceSaveKey: false,
		SaveKey:      key,
	}
	uptoken := putPolicy.UploadToken(d.mac)
	ret := storage.PutRet{}
	if err := d.uploader.Put(ctx, &ret, uptoken, key, file, written, &storage.PutExtra{Params: map[string]string{}}); err != nil {
		return nil, err
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	if expected := strings.TrimSpace(req.ChecksumSHA256); expected != "" && !strings.EqualFold(expected, checksum) {
		return nil, fmt.Errorf("upload checksum mismatch: got %s want %s", checksum, expected)
	}
	publicURL := d.buildPublicURL(key)
	return &storagecontract.PutObjectResult{
		Key:            firstNonEmpty(strings.TrimSpace(ret.Key), key),
		StorageName:    filepath.Base(key),
		URL:            publicURL,
		ETag:           firstNonEmpty(strings.TrimSpace(ret.Hash), checksum),
		Size:           written,
		ChecksumSHA256: checksum,
	}, nil
}

func (d *Driver) Get(ctx context.Context, key string) (io.ReadCloser, *storagecontract.ObjectInfo, error) {
	if d == nil {
		return nil, nil, fmt.Errorf("qiniu storage driver is not configured")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return nil, nil, err
	}
	info, err := d.bucketManager.Stat(d.cfg.Bucket, key)
	if err != nil {
		if isQiniuNotFound(err) {
			return nil, nil, storagecontractErrNotFound(key)
		}
		return nil, nil, err
	}
	reader, err := d.openDownloadReader(ctx, key)
	if err != nil {
		return nil, nil, err
	}
	return reader, &storagecontract.ObjectInfo{
		Key:         key,
		Size:        info.Fsize,
		ContentType: firstNonEmpty(info.MimeType, mime.TypeByExtension(filepath.Ext(key)), "application/octet-stream"),
		ETag:        info.Hash,
		ModTime:     time.Unix(0, info.PutTime*100),
		PublicURL:   d.buildPublicURL(key),
	}, nil
}

func (d *Driver) Delete(ctx context.Context, key string) error {
	if d == nil {
		return fmt.Errorf("qiniu storage driver is not configured")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return err
	}
	if err := d.bucketManager.Delete(d.cfg.Bucket, key); err != nil {
		if isQiniuNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func (d *Driver) Exists(ctx context.Context, key string) (bool, error) {
	if d == nil {
		return false, fmt.Errorf("qiniu storage driver is not configured")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return false, err
	}
	_, err = d.bucketManager.Stat(d.cfg.Bucket, key)
	if err != nil {
		if isQiniuNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *Driver) PublicURL(ctx context.Context, key string) (string, error) {
	if d == nil {
		return "", fmt.Errorf("qiniu storage driver is not configured")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return "", err
	}
	return d.buildPublicURL(key), nil
}

func (d *Driver) SignedURL(ctx context.Context, key string, opts storagecontract.SignedURLOptions) (string, error) {
	if d == nil {
		return "", fmt.Errorf("qiniu storage driver is not configured")
	}
	key, err := normalizeKey(key)
	if err != nil {
		return "", err
	}
	if opts.Expires <= 0 {
		return d.buildPublicURL(key), nil
	}
	deadline := time.Now().Add(opts.Expires).Unix()
	query := url.Values{}
	if contentType := strings.TrimSpace(opts.ResponseContentType); contentType != "" {
		query.Set("response-content-type", contentType)
	}
	if disposition := strings.TrimSpace(opts.ResponseContentDisposition); disposition != "" {
		query.Set("response-content-disposition", disposition)
	}
	if len(query) == 0 {
		return storage.MakePrivateURLv2(d.mac, d.buildPublicURL(key), key, deadline), nil
	}
	return storage.MakePrivateURLv2WithQuery(d.mac, d.buildPublicURL(key), key, query, deadline), nil
}

func (d *Driver) buildPublicURL(key string) string {
	base := strings.TrimRight(strings.TrimSpace(d.cfg.PublicBaseURL), "/")
	if base == "" {
		return ""
	}
	return storage.MakePublicURLv2(base, key)
}

func (d *Driver) openDownloadReader(ctx context.Context, key string) (io.ReadCloser, error) {
	url := d.buildPublicURL(key)
	if url == "" {
		return nil, fmt.Errorf("qiniu public_base_url is required")
	}
	if strings.Contains(strings.ToLower(url), "token=") {
		// already signed
	} else {
		signedURL, err := d.SignedURL(ctx, key, storagecontract.SignedURLOptions{Expires: time.Hour})
		if err != nil {
			return nil, err
		}
		url = signedURL
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		return nil, fmt.Errorf("qiniu download failed: %s", resp.Status)
	}
	return resp.Body, nil
}

func resolveZone(region string) *storage.Zone {
	switch strings.ToLower(strings.TrimSpace(region)) {
	case "huabei", "z1", "northchina":
		return &storage.ZoneHuabei
	case "huanan", "z2", "southchina":
		return &storage.ZoneHuanan
	case "beimei", "na0", "northamerica":
		return &storage.ZoneBeimei
	case "xinjiapo", "sg", "singapore":
		return &storage.ZoneXinjiapo
	case "huadong", "z0", "eastchina", "":
		fallthrough
	default:
		return &storage.ZoneHuadong
	}
}

func normalizeUploadURL(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return defaultQiniuUploadURL
	}
	return strings.TrimRight(trimmed, "/")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func isQiniuNotFound(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "no such file or directory") || strings.Contains(text, "no such key") || strings.Contains(text, "status code 612") || strings.Contains(text, "not found")
}

func storagecontractErrNotFound(key string) error {
	return fmt.Errorf("upload file asset %s not found", key)
}

func ensureParentDir(key string) error {
	if strings.Contains(key, "../") {
		return fmt.Errorf("invalid qiniu object key")
	}
	return nil
}

func normalizeKey(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", fmt.Errorf("upload storage key is required")
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = filepath.Clean(trimmed)
	if trimmed == "." || trimmed == "" {
		return "", fmt.Errorf("upload storage key is invalid")
	}
	if strings.HasPrefix(trimmed, "../") || trimmed == ".." || strings.Contains(trimmed, "/../") {
		return "", fmt.Errorf("upload storage key contains path traversal")
	}
	return strings.TrimPrefix(trimmed, "/"), nil
}

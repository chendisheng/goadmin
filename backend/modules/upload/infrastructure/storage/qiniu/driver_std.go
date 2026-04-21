package qiniu

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"goadmin/core/config"
	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
)

const (
	defaultQiniuUploadURL = "https://upload.qiniup.com"
	defaultQiniuRsURL     = "https://rs.qiniu.com"
)

type Driver struct {
	bucket        string
	accessKey     string
	secretKey     []byte
	uploadURL     string
	rsURL         string
	publicBaseURL string
	httpClient    *http.Client
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
	return &Driver{
		bucket:        bucket,
		accessKey:     accessKey,
		secretKey:     []byte(secretKey),
		uploadURL:     normalizeBaseURL(cfg.UploadURL, defaultQiniuUploadURL),
		rsURL:         defaultQiniuRsURL,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		httpClient:    &http.Client{Timeout: 60 * time.Second},
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
	if err := ensureSafeKey(key); err != nil {
		return nil, err
	}

	contentType := strings.TrimSpace(req.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(key))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
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

	uptoken, err := d.uploadToken(key, time.Now().Add(time.Hour))
	if err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("token", uptoken); err != nil {
		return nil, err
	}
	if err := writer.WriteField("key", key); err != nil {
		return nil, err
	}
	if visibility := strings.TrimSpace(req.Visibility); visibility != "" {
		_ = writer.WriteField("x:visibility", visibility)
	}
	for k, v := range req.Metadata {
		k = strings.TrimSpace(k)
		if k == "" || strings.TrimSpace(v) == "" {
			continue
		}
		if !strings.HasPrefix(k, "x:") {
			k = "x:" + k
		}
		_ = writer.WriteField(k, v)
	}
	part, err := writer.CreateFormFile("file", filepath.Base(key))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, d.uploadURL, body)
	if err != nil {
		return nil, err
	}
	reqHTTP.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := d.httpClient.Do(reqHTTP)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("qiniu upload failed: %s: %s", resp.Status, strings.TrimSpace(string(respBytes)))
	}

	var uploadResp struct {
		Key   string `json:"key"`
		Hash  string `json:"hash"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(respBytes, &uploadResp); err != nil {
		return nil, fmt.Errorf("decode qiniu upload response: %w", err)
	}
	if strings.TrimSpace(uploadResp.Error) != "" {
		return nil, fmt.Errorf("qiniu upload error: %s", uploadResp.Error)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	if expected := strings.TrimSpace(req.ChecksumSHA256); expected != "" && !strings.EqualFold(expected, checksum) {
		return nil, fmt.Errorf("upload checksum mismatch: got %s want %s", checksum, expected)
	}

	publicURL, _ := d.PublicURL(ctx, key)
	return &storagecontract.PutObjectResult{
		Key:            firstNonEmpty(strings.TrimSpace(uploadResp.Key), key),
		StorageName:    filepath.Base(key),
		URL:            publicURL,
		ETag:           firstNonEmpty(strings.TrimSpace(uploadResp.Hash), checksum),
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
	info, err := d.statObject(ctx, key)
	if err != nil {
		return nil, nil, err
	}

	reader, err := d.openDownloadReader(ctx, key)
	if err != nil {
		return nil, nil, err
	}
	return reader, &storagecontract.ObjectInfo{
		Key:         key,
		Size:        info.SizeBytes,
		ContentType: firstNonEmpty(info.ContentType, mime.TypeByExtension(filepath.Ext(key)), "application/octet-stream"),
		ETag:        info.Hash,
		ModTime:     time.Unix(0, info.PutTime*100),
		Metadata:    map[string]string{},
		PublicURL:   func() string { v, _ := d.PublicURL(ctx, key); return v }(),
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
	entry := base64.RawURLEncoding.EncodeToString([]byte(d.bucket + ":" + key))
	reqURL := strings.TrimRight(d.rsURL, "/") + "/delete/" + entry
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return err
	}
	reqHTTP.Header.Set("Authorization", d.qboxAuthorization(reqHTTP.URL.RequestURI(), nil))
	resp, err := d.httpClient.Do(reqHTTP)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qiniu delete failed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
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
	_, err = d.statObject(ctx, key)
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
	base := strings.TrimRight(d.publicBaseURL, "/")
	if base == "" {
		return "", fmt.Errorf("qiniu public_base_url is required")
	}
	return base + "/" + key, nil
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
		return d.PublicURL(ctx, key)
	}
	publicURL, err := d.PublicURL(ctx, key)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(publicURL)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	if v := strings.TrimSpace(opts.ResponseContentType); v != "" {
		query.Set("response-content-type", v)
	}
	if v := strings.TrimSpace(opts.ResponseContentDisposition); v != "" {
		query.Set("response-content-disposition", v)
	}
	deadline := time.Now().Add(opts.Expires).Unix()
	query.Set("e", strconv.FormatInt(deadline, 10))
	parsed.RawQuery = query.Encode()
	unsigned := parsed.String()
	token := d.qiniuSign([]byte(unsigned))
	if parsed.RawQuery == "" {
		return unsigned + "?token=" + token, nil
	}
	return unsigned + "&token=" + token, nil
}

func (d *Driver) openDownloadReader(ctx context.Context, key string) (io.ReadCloser, error) {
	urlStr, err := d.SignedURL(ctx, key, storagecontract.SignedURLOptions{Expires: time.Hour})
	if err != nil {
		return nil, err
	}
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.httpClient.Do(reqHTTP)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() { _ = resp.Body.Close() }()
		return nil, fmt.Errorf("qiniu download failed: %s", resp.Status)
	}
	return resp.Body, nil
}

type qiniuStatInfo struct {
	Hash        string `json:"hash"`
	SizeBytes   int64  `json:"fsize"`
	ContentType string `json:"mimeType"`
	PutTime     int64  `json:"putTime"`
}

func (d *Driver) statObject(ctx context.Context, key string) (*qiniuStatInfo, error) {
	entry := base64.RawURLEncoding.EncodeToString([]byte(d.bucket + ":" + key))
	reqURL := strings.TrimRight(d.rsURL, "/") + "/stat/" + entry
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return nil, err
	}
	reqHTTP.Header.Set("Authorization", d.qboxAuthorization(reqHTTP.URL.RequestURI(), nil))
	resp, err := d.httpClient.Do(reqHTTP)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("upload object %s not found", key)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("qiniu stat failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var direct qiniuStatInfo
	if err := json.Unmarshal(body, &direct); err == nil && (direct.Hash != "" || direct.SizeBytes > 0 || direct.PutTime > 0) {
		return &direct, nil
	}
	var wrapped struct {
		Data  qiniuStatInfo `json:"data"`
		Code  int           `json:"code"`
		Msg   string        `json:"msg"`
		Error string        `json:"error"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, err
	}
	if wrapped.Error != "" {
		return nil, fmt.Errorf("qiniu stat error: %s", wrapped.Error)
	}
	if wrapped.Msg != "" && wrapped.Code != 0 {
		return nil, fmt.Errorf("qiniu stat error: %s", wrapped.Msg)
	}
	return &wrapped.Data, nil
}

func (d *Driver) qboxAuthorization(requestURI string, body []byte) string {
	mac := d.qiniuSign([]byte(requestURI))
	return "QBox " + d.accessKey + ":" + mac
}

func (d *Driver) uploadToken(key string, deadline time.Time) (string, error) {
	policy := map[string]any{
		"scope":    d.bucket + ":" + key,
		"deadline": deadline.Unix(),
	}
	encodedPolicy, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(encodedPolicy)
	return d.accessKey + ":" + d.qiniuSign([]byte(encoded)) + ":" + encoded, nil
}

func (d *Driver) qiniuSign(data []byte) string {
	h := hmac.New(sha1.New, d.secretKey)
	_, _ = h.Write(data)
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func (d *Driver) buildPublicURL(key string) string {
	base := strings.TrimRight(d.publicBaseURL, "/")
	if base == "" {
		return ""
	}
	return base + "/" + key
}

func normalizeBaseURL(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = fallback
	}
	return strings.TrimRight(trimmed, "/")
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

func ensureSafeKey(key string) error {
	if strings.Contains(key, "\x00") {
		return fmt.Errorf("upload storage key contains invalid byte")
	}
	return nil
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
	return strings.Contains(text, "not found") || strings.Contains(text, "status code 612") || strings.Contains(text, "no such key")
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

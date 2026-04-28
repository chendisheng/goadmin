package objectstore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"goadmin/core/config"
	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
)

func TestS3CompatibleDriverLifecycle(t *testing.T) {
	t.Parallel()

	driver, err := NewS3CompatibleDriver(config.S3CompatibleConfig{
		Endpoint:        "https://s3.example.com",
		Region:          "us-east-1",
		Bucket:          "goadmin-upload-" + sanitizeSegment(t.Name()),
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		PublicBaseURL:   "https://cdn.example.com/uploads",
	})
	if err != nil {
		t.Fatalf("NewS3CompatibleDriver: %v", err)
	}
	assertObjectDriverLifecycle(t, driver, "s3-compatible")
}

func TestOSSDriverLifecycle(t *testing.T) {
	t.Parallel()

	driver, err := NewOSSDriver(config.OSSStorageConfig{
		Endpoint:        "https://oss.example.com",
		Bucket:          "goadmin-upload-" + sanitizeSegment(t.Name()),
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		PublicBaseURL:   "https://cdn.example.com/uploads",
	})
	if err != nil {
		t.Fatalf("NewOSSDriver: %v", err)
	}
	assertObjectDriverLifecycle(t, driver, "oss")
}

func TestCOSDriverLifecycle(t *testing.T) {
	t.Parallel()

	driver, err := NewCOSDriver(config.COSStorageConfig{
		Region:        "ap-guangzhou",
		Bucket:        "goadmin-upload-" + sanitizeSegment(t.Name()),
		SecretID:      "ak",
		SecretKey:     "sk",
		PublicBaseURL: "https://cdn.example.com/uploads",
	})
	if err != nil {
		t.Fatalf("NewCOSDriver: %v", err)
	}
	assertObjectDriverLifecycle(t, driver, "cos")
}

func TestMinIODriverLifecycle(t *testing.T) {
	t.Parallel()

	endpoint, cleanup := newMinIOTestServer(t)
	defer cleanup()

	driver, err := NewMinIODriver(config.MinIOStorageConfig{
		Endpoint:        endpoint,
		Bucket:          "goadmin-upload-" + sanitizeSegment(t.Name()),
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		UseSSL:          false,
		PathStyle:       true,
		PublicBaseURL:   "https://cdn.example.com/uploads",
	})
	if err != nil {
		t.Fatalf("NewMinIODriver: %v", err)
	}
	assertObjectDriverLifecycle(t, driver, "minio")
}

func TestSignedURLIncludesQueryParameters(t *testing.T) {
	t.Parallel()

	driver, err := NewS3CompatibleDriver(config.S3CompatibleConfig{
		Endpoint:        "https://s3.example.com",
		Region:          "us-east-1",
		Bucket:          "goadmin-upload-" + sanitizeSegment(t.Name()),
		AccessKeyID:     "ak",
		AccessKeySecret: "sk",
		PublicBaseURL:   "https://cdn.example.com/uploads",
	})
	if err != nil {
		t.Fatalf("NewS3CompatibleDriver: %v", err)
	}
	url, err := driver.SignedURL(context.Background(), "files/report.pdf", storagecontract.SignedURLOptions{
		Method:                     "get",
		Expires:                    time.Minute,
		ResponseContentType:        "application/pdf",
		ResponseContentDisposition: "attachment; filename=report.pdf",
	})
	if err != nil {
		t.Fatalf("SignedURL: %v", err)
	}
	for _, want := range []string{"method=GET", "expires=1m0s", "response_content_type=application%2Fpdf", "response_content_disposition=attachment%3B+filename%3Dreport.pdf"} {
		if !strings.Contains(url, want) {
			t.Fatalf("SignedURL(%q) missing %q", url, want)
		}
	}
}

func assertObjectDriverLifecycle(t *testing.T, driver *Driver, wantName string) {
	t.Helper()

	if got := driver.Name(); got != wantName {
		t.Fatalf("Name() = %q, want %q", got, wantName)
	}
	key := "folder/demo.txt"
	payload := []byte("hello object storage")
	result, err := driver.Put(context.Background(), storagecontract.PutObjectRequest{
		Key:         key,
		Reader:      bytes.NewReader(payload),
		Size:        int64(len(payload)),
		ContentType: "text/plain",
		Filename:    "demo.txt",
		Metadata: map[string]string{
			"source": t.Name(),
		},
		Visibility:         "private",
		ChecksumSHA256:     "",
		ContentDisposition: "attachment; filename=demo.txt",
	})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if result.Key != key {
		t.Fatalf("Put result key = %q, want %q", result.Key, key)
	}
	if result.URL == "" || result.ChecksumSHA256 == "" {
		t.Fatalf("unexpected Put result: %+v", result)
	}

	exists, err := driver.Exists(context.Background(), key)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Fatal("expected uploaded object to exist")
	}

	reader, info, err := driver.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read object: %v", err)
	}
	if string(data) != string(payload) {
		t.Fatalf("Get payload = %q, want %q", string(data), string(payload))
	}
	if info == nil || info.PublicURL == "" || info.ContentType != "text/plain" {
		t.Fatalf("unexpected info: %+v", info)
	}
	if got := info.Metadata["source"]; got != t.Name() {
		t.Fatalf("metadata[source] = %q, want %q", got, t.Name())
	}

	url, err := driver.PublicURL(context.Background(), key)
	if err != nil {
		t.Fatalf("PublicURL: %v", err)
	}
	if url == "" {
		t.Fatal("PublicURL returned empty string")
	}

	if err := driver.Delete(context.Background(), key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	exists, err = driver.Exists(context.Background(), key)
	if err != nil {
		t.Fatalf("Exists after delete: %v", err)
	}
	if exists {
		t.Fatal("expected deleted object to be removed")
	}
}

type minioTestServer struct {
	t       *testing.T
	mu      sync.RWMutex
	buckets map[string]map[string]*minioTestObject
}

type minioTestObject struct {
	data        []byte
	contentType string
	metadata    map[string]string
	modTime     time.Time
	etag        string
}

func newMinIOTestServer(t *testing.T) (string, func()) {
	t.Helper()
	server := &minioTestServer{t: t, buckets: make(map[string]map[string]*minioTestObject)}
	httpServer := httptest.NewServer(server)
	return httpServer.URL, httpServer.Close
}

func (s *minioTestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bucket, object, ok := splitMinIOPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch {
	case object == "":
		s.handleBucket(w, r, bucket)
	default:
		s.handleObject(w, r, bucket, object)
	}
}

func (s *minioTestServer) handleBucket(w http.ResponseWriter, r *http.Request, bucket string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch r.Method {
	case http.MethodHead:
		if _, ok := s.buckets[bucket]; !ok {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	case http.MethodPut:
		if _, ok := s.buckets[bucket]; !ok {
			s.buckets[bucket] = make(map[string]*minioTestObject)
		}
		w.Header().Set("x-amz-bucket-region", "us-east-1")
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		if _, ok := r.URL.Query()["location"]; ok {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/"></LocationConstraint>`))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *minioTestServer) handleObject(w http.ResponseWriter, r *http.Request, bucket, object string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	objects, ok := s.buckets[bucket]
	if !ok {
		if r.Method == http.MethodHead {
			http.NotFound(w, r)
			return
		}
		objects = make(map[string]*minioTestObject)
		s.buckets[bucket] = objects
	}
	switch r.Method {
	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		metadata := make(map[string]string)
		for key, values := range r.Header {
			lower := strings.ToLower(key)
			switch {
			case strings.HasPrefix(lower, "x-amz-meta-") && len(values) > 0:
				metadata[strings.TrimPrefix(lower, "x-amz-meta-")] = values[0]
			case lower == "content-disposition" && len(values) > 0:
				metadata["content_disposition"] = values[0]
			}
		}
		etag := fmt.Sprintf(`"%x"`, body)
		objects[object] = &minioTestObject{
			data:        body,
			contentType: r.Header.Get("Content-Type"),
			metadata:    metadata,
			modTime:     time.Now().UTC(),
			etag:        etag,
		}
		w.Header().Set("ETag", etag)
		w.WriteHeader(http.StatusOK)
	case http.MethodHead:
		obj, ok := objects[object]
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeMinIOObjectHeaders(w, obj)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(obj.data)))
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		obj, ok := objects[object]
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeMinIOObjectHeaders(w, obj)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(obj.data)))
		_, _ = w.Write(obj.data)
	case http.MethodDelete:
		delete(objects, object)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeMinIOObjectHeaders(w http.ResponseWriter, obj *minioTestObject) {
	if obj == nil {
		return
	}
	if obj.contentType != "" {
		w.Header().Set("Content-Type", obj.contentType)
	}
	if obj.etag != "" {
		w.Header().Set("ETag", obj.etag)
	}
	w.Header().Set("Last-Modified", obj.modTime.UTC().Format(http.TimeFormat))
	for key, value := range obj.metadata {
		w.Header().Set("x-amz-meta-"+key, value)
	}
}

func splitMinIOPath(p string) (string, string, bool) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(p), "/")
	if trimmed == "" {
		return "", "", false
	}
	parts := strings.SplitN(trimmed, "/", 2)
	bucket := parts[0]
	if bucket == "" {
		return "", "", false
	}
	if len(parts) == 1 {
		return bucket, "", true
	}
	return bucket, parts[1], true
}

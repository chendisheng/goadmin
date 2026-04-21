package storage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"goadmin/core/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewDriverSelectsObjectStorageImplementations(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		cfg  config.UploadStorageConfig
		want string
	}{
		{
			name: "db",
			cfg: config.UploadStorageConfig{
				Driver: "db",
			},
			want: "db",
		},
		{
			name: "database",
			cfg: config.UploadStorageConfig{
				Driver: "database",
			},
			want: "db",
		},
		{
			name: "s3-compatible",
			cfg: config.UploadStorageConfig{
				Driver: "s3-compatible",
				S3Compatible: config.S3CompatibleConfig{
					Endpoint:        "https://s3.example.com",
					Region:          "us-east-1",
					Bucket:          "goadmin-upload-s3",
					AccessKeyID:     "ak",
					AccessKeySecret: "sk",
				},
			},
			want: "s3-compatible",
		},
		{
			name: "oss",
			cfg: config.UploadStorageConfig{
				Driver: "oss",
				OSS: config.OSSStorageConfig{
					Endpoint:        "https://oss.example.com",
					Bucket:          "goadmin-upload-oss",
					AccessKeyID:     "ak",
					AccessKeySecret: "sk",
				},
			},
			want: "oss",
		},
		{
			name: "cos",
			cfg: config.UploadStorageConfig{
				Driver: "cos",
				COS: config.COSStorageConfig{
					Region:    "ap-guangzhou",
					Bucket:    "goadmin-upload-cos",
					SecretID:  "ak",
					SecretKey: "sk",
				},
			},
			want: "cos",
		},
		{
			name: "minio",
			cfg: config.UploadStorageConfig{
				Driver: "minio",
				MinIO: config.MinIOStorageConfig{
					Endpoint:        minIOFactoryEndpoint(t),
					Bucket:          "goadmin-upload-minio",
					AccessKeyID:     "ak",
					AccessKeySecret: "sk",
					UseSSL:          false,
					PathStyle:       true,
				},
			},
			want: "minio",
		},
		{
			name: "qiniu",
			cfg: config.UploadStorageConfig{
				Driver: "qiniu",
				Qiniu: config.QiniuStorageConfig{
					Bucket:          "goadmin-upload-qiniu",
					AccessKeyID:     "ak",
					AccessKeySecret: "sk",
					PublicBaseURL:   "https://cdn.example.com/uploads",
				},
			},
			want: "qiniu",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			driver, err := newTestDriver(tc.cfg)
			if err != nil {
				t.Fatalf("NewDriver returned error: %v", err)
			}
			if got := driver.Name(); got != tc.want {
				t.Fatalf("driver.Name() = %q, want %q", got, tc.want)
			}
		})
	}
}

func newTestDriver(cfg config.UploadStorageConfig) (interface{ Name() string }, error) {
	if strings.EqualFold(strings.TrimSpace(cfg.Driver), "db") || strings.EqualFold(strings.TrimSpace(cfg.Driver), "database") {
		db, err := gorm.Open(sqlite.Open(testSQLiteDSN(cfg.Driver)), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		return NewDriverWithDB(db, cfg)
	}
	return NewDriver(cfg)
}

func testSQLiteDSN(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "default"
	}
	name = strings.NewReplacer(" ", "_", "/", "_", "\\", "_").Replace(name)
	return fmt.Sprintf("file:%s-%s?mode=memory&cache=shared", name, strings.ToLower(name))
}

func minIOFactoryEndpoint(t *testing.T) string {
	t.Helper()
	server := &minIOFactoryServer{buckets: make(map[string]struct{})}
	t.Cleanup(func() {
		if server.httpServer != nil {
			server.httpServer.Close()
		}
	})
	server.httpServer = httptest.NewServer(server)
	return server.httpServer.URL
}

type minIOFactoryServer struct {
	httpServer *httptest.Server
	mu         sync.Mutex
	buckets    map[string]struct{}
}

func (s *minIOFactoryServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bucket := strings.Trim(strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)[0], "/")
	if bucket == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	switch r.Method {
	case http.MethodGet:
		if _, ok := r.URL.Query()["location"]; ok {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/"></LocationConstraint>`))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	case http.MethodHead:
		if _, ok := s.buckets[bucket]; !ok {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	case http.MethodPut:
		s.buckets[bucket] = struct{}{}
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func TestNewDriverRejectsMissingObjectStorageConfiguration(t *testing.T) {
	t.Parallel()

	_, err := NewDriver(config.UploadStorageConfig{Driver: "s3-compatible"})
	if err == nil {
		t.Fatal("expected missing s3-compatible configuration to fail")
	}
}

func TestNewDriverRejectsUnsupportedObjectStorageDriver(t *testing.T) {
	t.Parallel()

	_, err := NewDriver(config.UploadStorageConfig{Driver: "ftp"})
	if err == nil {
		t.Fatal("expected unsupported driver to fail")
	}
}

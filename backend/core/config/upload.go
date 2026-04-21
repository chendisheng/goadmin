package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type UploadConfig struct {
	Storage UploadStorageConfig `mapstructure:"storage" yaml:"storage"`
}

type UploadStorageConfig struct {
	Driver       string              `mapstructure:"driver" yaml:"driver"`
	Local        LocalStorageConfig  `mapstructure:"local" yaml:"local"`
	S3Compatible S3CompatibleConfig  `mapstructure:"s3_compatible" yaml:"s3_compatible"`
	OSS          OSSStorageConfig    `mapstructure:"oss" yaml:"oss"`
	COS          COSStorageConfig    `mapstructure:"cos" yaml:"cos"`
	Qiniu        QiniuStorageConfig  `mapstructure:"qiniu" yaml:"qiniu"`
	MinIO        MinIOStorageConfig  `mapstructure:"minio" yaml:"minio"`
	Policy       StoragePolicyConfig `mapstructure:"policy" yaml:"policy"`
}

type StoragePolicyConfig struct {
	MaxUploadSize     string   `mapstructure:"max_upload_size" yaml:"max_upload_size"`
	AllowedExtensions []string `mapstructure:"allowed_extensions" yaml:"allowed_extensions"`
	AllowedMIMETypes  []string `mapstructure:"allowed_mime_types" yaml:"allowed_mime_types"`
	VisibilityDefault string   `mapstructure:"visibility_default" yaml:"visibility_default"`
	PathPrefix        string   `mapstructure:"path_prefix" yaml:"path_prefix"`
}

type LocalStorageConfig struct {
	BaseDir          string `mapstructure:"base_dir" yaml:"base_dir"`
	PublicBaseURL    string `mapstructure:"public_base_url" yaml:"public_base_url"`
	UseProxyDownload bool   `mapstructure:"use_proxy_download" yaml:"use_proxy_download"`
}

type S3CompatibleConfig struct {
	Endpoint        string `mapstructure:"endpoint" yaml:"endpoint"`
	Region          string `mapstructure:"region" yaml:"region"`
	Bucket          string `mapstructure:"bucket" yaml:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" yaml:"access_key_secret"`
	UseSSL          bool   `mapstructure:"use_ssl" yaml:"use_ssl"`
	PathStyle       bool   `mapstructure:"path_style" yaml:"path_style"`
	PublicBaseURL   string `mapstructure:"public_base_url" yaml:"public_base_url"`
}

type OSSStorageConfig struct {
	Endpoint        string `mapstructure:"endpoint" yaml:"endpoint"`
	Bucket          string `mapstructure:"bucket" yaml:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" yaml:"access_key_secret"`
	PublicBaseURL   string `mapstructure:"public_base_url" yaml:"public_base_url"`
}

type COSStorageConfig struct {
	Region        string `mapstructure:"region" yaml:"region"`
	Bucket        string `mapstructure:"bucket" yaml:"bucket"`
	SecretID      string `mapstructure:"secret_id" yaml:"secret_id"`
	SecretKey     string `mapstructure:"secret_key" yaml:"secret_key"`
	PublicBaseURL string `mapstructure:"public_base_url" yaml:"public_base_url"`
}

type QiniuStorageConfig struct {
	Region          string `mapstructure:"region" yaml:"region"`
	Bucket          string `mapstructure:"bucket" yaml:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" yaml:"access_key_secret"`
	UploadURL       string `mapstructure:"upload_url" yaml:"upload_url"`
	PublicBaseURL   string `mapstructure:"public_base_url" yaml:"public_base_url"`
}

type MinIOStorageConfig struct {
	Endpoint        string `mapstructure:"endpoint" yaml:"endpoint"`
	Bucket          string `mapstructure:"bucket" yaml:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret" yaml:"access_key_secret"`
	UseSSL          bool   `mapstructure:"use_ssl" yaml:"use_ssl"`
	PathStyle       bool   `mapstructure:"path_style" yaml:"path_style"`
	PublicBaseURL   string `mapstructure:"public_base_url" yaml:"public_base_url"`
}

func DefaultUploadConfig() UploadConfig {
	cfg := UploadConfig{
		Storage: UploadStorageConfig{
			Driver: "local",
			Local: LocalStorageConfig{
				BaseDir:          filepath.Join(os.TempDir(), "goadmin", "uploads"),
				PublicBaseURL:    "/uploads/files",
				UseProxyDownload: true,
			},
			Policy: StoragePolicyConfig{
				MaxUploadSize:     "20mb",
				AllowedExtensions: []string{".png", ".jpg", ".jpeg", ".webp", ".pdf", ".txt"},
				AllowedMIMETypes:  []string{"image/png", "image/jpeg", "image/webp", "application/pdf", "text/plain"},
				VisibilityDefault: "private",
				PathPrefix:        "uploads",
			},
		},
	}
	cfg.Normalize()
	return cfg
}

func (c *UploadConfig) Normalize() {
	if c == nil {
		return
	}
	if strings.TrimSpace(c.Storage.Driver) == "" {
		c.Storage.Driver = "local"
	}
	if strings.TrimSpace(c.Storage.Local.BaseDir) == "" {
		c.Storage.Local.BaseDir = filepath.Join(os.TempDir(), "goadmin", "uploads")
	}
	if strings.TrimSpace(c.Storage.Local.PublicBaseURL) == "" {
		c.Storage.Local.PublicBaseURL = "/uploads/files"
	}
	if strings.TrimSpace(c.Storage.Policy.MaxUploadSize) == "" {
		c.Storage.Policy.MaxUploadSize = "20mb"
	}
	if len(c.Storage.Policy.AllowedExtensions) == 0 {
		c.Storage.Policy.AllowedExtensions = []string{".png", ".jpg", ".jpeg", ".webp", ".pdf", ".txt"}
	}
	if len(c.Storage.Policy.AllowedMIMETypes) == 0 {
		c.Storage.Policy.AllowedMIMETypes = []string{"image/png", "image/jpeg", "image/webp", "application/pdf", "text/plain"}
	}
	if strings.TrimSpace(c.Storage.Policy.VisibilityDefault) == "" {
		c.Storage.Policy.VisibilityDefault = "private"
	}
	if strings.TrimSpace(c.Storage.Policy.PathPrefix) == "" {
		c.Storage.Policy.PathPrefix = "uploads"
	}
}

func (c UploadConfig) Validate() error {
	cfg := c
	cfg.Normalize()
	return cfg.validate()
}

func (c UploadConfig) validate() error {
	driver := strings.ToLower(strings.TrimSpace(c.Storage.Driver))
	if driver == "" {
		return fmt.Errorf("upload.storage.driver is required")
	}
	switch driver {
	case "local":
		if strings.TrimSpace(c.Storage.Local.BaseDir) == "" {
			return fmt.Errorf("upload.storage.local.base_dir is required when upload.storage.driver=local")
		}
		if strings.TrimSpace(c.Storage.Local.PublicBaseURL) == "" {
			return fmt.Errorf("upload.storage.local.public_base_url is required when upload.storage.driver=local")
		}
	case "s3-compatible", "minio":
		if strings.TrimSpace(c.Storage.S3Compatible.Endpoint) == "" && driver == "s3-compatible" {
			return fmt.Errorf("upload.storage.s3_compatible.endpoint is required when upload.storage.driver=s3-compatible")
		}
		if strings.TrimSpace(c.Storage.S3Compatible.Bucket) == "" && driver == "s3-compatible" {
			return fmt.Errorf("upload.storage.s3_compatible.bucket is required when upload.storage.driver=s3-compatible")
		}
		if driver == "s3-compatible" {
			if strings.TrimSpace(c.Storage.S3Compatible.AccessKeyID) == "" {
				return fmt.Errorf("upload.storage.s3_compatible.access_key_id is required when upload.storage.driver=s3-compatible")
			}
			if strings.TrimSpace(c.Storage.S3Compatible.AccessKeySecret) == "" {
				return fmt.Errorf("upload.storage.s3_compatible.access_key_secret is required when upload.storage.driver=s3-compatible")
			}
		}
		if driver == "minio" {
			if strings.TrimSpace(c.Storage.MinIO.Endpoint) == "" {
				return fmt.Errorf("upload.storage.minio.endpoint is required when upload.storage.driver=minio")
			}
			if strings.TrimSpace(c.Storage.MinIO.Bucket) == "" {
				return fmt.Errorf("upload.storage.minio.bucket is required when upload.storage.driver=minio")
			}
			if strings.TrimSpace(c.Storage.MinIO.AccessKeyID) == "" {
				return fmt.Errorf("upload.storage.minio.access_key_id is required when upload.storage.driver=minio")
			}
			if strings.TrimSpace(c.Storage.MinIO.AccessKeySecret) == "" {
				return fmt.Errorf("upload.storage.minio.access_key_secret is required when upload.storage.driver=minio")
			}
		}
	case "oss":
		if strings.TrimSpace(c.Storage.OSS.Endpoint) == "" {
			return fmt.Errorf("upload.storage.oss.endpoint is required when upload.storage.driver=oss")
		}
		if strings.TrimSpace(c.Storage.OSS.Bucket) == "" {
			return fmt.Errorf("upload.storage.oss.bucket is required when upload.storage.driver=oss")
		}
		if strings.TrimSpace(c.Storage.OSS.AccessKeyID) == "" {
			return fmt.Errorf("upload.storage.oss.access_key_id is required when upload.storage.driver=oss")
		}
		if strings.TrimSpace(c.Storage.OSS.AccessKeySecret) == "" {
			return fmt.Errorf("upload.storage.oss.access_key_secret is required when upload.storage.driver=oss")
		}
	case "cos":
		if strings.TrimSpace(c.Storage.COS.Region) == "" {
			return fmt.Errorf("upload.storage.cos.region is required when upload.storage.driver=cos")
		}
		if strings.TrimSpace(c.Storage.COS.Bucket) == "" {
			return fmt.Errorf("upload.storage.cos.bucket is required when upload.storage.driver=cos")
		}
		if strings.TrimSpace(c.Storage.COS.SecretID) == "" {
			return fmt.Errorf("upload.storage.cos.secret_id is required when upload.storage.driver=cos")
		}
		if strings.TrimSpace(c.Storage.COS.SecretKey) == "" {
			return fmt.Errorf("upload.storage.cos.secret_key is required when upload.storage.driver=cos")
		}
	case "qiniu":
		if strings.TrimSpace(c.Storage.Qiniu.Bucket) == "" {
			return fmt.Errorf("upload.storage.qiniu.bucket is required when upload.storage.driver=qiniu")
		}
		if strings.TrimSpace(c.Storage.Qiniu.AccessKeyID) == "" {
			return fmt.Errorf("upload.storage.qiniu.access_key_id is required when upload.storage.driver=qiniu")
		}
		if strings.TrimSpace(c.Storage.Qiniu.AccessKeySecret) == "" {
			return fmt.Errorf("upload.storage.qiniu.access_key_secret is required when upload.storage.driver=qiniu")
		}
		if strings.TrimSpace(c.Storage.Qiniu.PublicBaseURL) == "" {
			return fmt.Errorf("upload.storage.qiniu.public_base_url is required when upload.storage.driver=qiniu")
		}
	default:
		return fmt.Errorf("upload.storage.driver must be local, s3-compatible, oss, cos, qiniu, or minio")
	}
	return nil
}

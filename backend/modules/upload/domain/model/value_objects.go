package model

import "time"

type FileBinding struct {
	BizModule string `json:"biz_module,omitempty"`
	BizType   string `json:"biz_type,omitempty"`
	BizId     string `json:"biz_id,omitempty"`
	BizField  string `json:"biz_field,omitempty"`
}

type StoragePolicy struct {
	MaxUploadSize     string         `json:"max_upload_size,omitempty"`
	AllowedExts       []string       `json:"allowed_exts,omitempty"`
	AllowedMIMETypes  []string       `json:"allowed_mime_types,omitempty"`
	DefaultVisibility FileVisibility `json:"default_visibility,omitempty"`
	PathPrefix        string         `json:"path_prefix,omitempty"`
}

type StorageLocation struct {
	Driver    string `json:"driver,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	Key       string `json:"key,omitempty"`
	Path      string `json:"path,omitempty"`
	PublicURL string `json:"public_url,omitempty"`
}

type FileAccessLink struct {
	URL        string         `json:"url,omitempty"`
	ExpiresAt  time.Time      `json:"expires_at,omitempty"`
	Visibility FileVisibility `json:"visibility,omitempty"`
}

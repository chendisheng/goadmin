package response

import "time"

type FileItem struct {
	Id             string    `json:"id,omitempty"`
	TenantId       string    `json:"tenant_id,omitempty"`
	OriginalName   string    `json:"original_name,omitempty"`
	StorageName    string    `json:"storage_name,omitempty"`
	StorageKey     string    `json:"storage_key,omitempty"`
	StorageDriver  string    `json:"storage_driver,omitempty"`
	StoragePath    string    `json:"storage_path,omitempty"`
	PublicURL      string    `json:"public_url,omitempty"`
	MimeType       string    `json:"mime_type,omitempty"`
	Extension      string    `json:"extension,omitempty"`
	SizeBytes      int64     `json:"size_bytes,omitempty"`
	ChecksumSHA256 string    `json:"checksum_sha256,omitempty"`
	Visibility     string    `json:"visibility,omitempty"`
	BizModule      string    `json:"biz_module,omitempty"`
	BizType        string    `json:"biz_type,omitempty"`
	BizId          string    `json:"biz_id,omitempty"`
	BizField       string    `json:"biz_field,omitempty"`
	UploadedBy     string    `json:"uploaded_by,omitempty"`
	Status         string    `json:"status,omitempty"`
	Remark         string    `json:"remark,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type List struct {
	Total int64      `json:"total"`
	Items []FileItem `json:"items"`
}

type StorageSetting struct {
	Driver string `json:"driver"`
}

type Preview struct {
	FileItem
	PreviewKind      string `json:"preview_kind,omitempty"`
	PreviewMode      string `json:"preview_mode,omitempty"`
	DownloadURL      string `json:"download_url,omitempty"`
	CanPreview       bool   `json:"can_preview,omitempty"`
	CanOpenInBrowser bool   `json:"can_open_in_browser,omitempty"`
}

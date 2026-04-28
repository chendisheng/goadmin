package model

import "time"

type FileVisibility string

const (
	FileVisibilityPrivate FileVisibility = "private"
	FileVisibilityPublic  FileVisibility = "public"
)

type FileStatus string

const (
	FileStatusActive   FileStatus = "active"
	FileStatusArchived FileStatus = "archived"
	FileStatusDeleted  FileStatus = "deleted"
)

type FileAsset struct {
	Id             string         `json:"id,omitempty" gorm:"column:id;primaryKey;type:varchar(64);size:64;comment:主键ID"`
	TenantId       string         `json:"tenant_id,omitempty" gorm:"column:tenant_id;type:varchar(64);size:64;comment:租户ID"`
	OriginalName   string         `json:"original_name,omitempty" gorm:"column:original_name;type:varchar(255);size:255;comment:原始文件名"`
	StorageName    string         `json:"storage_name,omitempty" gorm:"column:storage_name;type:varchar(255);size:255;comment:存储文件名"`
	StorageKey     string         `json:"storage_key,omitempty" gorm:"column:storage_key;type:varchar(255);size:255;comment:存储键"`
	StorageDriver  string         `json:"storage_driver,omitempty" gorm:"column:storage_driver;type:varchar(64);size:64;comment:存储驱动|local=本地,s3-compatible=S3兼容,oss=阿里云OSS,cos=腾讯云COS,minio=MinIO"`
	StoragePath    string         `json:"storage_path,omitempty" gorm:"column:storage_path;type:varchar(255);size:255;comment:物理存储路径"`
	PublicURL      string         `json:"public_url,omitempty" gorm:"column:public_url;type:varchar(512);size:512;comment:公开访问地址"`
	MimeType       string         `json:"mime_type,omitempty" gorm:"column:mime_type;type:varchar(128);size:128;comment:文件MIME类型"`
	Extension      string         `json:"extension,omitempty" gorm:"column:extension;type:varchar(32);size:32;comment:文件扩展名"`
	SizeBytes      int64          `json:"size_bytes,omitempty" gorm:"column:size_bytes;comment:文件大小(字节)"`
	ChecksumSHA256 string         `json:"checksum_sha256,omitempty" gorm:"column:checksum_sha256;type:varchar(128);size:128;comment:SHA256校验值"`
	Visibility     FileVisibility `json:"visibility,omitempty" gorm:"column:visibility;type:varchar(32);size:32;comment:可见性|private=私有,public=公开"`
	BizModule      string         `json:"biz_module,omitempty" gorm:"column:biz_module;type:varchar(64);size:64;comment:业务模块"`
	BizType        string         `json:"biz_type,omitempty" gorm:"column:biz_type;type:varchar(64);size:64;comment:业务类型"`
	BizId          string         `json:"biz_id,omitempty" gorm:"column:biz_id;type:varchar(64);size:64;comment:业务主键"`
	BizField       string         `json:"biz_field,omitempty" gorm:"column:biz_field;type:varchar(64);size:64;comment:业务字段"`
	UploadedBy     string         `json:"uploaded_by,omitempty" gorm:"column:uploaded_by;type:varchar(64);size:64;comment:上传人ID"`
	Status         FileStatus     `json:"status,omitempty" gorm:"column:status;type:varchar(32);size:32;comment:文件状态|active=有效,archived=已归档,deleted=已删除"`
	Remark         string         `json:"remark,omitempty" gorm:"column:remark;type:varchar(255);size:255;comment:备注"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      *time.Time     `json:"deleted_at,omitempty" gorm:"column:deleted_at;index;comment:删除时间"`
}

func (FileAsset) TableName() string {
	return "upload_file"
}

func (m FileAsset) Clone() FileAsset {
	clone := m
	return clone
}

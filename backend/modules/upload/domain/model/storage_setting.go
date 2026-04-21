package model

import "time"

type StorageSetting struct {
	Id           string    `json:"id,omitempty" gorm:"column:id;primaryKey;type:varchar(64);size:64;comment:主键ID"`
	SettingKey   string    `json:"setting_key,omitempty" gorm:"column:setting_key;type:varchar(64);size:64;uniqueIndex;comment:配置键"`
	SettingValue string    `json:"setting_value,omitempty" gorm:"column:setting_value;type:varchar(255);size:255;comment:配置值"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (StorageSetting) TableName() string {
	return "upload_storage_setting"
}

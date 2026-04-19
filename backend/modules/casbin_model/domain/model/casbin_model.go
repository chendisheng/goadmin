package model

import "time"

type CasbinModel struct {
	Name      string    `json:"name,omitempty" gorm:"column:name;primaryKey;type:varchar(64);size:64"`
	Content   string    `json:"content,omitempty" gorm:"column:content;type:longtext;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CasbinModel) TableName() string {
	return "casbin_model"
}

func (m CasbinModel) Clone() CasbinModel {
	clone := m
	return clone
}

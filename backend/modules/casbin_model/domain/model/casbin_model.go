package model

import "time"

type CasbinModel struct {
	Name      string    `json:"name,omitempty" gorm:"column:name;primaryKey;type:varchar(64);size:64"`
	Content   string    `json:"content,omitempty" gorm:"column:content;type:varchar(255);size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m CasbinModel) Clone() CasbinModel {
	clone := m
	return clone
}

package model

import (
	"time"
)

type CodegenConsole struct {
	Id        string    `json:"id,omitempty" gorm:"column:id;primaryKey;type:varchar(64);size:64"`
	Name      string    `json:"name,omitempty" gorm:"column:name;type:varchar(255);size:255"`
	Enabled   int64     `json:"enabled,omitempty" gorm:"column:enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m CodegenConsole) Clone() CodegenConsole {
	clone := m
	return clone
}

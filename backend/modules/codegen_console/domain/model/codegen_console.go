package model

import "time"

type CodegenConsole struct {
	ID        string    `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Name      string    `json:"name,omitempty" gorm:"column:name;index"`
	Enabled   bool      `json:"enabled,omitempty" gorm:"column:enabled;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (m CodegenConsole) Clone() CodegenConsole {
	clone := m
	return clone
}

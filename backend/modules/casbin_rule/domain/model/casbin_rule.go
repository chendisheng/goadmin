package model

import "time"

type CasbinRule struct {
	Id        int64     `json:"id,omitempty" gorm:"column:id;primaryKey;autoIncrement"`
	Ptype     string    `json:"ptype,omitempty" gorm:"column:ptype;type:varchar(32);size:32"`
	V0        string    `json:"v0,omitempty" gorm:"column:v0;type:varchar(191);size:191"`
	V1        string    `json:"v1,omitempty" gorm:"column:v1;type:varchar(191);size:191"`
	V2        string    `json:"v2,omitempty" gorm:"column:v2;type:varchar(191);size:191"`
	V3        string    `json:"v3,omitempty" gorm:"column:v3;type:varchar(255);size:255"`
	V4        string    `json:"v4,omitempty" gorm:"column:v4;type:varchar(255);size:255"`
	V5        string    `json:"v5,omitempty" gorm:"column:v5;type:varchar(255);size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CasbinRule) TableName() string {
	return "casbin_rule"
}

func (m CasbinRule) Clone() CasbinRule {
	clone := m
	return clone
}

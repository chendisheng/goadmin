package model

import (
	"time"

	pluginiface "goadmin/plugin/interface"
)

type Plugin struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Enabled     bool                     `json:"enabled"`
	Menus       []pluginiface.Menu       `json:"menus,omitempty"`
	Permissions []pluginiface.Permission `json:"permissions,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

func (Plugin) TableName() string {
	return "plugin"
}

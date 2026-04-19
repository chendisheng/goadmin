package model

import "time"

type Status string

const (
	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
)

type Type string

const (
	TypeDirectory Type = "directory"
	TypeMenu      Type = "menu"
	TypeButton    Type = "button"
)

type Menu struct {
	ID          string    `json:"id"`
	ParentID    string    `json:"parent_id,omitempty"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Component   string    `json:"component,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	Sort        int       `json:"sort"`
	Permission  string    `json:"permission,omitempty"`
	Type        Type      `json:"type"`
	Visible     bool      `json:"visible"`
	Enabled     bool      `json:"enabled"`
	Redirect    string    `json:"redirect,omitempty"`
	ExternalURL string    `json:"external_url,omitempty"`
	Children    []Menu    `json:"children,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Menu) TableName() string {
	return "menu"
}

func (m Menu) Clone() Menu {
	clone := m
	if m.Children != nil {
		clone.Children = make([]Menu, 0, len(m.Children))
		for _, child := range m.Children {
			clone.Children = append(clone.Children, child.Clone())
		}
	}
	return clone
}

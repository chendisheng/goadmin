package response

import "time"

type Menu struct {
	Plugin      string    `json:"plugin,omitempty"`
	ID          string    `json:"id"`
	ParentID    string    `json:"parent_id,omitempty"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Component   string    `json:"component,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	Sort        int       `json:"sort"`
	Permission  string    `json:"permission,omitempty"`
	Type        string    `json:"type"`
	Visible     bool      `json:"visible"`
	Enabled     bool      `json:"enabled"`
	Redirect    string    `json:"redirect,omitempty"`
	ExternalURL string    `json:"external_url,omitempty"`
	Children    []Menu    `json:"children,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Permission struct {
	Plugin      string `json:"plugin,omitempty"`
	Object      string `json:"object"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

type Item struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Enabled     bool         `json:"enabled"`
	Menus       []Menu       `json:"menus,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type List struct {
	Total int64  `json:"total"`
	Items []Item `json:"items"`
}

type MenuList struct {
	Items []Menu `json:"items"`
}

type PermissionList struct {
	Items []Permission `json:"items"`
}

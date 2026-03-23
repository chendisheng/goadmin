package response

import "time"

type Item struct {
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
	Children    []Item    `json:"children,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type List struct {
	Total int64  `json:"total"`
	Items []Item `json:"items"`
}

type Tree struct {
	Items []Item `json:"items"`
}

type RouteMeta struct {
	Title      string `json:"title"`
	Icon       string `json:"icon,omitempty"`
	Permission string `json:"permission,omitempty"`
	Hidden     bool   `json:"hidden"`
	NoCache    bool   `json:"noCache"`
	Affix      bool   `json:"affix"`
	Link       string `json:"link,omitempty"`
}

type Route struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Component  string    `json:"component,omitempty"`
	Redirect   string    `json:"redirect,omitempty"`
	Hidden     bool      `json:"hidden"`
	AlwaysShow bool      `json:"alwaysShow,omitempty"`
	Meta       RouteMeta `json:"meta"`
	Children   []Route   `json:"children,omitempty"`
}

type Routes struct {
	Items []Route `json:"items"`
}

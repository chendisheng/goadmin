package request

type Permission struct {
	Object      string `json:"object"`
	Action      string `json:"action"`
	Description string `json:"description"`
}

type Menu struct {
	Plugin       string `json:"plugin,omitempty"`
	ID           string `json:"id,omitempty"`
	ParentID     string `json:"parent_id,omitempty"`
	Name         string `json:"name"`
	TitleKey     string `json:"titleKey,omitempty"`
	TitleDefault string `json:"titleDefault,omitempty"`
	Path         string `json:"path"`
	Component    string `json:"component,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Sort         int    `json:"sort"`
	Permission   string `json:"permission,omitempty"`
	Type         string `json:"type,omitempty"`
	Visible      bool   `json:"visible"`
	Enabled      bool   `json:"enabled"`
	Redirect     string `json:"redirect,omitempty"`
	ExternalURL  string `json:"external_url,omitempty"`
}

type CreateRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Enabled     bool         `json:"enabled"`
	Menus       []Menu       `json:"menus,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
}

type UpdateRequest struct {
	Description *string      `json:"description,omitempty"`
	Enabled     *bool        `json:"enabled,omitempty"`
	Menus       []Menu       `json:"menus,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
}

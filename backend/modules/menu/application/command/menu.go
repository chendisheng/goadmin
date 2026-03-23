package command

type CreateMenu struct {
	ParentID    string `json:"parent_id,omitempty"`
	Name        string `json:"name" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Component   string `json:"component,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Sort        int    `json:"sort"`
	Permission  string `json:"permission,omitempty"`
	Type        string `json:"type,omitempty"`
	Visible     bool   `json:"visible"`
	Enabled     bool   `json:"enabled"`
	Redirect    string `json:"redirect,omitempty"`
	ExternalURL string `json:"external_url,omitempty"`
}

type UpdateMenu struct {
	ParentID    string `json:"parent_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Path        string `json:"path,omitempty"`
	Component   string `json:"component,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Sort        int    `json:"sort"`
	Permission  string `json:"permission,omitempty"`
	Type        string `json:"type,omitempty"`
	Visible     bool   `json:"visible"`
	Enabled     bool   `json:"enabled"`
	Redirect    string `json:"redirect,omitempty"`
	ExternalURL string `json:"external_url,omitempty"`
}

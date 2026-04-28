package request

type ListRequest struct {
	Keyword  string `form:"keyword"`
	ParentID string `form:"parent_id"`
	Visible  *bool  `form:"visible"`
	Enabled  *bool  `form:"enabled"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type CreateRequest struct {
	ParentID     string `json:"parent_id"`
	Name         string `json:"name" binding:"required"`
	TitleKey     string `json:"title_key"`
	TitleDefault string `json:"title_default"`
	Path         string `json:"path" binding:"required"`
	Component    string `json:"component"`
	Icon         string `json:"icon"`
	Sort         int    `json:"sort"`
	Permission   string `json:"permission"`
	Type         string `json:"type"`
	Visible      bool   `json:"visible"`
	Enabled      bool   `json:"enabled"`
	Redirect     string `json:"redirect"`
	ExternalURL  string `json:"external_url"`
}

type UpdateRequest struct {
	ParentID     string `json:"parent_id"`
	Name         string `json:"name"`
	TitleKey     string `json:"title_key"`
	TitleDefault string `json:"title_default"`
	Path         string `json:"path"`
	Component    string `json:"component"`
	Icon         string `json:"icon"`
	Sort         int    `json:"sort"`
	Permission   string `json:"permission"`
	Type         string `json:"type"`
	Visible      bool   `json:"visible"`
	Enabled      bool   `json:"enabled"`
	Redirect     string `json:"redirect"`
	ExternalURL  string `json:"external_url"`
}

package request

type ListRequest struct {
	Keyword  string `json:"keyword,omitempty" form:"keyword"`
	Page     int    `json:"page,omitempty" form:"page"`
	PageSize int    `json:"page_size,omitempty" form:"page_size"`
}

type CreateRequest struct {
	Name    string `json:"name,omitempty" form:"name"`
	Enabled bool   `json:"enabled,omitempty" form:"enabled"`
}

type UpdateRequest struct {
	Name    string `json:"name,omitempty" form:"name"`
	Enabled bool   `json:"enabled,omitempty" form:"enabled"`
}

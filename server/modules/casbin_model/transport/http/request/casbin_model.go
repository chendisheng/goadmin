package request

type ListRequest struct {
	Keyword  string `json:"keyword,omitempty" form:"keyword"`
	Page     int    `json:"page,omitempty" form:"page"`
	PageSize int    `json:"page_size,omitempty" form:"page_size"`
}

type CreateRequest struct {
	Content string `json:"content,omitempty" form:"content"`
}

type UpdateRequest struct {
	Content string `json:"content,omitempty" form:"content"`
}

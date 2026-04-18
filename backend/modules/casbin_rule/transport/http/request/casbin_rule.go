package request

type ListRequest struct {
	Keyword  string `json:"keyword,omitempty" form:"keyword"`
	Page     int    `json:"page,omitempty" form:"page"`
	PageSize int    `json:"page_size,omitempty" form:"page_size"`
}

type CreateRequest struct {
	Ptype string `json:"ptype,omitempty" form:"ptype"`
	V0    string `json:"v0,omitempty" form:"v0"`
	V1    string `json:"v1,omitempty" form:"v1"`
	V2    string `json:"v2,omitempty" form:"v2"`
	V3    string `json:"v3,omitempty" form:"v3"`
	V4    string `json:"v4,omitempty" form:"v4"`
	V5    string `json:"v5,omitempty" form:"v5"`
}

type UpdateRequest struct {
	Ptype string `json:"ptype,omitempty" form:"ptype"`
	V0    string `json:"v0,omitempty" form:"v0"`
	V1    string `json:"v1,omitempty" form:"v1"`
	V2    string `json:"v2,omitempty" form:"v2"`
	V3    string `json:"v3,omitempty" form:"v3"`
	V4    string `json:"v4,omitempty" form:"v4"`
	V5    string `json:"v5,omitempty" form:"v5"`
}

package response

type Item struct {
	Id    int64  `json:"id,omitempty"`
	Ptype string `json:"ptype,omitempty"`
	V0    string `json:"v0,omitempty"`
	V1    string `json:"v1,omitempty"`
	V2    string `json:"v2,omitempty"`
	V3    string `json:"v3,omitempty"`
	V4    string `json:"v4,omitempty"`
	V5    string `json:"v5,omitempty"`
}

type List struct {
	Total int64  `json:"total"`
	Items []Item `json:"items"`
}

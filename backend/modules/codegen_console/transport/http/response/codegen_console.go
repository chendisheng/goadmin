package response

type Item struct {
	Id      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Enabled int64  `json:"enabled,omitempty"`
}

type List struct {
	Total int64  `json:"total"`
	Items []Item `json:"items"`
}

package response

type Item struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

type List struct {
	Total int64  `json:"total"`
	Items []Item `json:"items"`
}

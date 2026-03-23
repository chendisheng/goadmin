package query

type ListMenus struct {
	Keyword  string `form:"keyword"`
	ParentID string `form:"parent_id"`
	Visible  *bool  `form:"visible"`
	Enabled  *bool  `form:"enabled"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

package query

type ListUsers struct {
	TenantID string `form:"tenant_id"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

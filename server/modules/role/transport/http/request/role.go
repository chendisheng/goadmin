package request

type ListRequest struct {
	TenantID string `form:"tenant_id"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type CreateRequest struct {
	TenantID string   `json:"tenant_id"`
	Name     string   `json:"name" binding:"required"`
	Code     string   `json:"code" binding:"required"`
	Status   string   `json:"status"`
	Remark   string   `json:"remark"`
	MenuIDs  []string `json:"menu_ids"`
}

type UpdateRequest struct {
	TenantID string   `json:"tenant_id"`
	Name     string   `json:"name"`
	Code     string   `json:"code"`
	Status   string   `json:"status"`
	Remark   string   `json:"remark"`
	MenuIDs  []string `json:"menu_ids"`
}

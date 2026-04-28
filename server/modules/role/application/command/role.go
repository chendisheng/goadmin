package command

type CreateRole struct {
	TenantID string   `json:"tenant_id,omitempty"`
	Name     string   `json:"name" binding:"required"`
	Code     string   `json:"code" binding:"required"`
	Status   string   `json:"status,omitempty"`
	Remark   string   `json:"remark,omitempty"`
	MenuIDs  []string `json:"menu_ids,omitempty"`
}

type UpdateRole struct {
	TenantID string   `json:"tenant_id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Code     string   `json:"code,omitempty"`
	Status   string   `json:"status,omitempty"`
	Remark   string   `json:"remark,omitempty"`
	MenuIDs  []string `json:"menu_ids,omitempty"`
}

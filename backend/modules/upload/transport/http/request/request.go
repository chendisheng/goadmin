package request

type ListRequest struct {
	Keyword    string `form:"keyword"`
	Visibility string `form:"visibility"`
	Status     string `form:"status"`
	BizModule  string `form:"biz_module"`
	BizType    string `form:"biz_type"`
	BizId      string `form:"biz_id"`
	UploadedBy string `form:"uploaded_by"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

type UploadRequest struct {
	Visibility string `form:"visibility" json:"visibility"`
	BizModule  string `form:"biz_module" json:"biz_module"`
	BizType    string `form:"biz_type" json:"biz_type"`
	BizId      string `form:"biz_id" json:"biz_id"`
	BizField   string `form:"biz_field" json:"biz_field"`
	Remark     string `form:"remark" json:"remark"`
}

type BindRequest struct {
	BizModule string `json:"biz_module"`
	BizType   string `json:"biz_type"`
	BizId     string `json:"biz_id"`
	BizField  string `json:"biz_field"`
}

type StorageSettingRequest struct {
	Driver string `json:"driver" form:"driver"`
}

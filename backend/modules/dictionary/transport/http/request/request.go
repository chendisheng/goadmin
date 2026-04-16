package request

type CategoryListRequest struct {
	Keyword  string `json:"keyword,omitempty" form:"keyword"`
	Status   string `json:"status,omitempty" form:"status"`
	Page     int    `json:"page,omitempty" form:"page"`
	PageSize int    `json:"page_size,omitempty" form:"page_size"`
}

type CategoryCreateRequest struct {
	ID          string `json:"id,omitempty" form:"id"`
	Code        string `json:"code,omitempty" form:"code"`
	Name        string `json:"name,omitempty" form:"name"`
	Description string `json:"description,omitempty" form:"description"`
	Status      string `json:"status,omitempty" form:"status"`
	Sort        int    `json:"sort,omitempty" form:"sort"`
	Remark      string `json:"remark,omitempty" form:"remark"`
}

type CategoryUpdateRequest struct {
	Code        string `json:"code,omitempty" form:"code"`
	Name        string `json:"name,omitempty" form:"name"`
	Description string `json:"description,omitempty" form:"description"`
	Status      string `json:"status,omitempty" form:"status"`
	Sort        int    `json:"sort,omitempty" form:"sort"`
	Remark      string `json:"remark,omitempty" form:"remark"`
}

type ItemListRequest struct {
	CategoryID   string `json:"category_id,omitempty" form:"category_id"`
	CategoryCode string `json:"category_code,omitempty" form:"category_code"`
	Keyword      string `json:"keyword,omitempty" form:"keyword"`
	Status       string `json:"status,omitempty" form:"status"`
	Page         int    `json:"page,omitempty" form:"page"`
	PageSize     int    `json:"page_size,omitempty" form:"page_size"`
}

type ItemCreateRequest struct {
	ID         string `json:"id,omitempty" form:"id"`
	CategoryID string `json:"category_id,omitempty" form:"category_id"`
	Value      string `json:"value,omitempty" form:"value"`
	Label      string `json:"label,omitempty" form:"label"`
	TagType    string `json:"tag_type,omitempty" form:"tag_type"`
	TagColor   string `json:"tag_color,omitempty" form:"tag_color"`
	Extra      string `json:"extra,omitempty" form:"extra"`
	IsDefault  bool   `json:"is_default,omitempty" form:"is_default"`
	Status     string `json:"status,omitempty" form:"status"`
	Sort       int    `json:"sort,omitempty" form:"sort"`
	Remark     string `json:"remark,omitempty" form:"remark"`
}

type ItemUpdateRequest struct {
	CategoryID string `json:"category_id,omitempty" form:"category_id"`
	Value      string `json:"value,omitempty" form:"value"`
	Label      string `json:"label,omitempty" form:"label"`
	TagType    string `json:"tag_type,omitempty" form:"tag_type"`
	TagColor   string `json:"tag_color,omitempty" form:"tag_color"`
	Extra      string `json:"extra,omitempty" form:"extra"`
	IsDefault  bool   `json:"is_default,omitempty" form:"is_default"`
	Status     string `json:"status,omitempty" form:"status"`
	Sort       int    `json:"sort,omitempty" form:"sort"`
	Remark     string `json:"remark,omitempty" form:"remark"`
}

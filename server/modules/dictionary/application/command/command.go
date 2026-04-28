package command

type CreateCategory struct {
	ID          string
	Code        string
	Name        string
	Description string
	Status      string
	Sort        int
	Remark      string
}

type UpdateCategory struct {
	ID          string
	Code        string
	Name        string
	Description string
	Status      string
	Sort        int
	Remark      string
}

type CreateItem struct {
	ID         string
	CategoryID string
	Value      string
	Label      string
	TagType    string
	TagColor   string
	Extra      string
	IsDefault  bool
	Status     string
	Sort       int
	Remark     string
}

type UpdateItem struct {
	ID         string
	CategoryID string
	Value      string
	Label      string
	TagType    string
	TagColor   string
	Extra      string
	IsDefault  bool
	Status     string
	Sort       int
	Remark     string
}

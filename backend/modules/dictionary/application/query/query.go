package query

type ListCategories struct {
	Keyword  string
	Status   string
	Page     int
	PageSize int
}

type ListItems struct {
	CategoryID   string
	CategoryCode string
	Keyword      string
	Status       string
	Page         int
	PageSize     int
}

type LookupItems struct {
	CategoryCode string
}

type LookupItem struct {
	CategoryCode string
	Value        string
}

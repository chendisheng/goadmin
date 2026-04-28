package model

type Field struct {
	Name    string
	Type    string
	Primary bool
	Index   bool
	Unique  bool
}

type Resource struct {
	Kind             string
	Name             string
	Fields           []Field
	GenerateFrontend bool
	GeneratePolicy   bool
	Force            bool
}

type Change struct {
	Path    string
	Action  string
	Content []byte
}

type Plan struct {
	Resources []Resource
	Changes   []Change
	Messages  []string
}

type IR struct {
	Source string
	Plan   Plan
}

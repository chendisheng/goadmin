package gin

import "goadmin/codegen/model"

type Generator struct{}

func New() Generator {
	return Generator{}
}

func (Generator) Framework() string {
	return "gin"
}

func (Generator) Describe(plan model.Plan) string {
	return "gin backend generator"
}

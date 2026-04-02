package vue3

import "goadmin/codegen/model"

type Generator struct{}

func New() Generator {
	return Generator{}
}

func (Generator) Framework() string {
	return "vue3"
}

func (Generator) Describe(plan model.Plan) string {
	return "vue3 frontend generator"
}

package policy

import "goadmin/codegen/model"

type Generator struct{}

func New() Generator {
	return Generator{}
}

func (Generator) Describe(plan model.Plan) string {
	return "policy generator"
}

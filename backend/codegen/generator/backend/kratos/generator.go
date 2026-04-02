package kratos

import "goadmin/codegen/model"

type Generator struct{}

func New() Generator {
	return Generator{}
}

func (Generator) Framework() string {
	return "kratos"
}

func (Generator) Describe(plan model.Plan) string {
	return "kratos backend generator"
}

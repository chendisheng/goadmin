package react

import "goadmin/codegen/model"

type Generator struct{}

func New() Generator {
	return Generator{}
}

func (Generator) Framework() string {
	return "react"
}

func (Generator) Describe(plan model.Plan) string {
	return "react frontend generator"
}

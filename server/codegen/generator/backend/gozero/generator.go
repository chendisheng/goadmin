package gozero

import "goadmin/codegen/model"

type Generator struct{}

func New() Generator {
	return Generator{}
}

func (Generator) Framework() string {
	return "go-zero"
}

func (Generator) Describe(plan model.Plan) string {
	return "go-zero server generator"
}

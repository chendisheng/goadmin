package planner

import (
	"fmt"
	"strings"

	"goadmin/codegen/model"
	"goadmin/codegen/schema"
)

type Planner interface {
	Plan(document schema.Document) (model.Plan, error)
}

type Default struct{}

func New() Default {
	return Default{}
}

func (Default) Plan(document schema.Document) (model.Plan, error) {
	resources, err := document.ResolveResources()
	if err != nil {
		return model.Plan{}, err
	}
	plan := model.Plan{
		Resources: make([]model.Resource, 0, len(resources)),
		Messages:  []string{"planned by default codegen planner"},
	}
	for _, resource := range resources {
		planned, err := planResource(resource)
		if err != nil {
			return model.Plan{}, err
		}
		plan.Resources = append(plan.Resources, planned)
	}
	return plan, nil
}

func planResource(resource schema.Resource) (model.Resource, error) {
	kind := strings.TrimSpace(string(resource.Kind))
	name := strings.TrimSpace(resource.Name)
	if kind == "" {
		return model.Resource{}, fmt.Errorf("resource kind is required")
	}
	if name == "" {
		return model.Resource{}, fmt.Errorf("resource name is required")
	}
	planned := model.Resource{
		Kind:             kind,
		Name:             name,
		GenerateFrontend: resource.GenerateFrontend,
		GeneratePolicy:   resource.GeneratePolicy,
		Force:            resource.Force,
	}
	if len(resource.Fields) > 0 {
		planned.Fields = make([]model.Field, 0, len(resource.Fields))
		for _, field := range resource.Fields {
			planned.Fields = append(planned.Fields, model.Field{
				Name:    strings.TrimSpace(field.Name),
				Type:    strings.TrimSpace(field.Type),
				Primary: field.Primary,
				Index:   field.Index,
				Unique:  field.Unique,
			})
		}
	}
	return planned, nil
}

type Conflict struct {
	Path   string
	Reason string
}

type Scope struct {
	Resources []string
	Paths     []string
}

func (s Scope) Contains(value string) bool {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return false
	}
	for _, candidate := range s.Resources {
		if strings.EqualFold(strings.TrimSpace(candidate), clean) {
			return true
		}
	}
	for _, candidate := range s.Paths {
		if strings.EqualFold(strings.TrimSpace(candidate), clean) {
			return true
		}
	}
	return false
}

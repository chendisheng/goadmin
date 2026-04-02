package schema

import (
	"fmt"
	"strings"
)

func (d Document) ResolveResources() ([]Resource, error) {
	if len(d.Resources) > 0 {
		resources := make([]Resource, 0, len(d.Resources))
		for i, resource := range d.Resources {
			normalized := resource.normalizedFromDocument(d)
			if err := normalized.Validate(); err != nil {
				return nil, fmt.Errorf("resource[%d]: %w", i, err)
			}
			resources = append(resources, normalized)
		}
		return resources, nil
	}

	resource := Resource{
		Kind:             d.inferKind(),
		Name:             d.inferName(),
		Module:           strings.TrimSpace(d.Module),
		Framework:        d.Framework,
		Entity:           d.Entity,
		Fields:           append([]Field(nil), d.Entity.Fields...),
		Pages:            append([]Page(nil), d.Pages...),
		Permissions:      append([]Permission(nil), d.Permissions...),
		Routes:           append([]Route(nil), d.Routes...),
		Plugin:           d.Plugin,
		GenerateFrontend: d.Framework.Frontend != "" || len(d.Pages) > 0,
		GeneratePolicy:   len(d.Permissions) > 0,
	}
	if resource.Kind == "" {
		return nil, fmt.Errorf("codegen document requires a kind or at least one resource")
	}
	if resource.Name == "" {
		return nil, fmt.Errorf("codegen document requires a module, entity name, or resource name")
	}
	if err := resource.Validate(); err != nil {
		return nil, err
	}
	return []Resource{resource}, nil
}

func (d Document) inferKind() Kind {
	if strings.TrimSpace(string(d.Kind)) != "" {
		return d.Kind
	}
	if d.Plugin != nil && strings.TrimSpace(d.Plugin.Name) != "" {
		return KindPlugin
	}
	if strings.TrimSpace(d.Entity.Name) != "" && len(d.Entity.Fields) > 0 {
		if strings.TrimSpace(d.Framework.Frontend) != "" || len(d.Pages) > 0 {
			return KindBusinessModule
		}
		return KindCRUD
	}
	if len(d.Pages) > 0 {
		return KindFrontendPage
	}
	if len(d.Permissions) > 0 {
		return KindPolicy
	}
	if strings.TrimSpace(d.Module) != "" {
		return KindModule
	}
	return ""
}

func (d Document) inferName() string {
	if strings.TrimSpace(d.Entity.Name) != "" {
		return strings.TrimSpace(d.Entity.Name)
	}
	if strings.TrimSpace(d.Module) != "" {
		return strings.TrimSpace(d.Module)
	}
	if d.Plugin != nil && strings.TrimSpace(d.Plugin.Name) != "" {
		return strings.TrimSpace(d.Plugin.Name)
	}
	return ""
}

func (r Resource) normalizedFromDocument(d Document) Resource {
	normalized := r
	if strings.TrimSpace(normalized.Module) == "" {
		normalized.Module = strings.TrimSpace(d.Module)
	}
	if normalized.Framework == (Framework{}) {
		normalized.Framework = d.Framework
	}
	if normalized.Entity.Name == "" {
		normalized.Entity.Name = d.Entity.Name
	}
	if len(normalized.Entity.Fields) == 0 && len(normalized.Fields) == 0 && len(d.Entity.Fields) > 0 {
		normalized.Fields = append([]Field(nil), d.Entity.Fields...)
	}
	if len(normalized.Pages) == 0 && len(d.Pages) > 0 {
		normalized.Pages = append([]Page(nil), d.Pages...)
	}
	if len(normalized.Permissions) == 0 && len(d.Permissions) > 0 {
		normalized.Permissions = append([]Permission(nil), d.Permissions...)
	}
	if len(normalized.Routes) == 0 && len(d.Routes) > 0 {
		normalized.Routes = append([]Route(nil), d.Routes...)
	}
	if normalized.Plugin == nil && d.Plugin != nil {
		normalized.Plugin = d.Plugin
	}
	if strings.TrimSpace(string(normalized.Kind)) == "" {
		normalized.Kind = d.inferKind()
	}
	if strings.TrimSpace(normalized.Name) == "" {
		normalized.Name = d.inferName()
	}
	if !normalized.GenerateFrontend {
		normalized.GenerateFrontend = normalized.Framework.Frontend != "" || len(normalized.Pages) > 0
	}
	if !normalized.GeneratePolicy {
		normalized.GeneratePolicy = len(normalized.Permissions) > 0
	}
	return normalized
}

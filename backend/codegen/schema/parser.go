package schema

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Parser struct {
	Strict bool
}

func NewParser() Parser {
	return Parser{Strict: true}
}

func ParseYAML(data []byte) (Document, error) {
	return NewParser().ParseYAML(data)
}

func ParseYAMLFile(path string) (Document, error) {
	return NewParser().ParseYAMLFile(path)
}

func (p Parser) ParseYAMLFile(path string) (Document, error) {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return Document{}, fmt.Errorf("dsl path is required")
	}
	content, err := os.ReadFile(clean)
	if err != nil {
		return Document{}, fmt.Errorf("read dsl file %s: %w", clean, err)
	}
	doc, err := p.ParseYAML(content)
	if err != nil {
		return Document{}, err
	}
	doc.Version = normalizeVersion(doc.Version)
	return doc, nil
}

func (p Parser) ParseYAML(data []byte) (Document, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return Document{}, fmt.Errorf("dsl content is empty")
	}
	var payload rawDocument
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if p.Strict {
		decoder.KnownFields(true)
	}
	if err := decoder.Decode(&payload); err != nil {
		return Document{}, fmt.Errorf("decode dsl yaml: %w", err)
	}
	doc := payload.toDocument()
	if err := doc.Validate(); err != nil {
		return Document{}, err
	}
	return doc, nil
}

type rawDocument struct {
	Version     string          `yaml:"version,omitempty"`
	Module      string          `yaml:"module,omitempty"`
	Kind        Kind            `yaml:"kind,omitempty"`
	Framework   rawFramework    `yaml:"framework,omitempty"`
	Entity      rawEntity       `yaml:"entity,omitempty"`
	Pages       []rawPage       `yaml:"pages,omitempty"`
	Permissions []rawPermission `yaml:"permissions,omitempty"`
	Routes      []Route         `yaml:"routes,omitempty"`
	Plugin      *rawPlugin      `yaml:"plugin,omitempty"`
	Resources   []rawResource   `yaml:"resources,omitempty"`
}

type rawFramework struct {
	Backend  string `yaml:"backend,omitempty"`
	Frontend string `yaml:"frontend,omitempty"`
}

type rawEntity struct {
	Name   string     `yaml:"name,omitempty"`
	Fields []rawField `yaml:"fields,omitempty"`
}

type rawField struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type,omitempty"`
	Primary  bool   `yaml:"primary,omitempty"`
	Index    bool   `yaml:"index,omitempty"`
	Unique   bool   `yaml:"unique,omitempty"`
	Required bool   `yaml:"required,omitempty"`
}

type rawPage struct {
	Name      string `yaml:"name,omitempty"`
	Type      string `yaml:"type,omitempty"`
	Path      string `yaml:"path,omitempty"`
	Component string `yaml:"component,omitempty"`
}

type rawPermission struct {
	Name     string `yaml:"name,omitempty"`
	Action   string `yaml:"action,omitempty"`
	Resource string `yaml:"resource,omitempty"`
}

type rawPlugin struct {
	Name  string `yaml:"name,omitempty"`
	Route string `yaml:"route,omitempty"`
	View  string `yaml:"view,omitempty"`
}

type rawResource struct {
	Kind             Kind            `yaml:"kind"`
	Name             string          `yaml:"name"`
	Module           string          `yaml:"module,omitempty"`
	Framework        rawFramework    `yaml:"framework,omitempty"`
	Entity           rawEntity       `yaml:"entity,omitempty"`
	Fields           []rawField      `yaml:"fields,omitempty"`
	Pages            []rawPage       `yaml:"pages,omitempty"`
	Permissions      []rawPermission `yaml:"permissions,omitempty"`
	Routes           []Route         `yaml:"routes,omitempty"`
	Plugin           *rawPlugin      `yaml:"plugin,omitempty"`
	GenerateFrontend bool            `yaml:"generate_frontend,omitempty"`
	GeneratePolicy   bool            `yaml:"generate_policy,omitempty"`
	Force            bool            `yaml:"force,omitempty"`
}

func (p *rawPage) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.ScalarNode:
		p.Name = strings.TrimSpace(node.Value)
		return nil
	case yaml.MappingNode:
		type plain rawPage
		return node.Decode((*plain)(p))
	default:
		return fmt.Errorf("page must be a string or mapping")
	}
}

func (p *rawPermission) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.ScalarNode:
		p.Name = strings.TrimSpace(node.Value)
		return nil
	case yaml.MappingNode:
		type plain rawPermission
		return node.Decode((*plain)(p))
	default:
		return fmt.Errorf("permission must be a string or mapping")
	}
}

func (f *rawField) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.ScalarNode:
		f.Name = strings.TrimSpace(node.Value)
		f.Type = "string"
		return nil
	case yaml.MappingNode:
		type plain rawField
		return node.Decode((*plain)(f))
	default:
		return fmt.Errorf("field must be a string or mapping")
	}
}

func (r rawDocument) toDocument() Document {
	doc := Document{
		Version:     normalizeVersion(r.Version),
		Module:      strings.TrimSpace(r.Module),
		Kind:        r.Kind,
		Framework:   r.Framework.toFramework(),
		Entity:      r.Entity.toEntity(),
		Pages:       toPages(r.Pages),
		Permissions: toPermissions(r.Permissions),
		Routes:      append([]Route(nil), r.Routes...),
		Plugin:      r.Plugin.toPlugin(),
	}
	if len(r.Resources) > 0 {
		doc.Resources = make([]Resource, 0, len(r.Resources))
		for _, resource := range r.Resources {
			doc.Resources = append(doc.Resources, resource.toResource(doc))
		}
		return doc
	}

	if strings.TrimSpace(doc.Module) != "" || strings.TrimSpace(doc.Entity.Name) != "" || doc.Plugin != nil || len(doc.Pages) > 0 || len(doc.Permissions) > 0 || len(doc.Routes) > 0 {
		resource := Resource{
			Kind:             doc.inferKind(),
			Name:             doc.inferName(),
			Module:           doc.Module,
			Framework:        doc.Framework,
			Entity:           doc.Entity,
			Fields:           append([]Field(nil), doc.Entity.Fields...),
			Pages:            append([]Page(nil), doc.Pages...),
			Permissions:      append([]Permission(nil), doc.Permissions...),
			Routes:           append([]Route(nil), doc.Routes...),
			Plugin:           doc.Plugin,
			GenerateFrontend: doc.Framework.Frontend != "" || len(doc.Pages) > 0,
			GeneratePolicy:   len(doc.Permissions) > 0,
		}
		doc.Resources = []Resource{resource}
	}
	return doc
}

func (r rawResource) toResource(doc Document) Resource {
	resource := Resource{
		Kind:             r.Kind,
		Name:             strings.TrimSpace(r.Name),
		Module:           strings.TrimSpace(r.Module),
		Framework:        r.Framework.toFramework(),
		Entity:           r.Entity.toEntity(),
		Fields:           toFields(r.Fields),
		Pages:            toPages(r.Pages),
		Permissions:      toPermissions(r.Permissions),
		Routes:           append([]Route(nil), r.Routes...),
		Plugin:           r.Plugin.toPlugin(),
		GenerateFrontend: r.GenerateFrontend,
		GeneratePolicy:   r.GeneratePolicy,
		Force:            r.Force,
	}
	if strings.TrimSpace(resource.Module) == "" {
		resource.Module = doc.Module
	}
	if resource.Framework == (Framework{}) {
		resource.Framework = doc.Framework
	}
	if resource.Entity.Name == "" {
		resource.Entity.Name = doc.Entity.Name
	}
	if len(resource.Fields) == 0 {
		resource.Fields = append([]Field(nil), doc.Entity.Fields...)
	}
	if len(resource.Pages) == 0 {
		resource.Pages = append([]Page(nil), doc.Pages...)
	}
	if len(resource.Permissions) == 0 {
		resource.Permissions = append([]Permission(nil), doc.Permissions...)
	}
	if len(resource.Routes) == 0 {
		resource.Routes = append([]Route(nil), doc.Routes...)
	}
	if resource.Plugin == nil {
		resource.Plugin = doc.Plugin
	}
	if strings.TrimSpace(string(resource.Kind)) == "" {
		resource.Kind = doc.inferKind()
	}
	if strings.TrimSpace(resource.Name) == "" {
		resource.Name = doc.inferName()
	}
	if !resource.GenerateFrontend {
		resource.GenerateFrontend = resource.Framework.Frontend != "" || len(resource.Pages) > 0
	}
	if !resource.GeneratePolicy {
		resource.GeneratePolicy = len(resource.Permissions) > 0
	}
	return resource
}

func (f rawFramework) toFramework() Framework {
	return Framework{Backend: strings.TrimSpace(f.Backend), Frontend: strings.TrimSpace(f.Frontend)}
}

func (e rawEntity) toEntity() Entity {
	return Entity{Name: strings.TrimSpace(e.Name), Fields: toFields(e.Fields)}
}

func (p *rawPlugin) toPlugin() *Plugin {
	if p == nil {
		return nil
	}
	plugin := Plugin{Name: strings.TrimSpace(p.Name), Route: strings.TrimSpace(p.Route), View: strings.TrimSpace(p.View)}
	if plugin == (Plugin{}) {
		return nil
	}
	return &plugin
}

func toFields(items []rawField) []Field {
	if len(items) == 0 {
		return nil
	}
	fields := make([]Field, 0, len(items))
	for _, item := range items {
		field := Field{
			Name:     strings.TrimSpace(item.Name),
			Type:     strings.TrimSpace(item.Type),
			Primary:  item.Primary,
			Index:    item.Index,
			Unique:   item.Unique,
			Required: item.Required,
		}
		if field.Type == "" {
			field.Type = "string"
		}
		fields = append(fields, field)
	}
	return fields
}

func toPages(items []rawPage) []Page {
	if len(items) == 0 {
		return nil
	}
	pages := make([]Page, 0, len(items))
	for _, item := range items {
		pages = append(pages, Page{
			Name:      strings.TrimSpace(item.Name),
			Type:      strings.TrimSpace(item.Type),
			Path:      strings.TrimSpace(item.Path),
			Component: strings.TrimSpace(item.Component),
		})
	}
	return pages
}

func toPermissions(items []rawPermission) []Permission {
	if len(items) == 0 {
		return nil
	}
	permissions := make([]Permission, 0, len(items))
	for _, item := range items {
		permissions = append(permissions, Permission{
			Name:     strings.TrimSpace(item.Name),
			Action:   strings.TrimSpace(item.Action),
			Resource: strings.TrimSpace(item.Resource),
		})
	}
	return permissions
}

func normalizeVersion(version string) string {
	if strings.TrimSpace(version) == "" {
		return "v1"
	}
	return strings.TrimSpace(version)
}

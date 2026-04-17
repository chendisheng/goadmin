package schema

import (
	"fmt"
	"strings"
)

type Kind string

const (
	KindModule              Kind = "module"
	KindCRUD                Kind = "crud"
	KindPlugin              Kind = "plugin"
	KindBusinessModule      Kind = "business-module"
	KindBackendModule       Kind = "backend-module"
	KindBackendCRUD         Kind = "backend-crud"
	KindBackendPlugin       Kind = "backend-plugin"
	KindFrontendPage        Kind = "frontend-page"
	KindFrontendModuleRoute Kind = "frontend-module-route"
	KindPolicy              Kind = "policy"
	KindManifest            Kind = "manifest"
	KindConfig              Kind = "config"
)

type Field struct {
	Name     string     `yaml:"name" json:"name"`
	Type     string     `yaml:"type" json:"type"`
	Comment  string     `yaml:"comment,omitempty" json:"comment,omitempty"`
	UIType   string     `yaml:"ui_type,omitempty" json:"ui_type,omitempty"`
	Enum     *EnumField `yaml:"enum,omitempty" json:"enum,omitempty"`
	Primary  bool       `yaml:"primary,omitempty" json:"primary,omitempty"`
	Index    bool       `yaml:"index,omitempty" json:"index,omitempty"`
	Unique   bool       `yaml:"unique,omitempty" json:"unique,omitempty"`
	Required bool       `yaml:"required,omitempty" json:"required,omitempty"`
}

type EnumField struct {
	Kind       string         `yaml:"kind,omitempty" json:"kind,omitempty"`
	Mode       string         `yaml:"mode,omitempty" json:"mode,omitempty"`
	Display    string         `yaml:"display,omitempty" json:"display,omitempty"`
	SourceRef  string         `yaml:"source_ref,omitempty" json:"source_ref,omitempty"`
	Confidence string         `yaml:"confidence,omitempty" json:"confidence,omitempty"`
	Fallback   string         `yaml:"fallback,omitempty" json:"fallback,omitempty"`
	Options    []EnumOption   `yaml:"options,omitempty" json:"options,omitempty"`
	Values     []string       `yaml:"values,omitempty" json:"values,omitempty"`
	LabelField string         `yaml:"label_field,omitempty" json:"label_field,omitempty"`
	ValueField string         `yaml:"value_field,omitempty" json:"value_field,omitempty"`
	RemotePath string         `yaml:"remote_path,omitempty" json:"remote_path,omitempty"`
	Metadata   map[string]any `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type EnumOption struct {
	Value    string         `yaml:"value,omitempty" json:"value,omitempty"`
	Label    string         `yaml:"label,omitempty" json:"label,omitempty"`
	Color    string         `yaml:"color,omitempty" json:"color,omitempty"`
	Disabled bool           `yaml:"disabled,omitempty" json:"disabled,omitempty"`
	Order    int            `yaml:"order,omitempty" json:"order,omitempty"`
	Metadata map[string]any `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Framework struct {
	Backend  string `yaml:"backend" json:"backend"`
	Frontend string `yaml:"frontend" json:"frontend"`
}

type Entity struct {
	Name   string  `yaml:"name" json:"name"`
	Fields []Field `yaml:"fields,omitempty" json:"fields,omitempty"`
}

type Page struct {
	Name      string `yaml:"name" json:"name"`
	Type      string `yaml:"type,omitempty" json:"type,omitempty"`
	Path      string `yaml:"path,omitempty" json:"path,omitempty"`
	Component string `yaml:"component,omitempty" json:"component,omitempty"`
}

type Route struct {
	Method string `yaml:"method" json:"method"`
	Path   string `yaml:"path" json:"path"`
	Name   string `yaml:"name,omitempty" json:"name,omitempty"`
}

type Permission struct {
	Name     string `yaml:"name" json:"name"`
	Action   string `yaml:"action,omitempty" json:"action,omitempty"`
	Resource string `yaml:"resource,omitempty" json:"resource,omitempty"`
}

type Plugin struct {
	Name  string `yaml:"name" json:"name"`
	Route string `yaml:"route,omitempty" json:"route,omitempty"`
	View  string `yaml:"view,omitempty" json:"view,omitempty"`
}

type Resource struct {
	Kind             Kind         `yaml:"kind" json:"kind"`
	Name             string       `yaml:"name" json:"name"`
	Module           string       `yaml:"module,omitempty" json:"module,omitempty"`
	Comment          string       `yaml:"comment,omitempty" json:"comment,omitempty"`
	Database         string       `yaml:"database,omitempty" json:"database,omitempty"`
	Schema           string       `yaml:"schema,omitempty" json:"schema,omitempty"`
	MountParentPath  string       `yaml:"mount_parent_path,omitempty" json:"mount_parent_path,omitempty"`
	Framework        Framework    `yaml:"framework,omitempty" json:"framework,omitempty"`
	Entity           Entity       `yaml:"entity,omitempty" json:"entity,omitempty"`
	Fields           []Field      `yaml:"fields,omitempty" json:"fields,omitempty"`
	Pages            []Page       `yaml:"pages,omitempty" json:"pages,omitempty"`
	Permissions      []Permission `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Routes           []Route      `yaml:"routes,omitempty" json:"routes,omitempty"`
	Plugin           *Plugin      `yaml:"plugin,omitempty" json:"plugin,omitempty"`
	GenerateFrontend bool         `yaml:"generate_frontend,omitempty" json:"generate_frontend,omitempty"`
	GeneratePolicy   bool         `yaml:"generate_policy,omitempty" json:"generate_policy,omitempty"`
	Force            bool         `yaml:"force,omitempty" json:"force,omitempty"`
}

type Document struct {
	Version     string       `yaml:"version,omitempty" json:"version,omitempty"`
	Module      string       `yaml:"module,omitempty" json:"module,omitempty"`
	Kind        Kind         `yaml:"kind,omitempty" json:"kind,omitempty"`
	Framework   Framework    `yaml:"framework,omitempty" json:"framework,omitempty"`
	Entity      Entity       `yaml:"entity,omitempty" json:"entity,omitempty"`
	Pages       []Page       `yaml:"pages,omitempty" json:"pages,omitempty"`
	Permissions []Permission `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Routes      []Route      `yaml:"routes,omitempty" json:"routes,omitempty"`
	Plugin      *Plugin      `yaml:"plugin,omitempty" json:"plugin,omitempty"`
	Resources   []Resource   `yaml:"resources,omitempty" json:"resources,omitempty"`
}

func (d Document) Validate() error {
	resources, err := d.ResolveResources()
	if err != nil {
		return err
	}
	for i, resource := range resources {
		if err := resource.Validate(); err != nil {
			return fmt.Errorf("resource[%d]: %w", i, err)
		}
	}
	return nil
}

func (r Resource) Validate() error {
	if strings.TrimSpace(string(r.Kind)) == "" {
		return fmt.Errorf("resource kind is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("resource name is required")
	}
	if err := r.Entity.Validate(); err != nil {
		return err
	}
	for i, field := range r.Fields {
		if err := field.Validate(); err != nil {
			return fmt.Errorf("field[%d]: %w", i, err)
		}
	}
	for i, page := range r.Pages {
		if err := page.Validate(); err != nil {
			return fmt.Errorf("page[%d]: %w", i, err)
		}
	}
	for i, permission := range r.Permissions {
		if err := permission.Validate(); err != nil {
			return fmt.Errorf("permission[%d]: %w", i, err)
		}
	}
	for i, route := range r.Routes {
		if err := route.Validate(); err != nil {
			return fmt.Errorf("route[%d]: %w", i, err)
		}
	}
	if r.Plugin != nil {
		if err := r.Plugin.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func ParseCSV(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func ParseFields(spec string) ([]Field, error) {
	items := ParseCSV(spec)
	if len(items) == 0 {
		return nil, nil
	}
	fields := make([]Field, 0, len(items))
	for _, item := range items {
		name := item
		typeName := "string"
		if left, right, ok := strings.Cut(item, ":"); ok {
			name = strings.TrimSpace(left)
			typeName = strings.TrimSpace(right)
		}
		if strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("field name is required in %q", item)
		}
		if strings.TrimSpace(typeName) == "" {
			typeName = "string"
		}
		fields = append(fields, Field{Name: strings.TrimSpace(name), Type: typeName})
	}
	return fields, nil
}

func NormalizeName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func NewDocument(resources ...Resource) Document {
	return Document{Resources: resources}
}

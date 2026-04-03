// Package ir defines the unified intermediate representation used by the
// CodeGen pipeline before template rendering and file generation.
package ir

// SourceKind identifies the origin of an IR document or resource.
type SourceKind string

const (
	SourceKindUnknown   SourceKind = ""
	SourceKindDatabase  SourceKind = "database"
	SourceKindDSL       SourceKind = "dsl"
	SourceKindAPISchema SourceKind = "api-schema"
)

// Field describes the unified field model used across database, DSL and API inputs.
type Field struct {
	Name         string         `json:"name"`
	ColumnName   string         `json:"column_name,omitempty"`
	GoType       string         `json:"go_type,omitempty"`
	DBType       string         `json:"db_type,omitempty"`
	Nullable     bool           `json:"nullable,omitempty"`
	Primary      bool           `json:"primary,omitempty"`
	Unique       bool           `json:"unique,omitempty"`
	Index        bool           `json:"index,omitempty"`
	Required     bool           `json:"required,omitempty"`
	UIType       string         `json:"ui_type,omitempty"`
	Label        string         `json:"label,omitempty"`
	Searchable   bool           `json:"searchable,omitempty"`
	Editable     bool           `json:"editable,omitempty"`
	Sortable     bool           `json:"sortable,omitempty"`
	SemanticType string         `json:"semantic_type,omitempty"`
	DefaultValue string         `json:"default_value,omitempty"`
	EnumValues   []string       `json:"enum_values,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// Semantic stores rule-driven semantic hints derived from raw structure.
type Semantic struct {
	HasSoftDelete bool                `json:"has_soft_delete,omitempty"`
	HasAudit      bool                `json:"has_audit,omitempty"`
	EnumFields    map[string][]string `json:"enum_fields,omitempty"`
	StatusField   string              `json:"status_field,omitempty"`
	MoneyFields   []string            `json:"money_fields,omitempty"`
	TimeFields    []string            `json:"time_fields,omitempty"`
	BoolFields    []string            `json:"bool_fields,omitempty"`
	Metadata      map[string]any      `json:"metadata,omitempty"`
}

// Relation describes a relationship between two resources.
type Relation struct {
	Type            string         `json:"type,omitempty"`
	Field           string         `json:"field,omitempty"`
	RefTable        string         `json:"ref_table,omitempty"`
	RefField        string         `json:"ref_field,omitempty"`
	UIHint          string         `json:"ui_hint,omitempty"`
	Cardinality     string         `json:"cardinality,omitempty"`
	RefDisplayField string         `json:"ref_display_field,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// Page describes a generated or suggested page.
type Page struct {
	Name      string         `json:"name,omitempty"`
	Type      string         `json:"type,omitempty"`
	Path      string         `json:"path,omitempty"`
	Component string         `json:"component,omitempty"`
	Title     string         `json:"title,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Permission describes a generated permission entry.
type Permission struct {
	Name     string         `json:"name,omitempty"`
	Action   string         `json:"action,omitempty"`
	Resource string         `json:"resource,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Route describes a generated or suggested route entry.
type Route struct {
	Method    string         `json:"method,omitempty"`
	Path      string         `json:"path,omitempty"`
	Name      string         `json:"name,omitempty"`
	Component string         `json:"component,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Resource is the unified IR resource model.
type Resource struct {
	Name        string         `json:"name,omitempty"`
	Module      string         `json:"module,omitempty"`
	EntityName  string         `json:"entity_name,omitempty"`
	TableName   string         `json:"table_name,omitempty"`
	Kind        string         `json:"kind,omitempty"`
	Source      SourceKind     `json:"source,omitempty"`
	Fields      []Field        `json:"fields,omitempty"`
	Relations   []Relation     `json:"relations,omitempty"`
	Semantic    Semantic       `json:"semantic,omitempty"`
	Pages       []Page         `json:"pages,omitempty"`
	Permissions []Permission   `json:"permissions,omitempty"`
	Routes      []Route        `json:"routes,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// Document groups a set of IR resources and top-level metadata.
type Document struct {
	Version   string         `json:"version,omitempty"`
	Resources []Resource     `json:"resources,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

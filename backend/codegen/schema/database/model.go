// Package database provides the normalized database metadata model used by
// the database-driven CodeGen pipeline.
package database

import "time"

// DriverKind identifies the database dialect or source family.
type DriverKind string

const (
	DriverKindUnknown    DriverKind = ""
	DriverKindMySQL      DriverKind = "mysql"
	DriverKindPostgreSQL DriverKind = "postgresql"
	DriverKindSQLite     DriverKind = "sqlite"
	DriverKindSQLServer  DriverKind = "sqlserver"
)

// Table describes a database table after inspection.
type Table struct {
	Schema      string         `json:"schema,omitempty"`
	Name        string         `json:"name"`
	Comment     string         `json:"comment,omitempty"`
	PrimaryKeys []string       `json:"primary_keys,omitempty"`
	Columns     []Column       `json:"columns,omitempty"`
	Indexes     []Index        `json:"indexes,omitempty"`
	ForeignKeys []ForeignKey   `json:"foreign_keys,omitempty"`
	Engine      string         `json:"engine,omitempty"`
	Charset     string         `json:"charset,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// Column describes a single column within a database table.
type Column struct {
	Name          string         `json:"name"`
	Type          string         `json:"type"`
	Nullable      bool           `json:"nullable,omitempty"`
	Default       string         `json:"default,omitempty"`
	Comment       string         `json:"comment,omitempty"`
	UIType        string         `json:"ui_type,omitempty"`
	Length        *int           `json:"length,omitempty"`
	Precision     *int           `json:"precision,omitempty"`
	Scale         *int           `json:"scale,omitempty"`
	Primary       bool           `json:"primary,omitempty"`
	Unique        bool           `json:"unique,omitempty"`
	Index         bool           `json:"index,omitempty"`
	AutoIncrement bool           `json:"auto_increment,omitempty"`
	Generated     bool           `json:"generated,omitempty"`
	EnumKind      string         `json:"enum_kind,omitempty"`
	EnumMode      string         `json:"enum_mode,omitempty"`
	EnumDisplay   string         `json:"enum_display,omitempty"`
	EnumSource    string         `json:"enum_source,omitempty"`
	EnumSourceRef string         `json:"enum_source_ref,omitempty"`
	EnumValues    []string       `json:"enum_values,omitempty"`
	EnumOptions   []EnumOption   `json:"enum_options,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// EnumOption describes a single enum entry discovered from a database column.
type EnumOption struct {
	Value    string         `json:"value,omitempty"`
	Label    string         `json:"label,omitempty"`
	Color    string         `json:"color,omitempty"`
	Disabled bool           `json:"disabled,omitempty"`
	Order    int            `json:"order,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Index describes an index definition.
type Index struct {
	Name     string         `json:"name"`
	Columns  []string       `json:"columns,omitempty"`
	Unique   bool           `json:"unique,omitempty"`
	Primary  bool           `json:"primary,omitempty"`
	Type     string         `json:"type,omitempty"`
	Comment  string         `json:"comment,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ForeignKey describes a foreign key constraint.
type ForeignKey struct {
	Name       string         `json:"name"`
	Columns    []string       `json:"columns,omitempty"`
	RefTable   string         `json:"ref_table,omitempty"`
	RefColumns []string       `json:"ref_columns,omitempty"`
	OnUpdate   string         `json:"on_update,omitempty"`
	OnDelete   string         `json:"on_delete,omitempty"`
	Comment    string         `json:"comment,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// Snapshot captures the inspected state for a database or schema.
type Snapshot struct {
	Driver    DriverKind     `json:"driver"`
	Database  string         `json:"database,omitempty"`
	Schema    string         `json:"schema,omitempty"`
	Tables    []Table        `json:"tables,omitempty"`
	FetchedAt time.Time      `json:"fetched_at,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

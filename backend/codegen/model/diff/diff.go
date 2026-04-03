// Package diff defines the incremental change model used by the CodeGen
// pipeline when comparing source state, IR state and generated output.
package diff

// Type identifies the kind of change captured in a diff item.
type Type string

const (
	TypeAddField         Type = "ADD_FIELD"
	TypeRemoveField      Type = "REMOVE_FIELD"
	TypeModifyField      Type = "MODIFY_FIELD"
	TypeAddRelation      Type = "ADD_RELATION"
	TypeModifyRelation   Type = "MODIFY_RELATION"
	TypeAddPage          Type = "ADD_PAGE"
	TypeRemovePage       Type = "REMOVE_PAGE"
	TypeModifyPermission Type = "MODIFY_PERMISSION"
)

// Severity describes how risky a change is for safe generation.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Item describes a single diff record.
type Item struct {
	Type     Type           `json:"type,omitempty"`
	Target   string         `json:"target,omitempty"`
	Before   any            `json:"before,omitempty"`
	After    any            `json:"after,omitempty"`
	Patch    string         `json:"patch,omitempty"`
	Severity Severity       `json:"severity,omitempty"`
	Conflict bool           `json:"conflict,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Document groups multiple diff items together.
type Document struct {
	Items    []Item         `json:"items,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Package diff defines the incremental change model used by the CodeGen
// pipeline when comparing source state, IR state and generated output.
package diff

import "strings"

// Type identifies the kind of change captured in a diff item.
type Type string

const (
	TypeAddField          Type = "ADD_FIELD"
	TypeRemoveField       Type = "REMOVE_FIELD"
	TypeModifyField       Type = "MODIFY_FIELD"
	TypeAddRelation       Type = "ADD_RELATION"
	TypeModifyRelation    Type = "MODIFY_RELATION"
	TypeAddPage           Type = "ADD_PAGE"
	TypeRemovePage        Type = "REMOVE_PAGE"
	TypeModifyPermission  Type = "MODIFY_PERMISSION"
	TypeAddFile           Type = "ADD_FILE"
	TypeRemoveFile        Type = "REMOVE_FILE"
	TypeModifyFile        Type = "MODIFY_FILE"
	TypeAddDeclaration    Type = "ADD_DECLARATION"
	TypeRemoveDeclaration Type = "REMOVE_DECLARATION"
	TypeModifyDeclaration Type = "MODIFY_DECLARATION"
	TypeAddImport         Type = "ADD_IMPORT"
	TypeRemoveImport      Type = "REMOVE_IMPORT"
	TypeModifyImport      Type = "MODIFY_IMPORT"
	TypeAddPolicy         Type = "ADD_POLICY"
	TypeModifyPolicy      Type = "MODIFY_POLICY"
	TypeRemovePolicy      Type = "REMOVE_POLICY"
	TypeMergeConflict     Type = "MERGE_CONFLICT"
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

// Append returns a new diff document with the provided items appended.
func (d Document) Append(items ...Item) Document {
	if len(items) == 0 {
		return d
	}
	result := Document{
		Items:    append([]Item(nil), d.Items...),
		Metadata: cloneMetadata(d.Metadata),
	}
	result.Items = append(result.Items, items...)
	return result
}

// HasConflict reports whether any item in the document is marked as a conflict.
func (d Document) HasConflict() bool {
	for _, item := range d.Items {
		if item.IsConflict() {
			return true
		}
	}
	return false
}

// ConflictCount returns the number of conflict items in the document.
func (d Document) ConflictCount() int {
	count := 0
	for _, item := range d.Items {
		if item.IsConflict() {
			count++
		}
	}
	return count
}

// IsConflict reports whether the item should be treated as a merge conflict.
func (i Item) IsConflict() bool {
	return i.Conflict || i.Type == TypeMergeConflict || i.Severity == SeverityCritical
}

// NewItem creates a diff item with trimmed target and patch fields.
func NewItem(kind Type, target string, severity Severity) Item {
	return Item{Type: kind, Target: strings.TrimSpace(target), Severity: severity}
}

func cloneMetadata(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

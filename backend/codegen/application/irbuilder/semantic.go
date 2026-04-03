package irbuilder

import (
	"strings"

	irmodel "goadmin/codegen/model/ir"
	dbschema "goadmin/codegen/schema/database"
)

// SemanticOptions customizes the rule-driven semantic layer built on top of IR.
type SemanticOptions struct {
	SoftDeleteFieldNames   []string
	AuditFieldNames        []string
	StatusFieldNames       []string
	MoneyNameKeywords      []string
	TimeNameKeywords       []string
	BoolNamePrefixes       []string
	DisplayFieldCandidates []string
	FieldOverrides         map[string]FieldSemanticOverride
	RelationOverrides      map[string]RelationSemanticOverride
}

// FieldSemanticOverride allows callers to override field semantic hints.
type FieldSemanticOverride struct {
	SemanticType string
	UIType       string
	Label        string
	Searchable   *bool
	Editable     *bool
	Sortable     *bool
}

// RelationSemanticOverride allows callers to override relation suggestions.
type RelationSemanticOverride struct {
	Type            string
	UIHint          string
	Cardinality     string
	RefDisplayField string
	Confidence      string
	Metadata        map[string]any
}

// DefaultSemanticOptions returns the default rule set used by the IR builder.
func DefaultSemanticOptions() *SemanticOptions {
	return &SemanticOptions{
		SoftDeleteFieldNames: []string{"deleted_at", "deleted_on", "is_deleted"},
		AuditFieldNames: []string{
			"created_at",
			"updated_at",
			"created_on",
			"updated_on",
			"created_by",
			"updated_by",
		},
		StatusFieldNames: []string{"status", "state", "stage"},
		MoneyNameKeywords: []string{
			"amount",
			"balance",
			"cost",
			"fee",
			"money",
			"price",
			"salary",
			"total",
		},
		TimeNameKeywords:       []string{"time", "date", "timestamp", "at", "on"},
		BoolNamePrefixes:       []string{"is_", "has_", "can_", "should_", "allow_", "enable_"},
		DisplayFieldCandidates: []string{"name", "title", "code", "label", "display_name", "username", "nickname", "slug"},
	}
}

func normalizeSemanticOptions(opts *SemanticOptions) SemanticOptions {
	base := DefaultSemanticOptions()
	if opts == nil {
		return *base
	}
	merged := *base
	if opts.SoftDeleteFieldNames != nil {
		merged.SoftDeleteFieldNames = cloneStrings(opts.SoftDeleteFieldNames)
	}
	if opts.AuditFieldNames != nil {
		merged.AuditFieldNames = cloneStrings(opts.AuditFieldNames)
	}
	if opts.StatusFieldNames != nil {
		merged.StatusFieldNames = cloneStrings(opts.StatusFieldNames)
	}
	if opts.MoneyNameKeywords != nil {
		merged.MoneyNameKeywords = cloneStrings(opts.MoneyNameKeywords)
	}
	if opts.TimeNameKeywords != nil {
		merged.TimeNameKeywords = cloneStrings(opts.TimeNameKeywords)
	}
	if opts.BoolNamePrefixes != nil {
		merged.BoolNamePrefixes = cloneStrings(opts.BoolNamePrefixes)
	}
	if opts.DisplayFieldCandidates != nil {
		merged.DisplayFieldCandidates = cloneStrings(opts.DisplayFieldCandidates)
	}
	if len(opts.FieldOverrides) > 0 {
		merged.FieldOverrides = cloneFieldSemanticOverrides(opts.FieldOverrides)
	}
	if len(opts.RelationOverrides) > 0 {
		merged.RelationOverrides = cloneRelationSemanticOverrides(opts.RelationOverrides)
	}
	return merged
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func cloneFieldSemanticOverrides(src map[string]FieldSemanticOverride) map[string]FieldSemanticOverride {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]FieldSemanticOverride, len(src))
	for key, value := range src {
		dst[normalizeRuleKey(key)] = value
	}
	return dst
}

func cloneRelationSemanticOverrides(src map[string]RelationSemanticOverride) map[string]RelationSemanticOverride {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]RelationSemanticOverride, len(src))
	for key, value := range src {
		dst[normalizeRuleKey(key)] = value
	}
	return dst
}

func normalizeRuleKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func fieldRuleKey(tableName string, columnName string) string {
	return normalizeRuleKey(normalizeSchemaName(tableName) + "." + normalizeSchemaName(columnName))
}

func relationRuleKey(tableName string, relationField string, fkName string) string {
	key := fieldRuleKey(tableName, relationField)
	if key != "." {
		return key
	}
	if key = normalizeRuleKey(fkName); key != "" {
		return key
	}
	return key
}

func stringMetadata(metadata map[string]any, key string) (string, bool) {
	if len(metadata) == 0 {
		return "", false
	}
	value, ok := metadata[key]
	if !ok {
		return "", false
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", false
	}
	return text, true
}

func boolMetadataValue(metadata map[string]any, key string) (bool, bool) {
	if len(metadata) == 0 {
		return false, false
	}
	value, ok := metadata[key]
	if !ok {
		return false, false
	}
	flag, ok := value.(bool)
	return flag, ok
}

func applyFieldSemanticHints(field *irmodel.Field, table dbschema.Table, column dbschema.Column, rules SemanticOptions) {
	if field == nil {
		return
	}
	if field.Metadata == nil {
		field.Metadata = map[string]any{}
	}
	fieldKey := fieldRuleKey(table.Name, column.Name)
	field.Metadata["semantic_key"] = fieldKey

	semanticType := inferFieldSemanticType(field, column, rules)
	uiType := inferFieldUIType(field, semanticType)
	semanticSource := "heuristic"
	label := field.Label

	if text, ok := stringMetadata(column.Metadata, "label"); ok {
		label = text
		semanticSource = "metadata"
	}
	if text, ok := stringMetadata(column.Metadata, "semantic_type"); ok {
		semanticType = text
		semanticSource = "metadata"
	}
	if text, ok := stringMetadata(column.Metadata, "ui_type"); ok {
		uiType = text
		semanticSource = "metadata"
	}
	if flag, ok := boolMetadataValue(column.Metadata, "searchable"); ok {
		field.Searchable = flag
	}
	if flag, ok := boolMetadataValue(column.Metadata, "editable"); ok {
		field.Editable = flag
	}
	if flag, ok := boolMetadataValue(column.Metadata, "sortable"); ok {
		field.Sortable = flag
	}
	if override, ok := getFieldOverride(rules, table.Name, column.Name); ok {
		if override.SemanticType != "" {
			semanticType = override.SemanticType
		}
		if override.UIType != "" {
			uiType = override.UIType
		}
		if override.Label != "" {
			label = override.Label
		}
		if override.Searchable != nil {
			field.Searchable = *override.Searchable
		}
		if override.Editable != nil {
			field.Editable = *override.Editable
		}
		if override.Sortable != nil {
			field.Sortable = *override.Sortable
		}
		semanticSource = "rule"
	}
	if semanticType == "" {
		semanticType = field.SemanticType
	}
	if uiType == "" {
		uiType = field.UIType
	}
	field.SemanticType = semanticType
	field.UIType = uiType
	if label != "" {
		field.Label = label
	}
	field.Metadata["semantic_type"] = semanticType
	field.Metadata["ui_type"] = field.UIType
	field.Metadata["semantic_source"] = semanticSource
	field.Metadata["table_name"] = table.Name
	field.Metadata["column_name"] = column.Name
	field.Metadata["semantic_rule_key"] = fieldKey
	if len(field.EnumValues) > 0 {
		field.Metadata["enum_values"] = append([]string(nil), field.EnumValues...)
	}
	if field.Primary {
		field.Metadata["primary"] = true
	}
}

func inferFieldSemanticType(field *irmodel.Field, column dbschema.Column, rules SemanticOptions) string {
	name := normalizeRuleKey(column.Name)
	if name == "" {
		name = normalizeRuleKey(field.Name)
	}
	if len(column.EnumValues) > 0 {
		return "enum"
	}
	if column.Primary && name == "id" {
		return "identifier"
	}
	if containsAny(name, rules.SoftDeleteFieldNames) {
		return "soft_delete"
	}
	if containsAny(name, rules.AuditFieldNames) {
		return "audit"
	}
	if containsAny(name, rules.StatusFieldNames) {
		return "status"
	}
	if containsAny(name, rules.MoneyNameKeywords) {
		return "money"
	}
	if field.GoType == "bool" || hasAnyPrefix(name, rules.BoolNamePrefixes) {
		return "boolean"
	}
	if field.GoType == "time.Time" || containsAny(name, rules.TimeNameKeywords) {
		return "datetime"
	}
	if column.Unique && strings.Contains(name, "code") {
		return "identifier"
	}
	if field.SemanticType != "" {
		return field.SemanticType
	}
	return "text"
}

func inferFieldUIType(field *irmodel.Field, semanticType string) string {
	switch semanticType {
	case "enum", "status":
		return "select"
	case "boolean":
		return "switch"
	case "money":
		return "number"
	case "datetime", "audit", "soft_delete":
		if field.GoType == "bool" {
			return "switch"
		}
		if field.GoType == "time.Time" || field.UIType == "datetime" {
			return "datetime"
		}
	case "identifier":
		return field.UIType
	}
	return field.UIType
}

func getFieldOverride(rules SemanticOptions, tableName string, columnName string) (FieldSemanticOverride, bool) {
	if len(rules.FieldOverrides) == 0 {
		return FieldSemanticOverride{}, false
	}
	if override, ok := rules.FieldOverrides[fieldRuleKey(tableName, columnName)]; ok {
		return override, true
	}
	if override, ok := rules.FieldOverrides[normalizeRuleKey(columnName)]; ok {
		return override, true
	}
	return FieldSemanticOverride{}, false
}

func getRelationOverride(rules SemanticOptions, tableName string, relationField string, fkName string) (RelationSemanticOverride, bool) {
	if len(rules.RelationOverrides) == 0 {
		return RelationSemanticOverride{}, false
	}
	if override, ok := rules.RelationOverrides[fieldRuleKey(tableName, relationField)]; ok {
		return override, true
	}
	if override, ok := rules.RelationOverrides[normalizeRuleKey(fkName)]; ok {
		return override, true
	}
	if override, ok := rules.RelationOverrides[normalizeRuleKey(relationField)]; ok {
		return override, true
	}
	return RelationSemanticOverride{}, false
}

func applyRelationSemanticHints(resource *irmodel.Resource, table dbschema.Table, columns []dbschema.Column, rules SemanticOptions) {
	if resource == nil || len(resource.Relations) == 0 {
		return
	}
	joinCandidate := isJoinTableCandidate(columns, resource.Relations)
	for idx := range resource.Relations {
		relation := &resource.Relations[idx]
		if relation.Metadata == nil {
			relation.Metadata = map[string]any{}
		}
		relationKey := fieldRuleKey(table.Name, relation.Field)
		relation.Metadata["relation_key"] = relationKey
		relation.Metadata["source_table"] = table.Name
		relation.Metadata["inferred"] = true
		relation.Metadata["override_allowed"] = true
		relation.Metadata["display_field_candidates"] = cloneStrings(rules.DisplayFieldCandidates)

		if override, ok := getRelationOverride(rules, table.Name, relation.Field, relation.Metadata["name"].(string)); ok {
			applyRelationOverride(relation, override)
			relation.Metadata["semantic_source"] = "rule"
			continue
		}

		if joinCandidate {
			relation.Type = "manyToMany"
			relation.Cardinality = "many-to-many"
			relation.UIHint = "transfer"
			relation.Metadata["join_table_candidate"] = true
			relation.Metadata["confidence"] = "suggested"
		} else {
			if relation.Type == "" {
				relation.Type = "belongsTo"
			}
			if relation.Cardinality == "" {
				if isUniqueRelation(table, columns, relation.Field) {
					relation.Cardinality = "one-to-one"
				} else {
					relation.Cardinality = "many-to-one"
				}
			}
			if relation.UIHint == "" {
				relation.UIHint = defaultUIHintForCardinality(relation.Cardinality)
			}
			relation.Metadata["confidence"] = "suggested"
		}
		if relation.RefDisplayField == "" {
			relation.RefDisplayField = guessDisplayField(rules.DisplayFieldCandidates)
		}
		relation.Metadata["ref_display_field_candidates"] = cloneStrings(rules.DisplayFieldCandidates)
	}
	if resource.Semantic.Metadata == nil {
		resource.Semantic.Metadata = map[string]any{}
	}
	resource.Semantic.Metadata["join_table_candidate"] = joinCandidate
}

func applyRelationOverride(relation *irmodel.Relation, override RelationSemanticOverride) {
	if relation == nil {
		return
	}
	if override.Type != "" {
		relation.Type = override.Type
	}
	if override.UIHint != "" {
		relation.UIHint = override.UIHint
	}
	if override.Cardinality != "" {
		relation.Cardinality = override.Cardinality
	}
	if override.RefDisplayField != "" {
		relation.RefDisplayField = override.RefDisplayField
	}
	if relation.Metadata == nil {
		relation.Metadata = map[string]any{}
	}
	if override.Confidence != "" {
		relation.Metadata["confidence"] = override.Confidence
	}
	for key, value := range override.Metadata {
		relation.Metadata[key] = value
	}
}

func buildResourceSemantic(resource irmodel.Resource, table dbschema.Table, rules SemanticOptions) irmodel.Semantic {
	semantic := irmodel.Semantic{
		EnumFields: map[string][]string{},
		Metadata:   map[string]any{},
	}
	fieldSemantics := map[string]string{}
	for _, field := range resource.Fields {
		semanticType := strings.TrimSpace(field.SemanticType)
		if semanticType == "" {
			semanticType = strings.TrimSpace(field.UIType)
		}
		fieldSemantics[field.Name] = semanticType
		switch semanticType {
		case "soft_delete":
			semantic.HasSoftDelete = true
			semantic.TimeFields = append(semantic.TimeFields, field.Name)
		case "audit":
			semantic.HasAudit = true
			semantic.TimeFields = append(semantic.TimeFields, field.Name)
		case "enum":
			semantic.EnumFields[field.Name] = append([]string(nil), field.EnumValues...)
			if semantic.StatusField == "" && containsAny(normalizeRuleKey(field.Name), rules.StatusFieldNames) {
				semantic.StatusField = field.Name
			}
			if semantic.StatusField == "" {
				semantic.StatusField = field.Name
			}
		case "status":
			if semantic.StatusField == "" {
				semantic.StatusField = field.Name
			}
		case "money":
			semantic.MoneyFields = append(semantic.MoneyFields, field.Name)
		case "datetime":
			semantic.TimeFields = append(semantic.TimeFields, field.Name)
		case "boolean":
			semantic.BoolFields = append(semantic.BoolFields, field.Name)
		case "identifier":
			if field.Primary {
				semantic.Metadata["primary_identifier"] = field.Name
			}
		}
		if field.Primary && semantic.Metadata["primary_identifier"] == nil {
			semantic.Metadata["primary_identifier"] = field.Name
		}
	}
	for _, relation := range resource.Relations {
		semantic.Metadata["relations"] = appendSemanticRelationSummary(semantic.Metadata["relations"], relation)
	}
	semantic.Metadata["field_semantics"] = fieldSemantics
	semantic.Metadata["field_count"] = len(resource.Fields)
	semantic.Metadata["relation_count"] = len(resource.Relations)
	semantic.Metadata["source_table"] = table.Name
	semantic.Metadata["display_field_candidates"] = cloneStrings(rules.DisplayFieldCandidates)
	if len(semantic.TimeFields) > 0 {
		semantic.HasAudit = semantic.HasAudit || containsAuditFieldNames(semantic.TimeFields)
	}
	if len(semantic.EnumFields) == 0 {
		semantic.EnumFields = nil
	}
	if len(semantic.TimeFields) == 0 {
		semantic.TimeFields = nil
	}
	if len(semantic.BoolFields) == 0 {
		semantic.BoolFields = nil
	}
	if len(semantic.MoneyFields) == 0 {
		semantic.MoneyFields = nil
	}
	if len(semantic.Metadata) == 0 {
		semantic.Metadata = nil
	}
	return semantic
}

func appendSemanticRelationSummary(existing any, relation irmodel.Relation) []map[string]any {
	summary, _ := existing.([]map[string]any)
	item := map[string]any{
		"field":             relation.Field,
		"ref_table":         relation.RefTable,
		"ref_field":         relation.RefField,
		"ui_hint":           relation.UIHint,
		"cardinality":       relation.Cardinality,
		"ref_display_field": relation.RefDisplayField,
	}
	return append(summary, item)
}

func guessDisplayField(candidates []string) string {
	for _, candidate := range candidates {
		name := strings.TrimSpace(candidate)
		if name == "" {
			continue
		}
		return toExportedName(name)
	}
	return "Name"
}

func defaultUIHintForCardinality(cardinality string) string {
	switch cardinality {
	case "one-to-one", "many-to-one", "many-to-many":
		return map[string]string{
			"one-to-one":   "select",
			"many-to-one":  "select",
			"many-to-many": "transfer",
		}[cardinality]
	default:
		return "select"
	}
}

func isUniqueRelation(table dbschema.Table, columns []dbschema.Column, relationField string) bool {
	candidate := normalizeRuleKey(relationField)
	for _, column := range columns {
		if normalizeRuleKey(toExportedName(column.Name)) != candidate {
			continue
		}
		if column.Primary || column.Unique {
			return true
		}
		if flag, ok := boolMetadataValue(column.Metadata, "unique"); ok && flag {
			return true
		}
	}
	return false
}

func isJoinTableCandidate(columns []dbschema.Column, relations []irmodel.Relation) bool {
	if len(relations) < 2 || len(columns) == 0 {
		return false
	}
	fkFields := make(map[string]struct{}, len(relations))
	for _, relation := range relations {
		if relation.Field == "" {
			continue
		}
		fkFields[normalizeRuleKey(relation.Field)] = struct{}{}
	}
	if len(fkFields) < 2 {
		return false
	}
	for _, column := range columns {
		if column.Primary {
			continue
		}
		if _, ok := fkFields[normalizeRuleKey(toExportedName(column.Name))]; ok {
			continue
		}
		if column.Unique || column.Index {
			continue
		}
		if flag, ok := boolMetadataValue(column.Metadata, "generated"); ok && flag {
			continue
		}
		return false
	}
	return true
}

func containsAny(value string, candidates []string) bool {
	if value == "" {
		return false
	}
	for _, candidate := range candidates {
		candidate = normalizeRuleKey(candidate)
		if candidate == "" {
			continue
		}
		if value == candidate || strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func hasAnyPrefix(value string, prefixes []string) bool {
	if value == "" {
		return false
	}
	for _, prefix := range prefixes {
		prefix = normalizeRuleKey(prefix)
		if prefix == "" {
			continue
		}
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func containsAuditFieldNames(values []string) bool {
	for _, value := range values {
		name := normalizeRuleKey(value)
		if name == "created_at" || name == "updated_at" || name == "created_on" || name == "updated_on" || name == "created_by" || name == "updated_by" {
			return true
		}
	}
	return false
}

package irbuilder

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	insp "goadmin/codegen/infrastructure/inspector"
	irmodel "goadmin/codegen/model/ir"
	codeschema "goadmin/codegen/schema"
	dbschema "goadmin/codegen/schema/database"

	"gorm.io/gorm"
)

const defaultIRVersion = "v3"

// Dependencies describes the collaborators required by the IR builder.
type Dependencies struct {
	InspectorService *insp.Service
}

// Service converts database inspection results into the unified IR model.
type Service struct {
	inspectors *insp.Service
}

// DatabaseBuildOptions controls how database inspection results are collected
// before they are converted into IR.
type DatabaseBuildOptions struct {
	Tables           []string
	Force            bool
	GenerateFrontend *bool
	GeneratePolicy   *bool
	MountParentPath  string
	Semantic         *SemanticOptions
}

// NewService creates an IR builder service with sensible defaults.
func NewService(deps Dependencies) *Service {
	inspectors := deps.InspectorService
	if inspectors == nil {
		inspectors = insp.NewService(nil)
	}
	return &Service{inspectors: inspectors}
}

// BuildFromDatabase inspects a database through the configured inspector
// service and converts the result into an IR document.
func (s *Service) BuildFromDatabase(db *gorm.DB, database string, schema string) (irmodel.Document, error) {
	return s.BuildFromDatabaseWithOptions(db, database, schema, DatabaseBuildOptions{})
}

// BuildFromDatabaseWithOptions inspects a database through the configured
// inspector service and converts the result into an IR document.
func (s *Service) BuildFromDatabaseWithOptions(db *gorm.DB, database string, schema string, opts DatabaseBuildOptions) (irmodel.Document, error) {
	if s == nil {
		return irmodel.Document{}, fmt.Errorf("ir builder service is required")
	}
	reader := s.inspectors.Open(db, database, schema)
	if reader == nil {
		return irmodel.Document{}, fmt.Errorf("inspector reader is required")
	}
	return s.BuildFromReaderWithOptions(reader, opts)
}

// BuildFromReader converts an already constructed inspector reader into IR.
func (s *Service) BuildFromReader(reader insp.Reader) (irmodel.Document, error) {
	return s.BuildFromReaderWithOptions(reader, DatabaseBuildOptions{})
}

// BuildFromReaderWithOptions converts an already constructed inspector reader
// into IR while allowing table filtering.
func (s *Service) BuildFromReaderWithOptions(reader insp.Reader, opts DatabaseBuildOptions) (irmodel.Document, error) {
	if s == nil {
		return irmodel.Document{}, fmt.Errorf("ir builder service is required")
	}
	if reader == nil {
		return irmodel.Document{}, fmt.Errorf("inspector reader is required")
	}
	tables, err := reader.InspectTables()
	if err != nil {
		return irmodel.Document{}, fmt.Errorf("inspect tables: %w", err)
	}
	tables = filterTables(tables, opts.Tables)
	doc := irmodel.Document{
		Version:   defaultIRVersion,
		Resources: make([]irmodel.Resource, 0, len(tables)),
		Metadata: map[string]any{
			"source": string(irmodel.SourceKindDatabase),
		},
	}
	for _, table := range tables {
		resource, err := buildResource(reader, table, opts)
		if err != nil {
			return irmodel.Document{}, err
		}
		doc.Resources = append(doc.Resources, resource)
	}
	return doc, nil
}

func buildResource(reader insp.Reader, table dbschema.Table, opts DatabaseBuildOptions) (irmodel.Resource, error) {
	rules := normalizeSemanticOptions(opts.Semantic)
	columns := append([]dbschema.Column(nil), table.Columns...)
	if len(columns) == 0 {
		var err error
		columns, err = reader.InspectColumns(table.Name)
		if err != nil {
			return irmodel.Resource{}, fmt.Errorf("inspect columns for %s: %w", table.Name, err)
		}
	}
	foreignKeys := append([]dbschema.ForeignKey(nil), table.ForeignKeys...)
	if len(foreignKeys) == 0 {
		var err error
		foreignKeys, err = reader.InspectRelations(table.Name)
		if err != nil {
			return irmodel.Resource{}, fmt.Errorf("inspect relations for %s: %w", table.Name, err)
		}
	}
	resourceName := tableEntityName(table.Name)
	moduleName := tableModuleName(table.Name)
	resource := irmodel.Resource{
		Name:       resourceName,
		Module:     moduleName,
		EntityName: resourceName,
		TableName:  table.Name,
		Kind:       "crud",
		Source:     irmodel.SourceKindDatabase,
		Fields:     make([]irmodel.Field, 0, len(columns)),
		Relations:  make([]irmodel.Relation, 0, len(foreignKeys)),
		Metadata:   map[string]any{},
	}
	for _, column := range columns {
		field := buildField(column)
		applyFieldSemanticHints(&field, table, column, rules)
		resource.Fields = append(resource.Fields, field)
	}
	for _, fk := range foreignKeys {
		resource.Relations = append(resource.Relations, buildRelation(fk))
	}
	applyRelationSemanticHints(&resource, table, columns, rules)
	resource.Semantic = buildResourceSemantic(resource, table, rules)
	resource.Metadata["source"] = string(irmodel.SourceKindDatabase)
	resource.Metadata["table_name"] = table.Name
	if table.Schema != "" {
		resource.Metadata["schema"] = table.Schema
	}
	if table.Metadata != nil {
		if database, ok := table.Metadata["database"]; ok {
			resource.Metadata["database"] = database
		}
		for key, value := range table.Metadata {
			if key == "database" {
				continue
			}
			if _, exists := resource.Metadata[key]; !exists {
				resource.Metadata[key] = value
			}
		}
	}
	if table.Comment != "" {
		resource.Metadata["comment"] = table.Comment
	}
	if table.Engine != "" {
		resource.Metadata["engine"] = table.Engine
	}
	if table.Charset != "" {
		resource.Metadata["charset"] = table.Charset
	}
	if len(table.PrimaryKeys) > 0 {
		resource.Metadata["primary_keys"] = append([]string(nil), table.PrimaryKeys...)
	}
	if len(table.Indexes) > 0 {
		resource.Metadata["indexes"] = append([]dbschema.Index(nil), table.Indexes...)
	}
	if len(table.ForeignKeys) > 0 {
		resource.Metadata["foreign_keys"] = append([]dbschema.ForeignKey(nil), table.ForeignKeys...)
	}
	if opts.Force {
		resource.Metadata["force"] = true
	}
	if opts.GenerateFrontend != nil {
		resource.Metadata["generate_frontend"] = *opts.GenerateFrontend
	}
	if opts.GeneratePolicy != nil {
		resource.Metadata["generate_policy"] = *opts.GeneratePolicy
	}
	if mountParentPath := strings.TrimSpace(opts.MountParentPath); mountParentPath != "" {
		resource.Metadata["mount_parent_path"] = mountParentPath
	}
	return resource, nil
}

func filterTables(tables []dbschema.Table, allowed []string) []dbschema.Table {
	if len(tables) == 0 || len(allowed) == 0 {
		return tables
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		key := strings.ToLower(strings.TrimSpace(name))
		if key == "" {
			continue
		}
		allowedSet[key] = struct{}{}
	}
	if len(allowedSet) == 0 {
		return tables
	}
	filtered := make([]dbschema.Table, 0, len(tables))
	for _, table := range tables {
		if _, ok := allowedSet[strings.ToLower(strings.TrimSpace(table.Name))]; ok {
			filtered = append(filtered, table)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

func buildField(column dbschema.Column) irmodel.Field {
	category := classifyColumn(column.Type)
	name := toExportedName(column.Name)
	enumValues := append([]string(nil), column.EnumValues...)
	enumOptions := convertDBEnumOptions(column.EnumOptions)
	field := irmodel.Field{
		Name:          name,
		ColumnName:    column.Name,
		GoType:        inferGoType(column.Type),
		DBType:        strings.TrimSpace(column.Type),
		Nullable:      column.Nullable,
		Primary:       column.Primary,
		Unique:        column.Unique,
		Index:         column.Index,
		Required:      !column.Nullable && !column.Primary,
		UIType:        uiTypeForColumn(column, category),
		Label:         humanizeName(column.Name),
		Searchable:    category != "binary",
		Editable:      !column.Primary && !column.AutoIncrement,
		Sortable:      category != "binary",
		SemanticType:  semanticTypeForColumn(column, category),
		DefaultValue:  column.Default,
		EnumKind:      strings.TrimSpace(column.EnumKind),
		EnumMode:      strings.TrimSpace(column.EnumMode),
		EnumDisplay:   strings.TrimSpace(column.EnumDisplay),
		EnumSource:    strings.TrimSpace(column.EnumSource),
		EnumSourceRef: strings.TrimSpace(column.EnumSourceRef),
		EnumValues:    enumValues,
		EnumOptions:   enumOptions,
		Metadata: map[string]any{
			"column_name": column.Name,
		},
	}
	if column.Comment != "" {
		field.Metadata["comment"] = column.Comment
	}
	if len(column.EnumValues) > 0 {
		field.Metadata["enum_values"] = append([]string(nil), column.EnumValues...)
	}
	if len(column.EnumOptions) > 0 {
		field.Metadata["enum_options"] = cloneIRDBEnumOptionMetadata(column.EnumOptions)
	}
	if field.EnumKind != "" {
		field.Metadata["enum_kind"] = field.EnumKind
	}
	if field.EnumMode != "" {
		field.Metadata["enum_mode"] = field.EnumMode
	}
	if field.EnumDisplay != "" {
		field.Metadata["enum_display"] = field.EnumDisplay
	}
	if field.EnumSource != "" {
		field.Metadata["enum_source"] = field.EnumSource
	}
	if field.EnumSourceRef != "" {
		field.Metadata["enum_source_ref"] = field.EnumSourceRef
	}
	if field.UIType != "" {
		field.Metadata["ui_type"] = field.UIType
	}
	if field.HasEnum() {
		field.Metadata["has_enum"] = true
		if strings.TrimSpace(column.Comment) != "" && field.Label == "" {
			field.Label = humanizeName(column.Name)
		}
	} else if column.Comment != "" {
		field.Label = column.Comment
	}
	if column.AutoIncrement {
		field.Metadata["auto_increment"] = true
	}
	if column.Generated {
		field.Metadata["generated"] = true
	}
	if column.Length != nil {
		field.Metadata["length"] = *column.Length
	}
	if column.Precision != nil {
		field.Metadata["precision"] = *column.Precision
	}
	if column.Scale != nil {
		field.Metadata["scale"] = *column.Scale
	}
	return field
}

func buildRelation(fk dbschema.ForeignKey) irmodel.Relation {
	relationField := ""
	if len(fk.Columns) > 0 {
		relationField = toExportedName(fk.Columns[0])
	}
	refField := ""
	if len(fk.RefColumns) > 0 {
		refField = toExportedName(fk.RefColumns[0])
	}
	return irmodel.Relation{
		Type:            "belongsTo",
		Field:           relationField,
		RefTable:        fk.RefTable,
		RefField:        refField,
		UIHint:          "select",
		Cardinality:     "many-to-one",
		RefDisplayField: "",
		Metadata: map[string]any{
			"name": fk.Name,
		},
	}
}

func tableModuleName(tableName string) string {
	return singularizeSnake(normalizeSnake(tableName))
}

func tableEntityName(tableName string) string {
	return toExportedName(singularizeSnake(normalizeSnake(tableName)))
}

func normalizeSnake(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return value
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case '_', '-', '.', '/', ' ':
			return true
		default:
			return false
		}
	})
	return strings.Join(parts, "_")
}

func singularizeSnake(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	switch {
	case strings.HasSuffix(value, "ies") && len(value) > 3:
		return strings.TrimSuffix(value, "ies") + "y"
	case strings.HasSuffix(value, "ches"), strings.HasSuffix(value, "shes"), strings.HasSuffix(value, "xes"), strings.HasSuffix(value, "zes"), strings.HasSuffix(value, "ses"):
		return strings.TrimSuffix(value, "es")
	case strings.HasSuffix(value, "s") && !strings.HasSuffix(value, "ss") && !strings.HasSuffix(value, "us") && !strings.HasSuffix(value, "is"):
		return strings.TrimSuffix(value, "s")
	default:
		return value
	}
}

func toExportedName(value string) string {
	parts := strings.FieldsFunc(strings.TrimSpace(value), func(r rune) bool {
		switch r {
		case '_', '-', '.', '/', ' ':
			return true
		default:
			return false
		}
	})
	if len(parts) == 0 {
		return "Resource"
	}
	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		r, size := utf8.DecodeRuneInString(lower)
		if r == utf8.RuneError {
			continue
		}
		builder.WriteRune(unicode.ToUpper(r))
		builder.WriteString(lower[size:])
	}
	result := builder.String()
	if result == "" {
		return "Resource"
	}
	r, _ := utf8.DecodeRuneInString(result)
	if !unicode.IsLetter(r) {
		return "Field" + result
	}
	return result
}

func cloneAnyMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	result := make(map[string]any, len(src))
	for key, value := range src {
		result[key] = value
	}
	return result
}

func humanizeName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case '_', '-', '.', '/', ' ':
			return true
		default:
			return false
		}
	})
	for i, part := range parts {
		parts[i] = strings.ToUpper(string(part[0])) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, " ")
}

func classifyColumn(columnType string) string {
	t := strings.ToLower(strings.TrimSpace(columnType))
	switch {
	case strings.Contains(t, "bool"):
		return "boolean"
	case strings.Contains(t, "date"), strings.Contains(t, "time"):
		return "datetime"
	case strings.Contains(t, "int"), strings.Contains(t, "dec"), strings.Contains(t, "num"), strings.Contains(t, "real"), strings.Contains(t, "float"), strings.Contains(t, "double"), strings.Contains(t, "serial"), strings.Contains(t, "money"):
		return "number"
	case strings.Contains(t, "blob"), strings.Contains(t, "binary"):
		return "binary"
	default:
		return "text"
	}
}

func inferGoType(columnType string) string {
	switch classifyColumn(columnType) {
	case "boolean":
		return "bool"
	case "datetime":
		return "time.Time"
	case "number":
		return "int64"
	case "binary":
		return "[]byte"
	default:
		return "string"
	}
}

func uiTypeForColumn(column dbschema.Column, category string) string {
	if uiType := codeschema.NormalizeUIType(column.UIType); uiType != "" {
		return uiType
	}
	if hasColumnEnum(column) {
		switch strings.ToLower(strings.TrimSpace(column.EnumMode)) {
		case "multiple":
			return "checkbox-group"
		}
		switch strings.ToLower(strings.TrimSpace(column.EnumDisplay)) {
		case "radio":
			return "radio"
		case "checkbox-group":
			return "checkbox-group"
		case "switch":
			return "switch"
		case "autocomplete":
			return "autocomplete"
		case "remote-select":
			return "select"
		}
		if strings.EqualFold(strings.TrimSpace(column.EnumKind), "dictionary") || strings.TrimSpace(column.EnumSourceRef) != "" {
			return "select"
		}
		return "select"
	}
	switch category {
	case "boolean":
		return "switch"
	case "datetime":
		return "datetime"
	case "number":
		return "number"
	case "binary":
		return "upload"
	default:
		return "input"
	}
}

func semanticTypeForColumn(column dbschema.Column, category string) string {
	if hasColumnEnum(column) {
		if isStatusFieldName(column.Name) || strings.EqualFold(strings.TrimSpace(column.EnumDisplay), "switch") {
			return "status"
		}
		return "enum"
	}
	return category
}

func hasColumnEnum(column dbschema.Column) bool {
	return len(column.EnumValues) > 0 || len(column.EnumOptions) > 0 || strings.TrimSpace(column.EnumKind) != "" || strings.TrimSpace(column.EnumDisplay) != "" || strings.TrimSpace(column.EnumSource) != "" || strings.TrimSpace(column.EnumSourceRef) != ""
}

func isStatusFieldName(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	return strings.Contains(normalized, "status") || strings.Contains(normalized, "state") || strings.Contains(normalized, "enabled") || strings.Contains(normalized, "disabled")
}

func convertDBEnumOptions(options []dbschema.EnumOption) []irmodel.EnumOption {
	if len(options) == 0 {
		return nil
	}
	result := make([]irmodel.EnumOption, 0, len(options))
	for _, option := range options {
		result = append(result, irmodel.EnumOption{
			Value:    strings.TrimSpace(option.Value),
			Label:    strings.TrimSpace(option.Label),
			Color:    strings.TrimSpace(option.Color),
			Disabled: option.Disabled,
			Order:    option.Order,
			Metadata: cloneAnyMap(option.Metadata),
		})
	}
	return result
}

func cloneIRDBEnumOptionMetadata(options []dbschema.EnumOption) []map[string]any {
	if len(options) == 0 {
		return nil
	}
	result := make([]map[string]any, 0, len(options))
	for _, option := range options {
		metadata := map[string]any{
			"value":    strings.TrimSpace(option.Value),
			"label":    strings.TrimSpace(option.Label),
			"color":    strings.TrimSpace(option.Color),
			"disabled": option.Disabled,
			"order":    option.Order,
		}
		if option.Metadata != nil {
			for key, value := range option.Metadata {
				metadata[key] = value
			}
		}
		result = append(result, metadata)
	}
	return result
}

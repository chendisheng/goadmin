package generate

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type Field struct {
	Name          string
	GoName        string
	JSONName      string
	GoType        string
	Column        string
	Comment       string
	Primary       bool
	Index         bool
	Unique        bool
	EnumKind      string
	EnumMode      string
	EnumDisplay   string
	EnumSource    string
	EnumSourceRef string
	EnumValues    []string
	EnumOptions   []EnumOption
}

type EnumOption struct {
	Value    string
	Label    string
	Color    string
	Disabled bool
	Order    int
	Metadata map[string]any
}

func (f Field) DisplayLabel() string {
	name := strings.TrimSpace(f.JSONName)
	if name == "" {
		name = strings.TrimSpace(f.GoName)
	}
	if name == "" {
		return "Field"
	}
	parts := strings.Split(name, "_")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func (f Field) TSValueType() string {
	if f.HasEnum() {
		if strings.EqualFold(strings.TrimSpace(f.EnumMode), "multiple") {
			return "string[]"
		}
		return "string"
	}
	switch f.GoType {
	case "bool":
		return "boolean"
	case "int", "int64", "int32", "float64":
		return "number"
	case "time.Time":
		return "string"
	case "[]string":
		return "string[]"
	case "[]int":
		return "number[]"
	case "[]int64":
		return "number[]"
	case "map[string]any":
		return "Record<string, any>"
	default:
		return "string"
	}
}

func ParseGoType(input string) string {
	return mapGoType(input)
}

func (f Field) TSFormValueType() string {
	if f.HasEnum() {
		if strings.EqualFold(strings.TrimSpace(f.EnumMode), "multiple") {
			return "string[]"
		}
		return "string"
	}
	switch {
	case f.IsTime():
		return "string"
	case f.GoType == "bool":
		return "boolean"
	case f.GoType == "int" || f.GoType == "int64" || f.GoType == "int32" || f.GoType == "float64":
		return "number"
	default:
		return "string"
	}
}

func (f Field) FormDefaultValue() string {
	if f.HasEnum() {
		if strings.EqualFold(strings.TrimSpace(f.EnumMode), "multiple") {
			return "[]"
		}
		return "''"
	}
	switch {
	case f.IsTime():
		return "''"
	case f.GoType == "bool":
		return "false"
	case f.GoType == "int" || f.GoType == "int64" || f.GoType == "int32" || f.GoType == "float64":
		return "0"
	default:
		return "''"
	}
}

func (f Field) TSDefaultValue() string {
	if f.HasEnum() {
		if strings.EqualFold(strings.TrimSpace(f.EnumMode), "multiple") {
			return "[]"
		}
		return "''"
	}
	switch f.GoType {
	case "bool":
		return "false"
	case "int", "int64", "int32", "float64":
		return "0"
	case "time.Time":
		return "''"
	case "[]string", "[]int", "[]int64":
		return "[]"
	case "map[string]any":
		return "{}"
	default:
		return "''"
	}
}

func (f Field) FrontendControl() string {
	if f.HasEnum() {
		switch strings.ToLower(strings.TrimSpace(f.EnumMode)) {
		case "multiple":
			return "checkbox-group"
		}
		switch strings.ToLower(strings.TrimSpace(f.EnumDisplay)) {
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
		return "select"
	}
	switch {
	case f.IsTime():
		return "datetime"
	case f.GoType == "bool":
		return "switch"
	case f.GoType == "int" || f.GoType == "int64" || f.GoType == "int32" || f.GoType == "float64":
		return "number"
	case strings.Contains(strings.ToLower(f.JSONName), "description") || strings.Contains(strings.ToLower(f.JSONName), "remark") || strings.Contains(strings.ToLower(f.JSONName), "content") || strings.Contains(strings.ToLower(f.JSONName), "detail") || strings.Contains(strings.ToLower(f.JSONName), "summary"):
		return "textarea"
	default:
		return "input"
	}
}

func (f Field) FrontendDisplayExpression() string {
	prop := "row." + f.JSONName
	if f.HasEnum() {
		return "{{ formatEnumLabel(" + prop + ", " + f.EnumValueMapName() + ") }}"
	}
	switch {
	case f.IsTime():
		return "{{ formatDateTime(" + prop + ") }}"
	case f.GoType == "bool":
		return "{{ " + prop + " ? 'Yes' : 'No' }}"
	case f.GoType == "[]string":
		return "{{ Array.isArray(" + prop + ") ? " + prop + ".join(', ') : ('" + "-" + "') }}"
	case f.GoType == "[]int" || f.GoType == "[]int64":
		return "{{ Array.isArray(" + prop + ") ? " + prop + ".join(', ') : ('" + "-" + "') }}"
	case f.GoType == "map[string]any":
		return "{{ JSON.stringify(" + prop + ") }}"
	default:
		return "{{ " + prop + " || '-' }}"
	}
}

func (f Field) FrontendEditExpression() string {
	prop := "row." + f.JSONName
	if f.HasEnum() {
		if strings.EqualFold(strings.TrimSpace(f.EnumMode), "multiple") {
			return "Array.isArray(" + prop + ") ? " + prop + " : []"
		}
		return prop + " ?? ''"
	}
	switch {
	case f.IsTime():
		return prop + " ?? ''"
	case f.GoType == "bool":
		return "Boolean(" + prop + ")"
	case f.GoType == "int" || f.GoType == "int64" || f.GoType == "int32" || f.GoType == "float64":
		return "Number(" + prop + " ?? 0)"
	default:
		return prop + " ?? ''"
	}
}

func (f Field) FrontendSubmitExpression() string {
	prop := "form." + f.JSONName
	if f.HasEnum() {
		if f.EnumMode == "multiple" {
			return prop + " ?? []"
		}
		return prop
	}
	switch {
	case f.IsTime():
		return prop
	case f.GoType == "bool":
		return "Boolean(" + prop + ")"
	case f.GoType == "int" || f.GoType == "int64" || f.GoType == "int32" || f.GoType == "float64":
		return "Number(" + prop + " ?? 0)"
	default:
		return prop + ".trim()"
	}
}

func (f Field) HasEnum() bool {
	return len(f.EnumValues) > 0 || len(f.EnumOptions) > 0 || strings.TrimSpace(f.EnumKind) != "" || strings.TrimSpace(f.EnumSource) != "" || strings.TrimSpace(f.EnumSourceRef) != "" || strings.TrimSpace(f.EnumDisplay) != ""
}

func (f Field) EnumDisplayOptions() []EnumOption {
	if len(f.EnumOptions) > 0 {
		options := make([]EnumOption, 0, len(f.EnumOptions))
		for _, option := range f.EnumOptions {
			clone := option
			if clone.Label == "" {
				clone.Label = clone.Value
			}
			options = append(options, clone)
		}
		return options
	}
	if len(f.EnumValues) == 0 {
		return nil
	}
	options := make([]EnumOption, 0, len(f.EnumValues))
	for i, value := range f.EnumValues {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		options = append(options, EnumOption{Value: trimmed, Label: trimmed, Order: i + 1})
	}
	return options
}

func (f Field) EnumValueMapName() string {
	if name := strings.TrimSpace(f.JSONName); name != "" {
		return name + "EnumLabelMap"
	}
	if name := strings.TrimSpace(f.GoName); name != "" {
		return strings.ToLower(name) + "EnumLabelMap"
	}
	return "enumLabelMap"
}

type Route struct {
	Method string
	Path   string
}

type ModuleOptions struct {
	Name  string
	Force bool
}

type CRUDOptions struct {
	Name                string
	Fields              []Field
	TableComment        string
	Database            string
	Schema              string
	GenerateFrontend    bool
	GeneratePolicy      bool
	ManifestRoutes      []ManifestRoute
	ManifestMenus       []ManifestMenu
	ManifestPermissions []ManifestPermission
	Force               bool
}

type PluginOptions struct {
	Name  string
	Force bool
}

type ManifestRoute struct {
	Method string
	Path   string
}

type ManifestMenu struct {
	Name       string
	Path       string
	ParentPath string
	Component  string
	Icon       string
	Permission string
	Type       string
	Redirect   string
	Visible    bool
	Enabled    bool
	Sort       int
}

type ManifestPermission struct {
	Object      string
	Action      string
	Description string
}

type ManifestOptions struct {
	Name        string
	Module      string
	Kind        string
	Routes      []ManifestRoute
	Menus       []ManifestMenu
	Permissions []ManifestPermission
	Force       bool
}

type ConfigOptions struct {
	Name   string
	Module string
	Force  bool
}

type PageOptions struct {
	ViewScope  string
	RouteScope string
	PageName   string
	PageSlug   string
	Title      string
	RoutePath  string
	Component  string
	Permission string
	Force      bool
}

type PermissionsOptions struct {
	Scope       string
	Permissions []string
	Force       bool
}

type CRUDData struct {
	Name             string
	PackageName      string
	Entity           string
	EntityLower      string
	EntityPlural     string
	Fields           []Field
	Routes           []Route
	GenerateFrontend bool
	GeneratePolicy   bool
	HasInputTime     bool
	HasModelTime     bool
}

func (d CRUDData) DisplayFields() []Field {
	return nonPrimaryFields(d.Fields)
}

func (d CRUDData) FormFields() []Field {
	return nonPrimaryFields(d.Fields)
}

type PluginData struct {
	Name        string
	PackageName string
	Title       string
	Lower       string
	RoutePrefix string
	ViewName    string
	Force       bool
}

func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func ToSnake(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var out []rune
	prevLower := false
	for _, r := range value {
		switch {
		case r == '-' || r == ' ' || r == '.':
			if len(out) > 0 && out[len(out)-1] != '_' {
				out = append(out, '_')
			}
			prevLower = false
		case unicode.IsUpper(r):
			if prevLower && len(out) > 0 && out[len(out)-1] != '_' {
				out = append(out, '_')
			}
			out = append(out, unicode.ToLower(r))
			prevLower = false
		case r == '_':
			if len(out) > 0 && out[len(out)-1] != '_' {
				out = append(out, '_')
			}
			prevLower = false
		default:
			out = append(out, unicode.ToLower(r))
			prevLower = unicode.IsLetter(r) && unicode.IsLower(r)
		}
	}
	result := strings.Trim(string(out), "_")
	if result == "" {
		return ""
	}
	return result
}

func ToCamel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})
	if len(parts) == 0 {
		return strings.ToUpper(value[:1]) + value[1:]
	}
	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		builder.WriteString(strings.ToUpper(string(runes[0])))
		if len(runes) > 1 {
			builder.WriteString(string(runes[1:]))
		}
	}
	return builder.String()
}

func Pluralize(value string) string {
	value = NormalizeName(value)
	if value == "" {
		return ""
	}
	if strings.HasSuffix(value, "y") && len(value) > 1 {
		prev := rune(value[len(value)-2])
		if !strings.ContainsRune("aeiou", unicode.ToLower(prev)) {
			return value[:len(value)-1] + "ies"
		}
	}
	for _, suffix := range []string{"s", "x", "z", "ch", "sh"} {
		if strings.HasSuffix(value, suffix) {
			return value + "es"
		}
	}
	return value + "s"
}

func ParseFields(fieldSpec, primarySpec, indexSpec, uniqueSpec string) ([]Field, error) {
	specs := splitCSV(fieldSpec)
	if len(specs) == 0 {
		return defaultFields(), nil
	}
	primary := makeSet(primarySpec)
	indexes := makeSet(indexSpec)
	uniques := makeSet(uniqueSpec)
	fields := make([]Field, 0, len(specs))
	for _, spec := range specs {
		parts := strings.SplitN(spec, ":", 2)
		name := strings.TrimSpace(parts[0])
		if name == "" {
			return nil, fmt.Errorf("field name is required in %q", spec)
		}
		typeName := "string"
		if len(parts) == 2 {
			typeName = strings.TrimSpace(parts[1])
		}
		field := Field{
			Name:     name,
			GoName:   ToCamel(name),
			JSONName: ToSnake(name),
			GoType:   mapGoType(typeName),
			Column:   ToSnake(name),
		}
		if field.GoName == "" {
			field.GoName = ToCamel(field.JSONName)
		}
		if field.JSONName == "" {
			field.JSONName = ToSnake(field.GoName)
		}
		if primary[field.Name] || primary[field.JSONName] || primary[field.GoName] {
			field.Primary = true
		}
		if indexes[field.Name] || indexes[field.JSONName] || indexes[field.GoName] {
			field.Index = true
		}
		if uniques[field.Name] || uniques[field.JSONName] || uniques[field.GoName] {
			field.Unique = true
		}
		fields = append(fields, field)
	}
	hasID := false
	for _, field := range fields {
		if field.JSONName == "id" || field.GoName == "ID" {
			hasID = true
			break
		}
	}
	if !hasID {
		fields = append([]Field{{Name: "id", GoName: "ID", JSONName: "id", GoType: "string", Column: "id", Primary: true}}, fields...)
	}
	if len(primary) == 0 {
		for i := range fields {
			if fields[i].JSONName == "id" || fields[i].GoName == "ID" {
				fields[i].Primary = true
				break
			}
		}
	}
	return fields, nil
}

func defaultFields() []Field {
	return []Field{
		{Name: "id", GoName: "ID", JSONName: "id", GoType: "string", Column: "id", Primary: true},
		{Name: "name", GoName: "Name", JSONName: "name", GoType: "string", Column: "name", Index: true},
		{Name: "enabled", GoName: "Enabled", JSONName: "enabled", GoType: "bool", Column: "enabled", Index: true},
	}
}

func splitCSV(input string) []string {
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func makeSet(input string) map[string]bool {
	set := make(map[string]bool)
	for _, item := range splitCSV(input) {
		set[item] = true
		set[NormalizeName(item)] = true
		set[ToSnake(item)] = true
		set[ToCamel(item)] = true
	}
	return set
}

func mapGoType(input string) string {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case "", "string":
		return "string"
	case "int", "int32":
		return "int"
	case "int64", "long":
		return "int64"
	case "float", "float32", "float64":
		return "float64"
	case "bool", "boolean":
		return "bool"
	case "time", "time.time", "datetime", "timestamp":
		return "time.Time"
	case "[]string", "strings":
		return "[]string"
	case "[]int", "[]int32":
		return "[]int"
	case "[]int64":
		return "[]int64"
	case "map", "json", "object", "map[string]any":
		return "map[string]any"
	default:
		if strings.Contains(input, ".") || strings.HasPrefix(input, "[]") {
			return input
		}
		return input
	}
}

func (f Field) JSONTag() string {
	if f.JSONName == "" {
		return "-"
	}
	return f.JSONName + ",omitempty"
}

func (f Field) GormTag() string {
	parts := []string{"column:" + f.Column}
	if f.Primary {
		parts = append(parts, "primaryKey")
		if f.IsAutoIncrementPrimary() {
			parts = append(parts, "autoIncrement")
		}
	}
	if f.GoType == "string" {
		size := f.GormStringSize()
		parts = append(parts, fmt.Sprintf("type:varchar(%d)", size))
		parts = append(parts, fmt.Sprintf("size:%d", size))
	}
	if f.Index {
		parts = append(parts, "index")
	}
	if f.Unique {
		parts = append(parts, "uniqueIndex")
	}
	if comment := strings.TrimSpace(f.Comment); comment != "" {
		parts = append(parts, "comment:"+escapeGormTagValue(comment))
	}
	return strings.Join(parts, ";")
}

func (f Field) GormStringSize() int {
	if f.Primary {
		return 64
	}
	if f.Index || f.Unique {
		return 191
	}
	return 255
}

func escapeGormTagValue(value string) string {
	value = strings.ReplaceAll(value, `\\`, `\\\\`)
	value = strings.ReplaceAll(value, `"`, `\\"`)
	return value
}

func (f Field) IsStringPrimaryKey() bool {
	return f.Primary && f.GoType == "string"
}

func (f Field) IsAutoIncrementPrimary() bool {
	return f.Primary && f.IsIntegerType()
}

func (f Field) IsIntegerType() bool {
	switch f.GoType {
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return true
	default:
		return false
	}
}

func (f Field) IsTime() bool {
	return f.GoType == "time.Time"
}

func (f Field) IsSlice() bool {
	return strings.HasPrefix(f.GoType, "[]")
}

func (d CRUDData) InputFields() []Field {
	return d.Fields
}

func (d CRUDData) HasInputTimeField() bool {
	for _, field := range d.Fields {
		if field.IsTime() {
			return true
		}
	}
	return false
}

func (d CRUDData) HasModelTimeField() bool {
	return true
}

func (d CRUDData) RouteList() []Route {
	return []Route{
		{Method: "GET", Path: "/api/v1/" + d.EntityPlural},
		{Method: "GET", Path: "/api/v1/" + d.EntityPlural + "/:id"},
		{Method: "POST", Path: "/api/v1/" + d.EntityPlural},
		{Method: "PUT", Path: "/api/v1/" + d.EntityPlural + "/:id"},
		{Method: "DELETE", Path: "/api/v1/" + d.EntityPlural + "/:id"},
	}
}

func (d CRUDData) PolicyLines() []string {
	lines := make([]string, 0, len(d.RouteList()))
	for _, route := range d.RouteList() {
		lines = append(lines, fmt.Sprintf("p, admin, %s, %s", route.Path, route.Method))
	}
	return lines
}

func (d CRUDData) FrontendRoutePath() string {
	return "/" + d.EntityPlural
}

func (d CRUDData) FrontendViewPath() string {
	return "view/" + d.EntityPlural + "/index"
}

func (d CRUDData) FrontendApiPath() string {
	return "@/api/" + d.EntityLower
}

func (d CRUDData) FrontendRouterFile() string {
	return "web/src/router/modules/" + d.EntityLower + ".ts"
}

func (d CRUDData) FrontendViewFile() string {
	return "web/src/views/" + d.EntityLower + "/index.vue"
}

func (d CRUDData) FrontendApiFile() string {
	return "web/src/api/" + d.EntityLower + ".ts"
}

func (d CRUDData) FrontendName() string {
	return d.Entity + "View"
}

func (d CRUDData) HasTimeImport() bool {
	return true
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

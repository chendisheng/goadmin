package generate

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	codegenmerger "goadmin/codegen/merger"
	codegenpostprocess "goadmin/codegen/postprocess"
)

type Generator struct {
	Root       string
	PolicyPath string
}

func renderManifestMenusOrDefault(menus []ManifestMenu, defaults []ManifestMenu) string {
	if len(menus) > 0 {
		return renderManifestMenus(menus)
	}
	return renderManifestMenus(defaults)
}

func buildModuleRenderData(name string, force bool) moduleRenderData {
	entityLower := ToSnake(name)
	if entityLower == "" {
		entityLower = NormalizeName(name)
	}
	if entityLower == "" {
		entityLower = "module"
	}
	kind := "business-module"
	if entityLower == "auth" {
		kind = "core-module"
	}
	return moduleRenderData{
		Name:                entityLower,
		PackageName:         entityLower,
		EntityLower:         entityLower,
		EntityPlural:        Pluralize(entityLower),
		Title:               ToCamel(entityLower),
		Module:              entityLower,
		Kind:                kind,
		ManifestRoutes:      renderCRUDManifestRoutes(buildCRUDManifestRoutes(Pluralize(entityLower))),
		ManifestMenus:       "",
		ManifestPermissions: "",
		Force:               force,
	}
}

func renderListRequestFields() string {
	return "\tKeyword string `json:\"keyword,omitempty\" form:\"keyword\"`\n\tPage int `json:\"page,omitempty\" form:\"page\"`\n\tPageSize int `json:\"page_size,omitempty\" form:\"page_size\"`\n"
}

func renderRequestFieldBlock(fields []Field) string {
	if len(fields) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, field := range fields {
		builder.WriteString("\t")
		builder.WriteString(field.GoName)
		builder.WriteString(" ")
		builder.WriteString(field.GoType)
		builder.WriteString(" `json:\"")
		builder.WriteString(field.JSONName)
		builder.WriteString(",omitempty\" form:\"")
		builder.WriteString(field.JSONName)
		builder.WriteString("\"`\n")
	}
	return builder.String()
}

func nonPrimaryFields(fields []Field) []Field {
	result := make([]Field, 0, len(fields))
	for _, field := range fields {
		if field.Primary {
			continue
		}
		result = append(result, field)
	}
	return result
}

func renderResponseFieldBlock(fields []Field) string {
	if len(fields) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, field := range fields {
		builder.WriteString("\t")
		builder.WriteString(field.GoName)
		builder.WriteString(" ")
		builder.WriteString(field.GoType)
		builder.WriteString(" `json:\"")
		builder.WriteString(field.JSONName)
		builder.WriteString(",omitempty\"`\n")
	}
	return builder.String()
}

func renderCloneBlock(fields []Field) string {
	var builder strings.Builder
	for _, field := range fields {
		if !field.IsSlice() {
			continue
		}
		elementType := strings.TrimPrefix(field.GoType, "[]")
		builder.WriteString("\tif m.")
		builder.WriteString(field.GoName)
		builder.WriteString(" != nil {\n")
		if isCloneableElement(elementType) {
			builder.WriteString("\t\tclone.")
			builder.WriteString(field.GoName)
			builder.WriteString(" = make([]")
			builder.WriteString(elementType)
			builder.WriteString(", 0, len(m.")
			builder.WriteString(field.GoName)
			builder.WriteString("))\n")
			builder.WriteString("\t\tfor _, child := range m.")
			builder.WriteString(field.GoName)
			builder.WriteString(" {\n")
			builder.WriteString("\t\t\tclone.")
			builder.WriteString(field.GoName)
			builder.WriteString(" = append(clone.")
			builder.WriteString(field.GoName)
			builder.WriteString(", child.Clone())\n")
			builder.WriteString("\t\t}\n")
		} else {
			builder.WriteString("\t\tclone.")
			builder.WriteString(field.GoName)
			builder.WriteString(" = append([]")
			builder.WriteString(elementType)
			builder.WriteString("(nil), m.")
			builder.WriteString(field.GoName)
			builder.WriteString("...)\n")
		}
		builder.WriteString("\t}\n")
	}
	return builder.String()
}

func isCloneableElement(elementType string) bool {
	if elementType == "" {
		return false
	}
	r := rune(elementType[0])
	return unicode.IsUpper(r)
}

func renderResponseAssignments(fields []Field) string {
	if len(fields) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, field := range fields {
		builder.WriteString("\t\t")
		builder.WriteString(field.GoName)
		builder.WriteString(": item.")
		builder.WriteString(field.GoName)
		builder.WriteString(",\n")
	}
	return builder.String()
}

func renderAssignmentBlock(fields []Field, target string) string {
	if len(fields) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, field := range fields {
		builder.WriteString("\t")
		builder.WriteString(target)
		builder.WriteString(".")
		builder.WriteString(field.GoName)
		builder.WriteString(" = input.")
		builder.WriteString(field.GoName)
		builder.WriteString("\n")
	}
	return builder.String()
}

type scaffoldData struct {
	Name                string
	PackageName         string
	Entity              string
	Title               string
	EntityLower         string
	EntityPlural        string
	Module              string
	Kind                string
	RouteTitleKey       string
	RouteTitleDefault   string
	TableComment        string
	Database            string
	Schema              string
	Fields              []Field
	ModelFields         string
	CloneBlock          string
	CommandFields       string
	RequestFields       string
	ListRequestFields   string
	ResponseFields      string
	ResponseAssignments string
	CreateAssignments   string
	UpdateAssignments   string
	HasInputTime        bool
	GenerateFrontend    bool
	GeneratePolicy      bool
	ManifestRoutes      string
	ManifestMenus       string
	ManifestPermissions string
	Force               bool
}

type manifestRenderData struct {
	Name                string
	Module              string
	EntityLower         string
	Kind                string
	ManifestRoutes      string
	ManifestMenus       string
	ManifestPermissions string
	Force               bool
}

type configRenderData struct {
	Name   string
	Module string
	Kind   string
	Force  bool
}

type pluginRenderData struct {
	Name        string
	PackageName string
	EntityLower string
	Title       string
	RoutePrefix string
	ViewPath    string
	Force       bool
}

type pageRenderData struct {
	ViewScope    string
	PageSlug     string
	PageName     string
	Title        string
	TitleKey     string
	TitleDefault string
	RoutePath    string
	RouteName    string
	Component    string
	Permission   string
	Force        bool
}

type moduleRenderData struct {
	Name                string
	PackageName         string
	EntityLower         string
	EntityPlural        string
	Title               string
	Module              string
	Kind                string
	ManifestRoutes      string
	ManifestMenus       string
	ManifestPermissions string
	Force               bool
}

func New(root string) *Generator {
	clean := strings.TrimSpace(root)
	if clean == "" {
		clean = "."
	}
	return &Generator{
		Root:       clean,
		PolicyPath: filepath.Join(clean, "server", "core", "auth", "casbin", "adapter", "policy.csv"),
	}
}

func (g *Generator) GenerateModule(opts ModuleOptions) error {
	if g == nil {
		return errors.New("generator is nil")
	}
	data := buildModuleRenderData(opts.Name, opts.Force)
	return g.writeModuleScaffold(data)
}

func (g *Generator) GenerateCRUD(opts CRUDOptions) error {
	if g == nil {
		return errors.New("generator is nil")
	}
	fields := opts.Fields
	if len(fields) == 0 {
		fields, _ = ParseFields("", "", "", "")
	}
	data := buildScaffoldData(opts.Name, fields, opts.TableComment, opts.Database, opts.Schema, opts.GenerateFrontend, opts.GeneratePolicy, opts.ManifestRoutes, opts.ManifestMenus, opts.ManifestPermissions, opts.Force)
	if err := g.writeScaffold(data); err != nil {
		return err
	}
	if err := g.refreshBootstrapRegistry(); err != nil {
		return err
	}
	if opts.GeneratePolicy {
		if err := g.appendPolicyLines(data.PolicyLines()); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) GeneratePlugin(opts PluginOptions) error {
	if g == nil {
		return errors.New("generator is nil")
	}
	data := buildPluginRenderData(opts.Name, opts.Force)
	if err := g.writePluginScaffold(data); err != nil {
		return err
	}
	if err := g.appendPolicyLines([]string{
		fmt.Sprintf("p, admin, %s, GET", data.RoutePrefix+"/ping"),
	}); err != nil {
		return err
	}
	return nil
}

func (g *Generator) GenerateManifest(opts ManifestOptions) error {
	if g == nil {
		return errors.New("generator is nil")
	}
	data := buildManifestRenderData(opts)
	return g.writeGoOrText(filepath.Join(g.Root, "server", "modules", data.Name, "manifest.yaml"), manifestTemplate, data, data.Force)
}

func (g *Generator) GenerateConfig(opts ConfigOptions) error {
	if g == nil {
		return errors.New("generator is nil")
	}
	data := buildConfigRenderData(opts)
	fileName := "config." + data.Name + ".yaml"
	if data.Name == "" {
		fileName = "config.generated.yaml"
	}
	return g.writeGoOrText(filepath.Join(g.Root, "server", "config", fileName), configTemplate, data, data.Force)
}

func (g *Generator) GeneratePage(opts PageOptions) error {
	if g == nil {
		return errors.New("generator is nil")
	}
	data := buildPageRenderData(opts)
	viewFile := filepath.Join(g.Root, "web", "src", "views", data.ViewScope, data.PageSlug+".vue")
	routeFile := filepath.Join(g.Root, "web", "src", "router", "modules", data.ViewScope+"-"+data.PageSlug+".ts")
	if err := g.writeGoOrText(viewFile, pageViewTemplate, data, data.Force); err != nil {
		return err
	}
	if err := g.writeGoOrText(routeFile, pageRouterTemplate, data, data.Force); err != nil {
		return err
	}
	return nil
}

func (g *Generator) AppendPolicyLines(lines []string) error {
	return g.appendPolicyLines(lines)
}

func buildScaffoldData(name string, fields []Field, tableComment, database, schemaName string, frontend, policy bool, manifestRoutes []ManifestRoute, manifestMenus []ManifestMenu, manifestPermissions []ManifestPermission, force bool) scaffoldData {
	entity := ToCamel(name)
	entityLower := ToSnake(name)
	if entityLower == "" {
		entityLower = NormalizeName(name)
	}
	if entityLower == "" {
		entityLower = "entity"
	}
	if entity == "" {
		entity = ToCamel(entityLower)
	}
	packageName := entityLower
	entityPlural := Pluralize(entityLower)
	sanitized := sanitizeFields(fields)
	requestFields := nonPrimaryFields(sanitized)
	return scaffoldData{
		Name:                entityLower,
		PackageName:         packageName,
		Entity:              entity,
		Title:               entity,
		EntityLower:         entityLower,
		EntityPlural:        entityPlural,
		Module:              entityLower,
		Kind:                "business-module",
		RouteTitleKey:       buildLocaleKey("route", entityLower),
		RouteTitleDefault:   entity + " Management",
		TableComment:        strings.TrimSpace(tableComment),
		Database:            strings.TrimSpace(database),
		Schema:              strings.TrimSpace(schemaName),
		Fields:              sanitized,
		ModelFields:         renderFieldBlock(sanitized, true),
		CloneBlock:          renderCloneBlock(sanitized),
		CommandFields:       renderFieldBlock(requestFields, false),
		RequestFields:       renderRequestFieldBlock(requestFields),
		ListRequestFields:   renderListRequestFields(),
		ResponseFields:      renderResponseFieldBlock(sanitized),
		ResponseAssignments: renderResponseAssignments(sanitized),
		CreateAssignments:   renderAssignmentBlock(requestFields, "item"),
		UpdateAssignments:   renderAssignmentBlock(requestFields, "item"),
		HasInputTime:        hasTimeField(sanitized),
		GenerateFrontend:    frontend,
		GeneratePolicy:      policy,
		ManifestRoutes:      renderManifestRoutesOrDefault(manifestRoutes, buildCRUDManifestRoutes(entityPlural)),
		ManifestMenus:       renderManifestMenusOrDefault(manifestMenus, buildCRUDManifestMenus(entityLower, entityPlural, frontend)),
		ManifestPermissions: renderManifestPermissionsOrDefault(manifestPermissions, buildCRUDManifestPermissions(entityLower, entityPlural)),
		Force:               force,
	}
}

func buildManifestRenderData(opts ManifestOptions) manifestRenderData {
	name := ToSnake(opts.Name)
	if name == "" {
		name = NormalizeName(opts.Module)
	}
	if name == "" {
		name = "manifest"
	}
	module := strings.TrimSpace(opts.Module)
	if module == "" {
		module = name
	}
	kind := strings.TrimSpace(opts.Kind)
	if kind == "" {
		kind = "business-module"
	}
	return manifestRenderData{
		Name:                name,
		Module:              module,
		EntityLower:         name,
		Kind:                kind,
		ManifestRoutes:      renderManifestRoutes(opts.Routes),
		ManifestMenus:       renderManifestMenus(opts.Menus),
		ManifestPermissions: renderManifestPermissions(opts.Permissions),
		Force:               opts.Force,
	}
}

func buildConfigRenderData(opts ConfigOptions) configRenderData {
	name := ToSnake(opts.Name)
	if name == "" {
		name = NormalizeName(opts.Module)
	}
	if name == "" {
		name = "generated"
	}
	module := strings.TrimSpace(opts.Module)
	if module == "" {
		module = name
	}
	return configRenderData{Name: name, Module: module, Kind: "config", Force: opts.Force}
}

func (d scaffoldData) PrimaryField() Field {
	for _, field := range d.Fields {
		if field.Primary {
			return field
		}
	}
	return Field{}
}

func (d scaffoldData) NeedsPrimaryIDGeneration() bool {
	return d.PrimaryField().IsStringPrimaryKey()
}

func (d scaffoldData) NeedsStringsImport() bool {
	if d.NeedsPrimaryIDGeneration() {
		return true
	}
	for _, field := range d.Fields {
		if field.Primary {
			continue
		}
		if field.GoType == "string" {
			return true
		}
	}
	return false
}

func (d scaffoldData) DisplayFields() []Field {
	return nonPrimaryFields(d.Fields)
}

func (d scaffoldData) FormFields() []Field {
	return nonPrimaryFields(d.Fields)
}

func (d scaffoldData) SearchFields() []Field {
	result := make([]Field, 0, len(d.Fields))
	for _, field := range d.Fields {
		if field.Primary {
			continue
		}
		if field.GoType != "string" {
			continue
		}
		result = append(result, field)
	}
	return result
}

func (d scaffoldData) SearchFilterBlock() string {
	fields := d.SearchFields()
	if len(fields) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("\tif kw := strings.TrimSpace(strings.ToLower(keyword)); kw != \"\" {\n")
	builder.WriteString("\t\tlike := \"%\" + kw + \"%\"\n")
	builder.WriteString("\t\tbase = base.Where(\n")
	builder.WriteString("\t\t\t\"")
	for i, field := range fields {
		if i > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteString("LOWER(")
		builder.WriteString(field.Column)
		builder.WriteString(") LIKE ?")
	}
	builder.WriteString("\",\n")
	for range fields {
		builder.WriteString("\t\t\tlike,\n")
	}
	builder.WriteString("\t\t)\n")
	builder.WriteString("\t}\n")
	return builder.String()
}

func (d scaffoldData) PolicyLines() []string {
	routes := []Route{
		{Method: "GET", Path: "/api/v1/" + d.EntityPlural},
		{Method: "GET", Path: "/api/v1/" + d.EntityPlural + "/:id"},
		{Method: "POST", Path: "/api/v1/" + d.EntityPlural},
		{Method: "PUT", Path: "/api/v1/" + d.EntityPlural + "/:id"},
		{Method: "DELETE", Path: "/api/v1/" + d.EntityPlural + "/:id"},
	}
	lines := make([]string, 0, len(routes))
	for _, route := range routes {
		lines = append(lines, fmt.Sprintf("p, admin, %s, %s", route.Path, route.Method))
	}
	return lines
}

func (d scaffoldData) SQLTableName() string {
	if strings.TrimSpace(d.EntityLower) != "" {
		return d.EntityLower
	}
	if strings.TrimSpace(d.Entity) != "" {
		return ToSnake(d.Entity)
	}
	return "entity"
}

func (d scaffoldData) SQLSchema() string {
	tableName := d.SQLTableName()
	definitions := make([]string, 0, len(d.Fields)+4)
	primaryColumns := make([]string, 0, 1)
	for _, field := range d.Fields {
		definitions = append(definitions, "  "+sqlColumnDefinition(field))
		if field.Primary {
			primaryColumns = append(primaryColumns, "`"+field.Column+"`")
		}
	}
	if len(primaryColumns) > 0 {
		definitions = append(definitions, "  PRIMARY KEY ("+strings.Join(primaryColumns, ", ")+")")
	}
	for _, field := range d.Fields {
		if field.Primary {
			continue
		}
		if field.Unique {
			definitions = append(definitions, fmt.Sprintf("  UNIQUE KEY `idx_%s_%s` (`%s`)", tableName, field.Column, field.Column))
			continue
		}
		if field.Index {
			definitions = append(definitions, fmt.Sprintf("  KEY `idx_%s_%s` (`%s`)", tableName, field.Column, field.Column))
		}
	}
	if len(definitions) == 0 {
		definitions = append(definitions,
			"  `id` bigint unsigned NOT NULL AUTO_INCREMENT",
			"  PRIMARY KEY (`id`)",
		)
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("-- Auto-generated schema for %s\n", tableName))
	if database := strings.TrimSpace(d.Database); database != "" {
		builder.WriteString(fmt.Sprintf("-- Database: %s\n", database))
	}
	if schemaName := strings.TrimSpace(d.Schema); schemaName != "" {
		builder.WriteString(fmt.Sprintf("-- Schema: %s\n", schemaName))
	}
	if comment := strings.TrimSpace(d.TableComment); comment != "" {
		builder.WriteString(fmt.Sprintf("-- Table Comment: %s\n", comment))
	}
	builder.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n", tableName))
	builder.WriteString(strings.Join(definitions, ",\n"))
	builder.WriteString("\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	if comment := strings.TrimSpace(d.TableComment); comment != "" {
		builder.WriteString(fmt.Sprintf(" COMMENT='%s'", escapeSQLComment(comment)))
	}
	builder.WriteString(";\n")
	return builder.String()
}

func sqlColumnDefinition(field Field) string {
	column := "`" + field.Column + "`"
	switch field.GoType {
	case "string":
		return renderSQLColumnWithComment(fmt.Sprintf("%s varchar(%d) NOT NULL", column, field.GormStringSize()), field.Comment)
	case "bool":
		return renderSQLColumnWithComment(fmt.Sprintf("%s tinyint(1) NOT NULL DEFAULT 0", column), field.Comment)
	case "int":
		if field.Primary {
			return renderSQLColumnWithComment(fmt.Sprintf("%s bigint unsigned NOT NULL AUTO_INCREMENT", column), field.Comment)
		}
		return renderSQLColumnWithComment(fmt.Sprintf("%s int NOT NULL DEFAULT 0", column), field.Comment)
	case "int32":
		if field.Primary {
			return renderSQLColumnWithComment(fmt.Sprintf("%s bigint unsigned NOT NULL AUTO_INCREMENT", column), field.Comment)
		}
		return renderSQLColumnWithComment(fmt.Sprintf("%s int NOT NULL DEFAULT 0", column), field.Comment)
	case "int64":
		if field.Primary {
			return renderSQLColumnWithComment(fmt.Sprintf("%s bigint unsigned NOT NULL AUTO_INCREMENT", column), field.Comment)
		}
		return renderSQLColumnWithComment(fmt.Sprintf("%s bigint NOT NULL DEFAULT 0", column), field.Comment)
	case "float64":
		return renderSQLColumnWithComment(fmt.Sprintf("%s decimal(18,4) NOT NULL DEFAULT 0", column), field.Comment)
	case "time.Time":
		return renderSQLColumnWithComment(fmt.Sprintf("%s datetime NULL", column), field.Comment)
	case "[]string", "[]int", "[]int64", "map[string]any":
		return renderSQLColumnWithComment(fmt.Sprintf("%s json NULL", column), field.Comment)
	default:
		return renderSQLColumnWithComment(fmt.Sprintf("%s text NULL", column), field.Comment)
	}
}

func renderSQLColumnWithComment(definition string, comment string) string {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return definition
	}
	return definition + fmt.Sprintf(" COMMENT '%s'", escapeSQLComment(comment))
}

func escapeSQLComment(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

type crudRouteSpec struct {
	Method string
	Path   string
	Action string
}

func buildCRUDManifestRoutes(entityPlural string) []crudRouteSpec {
	base := "/api/v1/" + entityPlural
	return []crudRouteSpec{
		{Method: "GET", Path: base, Action: "list"},
		{Method: "GET", Path: base + "/:id", Action: "view"},
		{Method: "POST", Path: base, Action: "create"},
		{Method: "PUT", Path: base + "/:id", Action: "update"},
		{Method: "DELETE", Path: base + "/:id", Action: "delete"},
	}
}

func buildCRUDManifestMenus(entityLower, entityPlural string, frontend bool) []ManifestMenu {
	if !frontend {
		return nil
	}
	rootPath := "/" + entityPlural
	listPath := rootPath + "/list"
	if entityPlural == "" {
		rootPath = "/" + entityLower
		listPath = rootPath
	}
	return []ManifestMenu{
		{Name: ToCamel(entityPlural), TitleKey: buildLocaleKey("route", entityLower), TitleDefault: ToCamel(entityPlural), Path: rootPath, Component: "Layout", Icon: "menu", Permission: entityLower + ":view", Type: "directory", Redirect: listPath, Visible: true, Enabled: true, Sort: 1},
		{Name: "List", TitleKey: buildLocaleKey("route", entityLower, "list"), TitleDefault: "List", Path: listPath, Component: "view/" + entityLower + "/index", Icon: "menu", Permission: entityLower + ":list", Type: "menu", Visible: true, Enabled: true, Sort: 2},
	}
}

func buildCRUDManifestPermissions(entityLower, entityPlural string) []ManifestPermission {
	if entityPlural == "" {
		entityPlural = entityLower
	}
	label := ToCamel(entityPlural)
	return []ManifestPermission{
		{Object: entityLower, Action: "list", Description: "List " + label},
		{Object: entityLower, Action: "view", Description: "View " + label},
		{Object: entityLower, Action: "create", Description: "Create " + label},
		{Object: entityLower, Action: "update", Description: "Update " + label},
		{Object: entityLower, Action: "delete", Description: "Delete " + label},
	}
}

func renderManifestRoutes(routes []ManifestRoute) string {
	if len(routes) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("routes:\n")
	for _, route := range routes {
		builder.WriteString("  - method: ")
		builder.WriteString(route.Method)
		builder.WriteString("\n")
		builder.WriteString("    path: ")
		builder.WriteString(route.Path)
		builder.WriteString("\n")
	}
	return builder.String()
}

func renderCRUDManifestRoutes(routes []crudRouteSpec) string {
	if len(routes) == 0 {
		return ""
	}
	manifestRoutes := make([]ManifestRoute, 0, len(routes))
	for _, route := range routes {
		manifestRoutes = append(manifestRoutes, ManifestRoute{Method: route.Method, Path: route.Path})
	}
	return renderManifestRoutes(manifestRoutes)
}

func renderManifestRoutesOrDefault(routes []ManifestRoute, defaults []crudRouteSpec) string {
	if len(routes) > 0 {
		return renderManifestRoutes(routes)
	}
	return renderCRUDManifestRoutes(defaults)
}

func renderManifestMenus(menus []ManifestMenu) string {
	if len(menus) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("menus:\n")
	for _, menu := range menus {
		builder.WriteString("  - name: ")
		builder.WriteString(menu.Name)
		builder.WriteString("\n")
		if menu.TitleKey != "" {
			builder.WriteString("    title_key: ")
			builder.WriteString(menu.TitleKey)
			builder.WriteString("\n")
		}
		if menu.TitleDefault != "" {
			builder.WriteString("    title_default: ")
			builder.WriteString(menu.TitleDefault)
			builder.WriteString("\n")
		}
		builder.WriteString("    path: ")
		builder.WriteString(menu.Path)
		builder.WriteString("\n")
		if menu.ParentPath != "" {
			builder.WriteString("    parent_path: ")
			builder.WriteString(menu.ParentPath)
			builder.WriteString("\n")
		}
		builder.WriteString("    component: ")
		builder.WriteString(menu.Component)
		builder.WriteString("\n")
		if menu.Icon != "" {
			builder.WriteString("    icon: ")
			builder.WriteString(menu.Icon)
			builder.WriteString("\n")
		}
		if menu.Permission != "" {
			builder.WriteString("    permission: ")
			builder.WriteString(menu.Permission)
			builder.WriteString("\n")
		}
		builder.WriteString("    type: ")
		builder.WriteString(menu.Type)
		builder.WriteString("\n")
		if menu.Redirect != "" {
			builder.WriteString("    redirect: ")
			builder.WriteString(menu.Redirect)
			builder.WriteString("\n")
		}
		builder.WriteString("    visible: ")
		builder.WriteString(fmt.Sprintf("%t", menu.Visible))
		builder.WriteString("\n")
		builder.WriteString("    enabled: ")
		builder.WriteString(fmt.Sprintf("%t", menu.Enabled))
		builder.WriteString("\n")
		builder.WriteString("    sort: ")
		builder.WriteString(fmt.Sprintf("%d", menu.Sort))
		builder.WriteString("\n")
	}
	return builder.String()
}

func renderManifestPermissions(permissions []ManifestPermission) string {
	if len(permissions) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("permissions:\n")
	for _, permission := range permissions {
		builder.WriteString("  - object: ")
		builder.WriteString(permission.Object)
		builder.WriteString("\n")
		builder.WriteString("    action: ")
		builder.WriteString(permission.Action)
		builder.WriteString("\n")
		builder.WriteString("    description: ")
		builder.WriteString(permission.Description)
		builder.WriteString("\n")
	}
	return builder.String()
}

func renderManifestPermissionsOrDefault(permissions []ManifestPermission, defaults []ManifestPermission) string {
	if len(permissions) > 0 {
		return renderManifestPermissions(permissions)
	}
	return renderManifestPermissions(defaults)
}

func buildLocaleKey(prefix string, parts ...string) string {
	prefix = strings.TrimSpace(prefix)
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		segment := NormalizeName(part)
		if segment == "" {
			continue
		}
		segments = append(segments, segment)
	}
	if prefix == "" {
		return strings.Join(segments, "_")
	}
	if len(segments) == 0 {
		return prefix
	}
	return prefix + "." + strings.Join(segments, "_")
}

func buildPluginRenderData(name string, force bool) pluginRenderData {
	entityLower := ToSnake(name)
	if entityLower == "" {
		entityLower = NormalizeName(name)
	}
	if entityLower == "" {
		entityLower = "plugin"
	}
	return pluginRenderData{
		Name:        entityLower,
		PackageName: entityLower,
		EntityLower: entityLower,
		Title:       ToCamel(entityLower),
		RoutePrefix: "/plugins/" + entityLower,
		ViewPath:    "view/plugin/" + entityLower + "/index",
		Force:       force,
	}
}

func buildPageRenderData(opts PageOptions) pageRenderData {
	viewScope := NormalizeName(opts.ViewScope)
	if viewScope == "" {
		viewScope = "page"
	}
	routeScope := NormalizeName(opts.RouteScope)
	if routeScope == "" {
		routeScope = viewScope
	}
	pageSlug := ToSnake(opts.PageSlug)
	if pageSlug == "" {
		pageSlug = ToSnake(opts.PageName)
	}
	if pageSlug == "" {
		pageSlug = "index"
	}
	title := strings.TrimSpace(opts.Title)
	if title == "" {
		title = strings.TrimSpace(opts.PageName)
	}
	if title == "" {
		title = ToCamel(pageSlug)
	}
	titleKey := strings.TrimSpace(opts.TitleKey)
	if titleKey == "" {
		titleKey = buildLocaleKey("route", routeScope, pageSlug)
	}
	titleDefault := strings.TrimSpace(opts.TitleDefault)
	if titleDefault == "" {
		titleDefault = title
	}
	routePath := normalizePath(opts.RoutePath)
	if routePath == "" {
		routePath = normalizePath("/" + routeScope + "/" + pageSlug)
	}
	component := normalizeViewComponent(opts.Component)
	if component == "" {
		component = viewScope + "/" + pageSlug
	}
	routeName := strings.TrimSpace(opts.RouteScope)
	if routeName == "" {
		routeName = viewScope
	}
	routeName = NormalizeName(routeName)
	if routeName == "" {
		routeName = viewScope
	}
	if routeName != "" {
		routeName += "-"
	}
	routeName += pageSlug
	return pageRenderData{
		ViewScope:    viewScope,
		PageSlug:     pageSlug,
		PageName:     strings.TrimSpace(opts.PageName),
		Title:        titleDefault,
		TitleKey:     titleKey,
		TitleDefault: titleDefault,
		RoutePath:    routePath,
		RouteName:    routeName,
		Component:    component,
		Permission:   strings.TrimSpace(opts.Permission),
		Force:        opts.Force,
	}
}

func normalizeViewComponent(component string) string {
	trimmed := strings.TrimSpace(component)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.TrimPrefix(trimmed, "@/views/")
	trimmed = strings.TrimPrefix(trimmed, "views/")
	trimmed = strings.TrimPrefix(trimmed, "view/")
	trimmed = strings.Trim(trimmed, "/")
	return trimmed
}

func normalizePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	if trimmed != "/" {
		trimmed = strings.TrimRight(trimmed, "/")
	}
	return trimmed
}

func (g *Generator) writeScaffold(data scaffoldData) error {
	base := filepath.Join(g.Root, "server", "modules", data.EntityLower)
	files := map[string]string{
		filepath.Join(base, "bootstrap.go"):                                          bootstrapTemplate,
		filepath.Join(base, "module.go"):                                             moduleTemplate,
		filepath.Join(base, "manifest.yaml"):                                         manifestTemplate,
		filepath.Join(base, "locales", "zh-CN", data.EntityLower+".yaml"):            crudLocaleZhTemplate,
		filepath.Join(base, "locales", "en-US", data.EntityLower+".yaml"):            crudLocaleEnTemplate,
		filepath.Join(base, "domain", "model", data.EntityLower+".go"):               modelTemplate,
		filepath.Join(base, "domain", "repository", "repository.go"):                 repositoryTemplate,
		filepath.Join(base, "application", "command", data.EntityLower+".go"):        commandTemplate,
		filepath.Join(base, "application", "query", data.EntityLower+".go"):          queryTemplate,
		filepath.Join(base, "application", "service", "service.go"):                  serviceTemplate,
		filepath.Join(base, "infrastructure", "repo", "gorm.go"):                     gormRepositoryTemplate,
		filepath.Join(base, "transport", "http", "request", data.EntityLower+".go"):  requestTemplate,
		filepath.Join(base, "transport", "http", "response", data.EntityLower+".go"): responseTemplate,
		filepath.Join(base, "transport", "http", "handler", "handler.go"):            handlerTemplate,
		filepath.Join(base, "transport", "http", "router.go"):                        routerTemplate,
	}
	for path, tmpl := range files {
		if err := g.writeGoOrText(path, tmpl, data, data.Force); err != nil {
			return err
		}
	}
	sqlPath := filepath.Join(base, "schema.sql")
	if err := g.writeGoOrText(sqlPath, "{{.SQLSchema}}", data, data.Force); err != nil {
		return err
	}
	if data.GenerateFrontend {
		frontendFiles := map[string]string{
			filepath.Join(g.Root, "web", "src", "i18n", "locales", "zh-CN", data.EntityLower+".json"): crudFrontendLocaleZhTemplate,
			filepath.Join(g.Root, "web", "src", "i18n", "locales", "en-US", data.EntityLower+".json"): crudFrontendLocaleEnTemplate,
			filepath.Join(g.Root, "web", "src", "api", data.EntityLower+".ts"):                        frontendApiTemplate,
			filepath.Join(g.Root, "web", "src", "router", "modules", data.EntityLower+".ts"):          frontendRouterTemplate,
			filepath.Join(g.Root, "web", "src", "views", data.EntityLower, "index.vue"):               frontendViewTemplate,
		}
		for path, tmpl := range frontendFiles {
			if err := g.writeGoOrText(path, tmpl, data, data.Force); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Generator) refreshBootstrapRegistry() error {
	if g == nil {
		return errors.New("generator is nil")
	}
	modulesDir := filepath.Join(g.Root, "server", "modules")
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return fmt.Errorf("scan modules dir: %w", err)
	}
	moduleNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		bootstrapPath := filepath.Join(modulesDir, name, "bootstrap.go")
		content, err := os.ReadFile(bootstrapPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("read %s: %w", bootstrapPath, err)
		}
		if codegenpostprocess.HasGeneratedMarkers(bootstrapPath, content) {
			moduleNames = append(moduleNames, name)
		}
	}
	sort.Strings(moduleNames)
	var builder strings.Builder
	builder.WriteString("package bootstrap\n\n")
	if len(moduleNames) > 0 {
		builder.WriteString("import (\n")
		for _, name := range moduleNames {
			builder.WriteString("\t\"")
			builder.WriteString("goadmin/modules/")
			builder.WriteString(name)
			builder.WriteString("\"\n")
		}
		builder.WriteString(")\n\n")
	}
	builder.WriteString("func generatedModules() []Module {\n")
	if len(moduleNames) == 0 {
		builder.WriteString("\treturn nil\n")
	} else {
		builder.WriteString("\treturn []Module{\n")
		for _, name := range moduleNames {
			builder.WriteString("\t\t")
			builder.WriteString(name)
			builder.WriteString(".NewBootstrap(),\n")
		}
		builder.WriteString("\t}\n")
	}
	builder.WriteString("}\n")
	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return fmt.Errorf("format generated bootstrap registry: %w\nsource:\n%s", err, builder.String())
	}
	return g.writeFile(filepath.Join(g.Root, "server", "core", "bootstrap", "modules_gen.go"), formatted, true)
}

func (g *Generator) writeModuleScaffold(data moduleRenderData) error {
	base := filepath.Join(g.Root, "server", "modules", data.Name)
	files := map[string]string{
		filepath.Join(base, "module.go"):                                  moduleTemplate,
		filepath.Join(base, "manifest.yaml"):                              manifestTemplate,
		filepath.Join(base, "locales", "zh-CN", data.EntityLower+".yaml"): moduleLocaleZhTemplate,
		filepath.Join(base, "locales", "en-US", data.EntityLower+".yaml"): moduleLocaleEnTemplate,
	}
	for path, tmpl := range files {
		if err := g.writeGoOrText(path, tmpl, data, data.Force); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) writePluginScaffold(data pluginRenderData) error {
	base := filepath.Join(g.Root, "server", "plugin", "builtin", data.EntityLower)
	files := map[string]string{
		filepath.Join(base, data.EntityLower+".go"):                                           pluginTemplate,
		filepath.Join(base, "locales", "zh-CN", "plugin.yaml"):                                pluginLocaleZhTemplate,
		filepath.Join(base, "locales", "en-US", "plugin.yaml"):                                pluginLocaleEnTemplate,
		filepath.Join(g.Root, "web", "src", "plugins", data.EntityLower+".ts"):                pluginFrontendTemplate,
		filepath.Join(g.Root, "web", "src", "views", "plugin", data.EntityLower, "index.vue"): pluginViewTemplate,
	}
	for path, tmpl := range files {
		if err := g.writeGoOrText(path, tmpl, data, data.Force); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) appendPolicyLines(lines []string) error {
	if g == nil || len(lines) == 0 {
		return nil
	}
	cleaned := codegenmerger.UniqueLines(lines)
	if len(cleaned) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(g.PolicyPath), 0o755); err != nil {
		return fmt.Errorf("create policy directory: %w", err)
	}
	existing := make(map[string]struct{})
	var kept []string
	if content, err := os.ReadFile(g.PolicyPath); err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if _, ok := existing[trimmed]; ok {
				continue
			}
			existing[trimmed] = struct{}{}
			kept = append(kept, trimmed)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read policy file: %w", err)
	}
	merged := append([]string(nil), kept...)
	merged = append(merged, cleaned...)
	output := strings.Join(codegenpostprocess.NormalizePolicyLines(merged), "\n")
	if output != "" {
		output += "\n"
	}
	return os.WriteFile(g.PolicyPath, []byte(output), 0o644)
}

func (g *Generator) writeGoOrText(path, tmpl string, data any, force bool) error {
	if strings.HasSuffix(path, ".go") {
		return g.writeGoTemplate(path, tmpl, data, force)
	}
	return g.writeTextTemplate(path, tmpl, data, force)
}

func (g *Generator) writeGoTemplate(path, tmpl string, data any, force bool) error {
	rendered, err := renderTemplate(tmpl, data)
	if err != nil {
		return err
	}
	formatted, err := format.Source([]byte(rendered))
	if err != nil {
		return fmt.Errorf("format %s: %w\nsource:\n%s", path, err, rendered)
	}
	return g.writeFile(path, formatted, force)
}

func (g *Generator) writeTextTemplate(path, tmpl string, data any, force bool) error {
	rendered, err := renderTemplate(tmpl, data)
	if err != nil {
		return err
	}
	content := []byte(rendered)
	if shouldWrapGeneratedText(path) {
		content = codegenpostprocess.WrapGeneratedContent(path, content)
	}
	content = codegenpostprocess.EnsureTrailingNewline(content)
	return g.writeFile(path, content, force)
}

func (g *Generator) writeFile(path string, content []byte, force bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", filepath.Dir(path), err)
	}
	current, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat file %s: %w", path, err)
	}
	merged, err := codegenmerger.MergeContent(path, current, content, force)
	if err != nil {
		return err
	}
	if merged.Conflict && !force {
		return fmt.Errorf("merge conflict for %s: %d conflict(s)", path, merged.Diff.ConflictCount())
	}
	if len(current) > 0 && !merged.Changed && !force {
		return nil
	}
	if err := os.WriteFile(path, merged.Content, 0o644); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	return nil
}

func shouldWrapGeneratedText(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml", ".csv", ".md":
		return true
	default:
		return false
	}
}

func vueStringLiteral(value string) string {
	quoted := strconv.Quote(value)
	inner := strings.ReplaceAll(quoted[1:len(quoted)-1], "'", `\'`)
	return "'" + inner + "'"
}

func yamlStringLiteral(value string) string {
	return strconv.Quote(value)
}

func localeFieldLabelZh(field Field) string {
	label := strings.TrimSpace(field.Comment)
	if label != "" {
		label = strings.TrimSpace(strings.SplitN(label, "|", 2)[0])
		label = strings.TrimSpace(strings.SplitN(label, "（", 2)[0])
		label = strings.TrimSpace(strings.SplitN(label, "(", 2)[0])
	}
	if label == "" {
		label = strings.TrimSpace(field.DisplayLabel())
	}
	if label == "" {
		return "字段"
	}
	return label
}

func localeFieldLabelEn(field Field) string {
	label := strings.TrimSpace(field.DisplayLabel())
	if label == "" {
		return "Field"
	}
	return label
}

func fieldLocaleName(field Field) string {
	name := strings.TrimSpace(field.JSONName)
	if name == "" {
		name = strings.TrimSpace(field.GoName)
	}
	name = NormalizeName(name)
	if name == "" {
		return "field"
	}
	return name
}

func fieldLocaleKey(entityLower string, field Field, suffix string) string {
	entityLower = NormalizeName(entityLower)
	if entityLower == "" {
		entityLower = "entity"
	}
	suffix = NormalizeName(suffix)
	if suffix == "" {
		suffix = "label"
	}
	return fmt.Sprintf("%s.field.%s.%s", entityLower, fieldLocaleName(field), suffix)
}

func frontendLocaleJSON(data scaffoldData, language string) string {
	entityLower := NormalizeName(data.EntityLower)
	if entityLower == "" {
		entityLower = "entity"
	}
	language = strings.ToLower(strings.TrimSpace(language))
	isChinese := strings.HasPrefix(language, "zh")

	entries := make([]struct{ key, value string }, 0, len(data.DisplayFields())*2+10)
	if isChinese {
		entries = append(entries,
			struct{ key, value string }{key: entityLower + ".page.description", value: "由 goadmin-cli 生成的 CRUD 页面，可用于列表、编辑和删除。"},
			struct{ key, value string }{key: entityLower + ".search.placeholder", value: "搜索记录"},
			struct{ key, value string }{key: entityLower + ".create_title", value: "新建记录"},
			struct{ key, value string }{key: entityLower + ".edit_title", value: "编辑记录"},
			struct{ key, value string }{key: entityLower + ".created", value: "已创建"},
			struct{ key, value string }{key: entityLower + ".updated", value: "已更新"},
			struct{ key, value string }{key: entityLower + ".deleted", value: "已删除"},
			struct{ key, value string }{key: entityLower + ".save_failed", value: "保存失败"},
			struct{ key, value string }{key: entityLower + ".delete_confirm", value: "确认删除记录 {name} 吗？"},
			struct{ key, value string }{key: entityLower + ".delete_title", value: "删除记录"},
		)
	} else {
		entries = append(entries,
			struct{ key, value string }{key: entityLower + ".page.description", value: "Generated CRUD page for listing, editing, and deleting records."},
			struct{ key, value string }{key: entityLower + ".search.placeholder", value: "Search records"},
			struct{ key, value string }{key: entityLower + ".create_title", value: "Create record"},
			struct{ key, value string }{key: entityLower + ".edit_title", value: "Edit record"},
			struct{ key, value string }{key: entityLower + ".created", value: "Created"},
			struct{ key, value string }{key: entityLower + ".updated", value: "Updated"},
			struct{ key, value string }{key: entityLower + ".deleted", value: "Deleted"},
			struct{ key, value string }{key: entityLower + ".save_failed", value: "Save failed"},
			struct{ key, value string }{key: entityLower + ".delete_confirm", value: "Delete record {name}?"},
			struct{ key, value string }{key: entityLower + ".delete_title", value: "Delete record"},
		)
	}

	for _, field := range data.DisplayFields() {
		label := localeFieldLabelEn(field)
		placeholder := fmt.Sprintf("Please enter %s", label)
		if isChinese {
			label = localeFieldLabelZh(field)
			placeholder = fmt.Sprintf("请输入%s", label)
		}
		entries = append(entries,
			struct{ key, value string }{key: fieldLocaleKey(entityLower, field, "label"), value: label},
			struct{ key, value string }{key: fieldLocaleKey(entityLower, field, "placeholder"), value: placeholder},
		)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key < entries[j].key
	})

	var builder strings.Builder
	builder.WriteString("{\n")
	for idx, entry := range entries {
		if idx > 0 {
			builder.WriteString(",\n")
		}
		builder.WriteString("  ")
		builder.WriteString(strconv.Quote(entry.key))
		builder.WriteString(": ")
		builder.WriteString(strconv.Quote(entry.value))
	}
	builder.WriteString("\n}\n")
	return builder.String()
}

func renderTemplate(tmpl string, data any) (string, error) {
	t, err := template.New("goadmin").Funcs(template.FuncMap{
		"yamlStringLiteral":  yamlStringLiteral,
		"localeFieldLabelZh": localeFieldLabelZh,
		"localeFieldLabelEn": localeFieldLabelEn,
		"fieldLocaleKey":     fieldLocaleKey,
		"frontendLocaleJSON": frontendLocaleJSON,
		"vueStringLiteral":   vueStringLiteral,
	}).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderFieldBlock(fields []Field, includeGorm bool) string {
	if len(fields) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, field := range fields {
		builder.WriteString("\t")
		builder.WriteString(field.GoName)
		builder.WriteString(" ")
		builder.WriteString(field.GoType)
		builder.WriteString(" `json:\"")
		builder.WriteString(field.JSONName)
		builder.WriteString(",omitempty\"")
		if includeGorm {
			builder.WriteString(" gorm:\"")
			builder.WriteString(field.GormTag())
			builder.WriteString("\"")
		}
		builder.WriteString("`\n")
	}
	return builder.String()
}

func sanitizeFields(fields []Field) []Field {
	result := make([]Field, 0, len(fields))
	for _, field := range fields {
		if field.JSONName == "created_at" || field.JSONName == "updated_at" {
			continue
		}
		result = append(result, field)
	}
	return result
}

func hasTimeField(fields []Field) bool {
	for _, field := range fields {
		if field.IsTime() {
			return true
		}
	}
	return false
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

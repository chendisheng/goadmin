package generate

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

type Generator struct {
	Root       string
	PolicyPath string
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
		Name:         entityLower,
		PackageName:  entityLower,
		EntityLower:  entityLower,
		EntityPlural: Pluralize(entityLower),
		Kind:         kind,
		Force:        force,
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
	EntityLower         string
	EntityPlural        string
	Kind                string
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
	Force               bool
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

type moduleRenderData struct {
	Name         string
	PackageName  string
	EntityLower  string
	EntityPlural string
	Kind         string
	Force        bool
}

func New(root string) *Generator {
	clean := strings.TrimSpace(root)
	if clean == "" {
		clean = "."
	}
	return &Generator{
		Root:       clean,
		PolicyPath: filepath.Join(clean, "backend", "core", "auth", "casbin", "adapter", "policy.csv"),
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
	data := buildScaffoldData(opts.Name, fields, opts.GenerateFrontend, opts.GeneratePolicy, opts.Force)
	if err := g.writeScaffold(data); err != nil {
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

func buildScaffoldData(name string, fields []Field, frontend, policy, force bool) scaffoldData {
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
		EntityLower:         entityLower,
		EntityPlural:        entityPlural,
		Kind:                "business-module",
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
		Force:               force,
	}
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

func (g *Generator) writeScaffold(data scaffoldData) error {
	base := filepath.Join(g.Root, "backend", "modules", data.EntityLower)
	files := map[string]string{
		filepath.Join(base, "module.go"):                                             moduleTemplate,
		filepath.Join(base, "manifest.yaml"):                                         manifestTemplate,
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
	if data.GenerateFrontend {
		frontendFiles := map[string]string{
			filepath.Join(g.Root, "backend", "web", "src", "api", data.EntityLower+".ts"):               frontendApiTemplate,
			filepath.Join(g.Root, "backend", "web", "src", "router", "modules", data.EntityLower+".ts"): frontendRouterTemplate,
			filepath.Join(g.Root, "backend", "web", "src", "views", data.EntityLower, "index.vue"):      frontendViewTemplate,
		}
		for path, tmpl := range frontendFiles {
			if err := g.writeGoOrText(path, tmpl, data, data.Force); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Generator) writeModuleScaffold(data moduleRenderData) error {
	base := filepath.Join(g.Root, "backend", "modules", data.EntityLower)
	files := map[string]string{
		filepath.Join(base, "module.go"):     moduleTemplate,
		filepath.Join(base, "manifest.yaml"): manifestTemplate,
	}
	for path, tmpl := range files {
		if err := g.writeGoOrText(path, tmpl, data, data.Force); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) writePluginScaffold(data pluginRenderData) error {
	base := filepath.Join(g.Root, "backend", "plugin", "builtin", data.EntityLower)
	files := map[string]string{
		filepath.Join(base, data.EntityLower+".go"):                                                      pluginTemplate,
		filepath.Join(g.Root, "backend", "web", "src", "plugins", data.EntityLower+".ts"):                pluginFrontendTemplate,
		filepath.Join(g.Root, "backend", "web", "src", "views", "plugin", data.EntityLower, "index.vue"): pluginViewTemplate,
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
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
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
	for _, line := range cleaned {
		if _, ok := existing[line]; ok {
			continue
		}
		existing[line] = struct{}{}
		kept = append(kept, line)
	}
	output := strings.Join(kept, "\n") + "\n"
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
	return g.writeFile(path, []byte(rendered), force)
}

func (g *Generator) writeFile(path string, content []byte, force bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", filepath.Dir(path), err)
	}
	if _, err := os.Stat(path); err == nil && !force {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat file %s: %w", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	return nil
}

func renderTemplate(tmpl string, data any) (string, error) {
	t, err := template.New("goadmin").Parse(tmpl)
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

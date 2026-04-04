package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	modelpkg "goadmin/codegen/model"
	irmodel "goadmin/codegen/model/ir"
	"goadmin/codegen/planner"
	"goadmin/codegen/schema"
)

type DatabasePreviewSource struct {
	Driver   string   `json:"driver"`
	Database string   `json:"database"`
	Schema   string   `json:"schema,omitempty"`
	Tables   []string `json:"tables,omitempty"`
}

type DatabasePreviewPlanField struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Primary bool   `json:"primary,omitempty"`
	Index   bool   `json:"index,omitempty"`
	Unique  bool   `json:"unique,omitempty"`
}

type DatabasePreviewPlanResource struct {
	Kind             string                     `json:"kind"`
	Name             string                     `json:"name"`
	GenerateFrontend bool                       `json:"generate_frontend,omitempty"`
	GeneratePolicy   bool                       `json:"generate_policy,omitempty"`
	Force            bool                       `json:"force,omitempty"`
	Fields           []DatabasePreviewPlanField `json:"fields,omitempty"`
}

type DatabasePreviewPlan struct {
	Messages  []string                      `json:"messages,omitempty"`
	Resources []DatabasePreviewPlanResource `json:"resources,omitempty"`
}

type DatabasePreviewField struct {
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

type DatabasePreviewRelation struct {
	Type            string         `json:"type,omitempty"`
	Field           string         `json:"field,omitempty"`
	RefTable        string         `json:"ref_table,omitempty"`
	RefField        string         `json:"ref_field,omitempty"`
	UIHint          string         `json:"ui_hint,omitempty"`
	Cardinality     string         `json:"cardinality,omitempty"`
	RefDisplayField string         `json:"ref_display_field,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

type DatabasePreviewPage struct {
	Name       string         `json:"name,omitempty"`
	Type       string         `json:"type,omitempty"`
	Path       string         `json:"path,omitempty"`
	Component  string         `json:"component,omitempty"`
	Title      string         `json:"title,omitempty"`
	Permission string         `json:"permission,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type DatabasePreviewPermission struct {
	Name     string         `json:"name,omitempty"`
	Action   string         `json:"action,omitempty"`
	Resource string         `json:"resource,omitempty"`
	Policy   string         `json:"policy,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type DatabasePreviewRoute struct {
	Method   string         `json:"method,omitempty"`
	Path     string         `json:"path,omitempty"`
	Name     string         `json:"name,omitempty"`
	Policy   string         `json:"policy,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type DatabasePreviewFile struct {
	Path     string `json:"path"`
	Kind     string `json:"kind,omitempty"`
	Action   string `json:"action"`
	Resource string `json:"resource,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Exists   bool   `json:"exists,omitempty"`
	Conflict bool   `json:"conflict,omitempty"`
}

type DatabasePreviewConflict struct {
	Path     string `json:"path"`
	Resource string `json:"resource,omitempty"`
	Reason   string `json:"reason"`
}

type DatabaseAuditInput struct {
	ProjectRoot      string   `json:"project_root,omitempty"`
	Driver           string   `json:"driver"`
	Database         string   `json:"database"`
	Schema           string   `json:"schema,omitempty"`
	Tables           []string `json:"tables,omitempty"`
	Force            bool     `json:"force,omitempty"`
	GenerateFrontend bool     `json:"generate_frontend,omitempty"`
	GeneratePolicy   bool     `json:"generate_policy,omitempty"`
	DryRun           bool     `json:"dry_run"`
}

type DatabaseAuditStep struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type DatabaseAuditOutput struct {
	Files         []DatabasePreviewFile     `json:"files,omitempty"`
	Conflicts     []DatabasePreviewConflict `json:"conflicts,omitempty"`
	FileCount     int                       `json:"file_count"`
	ConflictCount int                       `json:"conflict_count"`
}

type DatabaseAuditRecord struct {
	RecordedAt string              `json:"recorded_at"`
	Input      DatabaseAuditInput  `json:"input"`
	Steps      []DatabaseAuditStep `json:"steps,omitempty"`
	Output     DatabaseAuditOutput `json:"output"`
}

type DatabasePreviewResource struct {
	TableName   string                      `json:"table_name,omitempty"`
	Kind        string                      `json:"kind,omitempty"`
	Name        string                      `json:"name,omitempty"`
	Module      string                      `json:"module,omitempty"`
	EntityName  string                      `json:"entity_name,omitempty"`
	Semantic    irmodel.Semantic            `json:"semantic,omitempty"`
	Fields      []DatabasePreviewField      `json:"fields,omitempty"`
	Relations   []DatabasePreviewRelation   `json:"relations,omitempty"`
	Pages       []DatabasePreviewPage       `json:"pages,omitempty"`
	Permissions []DatabasePreviewPermission `json:"permissions,omitempty"`
	Routes      []DatabasePreviewRoute      `json:"routes,omitempty"`
	Files       []DatabasePreviewFile       `json:"files,omitempty"`
	Conflicts   []DatabasePreviewConflict   `json:"conflicts,omitempty"`
	Actions     []string                    `json:"actions,omitempty"`
}

type DatabasePreviewReport struct {
	DryRun    bool                      `json:"dry_run"`
	Source    DatabasePreviewSource     `json:"source"`
	Messages  []string                  `json:"messages,omitempty"`
	Planner   DatabasePreviewPlan       `json:"planner"`
	Resources []DatabasePreviewResource `json:"resources,omitempty"`
	Files     []DatabasePreviewFile     `json:"files,omitempty"`
	Conflicts []DatabasePreviewConflict `json:"conflicts,omitempty"`
	Audit     DatabaseAuditRecord       `json:"audit,omitempty"`
}

func BuildDatabasePreviewReport(root string, req DatabaseExecutionRequest, irDoc irmodel.Document, schemaDoc schema.Document, plan modelpkg.Plan, dryRun bool) (DatabasePreviewReport, error) {
	resolvedResources, err := schemaDoc.ResolveResources()
	if err != nil {
		return DatabasePreviewReport{}, err
	}
	resourceCount := len(irDoc.Resources)
	if len(resolvedResources) < resourceCount {
		resourceCount = len(resolvedResources)
	}
	if len(plan.Resources) < resourceCount {
		resourceCount = len(plan.Resources)
	}
	report := DatabasePreviewReport{
		DryRun: dryRun,
		Source: DatabasePreviewSource{
			Driver:   strings.TrimSpace(req.Driver),
			Database: strings.TrimSpace(req.Database),
			Schema:   strings.TrimSpace(req.Schema),
			Tables:   append([]string(nil), req.Tables...),
		},
		Messages: append([]string(nil), plan.Messages...),
		Planner: DatabasePreviewPlan{
			Messages:  append([]string(nil), plan.Messages...),
			Resources: make([]DatabasePreviewPlanResource, 0, len(plan.Resources)),
		},
		Resources: make([]DatabasePreviewResource, 0, resourceCount),
		Files:     make([]DatabasePreviewFile, 0, resourceCount*8),
		Conflicts: make([]DatabasePreviewConflict, 0),
	}
	for _, planned := range plan.Resources {
		report.Planner.Resources = append(report.Planner.Resources, toDatabasePreviewPlanResource(planned))
	}
	for index := 0; index < resourceCount; index++ {
		irResource := irDoc.Resources[index]
		schemaResource := resolvedResources[index]
		resource := buildDatabasePreviewResource(root, irResource, schemaResource, req.Force)
		report.Resources = append(report.Resources, resource)
		report.Files = append(report.Files, resource.Files...)
		report.Conflicts = append(report.Conflicts, resource.Conflicts...)
	}
	report.Files = sortDatabasePreviewFiles(report.Files)
	report.Conflicts = sortDatabasePreviewConflicts(report.Conflicts)
	report.Audit = buildDatabaseAuditRecord(root, req, dryRun, report.Files, report.Conflicts, len(irDoc.Resources), len(plan.Resources), len(report.Resources))
	return report, nil
}

func buildDatabaseAuditRecord(root string, req DatabaseExecutionRequest, dryRun bool, files []DatabasePreviewFile, conflicts []DatabasePreviewConflict, irCount int, planCount int, resourceCount int) DatabaseAuditRecord {
	steps := []DatabaseAuditStep{
		{
			Name:   "inspect-database",
			Status: "ok",
			Detail: fmt.Sprintf("driver=%s database=%s schema=%s tables=%d", strings.TrimSpace(req.Driver), strings.TrimSpace(req.Database), strings.TrimSpace(req.Schema), len(req.Tables)),
		},
		{
			Name:   "build-ir",
			Status: "ok",
			Detail: fmt.Sprintf("resources=%d", irCount),
		},
		{
			Name:   "plan-schema",
			Status: "ok",
			Detail: fmt.Sprintf("planned=%d", planCount),
		},
	}
	if dryRun {
		steps = append(steps, DatabaseAuditStep{Name: "preview-output", Status: "ok", Detail: fmt.Sprintf("dry-run resources=%d", resourceCount)})
	} else {
		steps = append(steps, DatabaseAuditStep{Name: "write-output", Status: "ok", Detail: fmt.Sprintf("generated resources=%d", resourceCount)})
	}
	return DatabaseAuditRecord{
		RecordedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Input: DatabaseAuditInput{
			ProjectRoot:      strings.TrimSpace(root),
			Driver:           strings.TrimSpace(req.Driver),
			Database:         strings.TrimSpace(req.Database),
			Schema:           strings.TrimSpace(req.Schema),
			Tables:           append([]string(nil), req.Tables...),
			Force:            req.Force,
			GenerateFrontend: req.GenerateFrontend != nil && *req.GenerateFrontend,
			GeneratePolicy:   req.GeneratePolicy != nil && *req.GeneratePolicy,
			DryRun:           dryRun,
		},
		Steps: steps,
		Output: DatabaseAuditOutput{
			Files:         append([]DatabasePreviewFile(nil), files...),
			Conflicts:     append([]DatabasePreviewConflict(nil), conflicts...),
			FileCount:     len(files),
			ConflictCount: len(conflicts),
		},
	}
}

func toDatabasePreviewPlanResource(resource modelpkg.Resource) DatabasePreviewPlanResource {
	planned := DatabasePreviewPlanResource{
		Kind:             strings.TrimSpace(resource.Kind),
		Name:             strings.TrimSpace(resource.Name),
		GenerateFrontend: resource.GenerateFrontend,
		GeneratePolicy:   resource.GeneratePolicy,
		Force:            resource.Force,
	}
	if len(resource.Fields) > 0 {
		planned.Fields = make([]DatabasePreviewPlanField, 0, len(resource.Fields))
		for _, field := range resource.Fields {
			planned.Fields = append(planned.Fields, DatabasePreviewPlanField{
				Name:    strings.TrimSpace(field.Name),
				Type:    strings.TrimSpace(field.Type),
				Primary: field.Primary,
				Index:   field.Index,
				Unique:  field.Unique,
			})
		}
	}
	return planned
}

func buildDatabasePreviewResource(root string, irResource irmodel.Resource, schemaResource schema.Resource, force bool) DatabasePreviewResource {
	scope := choosePageScope(schemaResource)
	resource := DatabasePreviewResource{
		TableName:   strings.TrimSpace(irResource.TableName),
		Kind:        string(schemaResource.Kind),
		Name:        strings.TrimSpace(schemaResource.Name),
		Module:      strings.TrimSpace(schemaResource.Module),
		EntityName:  strings.TrimSpace(schemaResource.Entity.Name),
		Semantic:    irResource.Semantic,
		Fields:      make([]DatabasePreviewField, 0, len(irResource.Fields)),
		Relations:   make([]DatabasePreviewRelation, 0, len(irResource.Relations)),
		Pages:       make([]DatabasePreviewPage, 0, len(irResource.Pages)),
		Permissions: make([]DatabasePreviewPermission, 0, len(irResource.Permissions)),
		Routes:      make([]DatabasePreviewRoute, 0, len(irResource.Routes)),
		Files:       previewSchemaResourceFiles(root, schemaResource, force),
		Actions:     describeSchemaResourceActions(schemaResource),
	}
	for _, field := range irResource.Fields {
		resource.Fields = append(resource.Fields, DatabasePreviewField{
			Name:         field.Name,
			ColumnName:   field.ColumnName,
			GoType:       field.GoType,
			DBType:       field.DBType,
			Nullable:     field.Nullable,
			Primary:      field.Primary,
			Unique:       field.Unique,
			Index:        field.Index,
			Required:     field.Required,
			UIType:       field.UIType,
			Label:        field.Label,
			Searchable:   field.Searchable,
			Editable:     field.Editable,
			Sortable:     field.Sortable,
			SemanticType: field.SemanticType,
			DefaultValue: field.DefaultValue,
			EnumValues:   append([]string(nil), field.EnumValues...),
			Metadata:     cloneAnyMap(field.Metadata),
		})
	}
	for _, relation := range irResource.Relations {
		resource.Relations = append(resource.Relations, DatabasePreviewRelation{
			Type:            relation.Type,
			Field:           relation.Field,
			RefTable:        relation.RefTable,
			RefField:        relation.RefField,
			UIHint:          relation.UIHint,
			Cardinality:     relation.Cardinality,
			RefDisplayField: relation.RefDisplayField,
			Metadata:        cloneAnyMap(relation.Metadata),
		})
	}
	manifestPermissions := buildManifestPermissions(schemaResource)
	resource.Permissions = make([]DatabasePreviewPermission, 0, len(manifestPermissions))
	for _, permission := range manifestPermissions {
		policy := ""
		if resourceName := strings.TrimSpace(permission.Object); resourceName != "" {
			policy = resourceName + ":" + strings.TrimSpace(permission.Action)
		}
		resource.Permissions = append(resource.Permissions, DatabasePreviewPermission{
			Name:     strings.TrimSpace(permission.Description),
			Action:   strings.TrimSpace(permission.Action),
			Resource: strings.TrimSpace(permission.Object),
			Policy:   policy,
			Metadata: map[string]any{
				"description": strings.TrimSpace(permission.Description),
			},
		})
	}
	pages := append([]schema.Page(nil), schemaResource.Pages...)
	if len(pages) == 0 && (schemaResource.GenerateFrontend || hasCRUDContent(schemaResource)) {
		pages = append(pages, schema.Page{
			Name:      "List",
			Type:      "list",
			Path:      "/" + scope + "/list",
			Component: "view/" + scope + "/index",
		})
	}
	for _, page := range pages {
		resource.Pages = append(resource.Pages, DatabasePreviewPage{
			Name:       page.Name,
			Type:       page.Type,
			Path:       page.RoutePath(scope),
			Component:  page.ComponentName(scope),
			Title:      page.Title(),
			Permission: choosePagePermission(schemaResource, page, scope),
		})
	}
	for _, permission := range schemaResource.Permissions {
		policy := ""
		if scopeName, action, ok := permission.PolicyParts(); ok {
			policy = scopeName + ":" + action
		}
		resource.Permissions = append(resource.Permissions, DatabasePreviewPermission{
			Name:     permission.Name,
			Action:   permission.Action,
			Resource: permission.Resource,
			Policy:   policy,
		})
	}
	for _, route := range schemaResource.Routes {
		policy := ""
		if path, method, ok := route.PolicyParts(); ok {
			policy = method + " " + path
		}
		resource.Routes = append(resource.Routes, DatabasePreviewRoute{
			Method: route.Method,
			Path:   route.Path,
			Name:   route.Name,
			Policy: policy,
		})
	}
	resource.Conflicts = collectPreviewConflicts(resource.Files, resource.Name)
	return resource
}

func previewSchemaResourceFiles(root string, resource schema.Resource, force bool) []DatabasePreviewFile {
	files := make([]DatabasePreviewFile, 0, 16)
	add := func(path, kind, action, reason, resourceName string) {
		path = filepath.ToSlash(strings.TrimSpace(path))
		if path == "" {
			return
		}
		exists := fileExists(path)
		fileAction := action
		conflict := false
		conflictReason := ""
		switch {
		case action == "append":
			// append actions are additive and do not conflict with existing files.
		case exists && force:
			fileAction = "overwrite"
		case exists && !force:
			fileAction = "skip"
			conflict = true
			if reason != "" {
				conflictReason = reason + "; exists and force=false"
			} else {
				conflictReason = "file exists and force=false"
			}
		}
		files = append(files, DatabasePreviewFile{
			Path:     path,
			Kind:     kind,
			Action:   fileAction,
			Resource: resourceName,
			Reason:   reason,
			Exists:   exists,
			Conflict: conflict,
		})
		if conflict {
			files[len(files)-1].Reason = conflictReason
		}
	}
	resourceName := chooseCRUDName(resource)
	if resourceName == "" {
		resourceName = chooseModuleName(resource)
	}
	if resourceName == "" {
		resourceName = choosePluginName(resource)
	}
	switch resource.Kind {
	case schema.KindModule, schema.KindBusinessModule, schema.KindBackendModule:
		if moduleName := chooseModuleName(resource); moduleName != "" {
			add(filepath.Join(root, "backend", "modules", moduleName, "module.go"), "module", "create", "module scaffold", moduleName)
			add(filepath.Join(root, "backend", "modules", moduleName, "manifest.yaml"), "manifest", "create", "module manifest", moduleName)
		}
		if hasCRUDContent(resource) {
			files = append(files, previewCRUDFiles(root, resource, force)...)
		}
	case schema.KindCRUD, schema.KindBackendCRUD:
		files = append(files, previewCRUDFiles(root, resource, force)...)
	case schema.KindPlugin, schema.KindBackendPlugin:
		if pluginName := choosePluginName(resource); pluginName != "" {
			add(filepath.Join(root, "backend", "plugin", "builtin", pluginName, pluginName+".go"), "plugin", "create", "plugin scaffold", pluginName)
			add(filepath.Join(root, "backend", "web", "src", "plugins", pluginName+".ts"), "plugin-frontend", "create", "plugin frontend entry", pluginName)
			add(filepath.Join(root, "backend", "web", "src", "views", "plugin", pluginName, "index.vue"), "plugin-view", "create", "plugin view", pluginName)
			add(filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv"), "policy", "append", "plugin ping policy", pluginName)
		}
	case schema.KindManifest:
		if moduleName := chooseModuleName(resource); moduleName != "" {
			add(filepath.Join(root, "backend", "modules", moduleName, "manifest.yaml"), "manifest", "create", "module manifest", moduleName)
		}
	case schema.KindConfig:
		name := strings.TrimSpace(resource.Name)
		if name == "" {
			name = chooseModuleName(resource)
		}
		if name == "" {
			name = "generated"
		}
		add(filepath.Join(root, "backend", "config", "config."+schema.NormalizeName(name)+".yaml"), "config", "create", "config profile", name)
	case schema.KindFrontendPage, schema.KindFrontendModuleRoute:
		if hasCRUDContent(resource) {
			files = append(files, previewCRUDFiles(root, resource, force)...)
		}
	default:
		if hasCRUDContent(resource) {
			files = append(files, previewCRUDFiles(root, resource, force)...)
		}
		if moduleName := chooseModuleName(resource); moduleName != "" {
			add(filepath.Join(root, "backend", "modules", moduleName, "module.go"), "module", "create", "module scaffold", moduleName)
			add(filepath.Join(root, "backend", "modules", moduleName, "manifest.yaml"), "manifest", "create", "module manifest", moduleName)
		}
		if pluginName := choosePluginName(resource); pluginName != "" {
			add(filepath.Join(root, "backend", "plugin", "builtin", pluginName, pluginName+".go"), "plugin", "create", "plugin scaffold", pluginName)
			add(filepath.Join(root, "backend", "web", "src", "plugins", pluginName+".ts"), "plugin-frontend", "create", "plugin frontend entry", pluginName)
			add(filepath.Join(root, "backend", "web", "src", "views", "plugin", pluginName, "index.vue"), "plugin-view", "create", "plugin view", pluginName)
		}
	}
	pages := resource.Pages
	if len(pages) == 0 && (resource.Kind == schema.KindFrontendPage || resource.Kind == schema.KindFrontendModuleRoute) {
		pages = []schema.Page{{Name: resource.Name, Type: string(resource.Kind)}}
	}
	if len(pages) > 0 {
		scope := choosePageScope(resource)
		for _, page := range pages {
			pageSlug := page.NormalizedName()
			if pageSlug == "" {
				pageSlug = "index"
			}
			add(filepath.Join(root, "backend", "web", "src", "views", scope, pageSlug+".vue"), "page-view", "create", page.Title(), scope)
			add(filepath.Join(root, "backend", "web", "src", "router", "modules", scope+"-"+pageSlug+".ts"), "page-router", "create", page.Title(), scope)
		}
		add(filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv"), "policy", "append", "page policies", resourceName)
	}
	if len(resource.Routes) > 0 {
		add(filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv"), "policy", "append", "route policies", resourceName)
	}
	if len(resource.Permissions) > 0 {
		add(filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv"), "policy", "append", "permission policies", resourceName)
	}
	return sortDatabasePreviewFiles(files)
}

func previewCRUDFiles(root string, resource schema.Resource, force bool) []DatabasePreviewFile {
	name := chooseCRUDName(resource)
	if name == "" {
		name = chooseModuleName(resource)
	}
	if name == "" {
		name = choosePluginName(resource)
	}
	entityLower := schema.NormalizeName(name)
	if entityLower == "" {
		entityLower = "entity"
	}
	base := filepath.Join(root, "backend", "modules", entityLower)
	files := []DatabasePreviewFile{
		previewFile(base, "module", "create", "module scaffold", entityLower, force, false),
		previewFile(filepath.Join(base, "manifest.yaml"), "manifest", "create", "module manifest", entityLower, force, false),
		previewFile(filepath.Join(base, "domain", "model", entityLower+".go"), "crud", "create", "model scaffold", entityLower, force, false),
		previewFile(filepath.Join(base, "domain", "repository", "repository.go"), "crud", "create", "repository scaffold", entityLower, force, false),
		previewFile(filepath.Join(base, "application", "command", entityLower+".go"), "crud", "create", "command scaffold", entityLower, force, false),
		previewFile(filepath.Join(base, "application", "query", entityLower+".go"), "crud", "create", "query scaffold", entityLower, force, false),
		previewFile(filepath.Join(base, "application", "service", "service.go"), "crud", "create", "service scaffold", entityLower, force, false),
		previewFile(filepath.Join(base, "infrastructure", "repo", "gorm.go"), "crud", "create", "gorm repository", entityLower, force, false),
		previewFile(filepath.Join(base, "transport", "http", "request", entityLower+".go"), "crud", "create", "http request", entityLower, force, false),
		previewFile(filepath.Join(base, "transport", "http", "response", entityLower+".go"), "crud", "create", "http response", entityLower, force, false),
		previewFile(filepath.Join(base, "transport", "http", "handler", "handler.go"), "crud", "create", "http handler", entityLower, force, false),
		previewFile(filepath.Join(base, "transport", "http", "router.go"), "crud", "create", "http router", entityLower, force, false),
	}
	if resource.GenerateFrontend || len(resource.Pages) > 0 || strings.TrimSpace(resource.Framework.Frontend) != "" {
		files = append(files,
			previewFile(filepath.Join(root, "backend", "web", "src", "api", entityLower+".ts"), "frontend-api", "create", "frontend api", entityLower, force, false),
			previewFile(filepath.Join(root, "backend", "web", "src", "router", "modules", entityLower+".ts"), "frontend-router", "create", "frontend router", entityLower, force, false),
			previewFile(filepath.Join(root, "backend", "web", "src", "views", entityLower, "index.vue"), "frontend-view", "create", "frontend view", entityLower, force, false),
		)
	}
	if resource.GeneratePolicy || len(resource.Routes) > 0 || len(resource.Pages) > 0 || len(resource.Permissions) > 0 {
		files = append(files, previewFile(filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv"), "policy", "append", "CRUD policy lines", entityLower, force, true))
	}
	return sortDatabasePreviewFiles(files)
}

func previewFile(path, kind, action, reason, resource string, force bool, appendOnly bool) DatabasePreviewFile {
	path = filepath.ToSlash(strings.TrimSpace(path))
	exists := fileExists(path)
	fileAction := action
	conflict := false
	if appendOnly {
		fileAction = "append"
	} else if exists && force {
		fileAction = "overwrite"
	} else if exists && !force {
		fileAction = "skip"
		conflict = true
	}
	return DatabasePreviewFile{
		Path:     path,
		Kind:     kind,
		Action:   fileAction,
		Resource: resource,
		Reason:   reason,
		Exists:   exists,
		Conflict: conflict,
	}
}

func collectPreviewConflicts(files []DatabasePreviewFile, resource string) []DatabasePreviewConflict {
	conflicts := make([]DatabasePreviewConflict, 0)
	for _, file := range files {
		if !file.Conflict {
			continue
		}
		reason := strings.TrimSpace(file.Reason)
		if reason == "" {
			reason = "file exists and generation would skip or overwrite"
		}
		conflicts = append(conflicts, DatabasePreviewConflict{
			Path:     file.Path,
			Resource: resource,
			Reason:   reason,
		})
	}
	return conflicts
}

func sortDatabasePreviewFiles(files []DatabasePreviewFile) []DatabasePreviewFile {
	if len(files) == 0 {
		return nil
	}
	result := append([]DatabasePreviewFile(nil), files...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Path == result[j].Path {
			if result[i].Resource == result[j].Resource {
				return result[i].Kind < result[j].Kind
			}
			return result[i].Resource < result[j].Resource
		}
		return result[i].Path < result[j].Path
	})
	return result
}

func sortDatabasePreviewConflicts(conflicts []DatabasePreviewConflict) []DatabasePreviewConflict {
	if len(conflicts) == 0 {
		return nil
	}
	result := append([]DatabasePreviewConflict(nil), conflicts...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Path == result[j].Path {
			if result[i].Resource == result[j].Resource {
				return result[i].Reason < result[j].Reason
			}
			return result[i].Resource < result[j].Resource
		}
		return result[i].Path < result[j].Path
	})
	return result
}

func cloneAnyMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]any, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func previewDatabaseReportText(report DatabasePreviewReport) string {
	var builder strings.Builder
	builder.WriteString("database preview report\n")
	builder.WriteString(fmt.Sprintf("source: %s/%s", report.Source.Driver, report.Source.Database))
	if report.Source.Schema != "" {
		builder.WriteString(" schema=")
		builder.WriteString(report.Source.Schema)
	}
	if report.DryRun {
		builder.WriteString(" [dry-run]")
	}
	builder.WriteString("\n")
	if report.Audit.RecordedAt != "" {
		builder.WriteString("audit:\n")
		builder.WriteString("  recorded_at: ")
		builder.WriteString(report.Audit.RecordedAt)
		builder.WriteString("\n")
		if report.Audit.Input.ProjectRoot != "" {
			builder.WriteString("  project_root: ")
			builder.WriteString(report.Audit.Input.ProjectRoot)
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("  input: driver=%s database=%s schema=%s force=%t frontend=%t policy=%t dry_run=%t\n", report.Audit.Input.Driver, report.Audit.Input.Database, report.Audit.Input.Schema, report.Audit.Input.Force, report.Audit.Input.GenerateFrontend, report.Audit.Input.GeneratePolicy, report.Audit.Input.DryRun))
		if len(report.Audit.Input.Tables) > 0 {
			builder.WriteString("  tables: ")
			builder.WriteString(strings.Join(report.Audit.Input.Tables, ", "))
			builder.WriteString("\n")
		}
		if len(report.Audit.Steps) > 0 {
			builder.WriteString("  steps:\n")
			for _, step := range report.Audit.Steps {
				builder.WriteString(fmt.Sprintf("    - %s [%s] %s\n", step.Name, step.Status, step.Detail))
			}
		}
		builder.WriteString(fmt.Sprintf("  outputs: files=%d conflicts=%d\n", report.Audit.Output.FileCount, report.Audit.Output.ConflictCount))
	}
	if len(report.Messages) > 0 {
		builder.WriteString("messages:\n")
		for _, msg := range report.Messages {
			builder.WriteString("- ")
			builder.WriteString(msg)
			builder.WriteString("\n")
		}
	}
	if len(report.Planner.Messages) > 0 || len(report.Planner.Resources) > 0 {
		builder.WriteString("planner:\n")
		for _, msg := range report.Planner.Messages {
			builder.WriteString("  - ")
			builder.WriteString(msg)
			builder.WriteString("\n")
		}
		for _, planned := range report.Planner.Resources {
			builder.WriteString(fmt.Sprintf("  resource %s [%s]\n", planned.Name, planned.Kind))
			for _, field := range planned.Fields {
				builder.WriteString(fmt.Sprintf("    field %s (%s) primary=%t index=%t unique=%t\n", field.Name, field.Type, field.Primary, field.Index, field.Unique))
			}
		}
	}
	for _, resource := range report.Resources {
		builder.WriteString(fmt.Sprintf("resource %s [%s] (%s) -> %s\n", resource.Name, resource.Kind, resource.TableName, strings.Join(resource.Actions, "; ")))
		for _, field := range resource.Fields {
			builder.WriteString(fmt.Sprintf("  field mapping %s <- %s [%s/%s] required=%t editable=%t sortable=%t\n", field.Name, field.ColumnName, field.SemanticType, field.UIType, field.Required, field.Editable, field.Sortable))
		}
		for _, relation := range resource.Relations {
			builder.WriteString(fmt.Sprintf("  relation %s -> %s.%s (%s/%s)\n", relation.Field, relation.RefTable, relation.RefField, relation.Type, relation.Cardinality))
		}
		for _, page := range resource.Pages {
			builder.WriteString(fmt.Sprintf("  page item %s -> %s (%s, permission=%s)\n", page.Title, page.Path, page.Component, page.Permission))
		}
		for _, permission := range resource.Permissions {
			builder.WriteString(fmt.Sprintf("  permission item %s %s (%s)\n", permission.Resource, permission.Action, permission.Name))
			if desc, ok := permission.Metadata["description"].(string); ok && strings.TrimSpace(desc) != "" {
				builder.WriteString("    description: ")
				builder.WriteString(strings.TrimSpace(desc))
				builder.WriteString("\n")
			}
		}
		for _, route := range resource.Routes {
			builder.WriteString(fmt.Sprintf("  route item %s %s (%s)\n", route.Method, route.Path, route.Policy))
		}
		for _, file := range resource.Files {
			builder.WriteString(fmt.Sprintf("  file plan %s [%s] (%s)\n", file.Path, file.Action, file.Kind))
		}
		for _, conflict := range resource.Conflicts {
			builder.WriteString(fmt.Sprintf("  conflict %s: %s\n", conflict.Path, conflict.Reason))
		}
	}
	if len(report.Files) > 0 {
		builder.WriteString("file plan:\n")
		for _, file := range report.Files {
			builder.WriteString(fmt.Sprintf("- %s [%s] (%s)\n", file.Path, file.Action, file.Kind))
		}
	}
	if len(report.Conflicts) > 0 {
		builder.WriteString("conflicts:\n")
		for _, conflict := range report.Conflicts {
			builder.WriteString(fmt.Sprintf("- %s %s\n", conflict.Path, conflict.Reason))
		}
	}
	return builder.String()
}

func previewPlanResourcePaths(plan modelpkg.Plan) []string {
	paths := make([]string, 0, len(plan.Resources))
	for _, resource := range plan.Resources {
		paths = append(paths, resource.Name)
	}
	return paths
}

func buildPlannerFromSchema(document schema.Document) (modelpkg.Plan, error) {
	return planner.New().Plan(document)
}

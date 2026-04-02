package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	legacygenerate "goadmin/cli/generate"
	"goadmin/codegen/planner"
	"goadmin/codegen/schema"
)

func Run(root string, args []string) error {
	if len(args) == 0 {
		return errors.New("generate requires a subcommand: module, crud, plugin, dsl")
	}
	gen := legacygenerate.New(root)
	plan := planner.New()

	switch args[0] {
	case "generate":
		return runGenerate(gen, plan, args[1:])
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func generateSchemaResourceManifest(gen *legacygenerate.Generator, resource schema.Resource, force bool) error {
	if shouldSkipManifest(resource) {
		return nil
	}
	return gen.GenerateManifest(buildManifestOptions(resource, force))
}

func generateSchemaResourceConfig(gen *legacygenerate.Generator, resource schema.Resource, force bool) error {
	if resource.Kind != schema.KindConfig {
		return nil
	}
	return gen.GenerateConfig(buildConfigOptions(resource, force))
}

func runGenerate(gen *legacygenerate.Generator, plan planner.Default, args []string) error {
	if len(args) == 0 {
		return errors.New("generate requires a subcommand: module, crud, plugin, dsl")
	}
	switch args[0] {
	case "module":
		return runGenerateModule(gen, plan, args[1:])
	case "crud":
		return runGenerateCRUD(gen, plan, args[1:])
	case "plugin":
		return runGeneratePlugin(gen, plan, args[1:])
	case "dsl":
		return runGenerateDSL(gen, plan, args[1:])
	default:
		return fmt.Errorf("unknown generate subcommand %q", args[0])
	}
}

func runGenerateModule(gen *legacygenerate.Generator, plan planner.Default, args []string) error {
	fs := flag.NewFlagSet("generate module", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite existing files")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("generate module requires a module name")
	}
	name := fs.Arg(0)
	if _, err := plan.Plan(schema.NewDocument(schema.Resource{Kind: schema.KindModule, Name: name, Force: *force})); err != nil {
		return err
	}
	return gen.GenerateModule(legacygenerate.ModuleOptions{Name: name, Force: *force})
}

func runGenerateCRUD(gen *legacygenerate.Generator, plan planner.Default, args []string) error {
	fs := flag.NewFlagSet("generate crud", flag.ContinueOnError)
	fields := fs.String("fields", "", "comma separated field definitions like name:string,status:string")
	primary := fs.String("primary", "", "comma separated primary key fields")
	indexes := fs.String("index", "", "comma separated indexed fields")
	uniques := fs.String("unique", "", "comma separated unique fields")
	frontend := fs.Bool("frontend", true, "generate frontend scaffolding")
	policy := fs.Bool("policy", true, "append Casbin policy lines")
	force := fs.Bool("force", false, "overwrite existing files")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("generate crud requires an entity name")
	}
	parsedFields, err := legacygenerate.ParseFields(*fields, *primary, *indexes, *uniques)
	if err != nil {
		return err
	}
	if _, err := plan.Plan(schema.NewDocument(schema.Resource{
		Kind:             schema.KindCRUD,
		Name:             fs.Arg(0),
		Fields:           toSchemaFields(parsedFields),
		GenerateFrontend: *frontend,
		GeneratePolicy:   *policy,
		Force:            *force,
	})); err != nil {
		return err
	}
	return gen.GenerateCRUD(legacygenerate.CRUDOptions{
		Name:             fs.Arg(0),
		Fields:           parsedFields,
		GenerateFrontend: *frontend,
		GeneratePolicy:   *policy,
		Force:            *force,
	})
}

func runGeneratePlugin(gen *legacygenerate.Generator, plan planner.Default, args []string) error {
	fs := flag.NewFlagSet("generate plugin", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite existing files")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("generate plugin requires a plugin name")
	}
	name := fs.Arg(0)
	if _, err := plan.Plan(schema.NewDocument(schema.Resource{Kind: schema.KindPlugin, Name: name, Force: *force})); err != nil {
		return err
	}
	return gen.GeneratePlugin(legacygenerate.PluginOptions{Name: name, Force: *force})
}

func runGenerateDSL(gen *legacygenerate.Generator, plan planner.Default, args []string) error {
	fs := flag.NewFlagSet("generate dsl", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite existing files")
	dryRun := fs.Bool("dry-run", false, "preview generation actions without writing files")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("generate dsl requires a YAML file path")
	}
	doc, resources, err := ParseDSLResourcesFromFile(fs.Arg(0))
	if err != nil {
		return err
	}
	if _, err := plan.Plan(doc); err != nil {
		return err
	}
	if *dryRun {
		return previewDSLResources(resources, *force)
	}
	if err := ExecuteDSLResources(gen.Root, resources, *force); err != nil {
		return err
	}
	return nil
}

func generateFromSchemaResource(gen *legacygenerate.Generator, resource schema.Resource, cliForce bool) error {
	force := cliForce || resource.Force
	if err := generateSchemaResourceScaffold(gen, resource, force); err != nil {
		return err
	}
	if err := generateSchemaResourcePages(gen, resource, force); err != nil {
		return err
	}
	if err := generateSchemaResourcePolicies(gen, resource); err != nil {
		return err
	}
	if err := generateSchemaResourceManifest(gen, resource, force); err != nil {
		return err
	}
	if err := generateSchemaResourceConfig(gen, resource, force); err != nil {
		return err
	}
	return nil
}

func generateCRUDFromSchemaResource(gen *legacygenerate.Generator, resource schema.Resource, force bool) error {
	name := chooseCRUDName(resource)
	if name == "" {
		return errors.New("dsl crud requires an entity or resource name")
	}
	fields, err := toLegacyFields(resource)
	if err != nil {
		return err
	}
	generateFrontend := resource.GenerateFrontend || len(resource.Pages) > 0 || strings.TrimSpace(resource.Framework.Frontend) != ""
	generatePolicy := resource.GeneratePolicy || len(resource.Permissions) > 0
	return gen.GenerateCRUD(legacygenerate.CRUDOptions{
		Name:                name,
		Fields:              fields,
		GenerateFrontend:    generateFrontend,
		GeneratePolicy:      generatePolicy,
		ManifestRoutes:      buildManifestRoutes(resource),
		ManifestMenus:       buildManifestMenus(resource),
		ManifestPermissions: buildManifestPermissions(resource),
		Force:               force,
	})
}

func generateSchemaResourceScaffold(gen *legacygenerate.Generator, resource schema.Resource, force bool) error {
	switch resource.Kind {
	case schema.KindModule, schema.KindBusinessModule, schema.KindBackendModule:
		moduleName := chooseModuleName(resource)
		if moduleName != "" {
			if err := gen.GenerateModule(legacygenerate.ModuleOptions{Name: moduleName, Force: force}); err != nil {
				return err
			}
		}
		if hasCRUDContent(resource) {
			return generateCRUDFromSchemaResource(gen, resource, force)
		}
		return nil
	case schema.KindCRUD, schema.KindBackendCRUD:
		return generateCRUDFromSchemaResource(gen, resource, force)
	case schema.KindPlugin, schema.KindBackendPlugin:
		pluginName := choosePluginName(resource)
		if pluginName == "" {
			return errors.New("dsl plugin requires a plugin name")
		}
		return gen.GeneratePlugin(legacygenerate.PluginOptions{Name: pluginName, Force: force})
	case schema.KindManifest:
		return nil
	case schema.KindConfig:
		return nil
	case schema.KindFrontendPage, schema.KindFrontendModuleRoute, schema.KindPolicy:
		if hasCRUDContent(resource) {
			return generateCRUDFromSchemaResource(gen, resource, force)
		}
		return nil
	default:
		if hasCRUDContent(resource) {
			return generateCRUDFromSchemaResource(gen, resource, force)
		}
		if chooseModuleName(resource) != "" {
			return gen.GenerateModule(legacygenerate.ModuleOptions{Name: chooseModuleName(resource), Force: force})
		}
		if choosePluginName(resource) != "" {
			return gen.GeneratePlugin(legacygenerate.PluginOptions{Name: choosePluginName(resource), Force: force})
		}
		return nil
	}
}

func generateSchemaResourcePages(gen *legacygenerate.Generator, resource schema.Resource, force bool) error {
	pages := resource.Pages
	if len(pages) == 0 && (resource.Kind == schema.KindFrontendPage || resource.Kind == schema.KindFrontendModuleRoute) {
		pages = []schema.Page{{Name: resource.Name, Type: string(resource.Kind)}}
	}
	if len(pages) == 0 {
		return nil
	}
	scope := choosePageScope(resource)
	for _, page := range pages {
		if err := gen.GeneratePage(legacygenerate.PageOptions{
			ViewScope:  scope,
			RouteScope: scope,
			PageName:   page.Title(),
			PageSlug:   page.NormalizedName(),
			Title:      buildPageTitle(resource, page, scope),
			RoutePath:  page.RoutePath(scope),
			Component:  page.ComponentName(scope),
			Permission: choosePagePermission(resource, page, scope),
			Force:      force,
		}); err != nil {
			return err
		}
	}
	return nil
}

func generateSchemaResourcePolicies(gen *legacygenerate.Generator, resource schema.Resource) error {
	lines := make([]string, 0)
	for _, route := range resource.Routes {
		path, method, ok := route.PolicyParts()
		if !ok {
			continue
		}
		lines = append(lines, fmt.Sprintf("p, admin, %s, %s", path, method))
	}
	scope := choosePageScope(resource)
	for _, page := range resource.Pages {
		path := page.RoutePath(scope)
		if path == "" {
			continue
		}
		method := pagePolicyMethod(page.PermissionAction())
		if method == "" {
			method = "GET"
		}
		lines = append(lines, fmt.Sprintf("p, admin, %s, %s", path, method))
	}
	if len(lines) == 0 {
		return nil
	}
	return gen.AppendPolicyLines(lines)
}

func toLegacyFields(resource schema.Resource) ([]legacygenerate.Field, error) {
	fields := resource.Fields
	if len(fields) == 0 {
		fields = resource.Entity.Fields
	}
	if len(fields) == 0 {
		return nil, nil
	}
	parts := make([]string, 0, len(fields))
	primary := make([]string, 0)
	indexes := make([]string, 0)
	uniques := make([]string, 0)
	for _, field := range fields {
		name := strings.TrimSpace(field.Name)
		if name == "" {
			return nil, errors.New("dsl field name is required")
		}
		typeName := strings.TrimSpace(field.Type)
		if typeName == "" {
			typeName = "string"
		}
		parts = append(parts, fmt.Sprintf("%s:%s", name, typeName))
		if field.Primary {
			primary = append(primary, name)
		}
		if field.Index {
			indexes = append(indexes, name)
		}
		if field.Unique {
			uniques = append(uniques, name)
		}
	}
	return legacygenerate.ParseFields(
		strings.Join(parts, ","),
		strings.Join(primary, ","),
		strings.Join(indexes, ","),
		strings.Join(uniques, ","),
	)
}

func chooseModuleName(resource schema.Resource) string {
	if name := strings.TrimSpace(resource.Module); name != "" {
		return name
	}
	if name := strings.TrimSpace(resource.Name); name != "" {
		return name
	}
	if name := strings.TrimSpace(resource.Entity.Name); name != "" {
		return name
	}
	if resource.Plugin != nil {
		return strings.TrimSpace(resource.Plugin.Name)
	}
	return ""
}

func chooseCRUDName(resource schema.Resource) string {
	if name := strings.TrimSpace(resource.Entity.Name); name != "" {
		return name
	}
	if name := strings.TrimSpace(resource.Name); name != "" {
		return name
	}
	return strings.TrimSpace(resource.Module)
}

func choosePluginName(resource schema.Resource) string {
	if resource.Plugin != nil {
		if name := strings.TrimSpace(resource.Plugin.Name); name != "" {
			return name
		}
	}
	return strings.TrimSpace(resource.Name)
}

func hasCRUDContent(resource schema.Resource) bool {
	return len(resource.Fields) > 0 || len(resource.Entity.Fields) > 0 || strings.TrimSpace(resource.Entity.Name) != "" || resource.GenerateFrontend || resource.GeneratePolicy || len(resource.Pages) > 0 || len(resource.Permissions) > 0
}

func choosePageScope(resource schema.Resource) string {
	if name := strings.TrimSpace(resource.Module); name != "" {
		return schema.NormalizeName(name)
	}
	if name := strings.TrimSpace(resource.Entity.Name); name != "" {
		return schema.NormalizeName(name)
	}
	if name := strings.TrimSpace(resource.Name); name != "" {
		return schema.NormalizeName(name)
	}
	if resource.Plugin != nil {
		return schema.NormalizeName(resource.Plugin.Name)
	}
	return "page"
}

func buildPageTitle(resource schema.Resource, page schema.Page, scope string) string {
	title := strings.TrimSpace(page.Title())
	if title == "" {
		title = strings.TrimSpace(page.NormalizedName())
	}
	if title == "" {
		title = "Page"
	}
	scopeTitle := strings.TrimSpace(scope)
	if scopeTitle != "" {
		scopeTitle = legacygenerate.ToCamel(scopeTitle)
	}
	if scopeTitle != "" && !strings.Contains(strings.ToLower(title), strings.ToLower(scopeTitle)) {
		return scopeTitle + " " + title
	}
	if name := strings.TrimSpace(resource.Name); name != "" && !strings.Contains(strings.ToLower(title), strings.ToLower(name)) {
		if scopeTitle != "" {
			return scopeTitle + " " + title
		}
	}
	return title
}

func choosePagePermission(resource schema.Resource, page schema.Page, scope string) string {
	action := page.PermissionAction()
	if action == "" {
		return ""
	}
	for _, permission := range resource.Permissions {
		permScope, permAction, ok := permission.PolicyParts()
		if !ok || permAction != action {
			continue
		}
		if permScope == "" || permScope == schema.NormalizeName(scope) || permScope == schema.NormalizeName(resource.Module) || permScope == schema.NormalizeName(resource.Entity.Name) || permScope == schema.NormalizeName(resource.Name) {
			return permScope + ":" + permAction
		}
	}
	if scope == "" {
		return action
	}
	return schema.NormalizeName(scope) + ":" + action
}

func previewDSLResources(resources []schema.Resource, force bool) error {
	report := BuildDSLExecutionReport(resources, force, true)
	for _, message := range report.Messages {
		if _, err := fmt.Fprintln(os.Stdout, message); err != nil {
			return err
		}
	}
	for _, item := range report.Items {
		if _, err := fmt.Fprintf(os.Stdout, "resource[%d] kind=%s name=%s\n", item.Index, item.Kind, item.Name); err != nil {
			return err
		}
		for _, action := range item.Actions {
			if _, err := fmt.Fprintf(os.Stdout, "  - %s\n", action); err != nil {
				return err
			}
		}
	}
	return nil
}

func describeSchemaResourceActions(resource schema.Resource) []string {
	actions := make([]string, 0)
	forceLabel := ""
	if resource.Force {
		forceLabel = " (force)"
	}
	if moduleName := chooseModuleName(resource); moduleName != "" {
		actions = append(actions, fmt.Sprintf("generate module %q%s", moduleName, forceLabel))
	}
	if hasCRUDContent(resource) {
		crudName := chooseCRUDName(resource)
		if crudName != "" {
			actions = append(actions, fmt.Sprintf("generate CRUD %q%s", crudName, forceLabel))
		}
	}
	if pluginName := choosePluginName(resource); pluginName != "" && (resource.Kind == schema.KindPlugin || resource.Kind == schema.KindBackendPlugin) {
		actions = append(actions, fmt.Sprintf("generate plugin %q%s", pluginName, forceLabel))
	}
	if len(resource.Pages) > 0 || resource.Kind == schema.KindFrontendPage || resource.Kind == schema.KindFrontendModuleRoute {
		scope := choosePageScope(resource)
		for _, page := range resource.Pages {
			actions = append(actions, fmt.Sprintf("generate page %q -> %s%s", page.Title(), page.RoutePath(scope), forceLabel))
		}
		if len(resource.Pages) == 0 {
			actions = append(actions, fmt.Sprintf("generate page scaffold for scope %q%s", scope, forceLabel))
		}
	}
	if len(resource.Routes) > 0 {
		for _, route := range resource.Routes {
			if path, method, ok := route.PolicyParts(); ok {
				actions = append(actions, fmt.Sprintf("append policy %s %s", method, path))
			}
		}
	}
	if len(resource.Permissions) > 0 {
		for _, permission := range resource.Permissions {
			if scope, action, ok := permission.PolicyParts(); ok {
				actions = append(actions, fmt.Sprintf("tag permission %s:%s", scope, action))
			}
		}
	}
	return actions
}

func toSchemaFields(fields []legacygenerate.Field) []schema.Field {
	if len(fields) == 0 {
		return nil
	}
	result := make([]schema.Field, 0, len(fields))
	for _, field := range fields {
		result = append(result, schema.Field{
			Name:    field.Name,
			Type:    field.GoType,
			Primary: field.Primary,
			Index:   field.Index,
			Unique:  field.Unique,
		})
	}
	return result
}

func usage() {
	fmt.Fprintln(os.Stderr, strings.TrimSpace(`
Usage:
  goadmin-cli generate module <name> [--force]
  goadmin-cli generate crud <name> [--fields name:string,status:string] [--primary id] [--index name] [--unique code] [--frontend] [--policy] [--force]
  goadmin-cli generate plugin <name> [--force]
  goadmin-cli generate dsl <dsl.yaml> [--force]

Examples:
  goadmin-cli generate module user
  goadmin-cli generate crud order --fields id:string,name:string,status:string --policy
  goadmin-cli generate plugin demo
  goadmin-cli generate dsl deploy/codegen/inventory.yaml
`))
}

func shouldSkipManifest(resource schema.Resource) bool {
	switch resource.Kind {
	case schema.KindPlugin, schema.KindBackendPlugin, schema.KindConfig:
		return true
	default:
		return false
	}
}

func buildManifestOptions(resource schema.Resource, force bool) legacygenerate.ManifestOptions {
	name := chooseModuleName(resource)
	if name == "" {
		name = chooseCRUDName(resource)
	}
	if name == "" {
		name = strings.TrimSpace(resource.Name)
	}
	if name == "" {
		name = "manifest"
	}
	return legacygenerate.ManifestOptions{
		Name:        name,
		Module:      chooseModuleName(resource),
		Kind:        manifestKind(resource),
		Routes:      buildManifestRoutes(resource),
		Menus:       buildManifestMenus(resource),
		Permissions: buildManifestPermissions(resource),
		Force:       force,
	}
}

func buildConfigOptions(resource schema.Resource, force bool) legacygenerate.ConfigOptions {
	name := strings.TrimSpace(resource.Name)
	if name == "" {
		name = chooseModuleName(resource)
	}
	if name == "" {
		name = "generated"
	}
	return legacygenerate.ConfigOptions{Name: name, Module: chooseModuleName(resource), Force: force}
}

func manifestKind(resource schema.Resource) string {
	switch resource.Kind {
	case schema.KindModule, schema.KindBusinessModule, schema.KindBackendModule:
		return "business-module"
	case schema.KindCRUD, schema.KindBackendCRUD:
		return "crud"
	case schema.KindFrontendPage, schema.KindFrontendModuleRoute:
		return "frontend-module-route"
	case schema.KindManifest:
		return "manifest"
	default:
		return "business-module"
	}
}

func buildManifestRoutes(resource schema.Resource) []legacygenerate.ManifestRoute {
	routes := make([]legacygenerate.ManifestRoute, 0, len(resource.Routes))
	for _, route := range resource.Routes {
		path, method, ok := route.PolicyParts()
		if !ok {
			continue
		}
		routes = append(routes, legacygenerate.ManifestRoute{Method: method, Path: path})
	}
	if len(routes) > 0 {
		return routes
	}
	if !hasCRUDContent(resource) {
		return nil
	}
	name := chooseCRUDName(resource)
	if name == "" {
		name = chooseModuleName(resource)
	}
	if name == "" {
		name = "item"
	}
	plural := legacygenerate.Pluralize(schema.NormalizeName(name))
	base := "/api/v1/" + plural
	return []legacygenerate.ManifestRoute{
		{Method: "GET", Path: base},
		{Method: "GET", Path: base + "/:id"},
		{Method: "POST", Path: base},
		{Method: "PUT", Path: base + "/:id"},
		{Method: "DELETE", Path: base + "/:id"},
	}
}

func buildManifestMenus(resource schema.Resource) []legacygenerate.ManifestMenu {
	pages := resource.Pages
	if len(pages) == 0 {
		if resource.Kind == schema.KindFrontendPage || resource.Kind == schema.KindFrontendModuleRoute {
			pages = []schema.Page{{Name: resource.Name, Type: string(resource.Kind)}}
		}
	}
	if len(pages) == 0 {
		return nil
	}
	scope := choosePageScope(resource)
	menus := make([]legacygenerate.ManifestMenu, 0, len(pages)+1)
	firstPath := pages[0].RoutePath(scope)
	if firstPath == "" {
		firstPath = "/" + scope
	}
	menus = append(menus, legacygenerate.ManifestMenu{
		Name:       titleizeScope(scope),
		Path:       "/" + scope,
		Component:  "Layout",
		Icon:       "menu",
		Permission: scope + ":view",
		Type:       "directory",
		Redirect:   firstPath,
		Visible:    true,
		Enabled:    true,
		Sort:       1,
	})
	for index, page := range pages {
		menus = append(menus, legacygenerate.ManifestMenu{
			Name:       page.Title(),
			Path:       page.RoutePath(scope),
			Component:  page.ComponentName(scope),
			Icon:       "menu",
			Permission: choosePagePermission(resource, page, scope),
			Type:       "menu",
			Visible:    true,
			Enabled:    true,
			Sort:       index + 2,
		})
	}
	return menus
}

func buildManifestPermissions(resource schema.Resource) []legacygenerate.ManifestPermission {
	permissions := make([]legacygenerate.ManifestPermission, 0)
	if len(resource.Permissions) > 0 {
		for _, permission := range resource.Permissions {
			resourceName, actions, ok := permission.StandardActions()
			if !ok {
				continue
			}
			description := strings.TrimSpace(permission.Name)
			if description == "" {
				description = legacygenerate.ToCamel(resourceName)
			}
			for _, action := range actions {
				permissions = append(permissions, legacygenerate.ManifestPermission{
					Object:      resourceName,
					Action:      action,
					Description: fmt.Sprintf("%s %s", titleizeAction(action), description),
				})
			}
		}
		if len(permissions) > 0 {
			return permissions
		}
	}
	name := chooseCRUDName(resource)
	if name == "" {
		name = chooseModuleName(resource)
	}
	if name == "" {
		name = strings.TrimSpace(resource.Name)
	}
	if name == "" {
		name = "item"
	}
	label := legacygenerate.ToCamel(name)
	return []legacygenerate.ManifestPermission{
		{Object: schema.NormalizeName(name), Action: "list", Description: "List " + label},
		{Object: schema.NormalizeName(name), Action: "view", Description: "View " + label},
		{Object: schema.NormalizeName(name), Action: "create", Description: "Create " + label},
		{Object: schema.NormalizeName(name), Action: "update", Description: "Update " + label},
		{Object: schema.NormalizeName(name), Action: "delete", Description: "Delete " + label},
	}
}

func titleizeScope(scope string) string {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		return "Page"
	}
	return legacygenerate.ToCamel(scope)
}

func titleizeAction(action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		return "Action"
	}
	return strings.ToUpper(action[:1]) + action[1:]
}

func pagePolicyMethod(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "list", "view", "export":
		return "GET"
	case "create":
		return "POST"
	case "update":
		return "PUT"
	case "delete":
		return "DELETE"
	default:
		return ""
	}
}

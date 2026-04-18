package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	legacygenerate "goadmin/cli/generate"
	deletionapp "goadmin/codegen/application/deletion"
	deletionmodel "goadmin/codegen/model/deletion"
	"goadmin/codegen/planner"
	"goadmin/codegen/schema"
	menuservice "goadmin/modules/menu/application/service"
)

type Dependencies struct {
	MenuService   *menuservice.Service
	PolicyCleanup *deletionapp.PolicyCleanupService
	PolicyStore   string
}

func Run(root string, args []string) error {
	return RunWithDependencies(root, args, Dependencies{})
}

func RunWithDependencies(root string, args []string, deps Dependencies) error {
	if len(args) == 0 {
		return errors.New("generate requires a subcommand: module, crud, plugin, dsl, db, remove")
	}
	gen := legacygenerate.New(root)
	plan := planner.New()
	deleteService := deletionapp.NewService(deletionapp.Dependencies{
		ProjectRoot:   root,
		BackendRoot:   filepath.Join(root, "backend"),
		PolicyStore:   deps.PolicyStore,
		MenuService:   deps.MenuService,
		PolicyCleanup: deps.PolicyCleanup,
	})

	switch args[0] {
	case "generate":
		return runGenerate(root, gen, plan, args[1:])
	case "remove":
		return runRemove(root, deleteService, args[1:])
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

func runGenerate(root string, gen *legacygenerate.Generator, plan planner.Default, args []string) error {
	if len(args) == 0 {
		return errors.New("generate requires a subcommand: module, crud, plugin, dsl, db")
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
	case "db":
		return runGenerateDB(root, args[1:])
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

func printDeletionResultReport(result deletionmodel.DeleteResult) error {
	if _, err := fmt.Fprintln(os.Stdout, "deletion execution result"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stdout, "module: %s status=%s started=%s finished=%s\n", result.Plan.Module, result.Status, result.StartedAt.Format(time.RFC3339), result.FinishedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if len(result.Warnings) > 0 {
		if _, err := fmt.Fprintln(os.Stdout, "warnings:"); err != nil {
			return err
		}
		for _, warning := range result.Warnings {
			if _, err := fmt.Fprintf(os.Stdout, "- %s\n", warning); err != nil {
				return err
			}
		}
	}
	if len(result.Failures) > 0 {
		if _, err := fmt.Fprintln(os.Stdout, "failures:"); err != nil {
			return err
		}
		for _, failure := range result.Failures {
			if _, err := fmt.Fprintf(os.Stdout, "- %s (%t)\n", failure.Reason, failure.Recoverable); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintf(os.Stdout, "summary: source=%d runtime=%d registry=%d policy=%d frontend=%d skipped=%d failed=%d total_deleted=%d elapsed_ms=%d\n", result.Summary.DeletedSourceFiles, result.Summary.DeletedRuntimeAssets, result.Summary.DeletedRegistryChanges, result.Summary.DeletedPolicyChanges, result.Summary.DeletedFrontendChanges, result.Summary.Skipped, result.Summary.Failed, result.Summary.TotalDeleted, result.Summary.ElapsedMillis); err != nil {
		return err
	}
	return nil
}

func runRemove(root string, deletion *deletionapp.Service, args []string) error {
	if len(args) == 0 {
		return errors.New("remove requires a subcommand: preview, execute")
	}
	switch args[0] {
	case "preview":
		return runRemovePreview(root, deletion, args[1:])
	case "execute":
		return runRemoveExecute(root, deletion, args[1:])
	default:
		return fmt.Errorf("unknown remove subcommand %q", args[0])
	}
}

func runRemovePreview(root string, deletion *deletionapp.Service, args []string) error {
	fs := flag.NewFlagSet("remove preview", flag.ContinueOnError)
	kind := fs.String("kind", "crud", "deletion kind (crud, module, plugin)")
	force := fs.Bool("force", false, "confirm deletion scope in preview output")
	withPolicy := fs.Bool("with-policy", true, "include policy cleanup candidates")
	withRuntime := fs.Bool("with-runtime", true, "include runtime cleanup candidates")
	withFrontend := fs.Bool("with-frontend", true, "include frontend cleanup candidates")
	withRegistry := fs.Bool("with-registry", true, "include bootstrap registry cleanup candidates")
	policyStore := fs.String("policy-store", "", "policy store backend (csv or db)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("remove preview requires a module name")
	}
	if deletion == nil {
		return errors.New("deletion preview service is required")
	}
	_ = root
	report, err := deletion.Preview(deletionmodel.DeleteRequest{
		Module:       fs.Arg(0),
		Kind:         strings.TrimSpace(*kind),
		DryRun:       true,
		Force:        *force,
		WithPolicy:   *withPolicy,
		WithRuntime:  *withRuntime,
		WithFrontend: *withFrontend,
		WithRegistry: *withRegistry,
		PolicyStore:  deletionmodel.NormalizePolicyStoreKind(*policyStore),
	})
	if err != nil {
		return err
	}
	return printDeletionPreviewReport(report)
}

func runRemoveExecute(root string, deletion *deletionapp.Service, args []string) error {
	fs := flag.NewFlagSet("remove execute", flag.ContinueOnError)
	kind := fs.String("kind", "crud", "deletion kind (crud, module, plugin)")
	force := fs.Bool("force", false, "confirm deletion scope in execution output")
	withPolicy := fs.Bool("with-policy", true, "include policy cleanup candidates")
	withRuntime := fs.Bool("with-runtime", true, "include runtime cleanup candidates")
	withFrontend := fs.Bool("with-frontend", true, "include frontend cleanup candidates")
	withRegistry := fs.Bool("with-registry", true, "include bootstrap registry cleanup candidates")
	policyStore := fs.String("policy-store", "", "policy store backend (csv or db)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("remove execute requires a module name")
	}
	if deletion == nil {
		return errors.New("deletion execution service is required")
	}
	_ = root
	result, err := deletion.Delete(deletionmodel.DeleteRequest{
		Module:       fs.Arg(0),
		Kind:         strings.TrimSpace(*kind),
		DryRun:       false,
		Force:        *force,
		WithPolicy:   *withPolicy,
		WithRuntime:  *withRuntime,
		WithFrontend: *withFrontend,
		WithRegistry: *withRegistry,
		PolicyStore:  deletionmodel.NormalizePolicyStoreKind(*policyStore),
	})
	if err != nil {
		return err
	}
	return printDeletionResultReport(result)
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
		TableComment:        strings.TrimSpace(resource.Comment),
		Database:            strings.TrimSpace(resource.Database),
		Schema:              strings.TrimSpace(resource.Schema),
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
	result := make([]legacygenerate.Field, 0, len(fields))
	for _, field := range fields {
		name := strings.TrimSpace(field.Name)
		if name == "" {
			return nil, errors.New("dsl field name is required")
		}
		typeName := strings.TrimSpace(field.Type)
		if typeName == "" {
			typeName = "string"
		}
		legacyField := legacygenerate.Field{
			Name:     name,
			GoName:   legacygenerate.ToCamel(name),
			JSONName: legacygenerate.ToSnake(name),
			GoType:   legacygenerate.ParseGoType(typeName),
			Column:   legacygenerate.ToSnake(name),
			Comment:  strings.TrimSpace(field.Comment),
			Primary:  field.Primary,
			Index:    field.Index,
			Unique:   field.Unique,
		}
		if legacyField.GoName == "" {
			legacyField.GoName = legacygenerate.ToCamel(legacyField.JSONName)
		}
		if legacyField.JSONName == "" {
			legacyField.JSONName = legacygenerate.ToSnake(legacyField.GoName)
		}
		if field.Enum != nil {
			legacyField.EnumKind = strings.TrimSpace(field.Enum.Kind)
			legacyField.EnumMode = strings.TrimSpace(field.Enum.Mode)
			legacyField.EnumDisplay = strings.TrimSpace(field.Enum.Display)
			legacyField.EnumSource = strings.TrimSpace(field.Enum.SourceRef)
			legacyField.EnumSourceRef = strings.TrimSpace(field.Enum.RemotePath)
			if len(field.Enum.Options) > 0 {
				legacyField.EnumOptions = make([]legacygenerate.EnumOption, 0, len(field.Enum.Options))
				for _, option := range field.Enum.Options {
					legacyField.EnumOptions = append(legacyField.EnumOptions, legacygenerate.EnumOption{
						Value:    strings.TrimSpace(option.Value),
						Label:    strings.TrimSpace(option.Label),
						Color:    strings.TrimSpace(option.Color),
						Disabled: option.Disabled,
						Order:    option.Order,
						Metadata: cloneAnyMap(option.Metadata),
					})
				}
			}
			if len(field.Enum.Values) > 0 {
				legacyField.EnumValues = append([]string(nil), field.Enum.Values...)
			}
		}
		if len(legacyField.EnumValues) == 0 && len(legacyField.EnumOptions) > 0 {
			legacyField.EnumValues = make([]string, 0, len(legacyField.EnumOptions))
			for _, option := range legacyField.EnumOptions {
				if value := strings.TrimSpace(option.Value); value != "" {
					legacyField.EnumValues = append(legacyField.EnumValues, value)
				}
			}
		}
		result = append(result, legacyField)
	}
	return result, nil
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
		var enum *schema.EnumField
		if field.HasEnum() {
			enum = &schema.EnumField{
				Kind:       strings.TrimSpace(field.EnumKind),
				Mode:       strings.TrimSpace(field.EnumMode),
				Display:    strings.TrimSpace(field.EnumDisplay),
				SourceRef:  strings.TrimSpace(field.EnumSource),
				RemotePath: strings.TrimSpace(field.EnumSourceRef),
				Values:     append([]string(nil), field.EnumValues...),
			}
			if len(field.EnumOptions) > 0 {
				enum.Options = make([]schema.EnumOption, 0, len(field.EnumOptions))
				for _, option := range field.EnumOptions {
					enum.Options = append(enum.Options, schema.EnumOption{
						Value:    strings.TrimSpace(option.Value),
						Label:    strings.TrimSpace(option.Label),
						Color:    strings.TrimSpace(option.Color),
						Disabled: option.Disabled,
						Order:    option.Order,
						Metadata: cloneAnyMap(option.Metadata),
					})
				}
			}
		}
		result = append(result, schema.Field{
			Name:    field.Name,
			Type:    field.GoType,
			Comment: field.Comment,
			Enum:    enum,
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
  goadmin-cli generate db preview --driver mysql --dsn "..." --database goadmin [--table books] [--schema public] [--generate_frontend] [--generate_policy]
  goadmin-cli generate db generate --driver mysql --dsn "..." --database goadmin [--table books] [--schema public] [--generate_frontend] [--generate_policy]
  goadmin-cli remove preview <module> [--kind crud] [--force] [--with-policy] [--with-runtime] [--with-frontend] [--with-registry] [--policy-store csv|db]
  goadmin-cli remove execute <module> [--kind crud] [--force] [--with-policy] [--with-runtime] [--with-frontend] [--with-registry] [--policy-store csv|db]

Examples:
  goadmin-cli generate module user
  goadmin-cli generate crud order --fields id:string,name:string,status:string --policy
  goadmin-cli generate plugin demo
  goadmin-cli generate dsl deploy/codegen/inventory.yaml
  goadmin-cli generate db preview --driver sqlite --dsn "file:./tmp/codegen.db?cache=shared&mode=rwc" --database codegen --table books --generate_frontend --generate_policy
  goadmin-cli remove preview book --with-policy --with-runtime --with-frontend --with-registry --policy-store db
  goadmin-cli remove execute book --with-frontend --with-registry
`))
}

func printDeletionPreviewReport(report deletionapp.PreviewReport) error {
	if _, err := fmt.Fprintln(os.Stdout, "deletion preview report"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stdout, "module: %s kind=%s dry-run=%t force=%t policy-store=%s\n", report.Plan.Module, report.Resolution.Kind, report.Plan.DryRun, report.Plan.Force, report.Plan.PolicyStore); err != nil {
		return err
	}
	if report.Resolution.ManifestPath != "" {
		if _, err := fmt.Fprintf(os.Stdout, "manifest: %s\n", report.Resolution.ManifestPath); err != nil {
			return err
		}
	}
	if len(report.Plan.Warnings) > 0 {
		if _, err := fmt.Fprintln(os.Stdout, "warnings:"); err != nil {
			return err
		}
		for _, warning := range report.Plan.Warnings {
			if _, err := fmt.Fprintf(os.Stdout, "- %s\n", warning); err != nil {
				return err
			}
		}
	}
	if len(report.Plan.Conflicts) > 0 {
		if _, err := fmt.Fprintln(os.Stdout, "conflicts:"); err != nil {
			return err
		}
		for _, conflict := range report.Plan.Conflicts {
			if _, err := fmt.Fprintf(os.Stdout, "- %s [%s] %s\n", conflict.Path, conflict.Severity, conflict.Message); err != nil {
				return err
			}
		}
	}
	printItems := func(title string, items []deletionmodel.DeleteItem) error {
		if len(items) == 0 {
			return nil
		}
		if _, err := fmt.Fprintf(os.Stdout, "%s:\n", title); err != nil {
			return err
		}
		for _, item := range items {
			if _, err := fmt.Fprintf(os.Stdout, "- %s [%s] origin=%s managed=%t", item.Path, item.Kind, item.Origin, item.Managed); err != nil {
				return err
			}
			if item.Ref != "" {
				if _, err := fmt.Fprintf(os.Stdout, " ref=%s", item.Ref); err != nil {
					return err
				}
			}
			if item.Store != "" {
				if _, err := fmt.Fprintf(os.Stdout, " store=%s", item.Store); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(os.Stdout); err != nil {
				return err
			}
		}
		return nil
	}
	if err := printItems("source files", report.Plan.SourceFiles); err != nil {
		return err
	}
	if err := printItems("runtime assets", report.Plan.RuntimeAssets); err != nil {
		return err
	}
	if err := printItems("registry changes", report.Plan.RegistryChanges); err != nil {
		return err
	}
	if err := printItems("policy changes", report.Plan.PolicyChanges); err != nil {
		return err
	}
	if err := printItems("frontend changes", report.Plan.FrontendChanges); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(os.Stdout, "summary: source=%d runtime=%d registry=%d policy=%d frontend=%d warnings=%d conflicts=%d total=%d\n", report.Plan.Summary.SourceFiles, report.Plan.Summary.RuntimeAssets, report.Plan.Summary.RegistryChanges, report.Plan.Summary.PolicyChanges, report.Plan.Summary.FrontendChanges, report.Plan.Summary.Warnings, report.Plan.Summary.Conflicts, report.Plan.Summary.Total); err != nil {
		return err
	}
	return nil
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
	mountParentPath := strings.TrimSpace(resource.MountParentPath)
	pages := resource.Pages
	if len(pages) == 0 {
		if resource.Kind == schema.KindFrontendPage || resource.Kind == schema.KindFrontendModuleRoute {
			pages = []schema.Page{{Name: resource.Name, Type: string(resource.Kind)}}
		}
	}
	if len(pages) == 0 && hasCRUDContent(resource) {
		return buildCRUDManifestMenus(resource, mountParentPath)
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
		ParentPath: mountParentPath,
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
			ParentPath: "/" + scope,
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

func buildCRUDManifestMenus(resource schema.Resource, mountParentPath string) []legacygenerate.ManifestMenu {
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
	entityLower := schema.NormalizeName(name)
	if entityLower == "" {
		entityLower = "item"
	}
	entityPlural := legacygenerate.Pluralize(entityLower)
	if entityPlural == "" {
		entityPlural = entityLower
	}
	rootPath := "/" + entityPlural
	listPath := rootPath + "/list"
	if entityPlural == "" {
		rootPath = "/" + entityLower
		listPath = rootPath
	}
	rootName := legacygenerate.ToCamel(entityPlural)
	if rootName == "" {
		rootName = legacygenerate.ToCamel(entityLower)
	}
	if rootName == "" {
		rootName = "Menu"
	}
	return []legacygenerate.ManifestMenu{
		{Name: rootName, Path: rootPath, ParentPath: mountParentPath, Component: "Layout", Icon: "menu", Permission: entityLower + ":view", Type: "directory", Redirect: listPath, Visible: true, Enabled: true, Sort: 1},
		{Name: "List", Path: listPath, ParentPath: rootPath, Component: "view/" + entityLower + "/index", Icon: "menu", Permission: entityLower + ":list", Type: "menu", Visible: true, Enabled: true, Sort: 2},
	}
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

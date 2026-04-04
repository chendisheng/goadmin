package irbuilder

import (
	"bytes"
	"fmt"
	"strings"

	insp "goadmin/codegen/infrastructure/inspector"
	modelpkg "goadmin/codegen/model"
	irmodel "goadmin/codegen/model/ir"
	"goadmin/codegen/planner"
	"goadmin/codegen/schema"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// BuildSchemaDocument inspects a database and converts the result into the
// schema.Document form consumed by the existing DSL and planner pipeline.
func (s *Service) BuildSchemaDocument(db *gorm.DB, database string, schemaName string) (schema.Document, error) {
	return s.BuildSchemaDocumentWithOptions(db, database, schemaName, DatabaseBuildOptions{})
}

// ConvertIRDocumentToSchemaDocument converts an IR document into the schema
// document form consumed by the existing DSL and planner pipeline.
func ConvertIRDocumentToSchemaDocument(doc irmodel.Document) schema.Document {
	return convertIRDocumentWithOptions(doc, DatabaseBuildOptions{})
}

// ConvertIRDocumentToSchemaDocumentWithOptions converts an IR document into
// the schema document form while applying optional table filtering.
func ConvertIRDocumentToSchemaDocumentWithOptions(doc irmodel.Document, opts DatabaseBuildOptions) schema.Document {
	return convertIRDocumentWithOptions(doc, opts)
}

// BuildSchemaDocumentWithOptions inspects a database and converts the result
// into the schema.Document form consumed by the existing DSL and planner
// pipeline.
func (s *Service) BuildSchemaDocumentWithOptions(db *gorm.DB, database string, schemaName string, opts DatabaseBuildOptions) (schema.Document, error) {
	irDoc, err := s.BuildFromDatabaseWithOptions(db, database, schemaName, opts)
	if err != nil {
		return schema.Document{}, err
	}
	return convertIRDocumentWithOptions(irDoc, opts), nil
}

// BuildSchemaDocumentFromReader converts an already-inspected reader into the
// schema.Document form consumed by the existing DSL and planner pipeline.
func (s *Service) BuildSchemaDocumentFromReader(reader insp.Reader) (schema.Document, error) {
	return s.BuildSchemaDocumentFromReaderWithOptions(reader, DatabaseBuildOptions{})
}

// BuildSchemaDocumentFromReaderWithOptions converts an already-inspected
// reader into the schema.Document form consumed by the existing DSL and
// planner pipeline.
func (s *Service) BuildSchemaDocumentFromReaderWithOptions(reader insp.Reader, opts DatabaseBuildOptions) (schema.Document, error) {
	irDoc, err := s.BuildFromReaderWithOptions(reader, opts)
	if err != nil {
		return schema.Document{}, err
	}
	return convertIRDocumentWithOptions(irDoc, opts), nil
}

// BuildDSLDocument serializes the schema.Document form into YAML DSL bytes.
func (s *Service) BuildDSLDocument(db *gorm.DB, database string, schemaName string) ([]byte, error) {
	return s.BuildDSLDocumentWithOptions(db, database, schemaName, DatabaseBuildOptions{})
}

// BuildDSLDocumentWithOptions serializes the schema.Document form into YAML DSL
// bytes while allowing table filtering.
func (s *Service) BuildDSLDocumentWithOptions(db *gorm.DB, database string, schemaName string, opts DatabaseBuildOptions) ([]byte, error) {
	doc, err := s.BuildSchemaDocumentWithOptions(db, database, schemaName, opts)
	if err != nil {
		return nil, err
	}
	return s.SerializeSchemaDocument(doc)
}

// BuildDSLDocumentFromReader serializes a reader-backed schema.Document into YAML DSL bytes.
func (s *Service) BuildDSLDocumentFromReader(reader insp.Reader) ([]byte, error) {
	return s.BuildDSLDocumentFromReaderWithOptions(reader, DatabaseBuildOptions{})
}

// BuildDSLDocumentFromReaderWithOptions serializes a reader-backed
// schema.Document into YAML DSL bytes while allowing table filtering.
func (s *Service) BuildDSLDocumentFromReaderWithOptions(reader insp.Reader, opts DatabaseBuildOptions) ([]byte, error) {
	doc, err := s.BuildSchemaDocumentFromReaderWithOptions(reader, opts)
	if err != nil {
		return nil, err
	}
	return s.SerializeSchemaDocument(doc)
}

// SerializeSchemaDocument turns a schema.Document into the YAML DSL format used
// by the existing planner and generator pipeline.
func (s *Service) SerializeSchemaDocument(doc schema.Document) ([]byte, error) {
	if s == nil {
		return nil, fmt.Errorf("ir builder service is required")
	}
	data, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal schema document: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("schema document marshaled to empty payload")
	}
	if !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, '\n')
	}
	return data, nil
}

// PlanSchemaDocument runs the existing planner over a schema.Document.
func (s *Service) PlanSchemaDocument(doc schema.Document) (modelpkg.Plan, error) {
	if s == nil {
		return modelpkg.Plan{}, fmt.Errorf("ir builder service is required")
	}
	return planner.New().Plan(doc)
}

// PlanSchemaDocumentWithOptions runs the existing planner over a schema.Document
// after applying optional database-style table filtering.
func (s *Service) PlanSchemaDocumentWithOptions(doc schema.Document, opts DatabaseBuildOptions) (modelpkg.Plan, error) {
	if s == nil {
		return modelpkg.Plan{}, fmt.Errorf("ir builder service is required")
	}
	if len(opts.Tables) > 0 {
		filtered := filterSchemaDocumentTables(doc, opts.Tables)
		doc.Resources = filtered
	}
	return planner.New().Plan(doc)
}

// PlanFromDatabase inspects a database, converts it to schema.Document, and
// then runs the existing planner.
func (s *Service) PlanFromDatabase(db *gorm.DB, database string, schemaName string) (modelpkg.Plan, error) {
	return s.PlanFromDatabaseWithOptions(db, database, schemaName, DatabaseBuildOptions{})
}

// PlanFromDatabaseWithOptions inspects a database, converts it to
// schema.Document, and then runs the existing planner.
func (s *Service) PlanFromDatabaseWithOptions(db *gorm.DB, database string, schemaName string, opts DatabaseBuildOptions) (modelpkg.Plan, error) {
	doc, err := s.BuildSchemaDocumentWithOptions(db, database, schemaName, opts)
	if err != nil {
		return modelpkg.Plan{}, err
	}
	return s.PlanSchemaDocument(doc)
}

// PlanFromReader converts an inspector reader into schema.Document and then
// runs the existing planner.
func (s *Service) PlanFromReader(reader insp.Reader) (modelpkg.Plan, error) {
	return s.PlanFromReaderWithOptions(reader, DatabaseBuildOptions{})
}

// PlanFromReaderWithOptions converts an inspector reader into schema.Document
// and then runs the existing planner.
func (s *Service) PlanFromReaderWithOptions(reader insp.Reader, opts DatabaseBuildOptions) (modelpkg.Plan, error) {
	doc, err := s.BuildSchemaDocumentFromReaderWithOptions(reader, opts)
	if err != nil {
		return modelpkg.Plan{}, err
	}
	return s.PlanSchemaDocument(doc)
}

func convertIRDocument(doc irmodel.Document) schema.Document {
	return convertIRDocumentWithOptions(doc, DatabaseBuildOptions{})
}

func convertIRDocumentWithOptions(doc irmodel.Document, opts DatabaseBuildOptions) schema.Document {
	result := schema.Document{
		Version: normalizeIRVersion(doc.Version),
	}
	if len(doc.Resources) == 0 {
		return result
	}
	result.Resources = make([]schema.Resource, 0, len(doc.Resources))
	for _, resource := range doc.Resources {
		result.Resources = append(result.Resources, convertIRResource(resource))
	}
	if len(opts.Tables) > 0 {
		result.Resources = filterSchemaDocumentTables(result, opts.Tables)
	}
	if len(result.Resources) == 1 {
		first := result.Resources[0]
		result.Module = first.Module
		result.Kind = first.Kind
		result.Framework = first.Framework
		result.Entity = first.Entity
		result.Pages = append([]schema.Page(nil), first.Pages...)
		result.Permissions = append([]schema.Permission(nil), first.Permissions...)
		result.Routes = append([]schema.Route(nil), first.Routes...)
		result.Plugin = first.Plugin
	}
	return result
}

func filterSchemaDocumentTables(doc schema.Document, allowed []string) []schema.Resource {
	resources, err := doc.ResolveResources()
	if err != nil || len(resources) == 0 {
		return doc.Resources
	}
	return filterSchemaResources(resources, allowed)
}

func filterSchemaResources(resources []schema.Resource, allowed []string) []schema.Resource {
	if len(resources) == 0 || len(allowed) == 0 {
		return resources
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
		return resources
	}
	filtered := make([]schema.Resource, 0, len(resources))
	for _, resource := range resources {
		if _, ok := allowedSet[strings.ToLower(strings.TrimSpace(resource.Name))]; ok {
			filtered = append(filtered, resource)
		}
	}
	if len(filtered) == 0 {
		return resources
	}
	return filtered
}

func convertIRResource(resource irmodel.Resource) schema.Resource {
	name := normalizeSchemaName(resource.Module)
	if name == "" {
		name = normalizeSchemaName(resource.Name)
	}
	if name == "" {
		name = "resource"
	}
	kind := strings.TrimSpace(resource.Kind)
	if kind == "" {
		kind = string(schema.KindCRUD)
	}
	result := schema.Resource{
		Kind:             schema.Kind(kind),
		Name:             name,
		Module:           normalizeSchemaName(resource.Module),
		MountParentPath:  strings.TrimSpace(func() string { text, _ := stringMetadata(resource.Metadata, "mount_parent_path"); return text }()),
		Entity:           schema.Entity{Name: name},
		Fields:           make([]schema.Field, 0, len(resource.Fields)),
		Pages:            make([]schema.Page, 0, len(resource.Pages)),
		Permissions:      make([]schema.Permission, 0, len(resource.Permissions)),
		Routes:           make([]schema.Route, 0, len(resource.Routes)),
		GenerateFrontend: shouldGenerateFrontend(resource),
		GeneratePolicy:   shouldGeneratePolicy(resource),
	}
	for _, field := range resource.Fields {
		result.Fields = append(result.Fields, convertIRField(field))
	}
	result.Entity.Fields = append([]schema.Field(nil), result.Fields...)
	for _, page := range resource.Pages {
		result.Pages = append(result.Pages, convertIRPage(page, result.Module))
	}
	for _, permission := range resource.Permissions {
		result.Permissions = append(result.Permissions, convertIRPermission(permission, result.Name))
	}
	for _, route := range resource.Routes {
		result.Routes = append(result.Routes, convertIRRoute(route))
	}
	if plugin := convertIRPlugin(resource); plugin != nil {
		result.Plugin = plugin
	}
	if force, ok := boolMetadata(resource.Metadata, "force"); ok {
		result.Force = force
	}
	if kind == string(schema.KindPlugin) && result.Plugin == nil {
		result.Plugin = &schema.Plugin{Name: result.Name}
	}
	return result
}

func convertIRField(field irmodel.Field) schema.Field {
	name := strings.TrimSpace(field.ColumnName)
	if name == "" {
		name = strings.TrimSpace(field.Name)
	}
	typeName := strings.TrimSpace(field.GoType)
	if typeName == "" {
		typeName = strings.TrimSpace(field.DBType)
	}
	if typeName == "" {
		typeName = "string"
	}
	return schema.Field{
		Name:     name,
		Type:     typeName,
		Primary:  field.Primary,
		Index:    field.Index,
		Unique:   field.Unique,
		Required: field.Required || (!field.Nullable && !field.Primary),
	}
}

func convertIRPage(page irmodel.Page, scope string) schema.Page {
	name := strings.TrimSpace(page.Name)
	if name == "" {
		name = strings.TrimSpace(page.Title)
	}
	if name == "" {
		name = strings.TrimSpace(page.Type)
	}
	if name == "" {
		name = "page"
	}
	result := schema.Page{
		Name:      name,
		Type:      strings.TrimSpace(page.Type),
		Path:      strings.TrimSpace(page.Path),
		Component: strings.TrimSpace(page.Component),
	}
	if result.Path == "" {
		result.Path = "/" + normalizeSchemaName(scope) + "/" + normalizeSchemaName(name)
	}
	if result.Component == "" {
		result.Component = "view/" + normalizeSchemaName(scope) + "/" + normalizeSchemaName(name)
	}
	return result
}

func convertIRPermission(permission irmodel.Permission, resourceName string) schema.Permission {
	name := strings.TrimSpace(permission.Name)
	if name == "" {
		action := strings.TrimSpace(permission.Action)
		if action == "" {
			action = "view"
		}
		name = fmt.Sprintf("%s:%s", normalizeSchemaName(resourceName), normalizeSchemaName(action))
	}
	result := schema.Permission{
		Name:     name,
		Action:   strings.TrimSpace(permission.Action),
		Resource: normalizeSchemaName(permission.Resource),
	}
	if result.Resource == "" {
		result.Resource = normalizeSchemaName(resourceName)
	}
	if result.Action == "" {
		if left, right, ok := strings.Cut(name, ":"); ok {
			result.Resource = normalizeSchemaName(left)
			result.Action = normalizeSchemaName(right)
		}
	}
	return result
}

func convertIRRoute(route irmodel.Route) schema.Route {
	method := strings.ToUpper(strings.TrimSpace(route.Method))
	if method == "" {
		method = "GET"
	}
	path := strings.TrimSpace(route.Path)
	if path == "" {
		path = "/"
	}
	name := strings.TrimSpace(route.Name)
	if name == "" {
		name = normalizeSchemaName(path)
	}
	return schema.Route{Method: method, Path: path, Name: name}
}

func convertIRPlugin(resource irmodel.Resource) *schema.Plugin {
	if strings.TrimSpace(resource.Kind) != string(schema.KindPlugin) {
		return nil
	}
	name := normalizeSchemaName(resource.Name)
	if name == "" {
		name = normalizeSchemaName(resource.Module)
	}
	if name == "" {
		name = "plugin"
	}
	return &schema.Plugin{Name: name}
}

func shouldGenerateFrontend(resource irmodel.Resource) bool {
	if flag, ok := boolMetadata(resource.Metadata, "generate_frontend"); ok {
		return flag
	}
	kind := strings.ToLower(strings.TrimSpace(resource.Kind))
	return len(resource.Fields) > 0 || len(resource.Pages) > 0 || kind == "crud" || kind == "business-module" || kind == "frontend-page" || kind == "frontend-module-route"
}

func shouldGeneratePolicy(resource irmodel.Resource) bool {
	if flag, ok := boolMetadata(resource.Metadata, "generate_policy"); ok {
		return flag
	}
	return len(resource.Permissions) > 0 || len(resource.Routes) > 0 || len(resource.Fields) > 0
}

func boolMetadata(metadata map[string]any, key string) (bool, bool) {
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

func normalizeIRVersion(value string) string {
	if version := strings.TrimSpace(value); version != "" {
		return version
	}
	return defaultIRVersion
}

func normalizeSchemaName(value string) string {
	return schema.NormalizeName(value)
}

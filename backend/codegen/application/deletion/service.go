package deletion

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	deletionmodel "goadmin/codegen/model/deletion"
	codegenpostprocess "goadmin/codegen/postprocess"

	"gopkg.in/yaml.v3"
)

type Service struct {
	projectRoot string
	backendRoot string
	policyStore deletionmodel.PolicyStoreKind
}

func NewService(deps Dependencies) *Service {
	projectRoot := filepath.Clean(strings.TrimSpace(deps.ProjectRoot))
	if projectRoot == "" || projectRoot == "." {
		projectRoot = "."
	}
	backendRoot := filepath.Clean(strings.TrimSpace(deps.BackendRoot))
	if backendRoot == "" || backendRoot == "." {
		backendRoot = resolveBackendRoot(projectRoot)
	}
	return &Service{
		projectRoot: projectRoot,
		backendRoot: backendRoot,
		policyStore: normalizePolicyStoreSource(deps.PolicyStore),
	}
}

func (s *Service) Preview(req deletionmodel.DeleteRequest) (PreviewReport, error) {
	if s == nil {
		return PreviewReport{}, errors.New("deletion planner service is required")
	}
	normalized := req.Normalize()
	moduleName := NormalizeModuleName(normalized.Module)
	if moduleName == "" {
		return PreviewReport{}, fmt.Errorf("module is required")
	}
	normalized.Module = moduleName
	if strings.TrimSpace(normalized.Kind) == "" {
		normalized.Kind = "crud"
	}
	resolution, err := s.resolveModule(moduleName, normalized)
	if err != nil {
		return PreviewReport{}, err
	}
	plan, err := s.buildPlan(normalized, resolution)
	if err != nil {
		return PreviewReport{}, err
	}
	return PreviewReport{
		Request:    normalized,
		Resolution: resolution,
		Plan:       plan,
	}, nil
}

func (s *Service) Plan(req deletionmodel.DeleteRequest) (deletionmodel.DeletePlan, error) {
	report, err := s.Preview(req)
	if err != nil {
		return deletionmodel.DeletePlan{}, err
	}
	return report.Plan, nil
}

func (s *Service) resolveModule(moduleName string, req deletionmodel.DeleteRequest) (ModuleResolution, error) {
	moduleDir, backendRoot, err := s.locateModuleDir(moduleName)
	if err != nil {
		return ModuleResolution{}, err
	}
	displayRoot := s.displayRoot()
	moduleGoPath := filepath.Join(moduleDir, "module.go")
	bootstrapPath := filepath.Join(moduleDir, "bootstrap.go")
	manifestPath := filepath.Join(moduleDir, "manifest.yaml")
	altManifestPath := filepath.Join(moduleDir, "manifest.yml")
	codegenManifestPath := filepath.Join(moduleDir, "codegen.manifest.json")
	if _, err := os.Stat(manifestPath); errors.Is(err, os.ErrNotExist) {
		if _, err := os.Stat(altManifestPath); err == nil {
			manifestPath = altManifestPath
		} else if _, err := os.Stat(codegenManifestPath); err == nil {
			manifestPath = codegenManifestPath
		} else {
			manifestPath = ""
		}
	}

	moduleGo := moduleMeta{}
	if content, err := os.ReadFile(moduleGoPath); err == nil {
		moduleGo = parseModuleGoMeta(content)
	}
	generatedBootstrap := hasGeneratedBootstrap(bootstrapPath)
	registryPath := filepath.Join(backendRoot, "core", "bootstrap", "modules_gen.go")
	builtinRegistryPath := filepath.Join(backendRoot, "core", "bootstrap", "modules_builtin.go")
	isBuiltin := moduleRegistryContains(builtinRegistryPath, moduleName)
	manifestDoc := moduleManifest{}
	if manifestPath != "" {
		if content, err := os.ReadFile(manifestPath); err == nil {
			if err := yaml.Unmarshal(content, &manifestDoc); err != nil {
				return ModuleResolution{}, fmt.Errorf("parse manifest %s: %w", displayPath(displayRoot, manifestPath), err)
			}
		}
	}
	resolved := ModuleResolution{
		Input:               req.Module,
		Module:              moduleName,
		Kind:                strings.TrimSpace(req.Kind),
		ProjectRoot:         displayPath(displayRoot, s.projectRoot),
		BackendRoot:         displayPath(displayRoot, backendRoot),
		ModuleDir:           displayPath(displayRoot, moduleDir),
		ManifestPath:        displayPath(displayRoot, manifestPath),
		ModuleGoPath:        displayPath(displayRoot, moduleGoPath),
		BootstrapPath:       displayPath(displayRoot, bootstrapPath),
		RegistryPath:        displayPath(displayRoot, registryPath),
		BuiltinRegistryPath: displayPath(displayRoot, builtinRegistryPath),
		ManifestName:        strings.TrimSpace(firstNonEmpty(manifestDoc.Name, moduleGo.Name, moduleName)),
		ManifestKind:        strings.TrimSpace(manifestDoc.Kind),
		ManifestVersion:     strings.TrimSpace(manifestDoc.Version),
		GeneratedBootstrap:  generatedBootstrap,
		HasManifest:         manifestPath != "",
		HasModuleGo:         moduleGoPath != "" && fileExists(moduleGoPath),
		IsBuiltin:           isBuiltin,
		PolicyStore:         s.resolvePolicyStore(req.PolicyStore),
		Compatibility:       req.Compatibility.Normalize(),
	}
	if resolved.Kind == "" {
		resolved.Kind = strings.TrimSpace(manifestDoc.Kind)
	}
	if resolved.Kind == "" && generatedBootstrap {
		resolved.Kind = "crud"
	}
	if resolved.Kind == "" {
		resolved.Kind = "business-module"
	}
	return resolved, nil
}

func (s *Service) buildPlan(req deletionmodel.DeleteRequest, resolution ModuleResolution) (deletionmodel.DeletePlan, error) {
	moduleDir, err := s.absoluteModuleDir(resolution.Module)
	if err != nil {
		return deletionmodel.DeletePlan{}, err
	}
	displayRoot := s.displayRoot()
	plan := deletionmodel.DeletePlan{
		Request:     req,
		Module:      resolution.Module,
		DryRun:      true,
		Force:       req.Force,
		PolicyStore: resolution.PolicyStore,
		Legacy:      req.Compatibility.Normalize(),
	}
	manifestDoc, manifestErr := s.loadManifest(moduleDir)
	if manifestErr != nil && !errors.Is(manifestErr, os.ErrNotExist) {
		return deletionmodel.DeletePlan{}, manifestErr
	}
	managedByCodeGen := resolution.GeneratedBootstrap
	if !managedByCodeGen {
		plan.Warnings = append(plan.Warnings, "bootstrap.go is not marked as generated; preview uses conservative inference")
		plan.Conflicts = append(plan.Conflicts, deletionmodel.DeleteConflict{
			Kind:     "legacy-module",
			Severity: conflictSeverityHigh,
			Message:  "module bootstrap is not generated; deletion requires explicit review",
			Path:     resolution.BootstrapPath,
		})
	}
	if resolution.IsBuiltin {
		plan.Warnings = append(plan.Warnings, "module is registered as a builtin module")
		plan.Conflicts = append(plan.Conflicts, deletionmodel.DeleteConflict{
			Kind:     "builtin-module",
			Severity: conflictSeverityHigh,
			Message:  "builtin modules are not treated as generated business modules",
			Path:     resolution.ModuleDir,
		})
	}
	if manifestErr != nil && errors.Is(manifestErr, os.ErrNotExist) {
		plan.Warnings = append(plan.Warnings, "manifest.yaml not found; runtime assets were inferred from module name")
	}
	unknownFiles, sourceFiles := s.collectSourceFiles(moduleDir, displayRoot, resolution.Module, managedByCodeGen)
	if len(unknownFiles) > 0 {
		plan.Conflicts = append(plan.Conflicts, unknownFiles...)
	}
	plan.SourceFiles = append(plan.SourceFiles, sourceFiles...)
	if len(sourceFiles) > 0 && managedByCodeGen && len(unknownFiles) == 0 {
		plan.SourceFiles = append(plan.SourceFiles, deletionmodel.DeleteItem{
			Module:   resolution.Module,
			Kind:     deletionmodel.AssetKindSourceDirectory,
			Path:     displayPath(displayRoot, moduleDir),
			Origin:   deletionmodel.AssetOriginGenerated,
			Managed:  true,
			Metadata: map[string]any{"purpose": "remove empty generated module directory"},
		})
	}
	manifestToUse := manifestDoc
	if manifestErr != nil || strings.TrimSpace(manifestToUse.Module) == "" {
		if managedByCodeGen {
			manifestToUse = inferCRUDManifest(resolution.Module)
			plan.Warnings = append(plan.Warnings, "manifest preview inferred from generated module conventions")
		}
	}
	if strings.TrimSpace(manifestToUse.Module) != "" && strings.TrimSpace(manifestToUse.Module) != resolution.Module {
		plan.Conflicts = append(plan.Conflicts, deletionmodel.DeleteConflict{
			Kind:     "manifest-module-mismatch",
			Severity: conflictSeverityHigh,
			Message:  fmt.Sprintf("manifest module %q does not match requested module %q", strings.TrimSpace(manifestToUse.Module), resolution.Module),
			Path:     displayPath(displayRoot, filepath.Join(moduleDir, "manifest.yaml")),
		})
	}
	ownership := deletionmodel.ModuleOwnership{
		Module:           resolution.Module,
		Kind:             resolution.Kind,
		GeneratorVersion: strings.TrimSpace(manifestToUse.Version),
		ManifestPath:     resolution.ManifestPath,
		ManifestFormat:   manifestFormatFromPath(resolution.ManifestPath),
		Source:           ownershipSource(resolution, manifestErr),
		Compatibility:    req.Compatibility.Normalize(),
		Metadata: map[string]any{
			"generated_bootstrap": managedByCodeGen,
			"has_manifest":        resolution.HasManifest,
			"is_builtin":          resolution.IsBuiltin,
			"policy_store":        resolution.PolicyStore,
		},
	}
	if req.WithRuntime && len(manifestToUse.Routes) > 0 {
		for _, route := range manifestToUse.Routes {
			path := strings.TrimSpace(route.Path)
			method := strings.ToUpper(strings.TrimSpace(route.Method))
			if path == "" || method == "" {
				plan.Conflicts = append(plan.Conflicts, deletionmodel.DeleteConflict{
					Kind:     "invalid-route",
					Severity: conflictSeverityWarning,
					Message:  "route entry is incomplete",
				})
				continue
			}
			asset := deletionmodel.DeleteItem{
				Module:  resolution.Module,
				Kind:    deletionmodel.AssetKindRuntimeRoute,
				Path:    path,
				Ref:     method,
				Store:   resolution.PolicyStore,
				Origin:  routeOrigin(managedByCodeGen),
				Managed: managedByCodeGen,
				Metadata: map[string]any{
					"method": method,
					"path":   path,
				},
			}
			plan.RuntimeAssets = append(plan.RuntimeAssets, asset)
			if resolution.PolicyStore.IsKnown() && req.WithPolicy {
				selector := deletionmodel.PolicySelector{
					Store:     resolution.PolicyStore,
					Module:    resolution.Module,
					SourceRef: strings.TrimSpace(method + " " + path),
					PType:     "p",
					V0:        "admin",
					V1:        path,
					V2:        method,
				}
				plan.PolicyChanges = append(plan.PolicyChanges, deletionmodel.DeleteItem{
					Module:   resolution.Module,
					Kind:     deletionmodel.AssetKindPolicyRule,
					Path:     path,
					Ref:      method,
					Store:    resolution.PolicyStore,
					Selector: &selector,
					Origin:   routeOrigin(managedByCodeGen),
					Managed:  managedByCodeGen,
				})
			}
		}
	}
	if req.WithRuntime && len(manifestToUse.Menus) > 0 {
		for _, menu := range manifestToUse.Menus {
			path := strings.TrimSpace(menu.Path)
			if path == "" {
				continue
			}
			plan.RuntimeAssets = append(plan.RuntimeAssets, deletionmodel.DeleteItem{
				Module:  resolution.Module,
				Kind:    deletionmodel.AssetKindRuntimeMenu,
				Path:    path,
				Ref:     strings.TrimSpace(menu.Permission),
				Origin:  routeOrigin(managedByCodeGen),
				Managed: managedByCodeGen,
				Metadata: map[string]any{
					"name":        strings.TrimSpace(menu.Name),
					"parent_path": strings.TrimSpace(menu.ParentPath),
					"component":   strings.TrimSpace(menu.Component),
				},
			})
		}
	}
	if req.WithRuntime && len(manifestToUse.Permissions) > 0 {
		for _, permission := range manifestToUse.Permissions {
			object := strings.TrimSpace(permission.Object)
			action := strings.TrimSpace(permission.Action)
			if object == "" && action == "" {
				continue
			}
			plan.RuntimeAssets = append(plan.RuntimeAssets, deletionmodel.DeleteItem{
				Module:  resolution.Module,
				Kind:    deletionmodel.AssetKindRuntimePermission,
				Ref:     strings.TrimSpace(object + ":" + action),
				Origin:  routeOrigin(managedByCodeGen),
				Managed: managedByCodeGen,
				Metadata: map[string]any{
					"object":      object,
					"action":      action,
					"description": strings.TrimSpace(permission.Description),
				},
			})
		}
	}
	if req.WithRegistry && resolution.PolicyStore.IsKnown() {
		plan.RegistryChanges = append(plan.RegistryChanges, deletionmodel.DeleteItem{
			Module:   resolution.Module,
			Kind:     deletionmodel.AssetKindRuntimeRegistry,
			Path:     resolution.RegistryPath,
			Ref:      resolution.Module,
			Store:    resolution.PolicyStore,
			Origin:   deletionmodel.AssetOriginGenerated,
			Managed:  true,
			Metadata: map[string]any{"registry": "generated_modules"},
		})
	} else if req.WithRegistry {
		plan.Warnings = append(plan.Warnings, "policy store is unknown; registry/policy cleanup is preview-only")
	}
	if req.WithFrontend {
		frontendAssets := s.collectFrontendCandidates(resolution.Module, managedByCodeGen, displayRoot)
		plan.FrontendChanges = append(plan.FrontendChanges, frontendAssets...)
	}
	if len(plan.PolicyChanges) == 0 && req.WithPolicy {
		plan.Warnings = append(plan.Warnings, "no policy rules could be inferred for deletion")
	}
	if len(plan.FrontendChanges) == 0 && req.WithFrontend {
		plan.Warnings = append(plan.Warnings, "no frontend assets were found for deletion")
	}
	plan.Ownership = ownership
	plan.Ownership.OwnedFiles = append([]deletionmodel.DeleteItem(nil), plan.SourceFiles...)
	plan.Ownership.RuntimeAssets = append([]deletionmodel.DeleteItem(nil), plan.RuntimeAssets...)
	plan.Ownership.PolicyAssets = make([]deletionmodel.PolicyAsset, 0, len(plan.PolicyChanges))
	for _, item := range plan.PolicyChanges {
		if item.Selector == nil {
			continue
		}
		plan.Ownership.PolicyAssets = append(plan.Ownership.PolicyAssets, deletionmodel.PolicyAsset{
			Store:     item.Selector.Store,
			Module:    resolution.Module,
			SourceRef: item.Selector.SourceRef,
			PType:     item.Selector.PType,
			V0:        item.Selector.V0,
			V1:        item.Selector.V1,
			V2:        item.Selector.V2,
			V3:        item.Selector.V3,
			V4:        item.Selector.V4,
			V5:        item.Selector.V5,
			Managed:   true,
		})
	}
	plan.Ownership.FrontendAssets = append([]deletionmodel.DeleteItem(nil), plan.FrontendChanges...)
	plan.Summary = deletionmodel.DeletePlanSummary{
		SourceFiles:     len(plan.SourceFiles),
		RuntimeAssets:   len(plan.RuntimeAssets),
		RegistryChanges: len(plan.RegistryChanges),
		PolicyChanges:   len(plan.PolicyChanges),
		FrontendChanges: len(plan.FrontendChanges),
		Warnings:        len(plan.Warnings),
		Conflicts:       len(plan.Conflicts),
		Total:           len(plan.SourceFiles) + len(plan.RuntimeAssets) + len(plan.RegistryChanges) + len(plan.PolicyChanges) + len(plan.FrontendChanges),
	}
	if len(plan.Conflicts) > 0 {
		sort.SliceStable(plan.Conflicts, func(i, j int) bool {
			return plan.Conflicts[i].Kind < plan.Conflicts[j].Kind || (plan.Conflicts[i].Kind == plan.Conflicts[j].Kind && plan.Conflicts[i].Path < plan.Conflicts[j].Path)
		})
	}
	sortDeleteItems(plan.SourceFiles)
	sortDeleteItems(plan.RuntimeAssets)
	sortDeleteItems(plan.RegistryChanges)
	sortDeleteItems(plan.PolicyChanges)
	sortDeleteItems(plan.FrontendChanges)
	return plan, nil
}

func (s *Service) collectSourceFiles(moduleDir, displayRoot, module string, managed bool) ([]deletionmodel.DeleteConflict, []deletionmodel.DeleteItem) {
	var conflicts []deletionmodel.DeleteConflict
	var items []deletionmodel.DeleteItem
	known := make(map[string]struct{})
	if err := filepath.WalkDir(moduleDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			return nil
		}
		rel, relErr := filepath.Rel(moduleDir, path)
		if relErr != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)
		if kind, ok := classifyModuleSourceFile(rel); ok {
			items = append(items, deletionmodel.DeleteItem{
				Module:   module,
				Kind:     kind,
				Path:     displayPath(displayRoot, path),
				Origin:   sourceOrigin(managed),
				Managed:  managed,
				Metadata: map[string]any{"relative_path": rel},
			})
			known[rel] = struct{}{}
			return nil
		}
		conflicts = append(conflicts, deletionmodel.DeleteConflict{
			Kind:     "unknown-owned-file",
			Severity: conflictSeverityHigh,
			Message:  "file is not recognized as a generated CodeGen asset",
			Path:     displayPath(displayRoot, path),
		})
		return nil
	}); err != nil {
		conflicts = append(conflicts, deletionmodel.DeleteConflict{
			Kind:     "scan-error",
			Severity: conflictSeverityWarning,
			Message:  err.Error(),
			Path:     displayPath(displayRoot, moduleDir),
		})
	}
	_ = known
	return conflicts, items
}

func (s *Service) collectFrontendCandidates(module string, managed bool, displayRoot string) []deletionmodel.DeleteItem {
	paths := []struct {
		path string
		kind deletionmodel.AssetKind
	}{
		{path: filepath.Join(s.backendRoot, "..", "web", "src", "api", module+".ts"), kind: deletionmodel.AssetKindFrontendFile},
		{path: filepath.Join(s.backendRoot, "..", "web", "src", "router", "modules", module+".ts"), kind: deletionmodel.AssetKindFrontendFile},
		{path: filepath.Join(s.backendRoot, "..", "web", "src", "views", module, "index.vue"), kind: deletionmodel.AssetKindFrontendFile},
	}
	items := make([]deletionmodel.DeleteItem, 0, len(paths))
	for _, candidate := range paths {
		abs := filepath.Clean(candidate.path)
		if !fileExists(abs) {
			continue
		}
		items = append(items, deletionmodel.DeleteItem{
			Module:   module,
			Kind:     candidate.kind,
			Path:     displayPath(displayRoot, abs),
			Origin:   sourceOrigin(managed),
			Managed:  managed,
			Metadata: map[string]any{"relative_path": displayPath(displayRoot, abs)},
		})
	}
	return items
}

func classifyModuleSourceFile(rel string) (deletionmodel.AssetKind, bool) {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	switch rel {
	case "module.go", "bootstrap.go", "manifest.yaml", "manifest.yml", "codegen.manifest.json", "schema.sql":
		return deletionmodel.AssetKindSourceFile, true
	case "transport/http/router.go":
		return deletionmodel.AssetKindSourceFile, true
	}
	switch {
	case strings.HasPrefix(rel, "application/command/") && strings.HasSuffix(rel, ".go"):
		return deletionmodel.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "application/query/") && strings.HasSuffix(rel, ".go"):
		return deletionmodel.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "application/service/") && strings.HasSuffix(rel, ".go"):
		return deletionmodel.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "domain/model/") && strings.HasSuffix(rel, ".go"):
		return deletionmodel.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "domain/repository/") && strings.HasSuffix(rel, ".go"):
		return deletionmodel.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "infrastructure/repo/") && strings.HasSuffix(rel, ".go"):
		return deletionmodel.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "transport/http/request/") && strings.HasSuffix(rel, ".go"):
		return deletionmodel.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "transport/http/response/") && strings.HasSuffix(rel, ".go"):
		return deletionmodel.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "transport/http/handler/") && strings.HasSuffix(rel, ".go"):
		return deletionmodel.AssetKindSourceFile, true
	case strings.HasSuffix(rel, "_test.go"):
		return deletionmodel.AssetKindSourceFile, true
	default:
		return deletionmodel.AssetKindUnknown, false
	}
}

type moduleMeta struct {
	Name         string
	ManifestPath string
}

func parseModuleGoMeta(content []byte) moduleMeta {
	text := string(content)
	return moduleMeta{
		Name:         extractStringConst(text, "Name"),
		ManifestPath: extractStringConst(text, "ManifestPath"),
	}
}

func extractStringConst(text, symbol string) string {
	re := regexp.MustCompile(`(?m)^\s*const\s+` + regexp.QuoteMeta(symbol) + `\s*=\s*"([^"]+)"\s*$`)
	match := re.FindStringSubmatch(text)
	if len(match) != 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func hasGeneratedBootstrap(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return codegenpostprocess.HasGeneratedMarkers(path, content)
}

func moduleRegistryContains(path string, module string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	needle := "goadmin/modules/" + module
	return strings.Contains(string(content), needle)
}

func inferCRUDManifest(module string) moduleManifest {
	plural := pluralize(module)
	label := titleFromModule(module)
	return moduleManifest{
		Name:    module,
		Version: "v1",
		Kind:    "crud",
		Module:  module,
		Routes: []moduleManifestRoute{
			{Method: "GET", Path: "/api/v1/" + plural},
			{Method: "GET", Path: "/api/v1/" + plural + "/:id"},
			{Method: "POST", Path: "/api/v1/" + plural},
			{Method: "PUT", Path: "/api/v1/" + plural + "/:id"},
			{Method: "DELETE", Path: "/api/v1/" + plural + "/:id"},
		},
		Menus: []moduleManifestMenu{
			{Name: label + "s", Path: "/" + plural, Component: "Layout", Permission: module + ":view", Type: "directory", Redirect: "/" + plural + "/list", Visible: true, Enabled: true, Sort: 1},
			{Name: "List", Path: "/" + plural + "/list", ParentPath: "/" + plural, Component: "view/" + module + "/index", Permission: module + ":list", Type: "menu", Visible: true, Enabled: true, Sort: 2},
		},
		Permissions: []moduleManifestPermission{
			{Object: module, Action: "list", Description: "List " + label},
			{Object: module, Action: "view", Description: "View " + label},
			{Object: module, Action: "create", Description: "Create " + label},
			{Object: module, Action: "update", Description: "Update " + label},
			{Object: module, Action: "delete", Description: "Delete " + label},
		},
		Capabilities: []string{"basic-crud", "policy-generated", "frontend-generated"},
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func manifestFormatFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return "json"
	case ".yml", ".yaml":
		return "yaml"
	default:
		return ""
	}
}

func ownershipSource(resolution ModuleResolution, manifestErr error) string {
	if resolution.GeneratedBootstrap {
		return "generated"
	}
	if manifestErr != nil {
		return "inferred"
	}
	return "manifest"
}

func routeOrigin(managed bool) deletionmodel.AssetOrigin {
	if managed {
		return deletionmodel.AssetOriginGenerated
	}
	return deletionmodel.AssetOriginInferred
}

func sourceOrigin(managed bool) deletionmodel.AssetOrigin {
	if managed {
		return deletionmodel.AssetOriginGenerated
	}
	return deletionmodel.AssetOriginInferred
}

func sortDeleteItems(items []deletionmodel.DeleteItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Path == items[j].Path {
			return items[i].Kind < items[j].Kind
		}
		return items[i].Path < items[j].Path
	})
}

func (s *Service) resolvePolicyStore(req deletionmodel.PolicyStoreKind) deletionmodel.PolicyStoreKind {
	if req.IsKnown() {
		return req
	}
	if resolved := normalizePolicyStoreSource(string(req)); resolved.IsKnown() {
		return resolved
	}
	if s != nil && s.policyStore.IsKnown() {
		return s.policyStore
	}
	return deletionmodel.PolicyStoreUnknown
}

func (s *Service) locateModuleDir(module string) (string, string, error) {
	candidates := []string{
		filepath.Join(s.backendRoot, "modules", module),
		filepath.Join(s.projectRoot, "backend", "modules", module),
		filepath.Join(s.projectRoot, "modules", module),
	}
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			backendRoot := s.backendRoot
			if backendRoot == "" {
				backendRoot = resolveBackendRoot(s.projectRoot)
			}
			return candidate, backendRoot, nil
		}
	}
	return "", "", fmt.Errorf("module %q not found under backend/modules", module)
}

func (s *Service) absoluteModuleDir(module string) (string, error) {
	moduleDir, _, err := s.locateModuleDir(module)
	return moduleDir, err
}

func (s *Service) displayRoot() string {
	if strings.TrimSpace(s.projectRoot) != "" {
		return s.projectRoot
	}
	return s.backendRoot
}

func displayPath(root, target string) string {
	root = filepath.Clean(strings.TrimSpace(root))
	target = filepath.Clean(strings.TrimSpace(target))
	if root == "" || target == "" || root == "." || target == "." {
		return target
	}
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return target
	}
	return filepath.ToSlash(rel)
}

func resolveBackendRoot(projectRoot string) string {
	projectRoot = filepath.Clean(strings.TrimSpace(projectRoot))
	if projectRoot == "" || projectRoot == "." {
		return "."
	}
	backendCandidate := filepath.Join(projectRoot, "backend")
	if info, err := os.Stat(filepath.Join(backendCandidate, "modules")); err == nil && info.IsDir() {
		return backendCandidate
	}
	if info, err := os.Stat(filepath.Join(projectRoot, "modules")); err == nil && info.IsDir() {
		return projectRoot
	}
	return backendCandidate
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func loadManifestFromPath(path string) (moduleManifest, error) {
	if path == "" {
		return moduleManifest{}, os.ErrNotExist
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return moduleManifest{}, err
	}
	var doc moduleManifest
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return moduleManifest{}, err
	}
	return doc, nil
}

func (s *Service) loadManifest(moduleDir string) (moduleManifest, error) {
	for _, name := range []string{"manifest.yaml", "manifest.yml", "codegen.manifest.json"} {
		path := filepath.Join(moduleDir, name)
		if !fileExists(path) {
			continue
		}
		return loadManifestFromPath(path)
	}
	return moduleManifest{}, os.ErrNotExist
}

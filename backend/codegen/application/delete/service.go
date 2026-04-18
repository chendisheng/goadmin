package deleteapp

import (
	"context"
	"errors"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	lifecycle "goadmin/codegen/model/lifecycle"
	codegenpostprocess "goadmin/codegen/postprocess"
	menuservice "goadmin/modules/menu/application/service"
	menuModel "goadmin/modules/menu/domain/model"

	"gopkg.in/yaml.v3"
)

type Service struct {
	projectRoot   string
	backendRoot   string
	policyStore   lifecycle.PolicyStoreKind
	menuService   *menuservice.Service
	policyCleanup *PolicyCleanupService
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
		projectRoot:   projectRoot,
		backendRoot:   backendRoot,
		policyStore:   normalizePolicyStoreSource(deps.PolicyStore),
		menuService:   deps.MenuService,
		policyCleanup: deps.PolicyCleanup,
	}
}

func (s *Service) Preview(req lifecycle.DeleteRequest) (PreviewReport, error) {
	if s == nil {
		return PreviewReport{}, errors.New("delete service is required")
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

func (s *Service) Plan(req lifecycle.DeleteRequest) (lifecycle.DeletePlan, error) {
	report, err := s.Preview(req)
	if err != nil {
		return lifecycle.DeletePlan{}, err
	}
	return report.Plan, nil
}

func (s *Service) Delete(req lifecycle.DeleteRequest) (lifecycle.DeleteResult, error) {
	if s == nil {
		return lifecycle.DeleteResult{}, errors.New("delete service is required")
	}
	startedAt := nowUTC()
	normalized := req.Normalize()
	result := lifecycle.DeleteResult{
		Request:   normalized,
		Status:    lifecycle.DeleteStatusPlanned,
		StartedAt: startedAt,
		Audit: lifecycle.DeleteAuditRecord{
			Operation:   "delete",
			Module:      normalized.Module,
			Kind:        normalized.Kind,
			PolicyStore: normalized.PolicyStore,
			DryRun:      normalized.DryRun,
			Force:       normalized.Force,
			StartedAt:   startedAt,
		},
	}
	if err := normalized.Validate(); err != nil {
		finishedAt := nowUTC()
		result.Status = lifecycle.DeleteStatusFailed
		result.FinishedAt = finishedAt
		result.Audit.FinishedAt = finishedAt
		result.Failures = append(result.Failures, lifecycle.DeleteFailure{
			Category:    lifecycle.DeleteFailureCategoryValidation,
			Stage:       lifecycle.DeleteFailureStageValidation,
			Reason:      err.Error(),
			Recoverable: false,
		})
		result.Summary = summarizeDeleteResult(result.StartedAt, result.FinishedAt, result.Deleted, result.Skipped, result.Failures)
		result.Audit.Result = result.Summary
		result.Audit.Failures = summarizeDeleteFailures(result.Failures)
		result.Validation = lifecycle.DeleteValidationReport{
			StartedAt:  finishedAt,
			FinishedAt: finishedAt,
			Status:     "failed",
			Verified:   false,
			Issues: []lifecycle.DeleteValidationIssue{{
				Category: lifecycle.DeleteFailureCategoryValidation,
				Stage:    lifecycle.DeleteFailureStageValidation,
				Message:  err.Error(),
			}},
		}
		return result, err
	}
	if strings.TrimSpace(normalized.Kind) == "" {
		normalized.Kind = "crud"
	}
	report, err := s.Preview(normalized)
	if err != nil {
		finishedAt := nowUTC()
		result.Status = lifecycle.DeleteStatusFailed
		result.FinishedAt = finishedAt
		result.Audit.FinishedAt = finishedAt
		result.Failures = append(result.Failures, lifecycle.DeleteFailure{
			Category:    lifecycle.DeleteFailureCategoryValidation,
			Stage:       lifecycle.DeleteFailureStageValidation,
			Reason:      err.Error(),
			Recoverable: false,
		})
		result.Summary = summarizeDeleteResult(result.StartedAt, result.FinishedAt, result.Deleted, result.Skipped, result.Failures)
		result.Audit.Result = result.Summary
		result.Audit.Failures = summarizeDeleteFailures(result.Failures)
		result.Validation = lifecycle.DeleteValidationReport{
			StartedAt:  finishedAt,
			FinishedAt: finishedAt,
			Status:     "failed",
			Verified:   false,
			Issues: []lifecycle.DeleteValidationIssue{{
				Category: lifecycle.DeleteFailureCategoryValidation,
				Stage:    lifecycle.DeleteFailureStageValidation,
				Message:  err.Error(),
			}},
		}
		return result, err
	}
	result.Request = report.Request
	result.Plan = report.Plan
	result.Warnings = append(result.Warnings, report.Plan.Warnings...)
	result.Audit.Module = report.Plan.Module
	result.Audit.Kind = report.Request.Kind
	result.Audit.PolicyStore = report.Plan.PolicyStore
	result.Audit.DryRun = report.Request.DryRun
	result.Audit.Force = report.Request.Force
	if normalized.DryRun {
		result.Status = lifecycle.DeleteStatusDryRun
		result.FinishedAt = nowUTC()
		result.Summary = summarizeDeleteResult(result.StartedAt, result.FinishedAt, result.Deleted, result.Skipped, result.Failures)
		result.Audit.FinishedAt = result.FinishedAt
		result.Audit.Result = result.Summary
		result.Audit.Failures = summarizeDeleteFailures(result.Failures)
		result.Validation = lifecycle.DeleteValidationReport{StartedAt: result.StartedAt, FinishedAt: result.FinishedAt, Status: "skipped", Verified: true}
		return result, nil
	}
	if len(report.Plan.Conflicts) > 0 {
		result.Status = lifecycle.DeleteStatusFailed
		result.FinishedAt = nowUTC()
		result.Failures = append(result.Failures, lifecycle.DeleteFailure{
			Category:    lifecycle.DeleteFailureCategoryValidation,
			Stage:       lifecycle.DeleteFailureStageValidation,
			Reason:      fmt.Sprintf("delete plan has %d blocking conflict(s)", len(report.Plan.Conflicts)),
			Recoverable: false,
		})
		result.Summary = summarizeDeleteResult(result.StartedAt, result.FinishedAt, result.Deleted, result.Skipped, result.Failures)
		result.Audit.FinishedAt = result.FinishedAt
		result.Audit.Result = result.Summary
		result.Audit.Failures = summarizeDeleteFailures(result.Failures)
		result.Validation = lifecycle.DeleteValidationReport{
			StartedAt:  result.StartedAt,
			FinishedAt: result.FinishedAt,
			Status:     "blocked",
			Verified:   false,
			Issues: []lifecycle.DeleteValidationIssue{{
				Category: lifecycle.DeleteFailureCategoryValidation,
				Stage:    lifecycle.DeleteFailureStageValidation,
				Message:  fmt.Sprintf("delete plan has %d blocking conflict(s)", len(report.Plan.Conflicts)),
			}},
		}
		return result, fmt.Errorf("delete plan has %d blocking conflict(s)", len(report.Plan.Conflicts))
	}
	deleted, skipped, failures, warnings := s.executeDeletePlan(context.Background(), report.Plan)
	result.Deleted = deleted
	result.Skipped = skipped
	result.Failures = failures
	result.Warnings = append(result.Warnings, warnings...)
	validation := s.validateDeleteExecution(context.Background(), report.Plan, deleted, skipped)
	result.Validation = validation
	if len(validation.Issues) > 0 {
		result.Failures = append(result.Failures, validationIssuesToFailures(validation.Issues)...)
	}
	result.FinishedAt = nowUTC()
	result.Summary = summarizeDeleteResult(result.StartedAt, result.FinishedAt, result.Deleted, result.Skipped, result.Failures)
	result.Audit.FinishedAt = result.FinishedAt
	result.Audit.Result = result.Summary
	result.Audit.Failures = summarizeDeleteFailures(result.Failures)
	switch {
	case len(result.Failures) > 0 && len(deleted) > 0:
		result.Status = lifecycle.DeleteStatusPartial
	case len(result.Failures) > 0:
		result.Status = lifecycle.DeleteStatusFailed
	case len(skipped) > 0:
		result.Status = lifecycle.DeleteStatusPartial
	default:
		result.Status = lifecycle.DeleteStatusSucceeded
	}
	return result, nil
}

func (s *Service) resolveModule(moduleName string, req lifecycle.DeleteRequest) (ModuleResolution, error) {
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

func (s *Service) buildPlan(req lifecycle.DeleteRequest, resolution ModuleResolution) (lifecycle.DeletePlan, error) {
	moduleDir, err := s.absoluteModuleDir(resolution.Module)
	if err != nil {
		return lifecycle.DeletePlan{}, err
	}
	displayRoot := s.displayRoot()
	plan := lifecycle.DeletePlan{
		Request:     req,
		Module:      resolution.Module,
		DryRun:      true,
		Force:       req.Force,
		PolicyStore: resolution.PolicyStore,
		Legacy:      req.Compatibility.Normalize(),
	}
	manifestDoc, manifestErr := s.loadManifest(moduleDir)
	if manifestErr != nil && !errors.Is(manifestErr, os.ErrNotExist) {
		return lifecycle.DeletePlan{}, manifestErr
	}
	managedByCodeGen := resolution.GeneratedBootstrap
	if !managedByCodeGen {
		plan.Warnings = append(plan.Warnings, "bootstrap.go is not marked as generated; preview uses conservative inference")
		plan.Conflicts = append(plan.Conflicts, lifecycle.DeleteConflict{
			Kind:     "legacy-module",
			Severity: conflictSeverityHigh,
			Message:  "module bootstrap is not generated; delete requires explicit review",
			Path:     resolution.BootstrapPath,
		})
	}
	if resolution.IsBuiltin {
		plan.Warnings = append(plan.Warnings, "module is registered as a builtin module")
		plan.Conflicts = append(plan.Conflicts, lifecycle.DeleteConflict{
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
		plan.SourceFiles = append(plan.SourceFiles, lifecycle.DeleteItem{
			Module:   resolution.Module,
			Kind:     lifecycle.AssetKindSourceDirectory,
			Path:     displayPath(displayRoot, moduleDir),
			Origin:   lifecycle.AssetOriginGenerated,
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
		plan.Conflicts = append(plan.Conflicts, lifecycle.DeleteConflict{
			Kind:     "manifest-module-mismatch",
			Severity: conflictSeverityHigh,
			Message:  fmt.Sprintf("manifest module %q does not match requested module %q", strings.TrimSpace(manifestToUse.Module), resolution.Module),
			Path:     displayPath(displayRoot, filepath.Join(moduleDir, "manifest.yaml")),
		})
	}
	manifestIndex, indexErr := loadModuleManifestReferenceIndex(s.backendRoot)
	if indexErr != nil {
		plan.Warnings = append(plan.Warnings, fmt.Sprintf("manifest reference index unavailable: %v", indexErr))
	}
	ownership := lifecycle.ModuleOwnership{
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
				plan.Conflicts = append(plan.Conflicts, lifecycle.DeleteConflict{
					Kind:     "invalid-route",
					Severity: conflictSeverityWarning,
					Message:  "route entry is incomplete",
				})
				continue
			}
			asset := lifecycle.DeleteItem{
				Module:  resolution.Module,
				Kind:    lifecycle.AssetKindRuntimeRoute,
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
			if manifestIndex != nil {
				owners := manifestIndex.routeOwnersFor(method, path)
				asset.Metadata["shared_with"] = append([]string(nil), owners...)
				asset.Metadata["reference_count"] = len(owners)
				if len(owners) > 1 {
					asset.Origin = lifecycle.AssetOriginShared
					plan.Warnings = append(plan.Warnings, fmt.Sprintf("route policy %s %s is shared by modules: %s", method, path, strings.Join(owners, ", ")))
				}
			}
			plan.RuntimeAssets = append(plan.RuntimeAssets, asset)
			if resolution.PolicyStore.IsKnown() && req.WithPolicy {
				selector := lifecycle.PolicySelector{
					Store:     resolution.PolicyStore,
					Module:    resolution.Module,
					SourceRef: strings.TrimSpace(method + " " + path),
					PType:     "p",
					V0:        "admin",
					V1:        path,
					V2:        method,
					Metadata: map[string]any{
						"module":  resolution.Module,
						"managed": managedByCodeGen,
						"origin":  string(routeOrigin(managedByCodeGen)),
						"path":    path,
						"method":  method,
					},
				}
				plan.PolicyChanges = append(plan.PolicyChanges, lifecycle.DeleteItem{
					Module:   resolution.Module,
					Kind:     lifecycle.AssetKindPolicyRule,
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
			asset := lifecycle.DeleteItem{
				Module:  resolution.Module,
				Kind:    lifecycle.AssetKindRuntimeMenu,
				Path:    path,
				Ref:     strings.TrimSpace(menu.Permission),
				Origin:  routeOrigin(managedByCodeGen),
				Managed: managedByCodeGen,
				Metadata: map[string]any{
					"name":        strings.TrimSpace(menu.Name),
					"parent_path": strings.TrimSpace(menu.ParentPath),
					"component":   strings.TrimSpace(menu.Component),
				},
			}
			if manifestIndex != nil {
				owners := manifestIndex.menuOwnersFor(path)
				asset.Metadata["shared_with"] = append([]string(nil), owners...)
				asset.Metadata["reference_count"] = len(owners)
				if len(owners) > 1 {
					asset.Origin = lifecycle.AssetOriginShared
					plan.Warnings = append(plan.Warnings, fmt.Sprintf("menu %s is shared by modules: %s", path, strings.Join(owners, ", ")))
				}
			}
			plan.RuntimeAssets = append(plan.RuntimeAssets, asset)
		}
	}
	if req.WithPolicy && resolution.PolicyStore.IsKnown() && len(manifestToUse.Permissions) > 0 {
		for _, permission := range manifestToUse.Permissions {
			object := strings.TrimSpace(permission.Object)
			action := strings.TrimSpace(permission.Action)
			if object == "" && action == "" {
				continue
			}
			selector := lifecycle.PolicySelector{
				Store:     resolution.PolicyStore,
				Module:    resolution.Module,
				SourceRef: strings.TrimSpace(object + " " + action),
				PType:     "p",
				V0:        "admin",
				V1:        object,
				V2:        action,
				Metadata: map[string]any{
					"module":      resolution.Module,
					"managed":     managedByCodeGen,
					"origin":      string(routeOrigin(managedByCodeGen)),
					"object":      object,
					"action":      action,
					"description": strings.TrimSpace(permission.Description),
				},
			}
			asset := lifecycle.DeleteItem{
				Module:   resolution.Module,
				Kind:     lifecycle.AssetKindPolicyRule,
				Store:    resolution.PolicyStore,
				Ref:      strings.TrimSpace(object + ":" + action),
				Selector: &selector,
				Origin:   routeOrigin(managedByCodeGen),
				Managed:  managedByCodeGen,
				Metadata: map[string]any{
					"object":      object,
					"action":      action,
					"description": strings.TrimSpace(permission.Description),
				},
			}
			if manifestIndex != nil {
				owners := manifestIndex.permissionOwnersFor(object, action)
				asset.Metadata["shared_with"] = append([]string(nil), owners...)
				asset.Metadata["reference_count"] = len(owners)
				if len(owners) > 1 {
					asset.Origin = lifecycle.AssetOriginShared
					plan.Warnings = append(plan.Warnings, fmt.Sprintf("permission policy %s:%s is shared by modules: %s", object, action, strings.Join(owners, ", ")))
				}
			}
			plan.PolicyChanges = append(plan.PolicyChanges, asset)
		}
	}
	if req.WithRegistry && resolution.PolicyStore.IsKnown() {
		plan.RegistryChanges = append(plan.RegistryChanges, lifecycle.DeleteItem{
			Module:   resolution.Module,
			Kind:     lifecycle.AssetKindRuntimeRegistry,
			Path:     resolution.RegistryPath,
			Ref:      resolution.Module,
			Store:    resolution.PolicyStore,
			Origin:   lifecycle.AssetOriginGenerated,
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
	plan.Ownership.OwnedFiles = append([]lifecycle.DeleteItem(nil), plan.SourceFiles...)
	plan.Ownership.RuntimeAssets = append([]lifecycle.DeleteItem(nil), plan.RuntimeAssets...)
	plan.Ownership.PolicyAssets = make([]lifecycle.PolicyAsset, 0, len(plan.PolicyChanges))
	for _, item := range plan.PolicyChanges {
		if item.Selector == nil {
			continue
		}
		plan.Ownership.PolicyAssets = append(plan.Ownership.PolicyAssets, lifecycle.PolicyAsset{
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
			Managed:   item.Managed,
			Metadata:  cloneAnyMap(item.Selector.Metadata),
		})
	}
	plan.Ownership.FrontendAssets = append([]lifecycle.DeleteItem(nil), plan.FrontendChanges...)
	plan.Summary = lifecycle.DeletePlanSummary{
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

func (s *Service) collectSourceFiles(moduleDir, displayRoot, module string, managed bool) ([]lifecycle.DeleteConflict, []lifecycle.DeleteItem) {
	var conflicts []lifecycle.DeleteConflict
	var items []lifecycle.DeleteItem
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
			items = append(items, lifecycle.DeleteItem{
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
		conflicts = append(conflicts, lifecycle.DeleteConflict{
			Kind:     "unknown-owned-file",
			Severity: conflictSeverityHigh,
			Message:  "file is not recognized as a generated CodeGen asset",
			Path:     displayPath(displayRoot, path),
		})
		return nil
	}); err != nil {
		conflicts = append(conflicts, lifecycle.DeleteConflict{
			Kind:     "scan-error",
			Severity: conflictSeverityWarning,
			Message:  err.Error(),
			Path:     displayPath(displayRoot, moduleDir),
		})
	}
	_ = known
	return conflicts, items
}

func (s *Service) collectFrontendCandidates(module string, managed bool, displayRoot string) []lifecycle.DeleteItem {
	paths := []struct {
		path string
		kind lifecycle.AssetKind
	}{
		{path: filepath.Join(s.backendRoot, "..", "web", "src", "api", module+".ts"), kind: lifecycle.AssetKindFrontendFile},
		{path: filepath.Join(s.backendRoot, "..", "web", "src", "router", "modules", module+".ts"), kind: lifecycle.AssetKindFrontendFile},
		{path: filepath.Join(s.backendRoot, "..", "web", "src", "views", module, "index.vue"), kind: lifecycle.AssetKindFrontendFile},
	}
	items := make([]lifecycle.DeleteItem, 0, len(paths))
	for _, candidate := range paths {
		abs := filepath.Clean(candidate.path)
		if !fileExists(abs) {
			continue
		}
		items = append(items, lifecycle.DeleteItem{
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

func classifyModuleSourceFile(rel string) (lifecycle.AssetKind, bool) {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	switch rel {
	case "module.go", "bootstrap.go", "manifest.yaml", "manifest.yml", "codegen.manifest.json", "schema.sql":
		return lifecycle.AssetKindSourceFile, true
	case "transport/http/router.go":
		return lifecycle.AssetKindSourceFile, true
	}
	switch {
	case strings.HasPrefix(rel, "application/command/") && strings.HasSuffix(rel, ".go"):
		return lifecycle.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "application/query/") && strings.HasSuffix(rel, ".go"):
		return lifecycle.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "application/service/") && strings.HasSuffix(rel, ".go"):
		return lifecycle.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "domain/model/") && strings.HasSuffix(rel, ".go"):
		return lifecycle.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "domain/repository/") && strings.HasSuffix(rel, ".go"):
		return lifecycle.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "infrastructure/repo/") && strings.HasSuffix(rel, ".go"):
		return lifecycle.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "transport/http/request/") && strings.HasSuffix(rel, ".go"):
		return lifecycle.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "transport/http/response/") && strings.HasSuffix(rel, ".go"):
		return lifecycle.AssetKindSourceFile, true
	case strings.HasPrefix(rel, "transport/http/handler/") && strings.HasSuffix(rel, ".go"):
		return lifecycle.AssetKindSourceFile, true
	case strings.HasSuffix(rel, "_test.go"):
		return lifecycle.AssetKindSourceFile, true
	default:
		return lifecycle.AssetKindUnknown, false
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

func routeOrigin(managed bool) lifecycle.AssetOrigin {
	if managed {
		return lifecycle.AssetOriginGenerated
	}
	return lifecycle.AssetOriginInferred
}

func sourceOrigin(managed bool) lifecycle.AssetOrigin {
	if managed {
		return lifecycle.AssetOriginGenerated
	}
	return lifecycle.AssetOriginInferred
}

func sortDeleteItems(items []lifecycle.DeleteItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Path == items[j].Path {
			return items[i].Kind < items[j].Kind
		}
		return items[i].Path < items[j].Path
	})
}

func (s *Service) resolvePolicyStore(req lifecycle.PolicyStoreKind) lifecycle.PolicyStoreKind {
	if req.IsKnown() {
		return req
	}
	if resolved := normalizePolicyStoreSource(string(req)); resolved.IsKnown() {
		return resolved
	}
	if s != nil && s.policyStore.IsKnown() {
		return s.policyStore
	}
	return lifecycle.PolicyStoreUnknown
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

func cloneAnyMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	clone := make(map[string]any, len(src))
	for key, value := range src {
		clone[key] = value
	}
	return clone
}

func (s *Service) executeDeletePlan(ctx context.Context, plan lifecycle.DeletePlan) ([]lifecycle.DeleteItem, []lifecycle.DeleteItem, []lifecycle.DeleteFailure, []string) {
	if ctx == nil {
		ctx = context.Background()
	}
	deleted := make([]lifecycle.DeleteItem, 0, len(plan.SourceFiles)+len(plan.RegistryChanges)+len(plan.FrontendChanges))
	skipped := make([]lifecycle.DeleteItem, 0, len(plan.RuntimeAssets)+len(plan.PolicyChanges))
	failures := make([]lifecycle.DeleteFailure, 0, 4)
	warnings := make([]string, 0, 4)

	for _, item := range plan.SourceFiles {
		if item.Kind == lifecycle.AssetKindSourceDirectory {
			if err := s.cleanupEmptyParents(s.resolveExecutionPath(item.Path), filepath.Join(s.backendRoot, "modules")); err != nil {
				failures = append(failures, newDeleteFailure(item, lifecycle.DeleteFailureCategoryFile, lifecycle.DeleteFailureStageFile, fmt.Sprintf("cleanup source directory: %v", err), true))
			}
			deleted = append(deleted, item)
			continue
		}
		if err := s.removeExecutionPath(item.Path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				skipped = append(skipped, item)
				warnings = append(warnings, fmt.Sprintf("source file already missing: %s", item.Path))
				continue
			}
			failures = append(failures, newDeleteFailure(item, lifecycle.DeleteFailureCategoryFile, lifecycle.DeleteFailureStageFile, fmt.Sprintf("delete source file: %v", err), true))
			continue
		}
		deleted = append(deleted, item)
		if err := s.cleanupGeneratedAncestors(item.Path); err != nil {
			failures = append(failures, newDeleteFailure(item, lifecycle.DeleteFailureCategoryFile, lifecycle.DeleteFailureStageFile, fmt.Sprintf("cleanup generated directories: %v", err), true))
		}
	}

	if len(plan.RegistryChanges) > 0 {
		if err := s.refreshBootstrapRegistry(); err != nil {
			for _, item := range plan.RegistryChanges {
				failures = append(failures, newDeleteFailure(item, lifecycle.DeleteFailureCategoryRegistry, lifecycle.DeleteFailureStageRegistry, fmt.Sprintf("refresh bootstrap registry: %v", err), true))
			}
		} else {
			deleted = append(deleted, plan.RegistryChanges...)
		}
	}

	for _, item := range plan.FrontendChanges {
		if err := s.removeExecutionPath(item.Path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				skipped = append(skipped, item)
				warnings = append(warnings, fmt.Sprintf("frontend file already missing: %s", item.Path))
				continue
			}
			failures = append(failures, newDeleteFailure(item, lifecycle.DeleteFailureCategoryFile, lifecycle.DeleteFailureStageFile, fmt.Sprintf("delete frontend file: %v", err), true))
			continue
		}
		deleted = append(deleted, item)
		if err := s.cleanupGeneratedAncestors(item.Path); err != nil {
			failures = append(failures, newDeleteFailure(item, lifecycle.DeleteFailureCategoryFile, lifecycle.DeleteFailureStageFile, fmt.Sprintf("cleanup frontend directories: %v", err), true))
		}
	}

	menuItems, runtimeItems := splitRuntimeDeleteItems(plan.RuntimeAssets)
	if len(menuItems) > 0 {
		if s.menuService == nil {
			skipped = append(skipped, menuItems...)
			warnings = append(warnings, fmt.Sprintf("menu cleanup service is not configured (%d item(s) skipped)", len(menuItems)))
		} else {
			deletedMenus, skippedMenus, menuFailures, menuWarnings := s.deleteRuntimeMenus(ctx, menuItems)
			deleted = append(deleted, deletedMenus...)
			skipped = append(skipped, skippedMenus...)
			failures = append(failures, menuFailures...)
			warnings = append(warnings, menuWarnings...)
		}
	}

	policyItems := filterExecutablePolicyItems(plan.PolicyChanges)
	if len(policyItems.executable) > 0 {
		if s.policyCleanup == nil {
			skipped = append(skipped, policyItems.executable...)
			warnings = append(warnings, fmt.Sprintf("policy cleanup service is not configured (%d item(s) skipped)", len(policyItems.executable)))
		} else {
			cleanupReq := BuildPolicyCleanupRequest(lifecycle.DeletePlan{
				Module:        plan.Module,
				PolicyStore:   plan.PolicyStore,
				PolicyChanges: policyItems.executable,
			})
			cleanupResult, err := s.policyCleanup.Delete(ctx, cleanupReq)
			if err != nil {
				failures = append(failures, lifecycle.DeleteFailure{
					Category:    policyFailureCategory(plan.PolicyStore),
					Stage:       lifecycle.DeleteFailureStagePolicy,
					Reason:      fmt.Sprintf("policy cleanup: %v", err),
					Recoverable: true,
				})
			}
			deleted = append(deleted, policyCleanupDeletedItems(cleanupResult.Deleted)...)
			skipped = append(skipped, policyCleanupSkippedItems(cleanupResult.Skipped)...)
			if len(cleanupResult.Failures) > 0 {
				failures = append(failures, cleanupResult.Failures...)
			}
			warnings = append(warnings, cleanupResult.Warnings...)
		}
	}
	if len(policyItems.skipped) > 0 {
		skipped = append(skipped, policyItems.skipped...)
	}
	if len(policyItems.deferred) > 0 {
		skipped = append(skipped, policyItems.deferred...)
		warnings = append(warnings, fmt.Sprintf("policy cleanup is deferred for %d shared item(s)", len(policyItems.deferred)))
	}
	if len(runtimeItems) > 0 {
		logicalDeleted, logicalSkipped, logicalFailures, logicalWarnings := s.cleanupLogicalRuntimeAssets(plan, runtimeItems, deleted, failures)
		deleted = append(deleted, logicalDeleted...)
		skipped = append(skipped, logicalSkipped...)
		failures = append(failures, logicalFailures...)
		warnings = append(warnings, logicalWarnings...)
	}
	return deleted, skipped, failures, warnings
}

type policyExecutionItems struct {
	executable []lifecycle.DeleteItem
	skipped    []lifecycle.DeleteItem
	deferred   []lifecycle.DeleteItem
}

func splitRuntimeDeleteItems(items []lifecycle.DeleteItem) ([]lifecycle.DeleteItem, []lifecycle.DeleteItem) {
	menus := make([]lifecycle.DeleteItem, 0, len(items))
	deferred := make([]lifecycle.DeleteItem, 0)
	for _, item := range items {
		if item.Kind == lifecycle.AssetKindRuntimeMenu {
			menus = append(menus, item)
			continue
		}
		deferred = append(deferred, item)
	}
	sort.SliceStable(menus, func(i, j int) bool {
		iDepth := strings.Count(normalizeRuntimePath(menus[i].Path), "/")
		jDepth := strings.Count(normalizeRuntimePath(menus[j].Path), "/")
		if iDepth == jDepth {
			return menus[i].Path > menus[j].Path
		}
		return iDepth > jDepth
	})
	return menus, deferred
}

func (s *Service) cleanupLogicalRuntimeAssets(plan lifecycle.DeletePlan, items []lifecycle.DeleteItem, deleted []lifecycle.DeleteItem, failures []lifecycle.DeleteFailure) ([]lifecycle.DeleteItem, []lifecycle.DeleteItem, []lifecycle.DeleteFailure, []string) {
	logicalDeleted := make([]lifecycle.DeleteItem, 0, len(items))
	logicalSkipped := make([]lifecycle.DeleteItem, 0)
	logicalFailures := make([]lifecycle.DeleteFailure, 0)
	warnings := make([]string, 0)

	sourceReady := len(plan.SourceFiles) > 0 && !hasDeleteFailureKind(failures, lifecycle.AssetKindSourceFile, lifecycle.AssetKindSourceDirectory)
	registryReady := len(plan.RegistryChanges) == 0 || (hasDeletedKind(deleted, lifecycle.AssetKindRuntimeRegistry) && !hasDeleteFailureKind(failures, lifecycle.AssetKindRuntimeRegistry))
	frontendReady := len(plan.FrontendChanges) > 0 && !hasDeleteFailureKind(failures, lifecycle.AssetKindFrontendFile)
	policyReady := s != nil && s.policyCleanup != nil && len(plan.PolicyChanges) > 0 && hasDeletedKind(deleted, lifecycle.AssetKindPolicyRule) && !hasDeleteFailureKind(failures, lifecycle.AssetKindPolicyRule)

	for _, item := range items {
		if metadataInt(item.Metadata, "reference_count") > 1 || item.Origin == lifecycle.AssetOriginShared {
			logicalSkipped = append(logicalSkipped, item)
			warnings = append(warnings, fmt.Sprintf("shared runtime asset %s is referenced by multiple modules; skipped", item.Path))
			continue
		}
		switch item.Kind {
		case lifecycle.AssetKindRuntimeRoute:
			if sourceReady && registryReady {
				logicalDeleted = append(logicalDeleted, item)
				continue
			}
			logicalSkipped = append(logicalSkipped, item)
			warnings = append(warnings, fmt.Sprintf("route cleanup deferred until generated source files and registry refresh complete: %s", item.Path))
		case lifecycle.AssetKindRuntimePermission:
			if policyReady {
				logicalDeleted = append(logicalDeleted, item)
				continue
			}
			logicalSkipped = append(logicalSkipped, item)
			warnings = append(warnings, fmt.Sprintf("permission cleanup deferred until policy cleanup completes: %s", item.Path))
		case lifecycle.AssetKindRuntimePage:
			if frontendReady {
				logicalDeleted = append(logicalDeleted, item)
				continue
			}
			logicalSkipped = append(logicalSkipped, item)
			warnings = append(warnings, fmt.Sprintf("page cleanup deferred until frontend cleanup completes: %s", item.Path))
		default:
			logicalSkipped = append(logicalSkipped, item)
			warnings = append(warnings, fmt.Sprintf("unsupported runtime cleanup asset kind %s for %s", item.Kind, item.Path))
		}
	}

	return logicalDeleted, logicalSkipped, logicalFailures, warnings
}

func (s *Service) validateDeleteExecution(ctx context.Context, plan lifecycle.DeletePlan, deleted, skipped []lifecycle.DeleteItem) lifecycle.DeleteValidationReport {
	startedAt := nowUTC()
	report := lifecycle.DeleteValidationReport{
		StartedAt: startedAt,
		Status:    "skipped",
		Verified:  true,
	}
	issues := make([]lifecycle.DeleteValidationIssue, 0, len(deleted))
	checked := 0
	registryChecks := make(map[string]struct{})
	menuItems := make([]lifecycle.DeleteItem, 0)
	policyItems := make([]lifecycle.DeleteItem, 0)
	for _, item := range deleted {
		checked++
		switch item.Kind {
		case lifecycle.AssetKindSourceFile, lifecycle.AssetKindSourceDirectory, lifecycle.AssetKindFrontendFile:
			absolute := s.resolveExecutionPath(item.Path)
			if absolute == "" {
				issues = append(issues, lifecycle.DeleteValidationIssue{
					Item:     item,
					Category: lifecycle.DeleteFailureCategoryFile,
					Stage:    lifecycle.DeleteFailureStageValidation,
					Message:  "deleted file path could not be resolved",
				})
				continue
			}
			if _, err := os.Stat(absolute); err == nil {
				issues = append(issues, lifecycle.DeleteValidationIssue{
					Item:     item,
					Category: lifecycle.DeleteFailureCategoryFile,
					Stage:    lifecycle.DeleteFailureStageValidation,
					Message:  "file still exists after deletion",
					Actual:   absolute,
				})
			} else if !errors.Is(err, os.ErrNotExist) {
				issues = append(issues, lifecycle.DeleteValidationIssue{
					Item:     item,
					Category: lifecycle.DeleteFailureCategoryFile,
					Stage:    lifecycle.DeleteFailureStageValidation,
					Message:  fmt.Sprintf("validate file removal: %v", err),
					Actual:   absolute,
				})
			}
		case lifecycle.AssetKindRuntimeRegistry:
			if _, ok := registryChecks[item.Path]; ok {
				continue
			}
			registryChecks[item.Path] = struct{}{}
			absolute := s.resolveExecutionPath(item.Path)
			if absolute == "" {
				issues = append(issues, lifecycle.DeleteValidationIssue{
					Item:     item,
					Category: lifecycle.DeleteFailureCategoryRegistry,
					Stage:    lifecycle.DeleteFailureStageValidation,
					Message:  "registry path could not be resolved",
				})
				continue
			}
			content, err := os.ReadFile(absolute)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				issues = append(issues, lifecycle.DeleteValidationIssue{
					Item:     item,
					Category: lifecycle.DeleteFailureCategoryRegistry,
					Stage:    lifecycle.DeleteFailureStageValidation,
					Message:  fmt.Sprintf("read registry file: %v", err),
					Actual:   absolute,
				})
				continue
			}
			if strings.Contains(string(content), "goadmin/modules/"+plan.Module) {
				issues = append(issues, lifecycle.DeleteValidationIssue{
					Item:     item,
					Category: lifecycle.DeleteFailureCategoryRegistry,
					Stage:    lifecycle.DeleteFailureStageValidation,
					Message:  "bootstrap registry still contains deleted module",
					Expected: "module removed from generated registry",
					Actual:   absolute,
				})
			}
		case lifecycle.AssetKindRuntimeMenu:
			menuItems = append(menuItems, item)
		case lifecycle.AssetKindPolicyRule:
			if item.Selector != nil {
				policyItems = append(policyItems, item)
			}
		case lifecycle.AssetKindRuntimeRoute, lifecycle.AssetKindRuntimePermission, lifecycle.AssetKindRuntimePage:
			continue
		default:
			continue
		}
	}
	if len(menuItems) > 0 {
		if s.menuService == nil {
			for _, item := range menuItems {
				issues = append(issues, lifecycle.DeleteValidationIssue{
					Item:     item,
					Category: lifecycle.DeleteFailureCategoryDatabase,
					Stage:    lifecycle.DeleteFailureStageValidation,
					Message:  "menu service is not configured for validation",
				})
			}
		} else if tree, err := s.menuService.Tree(ctx); err != nil {
			for _, item := range menuItems {
				issues = append(issues, lifecycle.DeleteValidationIssue{
					Item:     item,
					Category: lifecycle.DeleteFailureCategoryDatabase,
					Stage:    lifecycle.DeleteFailureStageValidation,
					Message:  fmt.Sprintf("load menu tree: %v", err),
				})
			}
		} else {
			menusByPath := flattenMenusByPath(tree)
			for _, item := range menuItems {
				if _, ok := menusByPath[normalizeRuntimePath(item.Path)]; ok {
					issues = append(issues, lifecycle.DeleteValidationIssue{
						Item:     item,
						Category: lifecycle.DeleteFailureCategoryDatabase,
						Stage:    lifecycle.DeleteFailureStageValidation,
						Message:  "menu still exists after deletion",
						Expected: "menu path removed",
						Actual:   item.Path,
					})
				}
			}
		}
	}
	if len(policyItems) > 0 && s.policyCleanup != nil {
		selectors := make([]lifecycle.PolicySelector, 0, len(policyItems))
		for _, item := range policyItems {
			if item.Selector == nil {
				continue
			}
			selectors = append(selectors, *item.Selector)
		}
		if len(selectors) > 0 {
			preview, err := s.policyCleanup.Preview(ctx, PolicyCleanupRequest{Module: plan.Module, Store: plan.PolicyStore, Selectors: selectors})
			if err != nil {
				for _, item := range policyItems {
					issues = append(issues, lifecycle.DeleteValidationIssue{
						Item:     item,
						Category: policyFailureCategory(plan.PolicyStore),
						Stage:    lifecycle.DeleteFailureStageValidation,
						Message:  fmt.Sprintf("validate policy cleanup: %v", err),
					})
				}
			} else {
				for _, item := range preview.Items {
					if item.Decision != "delete" {
						continue
					}
					issue := lifecycle.DeleteValidationIssue{
						Item:     policyCleanupItemToDeleteItem(item),
						Category: policyFailureCategory(plan.PolicyStore),
						Stage:    lifecycle.DeleteFailureStageValidation,
						Message:  "policy rule still exists after deletion",
						Expected: "policy rule removed",
						Actual:   item.Rule.SourceRef,
					}
					if item.Reason != "" {
						issue.Metadata = map[string]any{"reason": item.Reason}
					}
					issues = append(issues, issue)
				}
			}
		}
	}
	report.Checked = checked
	report.FinishedAt = nowUTC()
	report.Issues = issues
	if len(issues) == 0 {
		report.Status = "passed"
		report.Verified = true
		if len(deleted) == 0 {
			report.Status = "skipped"
		}
		return report
	}
	report.Status = "failed"
	report.Verified = false
	return report
}

func validationIssuesToFailures(issues []lifecycle.DeleteValidationIssue) []lifecycle.DeleteFailure {
	if len(issues) == 0 {
		return nil
	}
	failures := make([]lifecycle.DeleteFailure, 0, len(issues))
	for _, issue := range issues {
		failures = append(failures, lifecycle.DeleteFailure{
			Item:        issue.Item,
			Category:    issue.Category,
			Stage:       issue.Stage,
			Reason:      issue.Message,
			Recoverable: false,
			Metadata:    cloneAnyMap(issue.Metadata),
		})
	}
	return failures
}

func summarizeDeleteFailures(failures []lifecycle.DeleteFailure) lifecycle.DeleteFailureSummary {
	summary := lifecycle.DeleteFailureSummary{Total: len(failures)}
	for _, failure := range failures {
		switch failure.Category {
		case lifecycle.DeleteFailureCategoryFile:
			summary.File++
		case lifecycle.DeleteFailureCategoryPolicy:
			summary.Policy++
		case lifecycle.DeleteFailureCategoryDatabase:
			summary.Database++
		case lifecycle.DeleteFailureCategoryRegistry:
			summary.Registry++
		case lifecycle.DeleteFailureCategoryValidation:
			summary.Validation++
		}
		if failure.Recoverable {
			summary.Recoverable++
		} else {
			summary.Blocking++
		}
	}
	return summary
}

func newDeleteFailure(item lifecycle.DeleteItem, category lifecycle.DeleteFailureCategory, stage lifecycle.DeleteFailureStage, reason string, recoverable bool) lifecycle.DeleteFailure {
	return lifecycle.DeleteFailure{
		Item:        item,
		Category:    category,
		Stage:       stage,
		Reason:      reason,
		Recoverable: recoverable,
	}
}

func policyFailureCategory(store lifecycle.PolicyStoreKind) lifecycle.DeleteFailureCategory {
	switch store {
	case lifecycle.PolicyStoreDB:
		return lifecycle.DeleteFailureCategoryDatabase
	case lifecycle.PolicyStoreCSV:
		return lifecycle.DeleteFailureCategoryPolicy
	default:
		return lifecycle.DeleteFailureCategoryPolicy
	}
}

func hasDeletedKind(items []lifecycle.DeleteItem, kinds ...lifecycle.AssetKind) bool {
	if len(items) == 0 || len(kinds) == 0 {
		return false
	}
	wanted := make(map[lifecycle.AssetKind]struct{}, len(kinds))
	for _, kind := range kinds {
		w := kind
		wanted[w] = struct{}{}
	}
	for _, item := range items {
		if _, ok := wanted[item.Kind]; ok {
			return true
		}
	}
	return false
}

func hasDeleteFailureKind(failures []lifecycle.DeleteFailure, kinds ...lifecycle.AssetKind) bool {
	if len(failures) == 0 || len(kinds) == 0 {
		return false
	}
	wanted := make(map[lifecycle.AssetKind]struct{}, len(kinds))
	for _, kind := range kinds {
		w := kind
		wanted[w] = struct{}{}
	}
	for _, failure := range failures {
		if _, ok := wanted[failure.Item.Kind]; ok {
			return true
		}
	}
	return false
}

func filterExecutablePolicyItems(items []lifecycle.DeleteItem) policyExecutionItems {
	result := policyExecutionItems{executable: make([]lifecycle.DeleteItem, 0, len(items))}
	for _, item := range items {
		count := metadataInt(item.Metadata, "reference_count")
		if count > 1 || item.Origin == lifecycle.AssetOriginShared {
			result.deferred = append(result.deferred, item)
			continue
		}
		if item.Selector == nil {
			result.skipped = append(result.skipped, item)
			continue
		}
		result.executable = append(result.executable, item)
	}
	return result
}

func (s *Service) deleteRuntimeMenus(ctx context.Context, items []lifecycle.DeleteItem) ([]lifecycle.DeleteItem, []lifecycle.DeleteItem, []lifecycle.DeleteFailure, []string) {
	deleted := make([]lifecycle.DeleteItem, 0, len(items))
	skipped := make([]lifecycle.DeleteItem, 0)
	failures := make([]lifecycle.DeleteFailure, 0)
	warnings := make([]string, 0)
	tree, err := s.menuService.Tree(ctx)
	if err != nil {
		for _, item := range items {
			failures = append(failures, newDeleteFailure(item, lifecycle.DeleteFailureCategoryDatabase, lifecycle.DeleteFailureStageDatabase, fmt.Sprintf("load menu tree: %v", err), true))
		}
		return deleted, skipped, failures, warnings
	}
	menusByPath := flattenMenusByPath(tree)
	for _, item := range items {
		if metadataInt(item.Metadata, "reference_count") > 1 || item.Origin == lifecycle.AssetOriginShared {
			skipped = append(skipped, item)
			warnings = append(warnings, fmt.Sprintf("shared menu %s is referenced by multiple modules; skipped", item.Path))
			continue
		}
		menu := menusByPath[normalizeRuntimePath(item.Path)]
		if menu == nil {
			skipped = append(skipped, item)
			warnings = append(warnings, fmt.Sprintf("menu already missing: %s", item.Path))
			continue
		}
		if err := s.menuService.Delete(ctx, menu.ID); err != nil {
			failures = append(failures, newDeleteFailure(item, lifecycle.DeleteFailureCategoryDatabase, lifecycle.DeleteFailureStageDatabase, fmt.Sprintf("delete menu %s: %v", item.Path, err), true))
			continue
		}
		deleted = append(deleted, item)
		delete(menusByPath, normalizeRuntimePath(item.Path))
	}
	return deleted, skipped, failures, warnings
}

func (s *Service) validateMenuCleanup(ctx context.Context, deleted []lifecycle.DeleteItem) error {
	if s == nil || s.menuService == nil || len(deleted) == 0 {
		return nil
	}
	tree, err := s.menuService.Tree(ctx)
	if err != nil {
		return err
	}
	menusByPath := flattenMenusByPath(tree)
	for _, item := range deleted {
		if item.Kind != lifecycle.AssetKindRuntimeMenu {
			continue
		}
		if _, ok := menusByPath[normalizeRuntimePath(item.Path)]; ok {
			return fmt.Errorf("menu %s still exists after deletion", item.Path)
		}
	}
	return nil
}

func flattenMenusByPath(items []menuModel.Menu) map[string]*menuModel.Menu {
	result := make(map[string]*menuModel.Menu)
	var walk func([]menuModel.Menu)
	walk = func(list []menuModel.Menu) {
		for _, item := range list {
			clone := item.Clone()
			result[normalizeRuntimePath(clone.Path)] = &clone
			if len(clone.Children) > 0 {
				walk(clone.Children)
			}
		}
	}
	walk(items)
	return result
}

func policyCleanupDeletedItems(items []PolicyCleanupItem) []lifecycle.DeleteItem {
	result := make([]lifecycle.DeleteItem, 0, len(items))
	for _, item := range items {
		if item.Decision != "delete" {
			continue
		}
		result = append(result, policyCleanupItemToDeleteItem(item))
	}
	return result
}

func policyCleanupSkippedItems(items []PolicyCleanupItem) []lifecycle.DeleteItem {
	result := make([]lifecycle.DeleteItem, 0, len(items))
	for _, item := range items {
		if item.Decision == "delete" {
			continue
		}
		result = append(result, policyCleanupItemToDeleteItem(item))
	}
	return result
}

func policyCleanupItemToDeleteItem(item PolicyCleanupItem) lifecycle.DeleteItem {
	converted := lifecycle.DeleteItem{
		Module:   item.Selector.Module,
		Kind:     lifecycle.AssetKindPolicyRule,
		Store:    item.Selector.Store,
		Ref:      strings.TrimSpace(item.Rule.SourceRef),
		Selector: clonePolicySelector(item.Selector),
		Origin:   lifecycle.AssetOriginGenerated,
		Managed:  selectorManaged(item.Selector),
		Metadata: cloneAnyMap(item.Selector.Metadata),
	}
	if converted.Metadata == nil {
		converted.Metadata = map[string]any{}
	}
	converted.Metadata["reason"] = item.Reason
	converted.Metadata["decision"] = item.Decision
	converted.Metadata["match_count"] = item.MatchCount
	return converted
}

func clonePolicySelector(selector lifecycle.PolicySelector) *lifecycle.PolicySelector {
	copySelector := selector
	copySelector.Metadata = cloneAnyMap(selector.Metadata)
	return &copySelector
}

func metadataInt(metadata map[string]any, key string) int {
	if len(metadata) == 0 {
		return 0
	}
	value, ok := metadata[key]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
			return n
		}
	}
	return 0
}

func summarizeDeleteResult(startedAt, finishedAt time.Time, deleted, skipped []lifecycle.DeleteItem, failures []lifecycle.DeleteFailure) lifecycle.DeleteResultSummary {
	summary := lifecycle.DeleteResultSummary{
		Skipped: len(skipped),
		Failed:  len(failures),
	}
	for _, item := range deleted {
		switch item.Kind {
		case lifecycle.AssetKindSourceFile, lifecycle.AssetKindSourceDirectory:
			summary.DeletedSourceFiles++
		case lifecycle.AssetKindRuntimeRegistry:
			summary.DeletedRegistryChanges++
		case lifecycle.AssetKindPolicyRule:
			summary.DeletedPolicyChanges++
		case lifecycle.AssetKindFrontendFile:
			summary.DeletedFrontendChanges++
		case lifecycle.AssetKindRuntimeRoute, lifecycle.AssetKindRuntimeMenu, lifecycle.AssetKindRuntimePermission, lifecycle.AssetKindRuntimePage:
			summary.DeletedRuntimeAssets++
		}
	}
	summary.TotalDeleted = summary.DeletedSourceFiles + summary.DeletedRuntimeAssets + summary.DeletedRegistryChanges + summary.DeletedPolicyChanges + summary.DeletedFrontendChanges
	if !startedAt.IsZero() && !finishedAt.IsZero() && finishedAt.After(startedAt) {
		summary.ElapsedMillis = finishedAt.Sub(startedAt).Milliseconds()
	}
	return summary
}

func (s *Service) removeExecutionPath(displayPathValue string) error {
	absolute := s.resolveExecutionPath(displayPathValue)
	if absolute == "" {
		return os.ErrNotExist
	}
	if err := os.Remove(absolute); err != nil {
		return err
	}
	return nil
}

func (s *Service) cleanupGeneratedAncestors(displayPathValue string) error {
	absolute := s.resolveExecutionPath(displayPathValue)
	if absolute == "" {
		return nil
	}
	stopAt := filepath.Join(s.backendRoot, "modules")
	if strings.HasPrefix(filepath.ToSlash(absolute), filepath.ToSlash(filepath.Join(s.displayRoot(), "web", "src", "views"))) {
		stopAt = filepath.Join(s.displayRoot(), "web", "src", "views")
	} else if strings.HasPrefix(filepath.ToSlash(absolute), filepath.ToSlash(filepath.Join(s.displayRoot(), "web", "src", "router", "modules"))) {
		stopAt = filepath.Join(s.displayRoot(), "web", "src", "router", "modules")
	} else if strings.HasPrefix(filepath.ToSlash(absolute), filepath.ToSlash(filepath.Join(s.displayRoot(), "web", "src", "api"))) {
		stopAt = filepath.Join(s.displayRoot(), "web", "src", "api")
	}
	return s.cleanupEmptyParents(filepath.Dir(absolute), stopAt)
}

func (s *Service) cleanupEmptyParents(start, stopAt string) error {
	current := filepath.Clean(start)
	stopAt = filepath.Clean(stopAt)
	for {
		if current == "" || current == "." {
			return nil
		}
		if current == stopAt {
			return nil
		}
		entries, err := os.ReadDir(current)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if len(entries) > 0 {
			return nil
		}
		if err := os.Remove(current); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return nil
		}
		current = parent
	}
}

func (s *Service) resolveExecutionPath(displayPathValue string) string {
	target := filepath.Clean(strings.TrimSpace(displayPathValue))
	if target == "" {
		return ""
	}
	if filepath.IsAbs(target) {
		return target
	}
	root := s.displayRoot()
	if strings.TrimSpace(root) == "" {
		return target
	}
	return filepath.Join(root, target)
}

func (s *Service) refreshBootstrapRegistry() error {
	if s == nil {
		return errors.New("delete service is required")
	}
	modulesDir := filepath.Join(s.backendRoot, "modules")
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
	registryPath := filepath.Join(s.backendRoot, "core", "bootstrap", "modules_gen.go")
	if err := os.MkdirAll(filepath.Dir(registryPath), 0o755); err != nil {
		return fmt.Errorf("create registry directory: %w", err)
	}
	if err := os.WriteFile(registryPath, formatted, 0o644); err != nil {
		return fmt.Errorf("write registry file: %w", err)
	}
	return nil
}

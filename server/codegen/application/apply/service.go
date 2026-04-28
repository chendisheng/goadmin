package apply

import (
	"context"
	"errors"
	"fmt"
	installapp "goadmin/codegen/application/install"
	lifecycle "goadmin/codegen/model/lifecycle"
	"os"
	"path/filepath"
	"strings"
	"time"

	legacygenerate "goadmin/cli/generate"
)

type Service struct {
	projectRoot string
	generator   Generator
	installer   Installer
	refresher   CacheRefresher
	logger      Logger
}

func NewService(deps Dependencies) *Service {
	return &Service{
		projectRoot: strings.TrimSpace(deps.ProjectRoot),
		generator:   deps.Generator,
		installer:   deps.Installer,
		refresher:   deps.Refresher,
		logger:      deps.Logger,
	}
}

func (s *Service) Preview(workload Workload) (lifecycle.ApplyPlan, error) {
	if s == nil {
		return lifecycle.ApplyPlan{}, errors.New("apply service is required")
	}
	normalized := normalizeWorkload(workload)
	if err := normalized.Request.Validate(); err != nil {
		return lifecycle.ApplyPlan{}, err
	}
	plan := lifecycle.ApplyPlan{
		Request:     normalized.Request,
		Module:      normalized.Request.Module,
		Kind:        normalized.Request.Kind,
		DryRun:      normalized.Request.DryRun,
		Force:       normalized.Request.Force,
		PolicyStore: normalized.Request.PolicyStore,
		Warnings:    append([]string(nil), normalized.Request.Execution.Notes...),
		Pairing:     lifecycle.DefaultApplyDeletePairing(),
	}
	plan.Generation = buildApplyItems(normalized, lifecycle.ApplyStageGenerate, "generation")
	plan.Installation = buildApplyItems(normalized, lifecycle.ApplyStageInstall, "installation")
	plan.PolicySync = buildPolicySyncItems(normalized)
	if normalized.Request.Verify || normalized.Request.Execution.RequireValidation {
		plan.Validation = []lifecycle.ApplyValidationIssue{{Stage: lifecycle.ApplyStageValidate, Message: "validation will run after execution"}}
	}
	plan.Conflicts = nil
	plan.Summary = summarizeApplyPlan(plan)
	return plan, nil
}

func (s *Service) Execute(ctx context.Context, workload Workload) (lifecycle.ApplyResult, error) {
	if s == nil {
		return lifecycle.ApplyResult{}, errors.New("apply service is required")
	}
	normalized := normalizeWorkload(workload)
	plan, err := s.Preview(normalized)
	if err != nil {
		return lifecycle.ApplyResult{}, err
	}
	result := lifecycle.ApplyResult{
		Request:   normalized.Request,
		Plan:      plan,
		Status:    lifecycle.ApplyStatusPlanned,
		StartedAt: time.Now().UTC(),
		Pairing:   lifecycle.DefaultApplyDeletePairing(),
	}
	if normalized.Request.DryRun {
		result.Status = lifecycle.ApplyStatusDryRun
		result.FinishedAt = time.Now().UTC()
		result.Generation = lifecycle.ApplyGenerationReport{StartedAt: result.StartedAt, FinishedAt: result.FinishedAt, Status: "skipped", Skipped: append([]lifecycle.ApplyItem(nil), plan.Generation...)}
		result.PolicySync = lifecycle.ApplyPolicySyncReport{StartedAt: result.StartedAt, FinishedAt: result.FinishedAt, Status: "skipped", Store: normalized.Request.PolicyStore, Skipped: append([]lifecycle.ApplyItem(nil), plan.PolicySync...)}
		result.Installation = lifecycle.ApplyInstallationReport{StartedAt: result.StartedAt, FinishedAt: result.FinishedAt, Status: "skipped", Skipped: append([]lifecycle.ApplyItem(nil), plan.Installation...)}
		result.Validation = lifecycle.ApplyValidationReport{StartedAt: result.StartedAt, FinishedAt: result.FinishedAt, Status: "skipped", Verified: true, Checked: 0}
		result.Summary = summarizeApplyResult(result)
		return result, nil
	}

	genStarted := time.Now().UTC()
	if err := s.runGeneration(normalized); err != nil {
		result.FinishedAt = time.Now().UTC()
		result.Status = lifecycle.ApplyStatusFailed
		result.Failures = append(result.Failures, lifecycle.ApplyFailure{Category: lifecycle.ApplyFailureCategoryGeneration, Stage: lifecycle.ApplyFailureStageGenerate, Reason: err.Error(), Recoverable: false})
		result.Summary = summarizeApplyResult(result)
		return result, classifyExecutionError(lifecycle.ApplyFailureStageGenerate, lifecycle.ApplyFailureCategoryGeneration, err)
	}
	result.Generation = lifecycle.ApplyGenerationReport{StartedAt: genStarted, FinishedAt: time.Now().UTC(), Status: "succeeded", Created: applyItemActions(plan.Generation, "generated")}

	policyStarted := time.Now().UTC()
	policyExecuted, err := s.runPolicySync(normalized)
	if err != nil {
		result.FinishedAt = time.Now().UTC()
		result.Status = lifecycle.ApplyStatusFailed
		result.Failures = append(result.Failures, lifecycle.ApplyFailure{Category: lifecycle.ApplyFailureCategoryPolicy, Stage: lifecycle.ApplyFailureStagePolicySync, Reason: err.Error(), Recoverable: false})
		result.Summary = summarizeApplyResult(result)
		return result, classifyExecutionError(lifecycle.ApplyFailureStagePolicySync, lifecycle.ApplyFailureCategoryPolicy, err)
	}
	if policyExecuted {
		result.PolicySync = lifecycle.ApplyPolicySyncReport{StartedAt: policyStarted, FinishedAt: time.Now().UTC(), Status: "succeeded", Store: normalized.Request.PolicyStore, Synced: applyItemActions(plan.PolicySync, "synced")}
		if len(result.PolicySync.Synced) == 0 && normalized.Request.Kind == "plugin" {
			result.PolicySync.Synced = []lifecycle.ApplyItem{{Module: normalized.Request.Module, Kind: "policy-rule", Path: "/plugins/" + legacygenerate.ToSnake(normalized.Request.Module) + "/ping", Ref: "GET", Stage: lifecycle.ApplyStagePolicySync, Action: "synced", Managed: true}}
		}
	} else {
		result.PolicySync = lifecycle.ApplyPolicySyncReport{StartedAt: policyStarted, FinishedAt: time.Now().UTC(), Status: "skipped", Store: normalized.Request.PolicyStore, Skipped: append([]lifecycle.ApplyItem(nil), plan.PolicySync...)}
	}

	var installResult installapp.InstallResult
	if normalized.Request.Install {
		installStarted := time.Now().UTC()
		installed, installErr := s.runInstall(ctx, normalized)
		if installErr != nil {
			result.FinishedAt = time.Now().UTC()
			result.Status = lifecycle.ApplyStatusFailed
			result.Failures = append(result.Failures, lifecycle.ApplyFailure{Category: lifecycle.ApplyFailureCategoryInstall, Stage: lifecycle.ApplyFailureStageInstall, Reason: installErr.Error(), Recoverable: false})
			result.Summary = summarizeApplyResult(result)
			return result, classifyExecutionError(lifecycle.ApplyFailureStageInstall, lifecycle.ApplyFailureCategoryInstall, installErr)
		}
		installResult = installed
		if len(installResult.Messages) > 0 {
			result.Warnings = append(result.Warnings, installResult.Messages...)
		}
		result.Installation = lifecycle.ApplyInstallationReport{
			StartedAt:  installStarted,
			FinishedAt: time.Now().UTC(),
			Status:     "succeeded",
			Created:    buildInstalledItems(normalized, installResult),
			Warnings:   append([]string(nil), installResult.Messages...),
		}
		if len(result.Installation.Created) == 0 && installResult.ManifestPath != "" {
			result.Installation.Created = []lifecycle.ApplyItem{{
				Module:  normalized.Request.Module,
				Kind:    "manifest-install",
				Path:    installResult.ManifestPath,
				Stage:   lifecycle.ApplyStageInstall,
				Action:  "installed",
				Managed: true,
			}}
		}
	} else {
		result.Installation = lifecycle.ApplyInstallationReport{StartedAt: result.StartedAt, FinishedAt: time.Now().UTC(), Status: "skipped", Skipped: append([]lifecycle.ApplyItem(nil), plan.Installation...)}
	}

	if normalized.Request.Verify || normalized.Request.Execution.RequireValidation {
		result.Validation = s.buildValidationReport(normalized, result, installResult)
		if len(result.Validation.Issues) > 0 {
			for _, issue := range result.Validation.Issues {
				result.Failures = append(result.Failures, lifecycle.ApplyFailure{
					Item:        issue.Item,
					Category:    lifecycle.ApplyFailureCategoryValidation,
					Stage:       lifecycle.ApplyFailureStage(issue.Stage),
					Reason:      issue.Message,
					Recoverable: normalized.Request.Execution.AllowPartial,
				})
			}
			if !normalized.Request.Execution.AllowPartial {
				result.Status = lifecycle.ApplyStatusFailed
				result.FinishedAt = time.Now().UTC()
				result.Summary = summarizeApplyResult(result)
				return result, classifyExecutionError(lifecycle.ApplyFailureStageValidate, lifecycle.ApplyFailureCategoryValidation, fmt.Errorf("validation failed with %d issue(s)", len(result.Validation.Issues)))
			}
			result.Status = lifecycle.ApplyStatusPartial
		} else {
			result.Validation.Verified = true
			result.Validation.Status = "passed"
		}
	} else {
		result.Validation = lifecycle.ApplyValidationReport{StartedAt: result.StartedAt, FinishedAt: time.Now().UTC(), Status: "skipped", Verified: true, Checked: 0}
	}
	if s.refresher != nil && (normalized.Request.Refresh || normalized.Request.Install) {
		if err := s.refresher.Reload(); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("refresh cache failed: %v", err))
		}
	}
	if result.Status != lifecycle.ApplyStatusPartial {
		result.Status = lifecycle.ApplyStatusSucceeded
	}
	result.FinishedAt = time.Now().UTC()
	result.Summary = summarizeApplyResult(result)
	return result, nil
}

func (s *Service) runGeneration(workload Workload) error {
	if s.generator == nil {
		return errors.New("generator is not configured")
	}
	needsManifest := strings.TrimSpace(workload.Manifest.Name) != "" || strings.TrimSpace(workload.Manifest.Module) != "" || strings.TrimSpace(workload.Manifest.Kind) != ""
	switch normalizeKind(workload.Request.Kind) {
	case "module":
		if err := s.generator.GenerateModule(workload.Module); err != nil {
			return err
		}
	case "crud":
		if err := s.generator.GenerateCRUD(workload.CRUD); err != nil {
			return err
		}
	case "plugin":
		if err := s.generator.GeneratePlugin(workload.Plugin); err != nil {
			return err
		}
	default:
		if err := s.generator.GenerateModule(workload.Module); err != nil {
			return err
		}
	}
	if normalizeKind(workload.Request.Kind) != "crud" && needsManifest {
		if err := s.generator.GenerateManifest(workload.Manifest); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) runPolicySync(workload Workload) (bool, error) {
	switch normalizeKind(workload.Request.Kind) {
	case "crud":
		if workload.CRUD.GeneratePolicy {
			return true, nil
		}
		if len(workload.PolicyLines) == 0 {
			return false, nil
		}
		if s.generator == nil {
			return false, nil
		}
		return true, s.generator.AppendPolicyLines(workload.PolicyLines)
	case "plugin":
		return true, nil
	default:
		if s.generator == nil || len(workload.PolicyLines) == 0 {
			return false, nil
		}
		return true, s.generator.AppendPolicyLines(workload.PolicyLines)
	}
}

func (s *Service) runInstall(ctx context.Context, workload Workload) (installResult installapp.InstallResult, err error) {
	if !workload.Request.Install {
		return installResult, nil
	}
	manifestPath := strings.TrimSpace(workload.ManifestPath)
	if manifestPath == "" {
		manifestPath = filepath.Join(s.projectRoot, "server", "modules", workload.Request.Module, "manifest.yaml")
	} else if !filepath.IsAbs(manifestPath) && s.projectRoot != "" {
		manifestPath = filepath.Join(s.projectRoot, manifestPath)
	}
	if s.installer == nil {
		return installResult, errors.New("installer is not configured")
	}
	result, err := s.installer.InstallManifest(ctx, manifestPath)
	if err != nil {
		return installResult, err
	}
	return result, nil
}

func normalizeWorkload(workload Workload) Workload {
	workload.Request = workload.Request.Normalize()
	workload.Request.Module = normalizeName(workload.Request.Module)
	workload.Request.Kind = normalizeKind(workload.Request.Kind)
	if workload.Request.PolicyStore == lifecycle.PolicyStoreUnknown {
		workload.Request.PolicyStore = lifecycle.PolicyStoreCSV
	}
	if workload.Module.Name == "" {
		workload.Module.Name = workload.Request.Module
	}
	workload.Module.Force = workload.Request.Force
	if workload.CRUD.Name == "" {
		workload.CRUD.Name = workload.Request.Module
	}
	workload.CRUD.Force = workload.Request.Force
	workload.CRUD.GenerateFrontend = workload.Request.GenerateFrontend
	workload.CRUD.GeneratePolicy = workload.Request.GeneratePolicy
	if workload.Plugin.Name == "" {
		workload.Plugin.Name = workload.Request.Module
	}
	workload.Plugin.Force = workload.Request.Force
	if workload.Manifest.Module == "" {
		workload.Manifest.Module = workload.Request.Module
	}
	if workload.Manifest.Name == "" {
		workload.Manifest.Name = workload.Request.Module
	}
	if workload.Manifest.Kind == "" {
		workload.Manifest.Kind = workload.Request.Kind
	}
	workload.Manifest.Force = workload.Request.Force
	if workload.ManifestPath == "" && workload.Manifest.Module != "" {
		workload.ManifestPath = filepath.Join("server", "modules", workload.Manifest.Module, "manifest.yaml")
	}
	return workload
}

func buildApplyItems(workload Workload, stage lifecycle.ApplyStage, label string) []lifecycle.ApplyItem {
	items := make([]lifecycle.ApplyItem, 0, 3)
	module := workload.Request.Module
	if module == "" {
		module = workload.Manifest.Module
	}
	if module != "" {
		items = append(items, lifecycle.ApplyItem{Module: module, Kind: label, Stage: stage, Action: "planned", Managed: true})
	}
	return items
}

func applyItemActions(items []lifecycle.ApplyItem, action string) []lifecycle.ApplyItem {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]lifecycle.ApplyItem, 0, len(items))
	for _, item := range items {
		item.Action = action
		cloned = append(cloned, item)
	}
	return cloned
}

func (s *Service) buildValidationReport(workload Workload, result lifecycle.ApplyResult, installResult installapp.InstallResult) lifecycle.ApplyValidationReport {
	now := time.Now().UTC()
	report := lifecycle.ApplyValidationReport{StartedAt: now, FinishedAt: now, Status: "failed", Verified: false}
	manifestPath := strings.TrimSpace(workload.ManifestPath)
	if manifestPath != "" {
		if !filepath.IsAbs(manifestPath) && s.projectRoot != "" {
			manifestPath = filepath.Join(s.projectRoot, manifestPath)
		}
		if _, err := os.Stat(manifestPath); err == nil {
			report.Checked++
		} else {
			report.Issues = append(report.Issues, lifecycle.ApplyValidationIssue{Stage: lifecycle.ApplyStageGenerate, Message: fmt.Sprintf("manifest not found at %s", manifestPath), Actual: err.Error()})
		}
	}
	if len(result.PolicySync.Synced) > 0 || len(result.PolicySync.Skipped) > 0 {
		report.Checked++
	}
	if workload.Request.Install {
		if strings.TrimSpace(installResult.ManifestPath) != "" {
			if _, err := os.Stat(installResult.ManifestPath); err == nil {
				report.Checked++
			} else {
				report.Issues = append(report.Issues, lifecycle.ApplyValidationIssue{Stage: lifecycle.ApplyStageInstall, Message: fmt.Sprintf("installed manifest not found at %s", installResult.ManifestPath), Actual: err.Error()})
			}
		}
	}
	if len(report.Issues) == 0 {
		report.Verified = true
		report.Status = "passed"
	} else {
		report.Status = "failed"
	}
	return report
}

func buildPolicySyncItems(workload Workload) []lifecycle.ApplyItem {
	kind := normalizeKind(workload.Request.Kind)
	module := workload.Request.Module
	if module == "" {
		module = workload.Manifest.Module
	}
	items := make([]lifecycle.ApplyItem, 0, 4)
	switch kind {
	case "crud":
		if !workload.Request.GeneratePolicy && len(workload.PolicyLines) > 0 {
			for _, line := range workload.PolicyLines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				items = append(items, lifecycle.ApplyItem{Module: module, Kind: "policy-line", Ref: line, Stage: lifecycle.ApplyStagePolicySync, Action: "planned", Managed: true})
			}
			break
		}
		if !workload.Request.GeneratePolicy {
			break
		}
		routes := workload.CRUD.ManifestRoutes
		if len(routes) == 0 {
			routes = workload.Manifest.Routes
		}
		for _, route := range routes {
			items = append(items, lifecycle.ApplyItem{
				Module:  module,
				Kind:    "policy-rule",
				Path:    strings.TrimSpace(route.Path),
				Ref:     strings.TrimSpace(route.Method),
				Stage:   lifecycle.ApplyStagePolicySync,
				Action:  "planned",
				Managed: true,
			})
		}
		if len(items) == 0 && workload.Request.GeneratePolicy {
			items = append(items, lifecycle.ApplyItem{Module: module, Kind: "policy-rule", Stage: lifecycle.ApplyStagePolicySync, Action: "planned", Managed: true})
		}
	case "plugin":
		pluginName := strings.TrimSpace(workload.Plugin.Name)
		if pluginName == "" {
			pluginName = module
		}
		pluginName = legacygenerate.ToSnake(pluginName)
		items = append(items, lifecycle.ApplyItem{
			Module:  module,
			Kind:    "policy-rule",
			Path:    "/plugins/" + pluginName + "/ping",
			Ref:     "GET",
			Stage:   lifecycle.ApplyStagePolicySync,
			Action:  "planned",
			Managed: true,
		})
	default:
		for _, line := range workload.PolicyLines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			items = append(items, lifecycle.ApplyItem{Module: module, Kind: "policy-line", Ref: line, Stage: lifecycle.ApplyStagePolicySync, Action: "planned", Managed: true})
		}
	}
	return items
}

func buildInstalledItems(workload Workload, result installapp.InstallResult) []lifecycle.ApplyItem {
	module := workload.Request.Module
	if module == "" {
		module = workload.Manifest.Module
	}
	items := make([]lifecycle.ApplyItem, 0, len(result.Menus))
	for _, menu := range result.Menus {
		action := strings.TrimSpace(menu.Action)
		if action == "" {
			action = "installed"
		}
		items = append(items, lifecycle.ApplyItem{
			Module:  module,
			Kind:    "runtime-menu",
			Path:    strings.TrimSpace(menu.Path),
			Ref:     strings.TrimSpace(menu.MenuID),
			Stage:   lifecycle.ApplyStageInstall,
			Action:  action,
			Managed: true,
		})
	}
	if len(items) == 0 && strings.TrimSpace(result.ManifestPath) != "" {
		items = append(items, lifecycle.ApplyItem{
			Module:  module,
			Kind:    "manifest-install",
			Path:    strings.TrimSpace(result.ManifestPath),
			Stage:   lifecycle.ApplyStageInstall,
			Action:  "installed",
			Managed: true,
		})
	}
	return items
}

func summarizeApplyPlan(plan lifecycle.ApplyPlan) lifecycle.ApplyPlanSummary {
	return lifecycle.ApplyPlanSummary{
		Generation:   len(plan.Generation),
		Installation: len(plan.Installation),
		PolicySync:   len(plan.PolicySync),
		Validation:   len(plan.Validation),
		Warnings:     len(plan.Warnings),
		Conflicts:    len(plan.Conflicts),
		Total:        len(plan.Generation) + len(plan.Installation) + len(plan.PolicySync) + len(plan.Validation),
	}
}

func summarizeApplyResult(result lifecycle.ApplyResult) lifecycle.ApplyResultSummary {
	generated := len(result.Generation.Created) + len(result.Generation.Updated)
	installed := len(result.Installation.Created) + len(result.Installation.Updated)
	policySynced := len(result.PolicySync.Synced)
	validationChecks := result.Validation.Checked
	validationIssues := len(result.Validation.Issues)
	skipped := len(result.Generation.Skipped) + len(result.Installation.Skipped) + len(result.PolicySync.Skipped)
	failed := len(result.Failures)
	elapsed := int64(0)
	if !result.StartedAt.IsZero() && !result.FinishedAt.IsZero() {
		elapsed = result.FinishedAt.Sub(result.StartedAt).Milliseconds()
	}
	return lifecycle.ApplyResultSummary{
		Generated:        generated,
		Installed:        installed,
		PolicySynced:     policySynced,
		ValidationChecks: validationChecks,
		ValidationIssues: validationIssues,
		Skipped:          skipped,
		Failed:           failed,
		TotalApplied:     generated + installed + policySynced,
		ElapsedMillis:    elapsed,
	}
}

func normalizeName(value string) string {
	return strings.TrimSpace(value)
}

func normalizeKind(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

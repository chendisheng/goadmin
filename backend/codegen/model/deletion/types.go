package deletion

import (
	"fmt"
	"strings"
	"time"
)

type PolicyStoreKind string

const (
	PolicyStoreUnknown PolicyStoreKind = ""
	PolicyStoreCSV     PolicyStoreKind = "csv"
	PolicyStoreDB      PolicyStoreKind = "db"
)

func NormalizePolicyStoreKind(value string) PolicyStoreKind {
	switch normalizeToken(value) {
	case string(PolicyStoreCSV):
		return PolicyStoreCSV
	case string(PolicyStoreDB):
		return PolicyStoreDB
	default:
		return PolicyStoreUnknown
	}
}

func (k PolicyStoreKind) IsKnown() bool {
	return k == PolicyStoreCSV || k == PolicyStoreDB
}

type LegacyCompatibilityMode string

const (
	LegacyCompatibilityModeUnknown       LegacyCompatibilityMode = ""
	LegacyCompatibilityModeConservative  LegacyCompatibilityMode = "conservative"
	LegacyCompatibilityModeTemplateInfer LegacyCompatibilityMode = "template_infer"
)

func NormalizeLegacyCompatibilityMode(value string) LegacyCompatibilityMode {
	switch normalizeToken(value) {
	case string(LegacyCompatibilityModeConservative):
		return LegacyCompatibilityModeConservative
	case string(LegacyCompatibilityModeTemplateInfer):
		return LegacyCompatibilityModeTemplateInfer
	default:
		return LegacyCompatibilityModeUnknown
	}
}

func (m LegacyCompatibilityMode) IsKnown() bool {
	return m == LegacyCompatibilityModeConservative || m == LegacyCompatibilityModeTemplateInfer
}

func (m LegacyCompatibilityMode) IsPreviewOnly() bool {
	return m == LegacyCompatibilityModeConservative
}

func (m LegacyCompatibilityMode) AllowsExecution() bool {
	return m == LegacyCompatibilityModeTemplateInfer
}

type AssetOrigin string

const (
	AssetOriginUnknown   AssetOrigin = ""
	AssetOriginGenerated AssetOrigin = "generated"
	AssetOriginInferred  AssetOrigin = "inferred"
	AssetOriginManual    AssetOrigin = "manual"
	AssetOriginShared    AssetOrigin = "shared"
)

type AssetKind string

const (
	AssetKindUnknown           AssetKind = ""
	AssetKindSourceFile        AssetKind = "source-file"
	AssetKindSourceDirectory   AssetKind = "source-directory"
	AssetKindRuntimeRegistry   AssetKind = "runtime-registry"
	AssetKindRuntimeRoute      AssetKind = "runtime-route"
	AssetKindRuntimeMenu       AssetKind = "runtime-menu"
	AssetKindRuntimePermission AssetKind = "runtime-permission"
	AssetKindRuntimePage       AssetKind = "runtime-page"
	AssetKindPolicyRule        AssetKind = "policy-rule"
	AssetKindFrontendFile      AssetKind = "frontend-file"
)

type DeleteRequest struct {
	Module        string                  `json:"module,omitempty"`
	Kind          string                  `json:"kind,omitempty"`
	DryRun        bool                    `json:"dry_run,omitempty"`
	Force         bool                    `json:"force,omitempty"`
	WithPolicy    bool                    `json:"with_policy,omitempty"`
	WithRuntime   bool                    `json:"with_runtime,omitempty"`
	WithFrontend  bool                    `json:"with_frontend,omitempty"`
	WithRegistry  bool                    `json:"with_registry,omitempty"`
	PolicyStore   PolicyStoreKind         `json:"policy_store,omitempty"`
	Compatibility LegacyCompatibilityRule `json:"compatibility,omitempty"`
	MetadataHints map[string]any          `json:"metadata_hints,omitempty"`
}

func (r DeleteRequest) Normalize() DeleteRequest {
	r.Module = strings.TrimSpace(r.Module)
	r.Kind = strings.TrimSpace(r.Kind)
	r.PolicyStore = NormalizePolicyStoreKind(string(r.PolicyStore))
	r.Compatibility = r.Compatibility.Normalize()
	return r
}

func (r DeleteRequest) Validate() error {
	normalized := r.Normalize()
	if normalized.Module == "" {
		return fmt.Errorf("module is required")
	}
	if normalized.PolicyStore != PolicyStoreUnknown && !normalized.PolicyStore.IsKnown() {
		return fmt.Errorf("unknown policy store %q", normalized.PolicyStore)
	}
	if normalized.Compatibility.Mode != LegacyCompatibilityModeUnknown && !normalized.Compatibility.Mode.IsKnown() {
		return fmt.Errorf("unknown compatibility mode %q", normalized.Compatibility.Mode)
	}
	return nil
}

type LegacyCompatibilityRule struct {
	Mode                   LegacyCompatibilityMode `json:"mode,omitempty"`
	RequireManifest        bool                    `json:"require_manifest,omitempty"`
	RequireExplicitConfirm bool                    `json:"require_explicit_confirm,omitempty"`
	AllowPathInference     bool                    `json:"allow_path_inference,omitempty"`
	ManifestPaths          []string                `json:"manifest_paths,omitempty"`
	ModuleRoots            []string                `json:"module_roots,omitempty"`
	OwnedFilePatterns      []string                `json:"owned_file_patterns,omitempty"`
	FallbackPolicyStores   []PolicyStoreKind       `json:"fallback_policy_stores,omitempty"`
	Notes                  []string                `json:"notes,omitempty"`
}

func (r LegacyCompatibilityRule) Normalize() LegacyCompatibilityRule {
	r.Mode = NormalizeLegacyCompatibilityMode(string(r.Mode))
	r.ManifestPaths = normalizeStringSlice(r.ManifestPaths)
	r.ModuleRoots = normalizeStringSlice(r.ModuleRoots)
	r.OwnedFilePatterns = normalizeStringSlice(r.OwnedFilePatterns)
	r.Notes = normalizeStringSlice(r.Notes)
	if len(r.FallbackPolicyStores) > 0 {
		normalized := make([]PolicyStoreKind, 0, len(r.FallbackPolicyStores))
		for _, store := range r.FallbackPolicyStores {
			if normalizedStore := NormalizePolicyStoreKind(string(store)); normalizedStore.IsKnown() {
				normalized = append(normalized, normalizedStore)
			}
		}
		r.FallbackPolicyStores = normalized
	}
	return r
}

func (r LegacyCompatibilityRule) IsPreviewOnly() bool {
	return r.Normalize().Mode.IsPreviewOnly()
}

func (r LegacyCompatibilityRule) AllowsExecution() bool {
	return r.Normalize().Mode.AllowsExecution() && !r.RequireExplicitConfirm
}

type ModuleOwnership struct {
	Module           string                  `json:"module,omitempty"`
	Kind             string                  `json:"kind,omitempty"`
	GeneratorVersion string                  `json:"generator_version,omitempty"`
	GeneratedAt      time.Time               `json:"generated_at,omitempty"`
	ManifestPath     string                  `json:"manifest_path,omitempty"`
	ManifestFormat   string                  `json:"manifest_format,omitempty"`
	Source           string                  `json:"source,omitempty"`
	OwnedFiles       []DeleteItem            `json:"owned_files,omitempty"`
	RuntimeAssets    []DeleteItem            `json:"runtime_assets,omitempty"`
	PolicyAssets     []PolicyAsset           `json:"policy_assets,omitempty"`
	FrontendAssets   []DeleteItem            `json:"frontend_assets,omitempty"`
	Compatibility    LegacyCompatibilityRule `json:"compatibility,omitempty"`
	Metadata         map[string]any          `json:"metadata,omitempty"`
}

type PolicyAsset struct {
	Store       PolicyStoreKind `json:"store,omitempty"`
	Module      string          `json:"module,omitempty"`
	SourceRef   string          `json:"source_ref,omitempty"`
	PType       string          `json:"ptype,omitempty"`
	V0          string          `json:"v0,omitempty"`
	V1          string          `json:"v1,omitempty"`
	V2          string          `json:"v2,omitempty"`
	V3          string          `json:"v3,omitempty"`
	V4          string          `json:"v4,omitempty"`
	V5          string          `json:"v5,omitempty"`
	Managed     bool            `json:"managed,omitempty"`
	GeneratedAt time.Time       `json:"generated_at,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

func (a PolicyAsset) Selector() PolicySelector {
	return PolicySelector{
		Store:     a.Store,
		Module:    strings.TrimSpace(a.Module),
		SourceRef: strings.TrimSpace(a.SourceRef),
		PType:     strings.TrimSpace(a.PType),
		V0:        strings.TrimSpace(a.V0),
		V1:        strings.TrimSpace(a.V1),
		V2:        strings.TrimSpace(a.V2),
		V3:        strings.TrimSpace(a.V3),
		V4:        strings.TrimSpace(a.V4),
		V5:        strings.TrimSpace(a.V5),
	}
}

type PolicySelector struct {
	Store     PolicyStoreKind `json:"store,omitempty"`
	Module    string          `json:"module,omitempty"`
	SourceRef string          `json:"source_ref,omitempty"`
	PType     string          `json:"ptype,omitempty"`
	V0        string          `json:"v0,omitempty"`
	V1        string          `json:"v1,omitempty"`
	V2        string          `json:"v2,omitempty"`
	V3        string          `json:"v3,omitempty"`
	V4        string          `json:"v4,omitempty"`
	V5        string          `json:"v5,omitempty"`
	Metadata  map[string]any  `json:"metadata,omitempty"`
}

func (s PolicySelector) Values() [6]string {
	return [6]string{s.V0, s.V1, s.V2, s.V3, s.V4, s.V5}
}

type DeleteItem struct {
	Module   string          `json:"module,omitempty"`
	Kind     AssetKind       `json:"kind,omitempty"`
	Path     string          `json:"path,omitempty"`
	Ref      string          `json:"ref,omitempty"`
	Store    PolicyStoreKind `json:"store,omitempty"`
	Selector *PolicySelector `json:"selector,omitempty"`
	Origin   AssetOrigin     `json:"origin,omitempty"`
	Managed  bool            `json:"managed,omitempty"`
	Metadata map[string]any  `json:"metadata,omitempty"`
}

type DeleteConflict struct {
	Kind     string         `json:"kind,omitempty"`
	Severity string         `json:"severity,omitempty"`
	Message  string         `json:"message,omitempty"`
	Path     string         `json:"path,omitempty"`
	Ref      string         `json:"ref,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type DeletePlanSummary struct {
	SourceFiles     int `json:"source_files,omitempty"`
	RuntimeAssets   int `json:"runtime_assets,omitempty"`
	RegistryChanges int `json:"registry_changes,omitempty"`
	PolicyChanges   int `json:"policy_changes,omitempty"`
	FrontendChanges int `json:"frontend_changes,omitempty"`
	Warnings        int `json:"warnings,omitempty"`
	Conflicts       int `json:"conflicts,omitempty"`
	Total           int `json:"total,omitempty"`
}

type DeletePlan struct {
	Request         DeleteRequest           `json:"request,omitempty"`
	Ownership       ModuleOwnership         `json:"ownership,omitempty"`
	Module          string                  `json:"module,omitempty"`
	DryRun          bool                    `json:"dry_run,omitempty"`
	Force           bool                    `json:"force,omitempty"`
	PolicyStore     PolicyStoreKind         `json:"policy_store,omitempty"`
	PolicyStores    []PolicyStoreKind       `json:"policy_stores,omitempty"`
	SourceFiles     []DeleteItem            `json:"source_files,omitempty"`
	RuntimeAssets   []DeleteItem            `json:"runtime_assets,omitempty"`
	RegistryChanges []DeleteItem            `json:"registry_changes,omitempty"`
	PolicyChanges   []DeleteItem            `json:"policy_changes,omitempty"`
	FrontendChanges []DeleteItem            `json:"frontend_changes,omitempty"`
	Warnings        []string                `json:"warnings,omitempty"`
	Conflicts       []DeleteConflict        `json:"conflicts,omitempty"`
	Legacy          LegacyCompatibilityRule `json:"legacy,omitempty"`
	Summary         DeletePlanSummary       `json:"summary,omitempty"`
}

type DeleteStatus string

const (
	DeleteStatusUnknown   DeleteStatus = ""
	DeleteStatusPlanned   DeleteStatus = "planned"
	DeleteStatusDryRun    DeleteStatus = "dry_run"
	DeleteStatusSucceeded DeleteStatus = "succeeded"
	DeleteStatusPartial   DeleteStatus = "partial"
	DeleteStatusFailed    DeleteStatus = "failed"
)

type DeleteFailure struct {
	Item        DeleteItem `json:"item,omitempty"`
	Reason      string     `json:"reason,omitempty"`
	Recoverable bool       `json:"recoverable,omitempty"`
}

type DeleteResultSummary struct {
	DeletedSourceFiles     int   `json:"deleted_source_files,omitempty"`
	DeletedRuntimeAssets   int   `json:"deleted_runtime_assets,omitempty"`
	DeletedRegistryChanges int   `json:"deleted_registry_changes,omitempty"`
	DeletedPolicyChanges   int   `json:"deleted_policy_changes,omitempty"`
	DeletedFrontendChanges int   `json:"deleted_frontend_changes,omitempty"`
	Skipped                int   `json:"skipped,omitempty"`
	Failed                 int   `json:"failed,omitempty"`
	TotalDeleted           int   `json:"total_deleted,omitempty"`
	ElapsedMillis          int64 `json:"elapsed_millis,omitempty"`
}

type DeleteResult struct {
	Request    DeleteRequest       `json:"request,omitempty"`
	Plan       DeletePlan          `json:"plan,omitempty"`
	Status     DeleteStatus        `json:"status,omitempty"`
	StartedAt  time.Time           `json:"started_at,omitempty"`
	FinishedAt time.Time           `json:"finished_at,omitempty"`
	Deleted    []DeleteItem        `json:"deleted,omitempty"`
	Skipped    []DeleteItem        `json:"skipped,omitempty"`
	Failures   []DeleteFailure     `json:"failures,omitempty"`
	Warnings   []string            `json:"warnings,omitempty"`
	Summary    DeleteResultSummary `json:"summary,omitempty"`
}

func normalizeToken(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "-", "_")
	return strings.ToLower(value)
}

func normalizeStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

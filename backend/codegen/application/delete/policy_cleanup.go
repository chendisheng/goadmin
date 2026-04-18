package deleteapp

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	lifecycle "goadmin/codegen/model/lifecycle"

	"gorm.io/gorm"
)

type PolicyCacheRefresher interface {
	Reload() error
}

type PolicyStore interface {
	Kind() lifecycle.PolicyStoreKind
	ListByModule(ctx context.Context, module string) ([]lifecycle.PolicyAsset, error)
	DeleteBySelector(ctx context.Context, selector lifecycle.PolicySelector) (int, error)
	Validate(ctx context.Context) error
	Preview(ctx context.Context, req PolicyCleanupRequest) (PolicyCleanupPreview, error)
	Delete(ctx context.Context, req PolicyCleanupRequest) (PolicyCleanupResult, error)
}

type PolicyCleanupDependencies struct {
	ProjectRoot string
	BackendRoot string
	PolicyPath  string
	Store       lifecycle.PolicyStoreKind
	DB          *gorm.DB
	Refresher   PolicyCacheRefresher
}

type PolicyCleanupService struct {
	store     PolicyStore
	storeKind lifecycle.PolicyStoreKind
	refresher PolicyCacheRefresher
}

type PolicyCleanupRequest struct {
	Module            string                     `json:"module,omitempty"`
	Store             lifecycle.PolicyStoreKind  `json:"store,omitempty"`
	Selectors         []lifecycle.PolicySelector `json:"selectors,omitempty"`
	RequireManaged    bool                       `json:"require_managed,omitempty"`
	RequireValidation bool                       `json:"validate,omitempty"`
	Refresh           bool                       `json:"refresh,omitempty"`
}

type PolicyCleanupItem struct {
	Selector   lifecycle.PolicySelector `json:"selector,omitempty"`
	Rule       lifecycle.PolicyAsset    `json:"rule,omitempty"`
	Decision   string                   `json:"decision,omitempty"`
	Reason     string                   `json:"reason,omitempty"`
	MatchCount int                      `json:"match_count,omitempty"`
}

type PolicyCleanupSummary struct {
	Total     int    `json:"total,omitempty"`
	Selected  int    `json:"selected,omitempty"`
	Deleted   int    `json:"deleted,omitempty"`
	Skipped   int    `json:"skipped,omitempty"`
	Conflicts int    `json:"conflicts,omitempty"`
	Failures  int    `json:"failures,omitempty"`
	Backend   string `json:"backend,omitempty"`
	Validated bool   `json:"validated,omitempty"`
}

type PolicyCleanupAudit struct {
	StartedAt      time.Time `json:"started_at,omitempty"`
	FinishedAt     time.Time `json:"finished_at,omitempty"`
	Backend        string    `json:"backend,omitempty"`
	Operation      string    `json:"operation,omitempty"`
	Validation     string    `json:"validation,omitempty"`
	CacheRefreshed bool      `json:"cache_refreshed,omitempty"`
}

type PolicyCleanupPreview struct {
	Request   PolicyCleanupRequest       `json:"request,omitempty"`
	Store     lifecycle.PolicyStoreKind  `json:"store,omitempty"`
	Items     []PolicyCleanupItem        `json:"items,omitempty"`
	Conflicts []lifecycle.DeleteConflict `json:"conflicts,omitempty"`
	Warnings  []string                   `json:"warnings,omitempty"`
	Summary   PolicyCleanupSummary       `json:"summary,omitempty"`
}

type PolicyCleanupResult struct {
	Preview  PolicyCleanupPreview      `json:"preview,omitempty"`
	Deleted  []PolicyCleanupItem       `json:"deleted,omitempty"`
	Skipped  []PolicyCleanupItem       `json:"skipped,omitempty"`
	Failures []lifecycle.DeleteFailure `json:"failures,omitempty"`
	Warnings []string                  `json:"warnings,omitempty"`
	Audit    PolicyCleanupAudit        `json:"audit,omitempty"`
	Verified bool                      `json:"verified,omitempty"`
	Summary  PolicyCleanupSummary      `json:"summary,omitempty"`
}

func NewPolicyCleanupService(deps PolicyCleanupDependencies) (*PolicyCleanupService, error) {
	backendRoot := strings.TrimSpace(deps.BackendRoot)
	if backendRoot == "" {
		backendRoot = resolveBackendRoot(deps.ProjectRoot)
	}
	policyPath := strings.TrimSpace(deps.PolicyPath)
	if policyPath == "" {
		policyPath = filepath.Join(backendRoot, "core", "auth", "casbin", "adapter", "policy.csv")
	}
	storeKind := deps.Store
	if !storeKind.IsKnown() {
		if deps.DB != nil {
			storeKind = lifecycle.PolicyStoreDB
		} else {
			storeKind = lifecycle.PolicyStoreCSV
		}
	}
	var store PolicyStore
	var err error
	switch storeKind {
	case lifecycle.PolicyStoreCSV:
		store, err = NewCSVPolicyStore(policyPath)
	case lifecycle.PolicyStoreDB:
		store, err = NewDBPolicyStore(deps.DB, deps.Refresher)
	default:
		err = fmt.Errorf("unsupported policy store %q", storeKind)
	}
	if err != nil {
		return nil, err
	}
	return &PolicyCleanupService{store: store, storeKind: storeKind, refresher: deps.Refresher}, nil
}

func (s *PolicyCleanupService) Preview(ctx context.Context, req PolicyCleanupRequest) (PolicyCleanupPreview, error) {
	store, err := s.resolveStore(req.Store)
	if err != nil {
		return PolicyCleanupPreview{}, err
	}
	normalized := req.Normalize(store.Kind())
	return store.Preview(ctx, normalized)
}

func (s *PolicyCleanupService) Delete(ctx context.Context, req PolicyCleanupRequest) (PolicyCleanupResult, error) {
	store, err := s.resolveStore(req.Store)
	if err != nil {
		return PolicyCleanupResult{}, err
	}
	normalized := req.Normalize(store.Kind())
	return store.Delete(ctx, normalized)
}

func (s *PolicyCleanupService) resolveStore(kind lifecycle.PolicyStoreKind) (PolicyStore, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("policy cleanup service is not configured")
	}
	if kind.IsKnown() && s.store.Kind() != kind {
		return nil, fmt.Errorf("policy store %q is not configured", kind)
	}
	return s.store, nil
}

func (r PolicyCleanupRequest) Normalize(defaultStore lifecycle.PolicyStoreKind) PolicyCleanupRequest {
	r.Module = strings.TrimSpace(r.Module)
	r.Store = choosePolicyStoreKind(r.Store, defaultStore)
	if len(r.Selectors) > 0 {
		selectors := make([]lifecycle.PolicySelector, 0, len(r.Selectors))
		seen := make(map[string]struct{}, len(r.Selectors))
		for _, selector := range r.Selectors {
			selector.Store = choosePolicyStoreKind(selector.Store, r.Store)
			selector.Module = strings.TrimSpace(selector.Module)
			selector.SourceRef = strings.TrimSpace(selector.SourceRef)
			selector.PType = strings.TrimSpace(selector.PType)
			selector.V0 = strings.TrimSpace(selector.V0)
			selector.V1 = strings.TrimSpace(selector.V1)
			selector.V2 = strings.TrimSpace(selector.V2)
			selector.V3 = strings.TrimSpace(selector.V3)
			selector.V4 = strings.TrimSpace(selector.V4)
			selector.V5 = strings.TrimSpace(selector.V5)
			selector.Metadata = cloneAnyMap(selector.Metadata)
			key := policySelectorKey(selector)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			selectors = append(selectors, selector)
		}
		sort.SliceStable(selectors, func(i, j int) bool {
			return policySelectorKey(selectors[i]) < policySelectorKey(selectors[j])
		})
		r.Selectors = selectors
	}
	if !r.RequireManaged {
		r.RequireManaged = true
	}
	if !r.RequireValidation {
		r.RequireValidation = true
	}
	return r
}

func (r PolicyCleanupRequest) ValidateRequest() error {
	if r.Store == lifecycle.PolicyStoreUnknown {
		return fmt.Errorf("policy store is required")
	}
	if len(r.Selectors) == 0 {
		return fmt.Errorf("policy selectors are required")
	}
	return nil
}

func BuildPolicyCleanupRequest(plan lifecycle.DeletePlan) PolicyCleanupRequest {
	selectors := make([]lifecycle.PolicySelector, 0, len(plan.PolicyChanges))
	for _, item := range plan.PolicyChanges {
		if item.Selector == nil {
			continue
		}
		selectors = append(selectors, *item.Selector)
	}
	return PolicyCleanupRequest{
		Module:            plan.Module,
		Store:             plan.PolicyStore,
		Selectors:         selectors,
		RequireManaged:    true,
		RequireValidation: true,
		Refresh:           true,
	}
}

func choosePolicyStoreKind(value, fallback lifecycle.PolicyStoreKind) lifecycle.PolicyStoreKind {
	if value.IsKnown() {
		return value
	}
	if fallback.IsKnown() {
		return fallback
	}
	return lifecycle.PolicyStoreUnknown
}

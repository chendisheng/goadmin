package deleteapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	lifecycle "goadmin/codegen/model/lifecycle"

	"gorm.io/gorm"
)

type DBPolicyStore struct {
	db        *gorm.DB
	refresher PolicyCacheRefresher
}

func NewDBPolicyStore(db *gorm.DB, refresher PolicyCacheRefresher) (*DBPolicyStore, error) {
	if db == nil {
		return nil, fmt.Errorf("db policy store requires db")
	}
	return &DBPolicyStore{db: db, refresher: refresher}, nil
}

func (s *DBPolicyStore) Kind() lifecycle.PolicyStoreKind {
	return lifecycle.PolicyStoreDB
}

func (s *DBPolicyStore) ListByModule(ctx context.Context, module string) ([]lifecycle.PolicyAsset, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("db policy store is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	rules, err := loadDBRules(ctx, s.db)
	if err != nil {
		return nil, err
	}
	return policyAssetsByModule(rules, module), nil
}

func (s *DBPolicyStore) DeleteBySelector(ctx context.Context, selector lifecycle.PolicySelector) (int, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("db policy store is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	res := s.db.WithContext(ctx).Where(buildDBSelectorQuery(selector)).Delete(&dbPolicyRecord{})
	if res.Error != nil {
		return 0, fmt.Errorf("delete casbin policy: %w", res.Error)
	}
	return int(res.RowsAffected), nil
}

func (s *DBPolicyStore) Validate(ctx context.Context) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("db policy store is not configured")
	}
	return validatePolicyStoreRules(ctx, func(ctx context.Context) ([]lifecycle.PolicyAsset, error) {
		return loadDBRules(ctx, s.db)
	})
}

func (s *DBPolicyStore) Preview(ctx context.Context, req PolicyCleanupRequest) (PolicyCleanupPreview, error) {
	if s == nil || s.db == nil {
		return PolicyCleanupPreview{}, fmt.Errorf("db policy store is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req = req.Normalize(s.Kind())
	if err := req.ValidateRequest(); err != nil {
		return PolicyCleanupPreview{}, err
	}
	rules, err := s.ListByModule(ctx, req.Module)
	if err != nil {
		return PolicyCleanupPreview{}, err
	}
	return buildPolicyCleanupPreview(req, s.Kind(), rules), nil
}

func (s *DBPolicyStore) Delete(ctx context.Context, req PolicyCleanupRequest) (PolicyCleanupResult, error) {
	if s == nil || s.db == nil {
		return PolicyCleanupResult{}, fmt.Errorf("db policy store is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req = req.Normalize(s.Kind())
	if err := req.ValidateRequest(); err != nil {
		return PolicyCleanupResult{}, err
	}
	started := nowUTC()
	preview, err := s.Preview(ctx, req)
	result := newPolicyCleanupResult(preview, started)
	if err != nil {
		return result, err
	}
	if len(preview.Conflicts) > 0 {
		result.Warnings = append(result.Warnings, "policy cleanup aborted because preview contained conflicts")
		result.Audit.FinishedAt = nowUTC()
		result.Audit.Validation = "blocked"
		return result, fmt.Errorf("policy cleanup preview has conflicts")
	}
	if len(preview.Items) == 0 {
		finalizePolicyCleanupResult(&result, preview, nil, nil, true, "no-op", false)
		return result, nil
	}
	var deleted []PolicyCleanupItem
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, selector := range req.Selectors {
			if !selectorManaged(selector) || !selectorMatchesModule(selector, req.Module) {
				continue
			}
			affected, deleteErr := func() (int, error) {
				res := tx.Where(buildDBSelectorQuery(selector)).Delete(&dbPolicyRecord{})
				if res.Error != nil {
					return 0, fmt.Errorf("delete casbin policy: %w", res.Error)
				}
				return int(res.RowsAffected), nil
			}()
			if deleteErr != nil {
				return deleteErr
			}
			if affected == 0 {
				continue
			}
			deleted = append(deleted, PolicyCleanupItem{
				Selector:   selector,
				Rule:       selectorToPolicyAsset(selector, selector.SourceRef),
				Decision:   "delete",
				Reason:     "matched by structured selector",
				MatchCount: affected,
			})
		}
		return nil
	})
	if err != nil {
		result.Audit.FinishedAt = nowUTC()
		result.Audit.Validation = "rolled_back"
		return result, err
	}
	if req.Refresh && s.refresher != nil {
		if reloadErr := s.refresher.Reload(); reloadErr != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("policy cache reload failed: %v", reloadErr))
		} else {
			result.Audit.CacheRefreshed = true
		}
	}
	verifiedRules, err := loadDBRules(ctx, s.db)
	if err != nil {
		result.Audit.FinishedAt = nowUTC()
		result.Audit.Validation = "verify_failed"
		return result, err
	}
	result.Verified = verifyPolicySelections(verifiedRules, req.Selectors)
	if !result.Verified {
		result.Audit.FinishedAt = nowUTC()
		result.Audit.Validation = "verify_failed"
		return result, fmt.Errorf("policy cleanup verification failed")
	}
	finalizePolicyCleanupResult(&result, preview, deleted, collectSkippedItems(preview.Items, deleted), true, "ok", result.Audit.CacheRefreshed)
	return result, nil
}

type dbPolicyRecord struct {
	ID        uint   `gorm:"primaryKey"`
	PType     string `gorm:"column:ptype;type:varchar(32);not null"`
	V0        string `gorm:"column:v0;type:varchar(191)"`
	V1        string `gorm:"column:v1;type:varchar(191)"`
	V2        string `gorm:"column:v2;type:varchar(191)"`
	V3        string `gorm:"column:v3;type:varchar(255)"`
	V4        string `gorm:"column:v4;type:varchar(255)"`
	V5        string `gorm:"column:v5;type:varchar(255)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (dbPolicyRecord) TableName() string { return "casbin_rule" }

func loadDBRules(ctx context.Context, db *gorm.DB) ([]lifecycle.PolicyAsset, error) {
	if db == nil {
		return nil, fmt.Errorf("db policy store requires db")
	}
	var records []dbPolicyRecord
	if err := db.WithContext(ctx).Where("ptype = ?", "p").Order("id ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("load casbin policies: %w", err)
	}
	rules := make([]lifecycle.PolicyAsset, 0, len(records))
	for _, record := range records {
		if strings.TrimSpace(record.V0) == "" || strings.TrimSpace(record.V1) == "" || strings.TrimSpace(record.V2) == "" {
			continue
		}
		rules = append(rules, lifecycle.PolicyAsset{
			Store:     lifecycle.PolicyStoreDB,
			SourceRef: fmt.Sprintf("casbin_rule:%d", record.ID),
			PType:     "p",
			V0:        strings.TrimSpace(record.V0),
			V1:        strings.TrimSpace(record.V1),
			V2:        strings.TrimSpace(record.V2),
			V3:        strings.TrimSpace(record.V3),
			V4:        strings.TrimSpace(record.V4),
			V5:        strings.TrimSpace(record.V5),
		})
	}
	return rules, nil
}

func buildDBSelectorQuery(selector lifecycle.PolicySelector) map[string]any {
	query := map[string]any{"ptype": "p"}
	if strings.TrimSpace(selector.V0) != "" {
		query["v0"] = strings.TrimSpace(selector.V0)
	}
	if strings.TrimSpace(selector.V1) != "" {
		query["v1"] = strings.TrimSpace(selector.V1)
	}
	if strings.TrimSpace(selector.V2) != "" {
		query["v2"] = strings.TrimSpace(selector.V2)
	}
	if strings.TrimSpace(selector.V3) != "" {
		query["v3"] = strings.TrimSpace(selector.V3)
	}
	if strings.TrimSpace(selector.V4) != "" {
		query["v4"] = strings.TrimSpace(selector.V4)
	}
	if strings.TrimSpace(selector.V5) != "" {
		query["v5"] = strings.TrimSpace(selector.V5)
	}
	return query
}

func verifyPolicySelectionsDB(rules []lifecycle.PolicyAsset, selectors []lifecycle.PolicySelector) bool {
	return verifyPolicySelections(rules, selectors)
}

func (s *DBPolicyStore) reloadAfterDelete() error {
	if s == nil || s.refresher == nil {
		return nil
	}
	return s.refresher.Reload()
}

func selectorReason(selector lifecycle.PolicySelector, module string) (string, bool) {
	if !selectorManaged(selector) {
		return "selector is not marked as managed", false
	}
	if !selectorMatchesModule(selector, module) {
		return "selector module does not match request module", false
	}
	return "matched by structured selector", true
}

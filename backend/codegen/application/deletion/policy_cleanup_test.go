package deletion

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	deletionmodel "goadmin/codegen/model/deletion"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestPolicyCleanupRequestNormalizeAndBuildRequest(t *testing.T) {
	t.Parallel()

	req := PolicyCleanupRequest{
		Module: " book ",
		Store:  deletionmodel.PolicyStoreUnknown,
		Selectors: []deletionmodel.PolicySelector{
			{
				Store:    deletionmodel.PolicyStoreUnknown,
				Module:   " book ",
				PType:    " p ",
				V0:       " book ",
				V1:       " book ",
				V2:       " read ",
				Metadata: map[string]any{"managed": true},
			},
			{
				Store:    deletionmodel.PolicyStoreUnknown,
				Module:   "book",
				PType:    "p",
				V0:       "book",
				V1:       "book",
				V2:       "read",
				Metadata: map[string]any{"managed": true},
			},
		},
	}

	normalized := req.Normalize(deletionmodel.PolicyStoreCSV)
	if normalized.Module != "book" {
		t.Fatalf("normalized module = %q, want book", normalized.Module)
	}
	if normalized.Store != deletionmodel.PolicyStoreCSV {
		t.Fatalf("normalized store = %q, want csv", normalized.Store)
	}
	if !normalized.RequireManaged || !normalized.RequireValidation {
		t.Fatalf("expected managed/validation defaults to be enabled, got %#v", normalized)
	}
	if len(normalized.Selectors) != 1 {
		t.Fatalf("expected duplicate selectors to be deduplicated, got %#v", normalized.Selectors)
	}
	if err := normalized.ValidateRequest(); err != nil {
		t.Fatalf("ValidateRequest returned error: %v", err)
	}

	plan := deletionmodel.DeletePlan{
		Module:      "book",
		PolicyStore: deletionmodel.PolicyStoreCSV,
		PolicyChanges: []deletionmodel.DeleteItem{
			{Selector: &deletionmodel.PolicySelector{Store: deletionmodel.PolicyStoreCSV, Module: "book", Metadata: map[string]any{"managed": true}, PType: "p", V0: "book", V1: "book", V2: "read"}},
			{Selector: nil},
		},
	}
	cleanupReq := BuildPolicyCleanupRequest(plan)
	if cleanupReq.Store != deletionmodel.PolicyStoreCSV {
		t.Fatalf("cleanup request store = %q, want csv", cleanupReq.Store)
	}
	if len(cleanupReq.Selectors) != 1 {
		t.Fatalf("expected BuildPolicyCleanupRequest to skip nil selectors, got %#v", cleanupReq.Selectors)
	}
	if !cleanupReq.Refresh || !cleanupReq.RequireManaged || !cleanupReq.RequireValidation {
		t.Fatalf("expected cleanup request defaults to be enabled, got %#v", cleanupReq)
	}
}

func TestCSVPolicyStoreDeleteRemovesMatchedRuleAndValidates(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	policyPath := filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv")
	if err := os.MkdirAll(filepath.Dir(policyPath), 0o755); err != nil {
		t.Fatalf("mkdir policy dir: %v", err)
	}
	if err := os.WriteFile(policyPath, []byte(strings.Join([]string{
		"p, book, book, read",
		"p, role, role, read",
	}, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("write policy fixture: %v", err)
	}

	store, err := NewCSVPolicyStore(policyPath)
	if err != nil {
		t.Fatalf("NewCSVPolicyStore returned error: %v", err)
	}

	result, err := store.Delete(context.Background(), PolicyCleanupRequest{
		Module: "book",
		Selectors: []deletionmodel.PolicySelector{{
			Store:    deletionmodel.PolicyStoreCSV,
			Module:   "book",
			PType:    "p",
			V0:       "book",
			V1:       "book",
			V2:       "read",
			Metadata: map[string]any{"managed": true},
		}},
	})
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if !result.Verified {
		t.Fatalf("expected CSV delete to verify successfully, got %#v", result)
	}
	if len(result.Deleted) != 1 {
		t.Fatalf("deleted items = %d, want 1", len(result.Deleted))
	}
	content, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("read policy file: %v", err)
	}
	if strings.Contains(string(content), "book") {
		t.Fatalf("expected book policy rule to be removed, got %s", content)
	}
	if !strings.Contains(string(content), "role") {
		t.Fatalf("expected unrelated policy rule to remain, got %s", content)
	}
}

func TestDBPolicyStorePreviewDeleteAndValidate(t *testing.T) {
	t.Parallel()

	db := openPolicyCleanupSQLiteDB(t)
	if err := db.AutoMigrate(&dbPolicyRecord{}); err != nil {
		t.Fatalf("AutoMigrate db policy record: %v", err)
	}
	seedDBPolicyRule(t, db, "book", "book", "read")
	seedDBPolicyRule(t, db, "role", "role", "read")

	store, err := NewDBPolicyStore(db, &noopPolicyCacheRefresher{})
	if err != nil {
		t.Fatalf("NewDBPolicyStore returned error: %v", err)
	}

	preview, err := store.Preview(context.Background(), PolicyCleanupRequest{
		Module: "book",
		Selectors: []deletionmodel.PolicySelector{{
			Store:    deletionmodel.PolicyStoreDB,
			Module:   "book",
			PType:    "p",
			V0:       "book",
			V1:       "book",
			V2:       "read",
			Metadata: map[string]any{"managed": true},
		}},
	})
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}
	if preview.Store != deletionmodel.PolicyStoreDB {
		t.Fatalf("preview store = %q, want db", preview.Store)
	}
	if len(preview.Items) != 1 {
		t.Fatalf("preview items = %d, want 1", len(preview.Items))
	}
	if preview.Items[0].Decision != "delete" {
		t.Fatalf("preview decision = %q, want delete", preview.Items[0].Decision)
	}
	if preview.Summary.Deleted != 1 || preview.Summary.Selected != 1 {
		t.Fatalf("unexpected preview summary: %#v", preview.Summary)
	}

	result, err := store.Delete(context.Background(), PolicyCleanupRequest{
		Module: "book",
		Store:  deletionmodel.PolicyStoreDB,
		Selectors: []deletionmodel.PolicySelector{{
			Store:    deletionmodel.PolicyStoreDB,
			Module:   "book",
			PType:    "p",
			V0:       "book",
			V1:       "book",
			V2:       "read",
			Metadata: map[string]any{"managed": true},
		}},
		Refresh: true,
	})
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if !result.Verified {
		t.Fatalf("expected DB delete to verify successfully, got %#v", result)
	}
	if !result.Audit.CacheRefreshed {
		t.Fatalf("expected cache refresh to be recorded, got %#v", result.Audit)
	}
	if len(result.Deleted) != 1 {
		t.Fatalf("deleted items = %d, want 1", len(result.Deleted))
	}
	if err := store.Validate(context.Background()); err != nil {
		t.Fatalf("Validate returned error after delete: %v", err)
	}
	remaining, err := loadDBRules(context.Background(), db)
	if err != nil {
		t.Fatalf("loadDBRules returned error: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("remaining rules = %d, want 1", len(remaining))
	}
	if remaining[0].V0 != "role" {
		t.Fatalf("expected unrelated DB rule to remain, got %#v", remaining[0])
	}
}

func TestDBPolicyStoreDeleteRollsBackOnTransactionFailure(t *testing.T) {
	t.Parallel()

	db := openPolicyCleanupSQLiteDB(t)
	if err := db.AutoMigrate(&dbPolicyRecord{}); err != nil {
		t.Fatalf("AutoMigrate db policy record: %v", err)
	}
	seedDBPolicyRule(t, db, "book", "book", "read")
	seedDBPolicyRule(t, db, "role", "role", "read")

	store, err := NewDBPolicyStore(db, nil)
	if err != nil {
		t.Fatalf("NewDBPolicyStore returned error: %v", err)
	}
	cbName := "policy_cleanup_test:force_delete_failure"
	db.Callback().Delete().Before("gorm:delete").Register(cbName, func(tx *gorm.DB) {
		if tx.Statement != nil && tx.Statement.Table == "casbin_rule" {
			tx.AddError(errors.New("forced delete failure"))
		}
	})

	result, err := store.Delete(context.Background(), PolicyCleanupRequest{
		Module: "book",
		Store:  deletionmodel.PolicyStoreDB,
		Selectors: []deletionmodel.PolicySelector{{
			Store:    deletionmodel.PolicyStoreDB,
			Module:   "book",
			PType:    "p",
			V0:       "book",
			V1:       "book",
			V2:       "read",
			Metadata: map[string]any{"managed": true},
		}},
	})
	if err == nil {
		t.Fatal("expected Delete to fail and roll back")
	}
	if result.Audit.Validation != "rolled_back" {
		t.Fatalf("audit validation = %q, want rolled_back", result.Audit.Validation)
	}
	remaining, err := loadDBRules(context.Background(), db)
	if err != nil {
		t.Fatalf("loadDBRules returned error: %v", err)
	}
	if len(remaining) != 2 {
		t.Fatalf("remaining rules = %d, want 2 after rollback", len(remaining))
	}
	if !containsDBRule(remaining, "book", "book", "read") || !containsDBRule(remaining, "role", "role", "read") {
		t.Fatalf("expected both DB policy rows to remain after rollback, got %#v", remaining)
	}
}

func openPolicyCleanupSQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "policy_cleanup.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	return db
}

func seedDBPolicyRule(t *testing.T, db *gorm.DB, v0, v1, v2 string) {
	t.Helper()
	if err := db.Create(&dbPolicyRecord{PType: "p", V0: v0, V1: v1, V2: v2}).Error; err != nil {
		t.Fatalf("seed db policy rule: %v", err)
	}
}

func containsDBRule(rules []deletionmodel.PolicyAsset, v0, v1, v2 string) bool {
	for _, rule := range rules {
		if rule.V0 == v0 && rule.V1 == v1 && rule.V2 == v2 {
			return true
		}
	}
	return false
}

type noopPolicyCacheRefresher struct{}

func (noopPolicyCacheRefresher) Reload() error { return nil }

package adapter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// GormPolicyStore describes the database-facing behavior required by the adapter.
// A concrete Gorm-backed repository can satisfy this interface without coupling
// the casbin package to a specific ORM implementation.
type GormPolicyStore interface {
	LoadPolicies(ctx context.Context) ([]Rule, error)
	SavePolicies(ctx context.Context, rules []Rule) error
}

type GormPolicyAdapter struct {
	store GormPolicyStore
}

func NewGormPolicyAdapter(store GormPolicyStore) (*GormPolicyAdapter, error) {
	if store == nil {
		return nil, fmt.Errorf("gorm policy store is required")
	}
	return &GormPolicyAdapter{store: store}, nil
}

func (a *GormPolicyAdapter) LoadRules() ([]Rule, error) {
	if a == nil || a.store == nil {
		return nil, fmt.Errorf("gorm policy adapter is not configured")
	}
	return a.store.LoadPolicies(context.Background())
}

func (a *GormPolicyAdapter) SaveRules(rules []Rule) error {
	if a == nil || a.store == nil {
		return fmt.Errorf("gorm policy adapter is not configured")
	}
	return a.store.SavePolicies(context.Background(), rules)
}

type GormStore interface {
	LoadModel(ctx context.Context, name string) (string, error)
	SaveModel(ctx context.Context, name, content string) error
	LoadPolicies(ctx context.Context) ([]Rule, error)
	SavePolicies(ctx context.Context, rules []Rule) error
}

type GormCasbinStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) (*GormCasbinStore, error) {
	if db == nil {
		return nil, fmt.Errorf("gorm casbin store requires db")
	}
	return &GormCasbinStore{db: db}, nil
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("gorm casbin migrate requires db")
	}
	return db.AutoMigrate(&casbinModelRecord{}, &casbinPolicyRecord{})
}

func (s *GormCasbinStore) LoadModel(ctx context.Context, name string) (string, error) {
	if s == nil || s.db == nil {
		return "", fmt.Errorf("gorm casbin store is not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("casbin model name is required")
	}
	var record casbinModelRecord
	if err := s.db.WithContext(ctx).First(&record, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("load casbin model: %w", err)
	}
	return record.Content, nil
}

func (s *GormCasbinStore) SaveModel(ctx context.Context, name, content string) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("gorm casbin store is not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("casbin model name is required")
	}
	record := casbinModelRecord{Name: name, Content: content, UpdatedAt: time.Now().UTC()}
	if err := s.db.WithContext(ctx).Save(&record).Error; err != nil {
		return fmt.Errorf("save casbin model: %w", err)
	}
	return nil
}

func (s *GormCasbinStore) LoadPolicies(ctx context.Context) ([]Rule, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("gorm casbin store is not configured")
	}
	var records []casbinPolicyRecord
	if err := s.db.WithContext(ctx).Where("ptype = ?", "p").Order("id ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("load casbin policies: %w", err)
	}
	rules := make([]Rule, 0, len(records))
	for _, record := range records {
		if trimmed := record.toRule(); trimmed != (Rule{}) {
			rules = append(rules, trimmed)
		}
	}
	return rules, nil
}

func (s *GormCasbinStore) SavePolicies(ctx context.Context, rules []Rule) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("gorm casbin store is not configured")
	}
	desiredRules := normalizePolicyRules(rules)
	existingRules, err := s.LoadPolicies(ctx)
	if err != nil {
		return err
	}

	desired := make(map[string]Rule, len(desiredRules))
	for _, rule := range desiredRules {
		desired[policyRuleKey(rule)] = rule
	}

	existing := make(map[string]Rule, len(existingRules))
	for _, rule := range existingRules {
		if key := policyRuleKey(rule); key != "" {
			existing[key] = rule
		}
	}

	var toDelete []Rule
	for _, rule := range existingRules {
		if _, ok := desired[policyRuleKey(rule)]; !ok {
			toDelete = append(toDelete, rule)
		}
	}

	var toInsert []Rule
	for _, rule := range desiredRules {
		if _, ok := existing[policyRuleKey(rule)]; !ok {
			toInsert = append(toInsert, rule)
		}
	}

	if len(toDelete) == 0 && len(toInsert) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rule := range toDelete {
			if err := tx.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", rule.Subject, rule.Object, rule.Action).Delete(&casbinPolicyRecord{}).Error; err != nil {
				return fmt.Errorf("delete casbin policy: %w", err)
			}
		}
		for _, rule := range toInsert {
			record := casbinPolicyRecord{PType: "p", V0: rule.Subject, V1: rule.Object, V2: rule.Action}
			if err := tx.Create(&record).Error; err != nil {
				return fmt.Errorf("save casbin policy: %w", err)
			}
		}
		return nil
	})
}

func normalizePolicyRules(rules []Rule) []Rule {
	seen := make(map[string]struct{}, len(rules))
	normalized := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		normalizedRule, ok := normalizePolicyRule(rule)
		if !ok {
			continue
		}
		key := policyRuleKey(normalizedRule)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, normalizedRule)
	}
	return normalized
}

func normalizePolicyRule(rule Rule) (Rule, bool) {
	rule.Subject = strings.TrimSpace(rule.Subject)
	rule.Object = strings.TrimSpace(rule.Object)
	rule.Action = strings.TrimSpace(rule.Action)
	if rule.Subject == "" || rule.Object == "" || rule.Action == "" {
		return Rule{}, false
	}
	return rule, true
}

func policyRuleKey(rule Rule) string {
	subject := strings.TrimSpace(rule.Subject)
	object := strings.TrimSpace(rule.Object)
	action := strings.TrimSpace(rule.Action)
	if subject == "" || object == "" || action == "" {
		return ""
	}
	return subject + "\x1f" + object + "\x1f" + action
}

type casbinModelRecord struct {
	Name      string `gorm:"column:name;primaryKey"`
	Content   string `gorm:"column:content;type:longtext;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (casbinModelRecord) TableName() string { return "casbin_model" }

type casbinPolicyRecord struct {
	ID        uint   `gorm:"primaryKey"`
	PType     string `gorm:"column:ptype;type:varchar(32);not null;index:idx_casbin_rule,priority:1"`
	V0        string `gorm:"column:v0;type:varchar(191);index:idx_casbin_rule,priority:2"`
	V1        string `gorm:"column:v1;type:varchar(191);index:idx_casbin_rule,priority:3"`
	V2        string `gorm:"column:v2;type:varchar(191);index:idx_casbin_rule,priority:4"`
	V3        string `gorm:"column:v3;type:varchar(255)"`
	V4        string `gorm:"column:v4;type:varchar(255)"`
	V5        string `gorm:"column:v5;type:varchar(255)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (casbinPolicyRecord) TableName() string { return "casbin_rule" }

func (r casbinPolicyRecord) toRule() Rule {
	if strings.TrimSpace(r.V0) == "" || strings.TrimSpace(r.V1) == "" || strings.TrimSpace(r.V2) == "" {
		return Rule{}
	}
	return Rule{Subject: strings.TrimSpace(r.V0), Object: strings.TrimSpace(r.V1), Action: strings.TrimSpace(r.V2)}
}

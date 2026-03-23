package adapter

import (
	"context"
	"fmt"
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

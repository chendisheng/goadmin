package adapter

import (
	"context"
	"fmt"
	"strings"
)

type Rule struct {
	Subject string
	Object  string
	Action  string
}

type Adapter interface {
	LoadRules() ([]Rule, error)
	SaveRules([]Rule) error
}

type PolicyRepository interface {
	LoadPolicies(ctx context.Context) ([]Rule, error)
	SavePolicies(ctx context.Context, rules []Rule) error
}

type GormAdapter struct {
	repo PolicyRepository
}

func NewGormAdapter(repo PolicyRepository) (*GormAdapter, error) {
	if repo == nil {
		return nil, fmt.Errorf("policy repository is required")
	}
	return &GormAdapter{repo: repo}, nil
}

func (a *GormAdapter) LoadRules() ([]Rule, error) {
	if a == nil || a.repo == nil {
		return nil, fmt.Errorf("gorm adapter is not configured")
	}
	return a.repo.LoadPolicies(context.Background())
}

func (a *GormAdapter) SaveRules(rules []Rule) error {
	if a == nil || a.repo == nil {
		return fmt.Errorf("gorm adapter is not configured")
	}
	return a.repo.SavePolicies(context.Background(), rules)
}

func normalizeRuleLine(parts []string) (Rule, error) {
	if len(parts) == 4 && strings.TrimSpace(parts[0]) == "p" {
		return Rule{
			Subject: strings.TrimSpace(parts[1]),
			Object:  strings.TrimSpace(parts[2]),
			Action:  strings.TrimSpace(parts[3]),
		}, nil
	}
	if len(parts) == 3 {
		return Rule{
			Subject: strings.TrimSpace(parts[0]),
			Object:  strings.TrimSpace(parts[1]),
			Action:  strings.TrimSpace(parts[2]),
		}, nil
	}
	return Rule{}, fmt.Errorf("invalid policy row")
}

func formatRuleLine(rule Rule) string {
	return strings.TrimSpace(fmt.Sprintf("p, %s, %s, %s", rule.Subject, rule.Object, rule.Action))
}

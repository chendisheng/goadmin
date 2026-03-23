package enforcer

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	casbinadapter "goadmin/core/auth/casbin/adapter"
)

type Config struct {
	ModelPath string
	Adapter   casbinadapter.Adapter
}

type Enforcer struct {
	mu        sync.RWMutex
	modelPath string
	adapter   casbinadapter.Adapter
	rules     []casbinadapter.Rule
}

func (e *Enforcer) ModelPath() string {
	if e == nil {
		return ""
	}
	return e.modelPath
}

func New(cfg Config) (*Enforcer, error) {
	if strings.TrimSpace(cfg.ModelPath) == "" {
		return nil, fmt.Errorf("casbin model path is required")
	}
	if cfg.Adapter == nil {
		return nil, fmt.Errorf("casbin adapter is required")
	}
	if err := validateModel(cfg.ModelPath); err != nil {
		return nil, err
	}

	rules, err := cfg.Adapter.LoadRules()
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		rules = defaultRules()
		if err := cfg.Adapter.SaveRules(rules); err != nil {
			return nil, err
		}
	}

	return &Enforcer{
		modelPath: cfg.ModelPath,
		adapter:   cfg.Adapter,
		rules:     rules,
	}, nil
}

func (e *Enforcer) Reload() error {
	if e == nil {
		return nil
	}
	rules, err := e.adapter.LoadRules()
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.rules = rules
	e.mu.Unlock()
	return nil
}

func (e *Enforcer) Enforce(subject, object, action string) (bool, error) {
	if e == nil {
		return true, nil
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, rule := range e.rules {
		if matchSubject(subject, rule.Subject) && matchObject(object, rule.Object) && matchAction(action, rule.Action) {
			return true, nil
		}
	}
	return false, nil
}

func (e *Enforcer) AddPolicy(rule casbinadapter.Rule) error {
	if e == nil {
		return nil
	}
	if strings.TrimSpace(rule.Subject) == "" || strings.TrimSpace(rule.Object) == "" || strings.TrimSpace(rule.Action) == "" {
		return fmt.Errorf("policy fields are required")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, existing := range e.rules {
		if existing == rule {
			return nil
		}
	}
	e.rules = append(e.rules, rule)
	return e.adapter.SaveRules(e.rules)
}

func (e *Enforcer) RemovePolicy(rule casbinadapter.Rule) error {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	filtered := e.rules[:0]
	for _, existing := range e.rules {
		if existing == rule {
			continue
		}
		filtered = append(filtered, existing)
	}
	e.rules = append([]casbinadapter.Rule(nil), filtered...)
	return e.adapter.SaveRules(e.rules)
}

func validateModel(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read casbin model: %w", err)
	}
	text := string(data)
	for _, token := range []string{"[request_definition]", "[policy_definition]", "[policy_effect]", "[matchers]"} {
		if !strings.Contains(text, token) {
			return fmt.Errorf("invalid casbin model %s: missing %s", path, token)
		}
	}
	return nil
}

func defaultRules() []casbinadapter.Rule {
	return []casbinadapter.Rule{
		{Subject: "admin", Object: "/api/v1/auth/me", Action: "GET"},
		{Subject: "admin", Object: "/api/v1/auth/logout", Action: "POST"},
		{Subject: "user", Object: "/api/v1/auth/me", Action: "GET"},
		{Subject: "user", Object: "/api/v1/auth/logout", Action: "POST"},
	}
}

func matchSubject(subject, policy string) bool {
	if policy == "*" {
		return true
	}
	subject = strings.TrimSpace(subject)
	policy = strings.TrimSpace(policy)
	if strings.EqualFold(subject, policy) {
		return true
	}
	if strings.Contains(policy, "*") {
		ok, err := path.Match(policy, subject)
		return err == nil && ok
	}
	if looksLikeRegex(policy) {
		ok, err := regexp.MatchString(policy, subject)
		return err == nil && ok
	}
	return false
}

func matchAction(action, policy string) bool {
	if policy == "*" {
		return true
	}
	policy = strings.TrimSpace(policy)
	if looksLikeRegex(policy) {
		ok, err := regexp.MatchString(policy, strings.TrimSpace(action))
		return err == nil && ok
	}
	return strings.EqualFold(strings.TrimSpace(action), policy)
}

func matchObject(value, pattern string) bool {
	if pattern == "*" {
		return true
	}
	value = strings.TrimSpace(value)
	pattern = strings.TrimSpace(pattern)
	if value == pattern {
		return true
	}
	if strings.Contains(pattern, "*") {
		ok, err := path.Match(pattern, value)
		return err == nil && ok
	}
	return false
}

func looksLikeRegex(value string) bool {
	for _, ch := range []string{".", "+", "?", "[", "]", "(", ")", "|", "^", "$"} {
		if strings.Contains(value, ch) {
			return true
		}
	}
	return false
}

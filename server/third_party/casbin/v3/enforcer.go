package casbin

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
)

type SyncedEnforcer struct {
	mu         sync.RWMutex
	modelPath  string
	policyPath string
	policies   []policyRule
}

type policyRule struct {
	sub string
	obj string
	act string
}

func NewSyncedEnforcer(modelPath, policyPath string) (*SyncedEnforcer, error) {
	if strings.TrimSpace(modelPath) == "" {
		return nil, fmt.Errorf("model path is required")
	}
	if strings.TrimSpace(policyPath) == "" {
		return nil, fmt.Errorf("policy path is required")
	}
	if err := validateModel(modelPath); err != nil {
		return nil, err
	}
	e := &SyncedEnforcer{modelPath: modelPath, policyPath: policyPath}
	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *SyncedEnforcer) LoadPolicy() error {
	if e == nil {
		return fmt.Errorf("enforcer is nil")
	}
	rules, err := loadPolicyFile(e.policyPath)
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.policies = rules
	e.mu.Unlock()
	return nil
}

func (e *SyncedEnforcer) Enforce(params ...interface{}) (bool, error) {
	if e == nil {
		return false, fmt.Errorf("enforcer is nil")
	}
	if len(params) != 3 {
		return false, fmt.Errorf("casbin enforce expects 3 arguments")
	}
	sub, ok := params[0].(string)
	if !ok {
		return false, fmt.Errorf("subject must be a string")
	}
	obj, ok := params[1].(string)
	if !ok {
		return false, fmt.Errorf("object must be a string")
	}
	act, ok := params[2].(string)
	if !ok {
		return false, fmt.Errorf("action must be a string")
	}

	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, rule := range e.policies {
		if matchSubject(sub, rule.sub) && matchPath(obj, rule.obj) && matchAction(act, rule.act) {
			return true, nil
		}
	}
	return false, nil
}

func (e *SyncedEnforcer) SetAutoSave(_ bool) {}

func (e *SyncedEnforcer) AddPolicy(params ...interface{}) (bool, error) {
	if len(params) != 3 {
		return false, fmt.Errorf("casbin add policy expects 3 arguments")
	}
	rule := policyRule{}
	for i, raw := range params {
		value, ok := raw.(string)
		if !ok {
			return false, fmt.Errorf("policy argument %d must be a string", i)
		}
		switch i {
		case 0:
			rule.sub = value
		case 1:
			rule.obj = value
		case 2:
			rule.act = value
		}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, existing := range e.policies {
		if existing == rule {
			return false, nil
		}
	}
	e.policies = append(e.policies, rule)
	return true, nil
}

func validateModel(modelPath string) error {
	data, err := os.ReadFile(modelPath)
	if err != nil {
		return fmt.Errorf("read model file: %w", err)
	}
	text := string(data)
	required := []string{"[request_definition]", "[policy_definition]", "[policy_effect]", "[matchers]"}
	for _, token := range required {
		if !strings.Contains(text, token) {
			return fmt.Errorf("invalid model file %s: missing %s", modelPath, token)
		}
	}
	return nil
}

func loadPolicyFile(policyPath string) ([]policyRule, error) {
	file, err := os.Open(policyPath)
	if err != nil {
		return nil, fmt.Errorf("open policy file: %w", err)
	}
	defer file.Close()

	var rules []policyRule
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid policy line: %q", line)
		}
		if strings.TrimSpace(parts[0]) != "p" {
			return nil, fmt.Errorf("invalid policy prefix in line: %q", line)
		}
		rules = append(rules, policyRule{
			sub: strings.TrimSpace(parts[1]),
			obj: strings.TrimSpace(parts[2]),
			act: strings.TrimSpace(parts[3]),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read policy file: %w", err)
	}
	return rules, nil
}

func matchSubject(subject, policy string) bool {
	if policy == "*" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(subject), strings.TrimSpace(policy))
}

func matchAction(action, policy string) bool {
	if policy == "*" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(action), strings.TrimSpace(policy))
}

func matchPath(value, pattern string) bool {
	value = strings.TrimSpace(value)
	pattern = strings.TrimSpace(pattern)
	if pattern == "*" {
		return true
	}
	if value == pattern {
		return true
	}
	if strings.Contains(pattern, "*") {
		ok, err := path.Match(pattern, value)
		return err == nil && ok
	}
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(value, pattern)
	}
	return false
}

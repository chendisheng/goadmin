package adapter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileAdapter struct {
	path string
}

func NewFileAdapter(path string) (*FileAdapter, error) {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return nil, fmt.Errorf("casbin policy path is required")
	}
	return &FileAdapter{path: clean}, nil
}

func (a *FileAdapter) LoadRules() ([]Rule, error) {
	if a == nil {
		return nil, fmt.Errorf("file adapter is nil")
	}
	file, err := os.Open(a.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open casbin policy file: %w", err)
	}
	var rules []Rule
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		parts := splitLine(line)
		rule, err := normalizeRuleLine(parts)
		if err != nil {
			return nil, fmt.Errorf("invalid casbin policy line %q: %w", line, err)
		}
		rules = append(rules, rule)
	}
	if err := scanner.Err(); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("scan casbin policy file: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close casbin policy file: %w", err)
	}
	return rules, nil
}

func (a *FileAdapter) SaveRules(rules []Rule) error {
	if a == nil {
		return fmt.Errorf("file adapter is nil")
	}
	if err := os.MkdirAll(filepath.Dir(a.path), 0o755); err != nil {
		return fmt.Errorf("create casbin policy directory: %w", err)
	}
	tmp := a.path + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create casbin policy temp file: %w", err)
	}

	writer := bufio.NewWriter(file)
	for _, rule := range rules {
		if _, err := writer.WriteString(formatRuleLine(rule) + "\n"); err != nil {
			_ = file.Close()
			_ = os.Remove(tmp)
			return fmt.Errorf("write casbin policy file: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		_ = file.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("flush casbin policy file: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close casbin policy temp file: %w", err)
	}
	if err := os.Rename(tmp, a.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace casbin policy file: %w", err)
	}
	return nil
}

func splitLine(line string) []string {
	parts := strings.Split(line, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

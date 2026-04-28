package merger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Writer interface {
	Write(path string, content []byte, force bool) error
}

type FileWriter struct {
	Root string
}

func (w FileWriter) Write(path string, content []byte, force bool) error {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return fmt.Errorf("path is required")
	}
	if !filepath.IsAbs(clean) && strings.TrimSpace(w.Root) != "" {
		clean = filepath.Join(w.Root, clean)
	}
	if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", filepath.Dir(clean), err)
	}
	if _, err := os.Stat(clean); err == nil && !force {
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", clean, err)
	}
	if err := os.WriteFile(clean, content, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", clean, err)
	}
	return nil
}

func UniqueLines(lines []string) []string {
	seen := make(map[string]struct{}, len(lines))
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

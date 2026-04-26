package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadResourceRoots(roots ...string) error {
	return defaultRegistry.LoadResourceRoots(roots...)
}

func (r *Registry) LoadResourceRoots(roots ...string) error {
	if r == nil {
		return nil
	}
	for _, root := range roots {
		if err := r.LoadResourceRoot(root); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) LoadResourceRoot(root string) error {
	if r == nil {
		return nil
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return nil
	}
	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() || filepath.Base(path) != "locales" {
			return nil
		}
		return r.loadLocalesDirectory(path)
	})
}

func (r *Registry) loadLocalesDirectory(localesDir string) error {
	entries, err := os.ReadDir(localesDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		language := ResolveLanguage(entry.Name())
		if language == "" {
			continue
		}
		languageDir := filepath.Join(localesDir, entry.Name())
		if err := filepath.WalkDir(languageDir, func(path string, walkEntry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if walkEntry.IsDir() {
				return nil
			}
			if !isLocaleFile(path) {
				return nil
			}
			return r.loadLocaleFile(language, path)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) loadLocaleFile(language, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read locale file %s: %w", path, err)
	}
	entries, err := parseLocaleEntries(path, data)
	if err != nil {
		return fmt.Errorf("parse locale file %s: %w", path, err)
	}
	r.Register(language, entries)
	return nil
}

func isLocaleFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

func parseLocaleEntries(path string, data []byte) (map[string]string, error) {
	var payload any
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, err
		}
	default:
		if err := yaml.Unmarshal(data, &payload); err != nil {
			return nil, err
		}
	}
	if payload == nil {
		return map[string]string{}, nil
	}
	entries := make(map[string]string)
	if rootMap, ok := payload.(map[string]any); ok {
		if translation, ok := rootMap["translation"]; ok && len(rootMap) == 1 {
			flattenLocaleValue("", translation, entries)
			return entries, nil
		}
		flattenLocaleValue("", rootMap, entries)
		return entries, nil
	}
	if rootMap, ok := payload.(map[string]string); ok {
		for key, value := range rootMap {
			if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
				continue
			}
			entries[key] = value
		}
		return entries, nil
	}
	return nil, fmt.Errorf("unsupported locale payload type %T", payload)
}

func flattenLocaleValue(prefix string, value any, entries map[string]string) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			joined := key
			if prefix != "" {
				joined = prefix + "." + key
			}
			flattenLocaleValue(joined, child, entries)
		}
	case map[string]string:
		for key, child := range typed {
			joined := key
			if prefix != "" {
				joined = prefix + "." + key
			}
			if strings.TrimSpace(joined) == "" || strings.TrimSpace(child) == "" {
				continue
			}
			entries[joined] = child
		}
	case string:
		if strings.TrimSpace(prefix) == "" || strings.TrimSpace(typed) == "" {
			return
		}
		entries[prefix] = typed
	case fmt.Stringer:
		if strings.TrimSpace(prefix) == "" {
			return
		}
		value := strings.TrimSpace(typed.String())
		if value == "" {
			return
		}
		entries[prefix] = value
	default:
		// Ignore unsupported nested value types so resource files can carry
		// metadata only in translation entries without breaking the loader.
	}
}

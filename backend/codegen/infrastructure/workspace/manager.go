package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	baseDir string
}

func NewManager(baseDir string) *Manager {
	clean := filepath.Clean(strings.TrimSpace(baseDir))
	if clean == "" || clean == "." {
		clean = filepath.Join(os.TempDir(), "goadmin", "codegen")
	}
	return &Manager{baseDir: clean}
}

func (m *Manager) Create(taskID string) (string, error) {
	if strings.TrimSpace(taskID) == "" {
		return "", fmt.Errorf("task id is required")
	}
	path := filepath.Join(m.baseDir, "workspaces", taskID)
	if err := os.RemoveAll(path); err != nil {
		return "", fmt.Errorf("reset workspace %s: %w", path, err)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create workspace %s: %w", path, err)
	}
	return path, nil
}

func (m *Manager) Remove(taskID string) error {
	if strings.TrimSpace(taskID) == "" {
		return nil
	}
	return os.RemoveAll(filepath.Join(m.baseDir, "workspaces", taskID))
}

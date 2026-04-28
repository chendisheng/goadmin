package artifact

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrArtifactExpired = errors.New("artifact expired")

type Metadata struct {
	TaskID      string    `json:"task_id"`
	Filename    string    `json:"filename"`
	PackagePath string    `json:"package_path"`
	SizeBytes   int64     `json:"size_bytes"`
	FileCount   int       `json:"file_count"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) *LocalStorage {
	clean := filepath.Clean(strings.TrimSpace(baseDir))
	if clean == "" || clean == "." {
		clean = filepath.Join(os.TempDir(), "goadmin", "codegen")
	}
	return &LocalStorage{baseDir: clean}
}

func (s *LocalStorage) Save(taskID string, filename string, sourcePath string, sizeBytes int64, fileCount int, expiresAt time.Time) (Metadata, error) {
	if strings.TrimSpace(taskID) == "" {
		return Metadata{}, fmt.Errorf("task id is required")
	}
	if strings.TrimSpace(filename) == "" {
		return Metadata{}, fmt.Errorf("filename is required")
	}
	if strings.TrimSpace(sourcePath) == "" {
		return Metadata{}, fmt.Errorf("source path is required")
	}
	artifactsDir := filepath.Join(s.baseDir, "artifacts")
	if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
		return Metadata{}, fmt.Errorf("create artifacts directory: %w", err)
	}
	packagePath := filepath.Join(artifactsDir, taskID+".zip")
	if err := copyFile(sourcePath, packagePath); err != nil {
		return Metadata{}, err
	}
	meta := Metadata{
		TaskID:      taskID,
		Filename:    filename,
		PackagePath: packagePath,
		SizeBytes:   sizeBytes,
		FileCount:   fileCount,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   expiresAt.UTC(),
	}
	if err := s.writeMetadata(meta); err != nil {
		return Metadata{}, err
	}
	return meta, nil
}

func (s *LocalStorage) Load(taskID string) (Metadata, error) {
	meta, err := s.readMetadata(taskID)
	if err != nil {
		return Metadata{}, err
	}
	if !meta.ExpiresAt.IsZero() && time.Now().UTC().After(meta.ExpiresAt) {
		_ = s.Remove(taskID)
		return Metadata{}, ErrArtifactExpired
	}
	if _, err := os.Stat(meta.PackagePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Metadata{}, os.ErrNotExist
		}
		return Metadata{}, fmt.Errorf("stat package %s: %w", meta.PackagePath, err)
	}
	return meta, nil
}

func (s *LocalStorage) Remove(taskID string) error {
	var errs []error
	if err := os.Remove(filepath.Join(s.baseDir, "artifacts", taskID+".zip")); err != nil && !errors.Is(err, os.ErrNotExist) {
		errs = append(errs, err)
	}
	if err := os.Remove(filepath.Join(s.baseDir, "artifacts", taskID+".json")); err != nil && !errors.Is(err, os.ErrNotExist) {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (s *LocalStorage) CleanupExpired(now time.Time) error {
	artifactsDir := filepath.Join(s.baseDir, "artifacts")
	entries, err := os.ReadDir(artifactsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read artifacts directory: %w", err)
	}
	var errs []error
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		taskID := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		meta, err := s.readMetadata(taskID)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if meta.ExpiresAt.IsZero() || meta.ExpiresAt.After(now.UTC()) {
			continue
		}
		if err := s.Remove(taskID); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *LocalStorage) readMetadata(taskID string) (Metadata, error) {
	path := filepath.Join(s.baseDir, "artifacts", taskID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Metadata{}, os.ErrNotExist
		}
		return Metadata{}, fmt.Errorf("read metadata %s: %w", path, err)
	}
	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return Metadata{}, fmt.Errorf("unmarshal metadata %s: %w", path, err)
	}
	return meta, nil
}

func (s *LocalStorage) writeMetadata(meta Metadata) error {
	path := filepath.Join(s.baseDir, "artifacts", meta.TaskID+".json")
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal artifact metadata: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write metadata %s: %w", path, err)
	}
	return nil
}

func copyFile(sourcePath string, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open source file %s: %w", sourcePath, err)
	}
	defer source.Close()
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create target directory %s: %w", filepath.Dir(targetPath), err)
	}
	target, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("create target file %s: %w", targetPath, err)
	}
	defer target.Close()
	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("copy file %s -> %s: %w", sourcePath, targetPath, err)
	}
	return nil
}

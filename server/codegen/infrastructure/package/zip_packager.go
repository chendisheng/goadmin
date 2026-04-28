package packager

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ZIPPackager struct{}

func NewZIPPackager() *ZIPPackager {
	return &ZIPPackager{}
}

func (p *ZIPPackager) Package(sourceDir string, outputPath string) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return 0, fmt.Errorf("create package directory: %w", err)
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return 0, fmt.Errorf("create package file %s: %w", outputPath, err)
	}
	defer func() {
		_ = file.Close()
	}()
	writer := zip.NewWriter(file)
	if err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("build relative path for %s: %w", path, err)
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("build zip header for %s: %w", path, err)
		}
		header.Name = filepath.ToSlash(relPath)
		header.Method = zip.Deflate
		entry, err := writer.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("create zip entry %s: %w", relPath, err)
		}
		source, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open source file %s: %w", path, err)
		}
		if _, err := io.Copy(entry, source); err != nil {
			_ = source.Close()
			return fmt.Errorf("write zip entry %s: %w", relPath, err)
		}
		if err := source.Close(); err != nil {
			return fmt.Errorf("close source file %s: %w", path, err)
		}
		return nil
	}); err != nil {
		_ = writer.Close()
		return 0, err
	}
	if err := writer.Close(); err != nil {
		return 0, fmt.Errorf("close zip writer: %w", err)
	}
	stat, err := os.Stat(outputPath)
	if err != nil {
		return 0, fmt.Errorf("stat package file %s: %w", outputPath, err)
	}
	return stat.Size(), nil
}

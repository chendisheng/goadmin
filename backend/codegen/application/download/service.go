package download

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	codegencli "goadmin/codegen/driver/cli"
	artifactstore "goadmin/codegen/infrastructure/artifact"
	packager "goadmin/codegen/infrastructure/package"
	workspaceinfra "goadmin/codegen/infrastructure/workspace"
	apperrors "goadmin/core/errors"
)

type Dependencies struct {
	BaseDir string
	TTL     time.Duration
}

type Service struct {
	workspaces *workspaceinfra.Manager
	storage    *artifactstore.LocalStorage
	packager   *packager.ZIPPackager
	ttl        time.Duration
}

func NewService(deps Dependencies) *Service {
	ttl := deps.TTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &Service{
		workspaces: workspaceinfra.NewManager(deps.BaseDir),
		storage:    artifactstore.NewLocalStorage(deps.BaseDir),
		packager:   packager.NewZIPPackager(),
		ttl:        ttl,
	}
}

func (s *Service) Generate(req GenerateRequest) (ArtifactInfo, error) {
	if s == nil {
		return ArtifactInfo{}, apperrors.New(apperrors.CodeInternal, "download service is required")
	}
	if strings.TrimSpace(req.DSL) == "" {
		return ArtifactInfo{}, apperrors.New(apperrors.CodeBadRequest, "dsl is required")
	}
	if err := s.storage.CleanupExpired(time.Now().UTC()); err != nil {
		return ArtifactInfo{}, apperrors.Wrap(err, apperrors.CodeInternal, "cleanup expired artifacts")
	}
	taskID, err := newTaskID()
	if err != nil {
		return ArtifactInfo{}, apperrors.Wrap(err, apperrors.CodeInternal, "create download task id")
	}
	workspaceRoot, err := s.workspaces.Create(taskID)
	if err != nil {
		return ArtifactInfo{}, apperrors.Wrap(err, apperrors.CodeInternal, "create codegen workspace")
	}
	defer func() {
		_ = s.workspaces.Remove(taskID)
	}()
	report, err := codegencli.ExecuteDSLDocument(workspaceRoot, []byte(req.DSL), req.Force, false)
	if err != nil {
		return ArtifactInfo{}, err
	}
	generatedFiles, err := listFiles(workspaceRoot)
	if err != nil {
		return ArtifactInfo{}, apperrors.Wrap(err, apperrors.CodeInternal, "collect generated files")
	}
	if err := writeSupportFiles(workspaceRoot, req, report, generatedFiles); err != nil {
		return ArtifactInfo{}, apperrors.Wrap(err, apperrors.CodeInternal, "prepare download package")
	}
	filename := buildFilename(req.PackageName, report)
	tempPackage, err := os.CreateTemp("", taskID+"-*.zip")
	if err != nil {
		return ArtifactInfo{}, apperrors.Wrap(err, apperrors.CodeInternal, "create temporary package file")
	}
	tempPackagePath := tempPackage.Name()
	if err := tempPackage.Close(); err != nil {
		return ArtifactInfo{}, apperrors.Wrap(err, apperrors.CodeInternal, "close temporary package file")
	}
	defer func() {
		_ = os.Remove(tempPackagePath)
	}()
	sizeBytes, err := s.packager.Package(workspaceRoot, tempPackagePath)
	if err != nil {
		return ArtifactInfo{}, apperrors.Wrap(err, apperrors.CodeInternal, "package generated artifact")
	}
	expiresAt := time.Now().UTC().Add(s.ttl)
	meta, err := s.storage.Save(taskID, filename, tempPackagePath, sizeBytes, len(generatedFiles), expiresAt)
	if err != nil {
		return ArtifactInfo{}, apperrors.Wrap(err, apperrors.CodeInternal, "store generated artifact")
	}
	return ArtifactInfo{
		TaskID:      meta.TaskID,
		DownloadURL: "/api/v1/codegen/artifacts/" + meta.TaskID,
		Filename:    meta.Filename,
		SizeBytes:   meta.SizeBytes,
		FileCount:   meta.FileCount,
		ExpiresAt:   meta.ExpiresAt,
	}, nil
}

func (s *Service) Resolve(taskID string) (ResolvedArtifact, error) {
	if s == nil {
		return ResolvedArtifact{}, apperrors.New(apperrors.CodeInternal, "download service is required")
	}
	if strings.TrimSpace(taskID) == "" {
		return ResolvedArtifact{}, apperrors.New(apperrors.CodeBadRequest, "task id is required")
	}
	if err := s.storage.CleanupExpired(time.Now().UTC()); err != nil {
		return ResolvedArtifact{}, apperrors.Wrap(err, apperrors.CodeInternal, "cleanup expired artifacts")
	}
	meta, err := s.storage.Load(taskID)
	if err != nil {
		switch {
		case errors.Is(err, artifactstore.ErrArtifactExpired):
			return ResolvedArtifact{}, apperrors.New(apperrors.CodeGone, "artifact expired")
		case errors.Is(err, os.ErrNotExist):
			return ResolvedArtifact{}, apperrors.New(apperrors.CodeNotFound, "artifact not found")
		default:
			return ResolvedArtifact{}, apperrors.Wrap(err, apperrors.CodeInternal, "load artifact")
		}
	}
	return ResolvedArtifact{Filename: meta.Filename, PackagePath: meta.PackagePath}, nil
}

func writeSupportFiles(workspaceRoot string, req GenerateRequest, report codegencli.DSLExecutionReport, generatedFiles []string) error {
	if req.IncludeReadme {
		if err := os.WriteFile(filepath.Join(workspaceRoot, "README.md"), []byte(buildReadme(report, generatedFiles)), 0o644); err != nil {
			return fmt.Errorf("write README.md: %w", err)
		}
	}
	if req.IncludeDSL {
		if err := os.WriteFile(filepath.Join(workspaceRoot, "dsl.yaml"), []byte(strings.TrimSpace(req.DSL)+"\n"), 0o644); err != nil {
			return fmt.Errorf("write dsl.yaml: %w", err)
		}
	}
	if req.IncludeReport {
		reportPayload := struct {
			GeneratedAt time.Time                     `json:"generated_at"`
			FileCount   int                           `json:"file_count"`
			Files       []string                      `json:"files"`
			Report      codegencli.DSLExecutionReport `json:"report"`
		}{
			GeneratedAt: time.Now().UTC(),
			FileCount:   len(generatedFiles),
			Files:       generatedFiles,
			Report:      report,
		}
		data, err := json.MarshalIndent(reportPayload, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal generation-report.json: %w", err)
		}
		if err := os.WriteFile(filepath.Join(workspaceRoot, "generation-report.json"), append(data, '\n'), 0o644); err != nil {
			return fmt.Errorf("write generation-report.json: %w", err)
		}
	}
	changes := buildChanges(generatedFiles)
	if err := os.WriteFile(filepath.Join(workspaceRoot, "changes.txt"), []byte(changes), 0o644); err != nil {
		return fmt.Errorf("write changes.txt: %w", err)
	}
	return nil
}

func listFiles(root string) ([]string, error) {
	paths := make([]string, 0, 16)
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func buildReadme(report codegencli.DSLExecutionReport, generatedFiles []string) string {
	var builder strings.Builder
	builder.WriteString("# GoAdmin CodeGen 生成代码包\n\n")
	builder.WriteString("本包由 GoAdmin CodeGen 服务端下载能力生成。\n\n")
	builder.WriteString("## 生成摘要\n\n")
	builder.WriteString(fmt.Sprintf("- 资源数：%d\n", len(report.Items)))
	builder.WriteString(fmt.Sprintf("- 文件数：%d\n", len(generatedFiles)))
	builder.WriteString("\n## 使用说明\n\n")
	builder.WriteString("1. 解压压缩包到本地临时目录。\n")
	builder.WriteString("2. 先核对 generation-report.json 与 changes.txt。\n")
	builder.WriteString("3. 将 backend/ 下生成文件合并到本地项目对应目录。\n")
	builder.WriteString("4. 若存在同名文件，请先审查再决定是否覆盖。\n")
	builder.WriteString("5. 合并完成后执行 gofmt、测试与前端校验。\n")
	return builder.String()
}

func buildChanges(generatedFiles []string) string {
	var builder strings.Builder
	builder.WriteString("GoAdmin CodeGen 变更清单\n")
	builder.WriteString("\n新增文件：\n")
	if len(generatedFiles) == 0 {
		builder.WriteString("无\n")
		return builder.String()
	}
	for _, path := range generatedFiles {
		builder.WriteString("- ")
		builder.WriteString(path)
		builder.WriteString("\n")
	}
	return builder.String()
}

func buildFilename(packageName string, report codegencli.DSLExecutionReport) string {
	base := sanitizePackageName(packageName)
	if base == "" && len(report.Items) > 0 {
		base = sanitizePackageName(report.Items[0].Name)
	}
	if base == "" {
		base = "codegen-package"
	}
	return fmt.Sprintf("%s-%s.zip", base, time.Now().UTC().Format("20060102-150405"))
}

func sanitizePackageName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case r == '-', r == '_', r == ' ', r == '.':
			if builder.Len() == 0 || lastDash {
				continue
			}
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func newTaskID() (string, error) {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("cgen_%s_%s", time.Now().UTC().Format("20060102_150405"), hex.EncodeToString(buf)), nil
}

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

	"gorm.io/gorm"
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
		return ArtifactInfo{}, apperrors.NewWithKey(apperrors.CodeInternal, "codegen.download.service_required", "download service is required")
	}
	if strings.TrimSpace(req.DSL) == "" {
		return ArtifactInfo{}, apperrors.NewWithKey(apperrors.CodeBadRequest, "codegen.download.dsl_required", "dsl is required")
	}
	if err := s.storage.CleanupExpired(time.Now().UTC()); err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.cleanup_expired_failed", "cleanup expired artifacts")
	}
	taskID, err := newTaskID()
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.create_task_id_failed", "create download task id")
	}
	workspaceRoot, err := s.workspaces.Create(taskID)
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.create_workspace_failed", "create codegen workspace")
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
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.collect_files_failed", "collect generated files")
	}
	if err := writeSupportFiles(workspaceRoot, req, report, generatedFiles); err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.prepare_package_failed", "prepare download package")
	}
	filename := buildFilename(req.PackageName, report)
	tempPackage, err := os.CreateTemp("", taskID+"-*.zip")
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.create_package_file_failed", "create temporary package file")
	}
	tempPackagePath := tempPackage.Name()
	if err := tempPackage.Close(); err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.close_package_file_failed", "close temporary package file")
	}
	defer func() {
		_ = os.Remove(tempPackagePath)
	}()
	sizeBytes, err := s.packager.Package(workspaceRoot, tempPackagePath)
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.package_artifact_failed", "package generated artifact")
	}
	expiresAt := time.Now().UTC().Add(s.ttl)
	meta, err := s.storage.Save(taskID, filename, tempPackagePath, sizeBytes, len(generatedFiles), expiresAt)
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.store_artifact_failed", "store generated artifact")
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

func (s *Service) GenerateDatabase(db *gorm.DB, req codegencli.DatabaseExecutionRequest) (ArtifactInfo, error) {
	if s == nil {
		return ArtifactInfo{}, apperrors.NewWithKey(apperrors.CodeInternal, "codegen.download.service_required", "download service is required")
	}
	if db == nil {
		return ArtifactInfo{}, apperrors.NewWithKey(apperrors.CodeBadRequest, "codegen.download.database_required", "database connection is required")
	}
	if err := req.Validate(); err != nil {
		return ArtifactInfo{}, err
	}
	if err := s.storage.CleanupExpired(time.Now().UTC()); err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.cleanup_expired_failed", "cleanup expired artifacts")
	}
	taskID, err := newTaskID()
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.create_task_id_failed", "create download task id")
	}
	workspaceRoot, err := s.workspaces.Create(taskID)
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.create_workspace_failed", "create codegen workspace")
	}
	defer func() {
		_ = s.workspaces.Remove(taskID)
	}()
	report, err := codegencli.ExecuteDatabaseDocument(workspaceRoot, db, nil, req, false)
	if err != nil {
		return ArtifactInfo{}, err
	}
	generatedFiles, err := listFiles(workspaceRoot)
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.collect_files_failed", "collect generated files")
	}
	if err := writeDatabaseSupportFiles(workspaceRoot, req, report, generatedFiles); err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.prepare_package_failed", "prepare download package")
	}
	filename := buildDatabaseFilename(req, report)
	tempPackage, err := os.CreateTemp("", taskID+"-*.zip")
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.create_package_file_failed", "create temporary package file")
	}
	tempPackagePath := tempPackage.Name()
	if err := tempPackage.Close(); err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.close_package_file_failed", "close temporary package file")
	}
	defer func() {
		_ = os.Remove(tempPackagePath)
	}()
	sizeBytes, err := s.packager.Package(workspaceRoot, tempPackagePath)
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.package_artifact_failed", "package generated artifact")
	}
	expiresAt := time.Now().UTC().Add(s.ttl)
	meta, err := s.storage.Save(taskID, filename, tempPackagePath, sizeBytes, len(generatedFiles), expiresAt)
	if err != nil {
		return ArtifactInfo{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.store_artifact_failed", "store generated artifact")
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
		return ResolvedArtifact{}, apperrors.NewWithKey(apperrors.CodeInternal, "codegen.download.service_required", "download service is required")
	}
	if strings.TrimSpace(taskID) == "" {
		return ResolvedArtifact{}, apperrors.NewWithKey(apperrors.CodeBadRequest, "codegen.download.task_id_required", "task id is required")
	}
	if err := s.storage.CleanupExpired(time.Now().UTC()); err != nil {
		return ResolvedArtifact{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.cleanup_expired_failed", "cleanup expired artifacts")
	}
	meta, err := s.storage.Load(taskID)
	if err != nil {
		switch {
		case errors.Is(err, artifactstore.ErrArtifactExpired):
			return ResolvedArtifact{}, apperrors.NewWithKey(apperrors.CodeGone, "codegen.download.artifact_expired", "artifact expired")
		case errors.Is(err, os.ErrNotExist):
			return ResolvedArtifact{}, apperrors.NewWithKey(apperrors.CodeNotFound, "codegen.download.artifact_not_found", "artifact not found")
		default:
			return ResolvedArtifact{}, apperrors.WrapWithKey(err, apperrors.CodeInternal, "codegen.download.load_artifact_failed", "load artifact")
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

func writeDatabaseSupportFiles(workspaceRoot string, req codegencli.DatabaseExecutionRequest, report codegencli.DatabasePreviewReport, generatedFiles []string) error {
	readme := buildDatabaseReadme(report, generatedFiles)
	if mountParentPath := strings.TrimSpace(req.MountParentPath); mountParentPath != "" {
		readme += "\n## 挂载目录\n\n"
		readme += "- " + mountParentPath + "\n"
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "README.md"), []byte(readme), 0o644); err != nil {
		return fmt.Errorf("write README.md: %w", err)
	}
	sanitizedRequest := struct {
		Driver           string   `json:"driver"`
		Database         string   `json:"database"`
		Schema           string   `json:"schema,omitempty"`
		Tables           []string `json:"tables,omitempty"`
		Force            bool     `json:"force,omitempty"`
		GenerateFrontend bool     `json:"generate_frontend,omitempty"`
		GeneratePolicy   bool     `json:"generate_policy,omitempty"`
		MountParentPath  string   `json:"mount_parent_path,omitempty"`
	}{
		Driver:           strings.TrimSpace(req.Driver),
		Database:         strings.TrimSpace(req.Database),
		Schema:           strings.TrimSpace(req.Schema),
		Tables:           append([]string(nil), req.Tables...),
		Force:            req.Force,
		GenerateFrontend: boolPtrValue(req.GenerateFrontend, true),
		GeneratePolicy:   boolPtrValue(req.GeneratePolicy, true),
		MountParentPath:  strings.TrimSpace(req.MountParentPath),
	}
	requestData, err := json.MarshalIndent(sanitizedRequest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal database-request.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "database-request.json"), append(requestData, '\n'), 0o644); err != nil {
		return fmt.Errorf("write database-request.json: %w", err)
	}
	sanitizedReport := report
	sanitizedReport.Audit.Input.ProjectRoot = ""
	reportData, err := json.MarshalIndent(sanitizedReport, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal database-report.json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceRoot, "database-report.json"), append(reportData, '\n'), 0o644); err != nil {
		return fmt.Errorf("write database-report.json: %w", err)
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

func buildDatabaseFilename(req codegencli.DatabaseExecutionRequest, report codegencli.DatabasePreviewReport) string {
	base := sanitizePackageName(req.Database)
	if base == "" && len(req.Tables) > 0 {
		base = sanitizePackageName(req.Tables[0])
	}
	if base == "" && len(report.Resources) > 0 {
		base = sanitizePackageName(report.Resources[0].Name)
	}
	if base == "" {
		base = "codegen-database-package"
	}
	return fmt.Sprintf("%s-%s.zip", base, time.Now().UTC().Format("20060102-150405"))
}

func buildDatabaseReadme(report codegencli.DatabasePreviewReport, generatedFiles []string) string {
	var builder strings.Builder
	builder.WriteString("# GoAdmin CodeGen 数据库生成代码包\n\n")
	builder.WriteString("本包由 GoAdmin CodeGen 数据库驱动下载能力生成。\n\n")
	builder.WriteString("## 生成摘要\n\n")
	builder.WriteString(fmt.Sprintf("- 驱动：%s\n", report.Source.Driver))
	builder.WriteString(fmt.Sprintf("- 数据库：%s\n", report.Source.Database))
	if report.Source.Schema != "" {
		builder.WriteString(fmt.Sprintf("- Schema：%s\n", report.Source.Schema))
	}
	builder.WriteString(fmt.Sprintf("- 表数：%d\n", len(report.Source.Tables)))
	builder.WriteString(fmt.Sprintf("- 资源数：%d\n", len(report.Resources)))
	builder.WriteString(fmt.Sprintf("- 文件数：%d\n", len(generatedFiles)))
	if len(report.Source.Tables) > 0 {
		builder.WriteString("\n## 表范围\n\n")
		for _, tableName := range report.Source.Tables {
			builder.WriteString("- ")
			builder.WriteString(tableName)
			builder.WriteString("\n")
		}
	}
	builder.WriteString("\n## 使用说明\n\n")
	builder.WriteString("1. 解压压缩包到本地临时目录。\n")
	builder.WriteString("2. 先核对 database-report.json 与 changes.txt。\n")
	builder.WriteString("3. 将 backend/ 下生成文件合并到本地项目对应目录。\n")
	builder.WriteString("4. 若存在同名文件，请先审查再决定是否覆盖。\n")
	builder.WriteString("5. 合并完成后执行 gofmt、测试与前端校验。\n")
	return builder.String()
}

func boolPtrValue(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}

func sanitizePackageName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
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

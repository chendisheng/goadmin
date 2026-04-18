package http

import (
	stdhttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	deleteapp "goadmin/codegen/application/delete"
	downloadapp "goadmin/codegen/application/download"
	installapp "goadmin/codegen/application/install"
	irbuilderapp "goadmin/codegen/application/irbuilder"
	codegencli "goadmin/codegen/driver/cli"
	lifecycle "goadmin/codegen/model/lifecycle"
	apperrors "goadmin/core/errors"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	menuservice "goadmin/modules/menu/application/service"

	"gorm.io/gorm"
)

type Dependencies struct {
	ProjectRoot     string
	PolicyStore     string
	DB              *gorm.DB
	ArtifactEnabled bool
	ArtifactBaseDir string
	ArtifactTTL     time.Duration
	MenuService     *menuservice.Service
}

type Handler struct {
	projectRoot     string
	db              *gorm.DB
	artifactEnabled bool
	downloads       *downloadapp.Service
	dbgen           *irbuilderapp.Service
	installer       *installapp.Service
	deletion        *deleteapp.Service
}

type DSLRequest struct {
	DSL   string `json:"dsl"`
	Force bool   `json:"force,omitempty"`
}

type GenerateDownloadRequest struct {
	DSL           string `json:"dsl"`
	Force         bool   `json:"force,omitempty"`
	PackageName   string `json:"package_name,omitempty"`
	IncludeReadme *bool  `json:"include_readme,omitempty"`
	IncludeReport *bool  `json:"include_report,omitempty"`
	IncludeDSL    *bool  `json:"include_dsl,omitempty"`
}

type DatabaseRequest struct {
	Driver           string   `json:"driver"`
	Database         string   `json:"database"`
	Schema           string   `json:"schema,omitempty"`
	Tables           []string `json:"tables,omitempty"`
	Force            bool     `json:"force,omitempty"`
	GenerateFrontend *bool    `json:"generate_frontend,omitempty"`
	GeneratePolicy   *bool    `json:"generate_policy,omitempty"`
	MountParentPath  string   `json:"mount_parent_path,omitempty"`
}

type InstallManifestRequest struct {
	ManifestPath string `json:"manifest_path,omitempty"`
	Module       string `json:"module,omitempty"`
}

func (req DatabaseRequest) toExecutionRequest() codegencli.DatabaseExecutionRequest {
	return codegencli.DatabaseExecutionRequest{
		Driver:           req.Driver,
		Database:         req.Database,
		Schema:           req.Schema,
		Tables:           append([]string(nil), req.Tables...),
		Force:            req.Force,
		GenerateFrontend: req.GenerateFrontend,
		GeneratePolicy:   req.GeneratePolicy,
		MountParentPath:  req.MountParentPath,
	}
}

func NewHandler(deps Dependencies) *Handler {
	var downloads *downloadapp.Service
	if deps.ArtifactEnabled {
		downloads = downloadapp.NewService(downloadapp.Dependencies{
			BaseDir: deps.ArtifactBaseDir,
			TTL:     deps.ArtifactTTL,
		})
	}
	var installer *installapp.Service
	if deps.MenuService != nil {
		installer = installapp.NewService(installapp.Dependencies{MenuService: deps.MenuService})
	}
	var policyCleanup *deleteapp.PolicyCleanupService
	if cleanup, err := deleteapp.NewPolicyCleanupService(deleteapp.PolicyCleanupDependencies{
		ProjectRoot: deps.ProjectRoot,
		Store:       lifecycle.NormalizePolicyStoreKind(deps.PolicyStore),
		DB:          deps.DB,
	}); err == nil {
		policyCleanup = cleanup
	}
	deletionService := deleteapp.NewService(deleteapp.Dependencies{
		ProjectRoot:   deps.ProjectRoot,
		BackendRoot:   filepath.Join(deps.ProjectRoot, "backend"),
		PolicyStore:   deps.PolicyStore,
		MenuService:   deps.MenuService,
		PolicyCleanup: policyCleanup,
	})
	return &Handler{
		projectRoot:     strings.TrimSpace(deps.ProjectRoot),
		db:              deps.DB,
		artifactEnabled: deps.ArtifactEnabled,
		downloads:       downloads,
		dbgen:           irbuilderapp.NewService(irbuilderapp.Dependencies{}),
		installer:       installer,
		deletion:        deletionService,
	}
}

func (h *Handler) Preview(c coretransport.Context) {
	if err := h.ensureProjectRoot(); err != nil {
		h.writeError(c, err)
		return
	}
	var req DSLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "invalid request body"))
		return
	}
	if strings.TrimSpace(req.DSL) == "" {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "dsl is required"))
		return
	}
	report, err := codegencli.ExecuteDSLDocument(h.projectRoot, []byte(req.DSL), req.Force, true)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, response.Success(report, requestID(c)))
}

func (h *Handler) Generate(c coretransport.Context) {
	if err := h.ensureProjectRoot(); err != nil {
		h.writeError(c, err)
		return
	}
	var req DSLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "invalid request body"))
		return
	}
	if strings.TrimSpace(req.DSL) == "" {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "dsl is required"))
		return
	}
	report, err := codegencli.ExecuteDSLDocument(h.projectRoot, []byte(req.DSL), req.Force, false)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, response.Success(report, requestID(c)))
}

func (h *Handler) GenerateDownload(c coretransport.Context) {
	if err := h.ensureProjectRoot(); err != nil {
		h.writeError(c, err)
		return
	}
	if !h.artifactEnabled || h.downloads == nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "codegen artifact download is disabled"))
		return
	}
	var req GenerateDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "invalid request body"))
		return
	}
	if strings.TrimSpace(req.DSL) == "" {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "dsl is required"))
		return
	}
	artifact, err := h.downloads.Generate(downloadapp.GenerateRequest{
		DSL:           req.DSL,
		Force:         req.Force,
		PackageName:   req.PackageName,
		IncludeReadme: boolValue(req.IncludeReadme, true),
		IncludeReport: boolValue(req.IncludeReport, true),
		IncludeDSL:    boolValue(req.IncludeDSL, true),
	})
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, response.Success(artifact, requestID(c)))
}

func (h *Handler) InstallManifest(c coretransport.Context) {
	if err := h.ensureProjectRoot(); err != nil {
		h.writeError(c, err)
		return
	}
	if h.installer == nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "manifest install is disabled"))
		return
	}
	var req InstallManifestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "invalid request body"))
		return
	}
	manifestPath, err := h.resolveManifestPath(req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	result, err := h.installer.InstallManifest(c.RequestContext(), manifestPath)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, response.Success(result, requestID(c)))
}

func (h *Handler) GenerateDatabaseDownload(c coretransport.Context) {
	if err := h.ensureProjectRoot(); err != nil {
		h.writeError(c, err)
		return
	}
	if !h.artifactEnabled || h.downloads == nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "codegen artifact download is disabled"))
		return
	}
	var req DatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "invalid request body"))
		return
	}
	executionReq := req.toExecutionRequest()
	if err := executionReq.Validate(); err != nil {
		h.writeError(c, err)
		return
	}
	artifact, err := h.downloads.GenerateDatabase(h.db, executionReq)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, response.Success(artifact, requestID(c)))
}

func (h *Handler) PreviewDatabase(c coretransport.Context) {
	h.generateDatabase(c, true)
}

func (h *Handler) GenerateDatabase(c coretransport.Context) {
	h.generateDatabase(c, false)
}

func (h *Handler) PreviewDelete(c coretransport.Context) {
	if err := h.ensureProjectRoot(); err != nil {
		h.writeError(c, err)
		return
	}
	if h.deletion == nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "deletion preview is disabled"))
		return
	}
	var req lifecycle.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "invalid request body"))
		return
	}
	report, err := h.deletion.Preview(req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, response.Success(report, requestID(c)))
}

func (h *Handler) Delete(c coretransport.Context) {
	if err := h.ensureProjectRoot(); err != nil {
		h.writeError(c, err)
		return
	}
	if h.deletion == nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "deletion execution is disabled"))
		return
	}
	var req lifecycle.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "invalid request body"))
		return
	}
	result, err := h.deletion.Delete(req)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, response.Success(result, requestID(c)))
}

func (h *Handler) generateDatabase(c coretransport.Context, dryRun bool) {
	if err := h.ensureProjectRoot(); err != nil {
		h.writeError(c, err)
		return
	}
	var req DatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "invalid request body"))
		return
	}
	executionReq := req.toExecutionRequest()
	if err := executionReq.Validate(); err != nil {
		h.writeError(c, err)
		return
	}
	report, err := codegencli.ExecuteDatabaseDocument(h.projectRoot, h.db, h.dbgen, executionReq, dryRun)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, response.Success(report, requestID(c)))
}

func (h *Handler) DownloadArtifact(c coretransport.Context) {
	if !h.artifactEnabled || h.downloads == nil {
		h.writeError(c, apperrors.New(apperrors.CodeBadRequest, "codegen artifact download is disabled"))
		return
	}
	artifact, err := h.downloads.Resolve(c.Param("taskID"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.SetHeader("Cache-Control", "private, max-age=300")
	c.FileAttachment(artifact.PackagePath, artifact.Filename)
}

func (h *Handler) ensureProjectRoot() error {
	if strings.TrimSpace(h.projectRoot) == "" {
		return apperrors.New(apperrors.CodeBadRequest, "project root is required")
	}
	return nil
}

func (h *Handler) writeError(c coretransport.Context, err error) {
	status, body := response.Failure(err, requestID(c))

	c.JSON(status, body)
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get("request_id"); exists {
		if requestID, ok := value.(string); ok && strings.TrimSpace(requestID) != "" {
			return requestID
		}
	}
	return ""
}

func boolValue(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}

func (h *Handler) resolveManifestPath(req InstallManifestRequest) (string, error) {
	manifestPath := strings.TrimSpace(req.ManifestPath)
	if manifestPath != "" {
		if filepath.IsAbs(manifestPath) {
			return manifestPath, nil
		}
		return filepath.Join(h.projectRoot, manifestPath), nil
	}
	module := strings.TrimSpace(req.Module)
	if module == "" {
		return "", apperrors.New(apperrors.CodeBadRequest, "manifest path or module is required")
	}
	candidates := []string{
		filepath.Join(h.projectRoot, "backend", "modules", module, "manifest.yaml"),
		filepath.Join(h.projectRoot, "modules", module, "manifest.yaml"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", apperrors.New(apperrors.CodeNotFound, "manifest file not found")
}

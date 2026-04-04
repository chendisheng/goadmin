package http

import (
	stdhttp "net/http"
	"strings"
	"time"

	downloadapp "goadmin/codegen/application/download"
	irbuilderapp "goadmin/codegen/application/irbuilder"
	codegencli "goadmin/codegen/driver/cli"
	apperrors "goadmin/core/errors"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
)

type Dependencies struct {
	ProjectRoot     string
	ArtifactEnabled bool
	ArtifactBaseDir string
	ArtifactTTL     time.Duration
}

type Handler struct {
	projectRoot     string
	artifactEnabled bool
	downloads       *downloadapp.Service
	dbgen           *irbuilderapp.Service
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
	DSN              string   `json:"dsn"`
	Database         string   `json:"database"`
	Schema           string   `json:"schema,omitempty"`
	Tables           []string `json:"tables,omitempty"`
	Force            bool     `json:"force,omitempty"`
	GenerateFrontend *bool    `json:"generate_frontend,omitempty"`
	GeneratePolicy   *bool    `json:"generate_policy,omitempty"`
}

func (req DatabaseRequest) toExecutionRequest() codegencli.DatabaseExecutionRequest {
	return codegencli.DatabaseExecutionRequest{
		Driver:           req.Driver,
		DSN:              req.DSN,
		Database:         req.Database,
		Schema:           req.Schema,
		Tables:           append([]string(nil), req.Tables...),
		Force:            req.Force,
		GenerateFrontend: req.GenerateFrontend,
		GeneratePolicy:   req.GeneratePolicy,
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
	return &Handler{
		projectRoot:     strings.TrimSpace(deps.ProjectRoot),
		artifactEnabled: deps.ArtifactEnabled,
		downloads:       downloads,
		dbgen:           irbuilderapp.NewService(irbuilderapp.Dependencies{}),
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

func (h *Handler) PreviewDatabase(c coretransport.Context) {
	h.generateDatabase(c, true)
}

func (h *Handler) GenerateDatabase(c coretransport.Context) {
	h.generateDatabase(c, false)
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
	report, err := codegencli.ExecuteDatabaseDocument(h.projectRoot, h.dbgen, executionReq, dryRun)
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

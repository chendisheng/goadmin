package http

import (
	stdhttp "net/http"
	"strings"

	codegencli "goadmin/codegen/driver/cli"
	apperrors "goadmin/core/errors"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
)

type Dependencies struct {
	ProjectRoot string
}

type Handler struct {
	projectRoot string
}

type DSLRequest struct {
	DSL   string `json:"dsl"`
	Force bool   `json:"force,omitempty"`
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{projectRoot: strings.TrimSpace(deps.ProjectRoot)}
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

package handler

import (
	"net/http"

	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	casbinservice "goadmin/modules/casbin/application/service"

	"go.uber.org/zap"
)

type Handler struct {
	service *casbinservice.Service
	logger  *zap.Logger
}

func New(service *casbinservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Status(c coretransport.Context) {
	c.JSON(http.StatusOK, response.Success(h.service.Status(), requestID(c)))
}

func (h *Handler) Reload(c coretransport.Context) {
	if err := h.service.Reload(); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"reloaded": true}, requestID(c)))
}

func (h *Handler) Seed(c coretransport.Context) {
	if err := h.service.Seed(); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"seeded": true}, requestID(c)))
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

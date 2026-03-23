package handler

import (
	"net/http"
	"runtime"
	"time"

	"goadmin/core/config"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"

	"go.uber.org/zap"
)

type HealthHandler struct {
	cfg       *config.Config
	logger    *zap.Logger
	startedAt time.Time
}

func NewHealthHandler(cfg *config.Config, logger *zap.Logger, startedAt time.Time) *HealthHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &HealthHandler{cfg: cfg, logger: logger, startedAt: startedAt}
}

func (h *HealthHandler) Health(c coretransport.Context) {
	c.JSON(http.StatusOK, response.Success(map[string]any{
		"status":    "ok",
		"uptime":    time.Since(h.startedAt).String(),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, requestID(c)))
}

func (h *HealthHandler) Version(c coretransport.Context) {
	cfg := map[string]any{}
	if h.cfg != nil {
		cfg = map[string]any{
			"app": map[string]any{
				"name":    h.cfg.App.Name,
				"env":     h.cfg.App.Env,
				"version": h.cfg.App.Version,
			},
		}
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{
		"runtime": runtime.Version(),
		"config":  cfg,
	}, requestID(c)))
}

func (h *HealthHandler) Config(c coretransport.Context) {
	if h.cfg == nil {
		c.JSON(http.StatusOK, response.Success(map[string]any{}, requestID(c)))
		return
	}
	c.JSON(http.StatusOK, response.Success(h.cfg.Public(), requestID(c)))
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

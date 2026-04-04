package handler

import (
	"net/http"

	"go.uber.org/zap"
	coreerrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	"goadmin/modules/codegen_console/application/command"
	"goadmin/modules/codegen_console/application/query"
	codegen_consoleservice "goadmin/modules/codegen_console/application/service"
	"goadmin/modules/codegen_console/domain/model"
	codegen_consolereq "goadmin/modules/codegen_console/transport/http/request"
	codegen_consoleresp "goadmin/modules/codegen_console/transport/http/response"
)

type Handler struct {
	service *codegen_consoleservice.Service
	logger  *zap.Logger
}

func New(service *codegen_consoleservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c coretransport.Context) {
	var req codegen_consolereq.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.List(c.RequestContext(), query.Listcodegen_consoles{
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(codegen_consoleresp.List{
		Total: total,
		Items: mapItems(items),
	}, requestID(c)))
}

func (h *Handler) Get(c coretransport.Context) {
	item, err := h.service.Get(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Create(c coretransport.Context) {
	var req codegen_consolereq.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Create(c.RequestContext(), command.CreateCodegenConsole(req))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Update(c coretransport.Context) {
	var req codegen_consolereq.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Update(c.RequestContext(), c.Param("id"), command.UpdateCodegenConsole(req))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Delete(c coretransport.Context) {
	if err := h.service.Delete(c.RequestContext(), c.Param("id")); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"deleted": true}, requestID(c)))
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func mapItem(item model.CodegenConsole) codegen_consoleresp.Item {
	return codegen_consoleresp.Item{
		ID:      item.ID,
		Name:    item.Name,
		Enabled: item.Enabled,
	}
}

func mapItems(items []model.CodegenConsole) []codegen_consoleresp.Item {
	result := make([]codegen_consoleresp.Item, 0, len(items))
	for _, item := range items {
		result = append(result, mapItem(item))
	}
	return result
}

package handler

import (
	"net/http"

	"goadmin/core/errors"
	"goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	"goadmin/modules/role/application/command"
	"goadmin/modules/role/application/query"
	roleservice "goadmin/modules/role/application/service"
	"goadmin/modules/role/domain/model"
	rolereq "goadmin/modules/role/transport/http/request"
	roleresp "goadmin/modules/role/transport/http/response"

	"go.uber.org/zap"
)

type Handler struct {
	service *roleservice.Service
	logger  *zap.Logger
}

func New(service *roleservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c coretransport.Context) {
	var req rolereq.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(errors.Wrap(err, errors.CodeBadRequest, "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.List(c.RequestContext(), query.ListRoles{
		TenantID: req.TenantID,
		Keyword:  req.Keyword,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(roleresp.List{Total: total, Items: mapRoles(items)}, requestID(c)))
}

func (h *Handler) Get(c coretransport.Context) {
	item, err := h.service.Get(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapRole(*item), requestID(c)))
}

func (h *Handler) Create(c coretransport.Context) {
	var req rolereq.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(errors.Wrap(err, errors.CodeBadRequest, "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Create(c.RequestContext(), command.CreateRole{
		TenantID: req.TenantID,
		Name:     req.Name,
		Code:     req.Code,
		Status:   req.Status,
		Remark:   req.Remark,
		MenuIDs:  req.MenuIDs,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapRole(*item), requestID(c)))
}

func (h *Handler) Update(c coretransport.Context) {
	var req rolereq.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(errors.Wrap(err, errors.CodeBadRequest, "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Update(c.RequestContext(), c.Param("id"), command.UpdateRole{
		TenantID: req.TenantID,
		Name:     req.Name,
		Code:     req.Code,
		Status:   req.Status,
		Remark:   req.Remark,
		MenuIDs:  req.MenuIDs,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapRole(*item), requestID(c)))
}

func (h *Handler) Delete(c coretransport.Context) {
	if err := h.service.Delete(c.RequestContext(), c.Param("id")); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"deleted": true}, requestID(c)))
}

func mapRoles(items []model.Role) []roleresp.Item {
	result := make([]roleresp.Item, 0, len(items))
	for _, item := range items {
		result = append(result, mapRole(item))
	}
	return result
}

func mapRole(item model.Role) roleresp.Item {
	return roleresp.Item{
		ID:        item.ID,
		TenantID:  item.TenantID,
		Name:      item.Name,
		Code:      item.Code,
		Status:    string(item.Status),
		Remark:    item.Remark,
		MenuIDs:   append([]string(nil), item.MenuIDs...),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get(middleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

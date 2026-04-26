package handler

import (
	"net/http"

	apperrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	"goadmin/modules/user/application/command"
	"goadmin/modules/user/application/query"
	userservice "goadmin/modules/user/application/service"
	"goadmin/modules/user/domain/model"
	userhttpreq "goadmin/modules/user/transport/http/request"
	userhttpresp "goadmin/modules/user/transport/http/response"

	"go.uber.org/zap"
)

type Handler struct {
	service *userservice.Service
	logger  *zap.Logger
}

func New(service *userservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c coretransport.Context) {
	var req userhttpreq.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.List(c.RequestContext(), query.ListUsers{
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
	c.JSON(http.StatusOK, response.Success(userhttpresp.List{
		Total: total,
		Items: mapUsers(items),
	}, requestID(c)))
}

func (h *Handler) Get(c coretransport.Context) {
	item, err := h.service.Get(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapUser(*item), requestID(c)))
}

func (h *Handler) Create(c coretransport.Context) {
	var req userhttpreq.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Create(c.RequestContext(), command.CreateUser{
		TenantID:     req.TenantID,
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		Language:     req.Language,
		Mobile:       req.Mobile,
		Email:        req.Email,
		Status:       req.Status,
		RoleCodes:    req.RoleCodes,
		PasswordHash: req.PasswordHash,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapUser(*item), requestID(c)))
}

func (h *Handler) Update(c coretransport.Context) {
	var req userhttpreq.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Update(c.RequestContext(), c.Param("id"), command.UpdateUser{
		TenantID:     req.TenantID,
		Username:     req.Username,
		DisplayName:  req.DisplayName,
		Language:     req.Language,
		Mobile:       req.Mobile,
		Email:        req.Email,
		Status:       req.Status,
		RoleCodes:    req.RoleCodes,
		PasswordHash: req.PasswordHash,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapUser(*item), requestID(c)))
}

func (h *Handler) Delete(c coretransport.Context) {
	if err := h.service.Delete(c.RequestContext(), c.Param("id")); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"deleted": true}, requestID(c)))
}

func mapUsers(items []model.User) []userhttpresp.Item {
	result := make([]userhttpresp.Item, 0, len(items))
	for _, item := range items {
		result = append(result, mapUser(item))
	}
	return result
}

func mapUser(item model.User) userhttpresp.Item {
	return userhttpresp.Item{
		ID:          item.ID,
		TenantID:    item.TenantID,
		Username:    item.Username,
		DisplayName: item.DisplayName,
		Language:    item.Language,
		Mobile:      item.Mobile,
		Email:       item.Email,
		Status:      string(item.Status),
		RoleCodes:   append([]string(nil), item.RoleCodes...),
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

package handler

import (
	"net/http"

	coreerrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	"goadmin/modules/casbin_rule/application/command"
	"goadmin/modules/casbin_rule/application/query"
	casbin_ruleservice "goadmin/modules/casbin_rule/application/service"
	"goadmin/modules/casbin_rule/domain/model"
	casbin_rulereq "goadmin/modules/casbin_rule/transport/http/request"
	casbin_ruleresp "goadmin/modules/casbin_rule/transport/http/response"

	"go.uber.org/zap"
)

type Handler struct {
	service *casbin_ruleservice.Service
	logger  *zap.Logger
}

func New(service *casbin_ruleservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c coretransport.Context) {
	var req casbin_rulereq.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(coreerrors.WrapWithKey(err, coreerrors.CodeBadRequest, "casbin_rule.invalid_list_request", "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.List(c.RequestContext(), query.Listcasbin_rules{
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(casbin_ruleresp.List{
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
	var req casbin_rulereq.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(coreerrors.WrapWithKey(err, coreerrors.CodeBadRequest, "casbin_rule.invalid_create_request", "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Create(c.RequestContext(), command.CreateCasbinRule(req))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Update(c coretransport.Context) {
	var req casbin_rulereq.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(coreerrors.WrapWithKey(err, coreerrors.CodeBadRequest, "casbin_rule.invalid_update_request", "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Update(c.RequestContext(), c.Param("id"), command.UpdateCasbinRule(req))
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

func mapItem(item model.CasbinRule) casbin_ruleresp.Item {
	return casbin_ruleresp.Item{
		Id:        item.Id,
		Ptype:     item.Ptype,
		V0:        item.V0,
		V1:        item.V1,
		V2:        item.V2,
		V3:        item.V3,
		V4:        item.V4,
		V5:        item.V5,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func mapItems(items []model.CasbinRule) []casbin_ruleresp.Item {
	result := make([]casbin_ruleresp.Item, 0, len(items))
	for _, item := range items {
		result = append(result, mapItem(item))
	}
	return result
}

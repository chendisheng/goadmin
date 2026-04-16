package handler

import (
	"net/http"

	apperrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	"goadmin/modules/dictionary/application/command"
	"goadmin/modules/dictionary/application/query"
	dictionaryservice "goadmin/modules/dictionary/application/service"
	dictmodel "goadmin/modules/dictionary/domain/model"
	dictreq "goadmin/modules/dictionary/transport/http/request"
	dictresp "goadmin/modules/dictionary/transport/http/response"

	"go.uber.org/zap"
)

type Handler struct {
	service *dictionaryservice.Service
	logger  *zap.Logger
}

func New(service *dictionaryservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) ListCategories(c coretransport.Context) {
	var req dictreq.CategoryListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.ListCategories(c.RequestContext(), query.ListCategories{
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
	c.JSON(http.StatusOK, response.Success(dictresp.CategoryList{Total: total, Items: mapCategories(items)}, requestID(c)))
}

func (h *Handler) GetCategory(c coretransport.Context) {
	item, err := h.service.GetCategory(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapCategory(*item), requestID(c)))
}

func (h *Handler) CreateCategory(c coretransport.Context) {
	var req dictreq.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.CreateCategory(c.RequestContext(), command.CreateCategory{
		ID:          req.ID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Sort:        req.Sort,
		Remark:      req.Remark,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapCategory(*item), requestID(c)))
}

func (h *Handler) UpdateCategory(c coretransport.Context) {
	var req dictreq.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.UpdateCategory(c.RequestContext(), c.Param("id"), command.UpdateCategory{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Sort:        req.Sort,
		Remark:      req.Remark,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapCategory(*item), requestID(c)))
}

func (h *Handler) DeleteCategory(c coretransport.Context) {
	if err := h.service.DeleteCategory(c.RequestContext(), c.Param("id")); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"deleted": true}, requestID(c)))
}

func (h *Handler) ListItems(c coretransport.Context) {
	var req dictreq.ItemListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.ListItems(c.RequestContext(), query.ListItems{
		CategoryID:   req.CategoryID,
		CategoryCode: req.CategoryCode,
		Keyword:      req.Keyword,
		Status:       req.Status,
		Page:         req.Page,
		PageSize:     req.PageSize,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(dictresp.ItemList{Total: total, Items: mapItems(items)}, requestID(c)))
}

func (h *Handler) GetItem(c coretransport.Context) {
	item, err := h.service.GetItem(c.RequestContext(), c.Param("id"))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) CreateItem(c coretransport.Context) {
	var req dictreq.ItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.CreateItem(c.RequestContext(), command.CreateItem{
		ID:         req.ID,
		CategoryID: req.CategoryID,
		Value:      req.Value,
		Label:      req.Label,
		TagType:    req.TagType,
		TagColor:   req.TagColor,
		Extra:      req.Extra,
		IsDefault:  req.IsDefault,
		Status:     req.Status,
		Sort:       req.Sort,
		Remark:     req.Remark,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) UpdateItem(c coretransport.Context) {
	var req dictreq.ItemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.UpdateItem(c.RequestContext(), c.Param("id"), command.UpdateItem{
		CategoryID: req.CategoryID,
		Value:      req.Value,
		Label:      req.Label,
		TagType:    req.TagType,
		TagColor:   req.TagColor,
		Extra:      req.Extra,
		IsDefault:  req.IsDefault,
		Status:     req.Status,
		Sort:       req.Sort,
		Remark:     req.Remark,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) DeleteItem(c coretransport.Context) {
	if err := h.service.DeleteItem(c.RequestContext(), c.Param("id")); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"deleted": true}, requestID(c)))
}

func (h *Handler) LookupItems(c coretransport.Context) {
	items, err := h.service.LookupItems(c.RequestContext(), query.LookupItems{CategoryCode: c.Param("code")})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(dictresp.Lookup{Items: mapItems(items)}, requestID(c)))
}

func (h *Handler) LookupItem(c coretransport.Context) {
	item, err := h.service.LookupItem(c.RequestContext(), query.LookupItem{
		CategoryCode: c.Param("code"),
		Value:        c.Param("value"),
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapItem(*item), requestID(c)))
}

func mapCategories(items []dictmodel.Category) []dictresp.CategoryItem {
	result := make([]dictresp.CategoryItem, 0, len(items))
	for _, item := range items {
		result = append(result, mapCategory(item))
	}
	return result
}

func mapCategory(item dictmodel.Category) dictresp.CategoryItem {
	return dictresp.CategoryItem{
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		Status:      string(item.Status),
		Sort:        item.Sort,
		Remark:      item.Remark,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

func mapItems(items []dictmodel.Item) []dictresp.Item {
	result := make([]dictresp.Item, 0, len(items))
	for _, item := range items {
		result = append(result, mapItem(item))
	}
	return result
}

func mapItem(item dictmodel.Item) dictresp.Item {
	return dictresp.Item{
		ID:         item.ID,
		CategoryID: item.CategoryID,
		Value:      item.Value,
		Label:      item.Label,
		TagType:    item.TagType,
		TagColor:   item.TagColor,
		Extra:      item.Extra,
		IsDefault:  item.IsDefault,
		Status:     string(item.Status),
		Sort:       item.Sort,
		Remark:     item.Remark,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
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

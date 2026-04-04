package handler

import (
	"net/http"

	coreerrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	"goadmin/modules/book/application/command"
	"goadmin/modules/book/application/query"
	bookservice "goadmin/modules/book/application/service"
	"goadmin/modules/book/domain/model"
	bookreq "goadmin/modules/book/transport/http/request"
	bookresp "goadmin/modules/book/transport/http/response"

	"go.uber.org/zap"
)

type Handler struct {
	service *bookservice.Service
	logger  *zap.Logger
}

func New(service *bookservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c coretransport.Context) {
	var req bookreq.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.List(c.RequestContext(), query.Listbooks{
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(bookresp.List{
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
	var req bookreq.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Create(c.RequestContext(), command.CreateBook(req))
	if err != nil {
		h.logger.Error("create book failed",
			zap.String("request_id", requestID(c)),
			zap.String("tenant_id", req.TenantId),
			zap.String("title", req.Title),
			zap.Error(err),
		)
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Update(c coretransport.Context) {
	var req bookreq.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Update(c.RequestContext(), c.Param("id"), command.UpdateBook(req))
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

func mapItem(item model.Book) bookresp.Item {
	return bookresp.Item{
		Id:            item.Id,
		TenantId:      item.TenantId,
		Title:         item.Title,
		Author:        item.Author,
		Isbn:          item.Isbn,
		Publisher:     item.Publisher,
		PublishDate:   item.PublishDate,
		Category:      item.Category,
		Description:   item.Description,
		Status:        item.Status,
		Price:         item.Price,
		StockQuantity: item.StockQuantity,
		CoverImageUrl: item.CoverImageUrl,
		Tags:          item.Tags,
	}
}

func mapItems(items []model.Book) []bookresp.Item {
	result := make([]bookresp.Item, 0, len(items))
	for _, item := range items {
		result = append(result, mapItem(item))
	}
	return result
}

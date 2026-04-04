package handler

import (
	"net/http"

	"go.uber.org/zap"
	coreerrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	"goadmin/modules/order/application/command"
	"goadmin/modules/order/application/query"
	orderservice "goadmin/modules/order/application/service"
	"goadmin/modules/order/domain/model"
	orderreq "goadmin/modules/order/transport/http/request"
	orderresp "goadmin/modules/order/transport/http/response"
)

type Handler struct {
	service *orderservice.Service
	logger  *zap.Logger
}

func New(service *orderservice.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c coretransport.Context) {
	var req orderreq.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid list request"), requestID(c))
		c.JSON(status, body)
		return
	}
	items, total, err := h.service.List(c.RequestContext(), query.Listorders{
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(orderresp.List{
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
	var req orderreq.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid create request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Create(c.RequestContext(), command.CreateOrder(req))
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapItem(*item), requestID(c)))
}

func (h *Handler) Update(c coretransport.Context) {
	var req orderreq.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(coreerrors.Wrap(err, coreerrors.CodeBadRequest, "invalid update request"), requestID(c))
		c.JSON(status, body)
		return
	}
	item, err := h.service.Update(c.RequestContext(), c.Param("id"), command.UpdateOrder(req))
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

func mapItem(item model.Order) orderresp.Item {
	return orderresp.Item{
		Id:              item.Id,
		TenantId:        item.TenantId,
		OrderNo:         item.OrderNo,
		UserId:          item.UserId,
		CustomerName:    item.CustomerName,
		CustomerEmail:   item.CustomerEmail,
		CustomerPhone:   item.CustomerPhone,
		ShippingAddress: item.ShippingAddress,
		BillingAddress:  item.BillingAddress,
		OrderStatus:     item.OrderStatus,
		PaymentStatus:   item.PaymentStatus,
		PaymentMethod:   item.PaymentMethod,
		Currency:        item.Currency,
		TotalAmount:     item.TotalAmount,
		DiscountAmount:  item.DiscountAmount,
		TaxAmount:       item.TaxAmount,
		ShippingAmount:  item.ShippingAmount,
		FinalAmount:     item.FinalAmount,
		OrderDate:       item.OrderDate,
		ShippedDate:     item.ShippedDate,
		DeliveredDate:   item.DeliveredDate,
		Notes:           item.Notes,
		InternalNotes:   item.InternalNotes,
	}
}

func mapItems(items []model.Order) []orderresp.Item {
	result := make([]orderresp.Item, 0, len(items))
	for _, item := range items {
		result = append(result, mapItem(item))
	}
	return result
}

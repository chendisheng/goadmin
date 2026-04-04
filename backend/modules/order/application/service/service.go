package service

import (
	"context"
	"fmt"

	"goadmin/modules/order/application/command"
	"goadmin/modules/order/application/query"
	"goadmin/modules/order/domain/model"
	"goadmin/modules/order/domain/repository"
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("order repository is required")
	}
	return &Service{repo: repo}, nil
}

func (s *Service) List(ctx context.Context, q query.Listorders) ([]model.Order, int64, error) {
	if s == nil || s.repo == nil {
		return nil, 0, fmt.Errorf("order service is not configured")
	}
	return s.repo.List(ctx, q.Keyword, q.Page, q.PageSize)
}

func (s *Service) Get(ctx context.Context, id string) (*model.Order, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("order service is not configured")
	}
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input command.CreateOrder) (*model.Order, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("order service is not configured")
	}
	item := &model.Order{}
	item.TenantId = input.TenantId
	item.OrderNo = input.OrderNo
	item.UserId = input.UserId
	item.CustomerName = input.CustomerName
	item.CustomerEmail = input.CustomerEmail
	item.CustomerPhone = input.CustomerPhone
	item.ShippingAddress = input.ShippingAddress
	item.BillingAddress = input.BillingAddress
	item.OrderStatus = input.OrderStatus
	item.PaymentStatus = input.PaymentStatus
	item.PaymentMethod = input.PaymentMethod
	item.Currency = input.Currency
	item.TotalAmount = input.TotalAmount
	item.DiscountAmount = input.DiscountAmount
	item.TaxAmount = input.TaxAmount
	item.ShippingAmount = input.ShippingAmount
	item.FinalAmount = input.FinalAmount
	item.OrderDate = input.OrderDate
	item.ShippedDate = input.ShippedDate
	item.DeliveredDate = input.DeliveredDate
	item.Notes = input.Notes
	item.InternalNotes = input.InternalNotes
	return s.repo.Create(ctx, item)
}

func (s *Service) Update(ctx context.Context, id string, input command.UpdateOrder) (*model.Order, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("order service is not configured")
	}
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	cloned := item.Clone()
	item = &cloned
	item.TenantId = input.TenantId
	item.OrderNo = input.OrderNo
	item.UserId = input.UserId
	item.CustomerName = input.CustomerName
	item.CustomerEmail = input.CustomerEmail
	item.CustomerPhone = input.CustomerPhone
	item.ShippingAddress = input.ShippingAddress
	item.BillingAddress = input.BillingAddress
	item.OrderStatus = input.OrderStatus
	item.PaymentStatus = input.PaymentStatus
	item.PaymentMethod = input.PaymentMethod
	item.Currency = input.Currency
	item.TotalAmount = input.TotalAmount
	item.DiscountAmount = input.DiscountAmount
	item.TaxAmount = input.TaxAmount
	item.ShippingAmount = input.ShippingAmount
	item.FinalAmount = input.FinalAmount
	item.OrderDate = input.OrderDate
	item.ShippedDate = input.ShippedDate
	item.DeliveredDate = input.DeliveredDate
	item.Notes = input.Notes
	item.InternalNotes = input.InternalNotes
	return s.repo.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("order service is not configured")
	}
	return s.repo.Delete(ctx, id)
}

package repository

import (
	"context"
	"errors"

	"goadmin/modules/order/domain/model"
)

var ErrNotFound = errors.New("order not found")

type Repository interface {
	List(ctx context.Context, keyword string, page int, pageSize int) ([]model.Order, int64, error)
	Get(ctx context.Context, id string) (*model.Order, error)
	Create(ctx context.Context, item *model.Order) (*model.Order, error)
	Update(ctx context.Context, item *model.Order) (*model.Order, error)
	Delete(ctx context.Context, id string) error
}

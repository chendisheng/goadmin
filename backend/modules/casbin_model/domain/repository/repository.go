package repository

import (
	"context"
	"errors"

	"goadmin/modules/casbin_model/domain/model"
)

var ErrNotFound = errors.New("casbin_model not found")

type Repository interface {
	List(ctx context.Context, keyword string, page int, pageSize int) ([]model.CasbinModel, int64, error)
	Get(ctx context.Context, id string) (*model.CasbinModel, error)
	Create(ctx context.Context, item *model.CasbinModel) (*model.CasbinModel, error)
	Update(ctx context.Context, item *model.CasbinModel) (*model.CasbinModel, error)
	Delete(ctx context.Context, id string) error
}

package repository

import (
	"context"
	"errors"

	"goadmin/modules/casbin_rule/domain/model"
)

var ErrNotFound = errors.New("casbin_rule not found")

type Repository interface {
	List(ctx context.Context, keyword string, page int, pageSize int) ([]model.CasbinRule, int64, error)
	Get(ctx context.Context, id string) (*model.CasbinRule, error)
	Create(ctx context.Context, item *model.CasbinRule) (*model.CasbinRule, error)
	Update(ctx context.Context, item *model.CasbinRule) (*model.CasbinRule, error)
	Delete(ctx context.Context, id string) error
}

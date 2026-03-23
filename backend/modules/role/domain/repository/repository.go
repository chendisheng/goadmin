package repository

import (
	"context"
	"errors"

	"goadmin/modules/role/domain/model"
)

var (
	ErrNotFound = errors.New("role not found")
	ErrConflict = errors.New("role already exists")
)

type ListFilter struct {
	TenantID string
	Keyword  string
	Status   string
	Page     int
	PageSize int
}

type Repository interface {
	List(ctx context.Context, filter ListFilter) ([]model.Role, int64, error)
	Get(ctx context.Context, id string) (*model.Role, error)
	Create(ctx context.Context, role *model.Role) (*model.Role, error)
	Update(ctx context.Context, role *model.Role) (*model.Role, error)
	Delete(ctx context.Context, id string) error
}

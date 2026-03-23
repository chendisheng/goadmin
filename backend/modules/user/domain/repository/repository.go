package repository

import (
	"context"
	"errors"

	"goadmin/modules/user/domain/model"
)

var (
	ErrNotFound = errors.New("user not found")
	ErrConflict = errors.New("user already exists")
)

type ListFilter struct {
	TenantID string
	Keyword  string
	Status   string
	Page     int
	PageSize int
}

type Repository interface {
	List(ctx context.Context, filter ListFilter) ([]model.User, int64, error)
	Get(ctx context.Context, id string) (*model.User, error)
	Create(ctx context.Context, user *model.User) (*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	Delete(ctx context.Context, id string) error
}

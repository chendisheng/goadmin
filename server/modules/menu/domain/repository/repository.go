package repository

import (
	"context"
	"errors"

	"goadmin/modules/menu/domain/model"
)

var (
	ErrNotFound = errors.New("menu not found")
	ErrConflict = errors.New("menu already exists")
)

type ListFilter struct {
	Keyword  string
	Visible  *bool
	Enabled  *bool
	ParentID string
	Page     int
	PageSize int
}

type Repository interface {
	List(ctx context.Context, filter ListFilter) ([]model.Menu, int64, error)
	Get(ctx context.Context, id string) (*model.Menu, error)
	Create(ctx context.Context, menu *model.Menu) (*model.Menu, error)
	Update(ctx context.Context, menu *model.Menu) (*model.Menu, error)
	Delete(ctx context.Context, id string) error
	Tree(ctx context.Context) ([]model.Menu, error)
}

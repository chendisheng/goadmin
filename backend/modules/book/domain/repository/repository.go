package repository

import (
	"context"
	"errors"

	"goadmin/modules/book/domain/model"
)

var ErrNotFound = errors.New("book not found")

type Repository interface {
	List(ctx context.Context, keyword string, page int, pageSize int) ([]model.Book, int64, error)
	Get(ctx context.Context, id string) (*model.Book, error)
	Create(ctx context.Context, item *model.Book) (*model.Book, error)
	Update(ctx context.Context, item *model.Book) (*model.Book, error)
	Delete(ctx context.Context, id string) error
}

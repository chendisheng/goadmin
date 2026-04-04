package repository

import (
	"context"
	"errors"

	"goadmin/modules/codegen_console/domain/model"
)

var ErrNotFound = errors.New("codegen_console not found")

type Repository interface {
	List(ctx context.Context, keyword string, page int, pageSize int) ([]model.CodegenConsole, int64, error)
	Get(ctx context.Context, id string) (*model.CodegenConsole, error)
	Create(ctx context.Context, item *model.CodegenConsole) (*model.CodegenConsole, error)
	Update(ctx context.Context, item *model.CodegenConsole) (*model.CodegenConsole, error)
	Delete(ctx context.Context, id string) error
}

package repository

import (
	"context"
	"errors"

	"goadmin/modules/upload/domain/model"
)

var ErrNotFound = errors.New("upload file asset not found")

type ListFilter struct {
	Keyword    string
	Visibility string
	Status     string
	BizModule  string
	BizType    string
	BizId      string
	UploadedBy string
	Page       int
	PageSize   int
}

type Repository interface {
	List(ctx context.Context, filter ListFilter) ([]model.FileAsset, int64, error)
	Get(ctx context.Context, id string) (*model.FileAsset, error)
	Create(ctx context.Context, item *model.FileAsset) (*model.FileAsset, error)
	Update(ctx context.Context, item *model.FileAsset) (*model.FileAsset, error)
	Delete(ctx context.Context, id string) error
	Bind(ctx context.Context, id string, binding model.FileBinding) (*model.FileAsset, error)
	Unbind(ctx context.Context, id string) (*model.FileAsset, error)
	DefaultStorageDriver(ctx context.Context, fallback string) (string, error)
	SetDefaultStorageDriver(ctx context.Context, driver string) error
}

package repository

import (
	"context"
	"errors"

	pluginmodel "goadmin/plugin/domain/model"
)

var (
	ErrNotFound = errors.New("plugin not found")
	ErrConflict = errors.New("plugin already exists")
)

type Repository interface {
	List(ctx context.Context) ([]pluginmodel.Plugin, int64, error)
	Get(ctx context.Context, name string) (*pluginmodel.Plugin, error)
	Create(ctx context.Context, plugin *pluginmodel.Plugin) (*pluginmodel.Plugin, error)
	Update(ctx context.Context, plugin *pluginmodel.Plugin) (*pluginmodel.Plugin, error)
	Delete(ctx context.Context, name string) error
}

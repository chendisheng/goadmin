package storage

import (
	"fmt"
	"strings"

	"goadmin/core/config"
	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
	"goadmin/modules/upload/infrastructure/storage/local"
)

func NewDriver(cfg config.UploadStorageConfig) (storagecontract.Driver, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Driver)) {
	case "", "local":
		return local.NewDriver(cfg.Local)
	case "s3-compatible", "oss", "cos", "minio":
		return nil, fmt.Errorf("upload storage driver %q is not implemented yet", cfg.Driver)
	default:
		return nil, fmt.Errorf("unsupported upload storage driver %q", cfg.Driver)
	}
}

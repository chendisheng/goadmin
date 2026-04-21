package storage

import (
	"fmt"
	"strings"

	"goadmin/core/config"
	storagecontract "goadmin/modules/upload/infrastructure/storage/contract"
	"goadmin/modules/upload/infrastructure/storage/local"
	"goadmin/modules/upload/infrastructure/storage/objectstore"
	qiniustorage "goadmin/modules/upload/infrastructure/storage/qiniu"

	"gorm.io/gorm"
)

func NewDriver(cfg config.UploadStorageConfig) (storagecontract.Driver, error) {
	return NewDriverWithDB(nil, cfg)
}

func NewDriverWithDB(db *gorm.DB, cfg config.UploadStorageConfig) (storagecontract.Driver, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Driver)) {
	case "", "local":
		return local.NewDriver(cfg.Local)
	case "db", "database":
		driver := NewDatabaseDriver()
		if err := driver.SetDB(db); err != nil {
			return nil, err
		}
		return driver, nil
	case "s3-compatible":
		driver, err := objectstore.NewS3CompatibleDriver(cfg.S3Compatible)
		if err != nil {
			return nil, err
		}
		return driver, nil
	case "oss":
		driver, err := objectstore.NewOSSDriver(cfg.OSS)
		if err != nil {
			return nil, err
		}
		return driver, nil
	case "cos":
		driver, err := objectstore.NewCOSDriver(cfg.COS)
		if err != nil {
			return nil, err
		}
		return driver, nil
	case "qiniu":
		driver, err := qiniustorage.NewDriver(cfg.Qiniu)
		if err != nil {
			return nil, err
		}
		return driver, nil
	case "minio":
		driver, err := objectstore.NewMinIODriver(cfg.MinIO)
		if err != nil {
			return nil, err
		}
		return driver, nil
	default:
		return nil, fmt.Errorf("unsupported upload storage driver %q", cfg.Driver)
	}
}

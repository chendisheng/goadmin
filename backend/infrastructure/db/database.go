package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"goadmin/core/config"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Driver)) {
	case "", "mysql":
		dsn := strings.TrimSpace(cfg.DSN)
		if dsn == "" {
			return nil, fmt.Errorf("database.dsn is required")
		}
		return gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	case "postgres", "postgresql":
		dsn := strings.TrimSpace(cfg.DSN)
		if dsn == "" {
			return nil, fmt.Errorf("database.dsn is required")
		}
		return gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	case "sqlite":
		dsn := strings.TrimSpace(cfg.DSN)
		if dsn == "" {
			return nil, fmt.Errorf("database.dsn is required")
		}
		if err := ensureSQLiteDir(dsn); err != nil {
			return nil, err
		}
		return gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	default:
		return nil, fmt.Errorf("unsupported database driver %q", cfg.Driver)
	}
}

func AutoMigrate(db *gorm.DB, models ...any) error {
	if db == nil || len(models) == 0 {
		return nil
	}
	return db.AutoMigrate(models...)
}

func ensureSQLiteDir(dsn string) error {
	if !strings.HasPrefix(dsn, "file:") {
		return nil
	}
	path := strings.TrimPrefix(dsn, "file:")
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == string(filepath.Separator) {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

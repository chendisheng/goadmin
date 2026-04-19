package db

import (
	"testing"

	gormlogger "gorm.io/gorm/logger"
)

func TestSQLLogMode(t *testing.T) {
	t.Parallel()

	if got := sqlLogMode(true); got != gormlogger.Info {
		t.Fatalf("sqlLogMode(true) = %v, want %v", got, gormlogger.Info)
	}
	if got := sqlLogMode(false); got != gormlogger.Silent {
		t.Fatalf("sqlLogMode(false) = %v, want %v", got, gormlogger.Silent)
	}
}

// Package db defines the database-driven input adapter contracts.
package db

import "goadmin/codegen/schema/database"

// Inspector is the read-only contract implemented by database adapters.
// It is intentionally narrow so that database dialect details stay inside the
// infrastructure layer.
type Inspector interface {
	InspectTables() ([]database.Table, error)
	InspectColumns(table string) ([]database.Column, error)
	InspectRelations(table string) ([]database.ForeignKey, error)
}

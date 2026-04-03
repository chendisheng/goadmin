package model

import (
	diffmodel "goadmin/codegen/model/diff"
	irmodel "goadmin/codegen/model/ir"
	dbschema "goadmin/codegen/schema/database"
)

// Compatibility aliases that expose the v3 core models from the legacy model
// package. Existing planner and generator code can keep importing this package
// while the new subpackages become the canonical home for v3 structures.
type DatabaseDriverKind = dbschema.DriverKind

type DatabaseTable = dbschema.Table
type DatabaseColumn = dbschema.Column
type DatabaseIndex = dbschema.Index
type DatabaseForeignKey = dbschema.ForeignKey

type DatabaseSnapshot = dbschema.Snapshot

type IRSourceKind = irmodel.SourceKind

type IRField = irmodel.Field
type IRSemantic = irmodel.Semantic
type IRRelation = irmodel.Relation
type IRPage = irmodel.Page
type IRPermission = irmodel.Permission
type IRRoute = irmodel.Route
type IRResource = irmodel.Resource
type IRDocument = irmodel.Document
type IRDiffType = diffmodel.Type
type IRDiffSeverity = diffmodel.Severity
type IRDiffItem = diffmodel.Item
type IRDiffDocument = diffmodel.Document

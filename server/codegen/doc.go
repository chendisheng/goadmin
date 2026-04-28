// Package codegen defines the next-generation generation pipeline for GoAdmin.
//
// The package is intentionally layered so the current CLI generator can keep
// working while the project gradually adopts a schema -> model -> planner ->
// generator -> merger -> postprocess flow.
package codegen

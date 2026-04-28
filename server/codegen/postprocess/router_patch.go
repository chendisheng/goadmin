package postprocess

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
)

// PatchGoRouterRegistration merges an import and appends a registration call to
// the Register function of a Go router source file.
func PatchGoRouterRegistration(source []byte, importAlias, importPath, registrationCall string) ([]byte, error) {
	trimmed := strings.TrimSpace(string(source))
	if trimmed == "" {
		return nil, fmt.Errorf("source is required")
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "router.go", source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse router source: %w", err)
	}
	ensureImport(file, importAlias, importPath)
	if strings.TrimSpace(registrationCall) != "" {
		if err := appendRegistrationCall(file, registrationCall); err != nil {
			return nil, err
		}
	}
	var buf bytes.Buffer
	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	if err := cfg.Fprint(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("print router source: %w", err)
	}
	return EnsureTrailingNewline(buf.Bytes()), nil
}

func ensureImport(file *ast.File, importAlias, importPath string) {
	alias := strings.TrimSpace(importAlias)
	path := strings.TrimSpace(importPath)
	if file == nil || path == "" {
		return
	}
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}
		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok || importSpec.Path == nil {
				continue
			}
			if strings.Trim(importSpec.Path.Value, `"`) == path && importAliasMatches(importSpec.Name, alias) {
				return
			}
		}
	}
	importSpec := &ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", path)}}
	if alias != "" {
		importSpec.Name = ast.NewIdent(alias)
	}
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}
		genDecl.Specs = append(genDecl.Specs, importSpec)
		return
	}
	file.Decls = append([]ast.Decl{&ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{importSpec}}}, file.Decls...)
}

func importAliasMatches(name *ast.Ident, alias string) bool {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return name == nil
	}
	return name != nil && name.Name == alias
}

func appendRegistrationCall(file *ast.File, registrationCall string) error {
	stmt, err := parseStatement(registrationCall)
	if err != nil {
		return err
	}
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name == nil || funcDecl.Name.Name != "Register" || funcDecl.Body == nil {
			continue
		}
		if containsStatement(funcDecl.Body.List, registrationCall) {
			return nil
		}
		funcDecl.Body.List = append(funcDecl.Body.List, stmt)
		return nil
	}
	return fmt.Errorf("register function not found")
}

func parseStatement(statement string) (ast.Stmt, error) {
	source := "package main\nfunc _() {\n" + strings.TrimSpace(statement) + "\n}"
	file, err := parser.ParseFile(token.NewFileSet(), "stmt.go", source, 0)
	if err != nil {
		return nil, fmt.Errorf("parse registration statement: %w", err)
	}
	if len(file.Decls) == 0 {
		return nil, fmt.Errorf("statement parsing produced no declarations")
	}
	funcDecl, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok || funcDecl.Body == nil || len(funcDecl.Body.List) == 0 {
		return nil, fmt.Errorf("statement parsing produced no body statements")
	}
	return funcDecl.Body.List[0], nil
}

func containsStatement(stmts []ast.Stmt, text string) bool {
	needle := normalizeStatement(text)
	for _, stmt := range stmts {
		var buf bytes.Buffer
		if err := printer.Fprint(&buf, token.NewFileSet(), stmt); err != nil {
			continue
		}
		if normalizeStatement(buf.String()) == needle {
			return true
		}
	}
	return false
}

func normalizeStatement(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), "")
}

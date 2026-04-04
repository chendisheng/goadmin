package merger

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"sort"
	"strings"

	diffmodel "goadmin/codegen/model/diff"
	"goadmin/codegen/postprocess"
)

type Result struct {
	Content  []byte
	Diff     diffmodel.Document
	Changed  bool
	Conflict bool
}

func MergeContent(path string, current, generated []byte, force bool) (Result, error) {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return Result{}, fmt.Errorf("path is required")
	}
	generated = postprocess.EnsureTrailingNewline(generated)
	if len(current) == 0 || force {
		return Result{
			Content: generated,
			Diff: diffmodel.Document{
				Items: []diffmodel.Item{{
					Type:     diffmodel.TypeAddFile,
					Target:   cleanPath,
					Severity: diffmodel.SeverityLow,
					Metadata: map[string]any{"force": force},
				}},
			},
			Changed: true,
		}, nil
	}
	current = postprocess.EnsureTrailingNewline(current)
	if bytes.Equal(bytes.TrimSpace(current), bytes.TrimSpace(generated)) {
		return Result{Content: current}, nil
	}
	ext := strings.ToLower(filepath.Ext(cleanPath))
	switch ext {
	case ".go":
		return mergeGoContent(cleanPath, current, generated)
	case ".csv":
		return mergeDelimitedContent(cleanPath, current, generated, true), nil
	case ".yaml", ".yml":
		return mergeYAMLContent(cleanPath, current, generated), nil
	default:
		return Result{
			Content: current,
			Diff: diffmodel.Document{Items: []diffmodel.Item{{
				Type:     diffmodel.TypeMergeConflict,
				Target:   cleanPath,
				Patch:    "existing file differs from generated content and no merge strategy is registered",
				Severity: diffmodel.SeverityHigh,
				Conflict: true,
			}}},
			Changed:  false,
			Conflict: true,
		}, nil
	}
}

func mergeYAMLContent(path string, current, generated []byte) Result {
	content := postprocess.EnsureTrailingNewline(generated)
	return Result{
		Content: content,
		Diff: diffmodel.Document{Items: []diffmodel.Item{{
			Type:     diffmodel.TypeModifyFile,
			Target:   path,
			Severity: diffmodel.SeverityLow,
			Metadata: map[string]any{"strategy": "replace"},
		}}},
		Changed: !bytes.Equal(bytes.TrimSpace(current), bytes.TrimSpace(content)),
	}
}

func mergeGoContent(path string, current, generated []byte) (Result, error) {
	fset := token.NewFileSet()
	currentFile, err := parser.ParseFile(fset, path, current, parser.ParseComments)
	if err != nil {
		return Result{
			Content:  current,
			Diff:     conflictDiff(path, "current Go file is not parseable; preserving existing content"),
			Conflict: true,
		}, nil
	}
	generatedFile, err := parser.ParseFile(fset, path, generated, parser.ParseComments)
	if err != nil {
		return Result{}, fmt.Errorf("parse generated go source for %s: %w", path, err)
	}

	merged := &ast.File{Name: currentFile.Name}
	merged.Comments = currentFile.Comments
	merged.Decls = mergeGoDecls(currentFile.Decls, generatedFile.Decls)

	var buf bytes.Buffer
	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	if err := cfg.Fprint(&buf, fset, merged); err != nil {
		return Result{}, fmt.Errorf("print merged go source for %s: %w", path, err)
	}
	content := postprocess.EnsureTrailingNewline(buf.Bytes())
	return Result{
		Content: content,
		Diff:    buildGoDiff(path, currentFile, generatedFile, merged),
		Changed: !bytes.Equal(bytes.TrimSpace(current), bytes.TrimSpace(content)),
	}, nil
}

func mergeGoDecls(currentDecls, generatedDecls []ast.Decl) []ast.Decl {
	decls := append([]ast.Decl(nil), currentDecls...)
	importIndex := findImportDeclIndex(decls)
	currentImports := collectImportSpecs(currentDecls)
	generatedImports := collectImportSpecs(generatedDecls)
	if len(generatedImports) > 0 {
		mergedImportDecl := buildImportDecl(currentImports, generatedImports)
		if importIndex >= 0 {
			decls[importIndex] = mergedImportDecl
			decls = removeOtherImportDecls(decls, importIndex)
		} else {
			decls = append([]ast.Decl{mergedImportDecl}, decls...)
		}
	}
	for _, generatedDecl := range generatedDecls {
		if isImportDecl(generatedDecl) {
			continue
		}
		key, ok := declKey(generatedDecl)
		if !ok {
			decls = append(decls, generatedDecl)
			continue
		}
		if index := findDeclIndexByKey(decls, key); index >= 0 {
			decls[index] = generatedDecl
			continue
		}
		decls = append(decls, generatedDecl)
	}
	return decls
}

func buildGoDiff(path string, currentFile, generatedFile, merged *ast.File) diffmodel.Document {
	doc := diffmodel.Document{Metadata: map[string]any{"path": path}}
	currentKeys := topLevelKeys(currentFile.Decls)
	generatedKeys := topLevelKeys(generatedFile.Decls)
	mergedKeys := topLevelKeys(merged.Decls)

	for _, key := range mergedKeys {
		if containsString(currentKeys, key) && containsString(generatedKeys, key) {
			doc = doc.Append(diffmodel.NewItem(diffmodel.TypeModifyDeclaration, key, diffmodel.SeverityMedium))
			continue
		}
		if !containsString(currentKeys, key) && containsString(generatedKeys, key) {
			doc = doc.Append(diffmodel.NewItem(diffmodel.TypeAddDeclaration, key, diffmodel.SeverityLow))
		}
	}
	for _, key := range currentKeys {
		if !containsString(mergedKeys, key) {
			doc = doc.Append(diffmodel.NewItem(diffmodel.TypeRemoveDeclaration, key, diffmodel.SeverityMedium))
		}
	}
	currentImports := importSet(currentFile)
	generatedImports := importSet(generatedFile)
	mergedImports := importSet(merged)
	for key := range mergedImports {
		if _, ok := currentImports[key]; ok {
			if _, ok := generatedImports[key]; ok {
				doc = doc.Append(diffmodel.NewItem(diffmodel.TypeModifyImport, key, diffmodel.SeverityLow))
			}
			continue
		}
		if _, ok := generatedImports[key]; ok {
			doc = doc.Append(diffmodel.NewItem(diffmodel.TypeAddImport, key, diffmodel.SeverityLow))
		}
	}
	for key := range currentImports {
		if _, ok := mergedImports[key]; !ok {
			doc = doc.Append(diffmodel.NewItem(diffmodel.TypeRemoveImport, key, diffmodel.SeverityMedium))
		}
	}
	return doc
}

func mergeDelimitedContent(path string, current, generated []byte, policy bool) Result {
	currentLines := splitLines(string(current))
	generatedLines := splitLines(string(generated))
	if policy {
		merged := postprocess.NormalizePolicyLines(append(currentLines, generatedLines...))
		content := []byte(strings.Join(merged, "\n"))
		if len(content) > 0 {
			content = append(content, '\n')
		}
		return Result{
			Content: content,
			Diff: diffmodel.Document{Items: []diffmodel.Item{{
				Type:     diffmodel.TypeModifyPolicy,
				Target:   path,
				Severity: diffmodel.SeverityLow,
				Metadata: map[string]any{"lines": len(merged)},
			}}},
			Changed: !bytes.Equal(bytes.TrimSpace(current), bytes.TrimSpace(content)),
		}
	}
	merged := mergeUniqueLines(currentLines, generatedLines)
	content := []byte(strings.Join(merged, "\n"))
	if len(content) > 0 {
		content = append(content, '\n')
	}
	return Result{
		Content: content,
		Diff: diffmodel.Document{Items: []diffmodel.Item{{
			Type:     diffmodel.TypeModifyFile,
			Target:   path,
			Severity: diffmodel.SeverityLow,
			Metadata: map[string]any{"lines": len(merged)},
		}}},
		Changed: !bytes.Equal(bytes.TrimSpace(current), bytes.TrimSpace(content)),
	}
}

func mergeUniqueLines(currentLines, generatedLines []string) []string {
	seen := make(map[string]struct{}, len(currentLines)+len(generatedLines))
	result := make([]string, 0, len(currentLines)+len(generatedLines))
	for _, line := range currentLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	for _, line := range generatedLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func splitLines(content string) []string {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	return strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
}

func conflictDiff(path, reason string) diffmodel.Document {
	return diffmodel.Document{Items: []diffmodel.Item{{
		Type:     diffmodel.TypeMergeConflict,
		Target:   path,
		Patch:    reason,
		Severity: diffmodel.SeverityCritical,
		Conflict: true,
	}}}
}

func topLevelKeys(decls []ast.Decl) []string {
	keys := make([]string, 0, len(decls))
	for _, decl := range decls {
		if key, ok := declKey(decl); ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func declKey(decl ast.Decl) (string, bool) {
	switch node := decl.(type) {
	case *ast.FuncDecl:
		if node.Name == nil {
			return "", false
		}
		return "func:" + node.Name.Name, true
	case *ast.GenDecl:
		if node.Tok == token.IMPORT {
			return "", false
		}
		if len(node.Specs) == 0 {
			return "", false
		}
		names := make([]string, 0, len(node.Specs))
		for _, spec := range node.Specs {
			switch typed := spec.(type) {
			case *ast.TypeSpec:
				names = append(names, "type:"+typed.Name.Name)
			case *ast.ValueSpec:
				for _, name := range typed.Names {
					prefix := "var:"
					if node.Tok == token.CONST {
						prefix = "const:"
					}
					names = append(names, prefix+name.Name)
				}
			default:
				return "", false
			}
		}
		sort.Strings(names)
		return strings.Join(names, ","), true
	default:
		return "", false
	}
}

func findDeclIndexByKey(decls []ast.Decl, key string) int {
	for index, decl := range decls {
		if existingKey, ok := declKey(decl); ok && existingKey == key {
			return index
		}
	}
	return -1
}

func isImportDecl(decl ast.Decl) bool {
	genDecl, ok := decl.(*ast.GenDecl)
	return ok && genDecl.Tok == token.IMPORT
}

func findImportDeclIndex(decls []ast.Decl) int {
	for index, decl := range decls {
		if isImportDecl(decl) {
			return index
		}
	}
	return -1
}

func removeOtherImportDecls(decls []ast.Decl, keepIndex int) []ast.Decl {
	result := make([]ast.Decl, 0, len(decls))
	for index, decl := range decls {
		if isImportDecl(decl) && index != keepIndex {
			continue
		}
		result = append(result, decl)
	}
	return result
}

func collectImportSpecs(decls []ast.Decl) []*ast.ImportSpec {
	imports := make([]*ast.ImportSpec, 0)
	for _, decl := range decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}
		for _, spec := range genDecl.Specs {
			if importSpec, ok := spec.(*ast.ImportSpec); ok {
				imports = append(imports, importSpec)
			}
		}
	}
	return imports
}

func buildImportDecl(currentImports, generatedImports []*ast.ImportSpec) ast.Decl {
	unique := make(map[string]*ast.ImportSpec, len(currentImports)+len(generatedImports))
	ordered := make([]*ast.ImportSpec, 0, len(currentImports)+len(generatedImports))
	for _, spec := range currentImports {
		key := importSpecKey(spec)
		if _, ok := unique[key]; ok {
			continue
		}
		unique[key] = spec
		ordered = append(ordered, spec)
	}
	for _, spec := range generatedImports {
		key := importSpecKey(spec)
		if _, ok := unique[key]; ok {
			continue
		}
		unique[key] = spec
		ordered = append(ordered, spec)
	}
	if len(ordered) == 0 {
		return &ast.GenDecl{Tok: token.IMPORT}
	}
	return &ast.GenDecl{Tok: token.IMPORT, Lparen: 1, Specs: importSpecsToDecls(ordered)}
}

func importSpecsToDecls(specs []*ast.ImportSpec) []ast.Spec {
	decls := make([]ast.Spec, 0, len(specs))
	for _, spec := range specs {
		decls = append(decls, spec)
	}
	return decls
}

func importSpecKey(spec *ast.ImportSpec) string {
	if spec == nil || spec.Path == nil {
		return ""
	}
	alias := ""
	if spec.Name != nil {
		alias = spec.Name.Name
	}
	return alias + "|" + strings.Trim(spec.Path.Value, `"`)
}

func importSet(file *ast.File) map[string]struct{} {
	result := make(map[string]struct{})
	if file == nil {
		return result
	}
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}
		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}
			result[importSpecKey(importSpec)] = struct{}{}
		}
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

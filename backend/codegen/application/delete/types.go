package deleteapp

import (
	menuservice "goadmin/modules/menu/application/service"
	"path/filepath"
	"strings"
	"unicode"

	lifecycle "goadmin/codegen/model/lifecycle"
)

type Dependencies struct {
	ProjectRoot   string
	BackendRoot   string
	PolicyStore   string
	MenuService   *menuservice.Service
	PolicyCleanup *PolicyCleanupService
}

type PreviewReport struct {
	Request    lifecycle.DeleteRequest `json:"request,omitempty"`
	Resolution ModuleResolution        `json:"resolution,omitempty"`
	Plan       lifecycle.DeletePlan    `json:"plan,omitempty"`
}

type ModuleResolution struct {
	Input               string                            `json:"input,omitempty"`
	Module              string                            `json:"module,omitempty"`
	Kind                string                            `json:"kind,omitempty"`
	ProjectRoot         string                            `json:"project_root,omitempty"`
	BackendRoot         string                            `json:"backend_root,omitempty"`
	ModuleDir           string                            `json:"module_dir,omitempty"`
	ManifestPath        string                            `json:"manifest_path,omitempty"`
	ModuleGoPath        string                            `json:"module_go_path,omitempty"`
	BootstrapPath       string                            `json:"bootstrap_path,omitempty"`
	RegistryPath        string                            `json:"registry_path,omitempty"`
	BuiltinRegistryPath string                            `json:"builtin_registry_path,omitempty"`
	ManifestName        string                            `json:"manifest_name,omitempty"`
	ManifestKind        string                            `json:"manifest_kind,omitempty"`
	ManifestVersion     string                            `json:"manifest_version,omitempty"`
	GeneratedBootstrap  bool                              `json:"generated_bootstrap,omitempty"`
	HasManifest         bool                              `json:"has_manifest,omitempty"`
	HasModuleGo         bool                              `json:"has_module_go,omitempty"`
	IsBuiltin           bool                              `json:"is_builtin,omitempty"`
	PolicyStore         lifecycle.PolicyStoreKind         `json:"policy_store,omitempty"`
	Compatibility       lifecycle.LegacyCompatibilityRule `json:"compatibility,omitempty"`
}

type moduleManifest struct {
	Name         string                     `yaml:"name,omitempty"`
	Version      string                     `yaml:"version,omitempty"`
	Kind         string                     `yaml:"kind,omitempty"`
	Module       string                     `yaml:"module,omitempty"`
	Routes       []moduleManifestRoute      `yaml:"routes,omitempty"`
	Menus        []moduleManifestMenu       `yaml:"menus,omitempty"`
	Permissions  []moduleManifestPermission `yaml:"permissions,omitempty"`
	Capabilities []string                   `yaml:"capabilities,omitempty"`
}

type moduleManifestRoute struct {
	Method string `yaml:"method,omitempty"`
	Path   string `yaml:"path,omitempty"`
	Name   string `yaml:"name,omitempty"`
}

type moduleManifestMenu struct {
	Name       string `yaml:"name,omitempty"`
	Path       string `yaml:"path,omitempty"`
	ParentPath string `yaml:"parent_path,omitempty"`
	Component  string `yaml:"component,omitempty"`
	Permission string `yaml:"permission,omitempty"`
	Type       string `yaml:"type,omitempty"`
	Redirect   string `yaml:"redirect,omitempty"`
	Visible    bool   `yaml:"visible,omitempty"`
	Enabled    bool   `yaml:"enabled,omitempty"`
	Sort       int    `yaml:"sort,omitempty"`
}

type moduleManifestPermission struct {
	Object      string `yaml:"object,omitempty"`
	Action      string `yaml:"action,omitempty"`
	Description string `yaml:"description,omitempty"`
}

const (
	conflictSeverityInfo    = "info"
	conflictSeverityWarning = "warning"
	conflictSeverityHigh    = "high"
)

func NormalizeModuleName(input string) string {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return ""
	}
	raw = filepath.ToSlash(raw)
	raw = strings.TrimRight(raw, "/")
	lower := strings.ToLower(raw)
	for _, suffix := range []string{"/manifest.yaml", "/manifest.yml", "/module.go", "/bootstrap.go", "/schema.sql", "/codegen.manifest.json"} {
		if strings.HasSuffix(lower, suffix) {
			raw = raw[:len(raw)-len(suffix)]
			break
		}
	}
	if strings.Contains(raw, "/") {
		raw = filepath.Base(raw)
	}
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, ".go")
	raw = strings.TrimSuffix(raw, ".yaml")
	raw = strings.TrimSuffix(raw, ".yml")
	raw = strings.TrimSuffix(raw, ".json")
	raw = strings.Trim(raw, "._-/\\")
	if raw == "" {
		return ""
	}
	raw = strings.ReplaceAll(raw, "-", "_")
	raw = strings.ReplaceAll(raw, " ", "_")
	return toSnake(raw)
}

func normalizePolicyStoreSource(value string) lifecycle.PolicyStoreKind {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "file", "csv":
		return lifecycle.PolicyStoreCSV
	case "db":
		return lifecycle.PolicyStoreDB
	default:
		return lifecycle.PolicyStoreUnknown
	}
}

func toSnake(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(value) + 8)
	var prev rune
	writeUnderscore := func() {
		if builder.Len() == 0 {
			return
		}
		if strings.HasSuffix(builder.String(), "_") {
			return
		}
		builder.WriteByte('_')
	}
	for _, r := range value {
		switch {
		case r == '_' || r == '-' || r == '/' || r == '\\':
			writeUnderscore()
		case unicode.IsUpper(r):
			if prev != 0 && prev != '_' && prev != '-' && prev != '/' && prev != '\\' && (unicode.IsLower(prev) || unicode.IsDigit(prev)) {
				writeUnderscore()
			}
			builder.WriteRune(unicode.ToLower(r))
		default:
			builder.WriteRune(unicode.ToLower(r))
		}
		prev = r
	}
	result := strings.Trim(builder.String(), "_")
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	return result
}

func pluralize(value string) string {
	base := NormalizeModuleName(value)
	if base == "" {
		return ""
	}
	switch {
	case strings.HasSuffix(base, "ch"), strings.HasSuffix(base, "sh"), strings.HasSuffix(base, "s"), strings.HasSuffix(base, "x"), strings.HasSuffix(base, "z"):
		return base + "es"
	case strings.HasSuffix(base, "y") && len(base) > 1 && !isVowel(rune(base[len(base)-2])):
		return base[:len(base)-1] + "ies"
	default:
		return base + "s"
	}
}

func titleFromModule(value string) string {
	base := NormalizeModuleName(value)
	if base == "" {
		return "Module"
	}
	parts := strings.Split(base, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}

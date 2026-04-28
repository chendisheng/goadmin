package schema

import (
	"strings"
)

func (p Page) NormalizedName() string {
	if name := strings.TrimSpace(p.Name); name != "" {
		return NormalizeName(name)
	}
	if component := strings.TrimSpace(p.Component); component != "" {
		parts := strings.Split(component, "/")
		if len(parts) > 0 {
			return NormalizeName(parts[len(parts)-1])
		}
	}
	if pageType := strings.TrimSpace(p.Type); pageType != "" {
		return NormalizeName(pageType)
	}
	return ""
}

func (p Page) ComponentName(scope string) string {
	if component := strings.TrimSpace(p.Component); component != "" {
		return component
	}
	slug := p.NormalizedName()
	if slug == "" {
		slug = "index"
	}
	scope = NormalizeName(scope)
	if scope == "" {
		scope = "page"
	}
	return "view/" + scope + "/" + slug
}

func (p Page) RoutePath(base string) string {
	if path := strings.TrimSpace(p.Path); path != "" {
		return normalizePath(path)
	}
	slug := p.NormalizedName()
	if slug == "" {
		slug = "index"
	}
	base = NormalizeName(base)
	if base == "" {
		base = "page"
	}
	return normalizePath("/" + base + "/" + slug)
}

func (p Page) Title() string {
	if title := strings.TrimSpace(p.Name); title != "" {
		return title
	}
	if pageType := strings.TrimSpace(p.Type); pageType != "" {
		return pageType
	}
	return "Page"
}

func (p Page) PermissionAction() string {
	text := strings.ToLower(strings.Join([]string{p.Name, p.Type, p.Path, p.Component}, " "))
	switch {
	case strings.Contains(text, "delete"), strings.Contains(text, "remove"):
		return "delete"
	case strings.Contains(text, "create"), strings.Contains(text, "add"), strings.Contains(text, "new"):
		return "create"
	case strings.Contains(text, "edit"), strings.Contains(text, "update"), strings.Contains(text, "modify"), strings.Contains(text, "form"):
		return "update"
	case strings.Contains(text, "export"):
		return "export"
	case strings.Contains(text, "list"), strings.Contains(text, "index"), strings.Contains(text, "table"):
		return "list"
	default:
		return "view"
	}
}

func (p Permission) PolicyParts() (string, string, bool) {
	resource := strings.TrimSpace(p.Resource)
	action := strings.TrimSpace(p.Action)
	if resource == "" || action == "" {
		name := strings.TrimSpace(p.Name)
		if name != "" {
			if left, right, ok := strings.Cut(name, ":"); ok {
				if resource == "" {
					resource = NormalizeName(left)
				}
				if action == "" {
					action = NormalizeName(right)
				}
			}
		}
	}
	if resource == "" || action == "" {
		return "", "", false
	}
	return resource, action, true
}

func (p Permission) StandardActions() (string, []string, bool) {
	resource, action, ok := p.PolicyParts()
	if !ok {
		return "", nil, false
	}
	actions := expandCRUDAction(action)
	return resource, actions, true
}

func expandCRUDAction(action string) []string {
	normalized := strings.ToLower(strings.TrimSpace(action))
	switch normalized {
	case "", "view", "read", "show", "detail":
		return []string{"list", "view"}
	case "list", "index", "browse":
		return []string{"list"}
	case "create", "add", "new":
		return []string{"create"}
	case "update", "edit", "modify", "write":
		return []string{"update"}
	case "delete", "remove", "destroy":
		return []string{"delete"}
	case "export":
		return []string{"export"}
	case "crud", "manage", "all":
		return []string{"list", "view", "create", "update", "delete"}
	default:
		return []string{normalized}
	}
}

func (r Route) PolicyParts() (string, string, bool) {
	path := strings.TrimSpace(r.Path)
	method := strings.ToUpper(strings.TrimSpace(r.Method))
	if path == "" || method == "" {
		return "", "", false
	}
	return normalizePath(path), method, true
}

func normalizePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	if trimmed != "/" {
		trimmed = strings.TrimRight(trimmed, "/")
	}
	return trimmed
}

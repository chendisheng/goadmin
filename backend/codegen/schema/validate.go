package schema

import (
	"fmt"
	"strings"
)

func (f Field) Validate() error {
	if strings.TrimSpace(f.Name) == "" {
		return fmt.Errorf("field name is required")
	}
	return nil
}

func (e Entity) Validate() error {
	if strings.TrimSpace(e.Name) == "" && len(e.Fields) == 0 {
		return nil
	}
	if strings.TrimSpace(e.Name) == "" && len(e.Fields) > 0 {
		return fmt.Errorf("entity name is required when fields are defined")
	}
	for i, field := range e.Fields {
		if err := field.Validate(); err != nil {
			return fmt.Errorf("entity field[%d]: %w", i, err)
		}
	}
	return nil
}

func (p Page) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("page name is required")
	}
	return nil
}

func (r Route) Validate() error {
	if strings.TrimSpace(r.Method) == "" {
		return fmt.Errorf("route method is required")
	}
	if strings.TrimSpace(r.Path) == "" {
		return fmt.Errorf("route path is required")
	}
	return nil
}

func (p Permission) Validate() error {
	if strings.TrimSpace(p.Name) == "" && strings.TrimSpace(p.Action) == "" && strings.TrimSpace(p.Resource) == "" {
		return fmt.Errorf("permission is required")
	}
	return nil
}

func (p Plugin) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("plugin name is required")
	}
	return nil
}

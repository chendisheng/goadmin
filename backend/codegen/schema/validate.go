package schema

import (
	"fmt"
	"strings"
)

func (f Field) Validate() error {
	if strings.TrimSpace(f.Name) == "" {
		return fmt.Errorf("field name is required")
	}
	if f.Enum != nil {
		if err := f.Enum.Validate(); err != nil {
			return fmt.Errorf("field %s enum: %w", strings.TrimSpace(f.Name), err)
		}
	}
	return nil
}

func (e EnumField) Validate() error {
	if len(e.Values) == 0 && len(e.Options) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(e.Options)+len(e.Values))
	for i, option := range e.Options {
		if err := option.Validate(); err != nil {
			return fmt.Errorf("option[%d]: %w", i, err)
		}
		key := strings.ToLower(strings.TrimSpace(option.Value))
		if key == "" {
			return fmt.Errorf("option[%d] value is required", i)
		}
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate enum value %q", option.Value)
		}
		seen[key] = struct{}{}
	}
	for i, value := range e.Values {
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" {
			return fmt.Errorf("enum value[%d] is required", i)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
	}
	return nil
}

func (o EnumOption) Validate() error {
	if strings.TrimSpace(o.Value) == "" && strings.TrimSpace(o.Label) == "" {
		return fmt.Errorf("enum option requires value or label")
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

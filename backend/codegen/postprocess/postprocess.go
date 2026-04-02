package postprocess

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
)

func FormatGo(source []byte) ([]byte, error) {
	if len(source) == 0 {
		return source, nil
	}
	formatted, err := format.Source(source)
	if err != nil {
		return nil, fmt.Errorf("format go source: %w", err)
	}
	return formatted, nil
}

func EnsureTrailingNewline(source []byte) []byte {
	if len(source) == 0 || bytes.HasSuffix(source, []byte("\n")) {
		return source
	}
	return append(source, '\n')
}

func NormalizePolicyLines(lines []string) []string {
	seen := make(map[string]struct{}, len(lines))
	result := make([]string, 0, len(lines))
	for _, line := range lines {
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

func NormalizeRoutes(routes []string) []string {
	cleaned := make([]string, 0, len(routes))
	for _, route := range routes {
		trimmed := strings.TrimSpace(route)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}

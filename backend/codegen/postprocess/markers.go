package postprocess

import (
	"path/filepath"
	"strings"
)

type MarkerStyle struct {
	Begin string
	End   string
}

func MarkerStyleForPath(path string) MarkerStyle {
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(path))) {
	case ".go", ".ts", ".js", ".mjs", ".cjs":
		return MarkerStyle{Begin: "// codegen:begin", End: "// codegen:end"}
	case ".vue":
		return MarkerStyle{Begin: "<!-- codegen:begin -->", End: "<!-- codegen:end -->"}
	case ".yaml", ".yml", ".csv", ".md":
		return MarkerStyle{Begin: "# codegen:begin", End: "# codegen:end"}
	default:
		return MarkerStyle{Begin: "// codegen:begin", End: "// codegen:end"}
	}
}

func WrapGeneratedContent(path string, source []byte) []byte {
	trimmed := strings.TrimSpace(string(source))
	if trimmed == "" {
		return source
	}
	if HasGeneratedMarkers(path, source) {
		return source
	}
	style := MarkerStyleForPath(path)
	var builder strings.Builder
	builder.Grow(len(source) + len(style.Begin) + len(style.End) + 32)
	builder.WriteString(style.Begin)
	builder.WriteString("\n")
	builder.WriteString(strings.TrimRight(string(source), "\n"))
	builder.WriteString("\n")
	builder.WriteString(style.End)
	builder.WriteString("\n")
	return []byte(builder.String())
}

func HasGeneratedMarkers(path string, source []byte) bool {
	style := MarkerStyleForPath(path)
	content := string(source)
	return strings.Contains(content, style.Begin) && strings.Contains(content, style.End)
}

func UnwrapGeneratedContent(path string, source []byte) ([]byte, bool) {
	style := MarkerStyleForPath(path)
	content := string(source)
	begin := strings.Index(content, style.Begin)
	if begin < 0 {
		return source, false
	}
	end := strings.LastIndex(content, style.End)
	if end < 0 || end <= begin {
		return source, false
	}
	beginLineEnd := strings.Index(content[begin:], "\n")
	if beginLineEnd < 0 {
		return source, false
	}
	beginLineEnd += begin + 1
	inner := content[beginLineEnd:end]
	inner = strings.TrimPrefix(inner, "\n")
	inner = strings.TrimSuffix(inner, "\n")
	return []byte(inner), true
}

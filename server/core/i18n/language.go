package i18n

import (
	"net/http"
	"sort"
	"strings"
	"sync"
)

const (
	LanguageZHCN = "zh-CN"
	LanguageENUS = "en-US"
)

var (
	runtimeMu                 sync.RWMutex
	runtimeDefaultLanguage    = LanguageZHCN
	runtimeSupportedLanguages = []string{LanguageZHCN, LanguageENUS}
)

func DefaultLanguage() string {
	runtimeMu.RLock()
	defer runtimeMu.RUnlock()
	if runtimeDefaultLanguage == "" {
		return LanguageZHCN
	}
	return runtimeDefaultLanguage
}

func SupportedLanguages() []string {
	runtimeMu.RLock()
	defer runtimeMu.RUnlock()
	return append([]string(nil), runtimeSupportedLanguages...)
}

func IsSupported(language string) bool {
	language = NormalizeLanguage(language)
	if language == "" {
		return false
	}
	runtimeMu.RLock()
	defer runtimeMu.RUnlock()
	for _, supported := range runtimeSupportedLanguages {
		if strings.EqualFold(supported, language) {
			return true
		}
	}
	return false
}

// Configure updates the runtime default language and supported language list.
func Configure(defaultLanguage string, supportedLanguages []string) {
	normalizedSupported := normalizeSupportedLanguages(supportedLanguages)
	normalizedDefault := NormalizeLanguage(defaultLanguage)
	if normalizedDefault == "" {
		normalizedDefault = LanguageZHCN
	}
	if !containsLanguage(normalizedSupported, normalizedDefault) {
		if len(normalizedSupported) > 0 {
			normalizedDefault = normalizedSupported[0]
		} else {
			normalizedDefault = LanguageZHCN
		}
	}

	runtimeMu.Lock()
	runtimeDefaultLanguage = normalizedDefault
	runtimeSupportedLanguages = normalizedSupported
	runtimeMu.Unlock()

	if defaultRegistry != nil {
		defaultRegistry.mu.Lock()
		defaultRegistry.defaultLanguage = normalizedDefault
		defaultRegistry.mu.Unlock()
	}
}

func NormalizeLanguage(language string) string {
	language = strings.TrimSpace(language)
	if language == "" {
		return ""
	}
	language = strings.ReplaceAll(language, "_", "-")
	lower := strings.ToLower(language)
	if strings.HasPrefix(lower, "zh") {
		return LanguageZHCN
	}
	if strings.HasPrefix(lower, "en") {
		return LanguageENUS
	}
	parts := strings.Split(lower, "-")
	if len(parts) > 0 {
		switch parts[0] {
		case "zh":
			return LanguageZHCN
		case "en":
			return LanguageENUS
		}
	}
	return canonicalLanguage(language)
}

func canonicalLanguage(language string) string {
	language = strings.TrimSpace(language)
	if language == "" {
		return ""
	}
	if len(language) <= 2 {
		return strings.ToLower(language)
	}
	return strings.ToLower(language[:2]) + strings.ToUpper(language[2:3]) + strings.ToLower(language[3:])
}

func ResolveLanguage(candidates ...string) string {
	for _, candidate := range candidates {
		if language := NormalizeLanguage(candidate); language != "" && IsSupported(language) {
			return language
		}
	}
	return DefaultLanguage()
}

func ParseAcceptLanguage(header string) string {
	headers := strings.Split(header, ",")
	type item struct {
		language string
		quality  float64
	}
	items := make([]item, 0, len(headers))
	for _, part := range headers {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		quality := 1.0
		if idx := strings.Index(part, ";q="); idx >= 0 {
			value := strings.TrimSpace(part[idx+3:])
			part = strings.TrimSpace(part[:idx])
			if parsed := parseQuality(value); parsed >= 0 {
				quality = parsed
			}
		}
		items = append(items, item{language: NormalizeLanguage(part), quality: quality})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].quality == items[j].quality {
			return items[i].language < items[j].language
		}
		return items[i].quality > items[j].quality
	})
	for _, item := range items {
		if item.language != "" && IsSupported(item.language) {
			return item.language
		}
	}
	return DefaultLanguage()
}

func parseQuality(raw string) float64 {
	switch strings.TrimSpace(raw) {
	case "1", "1.0", "1.00":
		return 1
	case "0", "0.0", "0.00":
		return 0
	default:
		return 1
	}
}

func normalizeSupportedLanguages(languages []string) []string {
	normalized := make([]string, 0, len(languages))
	seen := make(map[string]struct{}, len(languages))
	for _, language := range languages {
		normalizedLanguage := NormalizeLanguage(language)
		if normalizedLanguage == "" {
			continue
		}
		if _, exists := seen[normalizedLanguage]; exists {
			continue
		}
		seen[normalizedLanguage] = struct{}{}
		normalized = append(normalized, normalizedLanguage)
	}
	if len(normalized) == 0 {
		return []string{LanguageZHCN, LanguageENUS}
	}
	return normalized
}

func containsLanguage(languages []string, language string) bool {
	for _, item := range languages {
		if strings.EqualFold(item, language) {
			return true
		}
	}
	return false
}

func HTTPHeaderLanguage(header string) string {
	if language := ParseAcceptLanguage(header); language != "" {
		return language
	}
	return DefaultLanguage()
}

func StatusText(language string, status int) string {
	language = ResolveLanguage(language)
	if language == LanguageENUS {
		return http.StatusText(status)
	}
	return http.StatusText(status)
}

package i18n

import (
	"context"
	"strings"
	"sync"
)

type contextKey string

const languageContextKey contextKey = "goadmin.i18n.language"

var requestLanguages sync.Map

func ContextWithLanguage(ctx context.Context, language string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	language = ResolveLanguage(language)
	if language == "" {
		language = DefaultLanguage()
	}
	return context.WithValue(ctx, languageContextKey, language)
}

func LanguageFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	language, ok := ctx.Value(languageContextKey).(string)
	if !ok || language == "" {
		return "", false
	}
	return ResolveLanguage(language), true
}

func LanguageOrDefault(ctx context.Context) string {
	if language, ok := LanguageFromContext(ctx); ok {
		return language
	}
	return DefaultLanguage()
}

func BindRequestLanguage(requestID, language string) {
	requestID = normalizeRequestID(requestID)
	if requestID == "" {
		return
	}
	language = ResolveLanguage(language)
	if language == "" {
		language = DefaultLanguage()
	}
	requestLanguages.Store(requestID, language)
}

func RequestLanguage(requestID string) string {
	requestID = normalizeRequestID(requestID)
	if requestID == "" {
		return DefaultLanguage()
	}
	if value, ok := requestLanguages.Load(requestID); ok {
		if language, ok := value.(string); ok && language != "" {
			return ResolveLanguage(language)
		}
	}
	return DefaultLanguage()
}

func ClearRequestLanguage(requestID string) {
	requestID = normalizeRequestID(requestID)
	if requestID == "" {
		return
	}
	requestLanguages.Delete(requestID)
}

func normalizeRequestID(requestID string) string {
	return strings.TrimSpace(requestID)
}

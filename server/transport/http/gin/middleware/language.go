package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	corei18n "goadmin/core/i18n"
)

func Language(defaultLanguage string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := requestIDFromGinContext(c)
		language := corei18n.ResolveLanguage(
			c.GetHeader("X-Language"),
			c.Query("lang"),
			c.GetHeader("Accept-Language"),
			defaultLanguage,
		)
		if language == "" {
			language = corei18n.DefaultLanguage()
		}
		corei18n.BindRequestLanguage(requestID, language)
		defer corei18n.ClearRequestLanguage(requestID)

		c.Set("i18n.language", language)
		if c.Request != nil {
			ctx := corei18n.ContextWithLanguage(c.Request.Context(), language)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}

func RequestLanguageFromContext(c *gin.Context) string {
	if value, exists := c.Get("i18n.language"); exists {
		if language, ok := value.(string); ok && strings.TrimSpace(language) != "" {
			return corei18n.ResolveLanguage(language)
		}
	}
	if c != nil && c.Request != nil {
		return corei18n.LanguageOrDefault(c.Request.Context())
	}
	return corei18n.DefaultLanguage()
}

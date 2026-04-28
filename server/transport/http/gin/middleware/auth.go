package middleware

import (
	"fmt"
	"net/http"
	"strings"

	coreauth "goadmin/core/auth"
	coreauthbootstrap "goadmin/core/auth/bootstrap"
	coreauthjwt "goadmin/core/auth/jwt"
	apperrors "goadmin/core/errors"
	corei18n "goadmin/core/i18n"
	"goadmin/core/response"
	coretenant "goadmin/core/tenant"
	coretransport "goadmin/core/transport"
)

func JWTAuth(manager interface {
	ParseAccessToken(string) (*coreauthjwt.Claims, error)
}, revocations coreauthbootstrap.RevocationStore) coretransport.Middleware {
	return func(next coretransport.HandlerFunc) coretransport.HandlerFunc {
		return func(c coretransport.Context) {
			token, ok := bearerToken(c.Header("Authorization"))
			if !ok {
				abortUnauthorized(c, fmt.Errorf("missing bearer token"))
				return
			}

			claims, err := manager.ParseAccessToken(token)
			if err != nil {
				abortUnauthorized(c, err)
				return
			}

			if revocations != nil {
				revoked, err := revocations.IsRevoked(c.RequestContext(), claims.RegisteredClaims.ID)
				if err != nil {
					abortInternal(c, err)
					return
				}
				if revoked {
					abortUnauthorized(c, fmt.Errorf("token revoked"))
					return
				}
			}

			reqCtx := coreauth.ContextWithClaims(c.RequestContext(), claims)
			requestID := requestIDFromTransportContext(c)
			language := corei18n.ResolveLanguage(claims.Language)
			if language == "" {
				language = corei18n.LanguageOrDefault(reqCtx)
			}
			corei18n.BindRequestLanguage(requestID, language)
			reqCtx = corei18n.ContextWithLanguage(reqCtx, language)
			reqCtx = coretenant.ContextWithTenant(reqCtx, coretenant.FromClaims(claims))
			c.SetRequestContext(reqCtx)
			c.Set("auth.claims", claims)
			c.Set("i18n.language", language)
			c.Set("tenant", coretenant.FromClaims(claims))
			next(c)
		}
	}
}

func RequirePermission(authorizer coreauthbootstrap.Authorizer) coretransport.Middleware {
	return func(next coretransport.HandlerFunc) coretransport.HandlerFunc {
		return func(c coretransport.Context) {
			claims, ok := claimsFromContext(c)
			if !ok {
				abortUnauthorized(c, fmt.Errorf("authentication required"))
				return
			}
			if authorizer == nil {
				next(c)
				return
			}
			obj := c.Path()
			if strings.TrimSpace(obj) == "" {
				obj = c.Path()
			}
			allowed, err := authorizer.EnforceClaims(claims, obj, c.Method())
			if err != nil {
				abortInternal(c, err)
				return
			}
			if !allowed {
				abortForbidden(c, fmt.Errorf("permission denied"))
				return
			}
			next(c)
		}
	}
}

func bearerToken(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", false
	}
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	if strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}

func claimsFromContext(c coretransport.Context) (*coreauthjwt.Claims, bool) {
	if value, exists := c.Get("auth.claims"); exists {
		if claims, ok := value.(*coreauthjwt.Claims); ok && claims != nil {
			return claims, true
		}
	}
	if claims, ok := coreauth.ClaimsFromContext(c.RequestContext()); ok {
		return claims, true
	}
	return nil, false
}

func abortUnauthorized(c coretransport.Context, err error) {
	status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeUnauthorized, http.StatusText(http.StatusUnauthorized)), requestIDFromTransportContext(c))
	c.AbortWithStatusJSON(status, body)
}

func abortForbidden(c coretransport.Context, err error) {
	status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeForbidden, http.StatusText(http.StatusForbidden)), requestIDFromTransportContext(c))
	c.AbortWithStatusJSON(status, body)
}

func abortInternal(c coretransport.Context, err error) {
	status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeInternal, http.StatusText(http.StatusInternalServerError)), requestIDFromTransportContext(c))
	c.AbortWithStatusJSON(status, body)
}

func requestIDFromTransportContext(c coretransport.Context) string {
	if value, exists := c.Get("request_id"); exists {
		if requestID, ok := value.(string); ok && requestID != "" {
			return requestID
		}
	}
	return ""
}

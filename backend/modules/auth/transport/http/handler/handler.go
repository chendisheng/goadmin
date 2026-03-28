package handler

import (
	"net/http"
	"time"

	coreauth "goadmin/core/auth"
	coreauthjwt "goadmin/core/auth/jwt"
	apperrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	"goadmin/modules/auth/application/service"
	"goadmin/modules/auth/domain/model"
	authhttpreq "goadmin/modules/auth/transport/http/request"
	authhttpresp "goadmin/modules/auth/transport/http/response"

	"go.uber.org/zap"
)

type Handler struct {
	service *service.Service
	logger  *zap.Logger
}

func New(service *service.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Login(c coretransport.Context) {
	var req authhttpreq.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := response.Failure(apperrors.Wrap(err, apperrors.CodeBadRequest, "invalid login request"), requestID(c))
		c.JSON(status, body)
		return
	}

	session, err := h.service.Login(c.RequestContext(), model.Credentials{Username: req.Username, Password: req.Password})
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}

	c.JSON(http.StatusOK, response.Success(authhttpresp.LoginResponse{
		AccessToken:      session.AccessToken,
		RefreshToken:     session.RefreshToken,
		TokenType:        string(coreauthjwt.TokenTypeAccess),
		ExpiresIn:        int64(time.Until(session.AccessExpiresAt).Seconds()),
		RefreshExpiresIn: int64(time.Until(session.RefreshExpiresAt).Seconds()),
		User: authhttpresp.UserInfo{
			UserID:      session.Identity.UserID,
			TenantID:    session.Identity.TenantID,
			Username:    session.Identity.Username,
			DisplayName: session.Identity.DisplayName,
			Roles:       append([]string(nil), session.Identity.Roles...),
			Permissions: append([]string(nil), session.Identity.Permissions...),
		},
	}, requestID(c)))
}

func (h *Handler) Logout(c coretransport.Context) {
	claims, ok := claimsFromContext(c)
	if !ok {
		status, body := response.Failure(apperrors.New(apperrors.CodeUnauthorized, "authentication required"), requestID(c))
		c.JSON(status, body)
		return
	}
	if err := h.service.Logout(c.RequestContext(), claims); err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(map[string]any{"logged_out": true}, requestID(c)))
}

func (h *Handler) Me(c coretransport.Context) {
	claims, ok := claimsFromContext(c)
	if !ok {
		status, body := response.Failure(apperrors.New(apperrors.CodeUnauthorized, "authentication required"), requestID(c))
		c.JSON(status, body)
		return
	}
	identity, err := h.service.Me(c.RequestContext(), claims)
	if err != nil {
		status, body := response.Failure(err, requestID(c))
		c.JSON(status, body)
		return
	}
	c.JSON(http.StatusOK, response.Success(authhttpresp.UserInfo{
		UserID:      identity.UserID,
		TenantID:    identity.TenantID,
		Username:    identity.Username,
		DisplayName: identity.DisplayName,
		Roles:       append([]string(nil), identity.Roles...),
		Permissions: append([]string(nil), identity.Permissions...),
	}, requestID(c)))
}

func claimsFromContext(c coretransport.Context) (*coreauthjwt.Claims, bool) {
	if claims, ok := coreauth.ClaimsFromContext(c.RequestContext()); ok {
		return claims, true
	}
	if value, exists := c.Get("auth.claims"); exists {
		if claims, ok := value.(*coreauthjwt.Claims); ok && claims != nil {
			return claims, true
		}
	}
	return nil, false
}

func requestID(c coretransport.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

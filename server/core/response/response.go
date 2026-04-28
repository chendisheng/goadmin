package response

import (
	"net/http"
	"strings"
	"time"

	apperrors "goadmin/core/errors"
	corei18n "goadmin/core/i18n"
)

type Envelope struct {
	Code      int    `json:"code"`
	Key       string `json:"key,omitempty"`
	Message   string `json:"msg"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

func Success(data any, requestID string) Envelope {
	key, message := translateMessage(requestID, "common.ok", "ok")
	return Envelope{
		Code:      int(apperrors.CodeOK),
		Key:       key,
		Message:   message,
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Unix(),
	}
}

func translateMessage(requestID, key, fallback string) (string, string) {
	if strings.TrimSpace(key) != "" {
		if translated := corei18n.TranslateRequest(requestID, key); translated != key && strings.TrimSpace(translated) != "" {
			return key, translated
		}
	}
	if strings.TrimSpace(fallback) == "" {
		return strings.TrimSpace(key), strings.TrimSpace(key)
	}
	if strings.TrimSpace(key) == "" {
		return "", fallback
	}
	return key, fallback
}

func translateErrorMessage(requestID string, appErr *apperrors.AppError) (string, string) {
	if appErr == nil {
		return "", ""
	}
	if strings.TrimSpace(appErr.Key) != "" {
		return translateMessage(requestID, appErr.Key, appErr.Message)
	}
	switch appErr.Code {
	case apperrors.CodeBadRequest:
		return translateMessage(requestID, "common.invalid_request", appErr.Message)
	case apperrors.CodeUnauthorized:
		lower := strings.ToLower(strings.TrimSpace(appErr.Message))
		if strings.Contains(lower, "credential") || strings.Contains(lower, "password") {
			return translateMessage(requestID, "auth.invalid_credentials", appErr.Message)
		}
		return translateMessage(requestID, "auth.authentication_required", appErr.Message)
	case apperrors.CodeForbidden:
		return translateMessage(requestID, "common.permission_denied", appErr.Message)
	case apperrors.CodeNotFound:
		return translateMessage(requestID, "common.route_not_found", appErr.Message)
	case apperrors.CodeInternal:
		return translateMessage(requestID, "common.internal_error", appErr.Message)
	default:
		if translated := corei18n.TranslateRequest(requestID, appErr.Message); translated != appErr.Message {
			return appErr.Message, translated
		}
		return "", appErr.Message
	}
}

func Failure(err error, requestID string) (int, Envelope) {
	appErr := apperrors.Resolve(err)
	if appErr == nil {
		appErr = apperrors.New(apperrors.CodeInternal, http.StatusText(http.StatusInternalServerError))
	}

	key, message := translateErrorMessage(requestID, appErr)
	if strings.TrimSpace(message) == "" {
		message = http.StatusText(appErr.Status())
	}

	return appErr.Status(), Envelope{
		Code:      int(appErr.Code),
		Key:       key,
		Message:   message,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Unix(),
	}
}

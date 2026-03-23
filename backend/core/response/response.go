package response

import (
	"net/http"
	"time"

	apperrors "goadmin/core/errors"
)

type Envelope struct {
	Code      int    `json:"code"`
	Message   string `json:"msg"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

func Success(data any, requestID string) Envelope {
	return Envelope{
		Code:      int(apperrors.CodeOK),
		Message:   "ok",
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Unix(),
	}
}

func Failure(err error, requestID string) (int, Envelope) {
	appErr := apperrors.Resolve(err)
	if appErr == nil {
		appErr = apperrors.New(apperrors.CodeInternal, http.StatusText(http.StatusInternalServerError))
	}

	message := appErr.Message
	if message == "" {
		message = http.StatusText(appErr.Status())
	}

	return appErr.Status(), Envelope{
		Code:      int(appErr.Code),
		Message:   message,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Unix(),
	}
}

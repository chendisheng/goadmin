package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type Code int

const (
	CodeOK Code = 200
)

const (
	CodeBadRequest Code = 40000 + iota
)

const (
	CodeUnauthorized Code = 40100
	CodeForbidden    Code = 40300
	CodeNotFound     Code = 40400
	CodeGone         Code = 41000
	CodeConflict     Code = 40900
	CodeInternal     Code = 50000
)

type AppError struct {
	Code       Code   `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"http_status"`
	Err        error  `json:"-"`
}

func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: defaultHTTPStatus(code)}
}

func Wrap(err error, code Code, message string) *AppError {
	if err == nil {
		return New(code, message)
	}
	return &AppError{Code: code, Message: message, HTTPStatus: defaultHTTPStatus(code), Err: err}
}

func Resolve(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Wrap(err, CodeInternal, http.StatusText(http.StatusInternalServerError))
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *AppError) Status() int {
	if e == nil {
		return http.StatusInternalServerError
	}
	if e.HTTPStatus > 0 {
		return e.HTTPStatus
	}
	return defaultHTTPStatus(e.Code)
}

func (e *AppError) WithCause(err error) *AppError {
	if e == nil {
		return nil
	}
	e.Err = err
	return e
}

func defaultHTTPStatus(code Code) int {
	switch code {
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeGone:
		return http.StatusGone
	case CodeConflict:
		return http.StatusConflict
	case CodeOK:
		return http.StatusOK
	default:
		return http.StatusInternalServerError
	}
}

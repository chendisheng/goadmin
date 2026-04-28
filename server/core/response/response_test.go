package response

import (
	"errors"
	"testing"

	apperrors "goadmin/core/errors"
	corei18n "goadmin/core/i18n"
)

func TestSuccessUsesRequestLanguage(t *testing.T) {
	t.Parallel()

	requestID := "resp-success"
	corei18n.BindRequestLanguage(requestID, corei18n.LanguageENUS)
	t.Cleanup(func() { corei18n.ClearRequestLanguage(requestID) })

	envelope := Success(map[string]string{"ok": "true"}, requestID)
	if envelope.Key != "common.ok" {
		t.Fatalf("Success() key = %q, want common.ok", envelope.Key)
	}
	if envelope.Message != "ok" {
		t.Fatalf("Success() message = %q, want ok", envelope.Message)
	}
	if envelope.RequestID != requestID {
		t.Fatalf("Success() request_id = %q, want %q", envelope.RequestID, requestID)
	}
}

func TestFailureTranslatesErrorMessageByRequestLanguage(t *testing.T) {
	t.Parallel()

	requestID := "resp-failure"
	corei18n.BindRequestLanguage(requestID, corei18n.LanguageENUS)
	t.Cleanup(func() { corei18n.ClearRequestLanguage(requestID) })

	status, envelope := Failure(apperrors.New(apperrors.CodeUnauthorized, "invalid credentials"), requestID)
	if status != 401 {
		t.Fatalf("Failure() status = %d, want 401", status)
	}
	if envelope.Key != "auth.invalid_credentials" {
		t.Fatalf("Failure() key = %q, want auth.invalid_credentials", envelope.Key)
	}
	if envelope.Message != "invalid username or password" {
		t.Fatalf("Failure() message = %q, want translated invalid credentials", envelope.Message)
	}
}

func TestFailurePreservesExplicitAppErrorKey(t *testing.T) {
	t.Parallel()

	requestID := "resp-keyed"
	corei18n.BindRequestLanguage(requestID, corei18n.LanguageENUS)
	t.Cleanup(func() { corei18n.ClearRequestLanguage(requestID) })

	status, envelope := Failure(apperrors.NewWithKey(apperrors.CodeInternal, "codegen.download.service_required", "download service is required"), requestID)
	if status != 500 {
		t.Fatalf("Failure() status = %d, want 500", status)
	}
	if envelope.Key != "codegen.download.service_required" {
		t.Fatalf("Failure() key = %q, want codegen.download.service_required", envelope.Key)
	}
	if envelope.Message != "download service is not configured" {
		t.Fatalf("Failure() message = %q, want translated codegen message", envelope.Message)
	}
}

func TestFailureFallsBackToAppErrorMessageWhenNoTranslationExists(t *testing.T) {
	t.Parallel()

	requestID := "resp-fallback"
	corei18n.BindRequestLanguage(requestID, corei18n.LanguageZHCN)
	t.Cleanup(func() { corei18n.ClearRequestLanguage(requestID) })

	status, envelope := Failure(errors.New("unexpected condition"), requestID)
	if status != 500 {
		t.Fatalf("Failure() status = %d, want 500", status)
	}
	if envelope.Key != "common.internal_error" {
		t.Fatalf("Failure() key = %q, want common.internal_error", envelope.Key)
	}
	if envelope.Message == "" {
		t.Fatal("Failure() message should not be empty")
	}
}

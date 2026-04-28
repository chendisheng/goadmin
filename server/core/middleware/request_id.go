package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

const (
	RequestIDHeader     = "X-Request-ID"
	RequestIDContextKey = "request_id"
)

func NormalizeRequestID(value string) string {
	return strings.TrimSpace(value)
}

func GenerateRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fallbackRequestID()
	}
	return hex.EncodeToString(b[:])
}

func fallbackRequestID() string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(strings.TrimSpace("fallback-request-id")), " ", "-"), "_", "-")
}

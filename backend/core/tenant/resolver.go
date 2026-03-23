package tenant

import (
	"context"
	"errors"
	"strings"
)

var ErrTenantMismatch = errors.New("tenant mismatch")

func ResolveTenantID(ctx context.Context, requested string) (string, error) {
	if !Enabled() {
		return "", nil
	}
	requested = strings.TrimSpace(requested)
	if tenantID, ok := TenantIDFromContext(ctx); ok && tenantID != "" {
		if requested != "" && requested != tenantID {
			return "", ErrTenantMismatch
		}
		return tenantID, nil
	}
	return requested, nil
}

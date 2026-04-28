package tenant

import (
	"context"
	"strings"
)

type contextKey string

const tenantContextKey contextKey = "goadmin.tenant"

func ContextWithTenant(ctx context.Context, value Tenant) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if !Enabled() {
		return ctx
	}
	return context.WithValue(ctx, tenantContextKey, value.Clone())
}

func ContextWithTenantID(ctx context.Context, tenantID string) context.Context {
	if !Enabled() {
		if ctx == nil {
			return context.Background()
		}
		return ctx
	}
	return ContextWithTenant(ctx, Tenant{ID: tenantID})
}

func TenantFromContext(ctx context.Context) (Tenant, bool) {
	if !Enabled() {
		return Tenant{}, false
	}
	if ctx == nil {
		return Tenant{}, false
	}
	value, ok := ctx.Value(tenantContextKey).(Tenant)
	if !ok {
		return Tenant{}, false
	}
	value = value.Clone()
	if strings.TrimSpace(value.ID) == "" && strings.TrimSpace(value.Code) == "" && strings.TrimSpace(value.Status) == "" {
		return Tenant{}, false
	}
	return value, true
}

func TenantIDFromContext(ctx context.Context) (string, bool) {
	if !Enabled() {
		return "", false
	}
	tenant, ok := TenantFromContext(ctx)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(tenant.ID), true
}

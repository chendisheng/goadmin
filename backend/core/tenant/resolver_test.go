package tenant

import (
	"context"
	"testing"
)

func TestContextWithTenantRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := ContextWithTenant(context.Background(), Tenant{ID: "tenant-a", Code: "alpha", Status: "active"})
	tenant, ok := TenantFromContext(ctx)
	if !ok {
		t.Fatal("expected tenant in context")
	}
	if tenant.ID != "tenant-a" || tenant.Code != "alpha" || tenant.Status != "active" {
		t.Fatalf("unexpected tenant: %+v", tenant)
	}
	if got, ok := TenantIDFromContext(ctx); !ok || got != "tenant-a" {
		t.Fatalf("TenantIDFromContext() = %q, %v; want tenant-a, true", got, ok)
	}
}

func TestResolveTenantID(t *testing.T) {
	t.Parallel()

	ctx := ContextWithTenantID(context.Background(), "tenant-a")
	got, err := ResolveTenantID(ctx, "")
	if err != nil {
		t.Fatalf("ResolveTenantID returned error: %v", err)
	}
	if got != "tenant-a" {
		t.Fatalf("ResolveTenantID() = %q, want tenant-a", got)
	}

	if _, err := ResolveTenantID(ctx, "tenant-b"); err == nil {
		t.Fatal("expected tenant mismatch error")
	}
}

func TestTenantDisabledIgnoresContextAndClaims(t *testing.T) {
	prev := Enabled()
	SetEnabled(false)
	t.Cleanup(func() { SetEnabled(prev) })

	ctx := ContextWithTenantID(context.Background(), "tenant-a")
	if _, ok := TenantIDFromContext(ctx); ok {
		t.Fatal("expected tenant context to be ignored when disabled")
	}
	got, err := ResolveTenantID(ctx, "tenant-b")
	if err != nil {
		t.Fatalf("ResolveTenantID returned error: %v", err)
	}
	if got != "" {
		t.Fatalf("ResolveTenantID() = %q, want empty string when tenant disabled", got)
	}
}

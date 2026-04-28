package tenant

import (
	"strings"
	"sync/atomic"

	corejwt "goadmin/core/auth/jwt"
)

var enabled atomic.Bool

func init() {
	enabled.Store(true)
}

type Tenant struct {
	ID     string
	Code   string
	Status string
}

func FromClaims(claims *corejwt.Claims) Tenant {
	if !Enabled() {
		return Tenant{}
	}
	if claims == nil {
		return Tenant{}
	}
	return Tenant{
		ID: strings.TrimSpace(claims.TenantID),
	}
}

func SetEnabled(value bool) {
	enabled.Store(value)
}

func Enabled() bool {
	return enabled.Load()
}

func (t Tenant) Clone() Tenant {
	return Tenant{
		ID:     strings.TrimSpace(t.ID),
		Code:   strings.TrimSpace(t.Code),
		Status: strings.TrimSpace(t.Status),
	}
}

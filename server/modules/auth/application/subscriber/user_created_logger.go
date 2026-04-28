package subscriber

import (
	"context"
	"strings"

	coreevent "goadmin/core/event"
	userevent "goadmin/modules/user/application/event"

	"go.uber.org/zap"
)

type UserCreatedLogger struct {
	logger *zap.Logger
}

func NewUserCreatedLogger(logger *zap.Logger) *UserCreatedLogger {
	return &UserCreatedLogger{logger: logger}
}

func (s *UserCreatedLogger) Handle(_ context.Context, evt coreevent.Event) error {
	if s == nil || s.logger == nil {
		return nil
	}
	created, ok := evt.(userevent.Created)
	if !ok {
		return nil
	}

	roles := strings.Join(created.RoleCodes, ",")
	if roles == "" {
		roles = "-"
	}

	s.logger.Info("user.created event received",
		zap.String("user_id", created.UserID),
		zap.String("tenant_id", created.TenantID),
		zap.String("username", created.Username),
		zap.String("display_name", created.DisplayName),
		zap.String("role_codes", roles),
		zap.Time("created_at", created.CreatedAt),
	)
	return nil
}

package event

import "time"

const CreatedTopic = "user.created"

type Created struct {
	UserID      string
	TenantID    string
	Username    string
	DisplayName string
	RoleCodes   []string
	CreatedAt   time.Time
}

func (Created) Topic() string {
	return CreatedTopic
}

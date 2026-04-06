package model

type Status string

const (
	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusEnabled, StatusDisabled:
		return true
	default:
		return false
	}
}

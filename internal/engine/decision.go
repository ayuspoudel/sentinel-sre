package engine

import "time"

type Decision struct {
	GuardName string
	Allowed   bool
	Phase     string
	Reason    string
	Timestamp time.Time
}

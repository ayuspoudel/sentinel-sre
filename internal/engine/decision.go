package engine

import "time"

type Decision struct {
	GuardName string
	Allowed   bool
	Phase     string
	Reason    string
	Timestamp time.Time
}

func (d *Decision) Update(allowed bool, phase string, reason string) {
	d.Allowed = allowed
	d.Phase = phase
	d.Reason = reason
	d.Timestamp = time.Now()
	return
}

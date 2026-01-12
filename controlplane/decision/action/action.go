package action

import "time"

type Type string

const (
	Allow    Type = "allow"
	Block    Type = "block"
	Rollback Type = "rollback"
	Promote  Type = "promote"
)

type Action struct {
	GuardName string
	Type      Type
	Phase     string
	Reason    string
	Timestamp time.Time
}

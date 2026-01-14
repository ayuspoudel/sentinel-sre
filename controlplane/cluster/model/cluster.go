package model

import "time"

type Cluster struct {
	Name          string            `json:"name" db:"name"`
	CredentialRef string            `json:"credential_ref" db:"credential_ref"`
	Labels        map[string]string `json:"labels" db:"labels"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
}

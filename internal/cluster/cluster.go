package cluster

import "time"

// Cluster struct will help us easily deal with json validations, encode and decode into the struct object
type Cluster struct {
	Name      string    `json:"name"`
	AgentID   string    `json:"agent_id"`
	Version   string    `json:"version"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}

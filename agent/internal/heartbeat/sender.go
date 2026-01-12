package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/logging"
)

type Payload struct {
	AgentID      string    `json:"agent_id"`
	ClusterName  string    `json:"cluster_name"`
	AgentVersion string    `json:"agent_version"`
	Timestamp    time.Time `json:"timestamp"`
}

func Send(ctx context.Context, controllerURL, agentId, clusterName, agentVersion string) error {
	log := logging.From(ctx)

	body, _ := json.Marshal(Payload{
		AgentID:      agentId,
		ClusterName:  clusterName,
		AgentVersion: agentVersion,
		Timestamp:    time.Now().UTC(),
	})

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		controllerURL+"/api/v1/agent/heartbeat",
		bytes.NewBuffer(body),
	)
	if err != nil {
		log.Warn("heartbeat request build failed", "error", err)
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn("heartbeat failed", "error", err)
		return err
	}
	defer resp.Body.Close()

	log.Info("heartbeat sent", "status", resp.StatusCode)
	return nil
}

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
	Cluster      string    `json:"cluster"`
	AgentId      string    `json:"agentId"`
	AgentVersion string    `json:"agentVersion"`
	Timestamp    time.Time `json:"timestamp"`
}

func Send(ctx context.Context, controllerURL, agentId, clusterName, agentVersion string) error {
	log := logging.From(ctx)

	body, _ := json.Marshal(Payload{
		Cluster:      clusterName,
		AgentId:      agentId,
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

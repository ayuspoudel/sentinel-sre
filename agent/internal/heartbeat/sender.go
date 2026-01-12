package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func Send(ctx context.Context, controllerURL, agentID, clusterName, agentVersion string) error {
	log := logging.From(ctx)

	body, err := json.Marshal(Payload{
		AgentID:      agentID,
		ClusterName:  clusterName,
		AgentVersion: agentVersion,
		Timestamp:    time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		controllerURL+"/api/v1/agent/heartbeat",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat rejected: status=%d", resp.StatusCode)
	}

	log.Info("heartbeat sent", "status", resp.StatusCode)
	return nil
}

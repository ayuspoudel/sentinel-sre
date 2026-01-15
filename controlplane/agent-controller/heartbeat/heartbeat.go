package heartbeat

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/logging"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/models/clusterStatusModel"
	"github.com/ayuspoudel/sentinel-sre/controlplane/agent-controller/store/clusterStatus"
)

type Handler struct {
	store *clusterStatus.StatusStore
}

func NewHandler(store *clusterStatus.StatusStore) *Handler {
	return &Handler{store: store}
}

type request struct {
	AgentID      string `json:"agent_id"`
	ClusterName  string `json:"cluster_name"`
	AgentVersion string `json:"agent_version"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logging.From(r.Context())

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("invalid heartbeat payload", "error", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if req.ClusterName == "" || req.AgentID == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	now := time.Now()
	healthy := true

	st := &clusterStatusModel.ClusterStatus{
		ClusterName:        req.ClusterName,
		AgentInstalled:     ptr(true),
		AgentHealthy:       &healthy,
		AgentLastHeartbeat: &now,
		AgentVersion:       &req.AgentVersion,
	}
	log.Info("agent heartbeat received", "cluster", req.ClusterName, "agent_id", req.AgentID, "agent_version", req.AgentVersion)

	if err := h.store.Upsert(r.Context(), st); err != nil {
		log.Error("failed to persist heartbeat", "error", err)
		http.Error(w, "failed to persist heartbeat", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func ptr[T any](v T) *T {
	return &v
}

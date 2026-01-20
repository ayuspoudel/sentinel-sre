package server

import (
	"net/http"

	"github.com/ayuspoudel/sentinel-sre/controlplane/api-server/config"
	"github.com/ayuspoudel/sentinel-sre/controlplane/api-server/proxy"
)

func registerRoutes(mux *http.ServeMux, cfg config.Config) {
	// system & lifecycle
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/readyz", readyz)

	// Cluster Registry
	mux.Handle("/v1/clusters/", proxy.To(cfg.ClusterRegistryURL))

	// Policy Registry
	mux.Handle("/v1/policies/", proxy.To(cfg.PolicyRegistryURL))

	// Agent Controller (heartbeats, presence, status)
	mux.Handle("/v1/agents/", proxy.To(cfg.AgentControllerURL))
}

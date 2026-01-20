package server

import (
	"net/http"
)

// healthz reports basic liveness.
// If this fails, the process is unhealthy.
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// readyz reports readiness to serve traffic.
// Later this can include dependency checks.
func readyz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

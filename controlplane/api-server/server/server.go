package server

import (
	"net/http"

	"github.com/ayuspoudel/sentinel-sre/controlplane/api-server/config"
)

func New(cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(mux, cfg)
	return withMiddleware(mux)
}

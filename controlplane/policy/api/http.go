package api

import (
	"net/http"
	"strings"

	"github.com/go-openapi/runtime/middleware"
)

func NewServer(addr string, h *Handler) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.Health)

	docsDir := "cmd/sentinel-policy-registry/docs"

	opts := middleware.RedocOpts{
		SpecURL:  "/openapi.yaml",
		Path:     "/docs",
		RedocURL: "https://rebilly.github.io/ReDoc/releases/latest/redoc.min.js",
	}
	swaggerOpts := middleware.SwaggerUIOpts{
		SpecURL: "/openapi.yaml",
		Path:    "/swagger",
	}

	swaggerHandler := middleware.SwaggerUI(swaggerOpts, nil)
	mux.Handle("/swagger", swaggerHandler)
	mux.Handle("/swagger/", swaggerHandler)

	redocHandler := middleware.Redoc(opts, nil)
	mux.Handle("/docs", redocHandler)
	mux.Handle("/docs/", redocHandler)

	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, docsDir+"/openapi.yaml")
	})

	mux.HandleFunc("/v1/policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.ListPolicies(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/v1/policies/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/policies/")
		parts := strings.Split(path, "/")

		if len(parts) == 0 || parts[0] == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		name := parts[0]

		if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodGet {
			h.GetStatus(w, r, name)
			return
		}

		switch r.Method {
		case http.MethodPut:
			h.ApplyPolicy(w, r, name)
		case http.MethodGet:
			h.GetPolicy(w, r, name)
		case http.MethodDelete:
			h.DeletePolicy(w, r, name)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

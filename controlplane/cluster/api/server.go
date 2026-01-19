package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
)

type Server struct {
	addr    string
	handler *Handler
}

func NewServer(addr string, handler *Handler) *Server {
	return &Server{addr: addr, handler: handler}
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	docsDir := "cmd/sentinel-cluster-registry/docs"

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
		http.ServeFile(w, r, docsDir+"/openapi1.yaml")
	})

	mux.HandleFunc("/health", s.handler.Health)
	mux.HandleFunc("/v1/clusters/register", s.handler.Register)
	mux.HandleFunc("/v1/clusters/register-with-credentials", s.handler.RegisterWithCredentials)
	mux.HandleFunc("/v1/clusters/list", s.handler.List)
	mux.HandleFunc("/v1/clusters/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handler.GetByName(w, r)
		case http.MethodDelete:
			s.handler.DeleteByName(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	httpServer := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		log.Println("sentinel cluster registry listening on", s.addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down cluster registry")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpServer.Shutdown(shutdownCtx)
}

package api

import (
	"context"
	"log"
	"net/http"
	"time"
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

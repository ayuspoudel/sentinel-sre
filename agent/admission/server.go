package admission

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Server struct {
	mux  *http.ServeMux
	http *http.Server
}

func NewServer(addr string) *Server {
	mux := http.NewServeMux()

	s := &Server{
		mux: mux,
		http: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
	}

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/validate", s.handleValidate)

	return s
}

func (s *Server) Start() error {
	log.Printf("starting admission server on %s", s.http.Addr)
	return s.http.ListenAndServeTLS("/tls/tls.crt", "/tls/tls.key")
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("stopping admission server")
	return s.http.Shutdown(ctx)
}

package cluster

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	addr    string
	handler *Handler
}

func NewServer(addr string, handler *Handler) *Server {
	return &Server{addr: addr, handler: handler}
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handler.Health)
	mux.HandleFunc("/clusters", s.handler.Register)
	mux.HandleFunc("/clusters/list", s.handler.List)
	mux.HandleFunc("/clusters/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.handler.GetByName(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			s.handler.DeleteByName(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	httpServer := &http.Server{Addr: s.addr, Handler: mux}

	go func() {
		log.Println("sentinel cluster registry listening on", s.addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("shutting down cluster registry")
	return httpServer.Shutdown(ctx)
}

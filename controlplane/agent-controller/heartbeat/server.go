package heartbeat

import (
	"context"
	"net/http"
	"time"

	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/logging"
	"github.com/ayuspoudel/sentinel-sre/controlplane/controller/agent/status"
)

type Server struct {
	addr  string
	store *status.StatusStore
}

func NewServer(addr string, store *status.StatusStore) *Server {
	return &Server{
		addr:  addr,
		store: store,
	}
}

func (s *Server) Start(ctx context.Context) error {
	log := logging.From(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.Handle("/api/v1/agent/heartbeat", NewHandler(s.store))

	srv := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Info("agent controller heartbeat server starting", "addr", s.addr)
	return srv.ListenAndServe()
}

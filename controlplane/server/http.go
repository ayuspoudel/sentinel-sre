package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ayuspoudel/sentinel-sre/internal/action"
	"github.com/ayuspoudel/sentinel-sre/internal/cluster"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	mux        *http.ServeMux
	httpServer *http.Server

	// store is a read-only reference to Sentinel's action store.
	// The server NEVER mutates state. It only exposes intent.
	store        *action.Store
	clusterStore *cluster.Store
}

/*
@ayuspoudel
This will be our http utility function which allows us to initialize a http server.
It takes in port number and an action store reference and returns a Server struct
with a custom handler initialized with:
- 5s read timeout
- 5s write timeout
- 30s idle timeout

The server does not make decisions.
It only exposes Sentinel's current intent in a read-only manner.
*/
func New(addr string, store *action.Store) *Server {
	/*
		@ayuspoudel
		It is always safe to register a custom mux (http multiplexer) instead of using
		DefaultMux() or http.HandleFunc() which internally uses a global mux.
		The problem with such mux is that when packages try to register endpoints in
		your go program, it might be able to register an endpoint you never wanted.
		For such reasons, it is safer to use a custom mux.
	*/
	mux := http.NewServeMux()

	/*
		@ayuspoudel
		Registers a new /health endpoint which is very important for servers of this kind.
		It indicates the liveness of the process and responds with "ok".
		This endpoint is intentionally dumb and fast.
	*/
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	/*
		@ayuspoudel
		Registers /metrics endpoint to expose Prometheus metrics.
		This is required for observability of Sentinel itself.
	*/
	mux.Handle("/metrics", promhttp.Handler())

	s := &Server{
		mux:   mux,
		store: store,
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
	}

	/*
		@ayuspoudel
		Action API endpoints.
		These endpoints expose Sentinel's current intent.
		No mutation or triggering is allowed via HTTP.
	*/
	mux.HandleFunc("/actions", s.handleActions)
	mux.HandleFunc("/actions/", s.handleActionByGuard)

	/*
		@ayuspoudel
		Cluster registration endpoints.
	*/
	s.clusterStore = cluster.NewStore()
	mux.HandleFunc("/clusters/register", cluster.RegisterHandler(s.clusterStore))
	mux.HandleFunc("/clusters", cluster.ListHandler(s.clusterStore))
	mux.HandleFunc("/clusters/", cluster.GetByNameHandler(s.clusterStore))
	return s
}

/*
	@ayuspoduel
	Below are methods for struct Server. These will be used to start and shut down
	the http server.
*/

func (s *Server) Start() error {
	log.Printf("starting http server on %s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("shutting down server")
	return s.httpServer.Shutdown(ctx)
}

/*
	@ayuspoudel
	Utility function to write JSON responses.
	This keeps response formatting consistent across handlers.
*/

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

/*
	@ayuspoudel
	GET /actions
	Returns the latest action for all active guards.
	This endpoint is expected to be consumed by CI/CD systems
	or humans inspecting Sentinel state.
*/

func (s *Server) handleActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actions := s.store.List()
	writeJSON(w, http.StatusOK, actions)
}

/*
	@ayuspoudel
	GET /actions/{guardName}
	Returns the latest action for a single guard.
	If the guard does not exist, 404 is returned.
*/

func (s *Server) handleActionByGuard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	guard := strings.TrimPrefix(r.URL.Path, "/actions/")
	if guard == "" {
		http.NotFound(w, r)
		return
	}

	a, ok := s.store.Get(guard)
	if !ok {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, http.StatusOK, a)
}

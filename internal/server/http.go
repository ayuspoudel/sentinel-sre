package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ayuspoudel/sentinel-sre/internal/engine"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	mux        *http.ServeMux
	httpServer *http.Server
	engine     *engine.Engine
}

/*
@ayuspoudel
This will be our http utility function which allows us to initialize a http server.
It takes in port number and engine reference and returns a Server struct with a
custom handler initialized with:
- 5s read timeout
- 5s write timeout
- 30s idle timeout

The engine reference is required so that this server can expose Sentinel's
decision state over HTTP in a read-only manner.
*/
func New(addr string, eng *engine.Engine) *Server {
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
		mux:    mux,
		engine: eng,
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
		Decision API endpoints.
		These endpoints expose Sentinel's current decisions in a read-only fashion.
		No mutation or triggering is allowed via HTTP.
	*/
	mux.HandleFunc("/decisions", s.handleDecisions)
	mux.HandleFunc("/decisions/", s.handleDecisionByGuard)

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
	Since we want other packages in this application to be able to register paths,
	we expose controlled helper functions instead of giving direct access to mux.
*/

func (s *Server) Handle(path string, handler http.Handler) {
	s.mux.Handle(path, handler)
}

func (s *Server) HandleFunc(path string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(path, handler)
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
	GET /decisions
	Returns the latest decision for all active guards.
	This endpoint is expected to be consumed by CI/CD systems
	or humans inspecting Sentinel state.
*/

func (s *Server) handleDecisions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	decisions := s.engine.Decisions()
	writeJSON(w, http.StatusOK, decisions)
}

/*
	@ayuspoudel
	GET /decisions/{guardName}
	Returns the latest decision for a single guard.
	If the guard does not exist, 404 is returned.
*/

func (s *Server) handleDecisionByGuard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	guard := strings.TrimPrefix(r.URL.Path, "/decisions/")
	if guard == "" {
		http.NotFound(w, r)
		return
	}

	d, ok := s.engine.DecisionFor(guard)
	if !ok {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, http.StatusOK, d)
}

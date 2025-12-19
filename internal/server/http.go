package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	mux        *http.ServeMux
	httpServer *http.Server
}

/*
@ayuspoudel
This will be our http utility function which allows us to initalize a http server
It takes in port number and returns a Server struct with a custom handler initialized
With 5s read, write and 30s idle timeout
*/
func New(addr string) *Server {
	/*
		@ayuspoudel
		It is always safe to register a custom mux (http multiplexer) instead of using DefaultMux() or http.HandleFunc()
		which internally uses a global mux. The problem with such mux is that when packages try to register endpoints in
		your go program, it might be able to register an endpoint. Boom, you get an endpoint you never wanted. For such
		reasons, it is safer to use a custom mux.
	*/
	mux := http.NewServeMux()

	/*
		@ayuspoudel
		Registers a new /health endpoint which is very important for servers of this kind
		It will indicate the current status of server by responing with "ok"
		This can be useful for automation systems to parse and what not.
	*/
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.Handle("/metrics", promhttp.Handler())
	return &Server{
		mux: mux,
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
	}
}

/*
	@ayuspoduel
	Below are methods for struct Server. These will be used to start and shut down
	the http servers
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
	Since we want other packages in this application to be able to register path
	we need the below functions. This gives us a controlled way to add routes.
*/

func (s *Server) Handle(path string, handler http.Handler) {
	s.mux.Handle(path, handler)
}

func (s *Server) HandleFunc(path string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(path, handler)
}

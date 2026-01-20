package server

import (
	"log"
	"net/http"
	"time"
)

// withMiddleware wraps the root handler with all global middleware.
// Order matters: outermost first.
func withMiddleware(h http.Handler) http.Handler {
	return requestLogging(h)
}

// requestLogging logs basic request metadata.
// This is intentionally simple and synchronous.
func requestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
}

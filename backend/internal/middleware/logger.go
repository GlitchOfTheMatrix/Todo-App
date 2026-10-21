package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logger logs every request with method, path, status, duration, and client IP
// in a structured format suitable for log aggregation at scale.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		ip := extractIP(r)

		log.Printf(
			`{"method":"%s","path":"%s","status":%d,"duration_ms":%d,"ip":"%s","user_agent":"%s"}`,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration.Milliseconds(),
			ip,
			r.UserAgent(),
		)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

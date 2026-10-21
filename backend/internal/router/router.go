package router

import (
	"net/http"
	"time"

	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/config"
	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/handler"
	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(authHandler *handler.AuthHandler, taskHandler *handler.TaskHandler, cfg *config.Config, db *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	// Health check – includes a DB ping so load balancers can detect unhealthy instances
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx := r.Context()
		if err := db.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","reason":"database unreachable"}`))
			return
		}

		stats := db.Stat()
		w.Write([]byte(`{"status":"ok","db_pool":{"total_conns":` +
			itoa(int(stats.TotalConns())) +
			`,"idle_conns":` + itoa(int(stats.IdleConns())) +
			`,"acquired_conns":` + itoa(int(stats.AcquiredConns())) +
			`}}`,
		))
	})

	// Auth routes (public)
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)

	// Auth routes (protected)
	authMux := http.NewServeMux()
	authMux.HandleFunc("GET /api/auth/me", authHandler.Me)

	// Task routes (protected)
	authMux.HandleFunc("GET /api/tasks", taskHandler.GetAll)
	authMux.HandleFunc("GET /api/tasks/{id}", taskHandler.GetOne)
	authMux.HandleFunc("POST /api/tasks", taskHandler.Create)
	authMux.HandleFunc("PATCH /api/tasks/{id}", taskHandler.Update)
	authMux.HandleFunc("DELETE /api/tasks/{id}", taskHandler.Delete)

	// Wrap protected routes with auth middleware
	mux.Handle("/api/", middleware.AuthMiddleware(cfg.JWTSecret)(authMux))

	// Build middleware chain (applied bottom-up)
	var h http.Handler = mux

	// 1. Request timeout – prevents goroutine leaks
	h = middleware.RequestTimeout(30 * time.Second)(h)

	// 2. Request body size limit
	h = middleware.MaxBodySize(cfg.MaxRequestBodySize)(h)

	// 3. Rate limiting
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	h = rateLimiter.Middleware(h)

	// 4. Structured logging
	h = middleware.Logger(h)

	// 5. CORS
	h = middleware.CORS(&middleware.CORSConfig{AllowedOrigins: cfg.AllowedOrigins})(h)

	return h
}

// itoa converts int to string without importing strconv in a hot path.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

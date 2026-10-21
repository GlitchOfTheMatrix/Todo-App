package router

import (
	"net/http"

	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/handler"
	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/middleware"
)

func New(authHandler *handler.AuthHandler, taskHandler *handler.TaskHandler, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
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
	mux.Handle("/api/", middleware.AuthMiddleware(jwtSecret)(authMux))

	// Apply global middleware
	var h http.Handler = mux
	h = middleware.Logger(h)
	h = middleware.CORS(h)

	return h
}

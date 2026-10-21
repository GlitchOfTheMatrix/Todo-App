package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/config"
	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/database"
	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/handler"
	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/repository"
	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/router"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (ignore error if not present)
	_ = godotenv.Load()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database with pool tuning
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.DatabaseUrl, &database.PoolConfig{
		MaxConns:          cfg.DBMaxConns,
		MinConns:          cfg.DBMinConns,
		MaxConnLifetime:   cfg.DBMaxConnLifetime,
		MaxConnIdleTime:   cfg.DBMaxConnIdleTime,
		HealthCheckPeriod: cfg.DBHealthCheckPeriod,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Printf("Connected to database (pool: max=%d, min=%d)", cfg.DBMaxConns, cfg.DBMinConns)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userRepo, cfg.JWTSecret)
	taskHandler := handler.NewTaskHandler(taskRepo)

	// Setup router with full middleware stack
	r := router.New(authHandler, taskHandler, cfg, db)

	// Create server with production timeouts
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
	}()

	log.Printf("Server starting on port %s (rate limit: %d rps, burst: %d)", cfg.Port, cfg.RateLimitRPS, cfg.RateLimitBurst)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped")
}

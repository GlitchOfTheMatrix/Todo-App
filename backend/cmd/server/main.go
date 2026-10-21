package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Connect to database
	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userRepo, cfg.JWTSecret)
	taskHandler := handler.NewTaskHandler(taskRepo)

	// Setup router
	r := router.New(authHandler, taskHandler, cfg.JWTSecret)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped")
}

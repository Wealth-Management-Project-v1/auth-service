package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/wmp/auth-service/internal/config"
	"github.com/wmp/auth-service/internal/database"
	"github.com/wmp/auth-service/internal/handler"
	"github.com/wmp/auth-service/internal/repository"
	"github.com/wmp/auth-service/internal/service"
	"github.com/wmp/auth-service/migrations"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := config.Load()
	ctx := context.Background()

	// Database
	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Run migrations
	migrationSQL, err := migrations.FS.ReadFile("001_create_auth_users.sql")
	if err != nil {
		slog.Error("failed to read migration file", "error", err)
		os.Exit(1)
	}
	if err := database.RunMigrations(ctx, pool, string(migrationSQL)); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations completed")

	// Dependencies
	userRepo := repository.NewUserRepository(pool)
	jwtService := service.NewJWTService(cfg)
	portfolioClient := service.NewPortfolioClient(cfg)
	authService := service.NewAuthService(userRepo, jwtService, portfolioClient)
	authHandler := handler.NewAuthHandler(authService)

	// Router
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(handler.LoggingMiddleware)
	r.Use(handler.RecoveryMiddleware)

	r.Get("/health", handler.HealthHandler())
	r.Post("/api/v1/auth/register", authHandler.Register)
	r.Post("/api/v1/auth/login", authHandler.Login)

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		slog.Info("auth-service starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
	slog.Info("server stopped")
}

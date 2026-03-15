package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/go-sql-driver/mysql"

	"clientes/internal/application"
	"clientes/internal/domain/repository"
	domainservice "clientes/internal/domain/service"
	"clientes/internal/infrastructure/config"
	httpinfra "clientes/internal/infrastructure/http"
	"clientes/internal/infrastructure/http/middleware"
	"clientes/internal/infrastructure/persistence/memory"
	mysqlrepo "clientes/internal/infrastructure/persistence/mysql"
)

func main() {
	cfg := config.LoadConfig()

	// -----------------------------------------------------------------------
	// Repositorio
	// -----------------------------------------------------------------------
	var repo repository.ClienteRepository

	if cfg.UseMockDB {
		slog.Info("usando repositorio in-memory (mock)")
		repo = memory.NewClienteRepo()
	} else {
		slog.Info("conectando a MySQL")
		db, err := sql.Open("mysql", cfg.DBDsn)
		if err != nil {
			slog.Error("no se pudo abrir conexion MySQL", "error", err)
			os.Exit(1)
		}
		defer db.Close()

		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		if err := db.Ping(); err != nil {
			slog.Error("no se pudo conectar a MySQL", "error", err)
			os.Exit(1)
		}
		slog.Info("conectado a MySQL")
		repo = mysqlrepo.NewClienteRepo(db)
	}

	// -----------------------------------------------------------------------
	// Wiring de dependencias
	// -----------------------------------------------------------------------
	domainSvc := domainservice.NewClienteDomainService(repo)
	appSvc := application.NewClienteAppService(repo, domainSvc)
	handler := httpinfra.NewHandler(appSvc)

	// -----------------------------------------------------------------------
	// Middleware stack
	// -----------------------------------------------------------------------
	jwtSecret := cfg.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-me"
	}
	authMw := middleware.JWTAuth(jwtSecret)
	rl := middleware.NewRateLimiter(30, time.Minute)

	// -----------------------------------------------------------------------
	// Router
	// -----------------------------------------------------------------------
	r := chi.NewRouter()
	r.Use(middleware.CORS)

	httpinfra.SetupRoutes(
		r,
		handler,
		authMw,
		middleware.Logging,
		middleware.RequestID,
		rl.Middleware,
		middleware.RequireRole,
	)

	// -----------------------------------------------------------------------
	// Server con graceful shutdown
	// -----------------------------------------------------------------------
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("servidor iniciado", "port", cfg.Port)
		errCh <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		slog.Info("senal recibida, apagando servidor", "signal", sig)
	case err := <-errCh:
		slog.Error("error en servidor", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("error al apagar servidor", "error", err)
		os.Exit(1)
	}
	slog.Info("servidor apagado correctamente")
}

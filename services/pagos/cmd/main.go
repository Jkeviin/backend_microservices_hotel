package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"

	"pagos/internal/application"
	"pagos/internal/domain/repository"
	"pagos/internal/infrastructure/config"
	infrahttp "pagos/internal/infrastructure/http"
	"pagos/internal/infrastructure/persistence/memory"
	"pagos/internal/infrastructure/persistence/mysql"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	var (
		pagoRepo       repository.PagoRepository
		reservaChecker application.ReservaChecker
	)

	if cfg.UseMockDB {
		slog.Info("usando base de datos en memoria (mock)")
		pagoRepo = memory.NewPagoRepo()
		reservaChecker = memory.NewReservaChecker()
	} else {
		slog.Info("conectando a MySQL", "dsn", maskDSN(cfg.DBDsn))
		db, err := sqlx.Connect("mysql", cfg.DBDsn)
		if err != nil {
			slog.Error("error conectando a MySQL", "error", err)
			os.Exit(1)
		}
		defer db.Close()

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		pagoRepo = mysql.NewPagoRepo(db)
		reservaChecker = mysql.NewReservaChecker(db)
	}

	app := application.NewPagoApp(pagoRepo, reservaChecker)
	handler := infrahttp.NewHandler(app)
	router := infrahttp.NewRouter(handler, cfg.JWTSecret)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("servicio de pagos iniciado", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("error en el servidor", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("apagando servicio...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("error en shutdown", "error", err)
	}

	slog.Info("servicio de pagos detenido")
}

// maskDSN oculta la contrasena del DSN para el log.
func maskDSN(dsn string) string {
	for i, c := range dsn {
		if c == ':' {
			for j := i + 1; j < len(dsn); j++ {
				if dsn[j] == '@' {
					return dsn[:i+1] + "****" + dsn[j:]
				}
			}
			break
		}
	}
	return dsn
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/config"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/server"
)

func main() {
	if err := ejecutar(); err != nil {
		slog.Error("el servidor terminó con error", "error", err)
		os.Exit(1)
	}
}

func ejecutar() error {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Warn("no se pudo leer el archivo .env", "error", err)
	}

	cfg, err := config.Cargar()
	if err != nil {
		return err
	}

	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()

	pool, err := db.Conectar(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	opcionesRedis, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("REDIS_URL inválida: %w", err)
	}

	cliente := asynq.NewClient(opcionesRedis)
	defer cliente.Close()

	manejador, err := server.NuevoRouter(ctx, pool, cfg, cliente)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Puerto),
		Handler:      manejador,
		ReadTimeout:  cfg.TiempoEsperaLectura,
		WriteTimeout: cfg.TiempoEsperaEscritura,
		IdleTimeout:  60 * time.Second,
	}

	errores := make(chan error, 1)

	go func() {
		slog.Info("servidor escuchando", "puerto", cfg.Puerto)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errores <- err
		}
	}()

	select {
	case err := <-errores:
		return err
	case <-ctx.Done():
		slog.Info("apagando el servidor")
	}

	ctxApagado, cancelar := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelar()

	if err := srv.Shutdown(ctxApagado); err != nil {
		return fmt.Errorf("apagado forzado del servidor: %w", err)
	}

	return nil
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/correspondence"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/dispatches"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/config"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/mailer"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/storage"
)

const (
	trabajadoresConcurrentes = 5
	envioPorSegundo          = 8
)

func main() {
	if err := ejecutar(); err != nil {
		slog.Error("el worker terminó con error", "error", err)
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

	proveedor, err := construirProveedor(cfg)
	if err != nil {
		return err
	}

	opcionesRedis, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("REDIS_URL inválida: %w", err)
	}

	servidor := asynq.NewServer(opcionesRedis, asynq.Config{
		Concurrency: trabajadoresConcurrentes,
		Queues:      map[string]int{dispatches.ColaEnvios: 1},
		Logger:      registrador{},
	})

	almacen, err := storage.NuevoAlmacen(cfg.RutaAlmacen, cfg.UrlPublicaArchivos)
	if err != nil {
		return err
	}

	worker := dispatches.NuevoWorker(
		dispatches.NuevoRepositorio(pool),
		correspondence.NuevoRepositorio(pool),
		proveedor,
		almacen,
		envioPorSegundo,
	)

	mux := asynq.NewServeMux()
	worker.Registrar(mux)

	errores := make(chan error, 1)

	go func() {
		slog.Info("worker escuchando la cola", "cola", dispatches.ColaEnvios, "proveedor", proveedor.Nombre())
		if err := servidor.Run(mux); err != nil {
			errores <- err
		}
	}()

	select {
	case err := <-errores:
		return err
	case <-ctx.Done():
		slog.Info("apagando el worker")
		servidor.Shutdown()
	}

	return nil
}

func construirProveedor(cfg config.Config) (mailer.Proveedor, error) {
	if cfg.ModoEnvio == config.ModoEnvioBitacora {
		return mailer.NuevoProveedorBitacora(), nil
	}

	return mailer.NuevoProveedorResend(cfg.ResendAPIKey, cfg.RemitenteCorreo, cfg.RemitenteNombre)
}

type registrador struct{}

func (registrador) Debug(args ...interface{}) { slog.Debug(fmt.Sprint(args...)) }
func (registrador) Info(args ...interface{})  { slog.Info(fmt.Sprint(args...)) }
func (registrador) Warn(args ...interface{})  { slog.Warn(fmt.Sprint(args...)) }
func (registrador) Error(args ...interface{}) { slog.Error(fmt.Sprint(args...)) }
func (registrador) Fatal(args ...interface{}) { slog.Error(fmt.Sprint(args...)) }

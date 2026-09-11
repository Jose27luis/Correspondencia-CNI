package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/assistant"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/correspondence"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/dispatches"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/lists"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/ai"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/auth"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/config"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/mailer"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/middleware"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/pdf"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/storage"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/suppression"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/users"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/webhooks"
)

type EstadoSalud struct {
	Estado      string `json:"estado"`
	BaseDeDatos string `json:"base_de_datos"`
}

func NuevoRouter(ctx context.Context, pool *pgxpool.Pool, cfg config.Config, cliente *asynq.Client) (http.Handler, error) {
	emisor := auth.NuevoEmisor(cfg.JWTSecret, auth.VigenciaDefecto)

	servicioUsuarios, err := users.NuevoServicio(users.NuevoRepositorio(pool), emisor)
	if err != nil {
		return nil, err
	}

	if err := servicioUsuarios.AsegurarUsuarioInicial(
		ctx,
		cfg.UsuarioInicialCorreo,
		cfg.UsuarioInicialClave,
	); err != nil {
		return nil, err
	}

	almacen, err := storage.NuevoAlmacen(cfg.RutaAlmacen, cfg.UrlPublicaArchivos)
	if err != nil {
		return nil, err
	}

	proveedor, err := construirProveedor(cfg)
	if err != nil {
		return nil, err
	}

	repositorioCorrespondencia := correspondence.NuevoRepositorio(pool)
	repositorioEnvios := dispatches.NuevoRepositorio(pool)
	repositorioExclusiones := suppression.NuevoRepositorio(pool)
	servicioExclusiones := suppression.NuevoServicio(repositorioExclusiones, cfg.JWTSecret, cfg.UrlPanel)
	handlerExclusiones := suppression.NuevoHandler(servicioExclusiones)

	asistente := ai.NuevoAsistente(cfg.AnthropicAPIKey)
	if !asistente.Disponible() {
		slog.Warn("el asistente de redacción queda deshabilitado", "motivo", "falta ANTHROPIC_API_KEY")
	}
	handlerAsistente := assistant.NuevoHandler(asistente)

	handlerUsuarios := users.NuevoHandler(servicioUsuarios, middleware.NuevoLimitadorIntentos())
	handlerContactos := contacts.NuevoHandler(contacts.NuevoServicio(contacts.NuevoRepositorio(pool)))
	handlerListas := lists.NuevoHandler(lists.NuevoServicio(lists.NuevoRepositorio(pool)))
	convertidor := pdf.NuevoConvertidor(cfg.SofficeRuta)
	if !convertidor.Disponible() {
		slog.Warn("la vista previa en PDF queda deshabilitada", "motivo", "no se encontró LibreOffice", "ruta", cfg.SofficeRuta)
	}
	handlerCorrespondencia := correspondence.NuevoHandler(
		correspondence.NuevoServicio(repositorioCorrespondencia, almacen, proveedor),
		convertidor,
	)
	handlerEnvios := dispatches.NuevoHandler(dispatches.NuevoServicio(repositorioEnvios, repositorioCorrespondencia, cliente))

	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))

	r.Get("/salud", manejarSalud(pool))

	r.Route("/api", func(api chi.Router) {
		api.Get("/salud", manejarSalud(pool))

		handlerUsuarios.RegistrarPublicas(api)
		registrarWebhooks(api, cfg, repositorioEnvios, repositorioExclusiones)
		handlerExclusiones.RegistrarPublicas(api)

		api.Group(func(protegidas chi.Router) {
			protegidas.Use(middleware.RequiereAutenticacion(emisor))
			handlerUsuarios.RegistrarProtegidas(protegidas)
			handlerContactos.Registrar(protegidas)
			handlerListas.Registrar(protegidas)
			handlerCorrespondencia.Registrar(protegidas)
			handlerEnvios.Registrar(protegidas)
			handlerExclusiones.RegistrarProtegidas(protegidas)
			handlerAsistente.Registrar(protegidas)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, http.StatusNotFound, "recurso no encontrado")
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, http.StatusMethodNotAllowed, "método no permitido")
	})

	return r, nil
}

func construirProveedor(cfg config.Config) (mailer.Proveedor, error) {
	if cfg.ModoEnvio == config.ModoEnvioBitacora {
		return mailer.NuevoProveedorBitacora(), nil
	}

	return mailer.NuevoProveedorResend(cfg.ResendAPIKey, cfg.RemitenteCorreo, cfg.RemitenteNombre)
}

func registrarWebhooks(
	api chi.Router,
	cfg config.Config,
	repositorio *dispatches.Repositorio,
	exclusiones *suppression.Repositorio,
) {
	verificador, err := webhooks.NuevoVerificador(cfg.ResendWebhookSecret)
	if err != nil {
		slog.Warn("el webhook de Resend queda deshabilitado", "motivo", err)
		return
	}

	webhooks.NuevoHandler(verificador, repositorio, exclusiones).Registrar(api)
}

func manejarSalud(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelar := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancelar()

		if err := pool.Ping(ctx); err != nil {
			httpx.JSON(w, http.StatusServiceUnavailable, EstadoSalud{
				Estado:      "degradado",
				BaseDeDatos: "sin conexión",
			})
			return
		}

		httpx.JSON(w, http.StatusOK, EstadoSalud{
			Estado:      "operativo",
			BaseDeDatos: "conectada",
		})
	}
}

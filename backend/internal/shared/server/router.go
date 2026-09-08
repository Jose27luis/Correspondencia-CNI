package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

type EstadoSalud struct {
	Estado      string `json:"estado"`
	BaseDeDatos string `json:"base_de_datos"`
}

func NuevoRouter(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/salud", manejarSalud(pool))

	r.Route("/api", func(api chi.Router) {
		api.Get("/salud", manejarSalud(pool))
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, http.StatusNotFound, "recurso no encontrado")
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, http.StatusMethodNotAllowed, "método no permitido")
	})

	return r
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

package directory

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/intake"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/middleware"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/storage"
)

type Handler struct {
	servicio  *Servicio
	almacen   *storage.Almacen
	limitador *middleware.LimitadorIntentos
}

func NuevoHandler(servicio *Servicio, almacen *storage.Almacen, limitador *middleware.LimitadorIntentos) *Handler {
	return &Handler{servicio: servicio, almacen: almacen, limitador: limitador}
}

func (h *Handler) RegistrarPublicas(r chi.Router) {
	r.Get("/publico/directorio", h.listarPublico)
	r.With(h.limitador.Middleware).Post("/publico/directorio", h.registrar)
}

func (h *Handler) RegistrarProtegidas(r chi.Router) {
	r.Route("/directorio", func(rd chi.Router) {
		rd.Get("/", h.listar)
		rd.Post("/{id}/aprobar", h.aprobar)
		rd.Post("/{id}/rechazar", h.rechazar)
	})
}

func (h *Handler) listarPublico(w http.ResponseWriter, r *http.Request) {
	datos, err := h.servicio.ListarPublico(r.Context())
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, datos)
}

func (h *Handler) registrar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, intake.TamanoMaximoFormulario)
	if err := r.ParseMultipartForm(intake.TamanoMaximoFormulario); err != nil {
		httpx.Error(w, http.StatusBadRequest, intake.ErrFormulario.Error())
		return
	}

	if intake.EsRobot(r) {
		httpx.JSON(w, http.StatusCreated, Registro{Recibido: true})
		return
	}

	entrada := EntradaEmpresa{
		Nombre:      intake.Texto(r, "nombre"),
		Ruc:         intake.Texto(r, "ruc"),
		Correo:      intake.Texto(r, "correo"),
		Direccion:   intake.Texto(r, "direccion"),
		Ciudad:      intake.Texto(r, "ciudad"),
		Telefono:    intake.TextoOpcional(r, "telefono"),
		Celular:     intake.TextoOpcional(r, "celular"),
		Facebook:    intake.TextoOpcional(r, "facebook"),
		PaginaWeb:   intake.TextoOpcional(r, "pagina_web"),
		Descripcion: intake.Texto(r, "descripcion"),
	}
	if err := h.servicio.Validar(&entrada); err != nil {
		responderError(w, err)
		return
	}

	logo, err := intake.GuardarImagen(h.almacen, r, "logo")
	if err != nil {
		responderError(w, err)
		return
	}
	entrada.LogoUrl = logo

	if err := h.servicio.Registrar(r.Context(), entrada); err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, Registro{Recibido: true})
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	listado, err := h.servicio.Listar(
		r.Context(),
		httpx.LeerTexto(r, "estado"),
		int32(httpx.LeerEntero(r, "limite", limitePorDefecto)),
		int32(httpx.LeerEntero(r, "desfase", 0)),
	)
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, listado)
}

func (h *Handler) aprobar(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	empresa, err := h.servicio.Aprobar(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, empresa)
}

func (h *Handler) rechazar(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	var entrada EntradaRechazo
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	empresa, err := h.servicio.Rechazar(r.Context(), id, entrada)
	if err != nil {
		responderError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, empresa)
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNoEncontrada):
		httpx.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrYaRevisada):
		httpx.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, intake.ErrImagenInvalida), errors.Is(err, intake.ErrFormulario):
		httpx.Error(w, http.StatusBadRequest, err.Error())
	default:
		var erroresValidacion validator.ValidationErrors
		if errors.As(err, &erroresValidacion) {
			httpx.ErrorConDetalle(w, http.StatusUnprocessableEntity, "datos inválidos", httpx.DetallarValidacion(erroresValidacion))
			return
		}
		slog.Error("error no controlado en el directorio", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

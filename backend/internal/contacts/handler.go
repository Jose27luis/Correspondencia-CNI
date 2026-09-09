package contacts

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

const tamanoMaximoCSV = 10 << 20

type Handler struct {
	servicio *Servicio
}

func NuevoHandler(servicio *Servicio) *Handler {
	return &Handler{servicio: servicio}
}

func (h *Handler) Registrar(r chi.Router) {
	r.Route("/contactos", func(rc chi.Router) {
		rc.Get("/", h.listar)
		rc.Post("/", h.crear)
		rc.Post("/importar", h.importar)
		rc.Get("/plantilla", h.plantilla)
		rc.Get("/{id}", h.obtener)
		rc.Put("/{id}", h.actualizar)
		rc.Delete("/{id}", h.eliminar)
	})
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	filtro := FiltroListado{
		Limite:  int32(httpx.LeerEntero(r, "limite", limitePorDefecto)),
		Desfase: int32(httpx.LeerEntero(r, "desfase", 0)),
	}

	filtro.Busqueda = httpx.LeerTexto(r, "busqueda")

	listado, err := h.servicio.Listar(r.Context(), filtro)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, listado)
}

func (h *Handler) crear(w http.ResponseWriter, r *http.Request) {
	var entrada EntradaContacto
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	contacto, err := h.servicio.Crear(r.Context(), entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, contacto)
}

func (h *Handler) obtener(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	contacto, err := h.servicio.Obtener(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, contacto)
}

func (h *Handler) actualizar(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	var entrada EntradaContacto
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	contacto, err := h.servicio.Actualizar(r.Context(), id, entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, contacto)
}

func (h *Handler) eliminar(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	if err := h.servicio.Eliminar(r.Context(), id); err != nil {
		responderError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) importar(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(tamanoMaximoCSV); err != nil {
		httpx.Error(w, http.StatusBadRequest, "no se pudo leer el archivo enviado")
		return
	}

	archivo, cabecera, err := r.FormFile("archivo")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "falta el archivo en el campo archivo")
		return
	}
	defer archivo.Close()

	if cabecera.Size > tamanoMaximoCSV {
		httpx.Error(w, http.StatusRequestEntityTooLarge, "el archivo supera los 10 MB")
		return
	}

	resumen, err := h.servicio.Importar(r.Context(), archivo, cabecera.Filename)
	if err != nil {
		if errors.Is(err, ErrCabeceraCSVInvalida) || errors.Is(err, ErrExcelSinFilas) {
			httpx.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, resumen)
}

func (h *Handler) plantilla(w http.ResponseWriter, _ *http.Request) {
	libro, err := GenerarPlantilla()
	if err != nil {
		slog.Error("no se pudo generar la plantilla", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "no se pudo generar la plantilla")
		return
	}
	defer libro.Close()

	w.Header().Set(
		"Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	)
	w.Header().Set("Content-Disposition", `attachment; filename="plantilla-contactos.xlsx"`)

	if err := libro.Write(w); err != nil {
		slog.Error("no se pudo enviar la plantilla", "error", err)
	}
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNoEncontrado):
		httpx.Error(w, http.StatusNotFound, "contacto no encontrado")
	case errors.Is(err, ErrCorreoEnUso):
		httpx.Error(w, http.StatusConflict, "ya existe un contacto con ese correo")
	default:
		var erroresValidacion validator.ValidationErrors
		if errors.As(err, &erroresValidacion) {
			httpx.ErrorConDetalle(w, http.StatusUnprocessableEntity, "datos inválidos", httpx.DetallarValidacion(erroresValidacion))
			return
		}
		slog.Error("error no controlado en contactos", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

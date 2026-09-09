package lists

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

type Handler struct {
	servicio *Servicio
}

func NuevoHandler(servicio *Servicio) *Handler {
	return &Handler{servicio: servicio}
}

func (h *Handler) Registrar(r chi.Router) {
	r.Route("/listas", func(rl chi.Router) {
		rl.Get("/", h.listar)
		rl.Post("/", h.crear)
		rl.Get("/{id}", h.obtener)
		rl.Put("/{id}", h.actualizar)
		rl.Delete("/{id}", h.eliminar)
		rl.Get("/{id}/contactos", h.listarMiembros)
		rl.Post("/{id}/contactos", h.agregarContactos)
		rl.Delete("/{id}/contactos/{contactoId}", h.quitarContacto)
	})
}

func (h *Handler) listar(w http.ResponseWriter, r *http.Request) {
	listado, err := h.servicio.Listar(
		r.Context(),
		int32(httpx.LeerEntero(r, "limite", limitePorDefecto)),
		int32(httpx.LeerEntero(r, "desfase", 0)),
	)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, listado)
}

func (h *Handler) crear(w http.ResponseWriter, r *http.Request) {
	var entrada EntradaLista
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	lista, err := h.servicio.Crear(r.Context(), entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, lista)
}

func (h *Handler) obtener(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	lista, err := h.servicio.Obtener(r.Context(), id)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, lista)
}

func (h *Handler) actualizar(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	var entrada EntradaLista
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	lista, err := h.servicio.Actualizar(r.Context(), id, entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, lista)
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

func (h *Handler) listarMiembros(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	listado, err := h.servicio.ListarMiembros(
		r.Context(),
		id,
		int32(httpx.LeerEntero(r, "limite", limitePorDefecto)),
		int32(httpx.LeerEntero(r, "desfase", 0)),
	)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, listado)
}

func (h *Handler) agregarContactos(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	var entrada EntradaMiembros
	if err := httpx.Decodificar(w, r, &entrada); err != nil {
		httpx.Error(w, http.StatusBadRequest, "el cuerpo de la petición no es válido")
		return
	}

	resultado, err := h.servicio.AgregarContactos(r.Context(), id, entrada)
	if err != nil {
		responderError(w, err)
		return
	}

	httpx.JSON(w, http.StatusOK, resultado)
}

func (h *Handler) quitarContacto(w http.ResponseWriter, r *http.Request) {
	listaID, err := httpx.LeerID(r, "id")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador inválido")
		return
	}

	contactoID, err := uuid.Parse(chi.URLParam(r, "contactoId"))
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "identificador de contacto inválido")
		return
	}

	if err := h.servicio.QuitarContacto(r.Context(), listaID, contactoID); err != nil {
		if errors.Is(err, ErrNoEncontrada) {
			httpx.Error(w, http.StatusNotFound, "el contacto no pertenece a la lista")
			return
		}
		responderError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNoEncontrada):
		httpx.Error(w, http.StatusNotFound, "lista no encontrada")
	case errors.Is(err, ErrContactoInvalido):
		httpx.Error(w, http.StatusUnprocessableEntity, "uno de los contactos no existe")
	default:
		var erroresValidacion validator.ValidationErrors
		if errors.As(err, &erroresValidacion) {
			httpx.ErrorConDetalle(w, http.StatusUnprocessableEntity, "datos inválidos", httpx.DetallarValidacion(erroresValidacion))
			return
		}
		slog.Error("error no controlado en listas", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "error interno del servidor")
	}
}

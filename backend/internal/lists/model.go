package lists

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

var (
	ErrNoEncontrada     = errors.New("lista no encontrada")
	ErrContactoInvalido = errors.New("uno de los contactos no existe")
)

type Lista struct {
	ID             uuid.UUID `json:"id"`
	Nombre         string    `json:"nombre"`
	Descripcion    *string   `json:"descripcion"`
	TotalContactos int64     `json:"total_contactos"`
	CreadoEn       time.Time `json:"creado_en"`
}

type EntradaLista struct {
	Nombre      string  `json:"nombre" validate:"required,min=2,max=200"`
	Descripcion *string `json:"descripcion" validate:"omitempty,max=500"`
}

type EntradaMiembros struct {
	ContactoIDs []uuid.UUID `json:"contacto_ids" validate:"required,min=1,max=5000,dive,required"`
}

type ResultadoMiembros struct {
	Agregados int64 `json:"agregados"`
}

type ListadoListas struct {
	Datos []Lista `json:"datos"`
	Total int64   `json:"total"`
}

type ListadoMiembros struct {
	Datos []contacts.Contacto `json:"datos"`
	Total int64               `json:"total"`
}

func (e *EntradaLista) Normalizar() {
	e.Nombre = strings.TrimSpace(e.Nombre)

	if e.Descripcion != nil {
		descripcion := strings.TrimSpace(*e.Descripcion)
		if descripcion == "" {
			e.Descripcion = nil
		} else {
			e.Descripcion = &descripcion
		}
	}
}

func desdeFila(fila sqlcgen.ListaContacto, totalContactos int64) Lista {
	return Lista{
		ID:             fila.ID,
		Nombre:         fila.Nombre,
		Descripcion:    fila.Descripcion,
		TotalContactos: totalContactos,
		CreadoEn:       fila.CreadoEn,
	}
}

package contacts

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

var (
	ErrNoEncontrado = errors.New("contacto no encontrado")
	ErrCorreoEnUso  = errors.New("ya existe un contacto con ese correo")
)

type Contacto struct {
	ID            uuid.UUID       `json:"id"`
	Nombre        string          `json:"nombre"`
	Empresa       string          `json:"empresa"`
	Correo        string          `json:"correo"`
	Pais          *string         `json:"pais"`
	CamposExtra   json.RawMessage `json:"campos_extra"`
	CreadoEn      time.Time       `json:"creado_en"`
	ActualizadoEn time.Time       `json:"actualizado_en"`
}

type EntradaContacto struct {
	Nombre      string          `json:"nombre" validate:"required,min=2,max=200"`
	Empresa     string          `json:"empresa" validate:"required,min=2,max=200"`
	Correo      string          `json:"correo" validate:"required,email,max=320"`
	Pais        *string         `json:"pais" validate:"omitempty,max=100"`
	CamposExtra json.RawMessage `json:"campos_extra" validate:"omitempty"`
}

type FiltroListado struct {
	Busqueda *string
	Limite   int32
	Desfase  int32
}

type ListadoContactos struct {
	Datos []Contacto `json:"datos"`
	Total int64      `json:"total"`
}

type ResumenImportacion struct {
	Creados      int      `json:"creados"`
	Actualizados int      `json:"actualizados"`
	Omitidos     int      `json:"omitidos"`
	Errores      []string `json:"errores"`
}

func (e *EntradaContacto) Normalizar() {
	e.Nombre = strings.TrimSpace(e.Nombre)
	e.Empresa = strings.TrimSpace(e.Empresa)
	e.Correo = strings.ToLower(strings.TrimSpace(e.Correo))

	if e.Pais != nil {
		pais := strings.TrimSpace(*e.Pais)
		if pais == "" {
			e.Pais = nil
		} else {
			e.Pais = &pais
		}
	}

	if len(e.CamposExtra) == 0 {
		e.CamposExtra = json.RawMessage("{}")
	}
}

func DesdeFila(fila sqlcgen.Contacto) Contacto {
	return desdeFila(fila)
}

func desdeFila(fila sqlcgen.Contacto) Contacto {
	return Contacto{
		ID:            fila.ID,
		Nombre:        fila.Nombre,
		Empresa:       fila.Empresa,
		Correo:        fila.Correo,
		Pais:          fila.Pais,
		CamposExtra:   json.RawMessage(fila.CamposExtra),
		CreadoEn:      fila.CreadoEn,
		ActualizadoEn: fila.ActualizadoEn,
	}
}

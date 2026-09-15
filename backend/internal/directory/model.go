package directory

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

const (
	EstadoPendiente  = "pendiente"
	EstadoAprobado   = "aprobado"
	EstadoRechazado  = "rechazado"
	NombreLista      = "Directorio C-E"
	DescripcionLista = "Empresas aprobadas en el Directorio de Comercio Exterior"
)

var (
	ErrNoEncontrada = errors.New("empresa no encontrada")
	ErrYaRevisada   = errors.New("la empresa ya fue revisada")
)

type Empresa struct {
	ID            uuid.UUID  `json:"id"`
	Nombre        string     `json:"nombre"`
	Ruc           string     `json:"ruc"`
	Correo        string     `json:"correo"`
	Direccion     string     `json:"direccion"`
	Ciudad        string     `json:"ciudad"`
	Telefono      *string    `json:"telefono"`
	Celular       *string    `json:"celular"`
	Facebook      *string    `json:"facebook"`
	PaginaWeb     *string    `json:"pagina_web"`
	Descripcion   string     `json:"descripcion"`
	LogoUrl       *string    `json:"logo_url"`
	Estado        string     `json:"estado"`
	MotivoRechazo *string    `json:"motivo_rechazo"`
	ContactoID    *uuid.UUID `json:"contacto_id"`
	CreadoEn      time.Time  `json:"creado_en"`
	RevisadoEn    *time.Time `json:"revisado_en"`
}

type EmpresaPublica struct {
	ID          uuid.UUID `json:"id"`
	Nombre      string    `json:"nombre"`
	Ciudad      string    `json:"ciudad"`
	Direccion   string    `json:"direccion"`
	Correo      string    `json:"correo"`
	Telefono    *string   `json:"telefono"`
	Celular     *string   `json:"celular"`
	Facebook    *string   `json:"facebook"`
	PaginaWeb   *string   `json:"pagina_web"`
	Descripcion string    `json:"descripcion"`
	LogoUrl     *string   `json:"logo_url"`
}

type EntradaEmpresa struct {
	Nombre      string  `validate:"required,min=2,max=200"`
	Ruc         string  `validate:"required,min=8,max=20"`
	Correo      string  `validate:"required,email,max=320"`
	Direccion   string  `validate:"required,min=3,max=300"`
	Ciudad      string  `validate:"required,min=2,max=120"`
	Telefono    *string `validate:"omitempty,max=40"`
	Celular     *string `validate:"omitempty,max=40"`
	Facebook    *string `validate:"omitempty,max=300"`
	PaginaWeb   *string `validate:"omitempty,max=300"`
	Descripcion string  `validate:"required,min=10,max=2000"`
	LogoUrl     *string
}

type EntradaRechazo struct {
	Motivo string `json:"motivo" validate:"required,min=3,max=500"`
}

type ListadoEmpresas struct {
	Datos []Empresa `json:"datos"`
	Total int64     `json:"total"`
}

type Registro struct {
	Recibido bool `json:"recibido"`
}

func desdeFila(fila sqlcgen.DirectorioEmpresa) Empresa {
	return Empresa{
		ID:            fila.ID,
		Nombre:        fila.Nombre,
		Ruc:           fila.Ruc,
		Correo:        fila.Correo,
		Direccion:     fila.Direccion,
		Ciudad:        fila.Ciudad,
		Telefono:      fila.Telefono,
		Celular:       fila.Celular,
		Facebook:      fila.Facebook,
		PaginaWeb:     fila.PaginaWeb,
		Descripcion:   fila.Descripcion,
		LogoUrl:       fila.LogoUrl,
		Estado:        fila.Estado,
		MotivoRechazo: fila.MotivoRechazo,
		ContactoID:    fila.ContactoID,
		CreadoEn:      fila.CreadoEn,
		RevisadoEn:    fila.RevisadoEn,
	}
}

func publicaDesdeFila(fila sqlcgen.DirectorioEmpresa) EmpresaPublica {
	return EmpresaPublica{
		ID:          fila.ID,
		Nombre:      fila.Nombre,
		Ciudad:      fila.Ciudad,
		Direccion:   fila.Direccion,
		Correo:      fila.Correo,
		Telefono:    fila.Telefono,
		Celular:     fila.Celular,
		Facebook:    fila.Facebook,
		PaginaWeb:   fila.PaginaWeb,
		Descripcion: fila.Descripcion,
		LogoUrl:     fila.LogoUrl,
	}
}

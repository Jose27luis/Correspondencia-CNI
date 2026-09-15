package businessround

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

const (
	EstadoPendiente        = "pendiente"
	EstadoAprobado         = "aprobado"
	EstadoRechazado        = "rechazado"
	TipoOferta             = "oferta"
	TipoDemanda            = "demanda"
	CoincidenciaSugerida   = "sugerida"
	CoincidenciaNotificada = "notificada"
	CoincidenciaDescartada = "descartada"
	NombreLista            = "Rueda de Negocios"
	DescripcionLista       = "Empresas aprobadas en la Rueda de Negocios"
)

var (
	ErrNoEncontrada     = errors.New("registro no encontrado")
	ErrYaRevisada       = errors.New("el registro ya fue revisado")
	ErrEmpresaPendiente = errors.New("primero aprueba la empresa de esta publicación")
	ErrNoAprobada       = errors.New("solo se buscan coincidencias de publicaciones aprobadas")
	ErrYaNotificada     = errors.New("la coincidencia ya fue notificada")
	ErrTipoInvalido     = errors.New("el tipo debe ser oferta o demanda")
)

type Empresa struct {
	ID               uuid.UUID  `json:"id"`
	RazonSocial      string     `json:"razon_social"`
	Ruc              string     `json:"ruc"`
	PersonaEncargada string     `json:"persona_encargada"`
	CargoEncargado   *string    `json:"cargo_encargado"`
	Correo           string     `json:"correo"`
	Direccion        *string    `json:"direccion"`
	Ciudad           *string    `json:"ciudad"`
	Region           *string    `json:"region"`
	Pais             string     `json:"pais"`
	CodigoPostal     *string    `json:"codigo_postal"`
	Telefono         *string    `json:"telefono"`
	Celular          *string    `json:"celular"`
	PaginaWeb        *string    `json:"pagina_web"`
	Estado           string     `json:"estado"`
	MotivoRechazo    *string    `json:"motivo_rechazo"`
	ContactoID       *uuid.UUID `json:"contacto_id"`
	CreadoEn         time.Time  `json:"creado_en"`
	RevisadoEn       *time.Time `json:"revisado_en"`
	Publicaciones    int64      `json:"publicaciones"`
}

type Publicacion struct {
	ID            uuid.UUID  `json:"id"`
	EmpresaID     uuid.UUID  `json:"empresa_id"`
	Tipo          string     `json:"tipo"`
	Titulo        string     `json:"titulo"`
	Descripcion   string     `json:"descripcion"`
	ImagenUrl     *string    `json:"imagen_url"`
	Estado        string     `json:"estado"`
	MotivoRechazo *string    `json:"motivo_rechazo"`
	CreadoEn      time.Time  `json:"creado_en"`
	RevisadoEn    *time.Time `json:"revisado_en"`
	RazonSocial   string     `json:"razon_social"`
	EmpresaCiudad *string    `json:"empresa_ciudad"`
	EmpresaPais   string     `json:"empresa_pais"`
	EmpresaCorreo string     `json:"empresa_correo"`
	EmpresaEstado string     `json:"empresa_estado"`
}

type PublicacionPublica struct {
	ID            uuid.UUID `json:"id"`
	Tipo          string    `json:"tipo"`
	Titulo        string    `json:"titulo"`
	Descripcion   string    `json:"descripcion"`
	ImagenUrl     *string   `json:"imagen_url"`
	CreadoEn      time.Time `json:"creado_en"`
	RazonSocial   string    `json:"razon_social"`
	EmpresaCiudad *string   `json:"empresa_ciudad"`
	EmpresaPais   string    `json:"empresa_pais"`
}

type Coincidencia struct {
	ID             uuid.UUID  `json:"id"`
	DemandaID      uuid.UUID  `json:"demanda_id"`
	OfertaID       uuid.UUID  `json:"oferta_id"`
	Puntaje        int32      `json:"puntaje"`
	Motivo         string     `json:"motivo"`
	Estado         string     `json:"estado"`
	CreadoEn       time.Time  `json:"creado_en"`
	NotificadaEn   *time.Time `json:"notificada_en"`
	DemandaTitulo  string     `json:"demanda_titulo"`
	DemandaEmpresa string     `json:"demanda_empresa"`
	DemandaCorreo  string     `json:"demanda_correo"`
	OfertaTitulo   string     `json:"oferta_titulo"`
	OfertaEmpresa  string     `json:"oferta_empresa"`
	OfertaCorreo   string     `json:"oferta_correo"`
}

type EntradaRegistro struct {
	RazonSocial      string  `validate:"required,min=2,max=200"`
	Ruc              string  `validate:"required,min=8,max=20"`
	PersonaEncargada string  `validate:"required,min=2,max=200"`
	CargoEncargado   *string `validate:"omitempty,max=120"`
	Correo           string  `validate:"required,email,max=320"`
	Direccion        *string `validate:"omitempty,max=300"`
	Ciudad           *string `validate:"omitempty,max=120"`
	Region           *string `validate:"omitempty,max=120"`
	Pais             string  `validate:"required,min=2,max=100"`
	CodigoPostal     *string `validate:"omitempty,max=20"`
	Telefono         *string `validate:"omitempty,max=40"`
	Celular          *string `validate:"omitempty,max=40"`
	PaginaWeb        *string `validate:"omitempty,max=300"`
	Tipo             string  `validate:"required,oneof=oferta demanda"`
	Titulo           string  `validate:"required,min=3,max=200"`
	Descripcion      string  `validate:"required,min=10,max=3000"`
	ImagenUrl        *string
}

type EntradaRechazo struct {
	Motivo string `json:"motivo" validate:"required,min=3,max=500"`
}

type ListadoEmpresas struct {
	Datos []Empresa `json:"datos"`
	Total int64     `json:"total"`
}

type ListadoPublicaciones struct {
	Datos []Publicacion `json:"datos"`
	Total int64         `json:"total"`
}

type ResultadoBusqueda struct {
	Encontradas int `json:"encontradas"`
}

type Registro struct {
	Recibido bool `json:"recibido"`
}

func empresaDesdeFila(fila sqlcgen.RuedaEmpresa, publicaciones int64) Empresa {
	return Empresa{
		ID:               fila.ID,
		RazonSocial:      fila.RazonSocial,
		Ruc:              fila.Ruc,
		PersonaEncargada: fila.PersonaEncargada,
		CargoEncargado:   fila.CargoEncargado,
		Correo:           fila.Correo,
		Direccion:        fila.Direccion,
		Ciudad:           fila.Ciudad,
		Region:           fila.Region,
		Pais:             fila.Pais,
		CodigoPostal:     fila.CodigoPostal,
		Telefono:         fila.Telefono,
		Celular:          fila.Celular,
		PaginaWeb:        fila.PaginaWeb,
		Estado:           fila.Estado,
		MotivoRechazo:    fila.MotivoRechazo,
		ContactoID:       fila.ContactoID,
		CreadoEn:         fila.CreadoEn,
		RevisadoEn:       fila.RevisadoEn,
		Publicaciones:    publicaciones,
	}
}

func publicacionDesdeFila(fila sqlcgen.ObtenerPublicacionRow) Publicacion {
	return Publicacion{
		ID:            fila.ID,
		EmpresaID:     fila.EmpresaID,
		Tipo:          fila.Tipo,
		Titulo:        fila.Titulo,
		Descripcion:   fila.Descripcion,
		ImagenUrl:     fila.ImagenUrl,
		Estado:        fila.Estado,
		MotivoRechazo: fila.MotivoRechazo,
		CreadoEn:      fila.CreadoEn,
		RevisadoEn:    fila.RevisadoEn,
		RazonSocial:   fila.RazonSocial,
		EmpresaCiudad: fila.EmpresaCiudad,
		EmpresaPais:   fila.EmpresaPais,
		EmpresaCorreo: fila.EmpresaCorreo,
		EmpresaEstado: fila.EmpresaEstado,
	}
}

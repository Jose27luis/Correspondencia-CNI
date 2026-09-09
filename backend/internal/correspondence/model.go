package correspondence

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

const (
	EstadoBorrador = "borrador"
	EstadoEncolada = "encolada"
	EstadoEnviando = "enviando"
	EstadoEnviada  = "enviada"
	EstadoFallida  = "fallida"
)

var (
	ErrNoEncontrada    = errors.New("correspondencia no encontrada")
	ErrNoEsBorrador    = errors.New("solo se puede modificar una correspondencia en borrador")
	ErrListaInvalida   = errors.New("la lista indicada no existe")
	ErrSinDestinatario = errors.New("la lista no tiene contactos para previsualizar")
	ErrAdjuntoInvalido = errors.New("adjunto no encontrado")
	ErrLimiteAdjuntos  = errors.New("los adjuntos superan el tamaño permitido")
)

type Correspondencia struct {
	ID            uuid.UUID  `json:"id"`
	UsuarioID     uuid.UUID  `json:"usuario_id"`
	ListaID       *uuid.UUID `json:"lista_id"`
	Asunto        string     `json:"asunto"`
	Cuerpo        string     `json:"cuerpo"`
	Estado        string     `json:"estado"`
	Adjuntos      []Adjunto  `json:"adjuntos"`
	CreadoEn      time.Time  `json:"creado_en"`
	ActualizadoEn time.Time  `json:"actualizado_en"`
}

type Adjunto struct {
	ID            uuid.UUID `json:"id"`
	NombreArchivo string    `json:"nombre_archivo"`
	UrlArchivo    string    `json:"url_archivo"`
	Tipo          string    `json:"tipo"`
	TamanoBytes   int64     `json:"tamano_bytes"`
	CreadoEn      time.Time `json:"creado_en"`
}

type EntradaCorrespondencia struct {
	ListaID *uuid.UUID `json:"lista_id" validate:"omitempty"`
	Asunto  string     `json:"asunto" validate:"required,min=3,max=300"`
	Cuerpo  string     `json:"cuerpo" validate:"required,min=10,max=100000"`
}

type EntradaAdjunto struct {
	NombreArchivo string `json:"nombre_archivo" validate:"required,max=255"`
	UrlArchivo    string `json:"url_archivo" validate:"required,url,max=2048"`
	Tipo          string `json:"tipo" validate:"required,max=150"`
	TamanoBytes   int64  `json:"tamano_bytes" validate:"required,min=1"`
}

type Previsualizacion struct {
	Asunto       string   `json:"asunto"`
	Cuerpo       string   `json:"cuerpo"`
	Destinatario string   `json:"destinatario"`
	Variables    []string `json:"variables_sin_valor"`
}

type ListadoCorrespondencia struct {
	Datos []Correspondencia `json:"datos"`
	Total int64             `json:"total"`
}

func (e *EntradaCorrespondencia) Normalizar() {
	e.Asunto = strings.TrimSpace(e.Asunto)
	e.Cuerpo = strings.TrimSpace(e.Cuerpo)
}

func (e *EntradaAdjunto) Normalizar() {
	e.NombreArchivo = strings.TrimSpace(e.NombreArchivo)
	e.UrlArchivo = strings.TrimSpace(e.UrlArchivo)
	e.Tipo = strings.TrimSpace(e.Tipo)
}

func desdeFila(fila sqlcgen.Correspondencia, adjuntos []Adjunto) Correspondencia {
	return Correspondencia{
		ID:            fila.ID,
		UsuarioID:     fila.UsuarioID,
		ListaID:       fila.ListaID,
		Asunto:        fila.Asunto,
		Cuerpo:        fila.Cuerpo,
		Estado:        fila.Estado,
		Adjuntos:      adjuntos,
		CreadoEn:      fila.CreadoEn,
		ActualizadoEn: fila.ActualizadoEn,
	}
}

func adjuntoDesdeFila(fila sqlcgen.Adjunto) Adjunto {
	return Adjunto{
		ID:            fila.ID,
		NombreArchivo: fila.NombreArchivo,
		UrlArchivo:    fila.UrlArchivo,
		Tipo:          fila.Tipo,
		TamanoBytes:   fila.TamanoBytes,
		CreadoEn:      fila.CreadoEn,
	}
}

package suppression

import (
	"errors"
	"time"
)

const (
	MotivoRebote = "rebote"
	MotivoBaja   = "baja"
	MotivoManual = "manual"
)

var (
	ErrNoEncontrado  = errors.New("el correo no está en la lista de exclusión")
	ErrTokenInvalido = errors.New("el enlace de baja no es válido")
)

type Supresion struct {
	Correo   string    `json:"correo"`
	Motivo   string    `json:"motivo"`
	Detalle  *string   `json:"detalle"`
	CreadoEn time.Time `json:"creado_en"`
}

type ListadoSupresiones struct {
	Datos []Supresion `json:"datos"`
	Total int64       `json:"total"`
}

type EntradaSupresion struct {
	Correo  string `json:"correo" validate:"required,email,max=320"`
	Detalle string `json:"detalle" validate:"omitempty,max=500"`
}

type ResultadoBaja struct {
	Correo  string `json:"correo"`
	Empresa string `json:"empresa"`
}

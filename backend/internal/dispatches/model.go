package dispatches

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	EstadoPendiente = "pendiente"
	EstadoEnviado   = "enviado"
	EstadoEntregado = "entregado"
	EstadoRebotado  = "rebotado"
	EstadoFallido   = "fallido"
)

var (
	ErrNoEncontrado      = errors.New("envío no encontrado")
	ErrSinLista          = errors.New("la correspondencia no tiene una lista asignada")
	ErrListaVacia        = errors.New("la lista no tiene contactos")
	ErrYaProcesada       = errors.New("la correspondencia ya fue encolada o enviada")
	ErrVariablesSinValor = errors.New("el mensaje tiene variables sin valor para algún destinatario")
)

type Envio struct {
	ID                 uuid.UUID  `json:"id"`
	Estado             string     `json:"estado"`
	ProveedorMensajeID *string    `json:"proveedor_mensaje_id"`
	FechaEnvio         *time.Time `json:"fecha_envio"`
	Nombre             string     `json:"nombre"`
	Empresa            string     `json:"empresa"`
	Correo             string     `json:"correo"`
}

type ListadoEnvios struct {
	Datos []Envio `json:"datos"`
	Total int64   `json:"total"`
}

type Resumen struct {
	Total     int64            `json:"total"`
	PorEstado map[string]int64 `json:"por_estado"`
}

type ResultadoEncolado struct {
	CorrespondenciaID uuid.UUID `json:"correspondencia_id"`
	Encolados         int64     `json:"encolados"`
	Estado            string    `json:"estado"`
}

type Evento struct {
	ID    uuid.UUID `json:"id"`
	Tipo  string    `json:"tipo"`
	Fecha time.Time `json:"fecha"`
}

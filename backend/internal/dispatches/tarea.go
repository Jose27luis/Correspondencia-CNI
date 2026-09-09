package dispatches

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TipoEnviarCorreo = "correspondencia:enviar"
	ColaEnvios       = "envios"
)

type CargaEnvio struct {
	EnvioID           uuid.UUID `json:"envio_id"`
	CorrespondenciaID uuid.UUID `json:"correspondencia_id"`
}

func NuevaTareaEnvio(carga CargaEnvio) (*asynq.Task, error) {
	cuerpo, err := json.Marshal(carga)
	if err != nil {
		return nil, fmt.Errorf("no se pudo serializar la tarea de envío: %w", err)
	}

	return asynq.NewTask(TipoEnviarCorreo, cuerpo, asynq.Queue(ColaEnvios), asynq.MaxRetry(3)), nil
}

func LeerCargaEnvio(tarea *asynq.Task) (CargaEnvio, error) {
	var carga CargaEnvio
	if err := json.Unmarshal(tarea.Payload(), &carga); err != nil {
		return CargaEnvio{}, fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	return carga, nil
}

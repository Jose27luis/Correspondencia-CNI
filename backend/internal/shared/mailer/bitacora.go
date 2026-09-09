package mailer

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type ProveedorBitacora struct{}

func NuevoProveedorBitacora() *ProveedorBitacora {
	slog.Warn("el envío de correo está en modo bitácora: no se entregará ningún correo real")
	return &ProveedorBitacora{}
}

func (p *ProveedorBitacora) Nombre() string {
	return "bitacora"
}

func (p *ProveedorBitacora) Enviar(_ context.Context, mensaje Mensaje) (Resultado, error) {
	mensajeID := uuid.NewString()

	slog.Info("correo no enviado, registrado en bitácora",
		"para", mensaje.Para,
		"asunto", mensaje.Asunto,
		"adjuntos", len(mensaje.Adjuntos),
		"mensaje_id", mensajeID,
	)

	return Resultado{MensajeID: mensajeID}, nil
}

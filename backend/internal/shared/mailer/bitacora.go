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

func contarGenerados(adjuntos []Adjunto) int {
	total := 0
	for _, adjunto := range adjuntos {
		if len(adjunto.Contenido) > 0 {
			total++
		}
	}
	return total
}

func (p *ProveedorBitacora) Enviar(_ context.Context, mensaje Mensaje) (Resultado, error) {
	mensajeID := uuid.NewString()

	slog.Info("correo no enviado, registrado en bitácora",
		"para", mensaje.Para,
		"asunto", mensaje.Asunto,
		"adjuntos", len(mensaje.Adjuntos),
		"generados", contarGenerados(mensaje.Adjuntos),
		"mensaje_id", mensajeID,
	)

	return Resultado{MensajeID: mensajeID}, nil
}

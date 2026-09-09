package mailer

import (
	"context"
	"fmt"
	"strings"

	"github.com/resend/resend-go/v2"
)

type ProveedorResend struct {
	cliente         *resend.Client
	remitenteCorreo string
	remitenteNombre string
}

func NuevoProveedorResend(apiKey string, remitenteCorreo string, remitenteNombre string) (*ProveedorResend, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("falta RESEND_API_KEY")
	}
	if strings.TrimSpace(remitenteCorreo) == "" {
		return nil, fmt.Errorf("falta REMITENTE_CORREO")
	}

	return &ProveedorResend{
		cliente:         resend.NewClient(apiKey),
		remitenteCorreo: remitenteCorreo,
		remitenteNombre: remitenteNombre,
	}, nil
}

func (p *ProveedorResend) Nombre() string {
	return "resend"
}

func (p *ProveedorResend) Enviar(ctx context.Context, mensaje Mensaje) (Resultado, error) {
	adjuntos := make([]*resend.Attachment, 0, len(mensaje.Adjuntos))
	for _, adjunto := range mensaje.Adjuntos {
		if len(adjunto.Contenido) > 0 {
			adjuntos = append(adjuntos, &resend.Attachment{
				Content:  adjunto.Contenido,
				Filename: adjunto.NombreArchivo,
			})
			continue
		}

		adjuntos = append(adjuntos, &resend.Attachment{
			Path:     adjunto.UrlArchivo,
			Filename: adjunto.NombreArchivo,
		})
	}

	peticion := &resend.SendEmailRequest{
		From:        p.remitente(),
		To:          []string{mensaje.Para},
		Subject:     mensaje.Asunto,
		Text:        mensaje.Cuerpo,
		Attachments: adjuntos,
	}

	respuesta, err := p.cliente.Emails.SendWithContext(ctx, peticion)
	if err != nil {
		return Resultado{}, fmt.Errorf("%w: %s", ErrEnvioRechazado, err.Error())
	}

	return Resultado{MensajeID: respuesta.Id}, nil
}

func (p *ProveedorResend) remitente() string {
	if p.remitenteNombre == "" {
		return p.remitenteCorreo
	}
	return fmt.Sprintf("%s <%s>", p.remitenteNombre, p.remitenteCorreo)
}

package mailer

import (
	"context"
	"errors"
)

var ErrEnvioRechazado = errors.New("el proveedor rechazó el correo")

type Adjunto struct {
	NombreArchivo string
	UrlArchivo    string
	Contenido     []byte
}

type Mensaje struct {
	Para     string
	Asunto   string
	Cuerpo   string
	Adjuntos []Adjunto
}

type Resultado struct {
	MensajeID string
}

type Proveedor interface {
	Enviar(ctx context.Context, mensaje Mensaje) (Resultado, error)
	Nombre() string
}

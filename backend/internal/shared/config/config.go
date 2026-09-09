package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Puerto                int
	DatabaseURL           string
	RedisURL              string
	ResendAPIKey          string
	ResendWebhookSecret   string
	JWTSecret             string
	RemitenteCorreo       string
	RemitenteNombre       string
	ModoEnvio             string
	RutaAlmacen           string
	UrlPublicaArchivos    string
	UsuarioInicialCorreo  string
	UsuarioInicialClave   string
	TiempoEsperaLectura   time.Duration
	TiempoEsperaEscritura time.Duration
}

const (
	ModoEnvioResend   = "resend"
	ModoEnvioBitacora = "bitacora"
)

var (
	ErrVariableFaltante  = errors.New("variable de entorno requerida no definida")
	ErrModoEnvioInvalido = errors.New("MODO_ENVIO debe ser resend o bitacora")
)

func Cargar() (Config, error) {
	cfg := Config{
		Puerto:                obtenerEntero("PORT", 8080),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		RedisURL:              obtenerTexto("REDIS_URL", "redis://127.0.0.1:6379/0"),
		ResendAPIKey:          os.Getenv("RESEND_API_KEY"),
		ResendWebhookSecret:   os.Getenv("RESEND_WEBHOOK_SECRET"),
		JWTSecret:             os.Getenv("JWT_SECRET"),
		RemitenteCorreo:       os.Getenv("REMITENTE_CORREO"),
		RemitenteNombre:       obtenerTexto("REMITENTE_NOMBRE", "CNI"),
		ModoEnvio:             obtenerTexto("MODO_ENVIO", ModoEnvioResend),
		RutaAlmacen:           obtenerTexto("RUTA_ALMACEN", "almacen/adjuntos"),
		UrlPublicaArchivos:    os.Getenv("URL_PUBLICA_ARCHIVOS"),
		UsuarioInicialCorreo:  os.Getenv("USUARIO_INICIAL_CORREO"),
		UsuarioInicialClave:   os.Getenv("USUARIO_INICIAL_CONTRASENA"),
		TiempoEsperaLectura:   15 * time.Second,
		TiempoEsperaEscritura: 30 * time.Second,
	}

	requeridas := map[string]string{
		"DATABASE_URL":         cfg.DatabaseURL,
		"JWT_SECRET":           cfg.JWTSecret,
		"URL_PUBLICA_ARCHIVOS": cfg.UrlPublicaArchivos,
	}
	for nombre, valor := range requeridas {
		if valor == "" {
			return Config{}, fmt.Errorf("%w: %s", ErrVariableFaltante, nombre)
		}
	}

	switch cfg.ModoEnvio {
	case ModoEnvioResend:
		if cfg.ResendAPIKey == "" {
			return Config{}, fmt.Errorf("%w: %s", ErrVariableFaltante, "RESEND_API_KEY")
		}
		if cfg.RemitenteCorreo == "" {
			return Config{}, fmt.Errorf("%w: %s", ErrVariableFaltante, "REMITENTE_CORREO")
		}
	case ModoEnvioBitacora:
	default:
		return Config{}, ErrModoEnvioInvalido
	}

	return cfg, nil
}

func obtenerTexto(clave string, porDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return porDefecto
}

func obtenerEntero(clave string, porDefecto int) int {
	valor := os.Getenv(clave)
	if valor == "" {
		return porDefecto
	}
	numero, err := strconv.Atoi(valor)
	if err != nil {
		return porDefecto
	}
	return numero
}

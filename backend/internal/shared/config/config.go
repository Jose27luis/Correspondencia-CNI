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
	TiempoEsperaLectura   time.Duration
	TiempoEsperaEscritura time.Duration
}

var ErrVariableFaltante = errors.New("variable de entorno requerida no definida")

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
		TiempoEsperaLectura:   15 * time.Second,
		TiempoEsperaEscritura: 30 * time.Second,
	}

	requeridas := map[string]string{
		"DATABASE_URL": cfg.DatabaseURL,
		"JWT_SECRET":   cfg.JWTSecret,
	}
	for nombre, valor := range requeridas {
		if valor == "" {
			return Config{}, fmt.Errorf("%w: %s", ErrVariableFaltante, nombre)
		}
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

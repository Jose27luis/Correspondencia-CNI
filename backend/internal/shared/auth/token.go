package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	emisor          = "correspondencia-cni"
	VigenciaDefecto = 12 * time.Hour
)

var (
	ErrTokenInvalido = errors.New("token inválido o expirado")
	ErrSinToken      = errors.New("no se envió el token de acceso")
)

type Identidad struct {
	UsuarioID uuid.UUID
	Correo    string
	Nombre    string
}

type Emisor struct {
	secreto  []byte
	vigencia time.Duration
}

func NuevoEmisor(secreto string, vigencia time.Duration) *Emisor {
	if vigencia <= 0 {
		vigencia = VigenciaDefecto
	}
	return &Emisor{secreto: []byte(secreto), vigencia: vigencia}
}

func (e *Emisor) Vigencia() time.Duration {
	return e.vigencia
}

func (e *Emisor) Generar(identidad Identidad) (string, time.Time, error) {
	ahora := time.Now()
	expiracion := ahora.Add(e.vigencia)

	afirmaciones := jwt.MapClaims{
		"sub":    identidad.UsuarioID.String(),
		"correo": identidad.Correo,
		"nombre": identidad.Nombre,
		"iss":    emisor,
		"iat":    ahora.Unix(),
		"exp":    expiracion.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, afirmaciones)

	firmado, err := token.SignedString(e.secreto)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("no se pudo firmar el token: %w", err)
	}

	return firmado, expiracion, nil
}

func (e *Emisor) Verificar(firmado string) (Identidad, error) {
	token, err := jwt.Parse(firmado, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalido
		}
		return e.secreto, nil
	}, jwt.WithIssuer(emisor), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil || !token.Valid {
		return Identidad{}, ErrTokenInvalido
	}

	afirmaciones, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Identidad{}, ErrTokenInvalido
	}

	sujeto, err := afirmaciones.GetSubject()
	if err != nil {
		return Identidad{}, ErrTokenInvalido
	}

	usuarioID, err := uuid.Parse(sujeto)
	if err != nil {
		return Identidad{}, ErrTokenInvalido
	}

	identidad := Identidad{UsuarioID: usuarioID}

	if correo, ok := afirmaciones["correo"].(string); ok {
		identidad.Correo = correo
	}
	if nombre, ok := afirmaciones["nombre"].(string); ok {
		identidad.Nombre = nombre
	}

	return identidad, nil
}

package suppression

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	limitePorDefecto = 50
	limiteMaximo     = 200
	longitudToken    = 24
)

type Servicio struct {
	repositorio *Repositorio
	validador   *validator.Validate
	secreto     []byte
	urlPanel    string
}

func NuevoServicio(repositorio *Repositorio, secreto string, urlPanel string) *Servicio {
	return &Servicio{
		repositorio: repositorio,
		validador:   validator.New(validator.WithRequiredStructEnabled()),
		secreto:     []byte(secreto),
		urlPanel:    strings.TrimSuffix(urlPanel, "/"),
	}
}

func (s *Servicio) TokenDeBaja(contactoID uuid.UUID) string {
	mac := hmac.New(sha256.New, s.secreto)
	mac.Write([]byte("baja:"))
	mac.Write([]byte(contactoID.String()))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))[:longitudToken]
}

func (s *Servicio) EnlaceDeBaja(contactoID uuid.UUID) string {
	return fmt.Sprintf("%s/baja?c=%s&t=%s", s.urlPanel, contactoID, s.TokenDeBaja(contactoID))
}

func (s *Servicio) DarDeBaja(ctx context.Context, contactoID uuid.UUID, token string) (ResultadoBaja, error) {
	esperado := s.TokenDeBaja(contactoID)
	if !hmac.Equal([]byte(token), []byte(esperado)) {
		return ResultadoBaja{}, ErrTokenInvalido
	}

	contacto, err := s.repositorio.ContactoPorID(ctx, contactoID)
	if err != nil {
		return ResultadoBaja{}, err
	}

	if err := s.repositorio.Suprimir(ctx, contacto.Correo, MotivoBaja, "solicitada por el destinatario"); err != nil {
		return ResultadoBaja{}, err
	}

	return ResultadoBaja{Correo: contacto.Correo, Empresa: contacto.Empresa}, nil
}

func (s *Servicio) Agregar(ctx context.Context, entrada EntradaSupresion) error {
	entrada.Correo = strings.ToLower(strings.TrimSpace(entrada.Correo))

	if err := s.validador.Struct(entrada); err != nil {
		return err
	}

	return s.repositorio.Suprimir(ctx, entrada.Correo, MotivoManual, entrada.Detalle)
}

func (s *Servicio) Listar(ctx context.Context, limite int32, desfase int32) (ListadoSupresiones, error) {
	if limite <= 0 {
		limite = limitePorDefecto
	}
	if limite > limiteMaximo {
		limite = limiteMaximo
	}
	if desfase < 0 {
		desfase = 0
	}

	return s.repositorio.Listar(ctx, limite, desfase)
}

func (s *Servicio) Quitar(ctx context.Context, correo string) error {
	return s.repositorio.Quitar(ctx, strings.ToLower(strings.TrimSpace(correo)))
}

func (s *Servicio) ContarEnLista(ctx context.Context, listaID uuid.UUID) (int64, error) {
	return s.repositorio.ContarEnLista(ctx, listaID)
}

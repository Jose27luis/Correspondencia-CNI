package users

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/auth"
)

var ErrRegistroNoPermitido = errors.New("solo un usuario autenticado puede registrar nuevas cuentas")

type Servicio struct {
	repositorio   *Repositorio
	emisor        *auth.Emisor
	validador     *validator.Validate
	hashDeRelleno []byte
}

func NuevoServicio(repositorio *Repositorio, emisor *auth.Emisor) (*Servicio, error) {
	relleno := make([]byte, 32)
	if _, err := rand.Read(relleno); err != nil {
		return nil, fmt.Errorf("no se pudo inicializar el servicio de usuarios: %w", err)
	}

	hashDeRelleno, err := bcrypt.GenerateFromPassword(relleno, bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("no se pudo inicializar el servicio de usuarios: %w", err)
	}

	return &Servicio{
		repositorio:   repositorio,
		emisor:        emisor,
		validador:     validator.New(validator.WithRequiredStructEnabled()),
		hashDeRelleno: hashDeRelleno,
	}, nil
}

func (s *Servicio) Registrar(ctx context.Context, entrada EntradaRegistro) (Usuario, error) {
	entrada.Normalizar()
	if err := s.validador.Struct(entrada); err != nil {
		return Usuario{}, err
	}

	total, err := s.repositorio.Contar(ctx)
	if err != nil {
		return Usuario{}, err
	}

	if total > 0 {
		if _, autenticado := auth.DesdeContexto(ctx); !autenticado {
			return Usuario{}, ErrRegistroNoPermitido
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(entrada.Contrasena), bcrypt.DefaultCost)
	if err != nil {
		return Usuario{}, fmt.Errorf("no se pudo cifrar la contraseña: %w", err)
	}

	return s.repositorio.Crear(ctx, entrada.Nombre, entrada.Correo, string(hash))
}

func (s *Servicio) Acceder(ctx context.Context, entrada EntradaAcceso) (Sesion, error) {
	entrada.Normalizar()
	if err := s.validador.Struct(entrada); err != nil {
		return Sesion{}, ErrCredencialesInvalidas
	}

	fila, err := s.repositorio.ObtenerPorCorreo(ctx, entrada.Correo)
	if err != nil {
		if errors.Is(err, ErrNoEncontrado) {
			bcrypt.CompareHashAndPassword(s.hashDeRelleno, []byte(entrada.Contrasena))
			return Sesion{}, ErrCredencialesInvalidas
		}
		return Sesion{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(fila.PasswordHash), []byte(entrada.Contrasena)); err != nil {
		return Sesion{}, ErrCredencialesInvalidas
	}

	usuario := desdeFila(fila)

	token, expiracion, err := s.emisor.Generar(auth.Identidad{
		UsuarioID: usuario.ID,
		Correo:    usuario.Correo,
		Nombre:    usuario.Nombre,
	})
	if err != nil {
		return Sesion{}, err
	}

	return Sesion{Token: token, ExpiraEn: expiracion, Usuario: usuario}, nil
}

func (s *Servicio) Perfil(ctx context.Context) (Usuario, error) {
	identidad, autenticado := auth.DesdeContexto(ctx)
	if !autenticado {
		return Usuario{}, auth.ErrSinToken
	}
	return s.repositorio.Obtener(ctx, identidad.UsuarioID)
}

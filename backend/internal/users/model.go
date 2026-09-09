package users

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

var (
	ErrCredencialesInvalidas = errors.New("correo o contraseña incorrectos")
	ErrNoEncontrado          = errors.New("usuario no encontrado")
	ErrCorreoEnUso           = errors.New("ya existe un usuario con ese correo")
)

type Usuario struct {
	ID       uuid.UUID `json:"id"`
	Nombre   string    `json:"nombre"`
	Correo   string    `json:"correo"`
	CreadoEn time.Time `json:"creado_en"`
}

type EntradaRegistro struct {
	Nombre     string `json:"nombre" validate:"required,min=2,max=200"`
	Correo     string `json:"correo" validate:"required,email,max=320"`
	Contrasena string `json:"contrasena" validate:"required,min=8,max=128"`
}

type EntradaAcceso struct {
	Correo     string `json:"correo" validate:"required,email,max=320"`
	Contrasena string `json:"contrasena" validate:"required,max=128"`
}

type Sesion struct {
	Token    string    `json:"token"`
	ExpiraEn time.Time `json:"expira_en"`
	Usuario  Usuario   `json:"usuario"`
}

func (e *EntradaRegistro) Normalizar() {
	e.Nombre = strings.TrimSpace(e.Nombre)
	e.Correo = strings.ToLower(strings.TrimSpace(e.Correo))
}

func (e *EntradaAcceso) Normalizar() {
	e.Correo = strings.ToLower(strings.TrimSpace(e.Correo))
}

func desdeFila(fila sqlcgen.Usuario) Usuario {
	return Usuario{
		ID:       fila.ID,
		Nombre:   fila.Nombre,
		Correo:   fila.Correo,
		CreadoEn: fila.CreadoEn,
	}
}

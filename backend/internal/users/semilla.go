package users

import (
	"context"
	"errors"
	"log/slog"
	"strings"
)

func (s *Servicio) AsegurarUsuarioInicial(ctx context.Context, correo string, contrasena string) error {
	correo = strings.TrimSpace(correo)

	if correo == "" || strings.TrimSpace(contrasena) == "" {
		return nil
	}

	total, err := s.repositorio.Contar(ctx)
	if err != nil {
		return err
	}

	if total > 0 {
		return nil
	}

	usuario, err := s.Registrar(ctx, EntradaRegistro{
		Nombre:     nombreDesdeCorreo(correo),
		Correo:     correo,
		Contrasena: contrasena,
	})
	if err != nil {
		if errors.Is(err, ErrCorreoEnUso) {
			return nil
		}
		return err
	}

	slog.Info("usuario inicial creado desde la configuración", "correo", usuario.Correo)

	return nil
}

func nombreDesdeCorreo(correo string) string {
	usuario, _, encontrado := strings.Cut(correo, "@")
	if !encontrado || usuario == "" {
		return "Usuario"
	}
	return usuario
}

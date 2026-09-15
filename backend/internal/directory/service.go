package directory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/intake"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/db/sqlcgen"
)

const (
	limitePorDefecto = 50
	limiteMaximo     = 200
)

type Servicio struct {
	consultas    *sqlcgen.Queries
	incorporador *intake.Incorporador
	validador    *validator.Validate
}

func NuevoServicio(pool *pgxpool.Pool, incorporador *intake.Incorporador) *Servicio {
	return &Servicio{
		consultas:    sqlcgen.New(pool),
		incorporador: incorporador,
		validador:    validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (s *Servicio) Validar(entrada *EntradaEmpresa) error {
	entrada.Correo = strings.ToLower(strings.TrimSpace(entrada.Correo))
	return s.validador.Struct(entrada)
}

func (s *Servicio) Registrar(ctx context.Context, entrada EntradaEmpresa) error {
	if _, err := s.consultas.CrearDirectorioEmpresa(ctx, sqlcgen.CrearDirectorioEmpresaParams{
		Nombre:      entrada.Nombre,
		Ruc:         entrada.Ruc,
		Correo:      entrada.Correo,
		Direccion:   entrada.Direccion,
		Ciudad:      entrada.Ciudad,
		Telefono:    entrada.Telefono,
		Celular:     entrada.Celular,
		Facebook:    entrada.Facebook,
		PaginaWeb:   entrada.PaginaWeb,
		Descripcion: entrada.Descripcion,
		LogoUrl:     entrada.LogoUrl,
	}); err != nil {
		return fmt.Errorf("no se pudo registrar la empresa: %w", err)
	}
	return nil
}

func (s *Servicio) ListarPublico(ctx context.Context) ([]EmpresaPublica, error) {
	filas, err := s.consultas.ListarDirectorioPublico(ctx)
	if err != nil {
		return nil, fmt.Errorf("no se pudo listar el directorio: %w", err)
	}

	datos := make([]EmpresaPublica, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, publicaDesdeFila(fila))
	}
	return datos, nil
}

func (s *Servicio) Listar(ctx context.Context, estado *string, limite int32, desfase int32) (ListadoEmpresas, error) {
	if limite <= 0 {
		limite = limitePorDefecto
	}
	if limite > limiteMaximo {
		limite = limiteMaximo
	}
	if desfase < 0 {
		desfase = 0
	}

	filas, err := s.consultas.ListarDirectorio(ctx, sqlcgen.ListarDirectorioParams{
		Limit:  limite,
		Offset: desfase,
		Estado: estado,
	})
	if err != nil {
		return ListadoEmpresas{}, fmt.Errorf("no se pudo listar el directorio: %w", err)
	}

	total, err := s.consultas.ContarDirectorio(ctx, estado)
	if err != nil {
		return ListadoEmpresas{}, fmt.Errorf("no se pudo contar el directorio: %w", err)
	}

	datos := make([]Empresa, 0, len(filas))
	for _, fila := range filas {
		datos = append(datos, desdeFila(fila))
	}
	return ListadoEmpresas{Datos: datos, Total: total}, nil
}

func (s *Servicio) Aprobar(ctx context.Context, id uuid.UUID) (Empresa, error) {
	fila, err := s.obtener(ctx, id)
	if err != nil {
		return Empresa{}, err
	}
	if fila.Estado == EstadoAprobado {
		return Empresa{}, ErrYaRevisada
	}

	contactoID, err := s.incorporador.Incorporar(ctx, intake.Solicitud{
		Nombre:  fila.Nombre,
		Empresa: fila.Nombre,
		Correo:  fila.Correo,
		Datos: map[string]string{
			"ruc":        fila.Ruc,
			"direccion":  fila.Direccion,
			"ciudad":     fila.Ciudad,
			"telefono":   valor(fila.Telefono),
			"celular":    valor(fila.Celular),
			"facebook":   valor(fila.Facebook),
			"pagina_web": valor(fila.PaginaWeb),
			"origen":     NombreLista,
		},
		NombreLista:      NombreLista,
		DescripcionLista: DescripcionLista,
	})
	if err != nil {
		return Empresa{}, err
	}

	aprobada, err := s.consultas.AprobarDirectorioEmpresa(ctx, sqlcgen.AprobarDirectorioEmpresaParams{
		ID:         id,
		ContactoID: &contactoID,
	})
	if err != nil {
		return Empresa{}, fmt.Errorf("no se pudo aprobar la empresa: %w", err)
	}
	return desdeFila(aprobada), nil
}

func (s *Servicio) Rechazar(ctx context.Context, id uuid.UUID, entrada EntradaRechazo) (Empresa, error) {
	entrada.Motivo = strings.TrimSpace(entrada.Motivo)
	if err := s.validador.Struct(entrada); err != nil {
		return Empresa{}, err
	}

	fila, err := s.obtener(ctx, id)
	if err != nil {
		return Empresa{}, err
	}
	if fila.Estado != EstadoPendiente {
		return Empresa{}, ErrYaRevisada
	}

	rechazada, err := s.consultas.RechazarDirectorioEmpresa(ctx, sqlcgen.RechazarDirectorioEmpresaParams{
		ID:            id,
		MotivoRechazo: &entrada.Motivo,
	})
	if err != nil {
		return Empresa{}, fmt.Errorf("no se pudo rechazar la empresa: %w", err)
	}
	return desdeFila(rechazada), nil
}

func (s *Servicio) obtener(ctx context.Context, id uuid.UUID) (sqlcgen.DirectorioEmpresa, error) {
	fila, err := s.consultas.ObtenerDirectorioEmpresa(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlcgen.DirectorioEmpresa{}, ErrNoEncontrada
		}
		return sqlcgen.DirectorioEmpresa{}, fmt.Errorf("no se pudo obtener la empresa: %w", err)
	}
	return fila, nil
}

func valor(texto *string) string {
	if texto == nil {
		return ""
	}
	return *texto
}

package contacts

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const (
	limitePorDefecto = 50
	limiteMaximo     = 200
	maximoFilasCSV   = 20000
)

var ErrCabeceraCSVInvalida = errors.New("el archivo debe tener las columnas nombre, empresa y correo")

type Servicio struct {
	repositorio *Repositorio
	validador   *validator.Validate
}

func NuevoServicio(repositorio *Repositorio) *Servicio {
	return &Servicio{
		repositorio: repositorio,
		validador:   validator.New(validator.WithRequiredStructEnabled()),
	}
}

func (s *Servicio) Crear(ctx context.Context, entrada EntradaContacto) (Contacto, error) {
	entrada.Normalizar()
	if err := s.validador.Struct(entrada); err != nil {
		return Contacto{}, err
	}
	return s.repositorio.Crear(ctx, entrada)
}

func (s *Servicio) Obtener(ctx context.Context, id uuid.UUID) (Contacto, error) {
	return s.repositorio.Obtener(ctx, id)
}

func (s *Servicio) Listar(ctx context.Context, filtro FiltroListado) (ListadoContactos, error) {
	if filtro.Limite <= 0 {
		filtro.Limite = limitePorDefecto
	}
	if filtro.Limite > limiteMaximo {
		filtro.Limite = limiteMaximo
	}
	if filtro.Desfase < 0 {
		filtro.Desfase = 0
	}
	return s.repositorio.Listar(ctx, filtro)
}

func (s *Servicio) Actualizar(ctx context.Context, id uuid.UUID, entrada EntradaContacto) (Contacto, error) {
	entrada.Normalizar()
	if err := s.validador.Struct(entrada); err != nil {
		return Contacto{}, err
	}
	return s.repositorio.Actualizar(ctx, id, entrada)
}

func (s *Servicio) Eliminar(ctx context.Context, id uuid.UUID) error {
	return s.repositorio.Eliminar(ctx, id)
}

func (s *Servicio) ImportarCSV(ctx context.Context, lector io.Reader) (ResumenImportacion, error) {
	lectorCSV := csv.NewReader(lector)
	lectorCSV.TrimLeadingSpace = true
	lectorCSV.FieldsPerRecord = -1

	cabecera, err := lectorCSV.Read()
	if err != nil {
		return ResumenImportacion{}, ErrCabeceraCSVInvalida
	}

	indices, err := mapearCabecera(cabecera)
	if err != nil {
		return ResumenImportacion{}, err
	}

	resumen := ResumenImportacion{Errores: []string{}}

	for numeroFila := 2; ; numeroFila++ {
		fila, err := lectorCSV.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			resumen.Omitidos++
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("fila %d: no se pudo leer", numeroFila))
			continue
		}

		if numeroFila-1 > maximoFilasCSV {
			return resumen, fmt.Errorf("el archivo supera el máximo de %d filas", maximoFilasCSV)
		}

		entrada, err := construirEntrada(fila, indices)
		if err != nil {
			resumen.Omitidos++
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("fila %d: %s", numeroFila, err.Error()))
			continue
		}

		if err := s.validador.Struct(entrada); err != nil {
			resumen.Omitidos++
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("fila %d: datos inválidos", numeroFila))
			continue
		}

		_, fueCreado, err := s.repositorio.Importar(ctx, entrada)
		if err != nil {
			resumen.Omitidos++
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("fila %d: no se pudo guardar", numeroFila))
			continue
		}

		if fueCreado {
			resumen.Creados++
		} else {
			resumen.Actualizados++
		}
	}

	return resumen, nil
}

type indicesCSV struct {
	nombre  int
	empresa int
	correo  int
	pais    int
	extras  map[string]int
}

func mapearCabecera(cabecera []string) (indicesCSV, error) {
	indices := indicesCSV{nombre: -1, empresa: -1, correo: -1, pais: -1, extras: map[string]int{}}

	for posicion, columna := range cabecera {
		nombreColumna := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(columna, "\ufeff")))

		switch nombreColumna {
		case "nombre":
			indices.nombre = posicion
		case "empresa":
			indices.empresa = posicion
		case "correo", "email":
			indices.correo = posicion
		case "pais", "país":
			indices.pais = posicion
		default:
			if clave := normalizarClave(nombreColumna); clave != "" {
				indices.extras[clave] = posicion
			}
		}
	}

	if indices.nombre < 0 || indices.empresa < 0 || indices.correo < 0 {
		return indicesCSV{}, ErrCabeceraCSVInvalida
	}

	return indices, nil
}

func normalizarClave(columna string) string {
	var construida strings.Builder

	for _, caracter := range columna {
		switch {
		case caracter >= 'a' && caracter <= 'z', caracter >= '0' && caracter <= '9':
			construida.WriteRune(caracter)
		case caracter == ' ', caracter == '-', caracter == '_':
			construida.WriteRune('_')
		case caracter == 'á':
			construida.WriteRune('a')
		case caracter == 'é':
			construida.WriteRune('e')
		case caracter == 'í':
			construida.WriteRune('i')
		case caracter == 'ó':
			construida.WriteRune('o')
		case caracter == 'ú':
			construida.WriteRune('u')
		case caracter == 'ñ':
			construida.WriteRune('n')
		}
	}

	return strings.Trim(construida.String(), "_")
}

func construirEntrada(fila []string, indices indicesCSV) (EntradaContacto, error) {
	valor := func(posicion int) string {
		if posicion < 0 || posicion >= len(fila) {
			return ""
		}
		return fila[posicion]
	}

	entrada := EntradaContacto{
		Nombre:  valor(indices.nombre),
		Empresa: valor(indices.empresa),
		Correo:  valor(indices.correo),
	}

	if pais := valor(indices.pais); pais != "" {
		entrada.Pais = &pais
	}

	if len(indices.extras) > 0 {
		extras := make(map[string]string, len(indices.extras))
		for clave, posicion := range indices.extras {
			if contenido := strings.TrimSpace(valor(posicion)); contenido != "" {
				extras[clave] = contenido
			}
		}

		if len(extras) > 0 {
			codificado, err := json.Marshal(extras)
			if err == nil {
				entrada.CamposExtra = codificado
			}
		}
	}

	entrada.Normalizar()

	if entrada.Correo == "" {
		return EntradaContacto{}, errors.New("correo vacío")
	}

	return entrada, nil
}

package ai

import (
	"context"
	"errors"
	"testing"
)

func candidatasDePrueba() []PublicacionResumen {
	return []PublicacionResumen{
		{ID: "oferta-1", Titulo: "Maíz amarillo duro", Empresa: "Agrosur"},
		{ID: "oferta-2", Titulo: "Transporte refrigerado", Empresa: "Frío Andino"},
	}
}

func TestFiltrarCoincidenciasDescartaIdentificadoresInventados(t *testing.T) {
	resultado := filtrarCoincidencias([]CoincidenciaSugerida{
		{ID: "oferta-1", Puntaje: 90, Motivo: "Ofrece el grano que se demanda."},
		{ID: "oferta-inventada", Puntaje: 95, Motivo: "No existe."},
	}, candidatasDePrueba())

	if len(resultado) != 1 || resultado[0].ID != "oferta-1" {
		t.Fatalf("se esperaba solo oferta-1, se obtuvo %+v", resultado)
	}
}

func TestFiltrarCoincidenciasDescartaPuntajesBajosYFueraDeRango(t *testing.T) {
	resultado := filtrarCoincidencias([]CoincidenciaSugerida{
		{ID: "oferta-1", Puntaje: 59, Motivo: "Débil."},
		{ID: "oferta-2", Puntaje: 140, Motivo: "Fuera de rango."},
	}, candidatasDePrueba())

	if len(resultado) != 0 {
		t.Fatalf("no se esperaban coincidencias, se obtuvo %+v", resultado)
	}
}

func TestFiltrarCoincidenciasDescartaDuplicadasYSinMotivo(t *testing.T) {
	resultado := filtrarCoincidencias([]CoincidenciaSugerida{
		{ID: "oferta-1", Puntaje: 80, Motivo: "  Encaja por producto.  "},
		{ID: "oferta-1", Puntaje: 85, Motivo: "Repetida."},
		{ID: "oferta-2", Puntaje: 75, Motivo: "   "},
	}, candidatasDePrueba())

	if len(resultado) != 1 {
		t.Fatalf("se esperaba una sola coincidencia, se obtuvo %+v", resultado)
	}

	if resultado[0].Motivo != "Encaja por producto." || resultado[0].Puntaje != 80 {
		t.Fatalf("se esperaba la primera versión recortada, se obtuvo %+v", resultado[0])
	}
}

func TestSugerirCoincidenciasSinClaveFalla(t *testing.T) {
	_, err := NuevoAsistente("").SugerirCoincidencias(
		context.Background(),
		PublicacionResumen{ID: "demanda-1"},
		"demanda",
		candidatasDePrueba(),
	)

	if !errors.Is(err, ErrNoConfigurado) {
		t.Fatalf("se esperaba ErrNoConfigurado, se obtuvo %v", err)
	}
}

func TestSugerirCoincidenciasSinCandidatasNoConsulta(t *testing.T) {
	resultado, err := NuevoAsistente("sk-ant-de-prueba").SugerirCoincidencias(
		context.Background(),
		PublicacionResumen{ID: "demanda-1"},
		"demanda",
		nil,
	)

	if err != nil || len(resultado) != 0 {
		t.Fatalf("sin candidatas se esperaba lista vacía sin error, se obtuvo %+v, %v", resultado, err)
	}
}

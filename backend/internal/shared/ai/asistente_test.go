package ai

import (
	"context"
	"errors"
	"testing"
)

func TestExtraerJSONQuitaBloqueDeCodigo(t *testing.T) {
	crudo := "```json\n{\"asunto\": \"Oferta\", \"cuerpo\": \"Texto\"}\n```"

	if extraido := extraerJSON(crudo); extraido != `{"asunto": "Oferta", "cuerpo": "Texto"}` {
		t.Fatalf("no se limpió el bloque de código: %q", extraido)
	}
}

func TestExtraerJSONQuitaTextoAlrededor(t *testing.T) {
	crudo := "Aquí tienes la carta:\n{\"asunto\": \"Oferta\"}\nEspero que sirva."

	if extraido := extraerJSON(crudo); extraido != `{"asunto": "Oferta"}` {
		t.Fatalf("no se aisló el objeto JSON: %q", extraido)
	}
}

func TestExtraerJSONConservaObjetosAnidados(t *testing.T) {
	crudo := `{"veredicto": "revisar", "hallazgos": [{"gravedad": "alta"}]}`

	if extraido := extraerJSON(crudo); extraido != crudo {
		t.Fatalf("se alteró un JSON válido: %q", extraido)
	}
}

func TestExtraerJSONSinLlavesDevuelveElTexto(t *testing.T) {
	if extraido := extraerJSON("  sin json aquí  "); extraido != "sin json aquí" {
		t.Fatalf("se esperaba el texto recortado, se obtuvo %q", extraido)
	}
}

func TestAsistenteSinClaveNoEstaDisponible(t *testing.T) {
	asistente := NuevoAsistente("   ")

	if asistente.Disponible() {
		t.Fatal("un asistente sin clave no debería estar disponible")
	}
}

func TestAsistenteSinClaveRechazaLasOperaciones(t *testing.T) {
	asistente := NuevoAsistente("")

	if _, err := asistente.Redactar(context.Background(), EntradaRedaccion{}); !errors.Is(err, ErrNoConfigurado) {
		t.Fatalf("se esperaba ErrNoConfigurado al redactar, se obtuvo %v", err)
	}

	if _, err := asistente.Revisar(context.Background(), "a", "b", nil); !errors.Is(err, ErrNoConfigurado) {
		t.Fatalf("se esperaba ErrNoConfigurado al revisar, se obtuvo %v", err)
	}
}

func TestAsistenteConClaveEstaDisponible(t *testing.T) {
	if !NuevoAsistente("sk-ant-de-prueba").Disponible() {
		t.Fatal("un asistente con clave debería estar disponible")
	}
}

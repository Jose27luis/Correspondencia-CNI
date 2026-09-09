package correspondence

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
)

func contactoDePrueba() contacts.Contacto {
	pais := "Perú"
	return contacts.Contacto{
		Nombre:      "Ana Quispe",
		Empresa:     "Agroindustrias del Sur",
		Correo:      "ana@agrosur.pe",
		Pais:        &pais,
		CamposExtra: json.RawMessage(`{"Cargo":"Gerente de Compras","volumen":1200}`),
	}
}

func TestRenderizarReemplazaVariablesConocidas(t *testing.T) {
	resultado := Renderizar("Estimada {nombre} de {empresa} ({pais})", contactoDePrueba())

	esperado := "Estimada Ana Quispe de Agroindustrias del Sur (Perú)"
	if resultado.Texto != esperado {
		t.Fatalf("se esperaba %q y se obtuvo %q", esperado, resultado.Texto)
	}

	if len(resultado.SinValor) != 0 {
		t.Fatalf("no se esperaban variables sin valor, se obtuvo %v", resultado.SinValor)
	}
}

func TestRenderizarUsaCamposExtra(t *testing.T) {
	resultado := Renderizar("Cargo: {cargo}, volumen: {volumen}", contactoDePrueba())

	esperado := "Cargo: Gerente de Compras, volumen: 1200"
	if resultado.Texto != esperado {
		t.Fatalf("se esperaba %q y se obtuvo %q", esperado, resultado.Texto)
	}
}

func TestRenderizarConservaVariablesSinValor(t *testing.T) {
	resultado := Renderizar("Hola {nombre}, su cupo es {cupo}", contactoDePrueba())

	esperado := "Hola Ana Quispe, su cupo es {cupo}"
	if resultado.Texto != esperado {
		t.Fatalf("se esperaba %q y se obtuvo %q", esperado, resultado.Texto)
	}

	if !reflect.DeepEqual(resultado.SinValor, []string{"cupo"}) {
		t.Fatalf("se esperaba [cupo] y se obtuvo %v", resultado.SinValor)
	}
}

func TestRenderizarNoDistingueMayusculas(t *testing.T) {
	resultado := Renderizar("{NOMBRE} de {Empresa}", contactoDePrueba())

	esperado := "Ana Quispe de Agroindustrias del Sur"
	if resultado.Texto != esperado {
		t.Fatalf("se esperaba %q y se obtuvo %q", esperado, resultado.Texto)
	}
}

func TestRenderizarSinPaisMarcaLaVariable(t *testing.T) {
	contacto := contactoDePrueba()
	contacto.Pais = nil

	resultado := Renderizar("Desde {pais}", contacto)

	if resultado.Texto != "Desde {pais}" {
		t.Fatalf("se esperaba conservar la variable, se obtuvo %q", resultado.Texto)
	}

	if !reflect.DeepEqual(resultado.SinValor, []string{"pais"}) {
		t.Fatalf("se esperaba [pais] y se obtuvo %v", resultado.SinValor)
	}
}

func TestRenderizarNoReemplazaValorDeVariable(t *testing.T) {
	contacto := contactoDePrueba()
	contacto.Nombre = "{empresa}"

	resultado := Renderizar("Hola {nombre}", contacto)

	if resultado.Texto != "Hola {empresa}" {
		t.Fatalf("el valor sustituido no debe volver a expandirse, se obtuvo %q", resultado.Texto)
	}
}

func TestVariablesDevuelveNombresUnicosOrdenados(t *testing.T) {
	nombres := Variables("{empresa} {nombre} {Empresa} {cupo}")

	if !reflect.DeepEqual(nombres, []string{"cupo", "empresa", "nombre"}) {
		t.Fatalf("se esperaba [cupo empresa nombre] y se obtuvo %v", nombres)
	}
}

func TestVariablesIgnoraLlavesSinNombreValido(t *testing.T) {
	nombres := Variables("{} {con-guion} {espacio libre} {ok}")

	if !reflect.DeepEqual(nombres, []string{"ok"}) {
		t.Fatalf("se esperaba [ok] y se obtuvo %v", nombres)
	}
}

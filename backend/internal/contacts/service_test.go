package contacts

import (
	"encoding/json"
	"testing"
)

func TestMapearCabeceraAceptaAliasYOrdenLibre(t *testing.T) {
	indices, err := mapearCabecera([]string{"País", "EMAIL", " Empresa ", "Nombre"})
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if indices.pais != 0 || indices.correo != 1 || indices.empresa != 2 || indices.nombre != 3 {
		t.Fatalf("las columnas no se mapearon en el orden correcto: %+v", indices)
	}
}

func TestExcelDeCNISinColumnaNombre(t *testing.T) {
	cabecera := []string{
		"EMPRESA", "RUC", "CIUDAD/REGIÓN", "TELEFONO FIJO",
		"TELEFONO MOVIL", "CORREO", "FACEBOOK", "PÁGINA WEB",
	}

	indices, err := mapearCabecera(cabecera)
	if err != nil {
		t.Fatalf("la cabecera del Excel de CNI debía aceptarse, se obtuvo %v", err)
	}

	for _, clave := range []string{"ruc", "ciudad_region", "telefono_fijo", "telefono_movil", "facebook", "pagina_web"} {
		if _, existe := indices.extras[clave]; !existe {
			t.Fatalf("falta la variable %s en %v", clave, indices.extras)
		}
	}

	entrada, err := construirEntrada([]string{
		"Molinos del Sur SAC", "20512345678", "Arequipa", "054123456",
		"987654321", "VENTAS@molinos.pe", "", "www.molinos.pe",
	}, indices)
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if entrada.Nombre != "Molinos del Sur SAC" {
		t.Fatalf("sin columna de nombre debía usarse la empresa, se obtuvo %q", entrada.Nombre)
	}

	if entrada.Correo != "ventas@molinos.pe" {
		t.Fatalf("el correo no se normalizó: %q", entrada.Correo)
	}

	var extras map[string]string
	if err := json.Unmarshal(entrada.CamposExtra, &extras); err != nil {
		t.Fatalf("campos extra inválidos: %v", err)
	}

	if extras["ruc"] != "20512345678" || extras["ciudad_region"] != "Arequipa" {
		t.Fatalf("no se guardaron RUC y ciudad: %v", extras)
	}

	if _, existe := extras["facebook"]; existe {
		t.Fatalf("una celda vacía no debía guardarse: %v", extras)
	}
}

func TestConstruirEntradaRechazaEmpresaVacia(t *testing.T) {
	indices, err := mapearCabecera([]string{"EMPRESA", "CORREO"})
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if _, err := construirEntrada([]string{"", "a@b.pe"}, indices); err == nil {
		t.Fatal("se esperaba error por empresa vacía")
	}
}

func TestMapearCabeceraExigeLasTresObligatorias(t *testing.T) {
	if _, err := mapearCabecera([]string{"nombre", "empresa"}); err != ErrCabeceraCSVInvalida {
		t.Fatalf("se esperaba ErrCabeceraCSVInvalida, se obtuvo %v", err)
	}
}

func TestMapearCabeceraIgnoraElBom(t *testing.T) {
	if _, err := mapearCabecera([]string{"\ufeffnombre", "empresa", "correo"}); err != nil {
		t.Fatalf("el BOM debería ignorarse, se obtuvo %v", err)
	}
}

func TestMapearCabeceraRegistraColumnasExtra(t *testing.T) {
	indices, err := mapearCabecera([]string{"nombre", "empresa", "correo", "Cargo", "N° de RUC"})
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if posicion, existe := indices.extras["cargo"]; !existe || posicion != 3 {
		t.Fatalf("no se registró la columna cargo: %+v", indices.extras)
	}

	if _, existe := indices.extras["n_de_ruc"]; !existe {
		t.Fatalf("no se normalizó la columna con símbolos: %+v", indices.extras)
	}
}

func TestConstruirEntradaGuardaLosExtras(t *testing.T) {
	indices, err := mapearCabecera([]string{"nombre", "empresa", "correo", "Cargo"})
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	entrada, err := construirEntrada(
		[]string{"Ana Quispe", "Agroindustrias del Sur", "ANA@agrosur.pe", "Gerente de Compras"},
		indices,
	)
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if entrada.Correo != "ana@agrosur.pe" {
		t.Fatalf("el correo debía normalizarse a minúsculas, se obtuvo %q", entrada.Correo)
	}

	var extras map[string]string
	if err := json.Unmarshal(entrada.CamposExtra, &extras); err != nil {
		t.Fatalf("los campos extra no son JSON válido: %v", err)
	}

	if extras["cargo"] != "Gerente de Compras" {
		t.Fatalf("se esperaba el cargo en los extras, se obtuvo %v", extras)
	}
}

func TestConstruirEntradaOmiteExtrasVacios(t *testing.T) {
	indices, err := mapearCabecera([]string{"nombre", "empresa", "correo", "Cargo"})
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	entrada, err := construirEntrada(
		[]string{"Luis Mamani", "Transportes Interoceánica", "luis@interoceanica.pe", "   "},
		indices,
	)
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if string(entrada.CamposExtra) != "{}" {
		t.Fatalf("un extra vacío no debía guardarse, se obtuvo %s", entrada.CamposExtra)
	}
}

func TestConstruirEntradaRechazaCorreoVacio(t *testing.T) {
	indices, err := mapearCabecera([]string{"nombre", "empresa", "correo"})
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if _, err := construirEntrada([]string{"Sin correo", "Empresa", ""}, indices); err == nil {
		t.Fatal("se esperaba error por correo vacío")
	}
}

func TestConstruirEntradaToleraFilasCortas(t *testing.T) {
	indices, err := mapearCabecera([]string{"nombre", "empresa", "correo", "pais"})
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	entrada, err := construirEntrada([]string{"Ana", "Agrosur", "ana@agrosur.pe"}, indices)
	if err != nil {
		t.Fatalf("una fila sin la última columna debía aceptarse, se obtuvo %v", err)
	}

	if entrada.Pais != nil {
		t.Fatalf("el país debía quedar vacío, se obtuvo %v", *entrada.Pais)
	}
}

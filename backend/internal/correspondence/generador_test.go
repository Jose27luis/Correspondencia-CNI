package correspondence

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/contacts"
)

func plantillaDePrueba(t *testing.T, cuerpoXML string) []byte {
	t.Helper()

	var almacen bytes.Buffer
	escritor := zip.NewWriter(&almacen)

	archivo, err := escritor.Create(documentoPrincipal)
	if err != nil {
		t.Fatalf("no se pudo crear el documento: %v", err)
	}

	encabezado := `<?xml version="1.0" encoding="UTF-8"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`

	if _, err := archivo.Write([]byte(encabezado + cuerpoXML + `</w:body></w:document>`)); err != nil {
		t.Fatalf("no se pudo escribir el documento: %v", err)
	}

	if err := escritor.Close(); err != nil {
		t.Fatalf("no se pudo cerrar el documento: %v", err)
	}

	return almacen.Bytes()
}

func textoDelDocumento(t *testing.T, documento []byte) string {
	t.Helper()

	contenido, err := LeerWord(bytes.NewReader(documento), int64(len(documento)))
	if err != nil {
		t.Fatalf("no se pudo leer el documento generado: %v", err)
	}

	return contenido.Cuerpo
}

func xmlDelDocumento(t *testing.T, documento []byte) string {
	t.Helper()

	lector, err := zip.NewReader(bytes.NewReader(documento), int64(len(documento)))
	if err != nil {
		t.Fatalf("el documento generado no es un zip válido: %v", err)
	}

	for _, entrada := range lector.File {
		if entrada.Name != documentoPrincipal {
			continue
		}

		abierto, err := entrada.Open()
		if err != nil {
			t.Fatalf("no se pudo abrir el documento: %v", err)
		}
		defer abierto.Close()

		crudo, err := io.ReadAll(abierto)
		if err != nil {
			t.Fatalf("no se pudo leer el documento: %v", err)
		}

		return string(crudo)
	}

	t.Fatal("el documento generado no contiene document.xml")
	return ""
}

func contactoGenerador() contacts.Contacto {
	pais := "Perú"
	return contacts.Contacto{
		Nombre:      "Ana Quispe",
		Empresa:     "Agroindustrias del Sur",
		Correo:      "ana@agrosur.pe",
		Pais:        &pais,
		CamposExtra: json.RawMessage(`{"cargo":"Gerente de Compras"}`),
	}
}

func TestGenerarDocumentoReemplazaVariables(t *testing.T) {
	plantilla := plantillaDePrueba(t,
		`<w:p><w:r><w:t>Señores {empresa}</w:t></w:r></w:p>`+
			`<w:p><w:r><w:t>Atención: {nombre}, {cargo}</w:t></w:r></w:p>`)

	generado, err := GenerarDocumento(plantilla, contactoGenerador())
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	texto := textoDelDocumento(t, generado)

	if !strings.Contains(texto, "Señores Agroindustrias del Sur") {
		t.Fatalf("no se reemplazó la empresa: %q", texto)
	}

	if !strings.Contains(texto, "Atención: Ana Quispe, Gerente de Compras") {
		t.Fatalf("no se reemplazaron nombre y cargo: %q", texto)
	}
}

func TestGenerarDocumentoUneVariablesPartidas(t *testing.T) {
	plantilla := plantillaDePrueba(t,
		`<w:p><w:r><w:rPr><w:b/></w:rPr><w:t>{emp</w:t></w:r><w:r><w:t>resa}</w:t></w:r></w:p>`)

	generado, err := GenerarDocumento(plantilla, contactoGenerador())
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if texto := textoDelDocumento(t, generado); texto != "Agroindustrias del Sur" {
		t.Fatalf("no se unió la variable partida: %q", texto)
	}
}

func TestGenerarDocumentoConservaElFormato(t *testing.T) {
	plantilla := plantillaDePrueba(t,
		`<w:p><w:pPr><w:jc w:val="center"/></w:pPr>`+
			`<w:r><w:rPr><w:b/><w:sz w:val="28"/></w:rPr><w:t>{empresa}</w:t></w:r></w:p>`)

	generado, err := GenerarDocumento(plantilla, contactoGenerador())
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	documento := xmlDelDocumento(t, generado)

	for _, marca := range []string{`<w:jc w:val="center"/>`, `<w:b/>`, `<w:sz w:val="28"/>`} {
		if !strings.Contains(documento, marca) {
			t.Fatalf("se perdió el formato %s en el documento generado", marca)
		}
	}
}

func TestGenerarDocumentoEscapaCaracteresEspeciales(t *testing.T) {
	contacto := contactoGenerador()
	contacto.Empresa = "Ferretería & Aceros <Sur>"

	plantilla := plantillaDePrueba(t, `<w:p><w:r><w:t>{empresa}</w:t></w:r></w:p>`)

	generado, err := GenerarDocumento(plantilla, contacto)
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	documento := xmlDelDocumento(t, generado)
	if strings.Contains(documento, "Aceros <Sur>") {
		t.Fatal("no se escaparon los caracteres especiales, el documento quedaría corrupto")
	}

	if texto := textoDelDocumento(t, generado); texto != "Ferretería & Aceros <Sur>" {
		t.Fatalf("el texto escapado no se recupera igual: %q", texto)
	}
}

func TestGenerarDocumentoDejaVariablesSinDato(t *testing.T) {
	plantilla := plantillaDePrueba(t, `<w:p><w:r><w:t>Cupo: {cupo}</w:t></w:r></w:p>`)

	generado, err := GenerarDocumento(plantilla, contactoGenerador())
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if texto := textoDelDocumento(t, generado); texto != "Cupo: {cupo}" {
		t.Fatalf("una variable sin dato debía conservarse: %q", texto)
	}
}

func TestGenerarDocumentoNoTocaParrafosSinVariables(t *testing.T) {
	plantilla := plantillaDePrueba(t,
		`<w:p><w:r><w:t>Texto fijo sin variables</w:t></w:r></w:p>`+
			`<w:p><w:r><w:t>{empresa}</w:t></w:r></w:p>`)

	generado, err := GenerarDocumento(plantilla, contactoGenerador())
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	texto := textoDelDocumento(t, generado)
	if !strings.Contains(texto, "Texto fijo sin variables") {
		t.Fatalf("se alteró un párrafo sin variables: %q", texto)
	}
}

func TestGenerarDocumentoRechazaArchivoInvalido(t *testing.T) {
	if _, err := GenerarDocumento([]byte("no es un docx"), contactoGenerador()); err == nil {
		t.Fatal("se esperaba error con un archivo inválido")
	}
}

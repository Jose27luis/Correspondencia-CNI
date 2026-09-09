package correspondence

import (
	"archive/zip"
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func documentoDePrueba(t *testing.T, cuerpoXML string) *bytes.Reader {
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

	return bytes.NewReader(almacen.Bytes())
}

func TestLeerWordExtraeParrafos(t *testing.T) {
	documento := documentoDePrueba(t,
		`<w:p><w:r><w:t>Estimada {nombre}:</w:t></w:r></w:p>`+
			`<w:p><w:r><w:t>Le ofrecemos maíz a {empresa}.</w:t></w:r></w:p>`)

	contenido, err := LeerWord(documento, int64(documento.Len()))
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	esperado := "Estimada {nombre}:\nLe ofrecemos maíz a {empresa}."
	if contenido.Cuerpo != esperado {
		t.Fatalf("se esperaba %q y se obtuvo %q", esperado, contenido.Cuerpo)
	}
}

func TestLeerWordUneVariablesPartidasEnVariosFragmentos(t *testing.T) {
	documento := documentoDePrueba(t,
		`<w:p><w:r><w:t>Hola {nom</w:t></w:r><w:r><w:t>bre}</w:t></w:r>`+
			`<w:r><w:t> de {emp</w:t></w:r><w:r><w:t>resa}</w:t></w:r></w:p>`)

	contenido, err := LeerWord(documento, int64(documento.Len()))
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if contenido.Cuerpo != "Hola {nombre} de {empresa}" {
		t.Fatalf("las variables partidas no se unieron: %q", contenido.Cuerpo)
	}

	if !reflect.DeepEqual(contenido.Variables, []string{"empresa", "nombre"}) {
		t.Fatalf("se esperaba [empresa nombre] y se obtuvo %v", contenido.Variables)
	}
}

func TestLeerWordConservaSaltosYTabulaciones(t *testing.T) {
	documento := documentoDePrueba(t,
		`<w:p><w:r><w:t>Primera</w:t><w:br/><w:t>Segunda</w:t><w:tab/><w:t>Tabulada</w:t></w:r></w:p>`)

	contenido, err := LeerWord(documento, int64(documento.Len()))
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if !strings.Contains(contenido.Cuerpo, "Primera\nSegunda\tTabulada") {
		t.Fatalf("no se respetaron los saltos: %q", contenido.Cuerpo)
	}
}

func TestLeerWordIgnoraPropiedadesDeFormato(t *testing.T) {
	documento := documentoDePrueba(t,
		`<w:p><w:pPr><w:jc w:val="center"/></w:pPr>`+
			`<w:r><w:rPr><w:b/></w:rPr><w:t>Solo este texto</w:t></w:r></w:p>`)

	contenido, err := LeerWord(documento, int64(documento.Len()))
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if contenido.Cuerpo != "Solo este texto" {
		t.Fatalf("se coló contenido de formato: %q", contenido.Cuerpo)
	}
}

func TestLeerWordRechazaArchivoQueNoEsWord(t *testing.T) {
	falso := bytes.NewReader([]byte("esto no es un docx"))

	if _, err := LeerWord(falso, int64(falso.Len())); !errors.Is(err, ErrWordInvalido) {
		t.Fatalf("se esperaba ErrWordInvalido, se obtuvo %v", err)
	}
}

func TestLeerWordRechazaDocumentoSinTexto(t *testing.T) {
	documento := documentoDePrueba(t, `<w:p><w:r><w:t>   </w:t></w:r></w:p>`)

	if _, err := LeerWord(documento, int64(documento.Len())); !errors.Is(err, ErrWordVacio) {
		t.Fatalf("se esperaba ErrWordVacio, se obtuvo %v", err)
	}
}

func TestLeerWordColapsaLineasVaciasSeguidas(t *testing.T) {
	documento := documentoDePrueba(t,
		`<w:p><w:r><w:t>Uno</w:t></w:r></w:p><w:p/><w:p/><w:p/>`+
			`<w:p><w:r><w:t>Dos</w:t></w:r></w:p>`)

	contenido, err := LeerWord(documento, int64(documento.Len()))
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	if strings.Contains(contenido.Cuerpo, "\n\n\n") {
		t.Fatalf("no se colapsaron las líneas vacías: %q", contenido.Cuerpo)
	}
}

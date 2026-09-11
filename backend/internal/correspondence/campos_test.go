package correspondence

import (
	"strings"
	"testing"
)

const campoDeCombinacion = `<w:p>` +
	`<w:r><w:fldChar w:fldCharType="begin"/></w:r>` +
	`<w:r><w:instrText xml:space="preserve"> MERGEFIELD NOMBRES </w:instrText></w:r>` +
	`<w:r><w:fldChar w:fldCharType="separate"/></w:r>` +
	`<w:r><w:rPr><w:b/></w:rPr><w:t>«NOMBRES»</w:t></w:r>` +
	`<w:r><w:fldChar w:fldCharType="end"/></w:r>` +
	`</w:p>`

func TestAplanarQuitaLasMarcasDelCampoDeCombinacion(t *testing.T) {
	resultado := string(aplanarCamposCombinacion([]byte(campoDeCombinacion)))

	for _, marca := range []string{"fldChar", "instrText", "MERGEFIELD"} {
		if strings.Contains(resultado, marca) {
			t.Fatalf("quedó %s en el resultado: %s", marca, resultado)
		}
	}

	if !strings.Contains(resultado, `<w:rPr><w:b/></w:rPr><w:t>«NOMBRES»</w:t>`) {
		t.Fatalf("se perdió el texto visible o su formato: %s", resultado)
	}
}

func TestAplanarConservaOtrosCampos(t *testing.T) {
	numeroDePagina := `<w:p>` +
		`<w:r><w:fldChar w:fldCharType="begin"/></w:r>` +
		`<w:r><w:instrText> PAGE </w:instrText></w:r>` +
		`<w:r><w:fldChar w:fldCharType="separate"/></w:r>` +
		`<w:r><w:t>1</w:t></w:r>` +
		`<w:r><w:fldChar w:fldCharType="end"/></w:r>` +
		`</w:p>`

	if resultado := string(aplanarCamposCombinacion([]byte(numeroDePagina))); resultado != numeroDePagina {
		t.Fatalf("se alteró un campo que no es de combinación: %s", resultado)
	}
}

func TestGenerarDocumentoReemplazaCampoDeCombinacionReal(t *testing.T) {
	plantilla := plantillaDePrueba(t, campoDeCombinacion)

	generado, err := GenerarDocumento(plantilla, contactoGenerador())
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo %v", err)
	}

	documento := xmlDelDocumento(t, generado)
	if strings.Contains(documento, "MERGEFIELD") || strings.Contains(documento, "«NOMBRES»") {
		t.Fatalf("el campo de combinación no se convirtió en texto: %s", documento)
	}

	if texto := textoDelDocumento(t, generado); texto != "Ana Quispe" {
		t.Fatalf("se esperaba el nombre del contacto, se obtuvo %q", texto)
	}
}

func TestQuitarCombinacionEliminaElVinculoAlExcel(t *testing.T) {
	configuracion := `<w:settings><w:zoom w:percent="100"/>` +
		`<w:mailMerge><w:mainDocumentType w:val="formLetters"/>` +
		`<w:connectString w:val="Data Source=C:\PRUEBA.xlsx"/></w:mailMerge>` +
		`<w:defaultTabStop w:val="708"/></w:settings>`

	resultado := string(quitarCombinacionCorrespondencia([]byte(configuracion)))

	if strings.Contains(resultado, "mailMerge") || strings.Contains(resultado, "PRUEBA.xlsx") {
		t.Fatalf("quedó el vínculo de combinación: %s", resultado)
	}

	if !strings.Contains(resultado, `<w:zoom w:percent="100"/>`) || !strings.Contains(resultado, "defaultTabStop") {
		t.Fatalf("se perdió el resto de la configuración: %s", resultado)
	}
}

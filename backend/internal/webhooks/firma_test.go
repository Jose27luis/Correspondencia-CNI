package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"
)

const secretoDePrueba = "whsec_MfKQ9r8GKYqrTwjUPD8ILPZIo2LaLaSw"

func cabecerasFirmadas(t *testing.T, id string, marca time.Time, cuerpo []byte) http.Header {
	t.Helper()

	crudo, err := base64.StdEncoding.DecodeString(secretoDePrueba[len(prefijoSecreto):])
	if err != nil {
		t.Fatalf("no se pudo decodificar el secreto: %v", err)
	}

	marcaTexto := strconv.FormatInt(marca.Unix(), 10)

	mac := hmac.New(sha256.New, crudo)
	mac.Write([]byte(fmt.Sprintf("%s.%s.%s", id, marcaTexto, cuerpo)))
	firma := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	cabeceras := http.Header{}
	cabeceras.Set("svix-id", id)
	cabeceras.Set("svix-timestamp", marcaTexto)
	cabeceras.Set("svix-signature", "v1,"+firma)

	return cabeceras
}

func TestVerificarAceptaFirmaCorrecta(t *testing.T) {
	verificador, err := NuevoVerificador(secretoDePrueba)
	if err != nil {
		t.Fatalf("no se pudo crear el verificador: %v", err)
	}

	cuerpo := []byte(`{"type":"email.delivered"}`)
	cabeceras := cabecerasFirmadas(t, "msg_123", time.Now(), cuerpo)

	if err := verificador.Verificar(cabeceras, cuerpo); err != nil {
		t.Fatalf("se esperaba firma válida, se obtuvo %v", err)
	}
}

func TestVerificarRechazaCuerpoAlterado(t *testing.T) {
	verificador, _ := NuevoVerificador(secretoDePrueba)

	cuerpo := []byte(`{"type":"email.delivered"}`)
	cabeceras := cabecerasFirmadas(t, "msg_123", time.Now(), cuerpo)

	if err := verificador.Verificar(cabeceras, []byte(`{"type":"email.bounced"}`)); !errors.Is(err, ErrFirmaInvalida) {
		t.Fatalf("se esperaba ErrFirmaInvalida, se obtuvo %v", err)
	}
}

func TestVerificarRechazaMarcaAntigua(t *testing.T) {
	verificador, _ := NuevoVerificador(secretoDePrueba)

	cuerpo := []byte(`{"type":"email.delivered"}`)
	cabeceras := cabecerasFirmadas(t, "msg_123", time.Now().Add(-30*time.Minute), cuerpo)

	if err := verificador.Verificar(cabeceras, cuerpo); !errors.Is(err, ErrFueraDeTiempo) {
		t.Fatalf("se esperaba ErrFueraDeTiempo, se obtuvo %v", err)
	}
}

func TestVerificarRechazaSinCabeceras(t *testing.T) {
	verificador, _ := NuevoVerificador(secretoDePrueba)

	if err := verificador.Verificar(http.Header{}, []byte("{}")); !errors.Is(err, ErrCabecerasFaltantes) {
		t.Fatalf("se esperaba ErrCabecerasFaltantes, se obtuvo %v", err)
	}
}

func TestVerificarAceptaVariasFirmas(t *testing.T) {
	verificador, _ := NuevoVerificador(secretoDePrueba)

	cuerpo := []byte(`{"type":"email.opened"}`)
	cabeceras := cabecerasFirmadas(t, "msg_456", time.Now(), cuerpo)
	cabeceras.Set("svix-signature", "v1,firmaajena "+cabeceras.Get("svix-signature"))

	if err := verificador.Verificar(cabeceras, cuerpo); err != nil {
		t.Fatalf("se esperaba aceptar una de las firmas, se obtuvo %v", err)
	}
}

func TestNuevoVerificadorRechazaSecretoVacio(t *testing.T) {
	if _, err := NuevoVerificador(""); err == nil {
		t.Fatal("se esperaba error con secreto vacío")
	}
}

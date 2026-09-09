package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	prefijoSecreto  = "whsec_"
	toleranciaReloj = 5 * time.Minute
)

var (
	ErrFirmaInvalida      = errors.New("la firma del webhook no es válida")
	ErrCabecerasFaltantes = errors.New("faltan cabeceras de firma en el webhook")
	ErrFueraDeTiempo      = errors.New("la marca de tiempo del webhook está fuera de rango")
)

type Verificador struct {
	secreto []byte
}

func NuevoVerificador(secreto string) (*Verificador, error) {
	crudo := strings.TrimPrefix(strings.TrimSpace(secreto), prefijoSecreto)
	if crudo == "" {
		return nil, errors.New("falta RESEND_WEBHOOK_SECRET")
	}

	decodificado, err := base64.StdEncoding.DecodeString(crudo)
	if err != nil {
		return nil, fmt.Errorf("el secreto del webhook no es base64 válido: %w", err)
	}

	return &Verificador{secreto: decodificado}, nil
}

func (v *Verificador) Verificar(cabeceras http.Header, cuerpo []byte) error {
	id := cabeceras.Get("svix-id")
	marca := cabeceras.Get("svix-timestamp")
	firmas := cabeceras.Get("svix-signature")

	if id == "" || marca == "" || firmas == "" {
		return ErrCabecerasFaltantes
	}

	if err := v.verificarMarca(marca); err != nil {
		return err
	}

	esperada := v.calcular(id, marca, cuerpo)

	for _, firma := range strings.Split(firmas, " ") {
		partes := strings.SplitN(firma, ",", 2)
		if len(partes) != 2 {
			continue
		}

		if hmac.Equal([]byte(partes[1]), []byte(esperada)) {
			return nil
		}
	}

	return ErrFirmaInvalida
}

func (v *Verificador) calcular(id string, marca string, cuerpo []byte) string {
	mac := hmac.New(sha256.New, v.secreto)
	mac.Write([]byte(id))
	mac.Write([]byte("."))
	mac.Write([]byte(marca))
	mac.Write([]byte("."))
	mac.Write(cuerpo)

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (v *Verificador) verificarMarca(marca string) error {
	segundos, err := strconv.ParseInt(marca, 10, 64)
	if err != nil {
		return ErrFueraDeTiempo
	}

	diferencia := time.Since(time.Unix(segundos, 0))
	if diferencia > toleranciaReloj || diferencia < -toleranciaReloj {
		return ErrFueraDeTiempo
	}

	return nil
}

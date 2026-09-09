package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

const tamanoMaximoCuerpo = 1 << 20

func Decodificar[T any](w http.ResponseWriter, r *http.Request, destino *T) error {
	decodificador := json.NewDecoder(http.MaxBytesReader(w, r.Body, tamanoMaximoCuerpo))
	decodificador.DisallowUnknownFields()
	return decodificador.Decode(destino)
}

func LeerID(r *http.Request, clave string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, clave))
}

func LeerEntero(r *http.Request, clave string, porDefecto int) int {
	valor := r.URL.Query().Get(clave)
	if valor == "" {
		return porDefecto
	}

	numero, err := strconv.Atoi(valor)
	if err != nil {
		return porDefecto
	}

	return numero
}

func LeerTexto(r *http.Request, clave string) *string {
	valor := strings.TrimSpace(r.URL.Query().Get(clave))
	if valor == "" {
		return nil
	}
	return &valor
}

func DetallarValidacion(erroresValidacion validator.ValidationErrors) map[string]string {
	detalle := make(map[string]string, len(erroresValidacion))
	for _, errorCampo := range erroresValidacion {
		detalle[strings.ToLower(errorCampo.Field())] = errorCampo.Tag()
	}
	return detalle
}

package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorRespuesta struct {
	Error   string            `json:"error"`
	Detalle map[string]string `json:"detalle,omitempty"`
}

func JSON[T any](w http.ResponseWriter, codigo int, cuerpo T) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)

	if err := json.NewEncoder(w).Encode(cuerpo); err != nil {
		slog.Error("no se pudo escribir la respuesta JSON", "error", err)
	}
}

func Error(w http.ResponseWriter, codigo int, mensaje string) {
	JSON(w, codigo, ErrorRespuesta{Error: mensaje})
}

func ErrorConDetalle(w http.ResponseWriter, codigo int, mensaje string, detalle map[string]string) {
	JSON(w, codigo, ErrorRespuesta{Error: mensaje, Detalle: detalle})
}

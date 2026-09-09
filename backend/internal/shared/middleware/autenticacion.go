package middleware

import (
	"net/http"
	"strings"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/auth"
	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

const prefijoBearer = "Bearer "

func RequiereAutenticacion(emisor *auth.Emisor) func(http.Handler) http.Handler {
	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cabecera := r.Header.Get("Authorization")
			if !strings.HasPrefix(cabecera, prefijoBearer) {
				httpx.Error(w, http.StatusUnauthorized, "no se envió el token de acceso")
				return
			}

			firmado := strings.TrimSpace(strings.TrimPrefix(cabecera, prefijoBearer))
			if firmado == "" {
				httpx.Error(w, http.StatusUnauthorized, "no se envió el token de acceso")
				return
			}

			identidad, err := emisor.Verificar(firmado)
			if err != nil {
				httpx.Error(w, http.StatusUnauthorized, "token inválido o expirado")
				return
			}

			siguiente.ServeHTTP(w, r.WithContext(auth.ConIdentidad(r.Context(), identidad)))
		})
	}
}

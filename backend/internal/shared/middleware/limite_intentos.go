package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Jose27luis/Correspondencia-CNI/backend/internal/shared/httpx"
)

const (
	intentosPermitidos = 8
	ventanaIntentos    = 5 * time.Minute
	limpiezaCada       = 10 * time.Minute
)

type registroIntentos struct {
	intentos int
	desde    time.Time
}

type LimitadorIntentos struct {
	mutex     sync.Mutex
	registros map[string]*registroIntentos
}

func NuevoLimitadorIntentos() *LimitadorIntentos {
	limitador := &LimitadorIntentos{registros: make(map[string]*registroIntentos)}
	go limitador.limpiarPeriodicamente()
	return limitador
}

func (l *LimitadorIntentos) Middleware(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origen := direccionDe(r)

		if !l.permitir(origen) {
			w.Header().Set("Retry-After", "300")
			httpx.Error(
				w,
				http.StatusTooManyRequests,
				"demasiados intentos fallidos, espere unos minutos antes de volver a intentar",
			)
			return
		}

		siguiente.ServeHTTP(w, r)
	})
}

func (l *LimitadorIntentos) permitir(origen string) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	ahora := time.Now()
	registro, existe := l.registros[origen]

	if !existe || ahora.Sub(registro.desde) > ventanaIntentos {
		l.registros[origen] = &registroIntentos{intentos: 1, desde: ahora}
		return true
	}

	registro.intentos++

	return registro.intentos <= intentosPermitidos
}

func (l *LimitadorIntentos) Perdonar(r *http.Request) {
	origen := direccionDe(r)

	l.mutex.Lock()
	defer l.mutex.Unlock()

	delete(l.registros, origen)
}

func (l *LimitadorIntentos) limpiarPeriodicamente() {
	for range time.Tick(limpiezaCada) {
		l.mutex.Lock()
		ahora := time.Now()

		for origen, registro := range l.registros {
			if ahora.Sub(registro.desde) > ventanaIntentos {
				delete(l.registros, origen)
			}
		}

		l.mutex.Unlock()
	}
}

func direccionDe(r *http.Request) string {
	if visitante := r.Header.Get("CF-Connecting-IP"); visitante != "" {
		return strings.TrimSpace(visitante)
	}

	if reenviada := r.Header.Get("X-Forwarded-For"); reenviada != "" {
		primera, _, _ := strings.Cut(reenviada, ",")
		if primera = strings.TrimSpace(primera); primera != "" {
			return primera
		}
	}

	direccion, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return direccion
}

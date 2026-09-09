package auth

import "context"

type claveContexto struct{}

var claveIdentidad = claveContexto{}

func ConIdentidad(ctx context.Context, identidad Identidad) context.Context {
	return context.WithValue(ctx, claveIdentidad, identidad)
}

func DesdeContexto(ctx context.Context) (Identidad, bool) {
	identidad, ok := ctx.Value(claveIdentidad).(Identidad)
	return identidad, ok
}

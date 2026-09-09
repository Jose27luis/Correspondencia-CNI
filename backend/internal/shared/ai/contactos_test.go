package ai

import (
	"context"
	"errors"
	"testing"
)

func TestLimpiarExtraidosDescartaSinCorreo(t *testing.T) {
	limpios := limpiarExtraidos([]ContactoExtraido{
		{Nombre: "Ana Quispe", Empresa: "Agrosur", Correo: "ana@agrosur.pe"},
		{Nombre: "Sin correo", Empresa: "Empresa X", Correo: ""},
		{Nombre: "Correo raro", Empresa: "Empresa Y", Correo: "esto-no-es-correo"},
	})

	if len(limpios) != 1 {
		t.Fatalf("se esperaba 1 contacto válido, se obtuvieron %d", len(limpios))
	}
}

func TestLimpiarExtraidosNormalizaYDeduplica(t *testing.T) {
	limpios := limpiarExtraidos([]ContactoExtraido{
		{Nombre: " Ana Quispe ", Empresa: " Agrosur ", Correo: "  ANA@Agrosur.PE  "},
		{Nombre: "Ana Q.", Empresa: "Agrosur SAC", Correo: "ana@agrosur.pe"},
	})

	if len(limpios) != 1 {
		t.Fatalf("el duplicado debía descartarse, se obtuvieron %d", len(limpios))
	}

	if limpios[0].Correo != "ana@agrosur.pe" {
		t.Fatalf("el correo no se normalizó: %q", limpios[0].Correo)
	}

	if limpios[0].Nombre != "Ana Quispe" || limpios[0].Empresa != "Agrosur" {
		t.Fatalf("no se recortaron los espacios: %+v", limpios[0])
	}
}

func TestLimpiarExtraidosDevuelveListaVaciaNoNula(t *testing.T) {
	if limpios := limpiarExtraidos(nil); limpios == nil || len(limpios) != 0 {
		t.Fatalf("se esperaba una lista vacía no nula, se obtuvo %v", limpios)
	}
}

func TestExtraerContactosSinClaveFalla(t *testing.T) {
	if _, err := NuevoAsistente("").ExtraerContactos(context.Background(), "texto"); !errors.Is(err, ErrNoConfigurado) {
		t.Fatalf("se esperaba ErrNoConfigurado, se obtuvo %v", err)
	}
}

func TestExtraerContactosRechazaTextoVacio(t *testing.T) {
	_, err := NuevoAsistente("sk-ant-de-prueba").ExtraerContactos(context.Background(), "   ")

	if !errors.Is(err, ErrSinRespuesta) {
		t.Fatalf("se esperaba ErrSinRespuesta con texto vacío, se obtuvo %v", err)
	}
}

func TestInterpretarBusquedaSinClaveFalla(t *testing.T) {
	if _, err := NuevoAsistente("").InterpretarBusqueda(context.Background(), "molinos"); !errors.Is(err, ErrNoConfigurado) {
		t.Fatalf("se esperaba ErrNoConfigurado, se obtuvo %v", err)
	}
}

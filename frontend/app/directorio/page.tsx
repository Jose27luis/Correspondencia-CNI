"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { Building2, CircleCheck, Globe, Loader2, Mail, MapPin, Phone } from "lucide-react";
import { useMemo, useState } from "react";

import { CampoTrampa, MarcoPublico } from "@/components/public-shell";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import { ErrorPeticion } from "@/lib/api-client";
import { apiCaptacion } from "@/lib/intake";

interface Campo {
  nombre: string;
  etiqueta: string;
  tipo?: string;
  requerido?: boolean;
}

const campos: Campo[] = [
  { nombre: "nombre", etiqueta: "Nombre de la empresa", requerido: true },
  { nombre: "ruc", etiqueta: "RUC", requerido: true },
  { nombre: "correo", etiqueta: "Correo", tipo: "email", requerido: true },
  { nombre: "direccion", etiqueta: "Dirección", requerido: true },
  { nombre: "ciudad", etiqueta: "Ciudad", requerido: true },
  { nombre: "telefono", etiqueta: "Teléfono" },
  { nombre: "celular", etiqueta: "Celular" },
  { nombre: "facebook", etiqueta: "Facebook" },
  { nombre: "pagina_web", etiqueta: "Página web" },
];

export default function PaginaDirectorio() {
  const [busqueda, setBusqueda] = useState("");

  const consulta = useQuery({
    queryKey: ["directorio-publico"],
    queryFn: () => apiCaptacion.directorioPublico(),
  });

  const registro = useMutation({
    mutationFn: (formulario: FormData) => apiCaptacion.registrarEnDirectorio(formulario),
  });

  const empresas = useMemo(() => {
    const termino = busqueda.trim().toLowerCase();
    const datos = consulta.data ?? [];
    if (!termino) {
      return datos;
    }
    return datos.filter((empresa) =>
      `${empresa.nombre} ${empresa.ciudad} ${empresa.descripcion}`.toLowerCase().includes(termino),
    );
  }, [consulta.data, busqueda]);

  return (
    <MarcoPublico
      titulo="Directorio de Comercio Exterior"
      descripcion="Empresas vinculadas al comercio entre Perú y Brasil. Registre la suya: el equipo de CNI revisa cada solicitud antes de publicarla."
    >
      <div className="space-y-1.5 sm:max-w-sm">
        <Label htmlFor="buscar-directorio">Buscar empresa</Label>
        <Input
          id="buscar-directorio"
          value={busqueda}
          onChange={(evento) => setBusqueda(evento.target.value)}
        />
      </div>

      {consulta.isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Skeleton className="h-48" />
          <Skeleton className="h-48" />
          <Skeleton className="h-48" />
        </div>
      ) : empresas.length === 0 ? (
        <p className="text-sm text-muted-foreground">Aún no hay empresas publicadas.</p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {empresas.map((empresa) => (
            <Card key={empresa.id}>
              <CardHeader className="flex flex-row items-center gap-3 space-y-0">
                {empresa.logo_url ? (
                  <img
                    src={empresa.logo_url}
                    alt={`Logo de ${empresa.nombre}`}
                    className="h-12 w-12 rounded-md object-contain"
                  />
                ) : (
                  <Building2 className="h-10 w-10 text-muted-foreground" aria-hidden="true" />
                )}
                <CardTitle className="text-base">{empresa.nombre}</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <p className="text-muted-foreground">{empresa.descripcion}</p>
                <p className="flex items-center gap-2">
                  <MapPin className="h-4 w-4 shrink-0" aria-hidden="true" />
                  {empresa.direccion}, {empresa.ciudad}
                </p>
                <p className="flex items-center gap-2">
                  <Mail className="h-4 w-4 shrink-0" aria-hidden="true" />
                  <a href={`mailto:${empresa.correo}`} className="hover:underline">
                    {empresa.correo}
                  </a>
                </p>
                {(empresa.telefono ?? empresa.celular) && (
                  <p className="flex items-center gap-2">
                    <Phone className="h-4 w-4 shrink-0" aria-hidden="true" />
                    {[empresa.telefono, empresa.celular].filter(Boolean).join(" / ")}
                  </p>
                )}
                {empresa.pagina_web && (
                  <p className="flex items-center gap-2">
                    <Globe className="h-4 w-4 shrink-0" aria-hidden="true" />
                    <span className="truncate">{empresa.pagina_web}</span>
                  </p>
                )}
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Registrar mi empresa</CardTitle>
        </CardHeader>
        <CardContent>
          {registro.isSuccess ? (
            <p className="flex items-center gap-2 text-sm text-emerald-700">
              <CircleCheck className="h-5 w-5" aria-hidden="true" />
              Solicitud recibida. La publicaremos cuando el equipo de CNI la revise.
            </p>
          ) : (
            <form
              className="relative grid gap-4 sm:grid-cols-2"
              onSubmit={(evento) => {
                evento.preventDefault();
                registro.mutate(new FormData(evento.currentTarget));
              }}
            >
              <CampoTrampa />
              {campos.map((campo) => (
                <div key={campo.nombre} className="space-y-1.5">
                  <Label htmlFor={`dir-${campo.nombre}`}>
                    {campo.etiqueta}
                    {campo.requerido && " *"}
                  </Label>
                  <Input
                    id={`dir-${campo.nombre}`}
                    name={campo.nombre}
                    type={campo.tipo ?? "text"}
                    required={campo.requerido}
                  />
                </div>
              ))}
              <div className="space-y-1.5">
                <Label htmlFor="dir-logo">Logo (PNG o JPG, hasta 5 MB)</Label>
                <Input id="dir-logo" name="logo" type="file" accept="image/png,image/jpeg" />
              </div>
              <div className="space-y-1.5 sm:col-span-2">
                <Label htmlFor="dir-descripcion">Descripción de la empresa *</Label>
                <Textarea id="dir-descripcion" name="descripcion" rows={4} required minLength={10} />
              </div>
              {registro.isError && (
                <p role="alert" className="text-sm text-red-700 sm:col-span-2">
                  {registro.error instanceof ErrorPeticion
                    ? registro.error.message
                    : "No se pudo enviar la solicitud."}
                </p>
              )}
              <div className="sm:col-span-2">
                <Button type="submit" disabled={registro.isPending} className="text-white">
                  {registro.isPending && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
                  )}
                  Enviar solicitud
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>
    </MarcoPublico>
  );
}

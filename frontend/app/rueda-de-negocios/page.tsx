"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { CircleCheck, Loader2, MapPin } from "lucide-react";
import { useState } from "react";

import { CampoTrampa, MarcoPublico } from "@/components/public-shell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import { ErrorPeticion } from "@/lib/api-client";
import { apiCaptacion, etiquetaTipo, type TipoPublicacion } from "@/lib/intake";
import { cn } from "@/lib/utils";

interface Campo {
  nombre: string;
  etiqueta: string;
  tipo?: string;
  requerido?: boolean;
}

const camposEmpresa: Campo[] = [
  { nombre: "razon_social", etiqueta: "Razón social", requerido: true },
  { nombre: "ruc", etiqueta: "RUC o CNPJ", requerido: true },
  { nombre: "persona_encargada", etiqueta: "Persona encargada", requerido: true },
  { nombre: "cargo_encargado", etiqueta: "Cargo" },
  { nombre: "correo", etiqueta: "Correo", tipo: "email", requerido: true },
  { nombre: "pais", etiqueta: "País", requerido: true },
  { nombre: "region", etiqueta: "Región o estado" },
  { nombre: "ciudad", etiqueta: "Ciudad" },
  { nombre: "direccion", etiqueta: "Dirección" },
  { nombre: "codigo_postal", etiqueta: "Código postal" },
  { nombre: "telefono", etiqueta: "Teléfono" },
  { nombre: "celular", etiqueta: "Celular" },
  { nombre: "pagina_web", etiqueta: "Página web" },
];

const filtros: { valor: TipoPublicacion | undefined; texto: string }[] = [
  { valor: undefined, texto: "Todas" },
  { valor: "oferta", texto: "Ofertas" },
  { valor: "demanda", texto: "Demandas" },
];

export default function PaginaRueda() {
  const [filtro, setFiltro] = useState<TipoPublicacion | undefined>(undefined);
  const [tipo, setTipo] = useState<TipoPublicacion>("oferta");

  const consulta = useQuery({
    queryKey: ["rueda-publica", filtro],
    queryFn: () => apiCaptacion.ruedaPublica(filtro),
  });

  const registro = useMutation({
    mutationFn: (formulario: FormData) => apiCaptacion.registrarEnRueda(formulario),
  });

  const publicaciones = consulta.data ?? [];

  return (
    <MarcoPublico
      titulo="Rueda de Negocios"
      descripcion="Publique lo que su empresa ofrece o lo que necesita. CNI revisa cada publicación y presenta entre sí a las empresas cuyas ofertas y demandas encajan."
    >
      <div className="flex flex-wrap gap-2">
        {filtros.map((opcion) => (
          <Button
            key={opcion.texto}
            type="button"
            variant={filtro === opcion.valor ? "default" : "outline"}
            onClick={() => setFiltro(opcion.valor)}
            className={cn(filtro === opcion.valor && "text-white")}
          >
            {opcion.texto}
          </Button>
        ))}
      </div>

      {consulta.isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Skeleton className="h-56" />
          <Skeleton className="h-56" />
          <Skeleton className="h-56" />
        </div>
      ) : publicaciones.length === 0 ? (
        <p className="text-sm text-muted-foreground">Aún no hay publicaciones.</p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {publicaciones.map((publicacion) => (
            <Card key={publicacion.id} className="overflow-hidden">
              {publicacion.imagen_url && (
                <img
                  src={publicacion.imagen_url}
                  alt={publicacion.titulo}
                  className="h-40 w-full object-cover"
                />
              )}
              <CardHeader className="space-y-2">
                <Badge
                  variant="secondary"
                  className={cn(
                    "w-fit",
                    publicacion.tipo === "oferta"
                      ? "bg-emerald-100 text-emerald-800 hover:bg-emerald-100"
                      : "bg-sky-100 text-sky-800 hover:bg-sky-100",
                  )}
                >
                  {etiquetaTipo[publicacion.tipo]}
                </Badge>
                <CardTitle className="text-base">{publicacion.titulo}</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <p className="text-muted-foreground">{publicacion.descripcion}</p>
                <p className="font-medium">{publicacion.razon_social}</p>
                <p className="flex items-center gap-2 text-muted-foreground">
                  <MapPin className="h-4 w-4 shrink-0" aria-hidden="true" />
                  {[publicacion.empresa_ciudad, publicacion.empresa_pais].filter(Boolean).join(", ")}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Participar en la rueda</CardTitle>
        </CardHeader>
        <CardContent>
          {registro.isSuccess ? (
            <p className="flex items-center gap-2 text-sm text-emerald-700">
              <CircleCheck className="h-5 w-5" aria-hidden="true" />
              Publicación recibida. Le escribiremos cuando encontremos una empresa afín.
            </p>
          ) : (
            <form
              className="relative grid gap-4 sm:grid-cols-2"
              onSubmit={(evento) => {
                evento.preventDefault();
                const formulario = new FormData(evento.currentTarget);
                formulario.set("tipo", tipo);
                registro.mutate(formulario);
              }}
            >
              <CampoTrampa />
              {camposEmpresa.map((campo) => (
                <div key={campo.nombre} className="space-y-1.5">
                  <Label htmlFor={`rueda-${campo.nombre}`}>
                    {campo.etiqueta}
                    {campo.requerido && " *"}
                  </Label>
                  <Input
                    id={`rueda-${campo.nombre}`}
                    name={campo.nombre}
                    type={campo.tipo ?? "text"}
                    required={campo.requerido}
                  />
                </div>
              ))}

              <fieldset className="space-y-2 sm:col-span-2">
                <legend className="text-sm font-medium">¿Qué desea publicar? *</legend>
                <div className="flex gap-2">
                  {(["oferta", "demanda"] as const).map((opcion) => (
                    <Button
                      key={opcion}
                      type="button"
                      variant={tipo === opcion ? "default" : "outline"}
                      aria-pressed={tipo === opcion}
                      onClick={() => setTipo(opcion)}
                      className={cn(tipo === opcion && "text-white")}
                    >
                      {opcion === "oferta" ? "Ofrezco un producto o servicio" : "Necesito un producto o servicio"}
                    </Button>
                  ))}
                </div>
              </fieldset>

              <div className="space-y-1.5 sm:col-span-2">
                <Label htmlFor="rueda-titulo">Título de la publicación *</Label>
                <Input id="rueda-titulo" name="titulo" required minLength={3} />
              </div>
              <div className="space-y-1.5 sm:col-span-2">
                <Label htmlFor="rueda-descripcion">Descripción *</Label>
                <Textarea id="rueda-descripcion" name="descripcion" rows={4} required minLength={10} />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="rueda-imagen">Imagen (PNG o JPG, hasta 5 MB)</Label>
                <Input id="rueda-imagen" name="imagen" type="file" accept="image/png,image/jpeg" />
              </div>

              {registro.isError && (
                <p role="alert" className="text-sm text-red-700 sm:col-span-2">
                  {registro.error instanceof ErrorPeticion
                    ? registro.error.message
                    : "No se pudo enviar la publicación."}
                </p>
              )}
              <div className="sm:col-span-2">
                <Button type="submit" disabled={registro.isPending} className="text-white">
                  {registro.isPending && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
                  )}
                  Enviar publicación
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>
    </MarcoPublico>
  );
}

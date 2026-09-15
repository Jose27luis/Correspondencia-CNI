"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, ExternalLink, Handshake, Loader2, Mail, Sparkles, X } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

import { DialogoRechazo } from "@/components/dialogo-rechazo";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ErrorPeticion } from "@/lib/api-client";
import {
  apiCaptacion,
  colorRevision,
  etiquetaRevision,
  etiquetaTipo,
  type EstadoCoincidencia,
  type EstadoRevision,
} from "@/lib/intake";
import { cn } from "@/lib/utils";

type Pestana = "empresas" | "publicaciones" | "coincidencias";

interface Rechazo {
  tipo: "empresa" | "publicacion";
  id: string;
  nombre: string;
}

const pestanas: { valor: Pestana; texto: string }[] = [
  { valor: "empresas", texto: "Empresas" },
  { valor: "publicaciones", texto: "Publicaciones" },
  { valor: "coincidencias", texto: "Coincidencias" },
];

const filtrosRevision: { valor: EstadoRevision | undefined; texto: string }[] = [
  { valor: "pendiente", texto: "Pendientes" },
  { valor: "aprobado", texto: "Aprobadas" },
  { valor: "rechazado", texto: "Rechazadas" },
  { valor: undefined, texto: "Todas" },
];

const filtrosCoincidencia: { valor: EstadoCoincidencia | undefined; texto: string }[] = [
  { valor: "sugerida", texto: "Sugeridas" },
  { valor: "notificada", texto: "Notificadas" },
  { valor: "descartada", texto: "Descartadas" },
  { valor: undefined, texto: "Todas" },
];

function avisarError(fallo: unknown, mensaje: string) {
  toast.error(fallo instanceof ErrorPeticion ? fallo.message : mensaje);
}

function Filtros<T extends string>({
  opciones,
  valor,
  alCambiar,
}: {
  opciones: { valor: T | undefined; texto: string }[];
  valor: T | undefined;
  alCambiar: (valor: T | undefined) => void;
}) {
  return (
    <div className="flex flex-wrap gap-2">
      {opciones.map((opcion) => (
        <Button
          key={opcion.texto}
          type="button"
          size="sm"
          variant={valor === opcion.valor ? "default" : "outline"}
          onClick={() => alCambiar(opcion.valor)}
          className={cn(valor === opcion.valor && "text-white")}
        >
          {opcion.texto}
        </Button>
      ))}
    </div>
  );
}

function Vacio({ texto }: { texto: string }) {
  return (
    <Card>
      <CardContent className="py-12 text-center">
        <Handshake className="mx-auto h-8 w-8 text-muted-foreground" aria-hidden="true" />
        <p className="mt-3 text-sm font-medium">{texto}</p>
      </CardContent>
    </Card>
  );
}

export default function PaginaRevisionRueda() {
  const [pestana, setPestana] = useState<Pestana>("empresas");
  const [estadoEmpresas, setEstadoEmpresas] = useState<EstadoRevision | undefined>("pendiente");
  const [estadoPublicaciones, setEstadoPublicaciones] = useState<EstadoRevision | undefined>(
    "pendiente",
  );
  const [estadoCoincidencias, setEstadoCoincidencias] = useState<EstadoCoincidencia | undefined>(
    "sugerida",
  );
  const [rechazo, setRechazo] = useState<Rechazo | null>(null);
  const clienteConsultas = useQueryClient();

  const empresas = useQuery({
    queryKey: ["rueda-empresas", estadoEmpresas],
    queryFn: () => apiCaptacion.listarEmpresasRueda(estadoEmpresas),
    enabled: pestana === "empresas",
  });

  const publicaciones = useQuery({
    queryKey: ["rueda-publicaciones", estadoPublicaciones],
    queryFn: () => apiCaptacion.listarPublicaciones(estadoPublicaciones),
    enabled: pestana === "publicaciones",
  });

  const coincidencias = useQuery({
    queryKey: ["rueda-coincidencias", estadoCoincidencias],
    queryFn: () => apiCaptacion.listarCoincidencias(estadoCoincidencias),
    enabled: pestana === "coincidencias",
  });

  function refrescar() {
    for (const clave of ["rueda-empresas", "rueda-publicaciones", "rueda-coincidencias"]) {
      void clienteConsultas.invalidateQueries({ queryKey: [clave] });
    }
  }

  const aprobarEmpresa = useMutation({
    mutationFn: (id: string) => apiCaptacion.aprobarEmpresaRueda(id),
    onSuccess: (empresa) => {
      refrescar();
      toast.success("Empresa aprobada", {
        description: `${empresa.razon_social} ya está en Contactos y en la lista Rueda de Negocios.`,
      });
    },
    onError: (fallo: unknown) => avisarError(fallo, "No se pudo aprobar la empresa"),
  });

  const aprobarPublicacion = useMutation({
    mutationFn: (id: string) => apiCaptacion.aprobarPublicacion(id),
    onSuccess: () => {
      refrescar();
      toast.success("Publicación aprobada", {
        description: "La IA está buscando coincidencias en segundo plano.",
      });
    },
    onError: (fallo: unknown) => avisarError(fallo, "No se pudo aprobar la publicación"),
  });

  const buscar = useMutation({
    mutationFn: (id: string) => apiCaptacion.buscarCoincidencias(id),
    onSuccess: (resultado) => {
      refrescar();
      toast.success(
        resultado.encontradas === 0
          ? "La IA no encontró coincidencias por ahora"
          : `La IA encontró ${resultado.encontradas} coincidencia(s)`,
      );
    },
    onError: (fallo: unknown) => avisarError(fallo, "No se pudieron buscar coincidencias"),
  });

  const rechazar = useMutation({
    mutationFn: async (datos: Rechazo & { motivo: string }): Promise<void> => {
      if (datos.tipo === "empresa") {
        await apiCaptacion.rechazarEmpresaRueda(datos.id, datos.motivo);
        return;
      }
      await apiCaptacion.rechazarPublicacion(datos.id, datos.motivo);
    },
    onSuccess: () => {
      refrescar();
      setRechazo(null);
      toast.success("Rechazado");
    },
    onError: (fallo: unknown) => avisarError(fallo, "No se pudo rechazar"),
  });

  const notificar = useMutation({
    mutationFn: (id: string) => apiCaptacion.notificarCoincidencia(id),
    onSuccess: () => {
      refrescar();
      toast.success("Coincidencia notificada", {
        description: "Ambas empresas recibieron un correo presentándose entre sí.",
      });
    },
    onError: (fallo: unknown) => avisarError(fallo, "No se pudo notificar"),
  });

  const descartar = useMutation({
    mutationFn: (id: string) => apiCaptacion.descartarCoincidencia(id),
    onSuccess: () => {
      refrescar();
      toast.success("Coincidencia descartada");
    },
    onError: (fallo: unknown) => avisarError(fallo, "No se pudo descartar"),
  });

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Rueda de Negocios</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Aprueba empresas y publicaciones. Al aprobar una publicación, la IA sugiere con qué
            ofertas o demandas encaja.
          </p>
        </div>
        <Button asChild variant="outline">
          <a href="/rueda-de-negocios" target="_blank" rel="noreferrer">
            <ExternalLink className="mr-2 h-4 w-4" aria-hidden="true" />
            Ver página pública
          </a>
        </Button>
      </div>

      <div role="tablist" className="flex gap-1 border-b">
        {pestanas.map((opcion) => (
          <button
            key={opcion.valor}
            type="button"
            role="tab"
            aria-selected={pestana === opcion.valor}
            onClick={() => setPestana(opcion.valor)}
            className={cn(
              "-mb-px border-b-2 px-4 py-2 text-sm transition-colors",
              pestana === opcion.valor
                ? "border-primary font-medium text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            {opcion.texto}
          </button>
        ))}
      </div>

      {pestana === "empresas" && (
        <div className="space-y-4">
          <Filtros opciones={filtrosRevision} valor={estadoEmpresas} alCambiar={setEstadoEmpresas} />
          {empresas.isLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : (empresas.data?.datos ?? []).length === 0 ? (
            <Vacio texto="No hay empresas en este estado" />
          ) : (
            (empresas.data?.datos ?? []).map((empresa) => (
              <Card key={empresa.id}>
                <CardContent className="flex flex-wrap gap-4 p-5">
                  <div className="min-w-0 flex-1 space-y-1 text-sm">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-semibold">{empresa.razon_social}</p>
                      <Badge variant="secondary" className={colorRevision[empresa.estado]}>
                        {etiquetaRevision[empresa.estado]}
                      </Badge>
                    </div>
                    <p className="text-muted-foreground">
                      {empresa.ruc} · {empresa.correo} ·{" "}
                      {[empresa.ciudad, empresa.region, empresa.pais].filter(Boolean).join(", ")}
                    </p>
                    <p>
                      {empresa.persona_encargada}
                      {empresa.cargo_encargado && `, ${empresa.cargo_encargado}`}
                    </p>
                    <p className="text-muted-foreground">
                      {[empresa.telefono, empresa.celular, empresa.pagina_web]
                        .filter(Boolean)
                        .join(" · ")}
                    </p>
                    <p className="text-muted-foreground">{empresa.publicaciones} publicación(es)</p>
                    {empresa.motivo_rechazo && (
                      <p className="text-red-700">Motivo del rechazo: {empresa.motivo_rechazo}</p>
                    )}
                  </div>
                  {empresa.estado !== "aprobado" && (
                    <div className="flex items-start gap-2">
                      <Button
                        type="button"
                        size="sm"
                        disabled={aprobarEmpresa.isPending}
                        onClick={() => aprobarEmpresa.mutate(empresa.id)}
                        className="text-white"
                      >
                        <Check className="mr-1 h-4 w-4" aria-hidden="true" />
                        Aprobar
                      </Button>
                      {empresa.estado === "pendiente" && (
                        <Button
                          type="button"
                          size="sm"
                          variant="outline"
                          onClick={() =>
                            setRechazo({ tipo: "empresa", id: empresa.id, nombre: empresa.razon_social })
                          }
                        >
                          <X className="mr-1 h-4 w-4" aria-hidden="true" />
                          Rechazar
                        </Button>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>
            ))
          )}
        </div>
      )}

      {pestana === "publicaciones" && (
        <div className="space-y-4">
          <Filtros
            opciones={filtrosRevision}
            valor={estadoPublicaciones}
            alCambiar={setEstadoPublicaciones}
          />
          {publicaciones.isLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : (publicaciones.data?.datos ?? []).length === 0 ? (
            <Vacio texto="No hay publicaciones en este estado" />
          ) : (
            (publicaciones.data?.datos ?? []).map((publicacion) => (
              <Card key={publicacion.id}>
                <CardContent className="flex flex-wrap gap-4 p-5">
                  {publicacion.imagen_url && (
                    <img
                      src={publicacion.imagen_url}
                      alt={publicacion.titulo}
                      className="h-20 w-28 rounded-md border object-cover"
                    />
                  )}
                  <div className="min-w-0 flex-1 space-y-1 text-sm">
                    <div className="flex flex-wrap items-center gap-2">
                      <Badge variant="outline">{etiquetaTipo[publicacion.tipo]}</Badge>
                      <p className="font-semibold">{publicacion.titulo}</p>
                      <Badge variant="secondary" className={colorRevision[publicacion.estado]}>
                        {etiquetaRevision[publicacion.estado]}
                      </Badge>
                    </div>
                    <p>{publicacion.descripcion}</p>
                    <p className="text-muted-foreground">
                      {publicacion.razon_social} ({etiquetaRevision[publicacion.empresa_estado].toLowerCase()})
                      · {publicacion.empresa_correo}
                    </p>
                    {publicacion.motivo_rechazo && (
                      <p className="text-red-700">Motivo del rechazo: {publicacion.motivo_rechazo}</p>
                    )}
                  </div>
                  <div className="flex items-start gap-2">
                    {publicacion.estado === "aprobado" ? (
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        disabled={buscar.isPending}
                        onClick={() => buscar.mutate(publicacion.id)}
                      >
                        {buscar.isPending && buscar.variables === publicacion.id ? (
                          <Loader2 className="mr-1 h-4 w-4 animate-spin" aria-hidden="true" />
                        ) : (
                          <Sparkles className="mr-1 h-4 w-4" aria-hidden="true" />
                        )}
                        Buscar coincidencias
                      </Button>
                    ) : (
                      <>
                        <Button
                          type="button"
                          size="sm"
                          disabled={aprobarPublicacion.isPending}
                          onClick={() => aprobarPublicacion.mutate(publicacion.id)}
                          className="text-white"
                        >
                          <Check className="mr-1 h-4 w-4" aria-hidden="true" />
                          Aprobar
                        </Button>
                        {publicacion.estado === "pendiente" && (
                          <Button
                            type="button"
                            size="sm"
                            variant="outline"
                            onClick={() =>
                              setRechazo({
                                tipo: "publicacion",
                                id: publicacion.id,
                                nombre: publicacion.titulo,
                              })
                            }
                          >
                            <X className="mr-1 h-4 w-4" aria-hidden="true" />
                            Rechazar
                          </Button>
                        )}
                      </>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))
          )}
        </div>
      )}

      {pestana === "coincidencias" && (
        <div className="space-y-4">
          <Filtros
            opciones={filtrosCoincidencia}
            valor={estadoCoincidencias}
            alCambiar={setEstadoCoincidencias}
          />
          {coincidencias.isLoading ? (
            <Skeleton className="h-40 w-full" />
          ) : (coincidencias.data ?? []).length === 0 ? (
            <Vacio texto="No hay coincidencias en este estado" />
          ) : (
            (coincidencias.data ?? []).map((coincidencia) => (
              <Card key={coincidencia.id}>
                <CardContent className="flex flex-wrap gap-4 p-5">
                  <div className="flex h-14 w-14 shrink-0 flex-col items-center justify-center rounded-lg bg-emerald-50 text-emerald-800">
                    <span className="text-lg font-semibold">{coincidencia.puntaje}</span>
                    <span className="text-[0.65rem]">de 100</span>
                  </div>
                  <div className="min-w-0 flex-1 space-y-2 text-sm">
                    <div className="grid gap-2 sm:grid-cols-2">
                      <div>
                        <p className="text-xs uppercase text-muted-foreground">Demanda</p>
                        <p className="font-medium">{coincidencia.demanda_titulo}</p>
                        <p className="text-muted-foreground">
                          {coincidencia.demanda_empresa} · {coincidencia.demanda_correo}
                        </p>
                      </div>
                      <div>
                        <p className="text-xs uppercase text-muted-foreground">Oferta</p>
                        <p className="font-medium">{coincidencia.oferta_titulo}</p>
                        <p className="text-muted-foreground">
                          {coincidencia.oferta_empresa} · {coincidencia.oferta_correo}
                        </p>
                      </div>
                    </div>
                    <p>{coincidencia.motivo}</p>
                    {coincidencia.notificada_en && (
                      <p className="text-muted-foreground">
                        Notificada el{" "}
                        {new Date(coincidencia.notificada_en).toLocaleDateString("es-PE")}
                      </p>
                    )}
                  </div>
                  {coincidencia.estado === "sugerida" && (
                    <div className="flex items-start gap-2">
                      <Button
                        type="button"
                        size="sm"
                        disabled={notificar.isPending}
                        onClick={() => notificar.mutate(coincidencia.id)}
                        className="text-white"
                      >
                        <Mail className="mr-1 h-4 w-4" aria-hidden="true" />
                        Presentar por correo
                      </Button>
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        disabled={descartar.isPending}
                        onClick={() => descartar.mutate(coincidencia.id)}
                      >
                        Descartar
                      </Button>
                    </div>
                  )}
                </CardContent>
              </Card>
            ))
          )}
        </div>
      )}

      <DialogoRechazo
        abierto={rechazo !== null}
        nombre={rechazo?.nombre ?? ""}
        pendiente={rechazar.isPending}
        alCerrar={() => setRechazo(null)}
        alConfirmar={(motivo) => {
          if (rechazo) {
            rechazar.mutate({ ...rechazo, motivo });
          }
        }}
      />
    </section>
  );
}

"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Download, Loader2, Pencil, Plus, Search, Sparkles, Trash2, Upload } from "lucide-react";
import { useRef, useState } from "react";
import { toast } from "sonner";

import { DialogoConfirmacion } from "@/components/dialogo-confirmacion";
import { PegarContactos } from "@/components/pegar-contactos";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { FormularioContacto } from "@/components/formulario-contacto";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";
import type { Contacto } from "@/lib/tipos";

const POR_PAGINA = 25;

function extra(contacto: Contacto, clave: string): string {
  const valor = contacto.campos_extra[clave];
  return typeof valor === "string" ? valor.trim() : "";
}

function primerEnlace(texto: string): string | null {
  const coincidencia = texto.match(/https?:\/\/[^\s]+/);
  return coincidencia ? coincidencia[0] : null;
}

function dominio(enlace: string): string {
  try {
    return new URL(enlace).hostname.replace(/^www\./, "");
  } catch {
    return enlace;
  }
}

export default function PaginaContactos() {
  const [busqueda, setBusqueda] = useState("");
  const [pagina, setPagina] = useState(0);
  const [formularioAbierto, setFormularioAbierto] = useState(false);
  const [porEliminar, setPorEliminar] = useState<Contacto | null>(null);
  const [pegarAbierto, setPegarAbierto] = useState(false);
  const [consultaIA, setConsultaIA] = useState("");
  const [explicacion, setExplicacion] = useState("");
  const [enEdicion, setEnEdicion] = useState<Contacto | null>(null);
  const referenciaArchivo = useRef<HTMLInputElement>(null);
  const clienteConsultas = useQueryClient();

  const consulta = useQuery({
    queryKey: ["contactos", busqueda, pagina],
    queryFn: () => api.listarContactos(busqueda || undefined, POR_PAGINA, pagina * POR_PAGINA),
  });

  function refrescar() {
    void clienteConsultas.invalidateQueries({ queryKey: ["contactos"] });
  }

  const importacion = useMutation({
    mutationFn: (archivo: File) => api.importarContactos(archivo),
    onSuccess: (datos) => {
      refrescar();
      toast.success(
        `Importación terminada: ${datos.creados} creados, ${datos.actualizados} actualizados`,
        datos.omitidos > 0
          ? { description: `${datos.omitidos} filas omitidas. ${datos.errores.slice(0, 3).join(" · ")}` }
          : undefined,
      );
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo importar el archivo");
    },
  });

  const descargaPlantilla = useMutation({
    mutationFn: () => api.descargarPlantilla(),
    onError: () => {
      toast.error("No se pudo descargar la plantilla");
    },
  });

  const asistente = useQuery({
    queryKey: ["asistente"],
    queryFn: () => api.estadoAsistente(),
    staleTime: Infinity,
  });

  const interpretacion = useMutation({
    mutationFn: () => api.interpretarBusqueda(consultaIA),
    onSuccess: (criterios) => {
      if (criterios.terminos.length === 0) {
        toast.warning("No se entendió qué buscar", {
          description: "Pruebe nombrando el rubro, el país o el tipo de empresa.",
        });
        return;
      }

      setBusqueda(criterios.terminos[0] ?? "");
      setPagina(0);
      setExplicacion(
        criterios.terminos.length > 1
          ? `${criterios.explicacion} Se buscó por "${criterios.terminos[0]}"; otros términos sugeridos: ${criterios.terminos.slice(1).join(", ")}.`
          : criterios.explicacion,
      );
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo interpretar");
    },
  });

  function abrirFormulario(contacto: Contacto | null) {
    setEnEdicion(contacto);
    setFormularioAbierto(true);
  }

  const eliminacion = useMutation({
    mutationFn: (id: string) => api.eliminarContacto(id),
    onSuccess: () => {
      refrescar();
      setPorEliminar(null);
      toast.success("Contacto eliminado");
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo eliminar");
    },
  });

  const total = consulta.data?.total ?? 0;
  const totalPaginas = Math.max(1, Math.ceil(total / POR_PAGINA));
  const contactos = consulta.data?.datos ?? [];

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Contactos</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {consulta.isLoading ? "Cargando" : `${total} empresas registradas`}
          </p>
        </div>

        <div className="flex gap-2">
          <input
            ref={referenciaArchivo}
            type="file"
            accept=".xlsx,.csv,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
            className="hidden"
            onChange={(evento) => {
              const archivo = evento.target.files?.[0];
              if (archivo) {
                importacion.mutate(archivo);
              }
              evento.target.value = "";
            }}
          />
          <Button
            type="button"
            variant="outline"
            onClick={() => referenciaArchivo.current?.click()}
            disabled={importacion.isPending}
          >
            {importacion.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
            ) : (
              <Upload className="mr-2 h-4 w-4" aria-hidden="true" />
            )}
            Importar archivo
          </Button>

          {asistente.data?.disponible && (
            <Button type="button" variant="outline" onClick={() => setPegarAbierto(true)}>
              <Sparkles className="mr-2 h-4 w-4" aria-hidden="true" />
              Pegar contactos
            </Button>
          )}

          <Button type="button" onClick={() => abrirFormulario(null)}>
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            Nuevo contacto
          </Button>
        </div>
      </div>

      <Card className="border-dashed bg-white">
        <CardContent className="flex flex-wrap items-start justify-between gap-4 p-5">
          <div className="max-w-2xl space-y-2">
            <p className="text-sm font-medium">Cómo preparar el archivo</p>
            <p className="text-sm text-muted-foreground">
              Suba su archivo de Excel (.xlsx) o un CSV. Solo son obligatorias dos columnas:{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">EMPRESA</code> y{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">CORREO</code>. El orden y las
              mayúsculas no importan. Si no hay columna de nombre de persona, se usa el nombre de la
              empresa.
            </p>
            <p className="text-sm text-muted-foreground">
              Las demás columnas, como{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">RUC</code>,{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">CIUDAD/REGIÓN</code>,{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">TELEFONO MOVIL</code> o{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">PÁGINA WEB</code>, se guardan y
              quedan disponibles como variables en las cartas:{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">{"{ruc}"}</code>,{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">{"{ciudad_region}"}</code>,{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">{"{telefono_movil}"}</code>,{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">{"{pagina_web}"}</code>.
            </p>
            <p className="text-sm text-muted-foreground">
              Si un correo ya existe, el contacto se actualiza en lugar de duplicarse.
            </p>
          </div>

          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => descargaPlantilla.mutate()}
            disabled={descargaPlantilla.isPending}
          >
            {descargaPlantilla.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
            ) : (
              <Download className="mr-2 h-4 w-4" aria-hidden="true" />
            )}
            Descargar plantilla Excel
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="pb-4">
          <div className="relative max-w-md">
            <Search
              className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden="true"
            />
            <Label htmlFor="busqueda" className="sr-only">
              Buscar por nombre, empresa o correo
            </Label>
            <Input
              id="busqueda"
              type="search"
              value={busqueda}
              onChange={(evento) => {
                setBusqueda(evento.target.value);
                setPagina(0);
              }}
              className="pl-9"
            />
          </div>

          {asistente.data?.disponible && (
            <div className="mt-3 space-y-2">
              <div className="flex flex-wrap items-end gap-2">
                <div className="min-w-64 flex-1 space-y-1.5">
                  <Label htmlFor="consulta" className="text-xs text-muted-foreground">
                    O descríbalo con sus palabras: molinos de Brasil, transportistas del sur
                  </Label>
                  <Input
                    id="consulta"
                    value={consultaIA}
                    onChange={(evento) => setConsultaIA(evento.target.value)}
                    onKeyDown={(evento) => {
                      if (evento.key === "Enter" && consultaIA.trim().length > 2) {
                        interpretacion.mutate();
                      }
                    }}
                  />
                </div>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => interpretacion.mutate()}
                  disabled={interpretacion.isPending || consultaIA.trim().length < 3}
                >
                  {interpretacion.isPending ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
                  ) : (
                    <Sparkles className="mr-2 h-4 w-4" aria-hidden="true" />
                  )}
                  Buscar
                </Button>
              </div>

              {explicacion && (
                <p className="text-xs text-emerald-800">{explicacion}</p>
              )}
            </div>
          )}
        </CardHeader>

        <CardContent>
          {consulta.isLoading ? (
            <div className="space-y-2">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : contactos.length === 0 ? (
            <div className="py-12 text-center">
              <p className="text-sm font-medium">
                {busqueda ? "Ninguna empresa coincide con la búsqueda" : "Todavía no hay contactos"}
              </p>
              <p className="mt-1 text-sm text-muted-foreground">
                {busqueda
                  ? "Pruebe con otro nombre, empresa o correo."
                  : "Importe su Excel como CSV con las columnas nombre, empresa y correo."}
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Empresa</TableHead>
                  <TableHead>Ciudad / región</TableHead>
                  <TableHead>Teléfonos</TableHead>
                  <TableHead>Correo</TableHead>
                  <TableHead>Web</TableHead>
                  <TableHead className="w-12" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {contactos.map((contacto) => (
                  <TableRow key={contacto.id}>
                    <TableCell>
                      <span className="block font-medium">{contacto.empresa}</span>
                      {extra(contacto, "ruc") && (
                        <span className="block text-xs text-muted-foreground">
                          RUC {extra(contacto, "ruc")}
                        </span>
                      )}
                      {contacto.nombre !== contacto.empresa && (
                        <span className="block text-xs text-muted-foreground">
                          {contacto.nombre}
                        </span>
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {extra(contacto, "ciudad_region") || contacto.pais || "—"}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {extra(contacto, "telefono_movil") || extra(contacto, "telefono_fijo") ? (
                        <>
                          {extra(contacto, "telefono_movil") && (
                            <span className="block whitespace-nowrap">
                              {extra(contacto, "telefono_movil")}
                            </span>
                          )}
                          {extra(contacto, "telefono_fijo") && (
                            <span className="block whitespace-nowrap text-xs">
                              {extra(contacto, "telefono_fijo")}
                            </span>
                          )}
                        </>
                      ) : (
                        "—"
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground">{contacto.correo}</TableCell>
                    <TableCell>
                      {(() => {
                        const enlace = primerEnlace(extra(contacto, "pagina_web"));
                        return enlace ? (
                          <a
                            href={enlace}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-sm text-emerald-700 hover:underline"
                          >
                            {dominio(enlace)}
                          </a>
                        ) : (
                          <span className="text-muted-foreground">—</span>
                        );
                      })()}
                    </TableCell>
                    <TableCell className="whitespace-nowrap text-right">
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => abrirFormulario(contacto)}
                        aria-label={`Editar ${contacto.empresa}`}
                        className="h-8 w-8 text-muted-foreground hover:text-foreground"
                      >
                        <Pencil className="h-4 w-4" aria-hidden="true" />
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => setPorEliminar(contacto)}
                        aria-label={`Eliminar ${contacto.empresa}`}
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                      >
                        <Trash2 className="h-4 w-4" aria-hidden="true" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}

          {total > POR_PAGINA && (
            <div className="flex items-center justify-between pt-4">
              <p className="text-sm text-muted-foreground">
                Página {pagina + 1} de {totalPaginas}
              </p>
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={pagina === 0}
                  onClick={() => setPagina((actual) => Math.max(0, actual - 1))}
                >
                  Anterior
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={pagina + 1 >= totalPaginas}
                  onClick={() => setPagina((actual) => actual + 1)}
                >
                  Siguiente
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <FormularioContacto
        abierto={formularioAbierto}
        contacto={enEdicion}
        alCambiar={(abierto) => {
          setFormularioAbierto(abierto);
          if (!abierto) {
            setEnEdicion(null);
          }
        }}
        alGuardar={refrescar}
      />

      <PegarContactos
        abierto={pegarAbierto}
        alCambiar={setPegarAbierto}
        alImportar={refrescar}
      />

      <DialogoConfirmacion
        abierto={porEliminar !== null}
        titulo="Eliminar contacto"
        descripcion={
          porEliminar
            ? `Se quitará ${porEliminar.empresa} (${porEliminar.correo}) de la base y de todas las listas donde esté. Esta acción no se puede deshacer.`
            : ""
        }
        procesando={eliminacion.isPending}
        alCambiar={(abierto) => {
          if (!abierto) {
            setPorEliminar(null);
          }
        }}
        alConfirmar={() => {
          if (porEliminar) {
            eliminacion.mutate(porEliminar.id);
          }
        }}
      />
    </section>
  );
}

"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Download, Loader2, Plus, Search, Sparkles, Trash2, Upload } from "lucide-react";
import { useRef, useState } from "react";
import { toast } from "sonner";

import { DialogoConfirmacion } from "@/components/dialogo-confirmacion";
import { PegarContactos } from "@/components/pegar-contactos";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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

export default function PaginaContactos() {
  const [busqueda, setBusqueda] = useState("");
  const [pagina, setPagina] = useState(0);
  const [formularioAbierto, setFormularioAbierto] = useState(false);
  const [porEliminar, setPorEliminar] = useState<Contacto | null>(null);
  const [pegarAbierto, setPegarAbierto] = useState(false);
  const [consultaIA, setConsultaIA] = useState("");
  const [explicacion, setExplicacion] = useState("");
  const [nombre, setNombre] = useState("");
  const [empresa, setEmpresa] = useState("");
  const [correo, setCorreo] = useState("");
  const [pais, setPais] = useState("");
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

  const creacion = useMutation({
    mutationFn: () =>
      api.crearContacto({ nombre, empresa, correo, pais: pais.trim() ? pais : null }),
    onSuccess: (contacto) => {
      refrescar();
      setFormularioAbierto(false);
      setNombre("");
      setEmpresa("");
      setCorreo("");
      setPais("");
      toast.success(`${contacto.empresa} agregada a la base de contactos`);
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo guardar el contacto");
    },
  });

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

          <Button type="button" onClick={() => setFormularioAbierto(true)}>
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
              Suba su archivo de Excel (.xlsx) o un CSV. Solo son obligatorias tres columnas:{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">nombre</code>,{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">empresa</code> y{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">correo</code>. El orden no
              importa y puede usar <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">email</code>{" "}
              en vez de correo.
            </p>
            <p className="text-sm text-muted-foreground">
              Cualquier otra columna, como{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">pais</code>,{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">cargo</code> o{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">ruc</code>, se guarda y queda
              disponible como variable en las cartas, por ejemplo{" "}
              <code className="rounded bg-zinc-100 px-1 py-0.5 text-xs">{"{cargo}"}</code>.
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
                  <TableHead>Contacto</TableHead>
                  <TableHead>Correo</TableHead>
                  <TableHead>País</TableHead>
                  <TableHead className="w-12" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {contactos.map((contacto) => (
                  <TableRow key={contacto.id}>
                    <TableCell className="font-medium">{contacto.empresa}</TableCell>
                    <TableCell>{contacto.nombre}</TableCell>
                    <TableCell className="text-muted-foreground">{contacto.correo}</TableCell>
                    <TableCell className="text-muted-foreground">{contacto.pais ?? "—"}</TableCell>
                    <TableCell>
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

      <Dialog open={formularioAbierto} onOpenChange={setFormularioAbierto}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Nuevo contacto</DialogTitle>
            <DialogDescription>
              Registre una empresa que no esté en el archivo de importación.
            </DialogDescription>
          </DialogHeader>

          <form
            id="formulario-contacto"
            onSubmit={(evento) => {
              evento.preventDefault();
              creacion.mutate();
            }}
            className="space-y-4"
          >
            <div className="space-y-1.5">
              <Label htmlFor="empresa">Empresa</Label>
              <Input
                id="empresa"
                value={empresa}
                onChange={(evento) => setEmpresa(evento.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="nombre">Nombre del contacto</Label>
              <Input
                id="nombre"
                value={nombre}
                onChange={(evento) => setNombre(evento.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="correo">Correo</Label>
              <Input
                id="correo"
                type="email"
                value={correo}
                onChange={(evento) => setCorreo(evento.target.value)}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="pais">País</Label>
              <Input id="pais" value={pais} onChange={(evento) => setPais(evento.target.value)} />
            </div>
          </form>

          <DialogFooter className="gap-2 sm:gap-2">
            <Button type="button" variant="outline" onClick={() => setFormularioAbierto(false)}>
              Cancelar
            </Button>
            <Button
              type="submit"
              form="formulario-contacto"
              disabled={
                creacion.isPending ||
                empresa.trim().length < 2 ||
                nombre.trim().length < 2 ||
                !correo.includes("@")
              }
              className="text-white"
            >
              {creacion.isPending && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
              )}
              Guardar contacto
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

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

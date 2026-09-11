"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AnimatePresence, motion } from "framer-motion";
import {
  ArrowLeft,
  FileText,
  FileUp,
  Loader2,
  MailCheck,
  Paperclip,
  Pencil,
  Send,
  Sparkles,
  Trash2,
  TriangleAlert,
  X,
} from "lucide-react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

import { DialogoConfirmacion } from "@/components/dialogo-confirmacion";
import { RevisionCarta } from "@/components/revision-carta";
import { VistaPreviaWord } from "@/components/vista-previa-word";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";
import {
  colorEstadoCorrespondencia,
  colorEstadoEnvio,
  etiquetaEstadoCorrespondencia,
  etiquetaEstadoEnvio,
} from "@/lib/estados";
import type { EstadoEnvio } from "@/lib/tipos";

const SIN_LISTA = "sin-lista";

export default function PaginaDetalleCorrespondencia() {
  const parametros = useParams<{ id: string }>();
  const id = parametros.id;
  const router = useRouter();
  const clienteConsultas = useQueryClient();

  const [editando, setEditando] = useState(false);
  const [asunto, setAsunto] = useState("");
  const [cuerpo, setCuerpo] = useState("");
  const [listaId, setListaId] = useState(SIN_LISTA);
  const [confirmarEnvio, setConfirmarEnvio] = useState(false);
  const [confirmarBorrado, setConfirmarBorrado] = useState(false);
  const referenciaArchivo = useRef<HTMLInputElement>(null);
  const referenciaPlantilla = useRef<HTMLInputElement>(null);

  const pieza = useQuery({
    queryKey: ["correspondencia", id],
    queryFn: () => api.obtenerCorrespondencia(id),
  });

  const listas = useQuery({ queryKey: ["listas"], queryFn: () => api.listarListas() });

  const esBorrador = pieza.data?.estado === "borrador";

  const previsualizacion = useQuery({
    queryKey: ["previsualizacion", id, pieza.data?.actualizado_en],
    queryFn: () => api.previsualizar(id),
    enabled: pieza.data !== undefined && pieza.data.lista_id !== null && !editando,
    retry: false,
  });

  const resumen = useQuery({
    queryKey: ["resumen", id],
    queryFn: () => api.resumenEnvios(id),
    enabled: pieza.data !== undefined && !esBorrador,
    refetchInterval: pieza.data && !esBorrador ? 5000 : false,
  });

  const envios = useQuery({
    queryKey: ["envios", id],
    queryFn: () => api.listarEnvios(id),
    enabled: pieza.data !== undefined && !esBorrador,
    refetchInterval: pieza.data && !esBorrador ? 5000 : false,
  });

  useEffect(() => {
    if (pieza.data && !editando) {
      setAsunto(pieza.data.asunto);
      setCuerpo(pieza.data.cuerpo);
      setListaId(pieza.data.lista_id ?? SIN_LISTA);
    }
  }, [pieza.data, editando]);

  function refrescar() {
    void clienteConsultas.invalidateQueries({ queryKey: ["correspondencia", id] });
    void clienteConsultas.invalidateQueries({ queryKey: ["previsualizacion", id] });
  }

  const guardado = useMutation({
    mutationFn: () =>
      api.actualizarCorrespondencia(id, {
        asunto,
        cuerpo,
        lista_id: listaId === SIN_LISTA ? null : listaId,
      }),
    onSuccess: () => {
      refrescar();
      setEditando(false);
      toast.success("Cambios guardados");
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo guardar");
    },
  });

  const subidaPlantilla = useMutation({
    mutationFn: (archivo: File) => api.subirPlantilla(id, archivo),
    onSuccess: () => {
      refrescar();
      toast.success("Plantilla cargada", {
        description: "Cada empresa recibirá su propia versión del documento.",
      });
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo subir la plantilla");
    },
  });

  const quitadoPlantilla = useMutation({
    mutationFn: () => api.quitarPlantilla(id),
    onSuccess: () => {
      refrescar();
      toast.success("Plantilla quitada");
    },
  });

  const subida = useMutation({
    mutationFn: (archivo: File) => api.subirAdjunto(id, archivo),
    onSuccess: (adjunto) => {
      refrescar();
      toast.success(`${adjunto.nombre_archivo} adjuntado`);
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo subir el archivo");
    },
  });

  const borradoAdjunto = useMutation({
    mutationFn: (adjuntoId: string) => api.eliminarAdjunto(id, adjuntoId),
    onSuccess: () => {
      refrescar();
      toast.success("Adjunto quitado");
    },
  });

  const borrado = useMutation({
    mutationFn: () => api.eliminarCorrespondencia(id),
    onSuccess: () => {
      void clienteConsultas.invalidateQueries({ queryKey: ["correspondencia"] });
      toast.success("Borrador eliminado");
      router.push("/correspondencia");
    },
    onError: (fallo: unknown) => {
      setConfirmarBorrado(false);
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo eliminar");
    },
  });

  const asistente = useQuery({
    queryKey: ["asistente"],
    queryFn: () => api.estadoAsistente(),
    staleTime: Infinity,
  });

  const revision = useMutation({
    mutationFn: () =>
      api.revisarConAsistente({
        asunto: pieza.data?.asunto ?? "",
        cuerpo: pieza.data?.cuerpo ?? "",
      }),
    onSuccess: (resultado) => {
      if (resultado.veredicto === "lista") {
        toast.success("La carta está lista para enviar");
        return;
      }
      toast.warning(
        resultado.veredicto === "no_enviar"
          ? "El asistente recomienda no enviarla todavía"
          : `Se encontraron ${resultado.hallazgos.length} detalles por revisar`,
      );
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo revisar la carta");
    },
  });

  const prueba = useMutation({
    mutationFn: () => api.enviarPrueba(id),
    onSuccess: (resultado) => {
      toast.success(
        resultado.simulada
          ? "Prueba simulada, no se envió ningún correo real"
          : `Prueba enviada a ${resultado.destinatario}`,
        {
          description: resultado.simulada
            ? "El sistema está en modo bitácora. Configure Resend para recibirla."
            : "Revise su bandeja antes de enviar a toda la lista.",
        },
      );
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo enviar la prueba");
    },
  });

  const envio = useMutation({
    mutationFn: () => api.enviar(id),
    onSuccess: (resultado) => {
      refrescar();
      setConfirmarEnvio(false);
      toast.success(`${resultado.encolados} correos encolados para envío`);
    },
    onError: (fallo: unknown) => {
      setConfirmarEnvio(false);
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo iniciar el envío");
    },
  });

  if (pieza.isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-9 w-64" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  if (!pieza.data) {
    return <p className="text-sm text-destructive">No se encontró la correspondencia</p>;
  }

  const datos = pieza.data;
  const variablesSinValor = previsualizacion.data?.variables_sin_valor ?? [];
  const listaAsignada = listas.data?.datos.find((lista) => lista.id === datos.lista_id);
  const listaVacia = listaAsignada !== undefined && listaAsignada.total_contactos === 0;

  return (
    <section className="space-y-6">
      <Button asChild variant="ghost" size="sm" className="-ml-2 text-muted-foreground">
        <Link href="/correspondencia">
          <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
          Volver a correspondencia
        </Link>
      </Button>

      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-semibold tracking-tight">{datos.asunto}</h1>
            <Badge variant="secondary" className={colorEstadoCorrespondencia[datos.estado]}>
              {etiquetaEstadoCorrespondencia[datos.estado]}
            </Badge>
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            {listaAsignada
              ? `Dirigida a ${listaAsignada.nombre} · ${listaAsignada.total_contactos} empresas`
              : "Sin lista de destinatarios asignada"}
          </p>
        </div>

        {esBorrador && !editando && (
          <div className="flex gap-2">
            {asistente.data?.disponible && (
              <Button
                type="button"
                variant="outline"
                onClick={() => revision.mutate()}
                disabled={revision.isPending}
              >
                {revision.isPending ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
                ) : (
                  <Sparkles className="mr-2 h-4 w-4" aria-hidden="true" />
                )}
                Revisar
              </Button>
            )}
            <Button
              type="button"
              variant="outline"
              onClick={() => prueba.mutate()}
              disabled={prueba.isPending}
            >
              {prueba.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
              ) : (
                <MailCheck className="mr-2 h-4 w-4" aria-hidden="true" />
              )}
              Enviar prueba
            </Button>
            <Button type="button" variant="outline" onClick={() => setEditando(true)}>
              <Pencil className="mr-2 h-4 w-4" aria-hidden="true" />
              Editar
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => setConfirmarBorrado(true)}
              className="text-destructive hover:text-destructive"
            >
              <Trash2 className="h-4 w-4" aria-hidden="true" />
              <span className="sr-only">Eliminar borrador</span>
            </Button>
            <Button
              type="button"
              onClick={() => setConfirmarEnvio(true)}
              disabled={datos.lista_id === null || listaVacia}
              className="text-white"
            >
              <Send className="mr-2 h-4 w-4" aria-hidden="true" />
              Enviar
            </Button>
          </div>
        )}
      </div>

      {esBorrador && datos.lista_id === null && (
        <Alert>
          <TriangleAlert className="h-4 w-4" aria-hidden="true" />
          <AlertDescription>
            Asigne una lista de destinatarios para poder previsualizar y enviar esta carta.
          </AlertDescription>
        </Alert>
      )}

      {esBorrador && listaVacia && (
        <Alert>
          <TriangleAlert className="h-4 w-4" aria-hidden="true" />
          <AlertDescription>
            La lista {listaAsignada?.nombre} no tiene empresas.{" "}
            <Link href={`/listas/${datos.lista_id}`} className="font-medium underline">
              Agregue destinatarios
            </Link>{" "}
            antes de enviar.
          </AlertDescription>
        </Alert>
      )}

      {variablesSinValor.length > 0 && (
        <Alert className="border-amber-200 bg-amber-50 text-amber-900">
          <TriangleAlert className="h-4 w-4" aria-hidden="true" />
          <AlertDescription>
            Estas variables quedarían sin reemplazar: {variablesSinValor.join(", ")}. Revise que los
            contactos tengan ese dato o quite la variable del texto.
          </AlertDescription>
        </Alert>
      )}

      {editando ? (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Editar borrador</CardTitle>
            <CardDescription>
              Puede usar campos de Word como «NOMBRES» o variables entre llaves como{" "}
              {"{empresa}"}.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="lista">Lista de destinatarios</Label>
              <Select value={listaId} onValueChange={setListaId}>
                <SelectTrigger id="lista">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={SIN_LISTA}>Sin lista asignada</SelectItem>
                  {listas.data?.datos.map((lista) => (
                    <SelectItem key={lista.id} value={lista.id}>
                      {lista.nombre} ({lista.total_contactos} empresas)
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="asunto">Asunto</Label>
              <Input
                id="asunto"
                value={asunto}
                onChange={(evento) => setAsunto(evento.target.value)}
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="cuerpo">Cuerpo del mensaje</Label>
              <Textarea
                id="cuerpo"
                rows={14}
                value={cuerpo}
                onChange={(evento) => setCuerpo(evento.target.value)}
                className="font-mono"
              />
            </div>

            <div className="flex gap-2">
              <Button
                type="button"
                onClick={() => guardado.mutate()}
                disabled={
                  guardado.isPending || asunto.trim().length < 3 || cuerpo.trim().length < 10
                }
                className="text-white"
              >
                {guardado.isPending && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
                )}
                Guardar cambios
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setEditando(false);
                  setAsunto(datos.asunto);
                  setCuerpo(datos.cuerpo);
                  setListaId(datos.lista_id ?? SIN_LISTA);
                }}
              >
                Cancelar
              </Button>
            </div>
          </CardContent>
        </Card>
      ) : (
        previsualizacion.data && (
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Previsualización</CardTitle>
              <CardDescription>
                Así lo recibirá {previsualizacion.data.destinatario}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm font-medium">{previsualizacion.data.asunto}</p>
              <pre className="mt-3 whitespace-pre-wrap font-sans text-sm text-zinc-700">
                {previsualizacion.data.cuerpo}
              </pre>
            </CardContent>
          </Card>
        )
      )}

      {revision.data && <RevisionCarta revision={revision.data} />}

      {esBorrador && !editando && (
        <Card className={datos.plantilla_url ? "border-emerald-200 bg-emerald-50/40" : ""}>
          <CardHeader>
            <CardTitle className="text-base">Carta en Word personalizada</CardTitle>
            <CardDescription>
              Suba el documento con el membrete de CNI. Cada empresa recibirá su propia copia, con
              los campos de combinación como «NOMBRES» ya reemplazados y el formato original
              intacto.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {datos.plantilla_nombre ? (
              <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-emerald-200 bg-white px-4 py-3">
                <span className="flex min-w-0 items-center gap-2">
                  <FileText className="h-4 w-4 shrink-0 text-emerald-700" aria-hidden="true" />
                  <span className="truncate text-sm font-medium">{datos.plantilla_nombre}</span>
                  <Badge variant="secondary" className="bg-emerald-100 text-emerald-800">
                    Se personaliza por empresa
                  </Badge>
                </span>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => quitadoPlantilla.mutate()}
                  disabled={quitadoPlantilla.isPending}
                  className="text-muted-foreground hover:text-destructive"
                >
                  Quitar
                </Button>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                Sin plantilla. Si no sube ninguna, la carta se enviará solo como texto del correo.
              </p>
            )}

            {datos.plantilla_nombre && (
              <VistaPreviaWord
                correspondenciaId={datos.id}
                listaId={datos.lista_id}
                nombrePlantilla={datos.plantilla_nombre}
                versionPlantilla={datos.actualizado_en}
              />
            )}

            <input
              ref={referenciaPlantilla}
              type="file"
              accept=".docx"
              className="hidden"
              onChange={(evento) => {
                const archivo = evento.target.files?.[0];
                if (archivo) {
                  subidaPlantilla.mutate(archivo);
                }
                evento.target.value = "";
              }}
            />
            <Button
              type="button"
              variant="outline"
              onClick={() => referenciaPlantilla.current?.click()}
              disabled={subidaPlantilla.isPending}
              className="mt-4"
            >
              {subidaPlantilla.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
              ) : (
                <FileUp className="mr-2 h-4 w-4" aria-hidden="true" />
              )}
              {datos.plantilla_nombre ? "Reemplazar plantilla" : "Subir carta en Word"}
            </Button>
          </CardContent>
        </Card>
      )}

      {esBorrador && !editando && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Documentos adjuntos</CardTitle>
            <CardDescription>
              Archivos que se envían idénticos a todas las empresas, sin personalizar.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {datos.adjuntos.length === 0 ? (
              <p className="text-sm text-muted-foreground">Sin documentos adjuntos</p>
            ) : (
              <ul className="divide-y divide-zinc-100">
                {datos.adjuntos.map((adjunto) => (
                  <li key={adjunto.id} className="flex items-center justify-between gap-3 py-2.5">
                    <span className="flex min-w-0 items-center gap-2">
                      <Paperclip
                        className="h-4 w-4 shrink-0 text-muted-foreground"
                        aria-hidden="true"
                      />
                      <span className="min-w-0">
                        <span className="block truncate text-sm">{adjunto.nombre_archivo}</span>
                        {adjunto.nombre_archivo.toLowerCase().endsWith(".docx") && (
                          <span className="block text-xs text-amber-700">
                            Este Word llega igual a todos: sus campos como «NOMBRES» no se
                            reemplazan. Para personalizarlo, súbalo en Carta en Word personalizada.
                          </span>
                        )}
                      </span>
                      <span className="shrink-0 text-xs text-muted-foreground">
                        {Math.round(adjunto.tamano_bytes / 1024)} KB
                      </span>
                    </span>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      onClick={() => borradoAdjunto.mutate(adjunto.id)}
                      aria-label={`Quitar ${adjunto.nombre_archivo}`}
                      className="h-8 w-8 shrink-0 text-muted-foreground hover:text-destructive"
                    >
                      <X className="h-4 w-4" aria-hidden="true" />
                    </Button>
                  </li>
                ))}
              </ul>
            )}

            <input
              ref={referenciaArchivo}
              type="file"
              accept=".pdf,.doc,.docx,.xls,.xlsx,.png,.jpg,.jpeg"
              className="hidden"
              onChange={(evento) => {
                const archivo = evento.target.files?.[0];
                if (archivo) {
                  subida.mutate(archivo);
                }
                evento.target.value = "";
              }}
            />
            <Button
              type="button"
              variant="outline"
              onClick={() => referenciaArchivo.current?.click()}
              disabled={subida.isPending}
              className="mt-4"
            >
              {subida.isPending ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
              ) : (
                <Paperclip className="mr-2 h-4 w-4" aria-hidden="true" />
              )}
              Adjuntar documento
            </Button>
          </CardContent>
        </Card>
      )}

      {!esBorrador && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Documentos de esta carta</CardTitle>
            <CardDescription>Lo que viajó adjunto en cada correo.</CardDescription>
          </CardHeader>
          <CardContent>
            {!datos.plantilla_nombre && datos.adjuntos.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                Esta carta se envió sin documentos adjuntos.
              </p>
            ) : (
              <ul className="divide-y divide-zinc-100">
                {datos.plantilla_nombre && datos.plantilla_url && (
                  <li className="flex flex-wrap items-center justify-between gap-3 py-2.5">
                    <span className="flex min-w-0 items-center gap-2">
                      <FileText className="h-4 w-4 shrink-0 text-emerald-700" aria-hidden="true" />
                      <span className="truncate text-sm font-medium">{datos.plantilla_nombre}</span>
                      <Badge variant="secondary" className="bg-emerald-100 text-emerald-800">
                        Personalizado por empresa
                      </Badge>
                    </span>
                    <a
                      href={datos.plantilla_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-emerald-700 hover:underline"
                    >
                      Descargar plantilla
                    </a>
                  </li>
                )}
                {datos.adjuntos.map((adjunto) => (
                  <li key={adjunto.id} className="flex flex-wrap items-center justify-between gap-3 py-2.5">
                    <span className="flex min-w-0 items-center gap-2">
                      <Paperclip className="h-4 w-4 shrink-0 text-muted-foreground" aria-hidden="true" />
                      <span className="truncate text-sm">{adjunto.nombre_archivo}</span>
                      <span className="shrink-0 text-xs text-muted-foreground">
                        {Math.round(adjunto.tamano_bytes / 1024)} KB · igual para todos
                      </span>
                    </span>
                    <a
                      href={adjunto.url_archivo}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-sm text-emerald-700 hover:underline"
                    >
                      Descargar
                    </a>
                  </li>
                ))}
              </ul>
            )}

            {datos.plantilla_nombre && (
              <VistaPreviaWord
                correspondenciaId={datos.id}
                listaId={datos.lista_id}
                nombrePlantilla={datos.plantilla_nombre}
                versionPlantilla={datos.actualizado_en}
              />
            )}
          </CardContent>
        </Card>
      )}

      <AnimatePresence>
        {!esBorrador && resumen.data && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.35 }}
            className="space-y-4"
          >
            <div>
              <h2 className="text-base font-semibold">Resultados del envío</h2>
              <div className="mt-3 flex flex-wrap gap-2">
                <Badge variant="outline" className="px-3 py-1.5 text-sm">
                  Total: {resumen.data.total}
                </Badge>
                {(Object.keys(etiquetaEstadoEnvio) as EstadoEnvio[])
                  .filter((estado) => resumen.data.por_estado[estado])
                  .map((estado) => (
                    <Badge
                      key={estado}
                      variant="secondary"
                      className={`px-3 py-1.5 text-sm ${colorEstadoEnvio[estado]}`}
                    >
                      {etiquetaEstadoEnvio[estado]}: {resumen.data.por_estado[estado]}
                    </Badge>
                  ))}
              </div>
            </div>

            <Card>
              <CardContent className="p-0">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Empresa</TableHead>
                      <TableHead>Correo</TableHead>
                      <TableHead>Estado</TableHead>
                      <TableHead>Fecha</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {envios.data?.datos.map((registro) => (
                      <TableRow key={registro.id}>
                        <TableCell className="font-medium">{registro.empresa}</TableCell>
                        <TableCell className="text-muted-foreground">{registro.correo}</TableCell>
                        <TableCell>
                          <Badge
                            variant="secondary"
                            className={colorEstadoEnvio[registro.estado]}
                          >
                            {etiquetaEstadoEnvio[registro.estado]}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {registro.fecha_envio
                            ? new Date(registro.fecha_envio).toLocaleString("es-PE")
                            : "—"}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          </motion.div>
        )}
      </AnimatePresence>

      <DialogoConfirmacion
        abierto={confirmarEnvio}
        titulo="Confirmar el envío"
        descripcion={`Se enviará esta carta a ${listaAsignada?.total_contactos ?? 0} empresas de la lista ${listaAsignada?.nombre ?? ""}. Una vez enviada no podrá modificarla.`}
        textoConfirmar="Enviar ahora"
        procesando={envio.isPending}
        alCambiar={setConfirmarEnvio}
        alConfirmar={() => envio.mutate()}
      />

      <DialogoConfirmacion
        abierto={confirmarBorrado}
        titulo="Eliminar borrador"
        descripcion="Se eliminará esta carta junto con sus documentos adjuntos. Esta acción no se puede deshacer."
        procesando={borrado.isPending}
        alCambiar={setConfirmarBorrado}
        alConfirmar={() => borrado.mutate()}
      />
    </section>
  );
}

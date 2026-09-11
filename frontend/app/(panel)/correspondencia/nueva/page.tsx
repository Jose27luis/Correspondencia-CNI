"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { ArrowLeft, FileUp, Loader2, Sparkles } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";
import { toast } from "sonner";

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
import { Textarea } from "@/components/ui/textarea";
import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";

const SIN_LISTA = "sin-lista";

const variablesDisponibles = ["nombre", "empresa", "correo", "pais"];

export default function PaginaNuevaCorrespondencia() {
  const router = useRouter();
  const [asunto, setAsunto] = useState("");
  const [cuerpo, setCuerpo] = useState("");
  const [listaId, setListaId] = useState(SIN_LISTA);
  const [variablesDetectadas, setVariablesDetectadas] = useState<string[]>([]);
  const [documentoSinVariables, setDocumentoSinVariables] = useState(false);
  const [textoDocumento, setTextoDocumento] = useState("");
  const [resumenSugerencia, setResumenSugerencia] = useState("");
  const [cuerpoAnterior, setCuerpoAnterior] = useState<string | null>(null);
  const [instruccion, setInstruccion] = useState("");
  const referenciaWord = useRef<HTMLInputElement>(null);
  const referenciaWordIA = useRef<HTMLInputElement>(null);
  const [archivoWord, setArchivoWord] = useState<File | null>(null);

  const listas = useQuery({ queryKey: ["listas"], queryFn: () => api.listarListas() });

  const asistente = useQuery({
    queryKey: ["asistente"],
    queryFn: () => api.estadoAsistente(),
    staleTime: Infinity,
  });

  const redaccion = useMutation({
    mutationFn: (mejorar: boolean) =>
      api.redactarConAsistente({
        instruccion,
        asunto: mejorar ? asunto : undefined,
        cuerpo: mejorar ? cuerpo : undefined,
        variables: variablesDisponibles,
      }),
    onSuccess: (borrador) => {
      if (borrador.asunto) {
        setAsunto(borrador.asunto);
      }
      setCuerpo(borrador.cuerpo);
      toast.success("Borrador preparado", {
        description: "Revíselo y ajústelo antes de guardarlo.",
      });
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo redactar el borrador");
    },
  });

  const lectura = useMutation({
    mutationFn: (archivo: File) => api.leerDocumento(archivo),
    onSuccess: (contenido, archivo) => {
      setCuerpo(contenido.cuerpo);
      setTextoDocumento(contenido.cuerpo);
      setArchivoWord(archivo);
      setResumenSugerencia("");
      setCuerpoAnterior(null);
      setVariablesDetectadas(contenido.variables);
      setDocumentoSinVariables(contenido.variables.length === 0);
      toast.success("Documento leído", {
        description:
          contenido.variables.length > 0
            ? `Se encontraron ${contenido.variables.length} variables.`
            : "No se encontraron variables en el texto.",
      });
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo leer el documento");
    },
  });

  const lecturaConIA = useMutation({
    mutationFn: async (archivo: File) => {
      const contenido = await api.leerDocumento(archivo);
      const resultado = await api.sugerirCuerpo({
        documento: contenido.cuerpo,
        asunto,
        variables: variablesDisponibles,
      });
      return { contenido, resultado };
    },
    onSuccess: ({ contenido, resultado }, archivo) => {
      setTextoDocumento(contenido.cuerpo);
      setVariablesDetectadas(contenido.variables);
      setDocumentoSinVariables(contenido.variables.length === 0);
      setArchivoWord(archivo);
      setCuerpoAnterior(cuerpo);
      setCuerpo(resultado.cuerpo);
      if (!asunto.trim() && resultado.asunto) {
        setAsunto(resultado.asunto);
      }
      setResumenSugerencia(resultado.resumen);
      toast.success("Documento analizado", {
        description: "Revise el cuerpo sugerido antes de guardar.",
      });
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo analizar el documento");
    },
  });

  const fuenteAnalisis = textoDocumento.trim() || cuerpo.trim();

  const sugerencia = useMutation({
    mutationFn: () =>
      api.sugerirCuerpo({
        documento: fuenteAnalisis,
        asunto,
        variables: variablesDisponibles,
      }),
    onSuccess: (resultado) => {
      setCuerpoAnterior(cuerpo);
      setCuerpo(resultado.cuerpo);
      if (!asunto.trim() && resultado.asunto) {
        setAsunto(resultado.asunto);
      }
      setResumenSugerencia(resultado.resumen);
      toast.success("Cuerpo sugerido", {
        description: "Revíselo antes de guardar. Puede deshacer el cambio.",
      });
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo analizar el documento");
    },
  });

  const creacion = useMutation({
    mutationFn: () =>
      api.crearCorrespondencia({
        asunto,
        cuerpo,
        lista_id: listaId === SIN_LISTA ? null : listaId,
      }),
    onSuccess: async (pieza) => {
      if (archivoWord) {
        try {
          await api.subirPlantilla(pieza.id, archivoWord);
          toast.success("Borrador guardado", {
            description: `${archivoWord.name} se adjuntará personalizado para cada empresa.`,
          });
        } catch (fallo: unknown) {
          toast.warning("Borrador guardado, pero no se pudo adjuntar el Word", {
            description:
              fallo instanceof ErrorPeticion
                ? fallo.message
                : "Súbalo en Carta en Word personalizada.",
          });
        }
      } else {
        toast.success("Borrador guardado");
      }
      router.push(`/correspondencia/${pieza.id}`);
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo guardar el borrador");
    },
  });

  function insertarVariable(variable: string) {
    setCuerpo((actual) => `${actual}{${variable}}`);
  }

  return (
    <section className="max-w-3xl space-y-6">
      <Button asChild variant="ghost" size="sm" className="-ml-2 text-muted-foreground">
        <Link href="/correspondencia">
          <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
          Volver a correspondencia
        </Link>
      </Button>

      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Nueva carta</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Redacte una vez y el sistema personalizará el texto para cada empresa.
        </p>
      </div>

      {asistente.data?.disponible && (
        <Card className="border-emerald-200 bg-emerald-50/40">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Sparkles className="h-4 w-4 text-emerald-700" aria-hidden="true" />
              Redactar con asistente
            </CardTitle>
            <CardDescription>
              Describa la oferta en una o dos frases y el asistente preparará un borrador con el
              tono comercial de CNI. Después puede editarlo libremente.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="space-y-1.5">
              <Label htmlFor="instruccion">Qué quiere comunicar</Label>
              <Textarea
                id="instruccion"
                rows={3}
                value={instruccion}
                onChange={(evento) => setInstruccion(evento.target.value)}
              />
            </div>

            <div className="flex flex-wrap gap-2">
              <Button
                type="button"
                onClick={() => redaccion.mutate(false)}
                disabled={redaccion.isPending || instruccion.trim().length < 10}
                className="text-white"
              >
                {redaccion.isPending ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
                ) : (
                  <Sparkles className="mr-2 h-4 w-4" aria-hidden="true" />
                )}
                Generar borrador
              </Button>

              {cuerpo.trim().length > 10 && (
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => redaccion.mutate(true)}
                  disabled={redaccion.isPending}
                >
                  Mejorar lo que ya escribí
                </Button>
              )}
            </div>

            <p className="text-xs text-muted-foreground">
              El asistente no inventa cifras ni plazos: si falta un dato, deja una variable para que
              usted la complete. Revise siempre el resultado antes de enviar.
            </p>
          </CardContent>
        </Card>
      )}

      <Card className="border-dashed">
        <CardHeader>
          <CardTitle className="text-base">Partir de una carta en Word</CardTitle>
          <CardDescription>
            Suba el documento .docx que ya tiene redactado. Reconoce tanto los campos de
            combinación de Word, como «NOMBRES», como las variables entre llaves, como {"{nombre}"}.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <input
            ref={referenciaWord}
            type="file"
            accept=".docx"
            className="hidden"
            onChange={(evento) => {
              const archivo = evento.target.files?.[0];
              if (archivo) {
                lectura.mutate(archivo);
              }
              evento.target.value = "";
            }}
          />
          <Button
            type="button"
            variant="outline"
            onClick={() => referenciaWord.current?.click()}
            disabled={lectura.isPending}
          >
            {lectura.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
            ) : (
              <FileUp className="mr-2 h-4 w-4" aria-hidden="true" />
            )}
            Subir documento de Word
          </Button>

          {variablesDetectadas.length > 0 && (
            <div className="mt-4 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3">
              <p className="text-sm font-medium text-emerald-900">
                Variables encontradas en el documento
              </p>
              <div className="mt-2 flex flex-wrap gap-1.5">
                {variablesDetectadas.map((variable) => (
                  <code
                    key={variable}
                    className="rounded bg-white px-1.5 py-0.5 font-mono text-xs text-emerald-800"
                  >
                    {`{${variable}}`}
                  </code>
                ))}
              </div>
              <p className="mt-2 text-xs text-emerald-800">
                Cada una se reemplazará con los datos de la empresa destinataria.
              </p>
            </div>
          )}

          {documentoSinVariables && (
            <div className="mt-4 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3">
              <p className="text-sm text-amber-900">
                El documento no tiene campos de combinación ni variables, así que todas las
                empresas recibirán el mismo texto. Escriba «NOMBRES» o {"{empresa}"} donde deba
                personalizarse.
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Contenido</CardTitle>
          <CardDescription>
            Se reemplazan los campos de Word como «NOMBRES» y las variables entre llaves como{" "}
            {"{empresa}"}.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={(evento) => {
              evento.preventDefault();
              creacion.mutate();
            }}
            className="space-y-4"
          >
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
              {listas.data?.datos.length === 0 && (
                <p className="text-xs text-muted-foreground">
                  Todavía no hay listas.{" "}
                  <Link href="/listas" className="font-medium underline">
                    Cree una lista
                  </Link>{" "}
                  para poder enviar.
                </p>
              )}
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
              <div className="flex flex-wrap items-end justify-between gap-2">
                <Label htmlFor="cuerpo">Cuerpo del mensaje</Label>
                {asistente.data?.disponible && (
                  <div className="flex gap-2">
                    {cuerpoAnterior !== null && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          setCuerpo(cuerpoAnterior);
                          setCuerpoAnterior(null);
                          setResumenSugerencia("");
                        }}
                      >
                        Deshacer
                      </Button>
                    )}
                    <Button
                      type="button"
                      size="sm"
                      onClick={() => sugerencia.mutate()}
                      disabled={sugerencia.isPending || fuenteAnalisis.length < 10}
                      className="text-white"
                    >
                      {sugerencia.isPending ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
                      ) : (
                        <Sparkles className="mr-2 h-4 w-4" aria-hidden="true" />
                      )}
                      {sugerencia.isPending ? "Analizando documento" : "Sugerir con IA"}
                    </Button>
                  </div>
                )}
              </div>

              {resumenSugerencia && (
                <p className="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs text-emerald-900">
                  <span className="font-medium">Lo que entendió del documento:</span>{" "}
                  {resumenSugerencia}{" "}
                  {archivoWord
                    ? "El Word se adjuntará personalizado para cada empresa al guardar."
                    : "Recuerde subir el Word como carta personalizada en el siguiente paso."}
                </p>
              )}

              {asistente.data?.disponible && fuenteAnalisis.length < 10 && (
                <p className="text-xs text-muted-foreground">
                  Suba un documento de Word o escriba la carta para que la IA sugiera un cuerpo
                  profesional.
                </p>
              )}

              <Textarea
                id="cuerpo"
                rows={14}
                value={cuerpo}
                onChange={(evento) => setCuerpo(evento.target.value)}
                className="font-mono"
              />
              <div className="flex flex-wrap items-center gap-2 pt-1">
                <span className="text-xs text-muted-foreground">Insertar variable:</span>
                {variablesDisponibles.map((variable) => (
                  <Button
                    key={variable}
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => insertarVariable(variable)}
                    className="h-7 font-mono text-xs"
                  >
                    {`{${variable}}`}
                  </Button>
                ))}
              </div>

              {asistente.data?.disponible && (
                <div className="mt-3 rounded-lg border border-dashed border-emerald-300 bg-emerald-50/40 p-4">
                  <p className="text-sm font-medium">Analizar una carta en Word con IA</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Suba el .docx: la IA lo lee y redacta arriba un cuerpo profesional para el correo.
                    El Word se adjuntará personalizado para cada empresa al guardar.
                  </p>
                  <input
                    ref={referenciaWordIA}
                    type="file"
                    accept=".docx"
                    className="hidden"
                    onChange={(evento) => {
                      const archivo = evento.target.files?.[0];
                      if (archivo) {
                        lecturaConIA.mutate(archivo);
                      }
                      evento.target.value = "";
                    }}
                  />
                  <div className="mt-3 flex flex-wrap items-center gap-3">
                    <Button
                      type="button"
                      size="sm"
                      onClick={() => referenciaWordIA.current?.click()}
                      disabled={lecturaConIA.isPending}
                      className="text-white"
                    >
                      {lecturaConIA.isPending ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
                      ) : (
                        <FileUp className="mr-2 h-4 w-4" aria-hidden="true" />
                      )}
                      {lecturaConIA.isPending ? "Analizando documento" : "Subir Word y sugerir"}
                    </Button>
                    {archivoWord && (
                      <span className="flex items-center gap-2 text-xs text-emerald-800">
                        {archivoWord.name}
                        <button
                          type="button"
                          onClick={() => setArchivoWord(null)}
                          className="text-muted-foreground underline hover:text-destructive"
                        >
                          no adjuntar
                        </button>
                      </span>
                    )}
                  </div>
                </div>
              )}
            </div>

            <div className="flex gap-2 pt-2">
              <Button
                type="submit"
                disabled={
                  creacion.isPending || asunto.trim().length < 3 || cuerpo.trim().length < 10
                }
                className="text-white"
              >
                {creacion.isPending && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />
                )}
                Guardar borrador
              </Button>
              <Button type="button" variant="outline" onClick={() => router.back()}>
                Cancelar
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </section>
  );
}

"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { ArrowLeft, FileUp, Loader2 } from "lucide-react";
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
  const referenciaWord = useRef<HTMLInputElement>(null);

  const listas = useQuery({ queryKey: ["listas"], queryFn: () => api.listarListas() });

  const lectura = useMutation({
    mutationFn: (archivo: File) => api.leerDocumento(archivo),
    onSuccess: (contenido) => {
      setCuerpo(contenido.cuerpo);
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

  const creacion = useMutation({
    mutationFn: () =>
      api.crearCorrespondencia({
        asunto,
        cuerpo,
        lista_id: listaId === SIN_LISTA ? null : listaId,
      }),
    onSuccess: (pieza) => {
      toast.success("Borrador guardado");
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

      <Card className="border-dashed">
        <CardHeader>
          <CardTitle className="text-base">Partir de una carta en Word</CardTitle>
          <CardDescription>
            Suba el documento .docx que ya tiene redactado. El sistema leerá su texto y lo cargará
            abajo para que revise las variables antes de enviar.
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
                El documento no tiene variables entre llaves, así que todas las empresas recibirán
                el mismo texto. Escriba {"{nombre}"} o {"{empresa}"} donde deba personalizarse.
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Contenido</CardTitle>
          <CardDescription>
            Las variables entre llaves se reemplazan con los datos de cada destinatario.
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
              <Label htmlFor="cuerpo">Cuerpo del mensaje</Label>
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

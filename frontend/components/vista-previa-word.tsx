"use client";

import { useQuery } from "@tanstack/react-query";
import { Download, Eye, EyeOff, Loader2 } from "lucide-react";
import { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";

const PRIMERO = "primero";

interface Props {
  correspondenciaId: string;
  listaId: string | null;
  nombrePlantilla: string;
  versionPlantilla: string;
}

export function VistaPreviaWord({
  correspondenciaId,
  listaId,
  nombrePlantilla,
  versionPlantilla,
}: Props) {
  const [visible, setVisible] = useState(false);
  const [contactoId, setContactoId] = useState(PRIMERO);
  const [errorRender, setErrorRender] = useState<string | null>(null);
  const contenedor = useRef<HTMLDivElement>(null);

  const miembros = useQuery({
    queryKey: ["miembros", listaId],
    queryFn: () => api.miembrosDeLista(listaId as string),
    enabled: visible && listaId !== null,
  });

  const contactoElegido = contactoId === PRIMERO ? undefined : contactoId;

  const documentoPdf = useQuery({
    queryKey: ["vista-previa-pdf", correspondenciaId, contactoId, versionPlantilla],
    queryFn: () => api.vistaPreviaPlantilla(correspondenciaId, contactoElegido, "pdf"),
    enabled: visible,
    retry: false,
  });

  const usarVisorWord = documentoPdf.isError;

  const documento = useQuery({
    queryKey: ["vista-previa-word", correspondenciaId, contactoId, versionPlantilla],
    queryFn: () => api.vistaPreviaPlantilla(correspondenciaId, contactoElegido),
    enabled: visible && usarVisorWord,
    retry: false,
  });

  const [urlPdf, setUrlPdf] = useState<string | null>(null);

  useEffect(() => {
    if (!documentoPdf.data) {
      setUrlPdf(null);
      return;
    }
    const url = URL.createObjectURL(documentoPdf.data);
    setUrlPdf(url);
    return () => URL.revokeObjectURL(url);
  }, [documentoPdf.data]);

  useEffect(() => {
    const destino = contenedor.current;
    if (!visible || !usarVisorWord || !documento.data || !destino) {
      return;
    }

    let cancelado = false;
    setErrorRender(null);

    import("docx-preview")
      .then(({ renderAsync }) => {
        if (cancelado) {
          return;
        }
        destino.innerHTML = "";
        return renderAsync(documento.data, destino, undefined, {
          className: "vista-word",
          inWrapper: true,
          ignoreWidth: false,
          ignoreHeight: true,
          breakPages: true,
          ignoreLastRenderedPageBreak: false,
          experimental: true,
          renderHeaders: true,
          renderFooters: true,
        });
      })
      .catch(() => {
        if (!cancelado) {
          setErrorRender("No se pudo dibujar el documento. Descárguelo para revisarlo en Word.");
        }
      });

    return () => {
      cancelado = true;
    };
  }, [documento.data, visible, usarVisorWord]);

  async function descargar() {
    const archivo =
      documento.data ?? (await api.vistaPreviaPlantilla(correspondenciaId, contactoElegido));
    const enlace = document.createElement("a");
    enlace.href = URL.createObjectURL(archivo);
    enlace.download = nombrePlantilla;
    document.body.appendChild(enlace);
    enlace.click();
    document.body.removeChild(enlace);
    URL.revokeObjectURL(enlace.href);
  }

  const empresas = miembros.data?.datos ?? [];

  return (
    <div className="mt-4 space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        <Button type="button" variant="outline" size="sm" onClick={() => setVisible((v) => !v)}>
          {visible ? (
            <EyeOff className="mr-2 h-4 w-4" aria-hidden="true" />
          ) : (
            <Eye className="mr-2 h-4 w-4" aria-hidden="true" />
          )}
          {visible ? "Ocultar vista previa" : "Ver cómo quedará la carta"}
        </Button>
        {visible && (urlPdf || documento.data) && (
          <Button type="button" variant="ghost" size="sm" onClick={descargar}>
            <Download className="mr-2 h-4 w-4" aria-hidden="true" />
            Descargar esta versión
          </Button>
        )}
      </div>

      {visible && (
        <div className="space-y-3">
          {listaId !== null && empresas.length > 0 && (
            <div className="max-w-sm space-y-1.5">
              <Label htmlFor="empresa-vista-previa" className="text-xs text-muted-foreground">
                Ver personalizada para
              </Label>
              <Select value={contactoId} onValueChange={setContactoId}>
                <SelectTrigger id="empresa-vista-previa">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={PRIMERO}>Primera empresa de la lista</SelectItem>
                  {empresas.map((contacto) => (
                    <SelectItem key={contacto.id} value={contacto.id}>
                      {contacto.empresa} · {contacto.correo}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          {listaId === null && (
            <p className="text-xs text-amber-700">
              La carta no tiene lista asignada: se muestra con datos de ejemplo.
            </p>
          )}

          {(documentoPdf.isFetching || documento.isFetching) && (
            <p className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
              Generando la carta personalizada
            </p>
          )}

          {documento.isError && (
            <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {documento.error instanceof ErrorPeticion
                ? documento.error.message
                : "No se pudo generar la vista previa."}
            </p>
          )}

          {errorRender && (
            <p className="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
              {errorRender}
            </p>
          )}

          {urlPdf && !documentoPdf.isFetching && (
            <iframe
              src={urlPdf}
              title={`Vista previa de ${nombrePlantilla}`}
              className="h-[80vh] w-full rounded-lg border border-zinc-200 bg-zinc-100"
            />
          )}

          {usarVisorWord && (
            <div className="max-h-[80vh] overflow-auto rounded-lg border border-zinc-200 bg-zinc-100">
              <div ref={contenedor} />
            </div>
          )}

          <p className="text-xs text-muted-foreground">
            {usarVisorWord
              ? "Vista aproximada generada en el navegador; algunos detalles de diseño pueden verse distintos. La descarga es el archivo exacto que se enviará."
              : "Vista con la maquetación real de la carta, página por página. La descarga es el archivo Word exacto que se enviará."}
          </p>
        </div>
      )}
    </div>
  );
}

"use client";

import { useMutation } from "@tanstack/react-query";
import { Loader2, Plus, X } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
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
import { Separator } from "@/components/ui/separator";
import { api } from "@/lib/api";
import { ErrorPeticion } from "@/lib/api-client";
import type { Contacto } from "@/lib/tipos";

const camposFijos = [
  { clave: "ruc", etiqueta: "RUC" },
  { clave: "ciudad_region", etiqueta: "Ciudad / región" },
  { clave: "telefono_fijo", etiqueta: "Teléfono fijo" },
  { clave: "telefono_movil", etiqueta: "Teléfono móvil" },
  { clave: "pagina_web", etiqueta: "Página web" },
  { clave: "facebook", etiqueta: "Facebook" },
] as const;

const clavesFijas = new Set<string>(camposFijos.map((campo) => campo.clave));

interface CampoPropio {
  id: number;
  nombre: string;
  valor: string;
}

export function normalizarClave(texto: string): string {
  return texto
    .toLowerCase()
    .normalize("NFD")
    .replace(/[̀-ͯ]/g, "")
    .replace(/[\s/-]+/g, "_")
    .replace(/[^a-z0-9_]/g, "")
    .replace(/^_+|_+$/g, "");
}

function textoDe(valor: unknown): string {
  if (typeof valor === "string") {
    return valor;
  }
  if (typeof valor === "number" || typeof valor === "boolean") {
    return String(valor);
  }
  return "";
}

interface Props {
  abierto: boolean;
  contacto: Contacto | null;
  alCambiar: (abierto: boolean) => void;
  alGuardar: () => void;
}

export function FormularioContacto({ abierto, contacto, alCambiar, alGuardar }: Props) {
  const [empresa, setEmpresa] = useState("");
  const [correo, setCorreo] = useState("");
  const [nombre, setNombre] = useState("");
  const [pais, setPais] = useState("");
  const [fijos, setFijos] = useState<Record<string, string>>({});
  const [propios, setPropios] = useState<CampoPropio[]>([]);
  const [siguienteId, setSiguienteId] = useState(1);

  useEffect(() => {
    if (!abierto) {
      return;
    }

    const extras = contacto?.campos_extra ?? {};
    const valoresFijos: Record<string, string> = {};
    const valoresPropios: CampoPropio[] = [];
    let contador = 1;

    for (const [clave, valor] of Object.entries(extras)) {
      if (clavesFijas.has(clave)) {
        valoresFijos[clave] = textoDe(valor);
      } else {
        valoresPropios.push({ id: contador, nombre: clave, valor: textoDe(valor) });
        contador += 1;
      }
    }

    setEmpresa(contacto?.empresa ?? "");
    setCorreo(contacto?.correo ?? "");
    setNombre(contacto && contacto.nombre !== contacto.empresa ? contacto.nombre : "");
    setPais(contacto?.pais ?? "");
    setFijos(valoresFijos);
    setPropios(valoresPropios);
    setSiguienteId(contador);
  }, [abierto, contacto]);

  const guardado = useMutation({
    mutationFn: () => {
      const camposExtra: Record<string, string> = {};

      for (const campo of camposFijos) {
        const valor = (fijos[campo.clave] ?? "").trim();
        if (valor) {
          camposExtra[campo.clave] = valor;
        }
      }

      for (const campo of propios) {
        const clave = normalizarClave(campo.nombre);
        const valor = campo.valor.trim();
        if (clave && valor) {
          camposExtra[clave] = valor;
        }
      }

      const datos = {
        empresa: empresa.trim(),
        correo: correo.trim(),
        nombre: nombre.trim(),
        pais: pais.trim() ? pais.trim() : null,
        campos_extra: camposExtra,
      };

      return contacto ? api.actualizarContacto(contacto.id, datos) : api.crearContacto(datos);
    },
    onSuccess: (guardadoContacto) => {
      toast.success(
        contacto
          ? `${guardadoContacto.empresa} actualizado`
          : `${guardadoContacto.empresa} agregado a la base de contactos`,
      );
      alCambiar(false);
      alGuardar();
    },
    onError: (fallo: unknown) => {
      toast.error(fallo instanceof ErrorPeticion ? fallo.message : "No se pudo guardar el contacto");
    },
  });

  function agregarCampo() {
    setPropios((actuales) => [...actuales, { id: siguienteId, nombre: "", valor: "" }]);
    setSiguienteId((actual) => actual + 1);
  }

  function actualizarPropio(id: number, cambios: Partial<Omit<CampoPropio, "id">>) {
    setPropios((actuales) =>
      actuales.map((campo) => (campo.id === id ? { ...campo, ...cambios } : campo)),
    );
  }

  const clavesPropias = propios.map((campo) => normalizarClave(campo.nombre)).filter(Boolean);
  const hayRepetidas =
    new Set(clavesPropias).size !== clavesPropias.length ||
    clavesPropias.some((clave) => clavesFijas.has(clave));

  const valido =
    empresa.trim().length >= 2 && correo.includes("@") && !hayRepetidas && !guardado.isPending;

  return (
    <Dialog open={abierto} onOpenChange={alCambiar}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{contacto ? "Editar contacto" : "Nuevo contacto"}</DialogTitle>
          <DialogDescription>
            Empresa y correo son obligatorios. Los demás datos quedan disponibles como variables en
            las cartas.
          </DialogDescription>
        </DialogHeader>

        <form
          id="formulario-contacto"
          onSubmit={(evento) => {
            evento.preventDefault();
            if (valido) {
              guardado.mutate();
            }
          }}
          className="space-y-5"
        >
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1.5 sm:col-span-2">
              <Label htmlFor="empresa">Empresa</Label>
              <Input id="empresa" value={empresa} onChange={(evento) => setEmpresa(evento.target.value)} />
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
              <Label htmlFor="nombre">Nombre del contacto (opcional)</Label>
              <Input id="nombre" value={nombre} onChange={(evento) => setNombre(evento.target.value)} />
            </div>
            {camposFijos.map((campo) => (
              <div key={campo.clave} className="space-y-1.5">
                <Label htmlFor={campo.clave}>{campo.etiqueta}</Label>
                <Input
                  id={campo.clave}
                  value={fijos[campo.clave] ?? ""}
                  onChange={(evento) =>
                    setFijos((actuales) => ({ ...actuales, [campo.clave]: evento.target.value }))
                  }
                />
              </div>
            ))}
            <div className="space-y-1.5">
              <Label htmlFor="pais">País</Label>
              <Input id="pais" value={pais} onChange={(evento) => setPais(evento.target.value)} />
            </div>
          </div>

          <Separator />

          <div className="space-y-3">
            <div className="flex items-center justify-between gap-3">
              <div>
                <p className="text-sm font-medium">Campos adicionales</p>
                <p className="text-xs text-muted-foreground">
                  Por ejemplo cargo o rubro. Se usan en las cartas como {"{cargo}"}.
                </p>
              </div>
              <Button type="button" variant="outline" size="sm" onClick={agregarCampo}>
                <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
                Agregar campo
              </Button>
            </div>

            {propios.map((campo) => {
              const clave = normalizarClave(campo.nombre);
              return (
                <div key={campo.id} className="grid grid-cols-[1fr_1fr_auto] items-end gap-2">
                  <div className="space-y-1.5">
                    <Label htmlFor={`campo-nombre-${campo.id}`} className="text-xs">
                      Nombre del campo
                    </Label>
                    <Input
                      id={`campo-nombre-${campo.id}`}
                      value={campo.nombre}
                      onChange={(evento) => actualizarPropio(campo.id, { nombre: evento.target.value })}
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor={`campo-valor-${campo.id}`} className="text-xs">
                      Valor{clave ? ` · variable {${clave}}` : ""}
                    </Label>
                    <Input
                      id={`campo-valor-${campo.id}`}
                      value={campo.valor}
                      onChange={(evento) => actualizarPropio(campo.id, { valor: evento.target.value })}
                    />
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() =>
                      setPropios((actuales) => actuales.filter((otro) => otro.id !== campo.id))
                    }
                    aria-label="Quitar este campo"
                    className="h-9 w-9 text-muted-foreground hover:text-destructive"
                  >
                    <X className="h-4 w-4" aria-hidden="true" />
                  </Button>
                </div>
              );
            })}

            {hayRepetidas && (
              <p className="text-xs text-destructive">
                Hay campos con el mismo nombre o que repiten uno de los campos de arriba.
              </p>
            )}
          </div>
        </form>

        <DialogFooter className="gap-2 sm:gap-2">
          <Button type="button" variant="outline" onClick={() => alCambiar(false)}>
            Cancelar
          </Button>
          <Button type="submit" form="formulario-contacto" disabled={!valido} className="text-white">
            {guardado.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />}
            {contacto ? "Guardar cambios" : "Guardar contacto"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

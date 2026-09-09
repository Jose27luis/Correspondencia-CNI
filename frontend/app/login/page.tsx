"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { ErrorPeticion } from "@/lib/api-client";
import { useSesion } from "@/lib/sesion";

const esquema = z.object({
  correo: z.string().min(1, "Ingrese su correo").email("El correo no es válido"),
  contrasena: z.string().min(1, "Ingrese su contraseña"),
});

type Formulario = z.infer<typeof esquema>;

export default function PaginaLogin() {
  const { acceder, usuario, cargando } = useSesion();
  const router = useRouter();
  const [errorGeneral, setErrorGeneral] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<Formulario>({ resolver: zodResolver(esquema) });

  useEffect(() => {
    if (!cargando && usuario) {
      router.replace("/correspondencia");
    }
  }, [cargando, usuario, router]);

  async function enviar(datos: Formulario) {
    setErrorGeneral(null);
    try {
      await acceder(datos.correo, datos.contrasena);
    } catch (error) {
      if (error instanceof ErrorPeticion) {
        setErrorGeneral(error.message);
        return;
      }
      setErrorGeneral("No se pudo conectar con el servidor");
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center px-4">
      <div className="w-full max-w-sm rounded-lg border border-slate-200 bg-white p-8 shadow-sm">
        <h1 className="text-xl font-semibold">Correspondencia CNI</h1>
        <p className="mt-1 text-sm text-slate-500">Acceso al panel interno</p>

        <form onSubmit={handleSubmit(enviar)} className="mt-6 space-y-4" noValidate>
          <div>
            <label htmlFor="correo" className="block text-sm font-medium">
              Correo
            </label>
            <input
              id="correo"
              type="email"
              autoComplete="email"
              className="mt-1 w-full rounded border border-slate-300 px-3 py-2 text-sm outline-none focus:border-slate-900"
              {...register("correo")}
            />
            {errors.correo && (
              <p className="mt-1 text-xs text-red-600">{errors.correo.message}</p>
            )}
          </div>

          <div>
            <label htmlFor="contrasena" className="block text-sm font-medium">
              Contraseña
            </label>
            <input
              id="contrasena"
              type="password"
              autoComplete="current-password"
              className="mt-1 w-full rounded border border-slate-300 px-3 py-2 text-sm outline-none focus:border-slate-900"
              {...register("contrasena")}
            />
            {errors.contrasena && (
              <p className="mt-1 text-xs text-red-600">{errors.contrasena.message}</p>
            )}
          </div>

          {errorGeneral && (
            <p className="rounded bg-red-50 px-3 py-2 text-xs text-red-700">{errorGeneral}</p>
          )}

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full rounded bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-60"
          >
            {isSubmitting ? "Ingresando" : "Ingresar"}
          </button>
        </form>
      </div>
    </main>
  );
}

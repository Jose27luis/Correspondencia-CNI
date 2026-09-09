"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { ErrorPeticion } from "@/lib/api-client";
import { useSesion } from "@/lib/sesion";

const esquema = z.object({
  correo: z.string().min(1, "Ingrese su correo").email("El correo no tiene un formato válido"),
  contrasena: z.string().min(1, "Ingrese su contraseña"),
});

type Formulario = z.infer<typeof esquema>;

function IconoOjoAbierto() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true" className="h-5 w-5">
      <path d="M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12Z" strokeLinecap="round" strokeLinejoin="round" />
      <circle cx="12" cy="12" r="3.2" />
    </svg>
  );
}

function IconoOjoCerrado() {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true" className="h-5 w-5">
      <path d="M3 3l18 18" strokeLinecap="round" />
      <path d="M10.6 6.2A9.6 9.6 0 0 1 12 6c6 0 9.5 6 9.5 6a17 17 0 0 1-3.3 3.9M6.4 8.1A17 17 0 0 0 2.5 12S6 18 12 18a9.3 9.3 0 0 0 3.5-.7" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M9.9 9.9a3.2 3.2 0 0 0 4.2 4.2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export default function PaginaLogin() {
  const { acceder, usuario, cargando } = useSesion();
  const router = useRouter();
  const [errorGeneral, setErrorGeneral] = useState<string | null>(null);
  const [verContrasena, setVerContrasena] = useState(false);
  const [mayusculasActivas, setMayusculasActivas] = useState(false);

  const {
    register,
    handleSubmit,
    setFocus,
    formState: { errors, isSubmitting },
  } = useForm<Formulario>({ resolver: zodResolver(esquema), mode: "onBlur" });

  useEffect(() => {
    if (!cargando && usuario) {
      router.replace("/correspondencia");
    }
  }, [cargando, usuario, router]);

  useEffect(() => {
    if (!cargando && !usuario) {
      setFocus("correo");
    }
  }, [cargando, usuario, setFocus]);

  function detectarMayusculas(evento: React.KeyboardEvent<HTMLInputElement>) {
    setMayusculasActivas(evento.getModifierState("CapsLock"));
  }

  async function enviar(datos: Formulario) {
    setErrorGeneral(null);
    try {
      await acceder(datos.correo, datos.contrasena);
    } catch (error) {
      if (error instanceof ErrorPeticion) {
        setErrorGeneral(error.message);
        return;
      }
      setErrorGeneral("No se pudo conectar con el servidor. Intente nuevamente.");
    }
  }

  return (
    <main className="grid min-h-screen lg:grid-cols-[1.05fr_1fr]">
      <section className="relative hidden overflow-hidden bg-slate-900 px-12 py-14 text-slate-100 lg:flex lg:flex-col lg:justify-between">
        <div
          aria-hidden="true"
          className="pointer-events-none absolute -right-24 -top-24 h-[26rem] w-[26rem] rounded-full bg-emerald-500/15 blur-3xl"
        />
        <div
          aria-hidden="true"
          className="pointer-events-none absolute -bottom-32 -left-20 h-[22rem] w-[22rem] rounded-full bg-sky-500/10 blur-3xl"
        />

        <div className="relative">
          <span className="inline-flex items-center rounded-full border border-white/15 bg-white/5 px-3 py-1 text-xs font-medium tracking-wide text-slate-200">
            Panel interno
          </span>
          <h1 className="mt-8 text-3xl font-semibold leading-tight">
            Correspondencia
            <span className="block text-emerald-400">CNI</span>
          </h1>
          <p className="mt-4 max-w-md text-sm leading-relaxed text-slate-300">
            Componga una carta comercial una sola vez, personalícela con los datos de cada empresa y
            envíela a toda una lista con seguimiento de entrega.
          </p>
        </div>

        <dl className="relative grid gap-6 sm:grid-cols-3">
          <div>
            <dt className="text-xs uppercase tracking-wide text-slate-400">Contactos</dt>
            <dd className="mt-1 text-sm text-slate-200">Una sola base de empresas</dd>
          </div>
          <div>
            <dt className="text-xs uppercase tracking-wide text-slate-400">Personalización</dt>
            <dd className="mt-1 text-sm text-slate-200">Variables por destinatario</dd>
          </div>
          <div>
            <dt className="text-xs uppercase tracking-wide text-slate-400">Seguimiento</dt>
            <dd className="mt-1 text-sm text-slate-200">Entregas y rebotes</dd>
          </div>
        </dl>
      </section>

      <section className="flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-sm">
          <div className="lg:hidden">
            <h1 className="text-2xl font-semibold">
              Correspondencia <span className="text-emerald-600">CNI</span>
            </h1>
          </div>

          <div className="mt-6 lg:mt-0">
            <h2 className="text-xl font-semibold text-slate-900">Iniciar sesión</h2>
            <p className="mt-1 text-sm text-slate-500">
              Use las credenciales que le asignó el equipo de CNI.
            </p>
          </div>

          <form onSubmit={handleSubmit(enviar)} className="mt-8 space-y-5" noValidate>
            <div>
              <label htmlFor="correo" className="block text-sm font-medium text-slate-800">
                Correo electrónico
              </label>
              <input
                id="correo"
                type="email"
                autoComplete="username"
                spellCheck={false}
                aria-invalid={errors.correo ? true : undefined}
                aria-describedby={errors.correo ? "error-correo" : undefined}
                className={`mt-1.5 w-full rounded-lg border px-3.5 py-2.5 text-sm text-slate-900 outline-none transition focus:ring-4 ${
                  errors.correo
                    ? "border-red-400 focus:border-red-500 focus:ring-red-100"
                    : "border-slate-300 focus:border-slate-900 focus:ring-slate-900/10"
                }`}
                {...register("correo")}
              />
              {errors.correo && (
                <p id="error-correo" className="mt-1.5 text-xs text-red-600">
                  {errors.correo.message}
                </p>
              )}
            </div>

            <div>
              <label htmlFor="contrasena" className="block text-sm font-medium text-slate-800">
                Contraseña
              </label>
              <div className="relative mt-1.5">
                <input
                  id="contrasena"
                  type={verContrasena ? "text" : "password"}
                  autoComplete="current-password"
                  onKeyUp={detectarMayusculas}
                  aria-invalid={errors.contrasena ? true : undefined}
                  aria-describedby={errors.contrasena ? "error-contrasena" : undefined}
                  className={`w-full rounded-lg border py-2.5 pl-3.5 pr-12 text-sm text-slate-900 outline-none transition focus:ring-4 ${
                    errors.contrasena
                      ? "border-red-400 focus:border-red-500 focus:ring-red-100"
                      : "border-slate-300 focus:border-slate-900 focus:ring-slate-900/10"
                  }`}
                  {...register("contrasena")}
                />
                <button
                  type="button"
                  onClick={() => setVerContrasena((estado) => !estado)}
                  aria-pressed={verContrasena}
                  aria-controls="contrasena"
                  aria-label={verContrasena ? "Ocultar la contraseña" : "Mostrar la contraseña"}
                  title={verContrasena ? "Ocultar la contraseña" : "Mostrar la contraseña"}
                  className="absolute inset-y-0 right-0 flex items-center rounded-r-lg px-3 text-slate-500 transition hover:text-slate-900 focus:outline-none focus-visible:ring-4 focus-visible:ring-slate-900/10"
                >
                  {verContrasena ? <IconoOjoCerrado /> : <IconoOjoAbierto />}
                </button>
              </div>

              {errors.contrasena && (
                <p id="error-contrasena" className="mt-1.5 text-xs text-red-600">
                  {errors.contrasena.message}
                </p>
              )}

              {mayusculasActivas && (
                <p className="mt-1.5 text-xs text-amber-700">
                  El bloqueo de mayúsculas está activado.
                </p>
              )}
            </div>

            {errorGeneral && (
              <p
                role="alert"
                className="rounded-lg border border-red-200 bg-red-50 px-3.5 py-2.5 text-sm text-red-700"
              >
                {errorGeneral}
              </p>
            )}

            <button
              type="submit"
              disabled={isSubmitting}
              className="flex w-full items-center justify-center gap-2 rounded-lg bg-slate-900 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-slate-800 focus:outline-none focus-visible:ring-4 focus-visible:ring-slate-900/20 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {isSubmitting && (
                <svg viewBox="0 0 24 24" aria-hidden="true" className="h-4 w-4 animate-spin">
                  <circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" strokeWidth="3" opacity="0.25" />
                  <path d="M21 12a9 9 0 0 0-9-9" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
                </svg>
              )}
              {isSubmitting ? "Verificando" : "Ingresar"}
            </button>
          </form>

          <p className="mt-8 text-xs leading-relaxed text-slate-500">
            Si olvidó su contraseña, solicite el restablecimiento al administrador del panel.
          </p>
        </div>
      </section>
    </main>
  );
}

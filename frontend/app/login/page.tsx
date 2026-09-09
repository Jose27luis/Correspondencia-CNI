"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import { Eye, EyeOff, Loader2, TriangleAlert } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ErrorPeticion } from "@/lib/api-client";
import { useSesion } from "@/lib/sesion";
import { cn } from "@/lib/utils";

const esquema = z.object({
  correo: z
    .string()
    .min(1, "Ingrese su correo")
    .email("El correo no tiene un formato válido"),
  contrasena: z.string().min(1, "Ingrese su contraseña"),
});

type Formulario = z.infer<typeof esquema>;

const ventajas = [
  { titulo: "Contactos", detalle: "Una sola base de empresas" },
  { titulo: "Personalización", detalle: "Variables por destinatario" },
  { titulo: "Seguimiento", detalle: "Entregas y rebotes" },
];

export default function PaginaLogin() {
  const { acceder, usuario, cargando } = useSesion();
  const router = useRouter();
  const animacionReducida = useReducedMotion();
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
      setErrorGeneral(
        "No se pudo conectar con el servidor. Intente nuevamente.",
      );
    }
  }

  const desplazamiento = animacionReducida ? 0 : 12;

  const contenedor = {
    oculto: {},
    visible: { transition: { staggerChildren: 0.07, delayChildren: 0.1 } },
  };

  const elemento = {
    oculto: { opacity: 0, y: desplazamiento },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: 0.45, ease: [0.22, 1, 0.36, 1] as const },
    },
  };

  return (
    <main className="grid min-h-screen lg:grid-cols-[1.05fr_1fr]">
      <section className="relative hidden overflow-hidden bg-gradient-to-br from-[#0b1220] via-[#0f1c2e] to-[#04140f] px-12 py-14 text-zinc-100 lg:flex lg:flex-col lg:justify-between">
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 opacity-[0.35] [background-image:radial-gradient(circle_at_1px_1px,rgba(255,255,255,0.14)_1px,transparent_0)] [background-size:22px_22px]"
        />
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-y-0 right-0 w-px bg-gradient-to-b from-transparent via-white/15 to-transparent"
        />
        <motion.div
          aria-hidden="true"
          className="pointer-events-none absolute -right-24 -top-24 h-[26rem] w-[26rem] rounded-full bg-emerald-500/15 blur-3xl"
          animate={
            animacionReducida
              ? undefined
              : { scale: [1, 1.12, 1], opacity: [0.5, 0.8, 0.5] }
          }
          transition={{ duration: 12, repeat: Infinity, ease: "easeInOut" }}
        />
        <motion.div
          aria-hidden="true"
          className="pointer-events-none absolute -bottom-32 -left-20 h-[22rem] w-[22rem] rounded-full bg-sky-500/10 blur-3xl"
          animate={
            animacionReducida
              ? undefined
              : { scale: [1, 1.18, 1], opacity: [0.4, 0.7, 0.4] }
          }
          transition={{
            duration: 15,
            repeat: Infinity,
            ease: "easeInOut",
            delay: 1.5,
          }}
        />

        <motion.div
          className="relative"
          variants={contenedor}
          initial="oculto"
          animate="visible"
        >
          <motion.span
            variants={elemento}
            className="inline-flex items-center rounded-full border border-white/15 bg-white/5 px-3 py-1 text-xs font-medium tracking-wide text-zinc-200"
          >
            Panel interno
          </motion.span>
          <motion.h1
            variants={elemento}
            className="mt-8 text-3xl font-semibold leading-tight"
          >
            Correspondencia
            <span className="block text-emerald-400">CNI</span>
          </motion.h1>
          <motion.p
            variants={elemento}
            className="mt-4 max-w-md text-sm leading-relaxed text-zinc-300"
          >
            Componga una carta comercial una sola vez, personalícela con los
            datos de cada empresa y envíela a toda una lista con seguimiento de
            entrega.
          </motion.p>
        </motion.div>

        <motion.dl
          className="relative grid gap-6 sm:grid-cols-3"
          variants={contenedor}
          initial="oculto"
          animate="visible"
        >
          {ventajas.map((ventaja) => (
            <motion.div key={ventaja.titulo} variants={elemento}>
              <dt className="text-xs uppercase tracking-wide text-zinc-400">
                {ventaja.titulo}
              </dt>
              <dd className="mt-1 text-sm text-zinc-200">{ventaja.detalle}</dd>
            </motion.div>
          ))}
        </motion.dl>
      </section>

      <section className="flex items-center justify-center bg-zinc-50 px-6 py-12">
        <motion.div
          className="w-full max-w-sm"
          initial={{ opacity: 0, y: desplazamiento }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
        >
          <div className="lg:hidden">
            <h1 className="text-center text-2xl font-semibold">
              Correspondencia <span className="text-emerald-600">CNI</span>
            </h1>
          </div>

          <div className="mt-6 rounded-2xl border border-zinc-200/80 bg-white p-8 shadow-[0_1px_2px_rgba(16,24,40,0.04),0_12px_32px_-8px_rgba(16,24,40,0.12),0_28px_60px_-20px_rgba(16,24,40,0.18)] lg:mt-0">
            <div>
              <h2 className="text-xl font-semibold tracking-tight">
                Iniciar sesión
              </h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Use las credenciales que le asignó el equipo de CNI.
              </p>
            </div>

            <form
              onSubmit={handleSubmit(enviar)}
              className="mt-7 space-y-5"
              noValidate
            >
              <div className="space-y-1.5">
                <Label htmlFor="correo">Correo electrónico</Label>
                <Input
                  id="correo"
                  type="email"
                  autoComplete="username"
                  spellCheck={false}
                  aria-invalid={errors.correo ? true : undefined}
                  aria-describedby={errors.correo ? "error-correo" : undefined}
                  className={cn(
                    errors.correo &&
                      "border-destructive focus-visible:ring-destructive/30",
                  )}
                  {...register("correo")}
                />
                <AnimatePresence initial={false}>
                  {errors.correo && (
                    <motion.p
                      id="error-correo"
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: "auto" }}
                      exit={{ opacity: 0, height: 0 }}
                      transition={{ duration: 0.2 }}
                      className="text-xs text-destructive"
                    >
                      {errors.correo.message}
                    </motion.p>
                  )}
                </AnimatePresence>
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="contrasena">Contraseña</Label>
                <div className="relative">
                  <Input
                    id="contrasena"
                    type={verContrasena ? "text" : "password"}
                    autoComplete="current-password"
                    onKeyUp={detectarMayusculas}
                    aria-invalid={errors.contrasena ? true : undefined}
                    aria-describedby={
                      errors.contrasena ? "error-contrasena" : undefined
                    }
                    className={cn(
                      "pr-11",
                      errors.contrasena &&
                        "border-destructive focus-visible:ring-destructive/30",
                    )}
                    {...register("contrasena")}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => setVerContrasena((estado) => !estado)}
                    aria-pressed={verContrasena}
                    aria-controls="contrasena"
                    aria-label={
                      verContrasena
                        ? "Ocultar la contraseña"
                        : "Mostrar la contraseña"
                    }
                    title={
                      verContrasena
                        ? "Ocultar la contraseña"
                        : "Mostrar la contraseña"
                    }
                    className="absolute right-1 top-1/2 h-8 w-8 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    <AnimatePresence mode="wait" initial={false}>
                      <motion.span
                        key={verContrasena ? "visible" : "oculta"}
                        initial={{ opacity: 0, scale: 0.7, rotate: -12 }}
                        animate={{ opacity: 1, scale: 1, rotate: 0 }}
                        exit={{ opacity: 0, scale: 0.7, rotate: 12 }}
                        transition={{ duration: 0.16 }}
                        className="flex"
                      >
                        {verContrasena ? (
                          <EyeOff className="h-4 w-4" aria-hidden="true" />
                        ) : (
                          <Eye className="h-4 w-4" aria-hidden="true" />
                        )}
                      </motion.span>
                    </AnimatePresence>
                  </Button>
                </div>

                <AnimatePresence initial={false}>
                  {errors.contrasena && (
                    <motion.p
                      id="error-contrasena"
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: "auto" }}
                      exit={{ opacity: 0, height: 0 }}
                      transition={{ duration: 0.2 }}
                      className="text-xs text-destructive"
                    >
                      {errors.contrasena.message}
                    </motion.p>
                  )}
                </AnimatePresence>

                <AnimatePresence initial={false}>
                  {mayusculasActivas && (
                    <motion.p
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: "auto" }}
                      exit={{ opacity: 0, height: 0 }}
                      transition={{ duration: 0.2 }}
                      className="flex items-center gap-1.5 text-xs text-amber-600"
                    >
                      <TriangleAlert
                        className="h-3.5 w-3.5"
                        aria-hidden="true"
                      />
                      El bloqueo de mayúsculas está activado.
                    </motion.p>
                  )}
                </AnimatePresence>
              </div>

              <AnimatePresence initial={false}>
                {errorGeneral && (
                  <motion.div
                    initial={{ opacity: 0, y: -6, height: 0 }}
                    animate={{ opacity: 1, y: 0, height: "auto" }}
                    exit={{ opacity: 0, y: -6, height: 0 }}
                    transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
                  >
                    <motion.div
                      animate={
                        animacionReducida
                          ? undefined
                          : { x: [0, -6, 6, -4, 4, 0] }
                      }
                      transition={{ duration: 0.4 }}
                    >
                      <Alert variant="destructive" role="alert">
                        <AlertDescription>{errorGeneral}</AlertDescription>
                      </Alert>
                    </motion.div>
                  </motion.div>
                )}
              </AnimatePresence>

              <motion.div
                whileHover={
                  animacionReducida || isSubmitting
                    ? undefined
                    : { scale: 1.015 }
                }
                whileTap={
                  animacionReducida || isSubmitting
                    ? undefined
                    : { scale: 0.985 }
                }
                transition={{ type: "spring", stiffness: 420, damping: 26 }}
              >
                <Button
                  type="submit"
                  disabled={isSubmitting}
                  className="group relative w-full overflow-hidden bg-zinc-900 shadow-[0_1px_2px_rgba(16,24,40,0.12)] transition-all duration-300 hover:bg-zinc-800 hover:shadow-[0_8px_20px_-6px_rgba(16,24,40,0.45)] focus-visible:ring-4 focus-visible:ring-zinc-900/15 disabled:shadow-none"
                >
                  <span
                    aria-hidden="true"
                    className="pointer-events-none absolute inset-0 -translate-x-full bg-gradient-to-r from-transparent via-white/20 to-transparent transition-transform duration-700 group-hover:translate-x-full"
                  />
                  <span className="relative flex items-center justify-center">
                    {isSubmitting && (
                      <Loader2
                        className="mr-2 h-4 w-4 animate-spin"
                        aria-hidden="true"
                      />
                    )}
                    {isSubmitting ? "Verificando" : "Ingresar"}
                  </span>
                </Button>
              </motion.div>
            </form>
          </div>

          <p className="mt-6 text-center text-xs leading-relaxed text-muted-foreground">
            Si olvidó su contraseña, solicite el restablecimiento al
            administrador del panel.
          </p>
        </motion.div>
      </section>
    </main>
  );
}

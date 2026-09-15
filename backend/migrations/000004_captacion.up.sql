CREATE TABLE directorio_empresa (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre text NOT NULL,
    ruc text NOT NULL,
    correo citext NOT NULL,
    direccion text NOT NULL,
    ciudad text NOT NULL,
    telefono text,
    celular text,
    facebook text,
    pagina_web text,
    descripcion text NOT NULL,
    logo_url text,
    estado text NOT NULL DEFAULT 'pendiente'
        CHECK (estado IN ('pendiente', 'aprobado', 'rechazado')),
    motivo_rechazo text,
    contacto_id uuid REFERENCES contacto (id) ON DELETE SET NULL,
    creado_en timestamptz NOT NULL DEFAULT now(),
    revisado_en timestamptz
);

CREATE INDEX idx_directorio_estado ON directorio_empresa (estado);

CREATE TABLE rueda_empresa (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    razon_social text NOT NULL,
    ruc text NOT NULL,
    persona_encargada text NOT NULL,
    cargo_encargado text,
    correo citext NOT NULL UNIQUE,
    direccion text,
    ciudad text,
    region text,
    pais text NOT NULL,
    codigo_postal text,
    telefono text,
    celular text,
    pagina_web text,
    estado text NOT NULL DEFAULT 'pendiente'
        CHECK (estado IN ('pendiente', 'aprobado', 'rechazado')),
    motivo_rechazo text,
    contacto_id uuid REFERENCES contacto (id) ON DELETE SET NULL,
    creado_en timestamptz NOT NULL DEFAULT now(),
    revisado_en timestamptz
);

CREATE INDEX idx_rueda_empresa_estado ON rueda_empresa (estado);

CREATE TABLE rueda_publicacion (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    empresa_id uuid NOT NULL REFERENCES rueda_empresa (id) ON DELETE CASCADE,
    tipo text NOT NULL CHECK (tipo IN ('oferta', 'demanda')),
    titulo text NOT NULL,
    descripcion text NOT NULL,
    imagen_url text,
    estado text NOT NULL DEFAULT 'pendiente'
        CHECK (estado IN ('pendiente', 'aprobado', 'rechazado')),
    motivo_rechazo text,
    creado_en timestamptz NOT NULL DEFAULT now(),
    revisado_en timestamptz
);

CREATE INDEX idx_rueda_publicacion_estado ON rueda_publicacion (estado, tipo);
CREATE INDEX idx_rueda_publicacion_empresa ON rueda_publicacion (empresa_id);

CREATE TABLE rueda_coincidencia (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    demanda_id uuid NOT NULL REFERENCES rueda_publicacion (id) ON DELETE CASCADE,
    oferta_id uuid NOT NULL REFERENCES rueda_publicacion (id) ON DELETE CASCADE,
    puntaje integer NOT NULL CHECK (puntaje BETWEEN 0 AND 100),
    motivo text NOT NULL,
    estado text NOT NULL DEFAULT 'sugerida'
        CHECK (estado IN ('sugerida', 'notificada', 'descartada')),
    creado_en timestamptz NOT NULL DEFAULT now(),
    notificada_en timestamptz,
    UNIQUE (demanda_id, oferta_id)
);

CREATE INDEX idx_rueda_coincidencia_estado ON rueda_coincidencia (estado);

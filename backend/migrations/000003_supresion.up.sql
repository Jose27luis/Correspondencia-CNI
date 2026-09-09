CREATE TABLE supresion (
    correo citext PRIMARY KEY,
    motivo text NOT NULL CHECK (motivo IN ('rebote', 'baja', 'manual')),
    detalle text,
    creado_en timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_supresion_motivo ON supresion (motivo);

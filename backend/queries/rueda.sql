-- name: RegistrarRuedaEmpresa :one
INSERT INTO rueda_empresa (
    razon_social, ruc, persona_encargada, cargo_encargado, correo, direccion,
    ciudad, region, pais, codigo_postal, telefono, celular, pagina_web
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (correo) DO UPDATE SET correo = EXCLUDED.correo
RETURNING *;

-- name: ObtenerRuedaEmpresa :one
SELECT * FROM rueda_empresa WHERE id = $1;

-- name: ListarRuedaEmpresas :many
SELECT e.*,
    (SELECT count(*) FROM rueda_publicacion p WHERE p.empresa_id = e.id) AS publicaciones
FROM rueda_empresa e
WHERE (sqlc.narg('estado')::text IS NULL OR e.estado = sqlc.narg('estado')::text)
ORDER BY e.creado_en DESC
LIMIT $1 OFFSET $2;

-- name: ContarRuedaEmpresas :one
SELECT count(*) FROM rueda_empresa
WHERE (sqlc.narg('estado')::text IS NULL OR estado = sqlc.narg('estado')::text);

-- name: AprobarRuedaEmpresa :one
UPDATE rueda_empresa
SET estado = 'aprobado',
    contacto_id = $2,
    motivo_rechazo = NULL,
    revisado_en = now()
WHERE id = $1
RETURNING *;

-- name: RechazarRuedaEmpresa :one
UPDATE rueda_empresa
SET estado = 'rechazado',
    motivo_rechazo = $2,
    revisado_en = now()
WHERE id = $1
RETURNING *;

-- name: CrearPublicacion :one
INSERT INTO rueda_publicacion (empresa_id, tipo, titulo, descripcion, imagen_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListarPublicaciones :many
SELECT p.*,
    e.razon_social, e.ciudad AS empresa_ciudad, e.pais AS empresa_pais,
    e.correo AS empresa_correo, e.estado AS empresa_estado
FROM rueda_publicacion p
JOIN rueda_empresa e ON e.id = p.empresa_id
WHERE (sqlc.narg('estado')::text IS NULL OR p.estado = sqlc.narg('estado')::text)
  AND (sqlc.narg('tipo')::text IS NULL OR p.tipo = sqlc.narg('tipo')::text)
ORDER BY p.creado_en DESC
LIMIT $1 OFFSET $2;

-- name: ContarPublicaciones :one
SELECT count(*) FROM rueda_publicacion p
WHERE (sqlc.narg('estado')::text IS NULL OR p.estado = sqlc.narg('estado')::text)
  AND (sqlc.narg('tipo')::text IS NULL OR p.tipo = sqlc.narg('tipo')::text);

-- name: ObtenerPublicacion :one
SELECT p.*,
    e.razon_social, e.ciudad AS empresa_ciudad, e.pais AS empresa_pais,
    e.correo AS empresa_correo, e.estado AS empresa_estado
FROM rueda_publicacion p
JOIN rueda_empresa e ON e.id = p.empresa_id
WHERE p.id = $1;

-- name: AprobarPublicacion :one
UPDATE rueda_publicacion
SET estado = 'aprobado',
    motivo_rechazo = NULL,
    revisado_en = now()
WHERE id = $1
RETURNING *;

-- name: RechazarPublicacion :one
UPDATE rueda_publicacion
SET estado = 'rechazado',
    motivo_rechazo = $2,
    revisado_en = now()
WHERE id = $1
RETURNING *;

-- name: ListarPublicacionesPublicas :many
SELECT p.id, p.tipo, p.titulo, p.descripcion, p.imagen_url, p.creado_en,
    e.razon_social, e.ciudad AS empresa_ciudad, e.pais AS empresa_pais
FROM rueda_publicacion p
JOIN rueda_empresa e ON e.id = p.empresa_id
WHERE p.estado = 'aprobado'
  AND e.estado = 'aprobado'
  AND (sqlc.narg('tipo')::text IS NULL OR p.tipo = sqlc.narg('tipo')::text)
ORDER BY p.creado_en DESC
LIMIT 200;

-- name: ListarCandidatas :many
SELECT p.id, p.titulo, p.descripcion,
    e.razon_social, e.ciudad AS empresa_ciudad, e.pais AS empresa_pais
FROM rueda_publicacion p
JOIN rueda_empresa e ON e.id = p.empresa_id
WHERE p.estado = 'aprobado'
  AND e.estado = 'aprobado'
  AND p.tipo = $1
  AND p.empresa_id <> $2
ORDER BY p.creado_en DESC
LIMIT 60;

-- name: GuardarCoincidencia :exec
INSERT INTO rueda_coincidencia (demanda_id, oferta_id, puntaje, motivo)
VALUES ($1, $2, $3, $4)
ON CONFLICT (demanda_id, oferta_id) DO UPDATE
SET puntaje = EXCLUDED.puntaje,
    motivo = EXCLUDED.motivo;

-- name: ListarCoincidencias :many
SELECT c.*,
    d.titulo AS demanda_titulo, de.razon_social AS demanda_empresa, de.correo AS demanda_correo,
    o.titulo AS oferta_titulo, oe.razon_social AS oferta_empresa, oe.correo AS oferta_correo
FROM rueda_coincidencia c
JOIN rueda_publicacion d ON d.id = c.demanda_id
JOIN rueda_empresa de ON de.id = d.empresa_id
JOIN rueda_publicacion o ON o.id = c.oferta_id
JOIN rueda_empresa oe ON oe.id = o.empresa_id
WHERE (sqlc.narg('estado')::text IS NULL OR c.estado = sqlc.narg('estado')::text)
ORDER BY c.puntaje DESC, c.creado_en DESC
LIMIT 200;

-- name: ObtenerCoincidencia :one
SELECT c.*,
    d.titulo AS demanda_titulo, d.descripcion AS demanda_descripcion,
    de.razon_social AS demanda_empresa, de.correo AS demanda_correo,
    de.persona_encargada AS demanda_encargado,
    o.titulo AS oferta_titulo, o.descripcion AS oferta_descripcion,
    oe.razon_social AS oferta_empresa, oe.correo AS oferta_correo,
    oe.persona_encargada AS oferta_encargado
FROM rueda_coincidencia c
JOIN rueda_publicacion d ON d.id = c.demanda_id
JOIN rueda_empresa de ON de.id = d.empresa_id
JOIN rueda_publicacion o ON o.id = c.oferta_id
JOIN rueda_empresa oe ON oe.id = o.empresa_id
WHERE c.id = $1;

-- name: MarcarCoincidencia :one
UPDATE rueda_coincidencia
SET estado = sqlc.arg('estado')::text,
    notificada_en = CASE
        WHEN sqlc.arg('estado')::text = 'notificada' THEN now()
        ELSE notificada_en
    END
WHERE id = sqlc.arg('id')
RETURNING *;

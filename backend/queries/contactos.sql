-- name: CrearContacto :one
INSERT INTO contacto (nombre, empresa, correo, pais, campos_extra)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ObtenerContacto :one
SELECT * FROM contacto WHERE id = $1;

-- name: ObtenerContactoPorCorreo :one
SELECT * FROM contacto WHERE correo = $1;

-- name: ListarContactos :many
SELECT * FROM contacto
WHERE (sqlc.narg('busqueda')::text IS NULL
   OR nombre ILIKE '%' || sqlc.narg('busqueda')::text || '%'
   OR empresa ILIKE '%' || sqlc.narg('busqueda')::text || '%'
   OR correo ILIKE '%' || sqlc.narg('busqueda')::text || '%')
ORDER BY creado_en DESC
LIMIT $1 OFFSET $2;

-- name: ContarContactos :one
SELECT count(*) FROM contacto
WHERE (sqlc.narg('busqueda')::text IS NULL
   OR nombre ILIKE '%' || sqlc.narg('busqueda')::text || '%'
   OR empresa ILIKE '%' || sqlc.narg('busqueda')::text || '%'
   OR correo ILIKE '%' || sqlc.narg('busqueda')::text || '%');

-- name: ActualizarContacto :one
UPDATE contacto
SET nombre = $2,
    empresa = $3,
    correo = $4,
    pais = $5,
    campos_extra = $6,
    actualizado_en = now()
WHERE id = $1
RETURNING *;

-- name: EliminarContacto :execrows
DELETE FROM contacto WHERE id = $1;

-- name: ImportarContacto :one
INSERT INTO contacto (nombre, empresa, correo, pais, campos_extra)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (correo) DO UPDATE
SET nombre = EXCLUDED.nombre,
    empresa = EXCLUDED.empresa,
    pais = EXCLUDED.pais,
    campos_extra = contacto.campos_extra || EXCLUDED.campos_extra,
    actualizado_en = now()
RETURNING *, (xmax = 0) AS fue_creado;

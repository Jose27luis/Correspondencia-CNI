-- name: SuprimirCorreo :exec
INSERT INTO supresion (correo, motivo, detalle)
VALUES ($1, $2, $3)
ON CONFLICT (correo) DO UPDATE
SET motivo = EXCLUDED.motivo,
    detalle = EXCLUDED.detalle;

-- name: EstaSuprimido :one
SELECT EXISTS (SELECT 1 FROM supresion WHERE correo = $1);

-- name: ListarSuprimidos :many
SELECT * FROM supresion ORDER BY creado_en DESC LIMIT $1 OFFSET $2;

-- name: ContarSuprimidos :one
SELECT count(*) FROM supresion;

-- name: QuitarSupresion :execrows
DELETE FROM supresion WHERE correo = $1;

-- name: SuprimirPorEnvio :exec
INSERT INTO supresion (correo, motivo, detalle)
SELECT c.correo, $2, $3
FROM envio e
JOIN contacto c ON c.id = e.contacto_id
WHERE e.id = $1
ON CONFLICT (correo) DO UPDATE
SET motivo = EXCLUDED.motivo,
    detalle = EXCLUDED.detalle;

-- name: ObtenerContactoPorID :one
SELECT * FROM contacto WHERE id = $1;

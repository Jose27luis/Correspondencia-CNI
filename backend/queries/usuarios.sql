-- name: CrearUsuario :one
INSERT INTO usuario (nombre, correo, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ObtenerUsuarioPorCorreo :one
SELECT * FROM usuario WHERE correo = $1;

-- name: ObtenerUsuario :one
SELECT * FROM usuario WHERE id = $1;

-- name: ContarUsuarios :one
SELECT count(*) FROM usuario;

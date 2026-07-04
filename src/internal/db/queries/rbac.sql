-- name: GetAllRolePermissions :many
SELECT r.code AS role_code, p.code AS permission_code
FROM roles r
JOIN role_permissions rp ON r.code = rp.role_code
JOIN permissions p ON rp.permission_code = p.code;

-- name: CreateRole :one
INSERT INTO roles (code, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreatePermission :one
INSERT INTO permissions (code, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: AssignPermissionToRole :exec
INSERT INTO role_permissions (role_code, permission_code)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

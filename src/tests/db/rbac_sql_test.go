package db_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"emc_lb/src/internal/db/sqlc"

	"github.com/pashagolub/pgxmock/v4"
)

func now() time.Time {
	return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
}

func TestCreateRoleAndPermission(t *testing.T) {
	q, mock := newQueries(t)

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO roles (code, name, description)
VALUES ($1, $2, $3)
RETURNING code, name, description, created_at`)).
		WithArgs("admin", "Admin", (*string)(nil)).
		WillReturnRows(pgxmock.NewRows([]string{"code", "name", "description", "created_at"}).
			AddRow("admin", "Admin", nil, now()))

	role, err := q.CreateRole(context.Background(), sqlc.CreateRoleParams{
		Code: "admin",
		Name: "Admin",
	})
	if err != nil {
		t.Fatalf("CreateRole returned error: %v", err)
	}
	if role.Code != "admin" {
		t.Fatalf("expected code admin, got %s", role.Code)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO permissions (code, name, description)
VALUES ($1, $2, $3)
RETURNING code, name, description, created_at`)).
		WithArgs("manage_orders", "Manage Orders", (*string)(nil)).
		WillReturnRows(pgxmock.NewRows([]string{"code", "name", "description", "created_at"}).
			AddRow("manage_orders", "Manage Orders", nil, now()))

	perm, err := q.CreatePermission(context.Background(), sqlc.CreatePermissionParams{
		Code: "manage_orders",
		Name: "Manage Orders",
	})
	if err != nil {
		t.Fatalf("CreatePermission returned error: %v", err)
	}
	if perm.Code != "manage_orders" {
		t.Fatalf("expected code manage_orders, got %s", perm.Code)
	}
}

func TestAssignPermissionToRole(t *testing.T) {
	q, mock := newQueries(t)

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO role_permissions (role_code, permission_code)
VALUES ($1, $2)
ON CONFLICT DO NOTHING`)).
		WithArgs("admin", "manage_orders").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := q.AssignPermissionToRole(context.Background(), sqlc.AssignPermissionToRoleParams{
		RoleCode:       "admin",
		PermissionCode: "manage_orders",
	}); err != nil {
		t.Fatalf("AssignPermissionToRole returned error: %v", err)
	}
}

func TestGetAllRolePermissions(t *testing.T) {
	q, mock := newQueries(t)

	rows := pgxmock.NewRows([]string{"role_code", "permission_code"}).
		AddRow("admin", "manage_orders").
		AddRow("admin", "manage_users").
		AddRow("customer", "view_orders")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT r.code AS role_code, p.code AS permission_code
FROM roles r
JOIN role_permissions rp ON r.code = rp.role_code
JOIN permissions p ON rp.permission_code = p.code`)).
		WillReturnRows(rows)

	items, err := q.GetAllRolePermissions(context.Background())
	if err != nil {
		t.Fatalf("GetAllRolePermissions returned error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(items))
	}

	permissionMap := make(map[string]bool)
	for _, rp := range items {
		permissionMap[rp.RoleCode+":"+rp.PermissionCode] = true
	}
	for _, key := range []string{"admin:manage_orders", "admin:manage_users", "customer:view_orders"} {
		if !permissionMap[key] {
			t.Fatalf("expected to find permission %s", key)
		}
	}
}

func TestGetAllRolePermissions_Empty(t *testing.T) {
	q, mock := newQueries(t)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT r.code AS role_code, p.code AS permission_code
FROM roles r
JOIN role_permissions rp ON r.code = rp.role_code
JOIN permissions p ON rp.permission_code = p.code`)).
		WillReturnRows(pgxmock.NewRows([]string{"role_code", "permission_code"}))

	items, err := q.GetAllRolePermissions(context.Background())
	if err != nil {
		t.Fatalf("GetAllRolePermissions returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(items))
	}
}

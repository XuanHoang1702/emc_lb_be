package service

import (
	"context"
	
	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/pkg/auth"
)

type RBACService interface {
	CreateRole(ctx context.Context, code, name, description string) (sqlc.Role, error)
	CreatePermission(ctx context.Context, code, name, description string) (sqlc.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleCode, permissionCode string) error
}

type rbacService struct {
	queries *sqlc.Queries
}

func NewRBACService(queries *sqlc.Queries) RBACService {
	return &rbacService{queries: queries}
}

func (s *rbacService) CreateRole(ctx context.Context, code, name, description string) (sqlc.Role, error) {
	return s.queries.CreateRole(ctx, sqlc.CreateRoleParams{
		Code:        code,
		Name:        name,
		Description: &description,
	})
}

func (s *rbacService) CreatePermission(ctx context.Context, code, name, description string) (sqlc.Permission, error) {
	return s.queries.CreatePermission(ctx, sqlc.CreatePermissionParams{
		Code:        code,
		Name:        name,
		Description: &description,
	})
}

func (s *rbacService) AssignPermissionToRole(ctx context.Context, roleCode, permissionCode string) error {
	err := s.queries.AssignPermissionToRole(ctx, sqlc.AssignPermissionToRoleParams{
		RoleCode:       roleCode,
		PermissionCode: permissionCode,
	})
	if err != nil {
		return err
	}
	
	// Notify other nodes to refresh their cache
	return auth.GetRBACManager().PublishCacheInvalidation(ctx)
}

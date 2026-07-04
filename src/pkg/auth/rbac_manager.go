package auth

import (
	"context"
	"fmt"
	"log"
	"sync"

	"emc_lb/src/internal/db/sqlc"
)

type RBACManager struct {
	mu          sync.RWMutex
	permissions map[string]map[string]bool // map[role_code]map[permission_code]bool
}

var (
	manager *RBACManager
	once    sync.Once
)

func InitRBACManager(queries *sqlc.Queries) {
	once.Do(func() {
		manager = &RBACManager{
			permissions: make(map[string]map[string]bool),
		}
		if err := manager.LoadPermissions(queries); err != nil {
			log.Fatalf("Failed to initialize RBAC Manager: %v", err)
		}
	})
}

func GetRBACManager() *RBACManager {
	if manager == nil {
		log.Fatal("RBACManager is not initialized. Call InitRBACManager first.")
	}
	return manager
}

func (m *RBACManager) LoadPermissions(queries *sqlc.Queries) error {
	ctx := context.Background()
	rolePerms, err := queries.GetAllRolePermissions(ctx)
	if err != nil {
		return fmt.Errorf("could not load role permissions from db: %w", err)
	}

	newPerms := make(map[string]map[string]bool)
	for _, rp := range rolePerms {
		if _, exists := newPerms[rp.RoleCode]; !exists {
			newPerms[rp.RoleCode] = make(map[string]bool)
		}
		newPerms[rp.RoleCode][rp.PermissionCode] = true
	}

	m.mu.Lock()
	m.permissions = newPerms
	m.mu.Unlock()

	return nil
}

func (m *RBACManager) HasPermission(roleCode string, permissionCode string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if rolePerms, ok := m.permissions[roleCode]; ok {
		return rolePerms[permissionCode]
	}
	return false
}

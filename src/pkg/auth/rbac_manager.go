package auth

import (
	"context"
	"fmt"
	"log"
	"sync"

	"emc_lb/src/internal/db/sqlc"

	"github.com/redis/go-redis/v9"
)

type RBACManager struct {
	mu          sync.RWMutex
	permissions map[string]map[string]bool // map[role_code]map[permission_code]bool
	redisClient *redis.Client
	queries     *sqlc.Queries
}

var (
	manager *RBACManager
	once    sync.Once
)

func InitRBACManager(queries *sqlc.Queries, redisClient *redis.Client) {
	once.Do(func() {
		manager = &RBACManager{
			permissions: make(map[string]map[string]bool),
			redisClient: redisClient,
			queries:     queries,
		}
		if err := manager.LoadPermissions(queries); err != nil {
			log.Fatalf("Failed to initialize RBAC Manager: %v", err)
		}

		if redisClient != nil {
			go manager.startCacheInvalidationListener()
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

func (m *RBACManager) RefreshPermissions(queries *sqlc.Queries) error {
	return m.LoadPermissions(queries)
}

func (m *RBACManager) HasPermission(roleCode string, permissionCode string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if rolePerms, ok := m.permissions[roleCode]; ok {
		return rolePerms[permissionCode]
	}
	return false
}

func (m *RBACManager) startCacheInvalidationListener() {
	ctx := context.Background()
	pubsub := m.redisClient.Subscribe(ctx, "emc_lb:rbac_cache_invalidate")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		log.Printf("Received RBAC cache invalidation signal: %s", msg.Payload)
		if err := m.RefreshPermissions(m.queries); err != nil {
			log.Printf("Failed to refresh RBAC permissions from DB: %v", err)
		}
	}
}

func (m *RBACManager) PublishCacheInvalidation(ctx context.Context) error {
	if m.redisClient == nil {
		return nil
	}
	return m.redisClient.Publish(ctx, "emc_lb:rbac_cache_invalidate", "refresh").Err()
}

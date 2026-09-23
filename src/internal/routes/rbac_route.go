package route

import (
	"emc_lb/src/internal/handler"
	"emc_lb/src/internal/middleware"
	"emc_lb/src/pkg/config"

	"github.com/gin-gonic/gin"
)

type RBACRoute struct {
	rbacHandler *handler.RBACHandler
	cfg         *config.AppConfig
}

func NewRBACRoute(rbacHandler *handler.RBACHandler, cfg *config.AppConfig) *RBACRoute {
	return &RBACRoute{
		rbacHandler: rbacHandler,
		cfg:         cfg,
	}
}

func (r *RBACRoute) RegisterPublic(router gin.IRouter) {
	// No public routes for RBAC
}

func (r *RBACRoute) RegisterProtected(router gin.IRouter) {
	rbacGroup := router.Group("/rbac")
	
	// Only superadmin or admin can manage RBAC
	// AccessTokenMiddleware is already applied by the protectedGroup
	rbacGroup.Use(middleware.RequirePermission("manage_rbac")) // Will be bypassed if role=admin/superadmin

	rbacGroup.POST("/roles", r.rbacHandler.HandleCreateRole)
	rbacGroup.POST("/permissions", r.rbacHandler.HandleCreatePermission)
	rbacGroup.POST("/roles/:role_code/permissions", r.rbacHandler.HandleAssignPermission)
}

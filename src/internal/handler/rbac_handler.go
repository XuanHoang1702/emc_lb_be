package handler

import (
	"net/http"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/res"
	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	rbacService service.RBACService
}

func NewRBACHandler(rbacService service.RBACService) *RBACHandler {
	return &RBACHandler{rbacService: rbacService}
}

func (h *RBACHandler) HandleCreateRole(ctx *gin.Context) {
	var req struct {
		Code        string `json:"code" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		res.Error(ctx, err)
		return
	}

	role, err := h.rbacService.CreateRole(ctx.Request.Context(), req.Code, req.Name, req.Description)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusCreated, "Role created successfully", role)
}

func (h *RBACHandler) HandleCreatePermission(ctx *gin.Context) {
	var req struct {
		Code        string `json:"code" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		res.Error(ctx, err)
		return
	}

	permission, err := h.rbacService.CreatePermission(ctx.Request.Context(), req.Code, req.Name, req.Description)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusCreated, "Permission created successfully", permission)
}

func (h *RBACHandler) HandleAssignPermission(ctx *gin.Context) {
	roleCode := ctx.Param("role_code")
	var req struct {
		PermissionCode string `json:"permission_code" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		res.Error(ctx, err)
		return
	}

	if err := h.rbacService.AssignPermissionToRole(ctx.Request.Context(), roleCode, req.PermissionCode); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, "Permission assigned to role successfully", nil)
}

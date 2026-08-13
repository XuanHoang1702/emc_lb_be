package handler

import (
	"net/http"

	"emc_lb/src/internal/middleware"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/auth"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/validation"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// HandleCreate godoc
// @Summary      Create order
// @Description  Create a new order from cart items. Orders are split by shop automatically.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        body body entities.CreateOrderRequest true "Order details with items and optional coupon"
// @Success      201  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/orders [post]
func (h *OrderHandler) HandleCreate(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		res.Error(ctx, &res.AppError{
			Message:    "Unauthorized",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	var createRequest entities.CreateOrderRequest
	if err := validation.BindJSON(ctx, &createRequest, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	orders, err := h.orderService.CreateOrder(ctx.Request.Context(), userID.(string), createRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusCreated, "Orders created successfully", orders)
}

// HandleMyOrders godoc
// @Summary      List my orders
// @Description  Get all orders for the authenticated user
// @Tags         Orders
// @Produce      json
// @Success      200  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/orders/my [get]
func (h *OrderHandler) HandleMyOrders(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		res.Error(ctx, &res.AppError{
			Message:    "Unauthorized",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	orders, err := h.orderService.ListOrders(ctx.Request.Context(), userID.(string))
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, orders)
}

// HandleGetAllOrders godoc
// @Summary      List all orders (admin)
// @Description  Get all orders in the system (requires manage_orders permission)
// @Tags         Orders
// @Produce      json
// @Success      200  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Failure      403  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/orders [get]
func (h *OrderHandler) HandleGetAllOrders(ctx *gin.Context) {
	orders, err := h.orderService.ListOrders(ctx.Request.Context(), "")
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, orders)
}

// HandleGetOrder godoc
// @Summary      Get order by ID
// @Description  Get a single order by its ID (owner or admin only)
// @Tags         Orders
// @Produce      json
// @Param        id   path      string  true  "Order ID"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      403  {object}  res.APIResponse
// @Failure      404  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/orders/{id} [get]
func (h *OrderHandler) HandleGetOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid order ID",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	order, err := h.orderService.GetOrder(ctx.Request.Context(), id)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	// Add check to ensure the user requesting this is either the owner or an admin
	userID, _ := ctx.Get(middleware.ContextUserIDKey)
	role, _ := ctx.Get(middleware.ContextRoleKey)
	if !auth.GetRBACManager().HasPermission(role.(string), "manage_orders") && order.UserID != userID.(string) {
		res.Error(ctx, &res.AppError{
			Message:    "Forbidden: You can only view your own orders",
			Code:       erres.CommonForbidden,
			StatusCode: http.StatusForbidden,
		})
		return
	}

	res.Success(ctx, http.StatusOK, order)
}

// HandleUpdateStatus godoc
// @Summary      Update order status
// @Description  Update an order's status (requires manage_orders permission)
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Order ID"
// @Param        body body object true "New status" SchemaExample({"status": "processing"})
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/orders/{id}/status [patch]
func (h *OrderHandler) HandleUpdateStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid order ID",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := validation.BindJSON(ctx, &req, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	if err := h.orderService.UpdateOrderStatus(ctx.Request.Context(), id, req.Status); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, map[string]string{"message": "Order status updated successfully"})
}

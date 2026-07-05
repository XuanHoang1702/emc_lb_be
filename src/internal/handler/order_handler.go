package handler

import (
	"net/http"

	"emc_lb/src/internal/middleware"
	"emc_lb/src/internal/service"
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

	order, err := h.orderService.CreateOrder(ctx.Request.Context(), userID.(string), createRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusCreated, order)
}

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

func (h *OrderHandler) HandleGetAllOrders(ctx *gin.Context) {
	orders, err := h.orderService.ListOrders(ctx.Request.Context(), "")
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, orders)
}

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
	
	if role != "admin" && order.UserID != userID.(string) {
		res.Error(ctx, &res.AppError{
			Message:    "Forbidden: You can only view your own orders",
			Code:       "FORBIDDEN",
			StatusCode: http.StatusForbidden,
		})
		return
	}

	res.Success(ctx, http.StatusOK, order)
}

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

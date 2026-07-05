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

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func (h *CartHandler) HandleGetCart(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		res.Error(ctx, &res.AppError{
			Message:    "Unauthorized",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	cart, err := h.cartService.GetCart(ctx.Request.Context(), userID.(string))
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, cart)
}

func (h *CartHandler) HandleAddItem(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		res.Error(ctx, &res.AppError{
			Message:    "Unauthorized",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	var req entities.AddCartItemRequest
	if err := validation.BindJSON(ctx, &req, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	cart, err := h.cartService.AddItem(ctx.Request.Context(), userID.(string), req)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, cart)
}

func (h *CartHandler) HandleUpdateItem(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		res.Error(ctx, &res.AppError{
			Message:    "Unauthorized",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	productID := ctx.Param("productId")
	if productID == "" {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid product ID",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	var req entities.UpdateCartItemRequest
	if err := validation.BindJSON(ctx, &req, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	cart, err := h.cartService.UpdateItem(ctx.Request.Context(), userID.(string), productID, req)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, cart)
}

func (h *CartHandler) HandleRemoveItem(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		res.Error(ctx, &res.AppError{
			Message:    "Unauthorized",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	productID := ctx.Param("productId")
	if productID == "" {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid product ID",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	cart, err := h.cartService.RemoveItem(ctx.Request.Context(), userID.(string), productID)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, cart)
}

func (h *CartHandler) HandleClearCart(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		res.Error(ctx, &res.AppError{
			Message:    "Unauthorized",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	if err := h.cartService.ClearCart(ctx.Request.Context(), userID.(string)); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, map[string]string{"message": "Cart cleared successfully"})
}

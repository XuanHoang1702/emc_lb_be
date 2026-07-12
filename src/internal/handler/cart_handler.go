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

// HandleGetCart godoc
// @Summary      Get cart
// @Description  Get the shopping cart for the authenticated user
// @Tags         Cart
// @Produce      json
// @Success      200  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/cart [get]
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

// HandleAddItem godoc
// @Summary      Add item to cart
// @Description  Add a product to the shopping cart
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Param        body body entities.AddCartItemRequest true "Item to add"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/cart/items [post]
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

// HandleUpdateItem godoc
// @Summary      Update cart item
// @Description  Update quantity of a product in the cart
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Param        productId path string true "Product ID"
// @Param        body body entities.UpdateCartItemRequest true "New quantity"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/cart/items/{productId} [put]
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

// HandleRemoveItem godoc
// @Summary      Remove item from cart
// @Description  Remove a product from the shopping cart
// @Tags         Cart
// @Produce      json
// @Param        productId path string true "Product ID"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/cart/items/{productId} [delete]
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

// HandleClearCart godoc
// @Summary      Clear cart
// @Description  Remove all items from the shopping cart
// @Tags         Cart
// @Produce      json
// @Success      200  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/cart [delete]
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

// HandleApplyCoupon godoc
// @Summary      Apply coupon to cart
// @Description  Apply a discount coupon code to the shopping cart
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Param        body body entities.ApplyCouponRequest true "Coupon code"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/cart/coupon [post]
func (h *CartHandler) HandleApplyCoupon(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		res.Error(ctx, &res.AppError{
			Message:    "Unauthorized",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	var req entities.ApplyCouponRequest
	if err := validation.BindJSON(ctx, &req, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	cart, err := h.cartService.ApplyCoupon(ctx.Request.Context(), userID.(string), req)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, cart)
}

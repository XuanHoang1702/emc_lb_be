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

type ShopHandler interface {
	CreateShop(ctx *gin.Context)
	GetMyShop(ctx *gin.Context)
}

type shopHandler struct {
	shopService service.ShopService
}

func NewShopHandler(shopService service.ShopService) ShopHandler {
	return &shopHandler{shopService: shopService}
}

// CreateShop godoc
// @Summary      Create a new shop
// @Description  Create a new shop for the current user
// @Tags         Shops
// @Accept       json
// @Produce      json
// @Param        body body entities.CreateShopRequest true "Shop details"
// @Success      201  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/shops [post]
func (h *shopHandler) CreateShop(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		res.Error(ctx, &res.AppError{
			Message:    "Unauthorized",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	var req entities.CreateShopRequest
	if err := validation.BindJSON(ctx, &req, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	shop, err := h.shopService.CreateShop(ctx.Request.Context(), userID.(string), req)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusCreated, "Shop created successfully", shop)
}

// GetMyShop godoc
// @Summary      Get my shop
// @Description  Get the shop details for the current user
// @Tags         Shops
// @Produce      json
// @Success      200  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Failure      404  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/shops/my-shop [get]
func (h *shopHandler) GetMyShop(ctx *gin.Context) {
	userID, exists := ctx.Get(middleware.ContextUserIDKey)
	if !exists {
		res.Error(ctx, &res.AppError{
			Message:    "Unauthorized",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	shop, err := h.shopService.GetShopByOwnerID(ctx.Request.Context(), userID.(string))
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, "Success", shop)
}

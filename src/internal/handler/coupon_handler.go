package handler

import (
	"net/http"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/validation"

	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	couponService service.CouponService
}

func NewCouponHandler(couponService service.CouponService) *CouponHandler {
	return &CouponHandler{couponService: couponService}
}

// HandleCreate godoc
// @Summary      Create coupon
// @Description  Create a new discount coupon (requires coupon:create permission)
// @Tags         Coupons
// @Accept       json
// @Produce      json
// @Param        body body entities.CreateCouponRequest true "Coupon details"
// @Success      201  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/coupons [post]
func (h *CouponHandler) HandleCreate(ctx *gin.Context) {
	var req entities.CreateCouponRequest
	if err := validation.BindJSON(ctx, &req, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	coupon, err := h.couponService.Create(ctx.Request.Context(), req)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusCreated, coupon)
}

// HandleList godoc
// @Summary      List coupons
// @Description  Get all coupons (requires coupon:read permission)
// @Tags         Coupons
// @Produce      json
// @Success      200  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/coupons [get]
func (h *CouponHandler) HandleList(ctx *gin.Context) {
	coupons, err := h.couponService.List(ctx.Request.Context())
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, coupons)
}

// HandleGet godoc
// @Summary      Get coupon by code
// @Description  Get a single coupon by its code (requires coupon:read permission)
// @Tags         Coupons
// @Produce      json
// @Param        code path      string  true  "Coupon code"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      404  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/coupons/{code} [get]
func (h *CouponHandler) HandleGet(ctx *gin.Context) {
	code := ctx.Param("code")
	if code == "" {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid coupon code",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	coupon, err := h.couponService.GetByCode(ctx.Request.Context(), code)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, coupon)
}

// HandleUpdate godoc
// @Summary      Update coupon
// @Description  Update an existing coupon (requires coupon:update permission)
// @Tags         Coupons
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Coupon ID"
// @Param        body body entities.UpdateCouponRequest true "Fields to update"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/coupons/{id} [patch]
func (h *CouponHandler) HandleUpdate(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid coupon id",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	var req entities.UpdateCouponRequest
	if err := validation.BindJSON(ctx, &req, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	coupon, err := h.couponService.Update(ctx.Request.Context(), id, req)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, coupon)
}

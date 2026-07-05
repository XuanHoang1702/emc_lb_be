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

func (h *CouponHandler) HandleList(ctx *gin.Context) {
	coupons, err := h.couponService.List(ctx.Request.Context())
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, coupons)
}

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

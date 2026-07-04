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

type BrandHandler struct {
	brandService service.BrandService
}

func NewBrandHandler(brandService service.BrandService) *BrandHandler {
	return &BrandHandler{brandService: brandService}
}

func (h *BrandHandler) HandleCreate(ctx *gin.Context) {
	var createRequest entities.CreateBrandRequest
	if err := validation.BindJSON(ctx, &createRequest, erres.BrandValidationFailed); err != nil {
		res.Error(ctx, err)
		return
	}

	brand, err := h.brandService.Create(ctx.Request.Context(), createRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusCreated, brand)
}

func (h *BrandHandler) HandleList(ctx *gin.Context) {
	brands, err := h.brandService.List(ctx.Request.Context())
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, brands)
}

func (h *BrandHandler) HandleGetByID(ctx *gin.Context) {
	var uriRequest entities.BrandURIRequest
	if err := validation.BindURI(ctx, &uriRequest, erres.BrandValidationFailed); err != nil {
		res.Error(ctx, err)
		return
	}

	brand, err := h.brandService.GetByID(ctx.Request.Context(), uriRequest.ID)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, brand)
}

func (h *BrandHandler) HandleUpdate(ctx *gin.Context) {
	var uriRequest entities.BrandURIRequest
	if err := validation.BindURI(ctx, &uriRequest, erres.BrandValidationFailed); err != nil {
		res.Error(ctx, err)
		return
	}

	var updateRequest entities.UpdateBrandRequest
	if err := validation.BindJSON(ctx, &updateRequest, erres.BrandValidationFailed); err != nil {
		res.Error(ctx, err)
		return
	}

	brand, err := h.brandService.Update(ctx.Request.Context(), uriRequest.ID, updateRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, brand)
}

func (h *BrandHandler) HandleDelete(ctx *gin.Context) {
	var uriRequest entities.BrandURIRequest
	if err := validation.BindURI(ctx, &uriRequest, erres.BrandValidationFailed); err != nil {
		res.Error(ctx, err)
		return
	}

	if err := h.brandService.Delete(ctx.Request.Context(), uriRequest.ID); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, "success_brand_deleted")
}

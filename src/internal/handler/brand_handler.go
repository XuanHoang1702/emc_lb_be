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

// HandleCreate godoc
// @Summary      Create brand
// @Description  Create a new brand
// @Tags         Brands
// @Accept       json
// @Produce      json
// @Param        body body entities.CreateBrandRequest true "Brand details"
// @Success      201  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/brands [post]
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

// HandleList godoc
// @Summary      List brands
// @Description  Get all brands
// @Tags         Brands
// @Produce      json
// @Success      200  {object}  res.APIResponse
// @Router       /api/v1/brands [get]
func (h *BrandHandler) HandleList(ctx *gin.Context) {
	brands, err := h.brandService.List(ctx.Request.Context())
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, brands)
}

// HandleGetByID godoc
// @Summary      Get brand by ID
// @Description  Get a single brand by its ID
// @Tags         Brands
// @Produce      json
// @Param        id   path      string  true  "Brand ID"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      404  {object}  res.APIResponse
// @Router       /api/v1/brands/{id} [get]
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

// HandleUpdate godoc
// @Summary      Update brand
// @Description  Update an existing brand
// @Tags         Brands
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Brand ID"
// @Param        body body entities.UpdateBrandRequest true "Fields to update"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/brands/{id} [patch]
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

// HandleDelete godoc
// @Summary      Delete brand
// @Description  Soft-delete a brand
// @Tags         Brands
// @Produce      json
// @Param        id   path      string  true  "Brand ID"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/brands/{id} [delete]
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

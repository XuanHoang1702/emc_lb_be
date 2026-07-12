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

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// HandleCreate godoc
// @Summary      Create product
// @Description  Create a new product (requires product:create permission)
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        body body entities.CreateProductRequest true "Product details"
// @Success      201  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/products [post]
func (h *ProductHandler) HandleCreate(ctx *gin.Context) {
	var createRequest entities.CreateProductRequest
	if err := validation.BindJSON(ctx, &createRequest, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	product, err := h.productService.Create(ctx.Request.Context(), createRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusCreated, product)
}

// HandleList godoc
// @Summary      List products
// @Description  Get all products
// @Tags         Products
// @Produce      json
// @Success      200  {object}  res.APIResponse
// @Failure      500  {object}  res.APIResponse
// @Router       /api/v1/products [get]
func (h *ProductHandler) HandleList(ctx *gin.Context) {
	products, err := h.productService.List(ctx.Request.Context())
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, products)
}

// HandleGetByID godoc
// @Summary      Get product by ID
// @Description  Get a single product by its ID
// @Tags         Products
// @Produce      json
// @Param        id   path      string  true  "Product ID"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      404  {object}  res.APIResponse
// @Router       /api/v1/products/{id} [get]
func (h *ProductHandler) HandleGetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid product ID",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	product, err := h.productService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, product)
}

// HandleUpdate godoc
// @Summary      Update product
// @Description  Update an existing product (requires product:update permission)
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Product ID"
// @Param        body body entities.UpdateProductRequest true "Fields to update"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/products/{id} [patch]
func (h *ProductHandler) HandleUpdate(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid product ID",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	var updateRequest entities.UpdateProductRequest
	if err := validation.BindJSON(ctx, &updateRequest, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	product, err := h.productService.Update(ctx.Request.Context(), id, updateRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, product)
}

// HandleDelete godoc
// @Summary      Delete product
// @Description  Soft-delete a product (requires product:delete permission)
// @Tags         Products
// @Produce      json
// @Param        id   path      string  true  "Product ID"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/products/{id} [delete]
func (h *ProductHandler) HandleDelete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid product ID",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	if err := h.productService.Delete(ctx.Request.Context(), id); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, map[string]interface{}{
		"message": "Product deleted successfully",
	})
}


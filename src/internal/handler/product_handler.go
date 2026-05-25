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

func (h *ProductHandler) HandleList(ctx *gin.Context) {
	products, err := h.productService.List(ctx.Request.Context())
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, products)
}

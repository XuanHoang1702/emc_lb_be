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

type CategoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) HandleCreate(ctx *gin.Context) {
	var createRequest entities.CreateCategoryRequest
	if err := validation.BindJSON(ctx, &createRequest, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	category, err := h.categoryService.Create(ctx.Request.Context(), createRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusCreated, category)
}

func (h *CategoryHandler) HandleList(ctx *gin.Context) {
	categories, err := h.categoryService.List(ctx.Request.Context())
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, categories)
}

func (h *CategoryHandler) HandleGetByID(ctx *gin.Context) {
	var uriRequest entities.CategoryURIRequest
	if err := validation.BindURI(ctx, &uriRequest, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	category, err := h.categoryService.GetByID(ctx.Request.Context(), uriRequest.ID)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, category)
}

func (h *CategoryHandler) HandleUpdate(ctx *gin.Context) {
	var uriRequest entities.CategoryURIRequest
	if err := validation.BindURI(ctx, &uriRequest, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	var updateRequest entities.UpdateCategoryRequest
	if err := validation.BindJSON(ctx, &updateRequest, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	category, err := h.categoryService.Update(ctx.Request.Context(), uriRequest.ID, updateRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, category)
}

func (h *CategoryHandler) HandleDelete(ctx *gin.Context) {
	var uriRequest entities.CategoryURIRequest
	if err := validation.BindURI(ctx, &uriRequest, erres.CommonBadRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	if err := h.categoryService.Delete(ctx.Request.Context(), uriRequest.ID); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, "success_category_deleted")
}

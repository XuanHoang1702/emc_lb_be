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

// HandleCreate godoc
// @Summary      Create category
// @Description  Create a new product category (requires category:create permission)
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        body body entities.CreateCategoryRequest true "Category details"
// @Success      201  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Failure      409  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/categories [post]
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

// HandleList godoc
// @Summary      List categories
// @Description  Get all product categories
// @Tags         Categories
// @Produce      json
// @Success      200  {object}  res.APIResponse
// @Failure      500  {object}  res.APIResponse
// @Router       /api/v1/categories [get]
func (h *CategoryHandler) HandleList(ctx *gin.Context) {
	categories, err := h.categoryService.List(ctx.Request.Context())
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, categories)
}

// HandleGetByID godoc
// @Summary      Get category by ID
// @Description  Get a single category by its ID
// @Tags         Categories
// @Produce      json
// @Param        id   path      string  true  "Category ID"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      404  {object}  res.APIResponse
// @Router       /api/v1/categories/{id} [get]
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

// HandleUpdate godoc
// @Summary      Update category
// @Description  Update an existing category (requires category:update permission)
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Category ID"
// @Param        body body entities.UpdateCategoryRequest true "Fields to update"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/categories/{id} [patch]
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

// HandleDelete godoc
// @Summary      Delete category
// @Description  Soft-delete a category (requires category:delete permission)
// @Tags         Categories
// @Produce      json
// @Param        id   path      string  true  "Category ID"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/categories/{id} [delete]
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


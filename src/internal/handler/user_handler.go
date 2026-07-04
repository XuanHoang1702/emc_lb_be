package handler

import (
	"io"
	"net/http"

	"emc_lb/src/internal/middleware"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/entities"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/validation"

	"emc_lb/src/pkg/errors"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) HandleRegister(ctx *gin.Context) {
	var registerRequest entities.RegisterUserRequest
	err := validation.BindJSON(ctx, &registerRequest, errors.UserInvalidFormat)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	user, err := h.userService.Register(ctx.Request.Context(), registerRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusCreated, user)
}

func (h *UserHandler) HandleLogin(ctx *gin.Context) {
	var loginRequest entities.LoginUserRequest
	err := validation.BindJSON(ctx, &loginRequest, errors.UserInvalidFormat)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	data, err := h.userService.Login(ctx.Request.Context(), loginRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, data)
}

func (h *UserHandler) HandleRefreshToken(ctx *gin.Context) {
	var refreshTokenRequest entities.RefreshTokenRequest
	err := validation.BindJSON(ctx, &refreshTokenRequest, errors.UserInvalidFormat)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	data, err := h.userService.RefreshToken(ctx.Request.Context(), refreshTokenRequest)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, data)
}

func (h *UserHandler) HandleLogout(ctx *gin.Context) {
	var logoutRequest entities.LogoutUserRequest
	err := validation.BindJSON(ctx, &logoutRequest, errors.UserInvalidFormat)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	if err := h.userService.Logout(ctx.Request.Context(), logoutRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, "success_logout")
}

func (h *UserHandler) HandleVerifyEmailOTP(ctx *gin.Context) {
	var verifyRequest entities.VerifyEmailOTPRequest
	err := validation.BindJSON(ctx, &verifyRequest, errors.UserInvalidFormat)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	if err := h.userService.VerifyEmailOTP(ctx.Request.Context(), verifyRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, "success_email_verified")
}

func (h *UserHandler) HandleDelete(ctx *gin.Context) {
	var deleteRequest entities.DeleteUserRequest
	err := validation.BindJSON(ctx, &deleteRequest, errors.UserInvalidFormat)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	if err := h.userService.Delete(ctx.Request.Context(), deleteRequest); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, "success_account_deleted")
}

func (h *UserHandler) HandleUpsertAvatar(ctx *gin.Context) {
	userID := ctx.GetString(middleware.ContextUserIDKey)
	if userID == "" {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid access token",
			Code:       errors.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	fileHeader, err := ctx.FormFile("avatar")
	if err != nil {
		res.Error(ctx, &res.AppError{
			Message:    "Avatar file is required",
			Code:       errors.UserInvalidFormat,
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		res.Error(ctx, err)
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		res.Error(ctx, err)
		return
	}

	data, err := h.userService.UpsertAvatar(ctx.Request.Context(), entities.UpsertAvatarRequest{
		UserID:      userID,
		FileName:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		FileData:    fileData,
	})
	if err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, data)
}

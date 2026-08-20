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
	"github.com/google/uuid"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// HandleRegister godoc
// @Summary      Register user
// @Description  Register a new user account
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body body entities.RegisterUserRequest true "Registration details"
// @Success      201  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      409  {object}  res.APIResponse
// @Router       /api/v1/user/register [post]
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

// HandleLogin godoc
// @Summary      Login
// @Description  Authenticate and receive access/refresh tokens
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body body entities.LoginUserRequest true "Login credentials"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Router       /api/v1/user/login [post]
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

// HandleRefreshToken godoc
// @Summary      Refresh token
// @Description  Get a new access token using a refresh token
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body body entities.RefreshTokenRequest true "Refresh token"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Router       /api/v1/user/refresh-token [post]
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

// HandleLogout godoc
// @Summary      Logout
// @Description  Invalidate the current session tokens
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body body entities.LogoutUserRequest true "Logout request"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/user/logout [post]
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

// HandleVerifyEmailOTP godoc
// @Summary      Verify email OTP
// @Description  Verify email address using OTP code
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body body entities.VerifyEmailOTPRequest true "OTP verification"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Router       /api/v1/user/verify-email-otp [post]
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

// HandleDelete godoc
// @Summary      Delete account
// @Description  Delete the currently authenticated user's account
// @Tags         Users
// @Accept       json
// @Produce      json
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Failure      404  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/user/delete [post]
func (h *UserHandler) HandleDelete(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.GetString(middleware.ContextUserIDKey))
	if err != nil {
		res.Error(ctx, &res.AppError{
			Message:    "Invalid access token",
			Code:       errors.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		})
		return
	}

	if err := h.userService.Delete(ctx.Request.Context(), userID); err != nil {
		res.Error(ctx, err)
		return
	}

	res.Success(ctx, http.StatusOK, "success_account_deleted")
}

// HandleUpsertAvatar godoc
// @Summary      Upload avatar
// @Description  Upload or update user avatar image
// @Tags         Users
// @Accept       multipart/form-data
// @Produce      json
// @Param        avatar formData file true "Avatar image"
// @Success      200  {object}  res.APIResponse
// @Failure      400  {object}  res.APIResponse
// @Failure      401  {object}  res.APIResponse
// @Security     BearerAuth
// @Router       /api/v1/user/avatar [put]
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
	defer func() { _ = file.Close() }()

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

package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/logs"
	"emc_lb/src/pkg/mail"
	"emc_lb/src/pkg/mapping"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/storage"
	"emc_lb/src/pkg/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type UserService interface {
	Register(context.Context, entities.RegisterUserRequest) (entities.RegisterUserResponse, error)
	Login(context.Context, entities.LoginUserRequest) (entities.LoginUserResponse, error)
	RefreshToken(context.Context, entities.RefreshTokenRequest) (entities.LoginUserResponse, error)
	Logout(context.Context, entities.LogoutUserRequest) error
	VerifyEmailOTP(context.Context, entities.VerifyEmailOTPRequest) error
	Delete(context.Context, entities.DeleteUserRequest) error
	UpsertAvatar(context.Context, entities.UpsertAvatarRequest) (entities.UpsertAvatarResponse, error)
}

type userService struct {
	userRepository    repository.UserRepository
	refreshTokenStore cache.RefreshTokenStore
	emailOTPStore     cache.EmailOTPStore
	mailer            mail.Mailer
	avatarStorage     storage.AvatarStorage
}

func NewUserService(userRepository repository.UserRepository, refreshTokenStore cache.RefreshTokenStore, emailOTPStore cache.EmailOTPStore, mailer mail.Mailer, avatarStorage storage.AvatarStorage) UserService {
	return &userService{
		userRepository:    userRepository,
		refreshTokenStore: refreshTokenStore,
		emailOTPStore:     emailOTPStore,
		mailer:            mailer,
		avatarStorage:     avatarStorage,
	}
}

func (s *userService) Register(ctx context.Context, req entities.RegisterUserRequest) (entities.RegisterUserResponse, error) {
	normalizedRequest := req
	utils.NormalizeEmail(&normalizedRequest.Email)
	utils.NormalizeStrings(
		&normalizedRequest.Password,
		&normalizedRequest.UserName,
		&normalizedRequest.Phone,
	)

	_, err := s.userRepository.GetByEmail(ctx, normalizedRequest.Email)
	if err == nil {
		return entities.RegisterUserResponse{}, &res.AppError{
			Message:    "Email already exists",
			Code:       erres.UserAlreadyExists,
			StatusCode: http.StatusConflict,
		}
	}
	// if err != nil && !errors.Is(err, pgx.ErrNoRows) {
	// 	return entities.RegisterUserResponse{}, res.WrapError(err, "Can not get account now", erres.UserGetFailed)
	// }

	passwordHash, err := utils.HashPassword(normalizedRequest.Password)
	if err != nil {
		return entities.RegisterUserResponse{}, res.WrapError(err, "Can not create account now", erres.CommonInternal)
	}

	userEntity := mapping.ToUserEntity(normalizedRequest, passwordHash)
	data, err := s.userRepository.Create(ctx, userEntity)
	if err != nil {
		return entities.RegisterUserResponse{}, res.WrapError(err, "Can not create account now", erres.UserCreateFailed)
	}

	otp, err := utils.GenerateOTP(6)
	if err != nil {
		return entities.RegisterUserResponse{}, res.WrapError(err, "Can not create account now", erres.CommonInternal)
	}

	otpTTL := utils.GetDurationFromEnv("EMAIL_OTP_TTL", 10*time.Minute)
	if err := s.emailOTPStore.Save(ctx, normalizedRequest.Email, otp, otpTTL); err != nil {
		return entities.RegisterUserResponse{}, res.WrapError(err, "Can not create account now", erres.CommonInternal)
	}

	if err := s.mailer.SendEmailVerificationOTP(ctx, normalizedRequest.Email, normalizedRequest.UserName, otp, int(otpTTL.Minutes())); err != nil {
		logs.LogError("mail", "send_verification_otp_failed", err, map[string]any{
			"email":     normalizedRequest.Email,
			"user_name": normalizedRequest.UserName,
		})
	}

	return mapping.ToRegisterUserResponse(data), nil
}

func (s *userService) Login(ctx context.Context, req entities.LoginUserRequest) (entities.LoginUserResponse, error) {
	normalizedRequest := req
	utils.NormalizeEmail(&normalizedRequest.Email)
	utils.NormalizeStrings(&normalizedRequest.Password)

	user, err := s.userRepository.GetByEmail(ctx, normalizedRequest.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.LoginUserResponse{}, &res.AppError{
				Message:    "Invalid email or password",
				Code:       erres.UserUnauthorized,
				StatusCode: http.StatusUnauthorized,
			}
		}

		return entities.LoginUserResponse{}, res.WrapError(err, "Can not get account now", erres.UserGetFailed)
	}

	if err := utils.CheckPassword(normalizedRequest.Password, user.PasswordHash); err != nil {
		return entities.LoginUserResponse{}, &res.AppError{
			Message:    "Invalid email or password",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		}
	}

	if !user.EmailVerified {
		return entities.LoginUserResponse{}, &res.AppError{
			Message:    "Email is not verified",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		}
	}

	tokenPair, err := utils.GenerateTokenPair(user.ID.String(), user.Role)
	if err != nil {
		return entities.LoginUserResponse{}, res.WrapError(err, "Can not login now", erres.CommonInternal)
	}

	if err := s.refreshTokenStore.Save(
		ctx,
		user.ID.String(),
		tokenPair.RefreshToken,
		time.Until(tokenPair.RefreshTokenExpiresAt),
	); err != nil {
		return entities.LoginUserResponse{}, res.WrapError(err, "Can not login now", erres.CommonInternal)
	}

	return entities.LoginUserResponse{
		AccessToken:           tokenPair.AccessToken,
		RefreshToken:          tokenPair.RefreshToken,
		AccessTokenExpiresAt:  tokenPair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenPair.RefreshTokenExpiresAt,
	}, nil
}

func (s *userService) RefreshToken(ctx context.Context, req entities.RefreshTokenRequest) (entities.LoginUserResponse, error) {
	normalizedRequest := req
	utils.NormalizeStrings(&normalizedRequest.RefreshToken)

	userID, err := s.refreshTokenStore.GetUserID(ctx, normalizedRequest.RefreshToken)
	if err != nil {
		return entities.LoginUserResponse{}, &res.AppError{
			Message:    "Invalid refresh token",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		}
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return entities.LoginUserResponse{}, res.WrapError(err, "Invalid user ID in refresh token", erres.CommonInternal)
	}

	user, err := s.userRepository.GetByID(ctx, uid)
	if err != nil {
		return entities.LoginUserResponse{}, res.WrapError(err, "Can not refresh token now", erres.CommonInternal)
	}

	tokenPair, err := utils.GenerateTokenPair(userID, user.Role)
	if err != nil {
		return entities.LoginUserResponse{}, res.WrapError(err, "Can not refresh token now", erres.CommonInternal)
	}

	if err := s.refreshTokenStore.Delete(ctx, normalizedRequest.RefreshToken); err != nil {
		return entities.LoginUserResponse{}, res.WrapError(err, "Can not refresh token now", erres.CommonInternal)
	}

	if err := s.refreshTokenStore.Save(
		ctx,
		userID,
		tokenPair.RefreshToken,
		time.Until(tokenPair.RefreshTokenExpiresAt),
	); err != nil {
		return entities.LoginUserResponse{}, res.WrapError(err, "Can not refresh token now", erres.CommonInternal)
	}

	return entities.LoginUserResponse{
		AccessToken:           tokenPair.AccessToken,
		RefreshToken:          tokenPair.RefreshToken,
		AccessTokenExpiresAt:  tokenPair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenPair.RefreshTokenExpiresAt,
	}, nil
}

func (s *userService) Logout(ctx context.Context, req entities.LogoutUserRequest) error {
	normalizedRequest := req
	utils.NormalizeStrings(&normalizedRequest.RefreshToken)

	if _, err := s.refreshTokenStore.GetUserID(ctx, normalizedRequest.RefreshToken); err != nil {
		return &res.AppError{
			Message:    "Invalid refresh token",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		}
	}

	if err := s.refreshTokenStore.Delete(ctx, normalizedRequest.RefreshToken); err != nil {
		return res.WrapError(err, "Can not logout now", erres.CommonInternal)
	}

	return nil
}

func (s *userService) VerifyEmailOTP(ctx context.Context, req entities.VerifyEmailOTPRequest) error {
	normalizedRequest := req
	utils.NormalizeEmail(&normalizedRequest.Email)
	utils.NormalizeStrings(&normalizedRequest.OTP)

	savedOTP, err := s.emailOTPStore.Get(ctx, normalizedRequest.Email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return &res.AppError{
				Message:    "Invalid or expired otp",
				Code:       erres.UserUnauthorized,
				StatusCode: http.StatusUnauthorized,
			}
		}

		return res.WrapError(err, "Can not verify email now", erres.CommonInternal)
	}

	if savedOTP != normalizedRequest.OTP {
		return &res.AppError{
			Message:    "Invalid or expired otp",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		}
	}

	if err := s.emailOTPStore.Delete(ctx, normalizedRequest.Email); err != nil {
		return res.WrapError(err, "Can not verify email now", erres.CommonInternal)
	}

	if err := s.userRepository.VerifyEmail(ctx, normalizedRequest.Email); err != nil {
		return res.WrapError(err, "Can not verify email now", erres.UserUpdateFailed)
	}

	return nil
}

func (s *userService) Delete(ctx context.Context, req entities.DeleteUserRequest) error {
	normalizedRequest := req
	utils.NormalizeEmail(&normalizedRequest.Email)

	if _, err := s.userRepository.GetByEmail(ctx, normalizedRequest.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &res.AppError{
				Message:    "User not found",
				Code:       erres.UserNotFound,
				StatusCode: http.StatusNotFound,
			}
		}

		return res.WrapError(err, "Can not get account now", erres.UserGetFailed)
	}

	if err := s.userRepository.SoftDeleteByEmail(ctx, normalizedRequest.Email); err != nil {
		return res.WrapError(err, "Can not delete account now", erres.UserUpdateFailed)
	}

	return nil
}

func (s *userService) UpsertAvatar(ctx context.Context, req entities.UpsertAvatarRequest) (entities.UpsertAvatarResponse, error) {
	normalizedRequest := req
	utils.NormalizeStrings(&normalizedRequest.UserID, &normalizedRequest.FileName, &normalizedRequest.ContentType)

	if normalizedRequest.UserID == "" {
		return entities.UpsertAvatarResponse{}, &res.AppError{
			Message:    "User id is required",
			Code:       erres.UserInvalidFormat,
			StatusCode: http.StatusBadRequest,
		}
	}

	userID, err := uuid.Parse(normalizedRequest.UserID)
	if err != nil {
		return entities.UpsertAvatarResponse{}, &res.AppError{
			Message:    "Invalid user id",
			Code:       erres.UserInvalidFormat,
			StatusCode: http.StatusBadRequest,
		}
	}

	if _, err := s.userRepository.GetIDByID(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.UpsertAvatarResponse{}, &res.AppError{
				Message:    "User not found",
				Code:       erres.UserNotFound,
				StatusCode: http.StatusNotFound,
			}
		}

		return entities.UpsertAvatarResponse{}, res.WrapError(err, "Can not get account now", erres.UserGetFailed)
	}

	if len(normalizedRequest.FileData) == 0 {
		return entities.UpsertAvatarResponse{}, &res.AppError{
			Message:    "Avatar file is required",
			Code:       erres.UserInvalidFormat,
			StatusCode: http.StatusBadRequest,
		}
	}

	if len(normalizedRequest.FileData) > 5*1024*1024 { // 5MB limit
		return entities.UpsertAvatarResponse{}, &res.AppError{
			Message:    "Avatar file is too large (max 5MB)",
			Code:       erres.UserInvalidFormat,
			StatusCode: http.StatusBadRequest,
		}
	}

	mimeType := http.DetectContentType(normalizedRequest.FileData)
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
		return entities.UpsertAvatarResponse{}, &res.AppError{
			Message:    "Invalid file type. Only JPEG, PNG, and WebP are allowed",
			Code:       erres.UserInvalidFormat,
			StatusCode: http.StatusBadRequest,
		}
	}

	avatarURL, err := s.avatarStorage.UploadAvatar(
		ctx,
		normalizedRequest.UserID,
		normalizedRequest.FileName,
		normalizedRequest.FileData,
		normalizedRequest.ContentType,
	)
	if err != nil {
		return entities.UpsertAvatarResponse{}, res.WrapError(err, "Can not upload avatar now", erres.CommonInternal)
	}

	if err := s.userRepository.UpdateAvatarByID(ctx, userID, avatarURL); err != nil {
		return entities.UpsertAvatarResponse{}, res.WrapError(err, "Can not update avatar now", erres.UserUpdateFailed)
	}

	return entities.UpsertAvatarResponse{
		AvatarURL: avatarURL,
	}, nil
}

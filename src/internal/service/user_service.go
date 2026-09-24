package service

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"time"

	"emc_lb/src/internal/db/sqlc"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/config"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/logs"
	"emc_lb/src/pkg/mapping"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/storage"
	"emc_lb/src/pkg/utils"
	"emc_lb/src/pkg/worker"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type UserService interface {
	Register(context.Context, entities.RegisterUserRequest) (entities.RegisterUserResponse, error)
	Login(context.Context, entities.LoginUserRequest) (entities.LoginUserResponse, error)
	RefreshToken(context.Context, entities.RefreshTokenRequest) (entities.LoginUserResponse, error)
	Logout(context.Context, entities.LogoutUserRequest) error
	VerifyEmailOTP(context.Context, entities.VerifyEmailOTPRequest) error
	Delete(context.Context, uuid.UUID) error
	UpsertAvatar(context.Context, entities.UpsertAvatarRequest) (entities.UpsertAvatarResponse, error)
	GetProfile(context.Context, uuid.UUID) (entities.UserProfileResponse, error)
	ChangePassword(context.Context, uuid.UUID, entities.ChangePasswordRequest) error
	AdminChangePassword(context.Context, uuid.UUID, entities.AdminChangePasswordRequest) error
	ForgotPassword(context.Context, entities.ForgotPasswordRequest) error
	ResetPassword(context.Context, entities.ResetPasswordRequest) error
}

type userService struct {
	cfg               *config.AppConfig
	pool              *pgxpool.Pool
	userRepository    repository.UserRepository
	refreshTokenStore cache.RefreshTokenStore
	emailOTPStore     cache.EmailOTPStore
	taskDistributor   worker.TaskDistributor
	avatarStorage     storage.AvatarStorage
}

func NewUserService(cfg *config.AppConfig, pool *pgxpool.Pool, userRepository repository.UserRepository, refreshTokenStore cache.RefreshTokenStore, emailOTPStore cache.EmailOTPStore, taskDistributor worker.TaskDistributor, avatarStorage storage.AvatarStorage) UserService {
	return &userService{
		cfg:               cfg,
		pool:              pool,
		userRepository:    userRepository,
		refreshTokenStore: refreshTokenStore,
		emailOTPStore:     emailOTPStore,
		taskDistributor:   taskDistributor,
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
	if !errors.Is(err, pgx.ErrNoRows) {
		return entities.RegisterUserResponse{}, res.WrapError(err, "Cannot check email availability", erres.CommonInternal)
	}

	passwordHash, err := utils.HashPassword(normalizedRequest.Password, s.cfg.App.SystemSecret)
	if err != nil {
		return entities.RegisterUserResponse{}, res.WrapError(err, "Can not create account now", erres.CommonInternal)
	}

	userEntity, userProfile := mapping.ToUserEntity(normalizedRequest, passwordHash)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return entities.RegisterUserResponse{}, res.WrapError(err, "Can not start transaction now", erres.CommonInternal)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	repoTx := s.userRepository.WithTx(tx)

	data, err := repoTx.Create(ctx, userEntity, userProfile)
	if err != nil {
		return entities.RegisterUserResponse{}, res.WrapError(err, "Can not create account now", erres.UserCreateFailed)
	}

	if err := tx.Commit(ctx); err != nil {
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

	payload := &worker.PayloadSendVerifyEmail{
		Email:    normalizedRequest.Email,
		UserName: normalizedRequest.UserName,
		OTP:      otp,
		TTL:      int(otpTTL.Minutes()),
	}

	if err := s.taskDistributor.DistributeTaskSendVerifyEmail(ctx, payload); err != nil {
		logs.LogError("worker", "enqueue_verify_email_task_failed", err, map[string]any{
			"email":     normalizedRequest.Email,
			"user_name": normalizedRequest.UserName,
		})
	}

	return mapping.ToRegisterUserResponse(data, userProfile), nil
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

	if user.IsBanned {
		return entities.LoginUserResponse{}, &res.AppError{
			Message:    "Account is banned",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusForbidden,
		}
	}

	if user.Status != "active" {
		return entities.LoginUserResponse{}, &res.AppError{
			Message:    "Account is inactive",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusForbidden,
		}
	}

	if time.Now().Before(user.LockedUntil) {
		return entities.LoginUserResponse{}, &res.AppError{
			Message:    "Account is temporarily locked. Please try again later.",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusLocked,
		}
	}

	if err = utils.CheckPassword(normalizedRequest.Password, user.PasswordHash, s.cfg.App.SystemSecret); err != nil {
		_ = s.userRepository.UpdateFailedLoginAttempts(ctx, user.ID)
		if user.FailedLoginAttempts+1 >= 5 {
			_ = s.userRepository.LockUserAccount(ctx, sqlc.LockUserAccountParams{
				ID:          user.ID,
				LockedUntil: time.Now().Add(15 * time.Minute),
			})
		}
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

	var lastLoginIPStr *string
	if req.ClientIP != "" {
		if _, err := netip.ParseAddr(req.ClientIP); err == nil {
			lastLoginIPStr = &req.ClientIP
		}
	}
	_ = s.userRepository.UpdateUserLoginStats(ctx, sqlc.UpdateUserLoginStatsParams{
		ID:          user.ID,
		LastLoginIp: lastLoginIPStr,
	})

	tokenPair, err := utils.GenerateTokenPair(user.UUID.String(), user.Role, s.cfg.JWT)
	if err != nil {
		return entities.LoginUserResponse{}, res.WrapError(err, "Can not login now", erres.CommonInternal)
	}

	if err := s.refreshTokenStore.Save(
		ctx,
		user.UUID.String(),
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

	user, err := s.userRepository.GetByUUID(ctx, uid)
	if err != nil {
		return entities.LoginUserResponse{}, res.WrapError(err, "Can not refresh token now", erres.CommonInternal)
	}

	tokenPair, err := utils.GenerateTokenPair(userID, user.Role, s.cfg.JWT)
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

func (s *userService) Delete(ctx context.Context, userID uuid.UUID) error {
	if _, err := s.userRepository.GetByUUID(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &res.AppError{
				Message:    "User not found",
				Code:       erres.UserNotFound,
				StatusCode: http.StatusNotFound,
			}
		}

		return res.WrapError(err, "Can not get account now", erres.UserGetFailed)
	}

	if err := s.userRepository.SoftDeleteByUUID(ctx, userID); err != nil {
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

	if _, err = s.userRepository.GetIDByUUID(ctx, userID); err != nil {
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

	if err := s.userRepository.UpdateAvatarByUUID(ctx, userID, avatarURL); err != nil {
		return entities.UpsertAvatarResponse{}, res.WrapError(err, "Can not update avatar now", erres.UserUpdateFailed)
	}

	return entities.UpsertAvatarResponse{
		AvatarURL: avatarURL,
	}, nil
}

func (s *userService) GetProfile(ctx context.Context, userID uuid.UUID) (entities.UserProfileResponse, error) {
	row, err := s.userRepository.GetUserProfileByUUID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entities.UserProfileResponse{}, &res.AppError{
				Message:    "User not found",
				Code:       erres.UserNotFound,
				StatusCode: http.StatusNotFound,
			}
		}
		return entities.UserProfileResponse{}, res.WrapError(err, "Can not get profile now", erres.UserGetFailed)
	}

	var avatarURL string
	if row.AvatarUrl != nil {
		avatarURL = *row.AvatarUrl
	}

	var phone string
	if row.Phone != nil {
		phone = *row.Phone
	}

	var fullName string
	if row.FullName != nil {
		fullName = *row.FullName
	}

	var userName string
	if row.UserName != nil {
		userName = *row.UserName
	}

	return entities.UserProfileResponse{
		ID:            row.Uuid,
		Email:         row.Email,
		Role:          row.Role,
		Status:        row.Status,
		EmailVerified: row.EmailVerified,
		CreatedAt:     row.CreatedAt,
		Profile: entities.ProfileData{
			UserName:  userName,
			FullName:  fullName,
			Phone:     phone,
			AvatarURL: avatarURL,
		},
	}, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID uuid.UUID, req entities.ChangePasswordRequest) error {
	normalizedRequest := req
	utils.NormalizeStrings(&normalizedRequest.OldPassword, &normalizedRequest.NewPassword)

	user, err := s.userRepository.GetByUUID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &res.AppError{
				Message:    "User not found",
				Code:       erres.UserNotFound,
				StatusCode: http.StatusNotFound,
			}
		}
		return res.WrapError(err, "Can not change password now", erres.UserGetFailed)
	}

	if err = utils.CheckPassword(normalizedRequest.OldPassword, user.PasswordHash, s.cfg.App.SystemSecret); err != nil {
		return &res.AppError{
			Message:    "Invalid old password",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusBadRequest,
		}
	}

	newPasswordHash, err := utils.HashPassword(normalizedRequest.NewPassword, s.cfg.App.SystemSecret)
	if err != nil {
		return res.WrapError(err, "Can not change password now", erres.CommonInternal)
	}

	if err := s.userRepository.UpdatePassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: newPasswordHash,
	}); err != nil {
		return res.WrapError(err, "Can not change password now", erres.UserUpdateFailed)
	}

	// Optionally invalidate all refresh tokens for this user here to force re-login on all devices
	// _ = s.refreshTokenStore.DeleteAllForUser(ctx, userID.String())

	return nil
}

func (s *userService) AdminChangePassword(ctx context.Context, targetUserUUID uuid.UUID, req entities.AdminChangePasswordRequest) error {
	normalizedRequest := req
	utils.NormalizeStrings(&normalizedRequest.NewPassword)

	user, err := s.userRepository.GetByUUID(ctx, targetUserUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &res.AppError{
				Message:    "User not found",
				Code:       erres.UserNotFound,
				StatusCode: http.StatusNotFound,
			}
		}
		return res.WrapError(err, "Can not change password now", erres.UserGetFailed)
	}

	newPasswordHash, err := utils.HashPassword(normalizedRequest.NewPassword, s.cfg.App.SystemSecret)
	if err != nil {
		return res.WrapError(err, "Can not change password now", erres.CommonInternal)
	}

	if err := s.userRepository.UpdatePassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: newPasswordHash,
	}); err != nil {
		return res.WrapError(err, "Can not change password now", erres.UserUpdateFailed)
	}

	return nil
}

func (s *userService) ForgotPassword(ctx context.Context, req entities.ForgotPasswordRequest) error {
	normalizedRequest := req
	utils.NormalizeEmail(&normalizedRequest.Email)

	user, err := s.userRepository.GetByEmail(ctx, normalizedRequest.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Do not leak if user exists
			return nil
		}
		return res.WrapError(err, "Can not process forgot password now", erres.UserGetFailed)
	}

	otp, err := utils.GenerateOTP(6)
	if err != nil {
		return res.WrapError(err, "Can not generate OTP now", erres.CommonInternal)
	}

	otpTTL := utils.GetDurationFromEnv("EMAIL_OTP_TTL", 10*time.Minute)
	// We prefix the email so it doesn't conflict with registration OTP
	resetKey := "pwd_reset:" + normalizedRequest.Email
	if err := s.emailOTPStore.Save(ctx, resetKey, otp, otpTTL); err != nil {
		return res.WrapError(err, "Can not save OTP now", erres.CommonInternal)
	}

	profile, _ := s.userRepository.GetUserProfileByUUID(ctx, user.UUID)
	userName := normalizedRequest.Email
	if profile.UserName != nil {
		userName = *profile.UserName
	}

	payload := &worker.PayloadSendPasswordResetEmail{
		Email:    normalizedRequest.Email,
		UserName: userName,
		OTP:      otp,
		TTL:      int(otpTTL.Minutes()),
	}

	if err := s.taskDistributor.DistributeTaskSendPasswordResetEmail(ctx, payload); err != nil {
		logs.LogError("worker", "enqueue_password_reset_email_task_failed", err, map[string]any{
			"email": normalizedRequest.Email,
		})
	}

	return nil
}

func (s *userService) ResetPassword(ctx context.Context, req entities.ResetPasswordRequest) error {
	normalizedRequest := req
	utils.NormalizeEmail(&normalizedRequest.Email)
	utils.NormalizeStrings(&normalizedRequest.OTP, &normalizedRequest.NewPassword)

	resetKey := "pwd_reset:" + normalizedRequest.Email
	savedOTP, err := s.emailOTPStore.Get(ctx, resetKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return &res.AppError{
				Message:    "Invalid or expired otp",
				Code:       erres.UserUnauthorized,
				StatusCode: http.StatusUnauthorized,
			}
		}
		return res.WrapError(err, "Can not verify OTP now", erres.CommonInternal)
	}

	if savedOTP != normalizedRequest.OTP {
		return &res.AppError{
			Message:    "Invalid or expired otp",
			Code:       erres.UserUnauthorized,
			StatusCode: http.StatusUnauthorized,
		}
	}

	user, err := s.userRepository.GetByEmail(ctx, normalizedRequest.Email)
	if err != nil {
		return res.WrapError(err, "Can not get user account", erres.UserGetFailed)
	}

	newPasswordHash, err := utils.HashPassword(normalizedRequest.NewPassword, s.cfg.App.SystemSecret)
	if err != nil {
		return res.WrapError(err, "Can not reset password now", erres.CommonInternal)
	}

	if err := s.userRepository.UpdatePassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: newPasswordHash,
	}); err != nil {
		return res.WrapError(err, "Can not reset password now", erres.UserUpdateFailed)
	}

	_ = s.emailOTPStore.Delete(ctx, resetKey)
	return nil
}

package service_test

import (
	"context"
	"errors"
	"testing"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/entities"
	errs "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mock_cache "emc_lb/src/tests/mocks/cache"
	mock_repository "emc_lb/src/tests/mocks/repository"
	mock_storage "emc_lb/src/tests/mocks/storage"
	mock_worker "emc_lb/src/tests/mocks/worker"
)

// newUserService creates a userService wired to mock dependencies.
func newUserService(
	t *testing.T,
	repo *mock_repository.MockUserRepository,
	tokenStore *mock_cache.MockRefreshTokenStore,
	otpStore *mock_cache.MockEmailOTPStore,
	distributor *mock_worker.MockTaskDistributor,
	avatar *mock_storage.MockAvatarStorage,
) service.UserService {
	t.Helper()
	return service.NewUserService(repo, tokenStore, otpStore, distributor, avatar)
}

// ---- Register ----

func TestUserService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock_repository.NewMockUserRepository(ctrl)
	tokenStore := mock_cache.NewMockRefreshTokenStore(ctrl)
	otpStore := mock_cache.NewMockEmailOTPStore(ctrl)
	distributor := mock_worker.NewMockTaskDistributor(ctrl)
	avatar := mock_storage.NewMockAvatarStorage(ctrl)

	ctx := context.Background()
	req := entities.RegisterUserRequest{
		Email:    "test@example.com",
		Password: "StrongPass1!",
		UserName: "testuser",
		Phone:    "0901234567",
	}

	createdUser := entities.User{
		ID:       uuid.New(),
		Email:    req.Email,
		UserName: req.UserName,
	}

	// Arrange
	repo.EXPECT().
		GetByEmail(ctx, req.Email).
		Return(entities.User{}, errors.New("not found"))

	repo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(createdUser, nil)

	otpStore.EXPECT().
		Save(ctx, req.Email, gomock.Any(), gomock.Any()).
		Return(nil)

	distributor.EXPECT().
		DistributeTaskSendVerifyEmail(ctx, gomock.Any()).
		Return(nil)

	svc := newUserService(t, repo, tokenStore, otpStore, distributor, avatar)

	// Act
	resp, err := svc.Register(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, createdUser.ID, resp.ID)
	assert.Equal(t, req.Email, resp.Email)
}

func TestUserService_Register_EmailAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock_repository.NewMockUserRepository(ctrl)
	tokenStore := mock_cache.NewMockRefreshTokenStore(ctrl)
	otpStore := mock_cache.NewMockEmailOTPStore(ctrl)
	distributor := mock_worker.NewMockTaskDistributor(ctrl)
	avatar := mock_storage.NewMockAvatarStorage(ctrl)

	ctx := context.Background()
	req := entities.RegisterUserRequest{
		Email:    "existing@example.com",
		Password: "StrongPass1!",
		UserName: "testuser",
	}

	// Arrange: email already exists
	repo.EXPECT().
		GetByEmail(ctx, req.Email).
		Return(entities.User{ID: uuid.New(), Email: req.Email}, nil)

	svc := newUserService(t, repo, tokenStore, otpStore, distributor, avatar)

	// Act
	_, err := svc.Register(ctx, req)

	// Assert
	var appErr *res.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, errs.UserAlreadyExists, appErr.Code)
}

// ---- Login ----

func TestUserService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock_repository.NewMockUserRepository(ctrl)
	tokenStore := mock_cache.NewMockRefreshTokenStore(ctrl)
	otpStore := mock_cache.NewMockEmailOTPStore(ctrl)
	distributor := mock_worker.NewMockTaskDistributor(ctrl)
	avatar := mock_storage.NewMockAvatarStorage(ctrl)

	ctx := context.Background()
	req := entities.LoginUserRequest{
		Email:    "nouser@example.com",
		Password: "SomePass1!",
	}

	// Arrange: pgx.ErrNoRows returned
	repo.EXPECT().
		GetByEmail(ctx, req.Email).
		Return(entities.User{}, errors.New("no rows"))

	svc := newUserService(t, repo, tokenStore, otpStore, distributor, avatar)

	// Act
	_, err := svc.Login(ctx, req)

	// Assert: must be an AppError
	var appErr *res.AppError
	require.ErrorAs(t, err, &appErr)
}

// ---- Logout ----

func TestUserService_Logout_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mock_repository.NewMockUserRepository(ctrl)
	tokenStore := mock_cache.NewMockRefreshTokenStore(ctrl)
	otpStore := mock_cache.NewMockEmailOTPStore(ctrl)
	distributor := mock_worker.NewMockTaskDistributor(ctrl)
	avatar := mock_storage.NewMockAvatarStorage(ctrl)

	ctx := context.Background()
	req := entities.LogoutUserRequest{RefreshToken: "invalid-token"}

	// Arrange: token not found in store
	tokenStore.EXPECT().
		GetUserID(ctx, req.RefreshToken).
		Return("", errors.New("not found"))

	svc := newUserService(t, repo, tokenStore, otpStore, distributor, avatar)

	// Act
	err := svc.Logout(ctx, req)

	// Assert
	var appErr *res.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, errs.UserUnauthorized, appErr.Code)
}

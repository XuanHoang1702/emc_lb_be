package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/pkg/entities"
	"emc_lb/src/pkg/mapping"
)

type UserRepository interface {
	WithTx(tx pgx.Tx) UserRepository
	GetByEmail(context.Context, string) (entities.User, error)
	GetByUUID(context.Context, uuid.UUID) (entities.User, error)
	GetIDByUUID(context.Context, uuid.UUID) (int64, error)
	Create(context.Context, entities.User, entities.UserProfile) (entities.User, error)
	VerifyEmail(context.Context, string) error
	SoftDeleteByUUID(context.Context, uuid.UUID) error
	UpdateAvatarByUUID(context.Context, uuid.UUID, string) error
	UpdateFailedLoginAttempts(context.Context, int64) error
	LockUserAccount(context.Context, sqlc.LockUserAccountParams) error
	UpdateUserLoginStats(context.Context, sqlc.UpdateUserLoginStatsParams) error
	GetUserProfileByUUID(context.Context, uuid.UUID) (sqlc.GetUserProfileByUUIDRow, error)
	UpdatePassword(context.Context, sqlc.UpdateUserPasswordParams) error
}

type userRepository struct {
	pool *pgxpool.Pool
	db   *sqlc.Queries
}

func NewUserRepository(pool *pgxpool.Pool, db sqlc.Querier) UserRepository {
	return &userRepository{
		pool: pool,
		db:   db.(*sqlc.Queries),
	}
}

func (r *userRepository) WithTx(tx pgx.Tx) UserRepository {
	return &userRepository{
		pool: r.pool,
		db:   r.db.WithTx(tx),
	}
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (entities.User, error) {
	row, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		return entities.User{}, err
	}
	return entities.User{
		ID:                  row.ID,
		UUID:                row.Uuid,
		Email:               row.Email,
		PasswordHash:        row.PasswordHash,
		EmailVerified:       row.EmailVerified,
		Role:                row.Role,
		Status:              row.Status,
		IsBanned:            row.IsBanned,
		LockedUntil:         row.LockedUntil,
		FailedLoginAttempts: row.FailedLoginAttempts,
	}, nil
}

func (r *userRepository) GetByUUID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	row, err := r.db.GetUserByUUID(ctx, id)
	if err != nil {
		return entities.User{}, err
	}
	return entities.User{
		ID:            row.ID,
		UUID:          row.Uuid,
		Email:         row.Email,
		PasswordHash:  row.PasswordHash,
		EmailVerified: row.EmailVerified,
		Role:          row.Role,
	}, nil
}

func (r *userRepository) GetIDByUUID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.db.GetUserIDByUUID(ctx, userID)
}

func (r *userRepository) Create(ctx context.Context, user entities.User, profile entities.UserProfile) (entities.User, error) {
	params := sqlc.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}
	
	row, err := r.db.CreateUser(ctx, params)
	if err != nil {
		return entities.User{}, err
	}

	profileParams := sqlc.CreateUserProfileParams{
		UserID:   row.ID,
		UserName: &profile.UserName,
		Phone:    profile.Phone,
	}

	if err := r.db.CreateUserProfile(ctx, profileParams); err != nil {
		return entities.User{}, err
	}

	if err := r.db.CreateCustomerStats(ctx, row.ID); err != nil {
		return entities.User{}, err
	}

	return mapping.ToUserEntityFromSqlc(row), nil
}

func (r *userRepository) VerifyEmail(ctx context.Context, email string) error {
	return r.db.VerifyUserEmail(ctx, email)
}

func (r *userRepository) SoftDeleteByUUID(ctx context.Context, id uuid.UUID) error {
	return r.db.SoftDeleteUserByUUID(ctx, id)
}

func (r *userRepository) UpdateAvatarByUUID(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	return r.db.UpdateUserAvatarByUUID(ctx, sqlc.UpdateUserAvatarByUUIDParams{
		Uuid:      userID,
		AvatarUrl: &avatarURL,
	})
}

func (r *userRepository) UpdateFailedLoginAttempts(ctx context.Context, id int64) error {
	return r.db.UpdateFailedLoginAttempts(ctx, id)
}

func (r *userRepository) LockUserAccount(ctx context.Context, arg sqlc.LockUserAccountParams) error {
	return r.db.LockUserAccount(ctx, arg)
}

func (r *userRepository) UpdateUserLoginStats(ctx context.Context, arg sqlc.UpdateUserLoginStatsParams) error {
	return r.db.UpdateUserLoginStats(ctx, arg)
}

func (r *userRepository) GetUserProfileByUUID(ctx context.Context, id uuid.UUID) (sqlc.GetUserProfileByUUIDRow, error) {
	return r.db.GetUserProfileByUUID(ctx, id)
}

func (r *userRepository) UpdatePassword(ctx context.Context, params sqlc.UpdateUserPasswordParams) error {
	return r.db.UpdateUserPassword(ctx, params)
}


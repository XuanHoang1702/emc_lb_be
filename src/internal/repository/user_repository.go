package repository

import (
	"context"

	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/pkg/entities"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetByEmail(context.Context, string) (entities.User, error)
	GetByID(context.Context, uuid.UUID) (entities.User, error)
	GetIDByID(context.Context, uuid.UUID) (uuid.UUID, error)
	Create(context.Context, entities.User) (entities.User, error)
	VerifyEmail(context.Context, string) error
	SoftDeleteByEmail(context.Context, string) error
	UpdateAvatarByID(context.Context, uuid.UUID, string) error
}

type userRepository struct {
	db sqlc.Querier
}

func NewUserRepository(db sqlc.Querier) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (entities.User, error) {
	row, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		return entities.User{}, err
	}
	return entities.User{
		ID:            row.ID,
		Email:         row.Email,
		PasswordHash:  row.PasswordHash,
		EmailVerified: row.EmailVerified,
		Role:          row.Role,
	}, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (entities.User, error) {
	row, err := r.db.GetUserByID(ctx, id)
	if err != nil {
		return entities.User{}, err
	}
	return entities.User{
		ID:            row.ID,
		Email:         row.Email,
		PasswordHash:  row.PasswordHash,
		EmailVerified: row.EmailVerified,
		Role:          row.Role,
	}, nil
}

func (r *userRepository) GetIDByID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return r.db.GetUserIDByID(ctx, userID)
}

func (r *userRepository) Create(ctx context.Context, user entities.User) (entities.User, error) {
	params := sqlc.CreateUserParams{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		UserName:     user.UserName,
		Phone:        user.Phone,
	}
	row, err := r.db.CreateUser(ctx, params)
	if err != nil {
		return entities.User{}, err
	}

	user.ID = row.ID
	user.Email = row.Email
	user.UserName = row.UserName
	user.Phone = row.Phone
	user.CreatedAt = row.CreatedAt

	return user, nil
}

func (r *userRepository) VerifyEmail(ctx context.Context, email string) error {
	return r.db.VerifyUserEmail(ctx, email)
}

func (r *userRepository) SoftDeleteByEmail(ctx context.Context, email string) error {
	return r.db.SoftDeleteUserByEmail(ctx, email)
}

func (r *userRepository) UpdateAvatarByID(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	return r.db.UpdateUserAvatarByID(ctx, sqlc.UpdateUserAvatarByIDParams{
		ID:        userID,
		AvatarUrl: &avatarURL,
	})
}

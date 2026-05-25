package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                  uuid.UUID
	Email               string
	Phone               *string
	FullName            *string
	UserName            string
	PasswordHash        string
	PasswordChangedAt   time.Time
	EmailVerified       bool
	EmailVerifiedAt     time.Time
	PhoneVerified       bool
	PhoneVerifiedAt     time.Time
	AvatarUrl           *string
	Gender              *string
	BirthDate           time.Time
	Status              string
	IsBanned            bool
	BannedReason        *string
	Role                string
	TotalOrders         int32
	TotalSpent          float64
	RewardPoints        int64
	LastLoginAt         time.Time
	LastLoginIp         *string
	FailedLoginAttempts int32
	LockedUntil         time.Time
	LanguageCode        *string
	Timezone            *string
	IsDeleted           bool
	DeletedAt           time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	UserName string `json:"user_name" binding:"required"`
	Phone    string `json:"phone"`
}

type RegisterUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	UserName  string    `json:"user_name,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginUserResponse struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutUserRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type VerifyEmailOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type DeleteUserRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type UpsertAvatarRequest struct {
	UserID      string
	FileName    string
	ContentType string
	FileData    []byte
}

type UpsertAvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
}

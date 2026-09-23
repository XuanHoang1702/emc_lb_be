package entities

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// User & UserProfile Entities
// ============================================================

type User struct {
	ID                  int64
	UUID                uuid.UUID
	Email               string
	PasswordHash        string
	PasswordChangedAt   time.Time
	EmailVerified       bool
	EmailVerifiedAt     time.Time
	Status              string
	IsBanned            bool
	BannedReason        *string
	Role                string
	LastLoginAt         time.Time
	LastLoginIP         *string
	FailedLoginAttempts int32
	LockedUntil         time.Time
	IsDeleted           bool
	DeletedAt           time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type UserProfile struct {
	UserID          int64
	UserName        string
	FullName        *string
	Phone           *string
	PhoneVerified   bool
	PhoneVerifiedAt time.Time
	AvatarURL       *string
	Gender          *string
	BirthDate       time.Time
	LanguageCode    *string
	Timezone        *string
	TotalOrders     int32
	TotalSpent      float64
	RewardPoints    int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ============================================================
// Auth: Register
// ============================================================

type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,password_strong"`
	UserName string `json:"user_name" binding:"required,min=3,max=254,regex=^[a-zA-Z0-9_]*$"`
	Phone    string `json:"phone" binding:"omitempty,min=10,max=15"`
}

type RegisterUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	UserName  string    `json:"user_name,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ============================================================
// Auth: Login / Token / Logout
// ============================================================

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	ClientIP string `json:"-"`
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

// ============================================================
// Verification
// ============================================================

type VerifyEmailOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

// ============================================================
// Avatar
// ============================================================

type UpsertAvatarRequest struct {
	UserID      string
	FileName    string
	ContentType string
	FileData    []byte
}

type UpsertAvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
}

// ============================================================
// Profile
// ============================================================

type ProfileData struct {
	UserName  string `json:"user_name,omitempty"`
	FullName  string `json:"full_name,omitempty"`
	Phone     string `json:"phone,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UserProfileResponse struct {
	ID            uuid.UUID   `json:"id"`
	Email         string      `json:"email"`
	Role          string      `json:"role"`
	Status        string      `json:"status"`
	EmailVerified bool        `json:"email_verified"`
	Profile       ProfileData `json:"profile"`
	CreatedAt     time.Time   `json:"created_at"`
}

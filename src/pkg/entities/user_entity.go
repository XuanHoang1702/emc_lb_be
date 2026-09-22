package entities

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// User Entity
// ============================================================

type User struct {
	// --- Identity ---
	ID       uuid.UUID
	Email    string
	Phone    *string
	UserName string
	FullName *string

	// --- Security ---
	PasswordHash        string
	PasswordChangedAt   time.Time
	FailedLoginAttempts int32
	LockedUntil         time.Time
	LastLoginAt         time.Time
	LastLoginIP         *string

	// --- Verification ---
	EmailVerified   bool
	EmailVerifiedAt time.Time
	PhoneVerified   bool
	PhoneVerifiedAt time.Time

	// --- Profile ---
	AvatarURL    *string
	Gender       *string
	BirthDate    time.Time
	LanguageCode *string
	Timezone     *string

	// --- Account status ---
	Status       string
	IsBanned     bool
	BannedReason *string
	Role         string

	// --- Stats ---
	TotalOrders  int32
	TotalSpent   float64
	RewardPoints int64

	// --- Audit ---
	IsDeleted bool
	DeletedAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
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

// Giới hạn độ dài tối đa của cột có thể ảnh hưởng đến khả năng lập index, giới hạn index key
// và một số quyết định lưu trữ; kích thước index thực tế phụ thuộc vào dữ liệu thực tế và database engine.

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

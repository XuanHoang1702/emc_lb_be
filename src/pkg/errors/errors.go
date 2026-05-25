package errors

import (
	"emc_lb/src/pkg/res"
)

type Code = res.ErrorCode

const (

	// COMOM ERROR
	CommonBadRequest      Code = res.ErrCodeBadRequest
	CommonNotFound        Code = res.ErrCodeNotFound
	CommonConflict        Code = res.ErrCodeConflict
	CommonInternal        Code = res.ErrCodeInternal
	CommonUnauthorized    Code = res.ErrCodeUnauthorized
	CommonForbidden       Code = res.ErrCodeForbidden
	CommonTooManyRequests Code = res.ErrCodeTooManyRequests

	// USER ERROR
	UserValidationFailed     Code = "USER_VALIDATION_FAILED"
	UserNotFound             Code = "USER_NOT_FOUND"
	UserAlreadyExists        Code = "USER_ALREADY_EXISTS"
	UserCreateFailed         Code = "USER_CREATE_FAILED"
	UserGetFailed            Code = "USER_GET_FAILED"
	UserUpdateFailed         Code = "USER_UPDATE_FAILED"
	UserStatusChangeFailed   Code = "USER_STATUS_CHANGE_FAILED"
	UserPasswordUpdateFailed Code = "USER_PASSWORD_UPDATE_FAILED"
	UserUnauthorized         Code = "USER_UNAUTHORIZED"
	UserInvalidCreatorId     Code = "USER_INVALID_CREATOR_ID"
	UserRoleUpdateForbidden  Code = "USER_ROLE_UPDATE_FORBIDDEN"
	UserAgentCreateForbidden Code = "USER_AGENT_CREATE_FORBIDDEN"
	UserInvalidAvatarUpload  Code = "USER_INVALID_AVATAR_UPLOAD"
	UserAvatarProcessFailed  Code = "USER_AVATAR_PROCESS_FAILED"
	UserInvalidFormat        Code = "USER_INVALID_FORMAT"

	// PRODUCTS

	// CATEGORIES
	CategoryGetFailed    Code = "CATEGORY_GET_FAILED"
	CategoryCreateFailed Code = "CATEGORY_CREATE_FALIED"
	CategoryAlreadyExist Code = "CATEGORY_ALREADY_EXIST"

	// BRANDS
	BrandValidationFailed Code = "BRAND_VALIDATION_FAILED"
	BrandNotFound         Code = "BRAND_NOT_FOUND"
	BrandAlreadyExists    Code = "BRAND_ALREADY_EXISTS"
	BrandCreateFailed     Code = "BRAND_CREATE_FAILED"
	BrandGetFailed        Code = "BRAND_GET_FAILED"
	BrandUpdateFailed     Code = "BRAND_UPDATE_FAILED"
	BrandDeleteFailed     Code = "BRAND_DELETE_FAILED"
)

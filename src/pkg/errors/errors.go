package errors

import (
	"emc_lb/src/pkg/res"
)

type Code = res.ErrorCode

const (
	// ===== COMMON =====
	CommonBadRequest      Code = res.ErrCodeBadRequest
	CommonNotFound        Code = res.ErrCodeNotFound
	CommonConflict        Code = res.ErrCodeConflict
	CommonInternal        Code = res.ErrCodeInternal
	CommonUnauthorized    Code = res.ErrCodeUnauthorized
	CommonForbidden       Code = res.ErrCodeForbidden
	CommonTooManyRequests Code = res.ErrCodeTooManyRequests

	// ===== USER =====
	UserValidationFailed     Code = "USER_VALIDATION_FAILED"
	UserNotFound             Code = "USER_NOT_FOUND"
	UserAlreadyExists        Code = "USER_ALREADY_EXISTS"
	UserCreateFailed         Code = "USER_CREATE_FAILED"
	UserGetFailed            Code = "USER_GET_FAILED"
	UserUpdateFailed         Code = "USER_UPDATE_FAILED"
	UserStatusChangeFailed   Code = "USER_STATUS_CHANGE_FAILED"
	UserPasswordUpdateFailed Code = "USER_PASSWORD_UPDATE_FAILED"
	UserUnauthorized         Code = "USER_UNAUTHORIZED"
	UserInvalidCreatorID     Code = "USER_INVALID_CREATOR_ID"
	UserRoleUpdateForbidden  Code = "USER_ROLE_UPDATE_FORBIDDEN"
	UserAgentCreateForbidden Code = "USER_AGENT_CREATE_FORBIDDEN"
	UserInvalidAvatarUpload  Code = "USER_INVALID_AVATAR_UPLOAD"
	UserAvatarProcessFailed  Code = "USER_AVATAR_PROCESS_FAILED"
	UserInvalidFormat        Code = "USER_INVALID_FORMAT"

	// ===== SHOP =====
	ShopNotFound      Code = "SHOP_NOT_FOUND"
	ShopAlreadyExists Code = "SHOP_ALREADY_EXISTS"
	ShopCreateFailed  Code = "SHOP_CREATE_FAILED"
	ShopUpdateFailed  Code = "SHOP_UPDATE_FAILED"
	ShopDeleteFailed  Code = "SHOP_DELETE_FAILED"
	ShopGetFailed     Code = "SHOP_GET_FAILED"
	ShopForbidden     Code = "SHOP_FORBIDDEN"

	// ===== PRODUCT =====
	ProductNotFound         Code = "PRODUCT_NOT_FOUND"
	ProductAlreadyExists    Code = "PRODUCT_ALREADY_EXISTS"
	ProductCreateFailed     Code = "PRODUCT_CREATE_FAILED"
	ProductUpdateFailed     Code = "PRODUCT_UPDATE_FAILED"
	ProductDeleteFailed     Code = "PRODUCT_DELETE_FAILED"
	ProductGetFailed        Code = "PRODUCT_GET_FAILED"
	ProductOutOfStock       Code = "PRODUCT_OUT_OF_STOCK"
	ProductInvalidVariant   Code = "PRODUCT_INVALID_VARIANT"
	ProductValidationFailed Code = "PRODUCT_VALIDATION_FAILED"

	// ===== CATEGORY =====
	CategoryGetFailed    Code = "CATEGORY_GET_FAILED"
	CategoryCreateFailed Code = "CATEGORY_CREATE_FAILED"
	CategoryAlreadyExist Code = "CATEGORY_ALREADY_EXIST"
	CategoryNotFound     Code = "CATEGORY_NOT_FOUND"
	CategoryUpdateFailed Code = "CATEGORY_UPDATE_FAILED"
	CategoryDeleteFailed Code = "CATEGORY_DELETE_FAILED"

	// ===== BRAND =====
	BrandValidationFailed Code = "BRAND_VALIDATION_FAILED"
	BrandNotFound         Code = "BRAND_NOT_FOUND"
	BrandAlreadyExists    Code = "BRAND_ALREADY_EXISTS"
	BrandCreateFailed     Code = "BRAND_CREATE_FAILED"
	BrandGetFailed        Code = "BRAND_GET_FAILED"
	BrandUpdateFailed     Code = "BRAND_UPDATE_FAILED"
	BrandDeleteFailed     Code = "BRAND_DELETE_FAILED"

	// ===== CART =====
	CartNotFound         Code = "CART_NOT_FOUND"
	CartGetFailed        Code = "CART_GET_FAILED"
	CartCreateFailed     Code = "CART_CREATE_FAILED"
	CartUpdateFailed     Code = "CART_UPDATE_FAILED"
	CartDeleteFailed     Code = "CART_DELETE_FAILED"
	CartEmpty            Code = "CART_EMPTY"
	CartItemNotFound     Code = "CART_ITEM_NOT_FOUND"
	CartItemInvalidQty   Code = "CART_ITEM_INVALID_QUANTITY"
	CartItemAddFailed    Code = "CART_ITEM_ADD_FAILED"
	CartItemRemoveFailed Code = "CART_ITEM_REMOVE_FAILED"
	CartCheckoutFailed   Code = "CART_CHECKOUT_FAILED"

	// ===== ORDER =====
	OrderNotFound         Code = "ORDER_NOT_FOUND"
	OrderGetFailed        Code = "ORDER_GET_FAILED"
	OrderCreateFailed     Code = "ORDER_CREATE_FAILED"
	OrderUpdateFailed     Code = "ORDER_UPDATE_FAILED"
	OrderCancelFailed     Code = "ORDER_CANCEL_FAILED"
	OrderStatusInvalid    Code = "ORDER_STATUS_INVALID"
	OrderForbidden        Code = "ORDER_FORBIDDEN"
	OrderAlreadyCancelled Code = "ORDER_ALREADY_CANCELLED"
	OrderAlreadyCompleted Code = "ORDER_ALREADY_COMPLETED"
	OrderItemNotFound     Code = "ORDER_ITEM_NOT_FOUND"

	// ===== PAYMENT =====
	PaymentNotFound        Code = "PAYMENT_NOT_FOUND"
	PaymentGetFailed       Code = "PAYMENT_GET_FAILED"
	PaymentCreateFailed    Code = "PAYMENT_CREATE_FAILED"
	PaymentFailed          Code = "PAYMENT_FAILED"
	PaymentInvalidProvider Code = "PAYMENT_INVALID_PROVIDER"
	PaymentAlreadyPaid     Code = "PAYMENT_ALREADY_PAID"
	PaymentExpired         Code = "PAYMENT_EXPIRED"
	PaymentVerifyFailed    Code = "PAYMENT_VERIFY_FAILED"

	// ===== COUPON =====
	CouponNotFound       Code = "COUPON_NOT_FOUND"
	CouponGetFailed      Code = "COUPON_GET_FAILED"
	CouponCreateFailed   Code = "COUPON_CREATE_FAILED"
	CouponUpdateFailed   Code = "COUPON_UPDATE_FAILED"
	CouponDeleteFailed   Code = "COUPON_DELETE_FAILED"
	CouponAlreadyUsed    Code = "COUPON_ALREADY_USED"
	CouponExpired        Code = "COUPON_EXPIRED"
	CouponInvalid        Code = "COUPON_INVALID"
	CouponMinOrderNotMet Code = "COUPON_MIN_ORDER_NOT_MET"
	CouponLimitReached   Code = "COUPON_LIMIT_REACHED"
)

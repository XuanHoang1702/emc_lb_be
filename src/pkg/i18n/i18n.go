package i18n

import (
	"fmt"
	"strings"
)

// #nosec G101
var messages = map[string]map[string]string{
	"en": {
		"val_gt":              "%s must be greater than %s",
		"val_lt":              "%s must be less than %s",
		"val_gte":             "%s must be greater than or equal to %s",
		"val_lte":             "%s must be less than or equal to %s",
		"val_uuid":            "%s must be a valid UUID",
		"val_slug":            "%s must contain only lowercase letters, numbers, hyphens, or dots",
		"val_min":             "%s must have at least %s characters",
		"val_max":             "%s must not exceed %s characters",
		"val_min_int":         "%s must have a value greater than or equal to %s",
		"val_max_int":         "%s must have a value less than or equal to %s",
		"val_oneof":           "%s must be one of the values: %s",
		"val_required":        "%s is required",
		"val_search":          "%s can only contain letters, numbers, and spaces",
		"val_email":           "%s must be a valid email format",
		"val_datetime":        "%s must be a valid date and time",
		"val_email_advanced":  "%s is in the blocked list",
		"val_password_strong": "%s must be at least 8 characters long and contain lowercase, uppercase, numbers, and special characters",
		"val_file_ext":        "%s only allows files with extensions: %s",
		"val_default":         "%s is invalid",

		"success_logout":           "Logout successfully",
		"success_email_verified":   "Email verified successfully",
		"success_account_deleted":  "Account deleted successfully",
		"success_category_deleted": "Category deleted successfully",
		"success_brand_deleted":    "Brand deleted successfully",
		"err_internal_server":      "Internal server error",
		"err_too_many_requests":    "You have sent too many requests, please try again later.",
		"err_unauthorized_role":    "Unauthorized: Invalid role",
		"err_forbidden_resource":   "Forbidden: You do not have permission to access this resource",

		// Coupon & Cart & Order messages
		"Coupon code already exists":                                     "Coupon code already exists",
		"Invalid coupon code":                                            "Invalid coupon code",
		"Coupon is not active":                                           "Coupon is not active",
		"Coupon is expired or not yet valid":                             "Coupon is expired or not yet valid",
		"Coupon usage limit exceeded":                                    "Coupon usage limit exceeded",
		"Order amount does not meet minimum requirement for this coupon": "Order amount does not meet minimum requirement for this coupon",
		"Cart cleared successfully":                                      "Cart cleared successfully",
		"Order must have at least one item":                              "Order must have at least one item",
	},
	"vi": {
		"val_gt":              "%s phải lớn hơn %s",
		"val_lt":              "%s phải nhỏ hơn %s",
		"val_gte":             "%s phải lớn hơn hoặc bằng %s",
		"val_lte":             "%s phải nhỏ hơn hoặc bằng %s",
		"val_uuid":            "%s phải là UUID hợp lệ",
		"val_slug":            "%s chỉ được chứa chữ thường, số, dấu gạch ngang hoặc dấu chấm",
		"val_min":             "%s phải có ít nhất %s ký tự",
		"val_max":             "%s không được vượt quá %s ký tự",
		"val_min_int":         "%s phải có giá trị lớn hơn hoặc bằng %s",
		"val_max_int":         "%s phải có giá trị nhỏ hơn hoặc bằng %s",
		"val_oneof":           "%s phải là một trong các giá trị: %s",
		"val_required":        "%s không được để trống",
		"val_search":          "%s chỉ được chứa chữ thường, in hoa, số và khoảng trắng",
		"val_email":           "%s phải đúng định dạng email",
		"val_datetime":        "%s phải là ngày giờ hợp lệ",
		"val_email_advanced":  "%s nằm trong danh sách bị cấm",
		"val_password_strong": "%s phải ít nhất 8 ký tự và gồm chữ thường, chữ in hoa, số và ký tự đặc biệt",
		"val_file_ext":        "%s chỉ cho phép file có đuôi: %s",
		"val_default":         "%s không hợp lệ",

		"success_logout":           "Đăng xuất thành công",
		"success_email_verified":   "Xác thực email thành công",
		"success_account_deleted":  "Xóa tài khoản thành công",
		"success_category_deleted": "Xóa danh mục thành công",
		"success_brand_deleted":    "Xóa thương hiệu thành công",
		"err_internal_server":      "Lỗi hệ thống nội bộ",
		"err_too_many_requests":    "Bạn đã gửi quá nhiều yêu cầu, vui lòng thử lại sau.",
		"err_unauthorized_role":    "Không có quyền: Role không hợp lệ",
		"err_forbidden_resource":   "Từ chối truy cập: Bạn không có quyền thao tác tài nguyên này",

		// Coupon & Cart & Order messages
		"Coupon code already exists":                                     "Mã giảm giá đã tồn tại",
		"Invalid coupon code":                                            "Mã giảm giá không hợp lệ",
		"Coupon is not active":                                           "Mã giảm giá chưa được kích hoạt",
		"Coupon is expired or not yet valid":                             "Mã giảm giá đã hết hạn hoặc chưa đến thời gian áp dụng",
		"Coupon usage limit exceeded":                                    "Mã giảm giá đã vượt quá giới hạn sử dụng",
		"Order amount does not meet minimum requirement for this coupon": "Giá trị đơn hàng chưa đạt mức tối thiểu để áp dụng mã này",
		"Cart cleared successfully":                                      "Xóa giỏ hàng thành công",
		"Order must have at least one item":                              "Đơn hàng phải có ít nhất một sản phẩm",
	},
}

// GetMessage retrieves the localized string.
func GetMessage(lang string, key string, args ...any) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if !strings.HasPrefix(lang, "vi") {
		lang = "en"
	} else {
		lang = "vi"
	}

	msgs, ok := messages[lang]
	if !ok {
		msgs = messages["en"]
	}

	msg, ok := msgs[key]
	if !ok {
		// Fallback to english if key doesn't exist in vi
		msg, ok = messages["en"][key]
		if !ok {
			return key // return raw key if not found
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}

	return msg
}

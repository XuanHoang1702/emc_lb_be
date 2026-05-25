package res

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ErrorCode string

const (
	ErrCodeBadRequest          ErrorCode = "COMMON_BAD_REQUEST"
	ErrCodeNotFound            ErrorCode = "COMMON_NOT_FOUND"
	ErrCodeConflict            ErrorCode = "COMMON_CONFLICT"
	ErrCodeInternal            ErrorCode = "COMMON_INTERNAL_SERVER_ERROR"
	ErrCodeUnauthorized        ErrorCode = "COMMON_UNAUTHORIZED"
	ErrCodeForbidden           ErrorCode = "COMMON_FORBIDDEN"
	ErrCodeTooManyRequests     ErrorCode = "COMMON_TOO_MANY_REQUESTS"
	ErrCodeInvalidIdentityType ErrorCode = "COMMON_INVALID_IDENTITY_TYPE"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type AppError struct {
	Message    string       `json:"message"`
	Code       ErrorCode    `json:"code"`
	StatusCode int          `json:"status"`
	Err        error        `json:"-"`
	Errors     []FieldError `json:"errors,omitempty"`
}

type APIResponse struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message,omitempty"`
	Data       any          `json:"data,omitempty"`
	Pagination any          `json:"pagination,omitempty"`
	Code       ErrorCode    `json:"code,omitempty"`
	Status     int          `json:"status,omitempty"`
	Errors     []FieldError `json:"errors,omitempty"`
}

func (ae *AppError) Error() string {
	if ae == nil {
		return ""
	}
	return ae.Message
}

func NewError(message string, code ErrorCode) error {
	return &AppError{Message: message, Code: code, StatusCode: httpStatusFromCode(code)}
}

func NewValidationError(message string, code ErrorCode, errs []FieldError) error {
	return &AppError{Message: message, Code: code, StatusCode: http.StatusBadRequest, Errors: errs}
}

func WrapError(err error, message string, code ErrorCode) error {
	return &AppError{Err: err, Message: message, Code: code, StatusCode: httpStatusFromCode(code)}
}

func Success(ctx *gin.Context, status int, args ...any) {
	resp := APIResponse{Success: true, Status: status}

	var payload any
	if len(args) > 0 {
		if msg, ok := args[0].(string); ok {
			resp.Message = capitalizeFirst(msg)
			if len(args) > 1 {
				payload = args[1]
			}
		} else {
			payload = args[0]
		}
	}

	if payload != nil {
		if m, ok := payload.(map[string]any); ok {
			if p, exists := m["pagination"]; exists {
				resp.Pagination = p
			}
			if d, exists := m["data"]; exists {
				resp.Data = d
			} else {
				resp.Data = m
			}
		} else {
			resp.Data = payload
		}
	}

	ctx.JSON(status, resp)
}

func Error(ctx *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		status := appErr.StatusCode
		if status == 0 {
			status = httpStatusFromCode(appErr.Code)
		}
		ctx.JSON(status, APIResponse{
			Success: false,
			Message: capitalizeFirst(appErr.Message),
			Code:    appErr.Code,
			Status:  status,
			Errors:  appErr.Errors,
		})
		return
	}
	ctx.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: "Internal server error",
		Code:    ErrCodeInternal,
		Status:  http.StatusInternalServerError,
	})
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func httpStatusFromCode(code ErrorCode) int {
	switch code {
	case ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

package res

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAppError_Error(t *testing.T) {
	var err *AppError
	if err.Error() != "" {
		t.Errorf("expected empty string for nil error, got %s", err.Error())
	}

	err = &AppError{Message: "test msg"}
	if err.Error() != "test msg" {
		t.Errorf("expected test msg, got %s", err.Error())
	}
}

func TestNewError(t *testing.T) {
	err := NewError("test", ErrCodeBadRequest)
	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatal("expected *AppError")
	}
	if appErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected %d, got %d", http.StatusBadRequest, appErr.StatusCode)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test", ErrCodeBadRequest, []FieldError{{Field: "f", Message: "m"}})
	appErr, _ := err.(*AppError)
	if len(appErr.Errors) != 1 {
		t.Errorf("expected 1 field error, got %d", len(appErr.Errors))
	}
}

func TestWrapError(t *testing.T) {
	inner := errors.New("inner")
	err := WrapError(inner, "wrapper", ErrCodeInternal)
	appErr, _ := err.(*AppError)
	if appErr.Err != inner {
		t.Errorf("expected inner error to match")
	}
}

func TestHttpStatusFromCode(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{ErrCodeBadRequest, http.StatusBadRequest},
		{ErrCodeNotFound, http.StatusNotFound},
		{ErrCodeConflict, http.StatusConflict},
		{ErrCodeUnauthorized, http.StatusUnauthorized},
		{ErrCodeForbidden, http.StatusForbidden},
		{ErrCodeTooManyRequests, http.StatusTooManyRequests},
		{ErrCodeServiceUnavailable, http.StatusServiceUnavailable},
		{"UNKNOWN", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		if httpStatusFromCode(tt.code) != tt.expected {
			t.Errorf("expected %d for %s", tt.expected, tt.code)
		}
	}
}

func TestCapitalizeFirst(t *testing.T) {
	if capitalizeFirst("") != "" {
		t.Error("expected empty")
	}
	if capitalizeFirst("hello") != "Hello" {
		t.Error("expected Hello")
	}
}

func TestSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{Header: make(http.Header)}

	Success(c, http.StatusOK, "msg", map[string]any{"data": "val", "pagination": "page"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test AppError
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{Header: make(http.Header)}
	Error(c, NewError("bad request", ErrCodeBadRequest))
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Test fallback standard error
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = &http.Request{Header: make(http.Header)}
	Error(c2, errors.New("unknown error"))
	if w2.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w2.Code)
	}
}

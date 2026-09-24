package validation

import (
	"errors"
	"net/http/httptest"
	"testing"

	"emc_lb/src/pkg/res"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type TestStruct struct {
	Name string `json:"name" validate:"required"`
}

func TestBindJSON_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/", nil) // No body
	
	err := BindJSON(c, &TestStruct{}, res.ErrCodeBadRequest)
	if err == nil {
		t.Error("expected error for nil body")
	}
}

func TestValidationMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	
	msg := validationMessage(c, "field", "required", "")
	if msg == "" {
		t.Error("expected message")
	}
	
	msg2 := validationMessage(c, "field", "oneof", "a b")
	if msg2 == "" {
		t.Error("expected message")
	}

	msg3 := validationMessage(c, "field", "unknown_tag", "")
	if msg3 == "" {
		t.Error("expected message")
	}
}

func TestHandleValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	
	err := HandleValidationErrors(c, errors.New("normal error"))
	appErr, ok := err.(*res.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}
	if len(appErr.Errors) != 1 || appErr.Errors[0].Message != "normal error" {
		t.Error("expected fallback for non-validation error")
	}
	
	v := validator.New()
	st := TestStruct{}
	valErr := v.Struct(st)
	
	err2 := HandleValidationErrors(c, valErr)
	appErr2, _ := err2.(*res.AppError)
	if len(appErr2.Errors) == 0 {
		t.Error("expected validation errors")
	}
}

package validation

import (
	"fmt"
	"reflect"
	"strings"

	"emc_lb/src/pkg/i18n"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func InitValidator() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return fmt.Errorf("failed to get validator engine")
	}

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		for _, key := range []string{"json", "uri", "form", "query"} {
			tag := fld.Tag.Get(key)
			if tag == "" {
				continue
			}

			name := strings.Split(tag, ",")[0]
			if name != "" && name != "-" {
				return name
			}
		}

		return fld.Name
	})

	RegisterCustomValidation(v)
	return nil
}

func HandleValidationErrors(ctx *gin.Context, err error) error {
	validationError, ok := err.(validator.ValidationErrors)
	if !ok {
		return res.NewValidationError("Validation failed", res.ErrCodeBadRequest, []res.FieldError{{Field: "request", Message: err.Error()}})
	}

	errors := make([]res.FieldError, 0, len(validationError))

	for _, e := range validationError {
		root := strings.Split(e.Namespace(), ".")[0]
		rawPath := strings.TrimPrefix(e.Namespace(), root+".")
		parts := strings.Split(rawPath, ".")

		for i, part := range parts {
			if strings.Contains(part, "[") {
				idx := strings.Index(part, "[")
				base := utils.CamelToSnake(part[:idx])
				index := part[idx:]
				parts[i] = base + index
			} else {
				parts[i] = utils.CamelToSnake(part)
			}
		}

		fieldPath := strings.Join(parts, ".")
		msg := validationMessage(ctx, fieldPath, e.Tag(), e.Param())
		errors = append(errors, res.FieldError{Field: fieldPath, Message: msg})
	}

	return res.NewValidationError("Validation failed", res.ErrCodeBadRequest, errors)
}

func validationMessage(ctx *gin.Context, fieldPath, tag, param string) string {
	lang := ctx.GetHeader("Accept-Language")
	switch tag {
	case "gt", "lt", "gte", "lte", "min", "max", "min_int", "max_int":
		return i18n.GetMessage(lang, "val_"+tag, fieldPath, param)
	case "uuid", "uuid4":
		return i18n.GetMessage(lang, "val_uuid", fieldPath)
	case "slug", "required", "search", "email", "datetime", "email_advanced", "password_strong":
		return i18n.GetMessage(lang, "val_"+tag, fieldPath)
	case "oneof":
		allowedValues := strings.Join(strings.Split(param, " "), ", ")
		return i18n.GetMessage(lang, "val_oneof", fieldPath, allowedValues)
	case "file_ext":
		allowedValues := strings.Join(strings.Split(param, " "), ", ")
		return i18n.GetMessage(lang, "val_file_ext", fieldPath, allowedValues)
	default:
		return i18n.GetMessage(lang, "val_default", fieldPath)
	}
}

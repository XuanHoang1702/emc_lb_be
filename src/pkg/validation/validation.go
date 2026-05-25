package validation

import (
	"fmt"
	"reflect"
	"strings"

	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"

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

func HandleValidationErrors(err error) error {
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
		msg := validationMessage(fieldPath, e.Tag(), e.Param())
		errors = append(errors, res.FieldError{Field: fieldPath, Message: msg})
	}

	return res.NewValidationError("Validation failed", res.ErrCodeBadRequest, errors)
}

func validationMessage(fieldPath, tag, param string) string {
	switch tag {
	case "gt":
		return fmt.Sprintf("%s phải lớn hơn %s", fieldPath, param)
	case "lt":
		return fmt.Sprintf("%s phải nhỏ hơn %s", fieldPath, param)
	case "gte":
		return fmt.Sprintf("%s phải lớn hơn hoặc bằng %s", fieldPath, param)
	case "lte":
		return fmt.Sprintf("%s phải nhỏ hơn hoặc bằng %s", fieldPath, param)
	case "uuid", "uuid4":
		return fmt.Sprintf("%s phải là UUID hợp lệ", fieldPath)
	case "slug":
		return fmt.Sprintf("%s chỉ được chứa chữ thường, số, dấu gạch ngang hoặc dấu chấm", fieldPath)
	case "min":
		return fmt.Sprintf("%s phải có ít nhất %s ký tự", fieldPath, param)
	case "max":
		return fmt.Sprintf("%s không được vượt quá %s ký tự", fieldPath, param)
	case "min_int":
		return fmt.Sprintf("%s phải có giá trị lớn hơn hoặc bằng %s", fieldPath, param)
	case "max_int":
		return fmt.Sprintf("%s phải có giá trị nhỏ hơn hoặc bằng %s", fieldPath, param)
	case "oneof":
		allowedValues := strings.Join(strings.Split(param, " "), ", ")
		return fmt.Sprintf("%s phải là một trong các giá trị: %s", fieldPath, allowedValues)
	case "required":
		return fmt.Sprintf("%s không được để trống", fieldPath)
	case "search":
		return fmt.Sprintf("%s chỉ được chứa chữ thường, in hoa, số và khoảng trắng", fieldPath)
	case "email":
		return fmt.Sprintf("%s phải đúng định dạng email", fieldPath)
	case "datetime":
		return fmt.Sprintf("%s phải là ngày giờ hợp lệ", fieldPath)
	case "email_advanced":
		return fmt.Sprintf("%s nằm trong danh sách bị cấm", fieldPath)
	case "password_strong":
		return fmt.Sprintf("%s phải ít nhất 8 ký tự và gồm chữ thường, chữ in hoa, số và ký tự đặc biệt", fieldPath)
	case "file_ext":
		allowedValues := strings.Join(strings.Split(param, " "), ", ")
		return fmt.Sprintf("%s chỉ cho phép file có đuôi: %s", fieldPath, allowedValues)
	default:
		return fmt.Sprintf("%s không hợp lệ", fieldPath)
	}
}

package validation

import (
	"emc_lb/src/pkg/res"

	"github.com/gin-gonic/gin"
)

func BindJSON(ctx *gin.Context, dst any, validationCode res.ErrorCode) error {
	if err := ctx.ShouldBindJSON(dst); err != nil {
		appErr := HandleValidationErrors(ctx, err).(*res.AppError)
		if validationCode != "" {
			appErr.Code = validationCode
			if appErr.StatusCode == 0 {
				appErr.StatusCode = 400
			}
		}
		return appErr
	}
	return nil
}

func BindURI(ctx *gin.Context, dst any, validationCode res.ErrorCode) error {
	if err := ctx.ShouldBindUri(dst); err != nil {
		appErr := HandleValidationErrors(ctx, err).(*res.AppError)
		if validationCode != "" {
			appErr.Code = validationCode
		}
		return appErr
	}
	return nil
}

func BindQuery(ctx *gin.Context, dst any, validationCode res.ErrorCode) error {
	if err := ctx.ShouldBindQuery(dst); err != nil {
		appErr := HandleValidationErrors(ctx, err).(*res.AppError)
		if validationCode != "" {
			appErr.Code = validationCode
		}
		return appErr
	}
	return nil
}

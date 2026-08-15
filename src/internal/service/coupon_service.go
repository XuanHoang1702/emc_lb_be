package service

import (
	"context"
	"net/http"
	"strings"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/res"
)

type CouponService interface {
	Create(ctx context.Context, req entities.CreateCouponRequest) (entities.CouponResponse, error)
	GetByCode(ctx context.Context, code string) (entities.CouponResponse, error)
	List(ctx context.Context) ([]entities.CouponResponse, error)
	Update(ctx context.Context, id string, req entities.UpdateCouponRequest) (entities.CouponResponse, error)
	ValidateCouponForAmount(ctx context.Context, code string, amount float64) (entities.Coupon, error)
	IncrementUsage(ctx context.Context, code string, count int64) error
}

type couponService struct {
	couponRepository repository.CouponRepository
}

func NewCouponService(repo repository.CouponRepository) CouponService {
	return &couponService{couponRepository: repo}
}

func (s *couponService) Create(ctx context.Context, req entities.CreateCouponRequest) (entities.CouponResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))

	// Check if already exists
	if existing, err := s.couponRepository.GetByCode(ctx, code); err == nil && existing.ID != "" {
		return entities.CouponResponse{}, &res.AppError{
			Message:    "Coupon code already exists",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	now := time.Now().UTC()
	coupon := entities.Coupon{
		Code:              code,
		Type:              req.Type,
		Value:             req.Value,
		MinOrderAmount:    req.MinOrderAmount,
		MaxDiscountAmount: req.MaxDiscountAmount,
		UsageLimit:        req.UsageLimit,
		UsageCount:        0,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		IsActive:          req.IsActive,
		IsDeleted:         false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	created, err := s.couponRepository.Create(ctx, coupon)
	if err != nil {
		return entities.CouponResponse{}, res.WrapError(err, "Failed to create coupon", erres.CommonInternal)
	}

	return s.toResponse(created), nil
}

func (s *couponService) GetByCode(ctx context.Context, code string) (entities.CouponResponse, error) {
	coupon, err := s.couponRepository.GetByCode(ctx, strings.ToUpper(code))
	if err != nil {
		return entities.CouponResponse{}, res.WrapError(err, "Coupon not found", erres.CommonNotFound)
	}
	return s.toResponse(coupon), nil
}

func (s *couponService) List(ctx context.Context) ([]entities.CouponResponse, error) {
	coupons, err := s.couponRepository.List(ctx)
	if err != nil {
		return nil, res.WrapError(err, "Failed to list coupons", erres.CommonInternal)
	}

	var responses []entities.CouponResponse
	for _, c := range coupons {
		responses = append(responses, s.toResponse(c))
	}
	return responses, nil
}

func (s *couponService) Update(ctx context.Context, id string, req entities.UpdateCouponRequest) (entities.CouponResponse, error) {
	update := make(map[string]any)

	if req.Type != "" {
		update["type"] = req.Type
	}
	if req.Value > 0 {
		update["value"] = req.Value
	}
	if req.MinOrderAmount > 0 {
		update["min_order_amount"] = req.MinOrderAmount
	}
	if req.MaxDiscountAmount > 0 {
		update["max_discount_amount"] = req.MaxDiscountAmount
	}
	if req.UsageLimit > 0 {
		update["usage_limit"] = req.UsageLimit
	}
	if !req.StartDate.IsZero() {
		update["start_date"] = req.StartDate
	}
	if !req.EndDate.IsZero() {
		update["end_date"] = req.EndDate
	}
	if req.IsActive != nil {
		update["is_active"] = *req.IsActive
	}

	if len(update) > 0 {
		update["updated_at"] = time.Now().UTC()
	}

	updated, err := s.couponRepository.Update(ctx, id, update)
	if err != nil {
		return entities.CouponResponse{}, res.WrapError(err, "Failed to update coupon", erres.CommonInternal)
	}

	return s.toResponse(updated), nil
}

func (s *couponService) ValidateCouponForAmount(ctx context.Context, code string, amount float64) (entities.Coupon, error) {
	coupon, err := s.couponRepository.GetByCode(ctx, strings.ToUpper(code))
	if err != nil {
		return entities.Coupon{}, &res.AppError{
			Message:    "Invalid coupon code",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	if !coupon.IsActive {
		return entities.Coupon{}, &res.AppError{
			Message:    "Coupon is not active",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	now := time.Now().UTC()
	if now.Before(coupon.StartDate) || now.After(coupon.EndDate) {
		return entities.Coupon{}, &res.AppError{
			Message:    "Coupon is expired or not yet valid",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	if coupon.UsageLimit > 0 && coupon.UsageCount >= coupon.UsageLimit {
		return entities.Coupon{}, &res.AppError{
			Message:    "Coupon usage limit exceeded",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	if amount < coupon.MinOrderAmount {
		return entities.Coupon{}, &res.AppError{
			Message:    "Order amount does not meet minimum requirement for this coupon",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	return coupon, nil
}

func (s *couponService) IncrementUsage(ctx context.Context, code string, count int64) error {
	return s.couponRepository.IncrementUsage(ctx, code, count)
}

func (s *couponService) toResponse(c entities.Coupon) entities.CouponResponse {
	return entities.CouponResponse{
		ID:                c.ID,
		Code:              c.Code,
		Type:              c.Type,
		Value:             c.Value,
		MinOrderAmount:    c.MinOrderAmount,
		MaxDiscountAmount: c.MaxDiscountAmount,
		UsageLimit:        c.UsageLimit,
		UsageCount:        c.UsageCount,
		StartDate:         c.StartDate,
		EndDate:           c.EndDate,
		IsActive:          c.IsActive,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/entities"
)

// stubCouponRepository is a minimal test double for CouponRepository
type stubCouponRepository struct {
	coupons     map[string]entities.Coupon
	createErr   error
	getByCodeFn func(code string) (entities.Coupon, error)
}

func newStubCouponRepo() *stubCouponRepository {
	return &stubCouponRepository{coupons: make(map[string]entities.Coupon)}
}

func (r *stubCouponRepository) Create(_ context.Context, c entities.Coupon) (entities.Coupon, error) {
	if r.createErr != nil {
		return entities.Coupon{}, r.createErr
	}
	c.ID = "coupon-id-1"
	r.coupons[c.Code] = c
	return c, nil
}

func (r *stubCouponRepository) GetByCode(_ context.Context, code string) (entities.Coupon, error) {
	if r.getByCodeFn != nil {
		return r.getByCodeFn(code)
	}
	c, ok := r.coupons[code]
	if !ok {
		return entities.Coupon{}, errors.New("not found")
	}
	return c, nil
}

func (r *stubCouponRepository) GetByID(_ context.Context, id string) (entities.Coupon, error) {
	for _, c := range r.coupons {
		if c.ID == id {
			return c, nil
		}
	}
	return entities.Coupon{}, errors.New("not found")
}

func (r *stubCouponRepository) List(_ context.Context) ([]entities.Coupon, error) {
	var result []entities.Coupon
	for _, c := range r.coupons {
		result = append(result, c)
	}
	return result, nil
}

func (r *stubCouponRepository) Update(_ context.Context, _ string, _ map[string]any) (entities.Coupon, error) {
	return entities.Coupon{}, nil
}

func (r *stubCouponRepository) IncrementUsage(_ context.Context, code string, count int64) error {
	if c, ok := r.coupons[code]; ok {
		c.UsageCount += count
		r.coupons[code] = c
	}
	return nil
}

func validCoupon() entities.Coupon {
	return entities.Coupon{
		ID:             "coupon-1",
		Code:           "SAVE10",
		Type:           "percentage",
		Value:          10,
		MinOrderAmount: 100,
		UsageLimit:     100,
		UsageCount:     0,
		StartDate:      time.Now().Add(-1 * time.Hour),
		EndDate:        time.Now().Add(24 * time.Hour),
		IsActive:       true,
	}
}

func TestValidateCouponForAmount_Success(t *testing.T) {
	repo := newStubCouponRepo()
	c := validCoupon()
	repo.coupons[c.Code] = c
	svc := service.NewCouponService(repo, nil)

	result, err := svc.ValidateCouponForAmount(context.Background(), "SAVE10", 200)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Code != "SAVE10" {
		t.Fatalf("expected code SAVE10, got %s", result.Code)
	}
}

func TestValidateCouponForAmount_Expired(t *testing.T) {
	repo := newStubCouponRepo()
	c := validCoupon()
	c.StartDate = time.Now().Add(-48 * time.Hour)
	c.EndDate = time.Now().Add(-24 * time.Hour) // expired yesterday
	repo.coupons[c.Code] = c
	svc := service.NewCouponService(repo, nil)

	_, err := svc.ValidateCouponForAmount(context.Background(), "SAVE10", 200)
	if err == nil {
		t.Fatal("expected error for expired coupon")
	}
}

func TestValidateCouponForAmount_Inactive(t *testing.T) {
	repo := newStubCouponRepo()
	c := validCoupon()
	c.IsActive = false
	repo.coupons[c.Code] = c
	svc := service.NewCouponService(repo, nil)

	_, err := svc.ValidateCouponForAmount(context.Background(), "SAVE10", 200)
	if err == nil {
		t.Fatal("expected error for inactive coupon")
	}
}

func TestValidateCouponForAmount_UsageLimitExceeded(t *testing.T) {
	repo := newStubCouponRepo()
	c := validCoupon()
	c.UsageLimit = 5
	c.UsageCount = 5
	repo.coupons[c.Code] = c
	svc := service.NewCouponService(repo, nil)

	_, err := svc.ValidateCouponForAmount(context.Background(), "SAVE10", 200)
	if err == nil {
		t.Fatal("expected error for usage limit exceeded")
	}
}

func TestValidateCouponForAmount_BelowMinOrder(t *testing.T) {
	repo := newStubCouponRepo()
	c := validCoupon()
	c.MinOrderAmount = 500
	repo.coupons[c.Code] = c
	svc := service.NewCouponService(repo, nil)

	_, err := svc.ValidateCouponForAmount(context.Background(), "SAVE10", 200)
	if err == nil {
		t.Fatal("expected error for order below minimum amount")
	}
}

func TestCreate_DuplicateCode(t *testing.T) {
	repo := newStubCouponRepo()
	c := validCoupon()
	repo.coupons[c.Code] = c
	svc := service.NewCouponService(repo, nil)

	_, err := svc.Create(context.Background(), entities.CreateCouponRequest{
		Code:  "SAVE10",
		Type:  "percentage",
		Value: 15,
	})
	if err == nil {
		t.Fatal("expected error for duplicate coupon code")
	}
}

func TestValidateCouponForAmount_InvalidCode(t *testing.T) {
	repo := newStubCouponRepo()
	svc := service.NewCouponService(repo, nil)

	_, err := svc.ValidateCouponForAmount(context.Background(), "NONEXISTENT", 200)
	if err == nil {
		t.Fatal("expected error for invalid coupon code")
	}
}

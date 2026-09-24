package mapping

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/pkg/entities"
)

func TestToBrandResponse(t *testing.T) {
	now := time.Now()
	brand := entities.Brand{
		ID:          "1",
		Name:        "Test Brand",
		Slug:        "test-brand",
		Description: "desc",
		CreatedAt:   now,
	}
	resp := ToBrandResponse(brand)
	if resp.ID != brand.ID || resp.Name != brand.Name || resp.CreatedAt != now {
		t.Errorf("ToBrandResponse failed to map correctly")
	}
}

func TestToCategoryResponse(t *testing.T) {
	now := time.Now()
	cat := entities.Category{
		ID:        "2",
		Name:      "Test Cat",
		Slug:      "test-cat",
		CreatedAt: now,
	}
	resp := ToCategoryResponse(cat)
	if resp.ID != cat.ID || resp.Name != cat.Name || resp.CreatedAt != now {
		t.Errorf("ToCategoryResponse failed to map correctly")
	}
}

func TestToOrderResponse(t *testing.T) {
	now := time.Now()
	order := entities.Order{
		ID:        "3",
		ShopID:    "s1",
		UserID:    "u1",
		Status:    "pending",
		CreatedAt: now,
	}
	resp := ToOrderResponse(order)
	if resp.ID != order.ID || resp.ShopID != order.ShopID || resp.Status != order.Status || resp.CreatedAt != now {
		t.Errorf("ToOrderResponse failed to map correctly")
	}
}

func TestToProductResponse(t *testing.T) {
	now := time.Now()
	prod := entities.Product{
		ID:        "4",
		Name:      "Prod",
		Price:     100,
		CreatedAt: now,
	}
	resp := ToProductResponse(prod)
	if resp.ID != prod.ID || resp.Name != prod.Name || resp.Price != prod.Price || resp.CreatedAt != now {
		t.Errorf("ToProductResponse failed to map correctly")
	}
}

func TestUserMapping(t *testing.T) {
	req := entities.RegisterUserRequest{
		Email:    "test@test.com",
		UserName: "tester",
		Phone:    "1234567890",
	}
	user, profile := ToUserEntity(req, "hashedpass")
	if user.Email != req.Email || user.PasswordHash != "hashedpass" {
		t.Errorf("ToUserEntity user mapping failed")
	}
	if profile.UserName != req.UserName || *profile.Phone != req.Phone {
		t.Errorf("ToUserEntity profile mapping failed")
	}

	req2 := entities.RegisterUserRequest{
		Email:    "test2@test.com",
		UserName: "tester2",
	}
	_, profile2 := ToUserEntity(req2, "hash")
	if profile2.Phone != nil {
		t.Errorf("expected phone to be nil")
	}

	resp := ToRegisterUserResponse(user, profile)
	if resp.Email != user.Email || resp.UserName != profile.UserName || resp.Phone != *profile.Phone {
		t.Errorf("ToRegisterUserResponse failed")
	}

	resp2 := ToRegisterUserResponse(user, profile2)
	if resp2.Phone != "" {
		t.Errorf("expected empty phone")
	}

	row := sqlc.CreateUserRow{
		ID:    1,
		Uuid:  uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Email: "test@test.com",
	}
	uEntity := ToUserEntityFromSqlc(row)
	if uEntity.ID != row.ID || uEntity.UUID != row.Uuid || uEntity.Email != row.Email {
		t.Errorf("ToUserEntityFromSqlc failed")
	}
}

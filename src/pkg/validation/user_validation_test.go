package validation

import (
	"testing"
	"github.com/go-playground/validator/v10"
)

func TestIsValidPhone(t *testing.T) {
	if !IsValidPhone("0123456789") {
		t.Error("expected true")
	}
	if IsValidPhone("0123") {
		t.Error("expected false for short phone")
	}
	if IsValidPhone("012345678a") {
		t.Error("expected false for non digits")
	}
}

type CustomValStruct struct {
	Email    string `validate:"email_advanced"`
	Password string `validate:"password_strong"`
	Phone    string `validate:"phone"`
	Slug     string `validate:"slug"`
	Regex    string `validate:"regex=^[a-z]+$"`
	Search   string `validate:"search"`
	MinInt   int    `validate:"min_int=5"`
	MaxInt   int    `validate:"max_int=10"`
	FileExt  string `validate:"file_ext=jpg png"`
}

func TestRegisterCustomValidation(t *testing.T) {
	v := validator.New()
	RegisterCustomValidation(v)

	tests := []struct {
		name    string
		st      CustomValStruct
		wantErr bool
	}{
		{"valid", CustomValStruct{
			Email: "test@gmail.com",
			Password: "Password123!",
			Phone: "+12345678901",
			Slug: "test-slug",
			Regex: "abc",
			Search: "test search 123",
			MinInt: 6,
			MaxInt: 9,
			FileExt: "image.jpg",
		}, false},
		{"invalid_email", CustomValStruct{Email: "test@abc.com", Password: "Password123!", Phone: "0123456789", Slug: "a", Regex: "a", Search: "a", MinInt: 5, MaxInt: 10, FileExt: "a.jpg"}, true},
		{"invalid_password", CustomValStruct{Email: "test@gmail.com", Password: "weak", Phone: "0123456789", Slug: "a", Regex: "a", Search: "a", MinInt: 5, MaxInt: 10, FileExt: "a.jpg"}, true},
		{"invalid_phone", CustomValStruct{Email: "test@gmail.com", Password: "Password123!", Phone: "invalid", Slug: "a", Regex: "a", Search: "a", MinInt: 5, MaxInt: 10, FileExt: "a.jpg"}, true},
		{"invalid_slug", CustomValStruct{Email: "test@gmail.com", Password: "Password123!", Phone: "0123456789", Slug: "Invalid Slug", Regex: "a", Search: "a", MinInt: 5, MaxInt: 10, FileExt: "a.jpg"}, true},
		{"invalid_regex", CustomValStruct{Email: "test@gmail.com", Password: "Password123!", Phone: "0123456789", Slug: "a", Regex: "123", Search: "a", MinInt: 5, MaxInt: 10, FileExt: "a.jpg"}, true},
		{"invalid_min", CustomValStruct{Email: "test@gmail.com", Password: "Password123!", Phone: "0123456789", Slug: "a", Regex: "a", Search: "a", MinInt: 3, MaxInt: 10, FileExt: "a.jpg"}, true},
		{"invalid_max", CustomValStruct{Email: "test@gmail.com", Password: "Password123!", Phone: "0123456789", Slug: "a", Regex: "a", Search: "a", MinInt: 5, MaxInt: 11, FileExt: "a.jpg"}, true},
		{"invalid_ext", CustomValStruct{Email: "test@gmail.com", Password: "Password123!", Phone: "0123456789", Slug: "a", Regex: "a", Search: "a", MinInt: 5, MaxInt: 10, FileExt: "a.gif"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.st)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

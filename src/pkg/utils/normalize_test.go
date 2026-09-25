package utils

import (
	"testing"
)

func TestNormalizeStrings(t *testing.T) {
	t.Run("trims spaces from all values", func(t *testing.T) {
		a := "  hello  "
		b := "  world  "
		NormalizeStrings(&a, &b)
		if a != "hello" {
			t.Errorf("a = %q; want %q", a, "hello")
		}
		if b != "world" {
			t.Errorf("b = %q; want %q", b, "world")
		}
	})

	t.Run("skips nil pointers", func(t *testing.T) {
		a := "  hello  "
		NormalizeStrings(nil, &a, nil)
		if a != "hello" {
			t.Errorf("a = %q; want %q", a, "hello")
		}
	})

	t.Run("handles empty variadic", func(t *testing.T) {
		// Should not panic.
		NormalizeStrings()
	})

	t.Run("preserves case", func(t *testing.T) {
		a := "  Hello World  "
		NormalizeStrings(&a)
		if a != "Hello World" {
			t.Errorf("a = %q; want %q", a, "Hello World")
		}
	})
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase and trim", "  User@Example.COM  ", "user@example.com"},
		{"already normalized", "user@example.com", "user@example.com"},
		{"only spaces", "   ", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := tt.input
			NormalizeEmail(&email)
			if email != tt.want {
				t.Errorf("NormalizeEmail(%q) = %q; want %q", tt.input, email, tt.want)
			}
		})
	}

	t.Run("nil pointer does not panic", func(t *testing.T) {
		NormalizeEmail(nil) // Should not panic.
	})
}

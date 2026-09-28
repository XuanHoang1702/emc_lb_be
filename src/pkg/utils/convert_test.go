package utils

import (
	"testing"
	"time"
)

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple camel", "CamelCase", "camel_case"},
		{"all caps word", "HTMLParser", "html_parser"},
		{"mixed", "getHTTPResponse", "get_http_response"},
		{"already snake", "already_snake", "already_snake"},
		{"single word", "hello", "hello"},
		{"single upper", "A", "a"},
		{"empty string", "", ""},
		{"acronym at end", "parseJSON", "parse_json"},
		{"consecutive acronyms", "XMLHTTPRequest", "xmlhttp_request"},
		{"with numbers", "field2Value", "field2_value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CamelToSnake(tt.input)
			if got != tt.want {
				t.Errorf("CamelToSnake(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"trims and lowercases", "  Hello World  ", "hello world"},
		{"already normalized", "hello", "hello"},
		{"all caps", "HELLO", "hello"},
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
		{"tabs and newlines", "\t hello \n", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeString(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeString(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvertToInt32Pointer(t *testing.T) {
	tests := []struct {
		name    string
		input   int32
		wantNil bool
		want    int32
	}{
		{"zero returns nil", 0, true, 0},
		{"positive returns pointer", 42, false, 42},
		{"negative returns pointer", -1, false, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToInt32Pointer(tt.input)
			if tt.wantNil {
				if got != nil {
					t.Errorf("ConvertToInt32Pointer(%d) = %v; want nil", tt.input, *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("ConvertToInt32Pointer(%d) = nil; want %d", tt.input, tt.want)
			}
			if *got != tt.want {
				t.Errorf("ConvertToInt32Pointer(%d) = %d; want %d", tt.input, *got, tt.want)
			}
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"lowercase first", "hello", "Hello"},
		{"already uppercase", "Hello", "Hello"},
		{"single char", "a", "A"},
		{"empty string", "", ""},
		{"all caps", "HELLO", "HELLO"},
		{"number first", "123abc", "123abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CapitalizeFirst(tt.input)
			if got != tt.want {
				t.Errorf("CapitalizeFirst(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToTimestamptz(t *testing.T) {
	now := time.Now()
	ts := ToTimestamptz(now)

	if !ts.Valid {
		t.Error("ToTimestamptz should set Valid = true")
	}
	if !ts.Time.Equal(now) {
		t.Errorf("ToTimestamptz time = %v; want %v", ts.Time, now)
	}
}

func TestToNumeric(t *testing.T) {
	n := ToNumeric(3.14)
	if !n.Valid {
		t.Error("ToNumeric(3.14) should produce valid Numeric")
	}
}

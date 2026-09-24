package pagination

import (
	"testing"
)

func TestOffsetRequest_Normalize(t *testing.T) {
	tests := []struct {
		name          string
		req           OffsetRequest
		expectedPage  int
		expectedLimit int
	}{
		{"zeros", OffsetRequest{}, 1, 20},
		{"valid", OffsetRequest{Page: 2, Limit: 50}, 2, 50},
		{"negative page", OffsetRequest{Page: -1, Limit: 10}, 1, 10},
		{"over max limit", OffsetRequest{Page: 1, Limit: 200}, 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.Normalize()
			if tt.req.Page != tt.expectedPage {
				t.Errorf("expected page %d, got %d", tt.expectedPage, tt.req.Page)
			}
			if tt.req.Limit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, tt.req.Limit)
			}
		})
	}
}

func TestOffsetRequest_Offset(t *testing.T) {
	req := OffsetRequest{Page: 3, Limit: 10}
	if req.Offset() != 20 {
		t.Errorf("expected 20, got %d", req.Offset())
	}
}

func TestNewOffsetMeta(t *testing.T) {
	meta := NewOffsetMeta(55, 2, 10)
	if meta.TotalPages != 6 {
		t.Errorf("expected 6 total pages, got %d", meta.TotalPages)
	}
}

func TestCursorRequest_Normalize(t *testing.T) {
	req := CursorRequest{Limit: 0}
	req.Normalize()
	if req.Limit != 20 {
		t.Errorf("expected limit 20, got %d", req.Limit)
	}
}

func TestEncodeDecodeCursor(t *testing.T) {
	raw := "cursor_value_123"
	encoded := EncodeCursor(raw)
	decoded, ok := DecodeCursor(encoded)
	if !ok {
		t.Fatal("expected DecodeCursor to succeed")
	}
	if decoded != raw {
		t.Errorf("expected %s, got %s", raw, decoded)
	}

	decodedEmpty, okEmpty := DecodeCursor("")
	if !okEmpty || decodedEmpty != "" {
		t.Error("expected empty string handling")
	}

	_, okBad := DecodeCursor("!@#$")
	if okBad {
		t.Error("expected failure on invalid base64")
	}
}

func TestNewOffsetPagedResponse(t *testing.T) {
	items := []int{1, 2, 3}
	resp := NewOffsetPagedResponse(items, 100, 1, 10)
	if len(resp.Data) != 3 {
		t.Errorf("expected 3 items, got %d", len(resp.Data))
	}
	meta, ok := resp.Pagination.(OffsetMeta)
	if !ok {
		t.Fatal("expected OffsetMeta")
	}
	if meta.TotalPages != 10 {
		t.Errorf("expected 10 total pages, got %d", meta.TotalPages)
	}
}

func TestNewCursorPagedResponse(t *testing.T) {
	items := []string{"a", "b", "c"}
	resp := NewCursorPagedResponse(items, 2, func(s string) string { return s })
	
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Data))
	}
	meta, ok := resp.Pagination.(CursorMeta)
	if !ok {
		t.Fatal("expected CursorMeta")
	}
	if !meta.HasMore {
		t.Error("expected HasMore true")
	}
	if meta.NextCursor != EncodeCursor("b") {
		t.Errorf("expected %s, got %s", EncodeCursor("b"), meta.NextCursor)
	}
}

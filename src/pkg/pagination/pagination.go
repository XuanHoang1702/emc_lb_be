// Package pagination provides standardized helpers for offset-based and
// cursor-based pagination used across all API endpoints.
package pagination

import (
	"encoding/base64"
	"math"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
	MinLimit     = 1
	MinPage      = 1
)

// ---- Offset-based pagination ----

// OffsetRequest is the query-param binding struct for offset pagination.
type OffsetRequest struct {
	Page  int `form:"page"  binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

// Normalize applies defaults when zero values are passed.
func (r *OffsetRequest) Normalize() {
	if r.Page < MinPage {
		r.Page = MinPage
	}

	if r.Limit < MinLimit || r.Limit > MaxLimit {
		r.Limit = DefaultLimit
	}
}

// Offset returns the SQL OFFSET value.
func (r *OffsetRequest) Offset() int {
	return (r.Page - 1) * r.Limit
}

// OffsetMeta is the pagination metadata included in list responses.
type OffsetMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// NewOffsetMeta constructs OffsetMeta from total count and the current request.
func NewOffsetMeta(total, page, limit int) OffsetMeta {
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return OffsetMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// ---- Cursor-based pagination ----

// CursorRequest is the query-param binding struct for cursor pagination.
type CursorRequest struct {
	Cursor string `form:"cursor"  binding:"omitempty"`
	Limit  int    `form:"limit"   binding:"omitempty,min=1,max=100"`
}

// Normalize applies defaults when zero values are passed.
func (r *CursorRequest) Normalize() {
	if r.Limit < MinLimit || r.Limit > MaxLimit {
		r.Limit = DefaultLimit
	}
}

// CursorMeta is the pagination metadata for cursor-based responses.
type CursorMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// EncodeCursor base64-encodes an opaque cursor value (e.g. an ID or timestamp string).
func EncodeCursor(raw string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor decodes a base64 cursor back to its raw value.
// Returns the original string and true on success, empty string and false on failure.
func DecodeCursor(encoded string) (string, bool) {
	if encoded == "" {
		return "", true
	}

	b, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", false
	}

	return string(b), true
}

// ---- Generic response wrapper ----

// PagedResponse is the standard list response envelope.
//
// Example usage:
//
//	res.Success(ctx, http.StatusOK, pagination.PagedResponse[ProductDTO]{
//	    Data:       products,
//	    Pagination: pagination.NewOffsetMeta(total, req.Page, req.Limit),
//	})
type PagedResponse[T any] struct {
	Data       []T `json:"data"`
	Pagination any `json:"pagination"`
}

// NewOffsetPagedResponse constructs a PagedResponse with OffsetMeta.
func NewOffsetPagedResponse[T any](items []T, total, page, limit int) PagedResponse[T] {
	return PagedResponse[T]{
		Data:       items,
		Pagination: NewOffsetMeta(total, page, limit),
	}
}

// NewCursorPagedResponse constructs a PagedResponse with CursorMeta.
// keyFn extracts the cursor key from the last item.
func NewCursorPagedResponse[T any](items []T, limit int, keyFn func(T) string) PagedResponse[T] {
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	var nextCursor string
	if hasMore && len(items) > 0 {
		nextCursor = EncodeCursor(keyFn(items[len(items)-1]))
	}

	return PagedResponse[T]{
		Data: items,
		Pagination: CursorMeta{
			NextCursor: nextCursor,
			HasMore:    hasMore,
		},
	}
}

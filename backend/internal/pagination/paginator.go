package pagination

import (
	"math"
	"net/http"
	"strconv"
)

// Params holds parsed pagination settings
type Params struct {
	Page    int
	Limit   int
	Offset  int
	Enabled bool
}

// Meta holds pagination metadata returned in API responses
type Meta struct {
	TotalCount int `json:"total_count"`
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Parse extracts page and limit query parameters from an HTTP request
func Parse(r *http.Request) Params {
	q := r.URL.Query()
	pageStr := q.Get("page")
	limitStr := q.Get("limit")

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 0
	enabled := false
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
		enabled = true
		if limit > 500 {
			limit = 500
		}
	} else if pageStr != "" {
		limit = 50
		enabled = true
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	return Params{
		Page:    page,
		Limit:   limit,
		Offset:  offset,
		Enabled: enabled,
	}
}

// PaginateSlice slices a dataset in-memory and computes accurate metadata
func PaginateSlice[T any](items []T, params Params) ([]T, Meta) {
	total := len(items)
	if !params.Enabled || params.Limit <= 0 {
		return items, Meta{TotalCount: total}
	}

	start := params.Offset
	if start >= total {
		totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))
		if totalPages == 0 {
			totalPages = 1
		}
		return []T{}, Meta{
			TotalCount: total,
			Page:       params.Page,
			Limit:      params.Limit,
			TotalPages: totalPages,
		}
	}

	end := start + params.Limit
	if end > total {
		end = total
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return items[start:end], Meta{
		TotalCount: total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}
}

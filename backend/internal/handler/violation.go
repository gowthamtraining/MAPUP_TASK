package handler

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"mapup-geofence/internal/service"
)

type ViolationHandler struct {
	service *service.AlertService
}

func NewViolationHandler(service *service.AlertService) *ViolationHandler {
	return &ViolationHandler{service: service}
}

// GetViolationHistory handles GET /violations/history?vehicle_id=...&geofence_id=...&start_date=...&end_date=...&limit=...&page=...
func (h *ViolationHandler) GetViolationHistory(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	q := r.URL.Query()
	vehicleID := q.Get("vehicle_id")
	geofenceID := q.Get("geofence_id")
	startDateStr := q.Get("start_date")
	endDateStr := q.Get("end_date")
	limitStr := q.Get("limit")
	pageStr := q.Get("page")

	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if limit > 500 {
		limit = 500
	}

	page := 1
	if pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	var startDate *time.Time
	if startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &t
		}
	}

	var endDate *time.Time
	if endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &t
		}
	}

	violations, totalCount, err := h.service.GetViolations(r.Context(), vehicleID, geofenceID, startDate, endDate, limit)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error(), start)
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	response := map[string]interface{}{
		"violations":  violations,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	}

	RespondJSON(w, http.StatusOK, response, start)
}

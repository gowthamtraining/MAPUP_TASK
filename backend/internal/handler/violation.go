package handler

import (
	"mapup-geofence/internal/service"
	"net/http"
	"strconv"
	"time"
)

type ViolationHandler struct {
	service *service.AlertService
}

func NewViolationHandler(service *service.AlertService) *ViolationHandler {
	return &ViolationHandler{service: service}
}

// GetViolationHistory handles GET /violations/history
func (h *ViolationHandler) GetViolationHistory(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	q := r.URL.Query()
	vehicleID := q.Get("vehicle_id")
	geofenceID := q.Get("geofence_id")
	startDateStr := q.Get("start_date")
	endDateStr := q.Get("end_date")
	limitStr := q.Get("limit")

	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
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

	response := map[string]interface{}{
		"violations":  violations,
		"total_count": totalCount,
	}

	RespondJSON(w, http.StatusOK, response, start)
}

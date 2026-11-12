package handler

import (
	"encoding/json"
	"mapup-geofence/internal/pagination"
	"mapup-geofence/internal/service"
	"net/http"
	"time"
)

type AlertHandler struct {
	service *service.AlertService
}

func NewAlertHandler(service *service.AlertService) *AlertHandler {
	return &AlertHandler{service: service}
}

// ConfigureAlert handles POST /alerts/configure
func (h *AlertHandler) ConfigureAlert(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req service.ConfigureAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid JSON payload: "+err.Error(), start)
		return
	}

	res, err := h.service.ConfigureAlert(r.Context(), req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error(), start)
		return
	}

	RespondJSON(w, http.StatusCreated, res, start)
}

// GetAlerts handles GET /alerts?geofence_id=...&vehicle_id=...&page=...&limit=...
func (h *AlertHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	geofenceID := r.URL.Query().Get("geofence_id")
	vehicleID := r.URL.Query().Get("vehicle_id")

	list, err := h.service.GetAlerts(r.Context(), geofenceID, vehicleID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error(), start)
		return
	}

	params := pagination.Parse(r)
	paginatedList, meta := pagination.PaginateSlice(list, params)

	response := map[string]interface{}{
		"alerts": paginatedList,
	}
	if params.Enabled {
		response["total_count"] = meta.TotalCount
		response["page"] = meta.Page
		response["limit"] = meta.Limit
		response["total_pages"] = meta.TotalPages
	}

	RespondJSON(w, http.StatusOK, response, start)
}

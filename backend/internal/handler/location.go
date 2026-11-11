package handler

import (
	"encoding/json"
	"mapup-geofence/internal/service"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type LocationHandler struct {
	service *service.TrackerService
}

func NewLocationHandler(service *service.TrackerService) *LocationHandler {
	return &LocationHandler{service: service}
}

// UpdateLocation handles POST /vehicles/location
func (h *LocationHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req service.LocationUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid JSON payload: "+err.Error(), start)
		return
	}

	res, err := h.service.ProcessLocation(r.Context(), req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error(), start)
		return
	}

	RespondJSON(w, http.StatusOK, res, start)
}

// GetVehicleLocation handles GET /vehicles/location/{vehicle_id}
func (h *LocationHandler) GetVehicleLocation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	vehicleID := chi.URLParam(r, "vehicle_id")
	if vehicleID == "" {
		RespondError(w, http.StatusBadRequest, "vehicle_id is required in URL path", start)
		return
	}

	res, err := h.service.GetVehicleLocationAndStatus(r.Context(), vehicleID)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error(), start)
		return
	}

	RespondJSON(w, http.StatusOK, res, start)
}

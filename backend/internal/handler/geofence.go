package handler

import (
	"encoding/json"
	"mapup-geofence/internal/geospatial"
	"mapup-geofence/internal/service"
	"net/http"
	"time"
)

type GeofenceHandler struct {
	service *service.GeofenceService
}

func NewGeofenceHandler(service *service.GeofenceService) *GeofenceHandler {
	return &GeofenceHandler{service: service}
}

// CreateGeofence handles POST /geofences
func (h *GeofenceHandler) CreateGeofence(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req service.CreateGeofenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid JSON payload: "+err.Error(), start)
		return
	}

	res, err := h.service.CreateGeofence(r.Context(), req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error(), start)
		return
	}

	RespondJSON(w, http.StatusCreated, res, start)
}

// GetGeofences handles GET /geofences?category=...
func (h *GeofenceHandler) GetGeofences(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	category := r.URL.Query().Get("category")
	if category != "" {
		if err := geospatial.ValidateCategory(category); err != nil {
			RespondError(w, http.StatusBadRequest, err.Error(), start)
			return
		}
	}

	list, err := h.service.GetGeofences(r.Context(), category)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error(), start)
		return
	}

	response := map[string]interface{}{
		"geofences": list,
	}

	RespondJSON(w, http.StatusOK, response, start)
}

package handler

import (
	"encoding/json"
	"mapup-geofence/internal/service"
	"net/http"
	"time"
)

type VehicleHandler struct {
	service *service.VehicleService
}

func NewVehicleHandler(service *service.VehicleService) *VehicleHandler {
	return &VehicleHandler{service: service}
}

// CreateVehicle handles POST /vehicles
func (h *VehicleHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req service.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid JSON payload: "+err.Error(), start)
		return
	}

	res, err := h.service.CreateVehicle(r.Context(), req)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error(), start)
		return
	}

	RespondJSON(w, http.StatusCreated, res, start)
}

// GetVehicles handles GET /vehicles
func (h *VehicleHandler) GetVehicles(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	list, err := h.service.GetVehicles(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error(), start)
		return
	}

	response := map[string]interface{}{
		"vehicles": list,
	}

	RespondJSON(w, http.StatusOK, response, start)
}

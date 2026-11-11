package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"mapup-geofence/internal/handler"
	"mapup-geofence/internal/repository"
	"mapup-geofence/internal/service"
	"mapup-geofence/internal/websocket"
)

func setupTestApp() http.Handler {
	repo := repository.NewMemoryRepository()
	hub := websocket.NewHub()
	go hub.Run()

	geofenceService := service.NewGeofenceService(repo)
	vehicleService := service.NewVehicleService(repo)
	alertService := service.NewAlertService(repo)
	trackerService := service.NewTrackerService(repo, hub)

	geofenceHandler := handler.NewGeofenceHandler(geofenceService)
	vehicleHandler := handler.NewVehicleHandler(vehicleService)
	locationHandler := handler.NewLocationHandler(trackerService)
	alertHandler := handler.NewAlertHandler(alertService)
	violationHandler := handler.NewViolationHandler(alertService)

	r := chi.NewRouter()
	r.Post("/geofences", geofenceHandler.CreateGeofence)
	r.Get("/geofences", geofenceHandler.GetGeofences)
	r.Post("/vehicles", vehicleHandler.CreateVehicle)
	r.Get("/vehicles", vehicleHandler.GetVehicles)
	r.Post("/vehicles/location", locationHandler.UpdateLocation)
	r.Get("/vehicles/location/{vehicle_id}", locationHandler.GetVehicleLocation)
	r.Post("/alerts/configure", alertHandler.ConfigureAlert)
	r.Get("/alerts", alertHandler.GetAlerts)
	r.Get("/violations/history", violationHandler.GetViolationHistory)

	return r
}

func TestEndToEndAPI(t *testing.T) {
	app := setupTestApp()

	// 1. Create a Geofence
	geoPayload := map[string]interface{}{
		"name":        "Downtown Delivery Zone",
		"description": "Main delivery area for downtown customers",
		"coordinates": [][]float64{
			{37.7749, -122.4194},
			{37.7849, -122.4194},
			{37.7849, -122.4094},
			{37.7749, -122.4094},
			{37.7749, -122.4194},
		},
		"category": "delivery_zone",
	}
	body, _ := json.Marshal(geoPayload)
	req := httptest.NewRequest("POST", "/geofences", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var geoResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &geoResp)
	if geoResp["time_ns"] == nil || geoResp["time_ns"] == "" {
		t.Fatalf("time_ns missing in response: %v", geoResp)
	}
	geoID := geoResp["id"].(string)

	// 2. Register a Vehicle
	vehPayload := map[string]interface{}{
		"vehicle_number": "KA-01-AB-1234",
		"driver_name":    "John Doe",
		"vehicle_type":   "truck",
		"phone":          "+1234567890",
	}
	body, _ = json.Marshal(vehPayload)
	req = httptest.NewRequest("POST", "/vehicles", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var vehResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &vehResp)
	if vehResp["time_ns"] == nil {
		t.Fatalf("time_ns missing in vehicle response: %v", vehResp)
	}
	vehID := vehResp["id"].(string)

	// 3. Configure Alert Rule for Entry
	alertPayload := map[string]interface{}{
		"geofence_id": geoID,
		"vehicle_id":  vehID,
		"event_type":  "entry",
	}
	body, _ = json.Marshal(alertPayload)
	req = httptest.NewRequest("POST", "/alerts/configure", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Update Vehicle Location INSIDE Geofence (Entry Event!)
	locPayload := map[string]interface{}{
		"vehicle_id": vehID,
		"latitude":   37.7800,
		"longitude":  -122.4150,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}
	body, _ = json.Marshal(locPayload)
	req = httptest.NewRequest("POST", "/vehicles/location", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var locResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &locResp)
	if locResp["time_ns"] == nil {
		t.Fatalf("time_ns missing in location response: %v", locResp)
	}
	geofencesInside := locResp["current_geofences"].([]interface{})
	if len(geofencesInside) != 1 {
		t.Fatalf("expected 1 geofence inside, got %d", len(geofencesInside))
	}

	// 5. Query Violations History
	req = httptest.NewRequest("GET", "/violations/history?vehicle_id="+vehID, nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for violations history, got %d", w.Code)
	}

	var violResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &violResp)
	if violResp["time_ns"] == nil {
		t.Fatalf("time_ns missing in violations response")
	}
	violations := violResp["violations"].([]interface{})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation recorded for entry, got %d", len(violations))
	}

	// 6. Test Error Response contains time_ns
	req = httptest.NewRequest("POST", "/geofences", bytes.NewBufferString("{invalid json"))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", w.Code)
	}
	var errResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp["time_ns"] == nil || errResp["error"] == nil {
		t.Fatalf("error response must contain time_ns and error message: %v", errResp)
	}
}

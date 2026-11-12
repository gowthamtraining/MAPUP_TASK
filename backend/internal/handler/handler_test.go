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

func TestAllRequiredEndpoints(t *testing.T) {
	app := setupTestApp()

	// ==========================================
	// 1. POST /geofences
	// ==========================================
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
		t.Fatalf("expected 201 Created for POST /geofences, got %d: %s", w.Code, w.Body.String())
	}
	var geoResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &geoResp)
	if geoResp["time_ns"] == nil || geoResp["time_ns"] == "" {
		t.Fatalf("time_ns missing in POST /geofences response")
	}
	geoID := geoResp["id"].(string)

	// ==========================================
	// 2. GET /geofences (with & without category filter)
	// ==========================================
	req = httptest.NewRequest("GET", "/geofences?category=delivery_zone", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /geofences?category=delivery_zone, got %d", w.Code)
	}
	var catResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &catResp)
	if catResp["time_ns"] == nil {
		t.Fatalf("time_ns missing in GET /geofences response")
	}
	filteredList := catResp["geofences"].([]interface{})
	if len(filteredList) != 1 {
		t.Fatalf("expected 1 delivery_zone geofence, got %d", len(filteredList))
	}

	// Filter by non-matching category
	req = httptest.NewRequest("GET", "/geofences?category=restricted_zone", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /geofences?category=restricted_zone, got %d", w.Code)
	}
	_ = json.Unmarshal(w.Body.Bytes(), &catResp)
	if len(catResp["geofences"].([]interface{})) != 0 {
		t.Fatalf("expected 0 restricted_zone geofences")
	}

	// ==========================================
	// 3. POST /vehicles
	// ==========================================
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
		t.Fatalf("expected 201 Created for POST /vehicles, got %d", w.Code)
	}
	var vehResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &vehResp)
	if vehResp["time_ns"] == nil {
		t.Fatalf("time_ns missing in POST /vehicles response")
	}
	vehID := vehResp["id"].(string)

	// ==========================================
	// 4. GET /vehicles
	// ==========================================
	req = httptest.NewRequest("GET", "/vehicles", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /vehicles, got %d", w.Code)
	}
	var vehiclesListResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &vehiclesListResp)
	if vehiclesListResp["time_ns"] == nil {
		t.Fatalf("time_ns missing in GET /vehicles response")
	}
	if len(vehiclesListResp["vehicles"].([]interface{})) != 1 {
		t.Fatalf("expected 1 vehicle in GET /vehicles")
	}

	// ==========================================
	// 5. POST /alerts/configure
	// ==========================================
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
		t.Fatalf("expected 201 Created for POST /alerts/configure, got %d", w.Code)
	}
	var alertResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &alertResp)
	if alertResp["time_ns"] == nil || alertResp["alert_id"] == nil {
		t.Fatalf("time_ns or alert_id missing in POST /alerts/configure response: %v", alertResp)
	}

	// ==========================================
	// 6. GET /alerts
	// ==========================================
	req = httptest.NewRequest("GET", "/alerts?geofence_id="+geoID+"&vehicle_id="+vehID, nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /alerts, got %d", w.Code)
	}
	var alertsListResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &alertsListResp)
	if alertsListResp["time_ns"] == nil {
		t.Fatalf("time_ns missing in GET /alerts response")
	}
	if len(alertsListResp["alerts"].([]interface{})) != 1 {
		t.Fatalf("expected 1 alert in GET /alerts, got %v", alertsListResp)
	}

	// ==========================================
	// 7. POST /vehicles/location (Trigger Entry Event!)
	// ==========================================
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
		t.Fatalf("expected 200 OK for POST /vehicles/location, got %d", w.Code)
	}
	var locResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &locResp)
	if locResp["time_ns"] == nil || locResp["location_updated"] != true {
		t.Fatalf("invalid POST /vehicles/location response: %v", locResp)
	}

	// ==========================================
	// 8. GET /vehicles/location/{vehicle_id}
	// ==========================================
	req = httptest.NewRequest("GET", "/vehicles/location/"+vehID, nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /vehicles/location/{vehicle_id}, got %d", w.Code)
	}
	var vehLocResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &vehLocResp)
	if vehLocResp["time_ns"] == nil || vehLocResp["vehicle_number"] != "KA-01-AB-1234" {
		t.Fatalf("invalid GET /vehicles/location response: %v", vehLocResp)
	}

	// ==========================================
	// 9. GET /violations/history
	// ==========================================
	req = httptest.NewRequest("GET", "/violations/history?vehicle_id="+vehID+"&limit=50", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /violations/history, got %d", w.Code)
	}
	var violResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &violResp)
	if violResp["time_ns"] == nil || violResp["total_count"] == nil {
		t.Fatalf("time_ns or total_count missing in violations response: %v", violResp)
	}
	violations := violResp["violations"].([]interface{})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation recorded for entry event, got %d", len(violations))
	}
	v0 := violations[0].(map[string]interface{})
	if v0["event_type"] != "entry" || v0["geofence_id"] != geoID {
		t.Fatalf("violation fields mismatch: %v", v0)
	}

	// ==========================================
	// 10. Verify Error Responses Include time_ns
	// ==========================================
	req = httptest.NewRequest("POST", "/geofences", bytes.NewBufferString("{bad json"))
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", w.Code)
	}
	var errResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp["time_ns"] == nil || errResp["error"] == nil {
		t.Fatalf("error response must contain time_ns: %v", errResp)
	}
}

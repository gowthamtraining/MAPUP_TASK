package service_test

import (
	"context"
	"testing"
	"time"

	"mapup-geofence/internal/domain"
	"mapup-geofence/internal/repository"
	"mapup-geofence/internal/service"
	"mapup-geofence/internal/websocket"
)

func setupTestServices() (*service.TrackerService, repository.Repository, *websocket.Hub) {
	repo := repository.NewMemoryRepository()
	hub := websocket.NewHub()
	go hub.Run()
	tracker := service.NewTrackerService(repo, hub)
	return tracker, repo, hub
}

func TestTrackerService_TransitionsAndRestrictedAlerts(t *testing.T) {
	ctx := context.Background()
	tracker, repo, _ := setupTestServices()

	// 1. Create a Restricted Zone
	restrictedZone := &domain.Geofence{
		ID:          "geo_restricted_test",
		Name:        "Nuclear Plant Perimeter",
		Category:    domain.CategoryRestrictedZone,
		Status:      "active",
		CreatedAt:   time.Now().UTC(),
		Coordinates: []domain.Coordinate{
			{10.0, 10.0},
			{11.0, 10.0},
			{11.0, 11.0},
			{10.0, 11.0},
			{10.0, 10.0},
		},
	}
	_ = repo.CreateGeofence(ctx, restrictedZone)

	// 2. Register a Vehicle
	veh := &domain.Vehicle{
		ID:            "veh_001",
		VehicleNumber: "TEST-VEH-01",
		DriverName:    "Alice",
		VehicleType:   "truck",
		Phone:         "+123456789",
		Status:        "active",
		CreatedAt:     time.Now().UTC(),
	}
	_ = repo.CreateVehicle(ctx, veh)

	// Ping 1: Vehicle is outside (5.0, 5.0)
	res1, err := tracker.ProcessLocation(ctx, service.LocationUpdateRequest{
		VehicleID: veh.ID,
		Latitude:  5.0,
		Longitude: 5.0,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("unexpected error on ping 1: %v", err)
	}
	if len(res1.CurrentGeofences) != 0 {
		t.Fatalf("expected vehicle outside all zones, got %d", len(res1.CurrentGeofences))
	}

	// Verify no violations recorded yet
	violations, count, _ := repo.GetViolations(ctx, veh.ID, "", nil, nil, 10)
	if count != 0 {
		t.Fatalf("expected 0 violations outside, got %d", count)
	}

	// Ping 2: Vehicle ENTERS Restricted Zone (10.5, 10.5)
	res2, err := tracker.ProcessLocation(ctx, service.LocationUpdateRequest{
		VehicleID: veh.ID,
		Latitude:  10.5,
		Longitude: 10.5,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("unexpected error on ping 2: %v", err)
	}
	if len(res2.CurrentGeofences) != 1 {
		t.Fatalf("expected vehicle inside 1 zone, got %d", len(res2.CurrentGeofences))
	}

	// Verify Restricted Zone breach alert was automatically created in violations
	violations, count, _ = repo.GetViolations(ctx, veh.ID, "", nil, nil, 10)
	if count != 1 {
		t.Fatalf("expected 1 breach violation for restricted zone entry, got %d", count)
	}
	if violations[0].EventType != domain.EventTypeEntry {
		t.Fatalf("expected event_type 'entry', got '%s'", violations[0].EventType)
	}

	// Ping 3: Vehicle stays inside Restricted Zone (10.6, 10.6) - Should NOT duplicate alert!
	_, err = tracker.ProcessLocation(ctx, service.LocationUpdateRequest{
		VehicleID: veh.ID,
		Latitude:  10.6,
		Longitude: 10.6,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("unexpected error on ping 3: %v", err)
	}

	violations, count, _ = repo.GetViolations(ctx, veh.ID, "", nil, nil, 10)
	if count != 1 {
		t.Fatalf("expected still 1 violation (no duplicate entry event), got %d", count)
	}

	// Ping 4: Vehicle EXITS Restricted Zone (5.0, 5.0)
	res4, err := tracker.ProcessLocation(ctx, service.LocationUpdateRequest{
		VehicleID: veh.ID,
		Latitude:  5.0,
		Longitude: 5.0,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("unexpected error on ping 4: %v", err)
	}
	if len(res4.CurrentGeofences) != 0 {
		t.Fatalf("expected 0 zones after exit, got %d", len(res4.CurrentGeofences))
	}
}

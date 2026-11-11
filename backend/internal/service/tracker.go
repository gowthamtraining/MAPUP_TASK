package service

import (
	"context"
	"fmt"
	"log"
	"mapup-geofence/internal/domain"
	"mapup-geofence/internal/geospatial"
	"mapup-geofence/internal/repository"
	"mapup-geofence/internal/websocket"
	"time"

	"github.com/google/uuid"
)

type TrackerService struct {
	repo repository.Repository
	hub  *websocket.Hub
}

func NewTrackerService(repo repository.Repository, hub *websocket.Hub) *TrackerService {
	return &TrackerService{
		repo: repo,
		hub:  hub,
	}
}

type LocationUpdateRequest struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp string  `json:"timestamp"`
}

type LocationUpdateResponse struct {
	VehicleID        string                         `json:"vehicle_id"`
	LocationUpdated  bool                           `json:"location_updated"`
	CurrentGeofences []domain.CurrentGeofenceStatus `json:"current_geofences"`
}

type VehicleLocationDetailResponse struct {
	VehicleID        string                         `json:"vehicle_id"`
	VehicleNumber    string                         `json:"vehicle_number"`
	CurrentLocation  CurrentLocationDetail          `json:"current_location"`
	CurrentGeofences []domain.VehicleGeofenceDetail `json:"current_geofences"`
}

type CurrentLocationDetail struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp string  `json:"timestamp"`
}

// ProcessLocation handles the location ingestion, containment calculation, transition detection, and async alert dispatch
func (s *TrackerService) ProcessLocation(ctx context.Context, req LocationUpdateRequest) (*LocationUpdateResponse, error) {
	if req.VehicleID == "" {
		return nil, fmt.Errorf("vehicle_id is required")
	}
	if req.Latitude < -90.0 || req.Latitude > 90.0 {
		return nil, fmt.Errorf("latitude must be between -90 and 90")
	}
	if req.Longitude < -180.0 || req.Longitude > 180.0 {
		return nil, fmt.Errorf("longitude must be between -180 and 180")
	}

	ts, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		ts = time.Now().UTC()
	}

	vehicle, err := s.repo.GetVehicleByID(ctx, req.VehicleID)
	if err != nil {
		return nil, fmt.Errorf("vehicle not found: %w", err)
	}

	// 1. Store location update in database
	locRecord := &domain.LocationRecord{
		VehicleID: req.VehicleID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Timestamp: ts,
	}
	if err := s.repo.RecordLocation(ctx, locRecord); err != nil {
		return nil, fmt.Errorf("failed to record location: %w", err)
	}

	// 2. Fetch all active geofences
	geofences, err := s.repo.GetAllActiveGeofences(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve active geofences: %w", err)
	}

	// 3. Check containment for each geofence
	currentInsideGeofences := make([]domain.Geofence, 0)
	currentInsideIDs := make([]string, 0)
	currentGeofenceStatuses := make([]domain.CurrentGeofenceStatus, 0)

	for _, g := range geofences {
		if geospatial.ContainsPoint(g.Coordinates, req.Latitude, req.Longitude) {
			currentInsideGeofences = append(currentInsideGeofences, g)
			currentInsideIDs = append(currentInsideIDs, g.ID)
			currentGeofenceStatuses = append(currentGeofenceStatuses, domain.CurrentGeofenceStatus{
				GeofenceID:   g.ID,
				GeofenceName: g.Name,
				Status:       "inside",
			})
		}
	}

	// 4. Detect entry/exit transitions by comparing with previous location state
	prevIDs, err := s.repo.GetVehicleCurrentGeofenceIDs(ctx, req.VehicleID)
	if err != nil {
		prevIDs = []string{}
	}

	prevSet := make(map[string]bool)
	for _, id := range prevIDs {
		prevSet[id] = true
	}

	currSet := make(map[string]bool)
	for _, id := range currentInsideIDs {
		currSet[id] = true
	}

	// Find entries: currently inside, but was not inside previously
	entries := make([]domain.Geofence, 0)
	for _, g := range currentInsideGeofences {
		if !prevSet[g.ID] {
			entries = append(entries, g)
		}
	}

	// Find exits: was inside previously, but not inside now
	exits := make([]domain.Geofence, 0)
	for _, prevID := range prevIDs {
		if !currSet[prevID] {
			// Find geofence entity
			for _, g := range geofences {
				if g.ID == prevID {
					exits = append(exits, g)
					break
				}
			}
		}
	}

	// Update vehicle's current geofence state in DB
	_ = s.repo.UpdateVehicleCurrentGeofenceIDs(ctx, req.VehicleID, currentInsideIDs)

	// 5. Evaluate alert rules & trigger real-time alerts asynchronously
	allRules, _ := s.repo.GetAllAlertRules(ctx)

	// Process Entries
	for _, g := range entries {
		s.handleTransitionEvent(ctx, vehicle, g, domain.EventTypeEntry, req.Latitude, req.Longitude, ts, allRules)
	}

	// Process Exits
	for _, g := range exits {
		s.handleTransitionEvent(ctx, vehicle, g, domain.EventTypeExit, req.Latitude, req.Longitude, ts, allRules)
	}

	return &LocationUpdateResponse{
		VehicleID:        req.VehicleID,
		LocationUpdated:  true,
		CurrentGeofences: currentGeofenceStatuses,
	}, nil
}

// handleTransitionEvent evaluates matching rules, records violation, and asynchronously broadcasts alert
func (s *TrackerService) handleTransitionEvent(
	ctx context.Context,
	vehicle *domain.Vehicle,
	geofence domain.Geofence,
	eventType string,
	lat, lon float64,
	ts time.Time,
	rules []domain.AlertRule,
) {
	// Rule matching logic:
	// 1. Mandatory alert for restricted zones on entry
	// 2. Or configured alert rule matching this geofence, event type, and vehicle
	shouldAlert := false

	if geofence.Category == domain.CategoryRestrictedZone && eventType == domain.EventTypeEntry {
		shouldAlert = true
	}

	for _, rule := range rules {
		if rule.Status != "active" {
			continue
		}
		if rule.GeofenceID != geofence.ID {
			continue
		}
		// Vehicle check: if rule vehicle is set, must match; if nil/empty, matches all vehicles
		if rule.VehicleID != nil && *rule.VehicleID != "" && *rule.VehicleID != vehicle.ID {
			continue
		}
		// Event check: entry, exit, or both
		if rule.EventType == domain.EventTypeBoth || rule.EventType == eventType {
			shouldAlert = true
			break
		}
	}

	if !shouldAlert {
		return
	}

	// Generate record IDs
	shortID := uuid.New().String()[:8]
	violID := fmt.Sprintf("viol_%s", shortID)
	evtID := fmt.Sprintf("evt_%s", shortID)

	violation := domain.Violation{
		ID:            violID,
		VehicleID:     vehicle.ID,
		VehicleNumber: vehicle.VehicleNumber,
		GeofenceID:    geofence.ID,
		GeofenceName:  geofence.Name,
		EventType:     eventType,
		Latitude:      lat,
		Longitude:     lon,
		Timestamp:     ts,
	}

	// Store all alerts in the database for historical records
	if err := s.repo.RecordViolation(ctx, &violation); err != nil {
		log.Printf("failed to record violation: %v", err)
	}

	// Asynchronously deliver the alert to connected clients in real-time
	wsAlert := domain.WebSocketAlertMessage{
		EventID:   evtID,
		EventType: eventType,
		Timestamp: ts.Format(time.RFC3339),
		Vehicle: domain.VehicleAlertSummary{
			VehicleID:     vehicle.ID,
			VehicleNumber: vehicle.VehicleNumber,
			DriverName:    vehicle.DriverName,
		},
		Geofence: domain.GeofenceAlertSummary{
			GeofenceID:   geofence.ID,
			GeofenceName: geofence.Name,
			Category:     geofence.Category,
		},
		Location: domain.LocationAlertSummary{
			Latitude:  lat,
			Longitude: lon,
		},
	}

	go func(alert domain.WebSocketAlertMessage) {
		s.hub.BroadcastAlert(alert)
	}(wsAlert)
}

// GetVehicleLocationAndStatus retrieves current location and geofence status for GET /vehicles/location/{vehicle_id}
func (s *TrackerService) GetVehicleLocationAndStatus(ctx context.Context, vehicleID string) (*VehicleLocationDetailResponse, error) {
	vehicle, err := s.repo.GetVehicleByID(ctx, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("vehicle not found: %w", err)
	}

	loc, err := s.repo.GetLatestLocation(ctx, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("location not found for vehicle: %w", err)
	}

	// Find current geofences
	geofenceIDs, _ := s.repo.GetVehicleCurrentGeofenceIDs(ctx, vehicleID)
	currentGeofences := make([]domain.VehicleGeofenceDetail, 0)

	for _, gID := range geofenceIDs {
		if g, err := s.repo.GetGeofenceByID(ctx, gID); err == nil {
			currentGeofences = append(currentGeofences, domain.VehicleGeofenceDetail{
				GeofenceID:   g.ID,
				GeofenceName: g.Name,
				Category:     g.Category,
			})
		}
	}

	return &VehicleLocationDetailResponse{
		VehicleID:     vehicle.ID,
		VehicleNumber: vehicle.VehicleNumber,
		CurrentLocation: CurrentLocationDetail{
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Timestamp: loc.Timestamp.Format(time.RFC3339),
		},
		CurrentGeofences: currentGeofences,
	}, nil
}

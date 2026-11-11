package service

import (
	"context"
	"fmt"
	"mapup-geofence/internal/domain"
	"mapup-geofence/internal/repository"
	"time"

	"github.com/google/uuid"
)

type AlertService struct {
	repo repository.Repository
}

func NewAlertService(repo repository.Repository) *AlertService {
	return &AlertService{repo: repo}
}

type ConfigureAlertRequest struct {
	GeofenceID string  `json:"geofence_id"`
	VehicleID   *string `json:"vehicle_id,omitempty"`
	EventType   string  `json:"event_type"` // entry, exit, both
}

type ConfigureAlertResponse struct {
	AlertID    string  `json:"alert_id"`
	GeofenceID string  `json:"geofence_id"`
	VehicleID  *string `json:"vehicle_id,omitempty"`
	EventType  string  `json:"event_type"`
	Status     string  `json:"status"`
}

func (s *AlertService) ConfigureAlert(ctx context.Context, req ConfigureAlertRequest) (*ConfigureAlertResponse, error) {
	if req.GeofenceID == "" {
		return nil, fmt.Errorf("geofence_id is required")
	}
	switch req.EventType {
	case domain.EventTypeEntry, domain.EventTypeExit, domain.EventTypeBoth:
		// valid
	default:
		return nil, fmt.Errorf("event_type must be one of: entry, exit, both")
	}

	geofence, err := s.repo.GetGeofenceByID(ctx, req.GeofenceID)
	if err != nil {
		return nil, fmt.Errorf("geofence not found: %w", err)
	}

	var vehicleNumber *string
	if req.VehicleID != nil && *req.VehicleID != "" {
		v, err := s.repo.GetVehicleByID(ctx, *req.VehicleID)
		if err != nil {
			return nil, fmt.Errorf("vehicle not found: %w", err)
		}
		vehicleNumber = &v.VehicleNumber
	}

	alertID := fmt.Sprintf("alert_%s", uuid.New().String()[:8])
	rule := &domain.AlertRule{
		AlertID:       alertID,
		GeofenceID:    geofence.ID,
		GeofenceName:  geofence.Name,
		VehicleID:     req.VehicleID,
		VehicleNumber: vehicleNumber,
		EventType:     req.EventType,
		Status:        "active",
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.repo.CreateAlertRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to configure alert: %w", err)
	}

	return &ConfigureAlertResponse{
		AlertID:    rule.AlertID,
		GeofenceID: rule.GeofenceID,
		VehicleID:  rule.VehicleID,
		EventType:  rule.EventType,
		Status:     rule.Status,
	}, nil
}

func (s *AlertService) GetAlerts(ctx context.Context, geofenceID, vehicleID string) ([]domain.AlertRule, error) {
	return s.repo.GetAlertRules(ctx, geofenceID, vehicleID)
}

func (s *AlertService) GetViolations(ctx context.Context, vehicleID, geofenceID string, startDate, endDate *time.Time, limit int) ([]domain.Violation, int, error) {
	return s.repo.GetViolations(ctx, vehicleID, geofenceID, startDate, endDate, limit)
}

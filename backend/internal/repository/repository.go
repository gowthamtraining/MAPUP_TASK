package repository

import (
	"context"
	"mapup-geofence/internal/domain"
	"time"
)

// Repository specifies all persistence operations for the geofencing system
type Repository interface {
	// Geofence operations
	CreateGeofence(ctx context.Context, g *domain.Geofence) error
	GetGeofences(ctx context.Context, category string) ([]domain.Geofence, error)
	GetGeofenceByID(ctx context.Context, id string) (*domain.Geofence, error)
	GetAllActiveGeofences(ctx context.Context) ([]domain.Geofence, error)

	// Vehicle operations
	CreateVehicle(ctx context.Context, v *domain.Vehicle) error
	GetVehicles(ctx context.Context) ([]domain.Vehicle, error)
	GetVehicleByID(ctx context.Context, id string) (*domain.Vehicle, error)

	// Location tracking
	RecordLocation(ctx context.Context, loc *domain.LocationRecord) error
	GetLatestLocation(ctx context.Context, vehicleID string) (*domain.LocationRecord, error)
	GetVehicleCurrentGeofenceIDs(ctx context.Context, vehicleID string) ([]string, error)
	UpdateVehicleCurrentGeofenceIDs(ctx context.Context, vehicleID string, geofenceIDs []string) error

	// Alert rules
	CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error
	GetAlertRules(ctx context.Context, geofenceID, vehicleID string) ([]domain.AlertRule, error)
	GetAllAlertRules(ctx context.Context) ([]domain.AlertRule, error)

	// Violations / History
	RecordViolation(ctx context.Context, v *domain.Violation) error
	GetViolations(ctx context.Context, vehicleID, geofenceID string, startDate, endDate *time.Time, limit int) ([]domain.Violation, int, error)

	// Close connections if any
	Close() error
}

package repository

import (
	"context"
	"fmt"
	"mapup-geofence/internal/domain"
	"sort"
	"strings"
	"sync"
	"time"
)

type MemoryRepository struct {
	mu sync.RWMutex

	geofences       map[string]domain.Geofence
	vehicles        map[string]domain.Vehicle
	locations       map[string][]domain.LocationRecord // vehicleID -> list
	currentGeofences map[string][]string               // vehicleID -> list of geofence IDs
	alertRules      map[string]domain.AlertRule
	violations      []domain.Violation
}

// NewMemoryRepository initializes an in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		geofences:        make(map[string]domain.Geofence),
		vehicles:         make(map[string]domain.Vehicle),
		locations:        make(map[string][]domain.LocationRecord),
		currentGeofences: make(map[string][]string),
		alertRules:       make(map[string]domain.AlertRule),
		violations:       make([]domain.Violation, 0),
	}
}

func (m *MemoryRepository) Close() error {
	return nil
}

// CreateGeofence stores a geofence
func (m *MemoryRepository) CreateGeofence(ctx context.Context, g *domain.Geofence) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.geofences[g.ID] = *g
	return nil
}

// GetGeofences retrieves geofences, filtered by category if provided
func (m *MemoryRepository) GetGeofences(ctx context.Context, category string) ([]domain.Geofence, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]domain.Geofence, 0)
	for _, g := range m.geofences {
		if category == "" || strings.EqualFold(g.Category, category) {
			result = append(result, g)
		}
	}
	// Sort by created_at descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result, nil
}

// GetGeofenceByID returns a single geofence
func (m *MemoryRepository) GetGeofenceByID(ctx context.Context, id string) (*domain.Geofence, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	g, exists := m.geofences[id]
	if !exists {
		return nil, fmt.Errorf("geofence with ID %s not found", id)
	}
	return &g, nil
}

// GetAllActiveGeofences returns all active geofences
func (m *MemoryRepository) GetAllActiveGeofences(ctx context.Context) ([]domain.Geofence, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]domain.Geofence, 0, len(m.geofences))
	for _, g := range m.geofences {
		if g.Status == "active" {
			result = append(result, g)
		}
	}
	return result, nil
}

// CreateVehicle registers a new vehicle
func (m *MemoryRepository) CreateVehicle(ctx context.Context, v *domain.Vehicle) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.vehicles[v.ID] = *v
	return nil
}

// GetVehicles retrieves all vehicles
func (m *MemoryRepository) GetVehicles(ctx context.Context) ([]domain.Vehicle, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]domain.Vehicle, 0, len(m.vehicles))
	for _, v := range m.vehicles {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result, nil
}

// GetVehicleByID returns vehicle by id
func (m *MemoryRepository) GetVehicleByID(ctx context.Context, id string) (*domain.Vehicle, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, exists := m.vehicles[id]
	if !exists {
		return nil, fmt.Errorf("vehicle with ID %s not found", id)
	}
	return &v, nil
}

// RecordLocation saves a vehicle location ping
func (m *MemoryRepository) RecordLocation(ctx context.Context, loc *domain.LocationRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.locations[loc.VehicleID] = append(m.locations[loc.VehicleID], *loc)
	return nil
}

// GetLatestLocation gets latest location for vehicle
func (m *MemoryRepository) GetLatestLocation(ctx context.Context, vehicleID string) (*domain.LocationRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list, exists := m.locations[vehicleID]
	if !exists || len(list) == 0 {
		return nil, fmt.Errorf("no location recorded for vehicle %s", vehicleID)
	}
	return &list[len(list)-1], nil
}

// GetVehicleCurrentGeofenceIDs retrieves the list of geofences the vehicle was last known to be inside
func (m *MemoryRepository) GetVehicleCurrentGeofenceIDs(ctx context.Context, vehicleID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids, exists := m.currentGeofences[vehicleID]
	if !exists {
		return []string{}, nil
	}
	// Return a copy
	res := make([]string, len(ids))
	copy(res, ids)
	return res, nil
}

// UpdateVehicleCurrentGeofenceIDs updates the current geofence IDs for the vehicle
func (m *MemoryRepository) UpdateVehicleCurrentGeofenceIDs(ctx context.Context, vehicleID string, geofenceIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	res := make([]string, len(geofenceIDs))
	copy(res, geofenceIDs)
	m.currentGeofences[vehicleID] = res
	return nil
}

// CreateAlertRule saves a rule
func (m *MemoryRepository) CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertRules[rule.AlertID] = *rule
	return nil
}

// GetAlertRules retrieves alert rules, optionally filtered
func (m *MemoryRepository) GetAlertRules(ctx context.Context, geofenceID, vehicleID string) ([]domain.AlertRule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]domain.AlertRule, 0)
	for _, rule := range m.alertRules {
		if geofenceID != "" && rule.GeofenceID != geofenceID {
			continue
		}
		if vehicleID != "" {
			// If rule is vehicle specific, must match; if rule is for all vehicles (nil/empty), it matches any query or specific filter
			if rule.VehicleID != nil && *rule.VehicleID != "" && *rule.VehicleID != vehicleID {
				continue
			}
		}
		result = append(result, rule)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result, nil
}

// GetAllAlertRules returns all active alert rules
func (m *MemoryRepository) GetAllAlertRules(ctx context.Context) ([]domain.AlertRule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]domain.AlertRule, 0, len(m.alertRules))
	for _, rule := range m.alertRules {
		result = append(result, rule)
	}
	return result, nil
}

// RecordViolation saves an entry/exit event in history
func (m *MemoryRepository) RecordViolation(ctx context.Context, v *domain.Violation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.violations = append(m.violations, *v)
	return nil
}

// GetViolations returns violations matching filters and limit
func (m *MemoryRepository) GetViolations(ctx context.Context, vehicleID, geofenceID string, startDate, endDate *time.Time, limit int) ([]domain.Violation, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	matched := make([]domain.Violation, 0)
	for _, v := range m.violations {
		if vehicleID != "" && v.VehicleID != vehicleID {
			continue
		}
		if geofenceID != "" && v.GeofenceID != geofenceID {
			continue
		}
		if startDate != nil && v.Timestamp.Before(*startDate) {
			continue
		}
		if endDate != nil && v.Timestamp.After(*endDate) {
			continue
		}
		matched = append(matched, v)
	}

	totalCount := len(matched)

	// Sort latest first
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Timestamp.After(matched[j].Timestamp)
	})

	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	if limit < len(matched) {
		matched = matched[:limit]
	}

	return matched, totalCount, nil
}

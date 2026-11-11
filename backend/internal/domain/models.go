package domain

import (
	"time"
)

// Allowed geofence categories
const (
	CategoryDeliveryZone   = "delivery_zone"
	CategoryRestrictedZone = "restricted_zone"
	CategoryTollZone       = "toll_zone"
	CategoryCustomerArea   = "customer_area"
)

// Allowed alert event types
const (
	EventTypeEntry = "entry"
	EventTypeExit  = "exit"
	EventTypeBoth  = "both"
)

// Coordinate represents [latitude, longitude]
type Coordinate [2]float64

func (c Coordinate) Lat() float64 { return c[0] }
func (c Coordinate) Lon() float64 { return c[1] }

// Geofence entity
type Geofence struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Coordinates []Coordinate `json:"coordinates"`
	Category    string       `json:"category"`
	Status      string       `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
}

// Vehicle entity
type Vehicle struct {
	ID            string    `json:"id"`
	VehicleNumber string    `json:"vehicle_number"`
	DriverName    string    `json:"driver_name"`
	VehicleType   string    `json:"vehicle_type"`
	Phone         string    `json:"phone"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// LocationRecord tracks a vehicle's historical position
type LocationRecord struct {
	VehicleID string    `json:"vehicle_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

// CurrentGeofenceStatus is used in POST /vehicles/location response
type CurrentGeofenceStatus struct {
	GeofenceID   string `json:"geofence_id"`
	GeofenceName string `json:"geofence_name"`
	Status       string `json:"status"` // e.g. "inside"
}

// VehicleGeofenceDetail is used in GET /vehicles/location/{vehicle_id} response
type VehicleGeofenceDetail struct {
	GeofenceID   string `json:"geofence_id"`
	GeofenceName string `json:"geofence_name"`
	Category     string `json:"category"`
}

// AlertRule represents a configured monitoring rule
type AlertRule struct {
	AlertID       string    `json:"alert_id"`
	GeofenceID    string    `json:"geofence_id"`
	GeofenceName  string    `json:"geofence_name,omitempty"`
	VehicleID     *string   `json:"vehicle_id,omitempty"` // null if applies to all vehicles
	VehicleNumber *string   `json:"vehicle_number,omitempty"`
	EventType     string    `json:"event_type"` // entry, exit, both
	Status        string    `json:"status"`     // active
	CreatedAt     time.Time `json:"created_at"`
}

// Violation represents historical entry/exit event records
type Violation struct {
	ID            string    `json:"id"`
	VehicleID     string    `json:"vehicle_id"`
	VehicleNumber string    `json:"vehicle_number"`
	GeofenceID    string    `json:"geofence_id"`
	GeofenceName  string    `json:"geofence_name"`
	EventType     string    `json:"event_type"` // entry or exit
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Timestamp     time.Time `json:"timestamp"`
}

// WebSocketAlertMessage is the exact payload sent over /ws/alerts
type WebSocketAlertMessage struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	Timestamp string                 `json:"timestamp"`
	Vehicle   VehicleAlertSummary    `json:"vehicle"`
	Geofence  GeofenceAlertSummary   `json:"geofence"`
	Location  LocationAlertSummary   `json:"location"`
}

type VehicleAlertSummary struct {
	VehicleID     string `json:"vehicle_id"`
	VehicleNumber string `json:"vehicle_number"`
	DriverName    string `json:"driver_name"`
}

type GeofenceAlertSummary struct {
	GeofenceID   string `json:"geofence_id"`
	GeofenceName string `json:"geofence_name"`
	Category     string `json:"category"`
}

type LocationAlertSummary struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

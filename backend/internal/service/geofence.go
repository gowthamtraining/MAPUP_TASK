package service

import (
	"context"
	"fmt"
	"mapup-geofence/internal/domain"
	"mapup-geofence/internal/geospatial"
	"mapup-geofence/internal/repository"
	"time"

	"github.com/google/uuid"
)

type GeofenceService struct {
	repo repository.Repository
}

func NewGeofenceService(repo repository.Repository) *GeofenceService {
	return &GeofenceService{repo: repo}
}

type CreateGeofenceRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Coordinates []domain.Coordinate `json:"coordinates"`
	Category    string              `json:"category"`
}

type CreateGeofenceResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (s *GeofenceService) CreateGeofence(ctx context.Context, req CreateGeofenceRequest) (*CreateGeofenceResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if err := geospatial.ValidateCategory(req.Category); err != nil {
		return nil, err
	}
	if err := geospatial.ValidatePolygonCoordinates(req.Coordinates); err != nil {
		return nil, err
	}

	geoID := fmt.Sprintf("geo_%s", uuid.New().String()[:8])
	geofence := &domain.Geofence{
		ID:          geoID,
		Name:        req.Name,
		Description: req.Description,
		Coordinates: req.Coordinates,
		Category:    req.Category,
		Status:      "active",
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.CreateGeofence(ctx, geofence); err != nil {
		return nil, fmt.Errorf("failed to save geofence: %w", err)
	}

	return &CreateGeofenceResponse{
		ID:     geofence.ID,
		Name:   geofence.Name,
		Status: geofence.Status,
	}, nil
}

func (s *GeofenceService) GetGeofences(ctx context.Context, category string) ([]domain.Geofence, error) {
	return s.repo.GetGeofences(ctx, category)
}

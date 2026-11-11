package service

import (
	"context"
	"fmt"
	"mapup-geofence/internal/domain"
	"mapup-geofence/internal/repository"
	"time"

	"github.com/google/uuid"
)

type VehicleService struct {
	repo repository.Repository
}

func NewVehicleService(repo repository.Repository) *VehicleService {
	return &VehicleService{repo: repo}
}

type CreateVehicleRequest struct {
	VehicleNumber string `json:"vehicle_number"`
	DriverName    string `json:"driver_name"`
	VehicleType   string `json:"vehicle_type"`
	Phone         string `json:"phone"`
}

type CreateVehicleResponse struct {
	ID            string `json:"id"`
	VehicleNumber string `json:"vehicle_number"`
	Status        string `json:"status"`
}

func (s *VehicleService) CreateVehicle(ctx context.Context, req CreateVehicleRequest) (*CreateVehicleResponse, error) {
	if req.VehicleNumber == "" {
		return nil, fmt.Errorf("vehicle_number is required")
	}
	if req.DriverName == "" {
		return nil, fmt.Errorf("driver_name is required")
	}
	if req.VehicleType == "" {
		return nil, fmt.Errorf("vehicle_type is required")
	}
	if req.Phone == "" {
		return nil, fmt.Errorf("phone is required")
	}

	vehID := fmt.Sprintf("veh_%s", uuid.New().String()[:8])
	vehicle := &domain.Vehicle{
		ID:            vehID,
		VehicleNumber: req.VehicleNumber,
		DriverName:    req.DriverName,
		VehicleType:   req.VehicleType,
		Phone:         req.Phone,
		Status:        "active",
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.repo.CreateVehicle(ctx, vehicle); err != nil {
		return nil, fmt.Errorf("failed to register vehicle: %w", err)
	}

	return &CreateVehicleResponse{
		ID:            vehicle.ID,
		VehicleNumber: vehicle.VehicleNumber,
		Status:        vehicle.Status,
	}, nil
}

func (s *VehicleService) GetVehicles(ctx context.Context) ([]domain.Vehicle, error) {
	return s.repo.GetVehicles(ctx)
}

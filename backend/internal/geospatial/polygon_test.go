package geospatial

import (
	"mapup-geofence/internal/domain"
	"testing"
)

func TestPolygonValidation(t *testing.T) {
	validSquare := []domain.Coordinate{
		{37.7749, -122.4194},
		{37.7849, -122.4194},
		{37.7849, -122.4094},
		{37.7749, -122.4094},
		{37.7749, -122.4194},
	}

	if err := ValidatePolygonCoordinates(validSquare); err != nil {
		t.Fatalf("expected valid polygon, got error: %v", err)
	}

	// Not closed
	unclosed := []domain.Coordinate{
		{37.7749, -122.4194},
		{37.7849, -122.4194},
		{37.7849, -122.4094},
		{37.7749, -122.4094},
	}
	if err := ValidatePolygonCoordinates(unclosed); err != ErrPolygonNotClosed {
		t.Fatalf("expected ErrPolygonNotClosed, got: %v", err)
	}

	// Insufficient points
	short := []domain.Coordinate{
		{37.7749, -122.4194},
		{37.7849, -122.4194},
		{37.7749, -122.4194},
	}
	if err := ValidatePolygonCoordinates(short); err != ErrInsufficientPoints {
		t.Fatalf("expected ErrInsufficientPoints, got: %v", err)
	}

	// Out of range coordinates
	outOfRange := []domain.Coordinate{
		{95.0, -122.4194},
		{37.7849, -122.4194},
		{37.7849, -122.4094},
		{95.0, -122.4194},
	}
	if err := ValidatePolygonCoordinates(outOfRange); err == nil {
		t.Fatalf("expected out of range latitude error, got nil")
	}
}

func TestContainsPoint(t *testing.T) {
	// A square polygon around Downtown SF
	square := []domain.Coordinate{
		{37.7700, -122.4200},
		{37.7800, -122.4200},
		{37.7800, -122.4100},
		{37.7700, -122.4100},
		{37.7700, -122.4200},
	}

	// Inside point
	if !ContainsPoint(square, 37.7750, -122.4150) {
		t.Errorf("expected point (37.7750, -122.4150) to be inside")
	}

	// Outside points
	if ContainsPoint(square, 37.7650, -122.4150) {
		t.Errorf("expected point (37.7650, -122.4150) to be outside")
	}
	if ContainsPoint(square, 37.7750, -122.4050) {
		t.Errorf("expected point (37.7750, -122.4050) to be outside")
	}

	// Boundary point (on the east edge: lon = -122.4100, lat = 37.7750)
	if !ContainsPoint(square, 37.7750, -122.4100) {
		t.Errorf("expected boundary point to be inside (ST_Covers semantics)")
	}
}

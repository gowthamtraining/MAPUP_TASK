package geospatial

import (
	"errors"
	"fmt"
	"mapup-geofence/internal/domain"
	"math"
)

var (
	ErrInvalidCategory      = errors.New("invalid category: must be delivery_zone, restricted_zone, toll_zone, or customer_area")
	ErrInsufficientPoints    = errors.New("polygon must have at least 4 coordinates (3 unique vertices + 1 closing point)")
	ErrPolygonNotClosed      = errors.New("first and last coordinates must be identical to form a closed polygon")
	ErrLatitudeOutOfRange    = errors.New("latitude must be between -90 and 90 degrees")
	ErrLongitudeOutOfRange   = errors.New("longitude must be between -180 and 180 degrees")
	ErrDuplicateUniquePoints = errors.New("polygon must have at least 3 distinct vertices")
)

// ValidateCategory checks if the category is one of the four allowed types
func ValidateCategory(cat string) error {
	switch cat {
	case domain.CategoryDeliveryZone,
		domain.CategoryRestrictedZone,
		domain.CategoryTollZone,
		domain.CategoryCustomerArea:
		return nil
	default:
		return ErrInvalidCategory
	}
}

// ValidatePolygonCoordinates validates coordinates according to requirements
func ValidatePolygonCoordinates(coords []domain.Coordinate) error {
	if len(coords) < 4 {
		return ErrInsufficientPoints
	}

	first := coords[0]
	last := coords[len(coords)-1]

	const epsilon = 1e-7
	if math.Abs(first.Lat()-last.Lat()) > epsilon || math.Abs(first.Lon()-last.Lon()) > epsilon {
		return ErrPolygonNotClosed
	}

	uniquePoints := make(map[string]bool)
	for i, pt := range coords {
		lat := pt.Lat()
		lon := pt.Lon()

		if lat < -90.0 || lat > 90.0 {
			return fmt.Errorf("%w at index %d: %f", ErrLatitudeOutOfRange, i, lat)
		}
		if lon < -180.0 || lon > 180.0 {
			return fmt.Errorf("%w at index %d: %f", ErrLongitudeOutOfRange, i, lon)
		}

		if i < len(coords)-1 {
			key := fmt.Sprintf("%.7f,%.7f", lat, lon)
			uniquePoints[key] = true
		}
	}

	if len(uniquePoints) < 3 {
		return ErrDuplicateUniquePoints
	}

	return nil
}

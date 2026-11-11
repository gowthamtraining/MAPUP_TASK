package geospatial

import (
	"mapup-geofence/internal/domain"
	"math"
)

// BoundingBox defines min and max latitude and longitude for fast pre-filtering
type BoundingBox struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

// ComputeBoundingBox calculates the bounding box of a polygon
func ComputeBoundingBox(coords []domain.Coordinate) BoundingBox {
	if len(coords) == 0 {
		return BoundingBox{}
	}
	minLat, maxLat := coords[0].Lat(), coords[0].Lat()
	minLon, maxLon := coords[0].Lon(), coords[0].Lon()

	for _, pt := range coords {
		lat := pt.Lat()
		lon := pt.Lon()
		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
		if lon < minLon {
			minLon = lon
		}
		if lon > maxLon {
			maxLon = lon
		}
	}
	return BoundingBox{
		MinLat: minLat,
		MaxLat: maxLat,
		MinLon: minLon,
		MaxLon: maxLon,
	}
}

// ContainsPoint checks if point (lat, lon) is inside polygon using Ray-Casting algorithm.
// Boundary Decision: Points lying directly on the perimeter boundary are considered INSIDE
// adhering to PostGIS ST_Covers semantics (inclusive boundary).
func ContainsPoint(polygon []domain.Coordinate, lat, lon float64) bool {
	if len(polygon) < 4 {
		return false
	}

	// Step 1: O(1) Bounding box pre-check
	bbox := ComputeBoundingBox(polygon)
	const tolerance = 1e-9
	if lat < bbox.MinLat-tolerance || lat > bbox.MaxLat+tolerance ||
		lon < bbox.MinLon-tolerance || lon > bbox.MaxLon+tolerance {
		return false
	}

	// Step 2: Ray-Casting Algorithm with boundary check
	inside := false
	n := len(polygon)

	for i := 0; i < n-1; i++ {
		p1 := polygon[i]
		p2 := polygon[i+1]

		lat1, lon1 := p1.Lat(), p1.Lon()
		lat2, lon2 := p2.Lat(), p2.Lon()

		// Check if point lies exactly on this segment
		if isPointOnSegment(lat, lon, lat1, lon1, lat2, lon2) {
			return true // Boundary is treated as inside (ST_Covers semantics)
		}

		// Ray casting test: ray extends from (lat, lon) along positive longitude axis
		// Check if the edge crosses the horizontal ray
		if (lat1 > lat) != (lat2 > lat) {
			// Calculate the longitude of the intersection of the edge with the horizontal line at lat
			lonIntersect := lon1 + (lat-lat1)*(lon2-lon1)/(lat2-lat1)
			if lon < lonIntersect {
				inside = !inside
			}
		}
	}

	return inside
}

// isPointOnSegment checks if (lat, lon) lies on segment between (lat1, lon1) and (lat2, lon2)
func isPointOnSegment(lat, lon, lat1, lon1, lat2, lon2 float64) bool {
	const epsilon = 1e-7

	// Check bounding box of segment
	minLat := math.Min(lat1, lat2) - epsilon
	maxLat := math.Max(lat1, lat2) + epsilon
	minLon := math.Min(lon1, lon2) - epsilon
	maxLon := math.Max(lon1, lon2) + epsilon

	if lat < minLat || lat > maxLat || lon < minLon || lon > maxLon {
		return false
	}

	// Cross product to check collinearity: (lat - lat1)*(lon2 - lon1) - (lon - lon1)*(lat2 - lat1)
	crossProduct := (lat-lat1)*(lon2-lon1) - (lon-lon1)*(lat2-lat1)
	if math.Abs(crossProduct) > epsilon {
		return false
	}

	return true
}

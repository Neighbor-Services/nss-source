package geo

import (
	"math"
)

// EarthRadiusKm represents the mean radius of the Earth in kilometers.
const EarthRadiusKm = 6371.0

// HaversineDistance computes the great-circle distance between two geographic coordinates in kilometers.
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	rLat1 := lat1 * (math.Pi / 180.0)
	rLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2.0)*math.Sin(dLat/2.0) +
		math.Cos(rLat1)*math.Cos(rLat2)*
			math.Sin(dLon/2.0)*math.Sin(dLon/2.0)

	c := 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1.0-a))
	return EarthRadiusKm * c
}

// IsWithinRadius checks if the target point (lat2, lon2) is within maxRadiusKm from the source point (lat1, lon1).
func IsWithinRadius(lat1, lon1, lat2, lon2, maxRadiusKm float64) (bool, float64) {
	dist := HaversineDistance(lat1, lon1, lat2, lon2)
	return dist <= maxRadiusKm, dist
}

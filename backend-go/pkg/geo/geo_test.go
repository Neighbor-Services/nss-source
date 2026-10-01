package geo_test

import (
	"math"
	"testing"

	"backend-go/pkg/geo"
)

func TestHaversineDistance(t *testing.T) {
	// London (51.5074, -0.1278) to Paris (48.8566, 2.3522) is approx 213.4 miles (~343 km)
	dist := geo.HaversineDistance(51.5074, -0.1278, 48.8566, 2.3522)
	if math.Abs(dist-213.4) > 5.0 {
		t.Errorf("Expected London-Paris distance ~213.4 miles, got %.2f miles", dist)
	}

	// Same point distance must be 0
	dist0 := geo.HaversineDistance(40.7128, -74.0060, 40.7128, -74.0060)
	if dist0 != 0.0 {
		t.Errorf("Expected distance 0, got %f", dist0)
	}

	// Within 25 miles test
	within, d := geo.IsWithinRadius(51.5074, -0.1278, 51.5200, -0.1200, 25.0)
	if !within {
		t.Errorf("Expected points to be within 25 miles, calculated %.2f miles", d)
	}

	// Outside 25 miles test
	outside, d2 := geo.IsWithinRadius(51.5074, -0.1278, 52.0000, 0.5000, 25.0)
	if outside {
		t.Errorf("Expected points to be outside 25 miles, calculated %.2f miles", d2)
	}
}

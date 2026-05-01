package basic

import (
	"math"
	"testing"
)

func TestHeightDegreeMatchesSphericalArc(t *testing.T) {
	height := 10000.0
	got := HeightDegree(height)
	want := HeightDistance(height) / EARTH_AVERAGE_RADIUS * 180 / math.Pi
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("HeightDegree mismatch: got %.15f want %.15f", got, want)
	}
}

func TestHeightDegreeByLatMatchesSphericalArc(t *testing.T) {
	height := 10000.0
	lat := 45.0
	radius := GeocentricRadius(lat)
	got := HeightDegreeByLat(height, lat)
	want := HeightDistanceByLat(height, lat) / radius * 180 / math.Pi
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("HeightDegreeByLat mismatch: got %.15f want %.15f", got, want)
	}
}

func TestHeightDegreeReferenceValue(t *testing.T) {
	got := HeightDegree(10000)
	want := 3.20801665537668
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("HeightDegree(10000) = %.15f want %.15f", got, want)
	}
}

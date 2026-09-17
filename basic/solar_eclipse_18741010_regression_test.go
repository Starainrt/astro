package basic

import (
	"testing"

	"github.com/starainrt/astro/internal/geodata"
)

// TestSolarEclipse18741010GrazingAnnularBandContainsSweep pins the other end of
// the grazing annular family: a one-limit path whose analytic sweep returns a
// band that still terminates 48 km inside its own sampled umbra. The band must
// be rebuilt from those samples instead of being exported as the short ribbon.
func TestSolarEclipse18741010GrazingAnnularBandContainsSweep(t *testing.T) {
	result := SolarEclipsePartialFootprints(
		JDECalc(1874, 10, 10),
		SolarEclipsePartialFootprintOptions{
			StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
		},
	)
	if result.Eclipse.Type != SolarEclipseAnnular ||
		result.Eclipse.Centrality != SolarEclipseCentralOneLimit {
		t.Fatalf("type=%s centrality=%s, want a one-limit annular eclipse",
			result.Eclipse.Type, result.Eclipse.Centrality)
	}
	if len(result.CentralBandSegments) != 1 {
		t.Fatalf("central-band segments=%d, want one closed band", len(result.CentralBandSegments))
	}
	ring := result.CentralBandSegments[0]
	if len(ring) < 4 || solarEclipsePathDistanceKM(ring[0], ring[len(ring)-1]) > 0.01 {
		t.Fatal("central-band envelope is not closed")
	}
	if !solarEclipseBandContainsFootprintsWithinKM(
		result.CentralBandSegments, result.CentralBandFootprints,
		solarEclipseCentralBandUnionContainmentToleranceKM,
	) {
		t.Fatal("central band does not contain the sampled umbral footprints")
	}
	// A station with two minutes of annular phase, 14 degrees above the horizon,
	// used to fall outside the exported band by more than 400 km.
	matrix := [][]geodata.GeoPoint{geodataRingFromPath(ring)}
	point := []geodata.GeoPoint{{Longitude: 60.5, Latitude: 57.5}}
	if !geodata.SphericalPolygonsContainPoints(matrix, point)[0] {
		t.Fatal("central band misses the visible annular region at (60.5,57.5)")
	}
}

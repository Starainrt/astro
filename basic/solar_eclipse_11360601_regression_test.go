package basic

import (
	"testing"

	"github.com/starainrt/astro/internal/geodata"
)

// TestSolarEclipseGrazingClosureRootsAreRecovered pins the events whose
// greatest-at-horizon closure arcs the analytic seedings miss: a grazing
// closure root can sit outside the sampled horizon branches, and the local
// classification then refuses the solved root. Recovering both from the sampled
// sweep restores the analytic critical envelope, which covers the visible
// annulus better than the sampled union that used to replace it.
func TestSolarEclipseGrazingClosureRootsAreRecovered(t *testing.T) {
	for _, date := range [][3]int{{1136, 6, 1}, {-1480, 12, 27}, {5705, 6, 17}} {
		seed := JDECalc(date[0], date[1], float64(date[2]))
		result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
			StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
		})
		if result.Eclipse.Centrality != SolarEclipseCentralTwoLimits {
			t.Fatalf("%04d-%02d-%02d centrality=%s, want two limits",
				date[0], date[1], date[2], result.Eclipse.Centrality)
		}
		if len(result.CentralBandHorizonClosures) != 2 {
			t.Fatalf("%04d-%02d-%02d horizon closures=%d, want two",
				date[0], date[1], date[2], len(result.CentralBandHorizonClosures))
		}
		if result.CentralBandSampled {
			t.Fatalf("%04d-%02d-%02d fell back to the sampled footprint union",
				date[0], date[1], date[2])
		}
	}
}

// TestSolarEclipseGrazingEventsWithoutClosuresStaySampled pins the other half
// of the criterion: a two-limit grazing event whose caps are not bounded by the
// greatest-at-horizon condition has no closure arc at all, and its band must
// stay a valid closed reconstruction instead of silently disappearing. The
// reasons are measured, not assumed: 4862-09-28 ends 40 km inside the horizon
// (+0.36 degrees at the cap, so the umbral rim bounds it) and 1552-07-21 has its
// boundary running along the horizon (+0.004 then -0.000 degrees), which makes
// the arc degenerate.
func TestSolarEclipseGrazingEventsWithoutClosuresStaySampled(t *testing.T) {
	for _, date := range [][3]int{{4862, 9, 28}, {1552, 7, 21}} {
		seed := JDECalc(date[0], date[1], float64(date[2]))
		result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
			StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
		})
		if result.Eclipse.Centrality != SolarEclipseCentralTwoLimits {
			t.Fatalf("%04d-%02d-%02d centrality=%s, want two limits",
				date[0], date[1], date[2], result.Eclipse.Centrality)
		}
		if len(result.CentralBandHorizonClosures) != 0 {
			t.Fatalf("%04d-%02d-%02d closures=%d, expected none",
				date[0], date[1], date[2], len(result.CentralBandHorizonClosures))
		}
		if len(result.CentralBandSegments) == 0 {
			t.Fatalf("%04d-%02d-%02d exported no central band", date[0], date[1], date[2])
		}
		ring := result.CentralBandSegments[0]
		if len(ring) < 4 || solarEclipsePathDistanceKM(ring[0], ring[len(ring)-1]) > 0.01 {
			t.Fatalf("%04d-%02d-%02d central band is not closed", date[0], date[1], date[2])
		}
	}
}

// TestSolarEclipse11360601GrazingAnnularBandContainsSweep pins the grazing polar
// annular event whose shadow axis runs almost parallel to the surface. Its
// instantaneous antumbral footprint is a long spindle: the rim is cut by the
// horizon over the contact intervals and fully closed over the middle of the
// path. The analytic envelope has no horizon roots there, and the open-arc
// sweep drops the closed middle samples, so the band used to be exported as a
// chordal ribbon hundreds of kilometres smaller than the umbra it describes.
func TestSolarEclipse11360601GrazingAnnularBandContainsSweep(t *testing.T) {
	result := SolarEclipsePartialFootprints(
		JDECalc(1136, 6, 1),
		SolarEclipsePartialFootprintOptions{
			StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
		},
	)
	if result.Eclipse.Type != SolarEclipseAnnular ||
		result.Eclipse.Centrality != SolarEclipseCentralTwoLimits {
		t.Fatalf("type=%s centrality=%s, want a two-limit annular eclipse",
			result.Eclipse.Type, result.Eclipse.Centrality)
	}
	// The band must be the analytic critical envelope, which needs both
	// greatest-at-horizon closure arcs. Recovering the second root from the
	// sampled sweep is what restored them; without it the event silently falls
	// back to a sampled union that leaves 0.3% of the visible annulus uncovered.
	if len(result.CentralBandHorizonClosures) != 2 {
		t.Fatalf("horizon closures=%d, want two", len(result.CentralBandHorizonClosures))
	}
	if result.CentralBandSampled {
		t.Fatal("central band fell back to the sampled footprint union")
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
	// The spindle reaches the poleward tip far above the center-line interval
	// the old ribbon covered; a point on that tip must stay inside the band.
	for _, probe := range []SolarEclipsePathPoint{
		{Longitude: 120, Latitude: 73},
		{Longitude: 125, Latitude: 71},
	} {
		matrix := [][]geodata.GeoPoint{geodataRingFromPath(ring)}
		point := []geodata.GeoPoint{{Longitude: probe.Longitude, Latitude: probe.Latitude}}
		if !geodata.SphericalPolygonsContainPoints(matrix, point)[0] {
			t.Fatalf("central band misses (%g,%g)", probe.Longitude, probe.Latitude)
		}
	}
}

func geodataRingFromPath(points []SolarEclipsePathPoint) []geodata.GeoPoint {
	ring := make([]geodata.GeoPoint, len(points))
	for index, point := range points {
		ring[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	return ring
}

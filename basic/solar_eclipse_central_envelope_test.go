package basic

import (
	"math"
	"testing"
)

func TestSolarEclipse20120521CentralBandIsContinuousCriticalEnvelope(t *testing.T) {
	seed := JDECalc(2012, 5, 21)
	result := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96,
	})
	if result.Eclipse.Type != SolarEclipseAnnular ||
		result.Eclipse.Centrality != SolarEclipseCentralTwoLimits {
		t.Fatalf("eclipse=%s/%s, want two-limit annular", result.Eclipse.Type, result.Eclipse.Centrality)
	}
	if len(result.CentralBandSegments) != 1 {
		t.Fatalf("central-band segments=%d, want one continuous envelope", len(result.CentralBandSegments))
	}

	ring := result.CentralBandSegments[0]
	if len(ring) < 4 || solarEclipsePathDistanceKM(ring[0], ring[len(ring)-1]) > 0.01 {
		t.Fatal("central-band envelope is not closed")
	}
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	interiorPoints := 0
	for index, point := range ring {
		if index > 0 {
			if distance := solarEclipsePathDistanceKM(ring[index-1], point); distance >
				solarEclipseCentralEnvelopeMaxSpacingKM+1e-6 {
				t.Fatalf("central-band edge %d is %.3f km, want at most %.3f km",
					index-1, distance, solarEclipseCentralEnvelopeMaxSpacingKM)
			}
		}
		if solarEclipseCentralEnvelopePointOnHorizonClosure(
			point, result.CentralBandHorizonClosures,
		) {
			continue
		}
		residual, ok := solarEclipseNonCentralBandBoundaryResidualAt(
			solver.magnitudeEvaluationAt(point.JDE), point.Longitude, point.Latitude,
		)
		if !ok || math.Abs(residual[0]) > 1e-6 ||
			math.Abs(residual[1]) > solarEclipseNonCentralBandDerivativeTolerance {
			t.Fatalf("central-band point %d altitude=%g residual=(%g,%g), want critical-envelope solution",
				index, point.SunAltitude, residual[0], residual[1])
		}
		interiorPoints++
	}
	if interiorPoints < 100 {
		t.Fatalf("critical-envelope points=%d, want a densely sampled continuous boundary", interiorPoints)
	}

	// NADC 3D eclipse data publishes these samples on the two 2012 central
	// limits. The allowance covers ephemeris/delta-T model differences and the
	// finite spacing between our continuation samples.
	for _, anchor := range []struct {
		name  string
		point SolarEclipsePathPoint
	}{
		{name: "north limit", point: SolarEclipsePathPoint{Longitude: 120.9238, Latitude: 24.6531}},
		{name: "south limit", point: SolarEclipsePathPoint{Longitude: 120.3071, Latitude: 27.5328}},
	} {
		nearest := math.Inf(1)
		for _, point := range ring {
			nearest = math.Min(nearest, solarEclipsePathDistanceKM(point, anchor.point))
		}
		if nearest > 60 {
			t.Fatalf("central-band %s is %.3f km from the NADC reference sample", anchor.name, nearest)
		}
	}
}

func TestSolarEclipseCentralEnvelopeIndependentOfRiseSetOutput(t *testing.T) {
	seed := JDECalc(2012, 5, 21)
	withCurves := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{StepDays: 2.0 / 1440, BoundaryPoints: 96})
	withoutCurves := SolarEclipsePartialFootprints(seed, SolarEclipsePartialFootprintOptions{StepDays: 2.0 / 1440, BoundaryPoints: 96, DisableRiseSetCurves: true})
	if len(withoutCurves.CentralBandSegments) == 0 {
		t.Fatal("disabling output curves removed the authoritative central envelope")
	}
	if len(withoutCurves.RiseSetCurves) != 0 {
		t.Fatal("disabled rise/set curves were returned")
	}
	if len(withCurves.CentralBandSegments) != len(withoutCurves.CentralBandSegments) {
		t.Fatalf("central envelope segment count changed: %d vs %d", len(withCurves.CentralBandSegments), len(withoutCurves.CentralBandSegments))
	}
}

func solarEclipseCentralEnvelopePointOnHorizonClosure(
	point SolarEclipsePathPoint,
	closures [][]SolarEclipsePathPoint,
) bool {
	for _, closure := range closures {
		for _, horizonPoint := range closure {
			if solarEclipsePathDistanceKM(point, horizonPoint) <= 0.001 &&
				math.Abs(point.JDE-horizonPoint.JDE) <= solarEclipsePathDuplicateTimeDays {
				return true
			}
		}
	}
	return false
}

package basic

import (
	"math"
	"testing"
)

func TestSolarEclipseCentralPathMatchesGlobalGreatest(t *testing.T) {
	seedJDE := JDECalc(2024, 4, 8)
	global := SolarEclipse(seedJDE)
	path := SolarEclipseCentralPath(seedJDE, SolarEclipsePathOptions{StepDays: 5.0 / 1440.0})

	if path.Eclipse.Type != global.Type {
		t.Fatalf("type mismatch: got %s want %s", path.Eclipse.Type, global.Type)
	}
	if !path.Eclipse.HasCentral {
		t.Fatalf("expected central eclipse path")
	}
	if len(path.CenterLine) == 0 {
		t.Fatalf("expected center line points")
	}
	if len(path.NorthernLimit) == 0 || len(path.SouthernLimit) == 0 {
		t.Fatalf("expected central path limits: north=%d south=%d", len(path.NorthernLimit), len(path.SouthernLimit))
	}

	assertSolarEclipseJDEClose(t, "Greatest.JDE", path.Greatest.JDE, global.GreatestEclipse, 1e-8)
	assertSolarEclipseFloatClose(t, "Greatest.Longitude", path.Greatest.Longitude, global.GreatestLongitude, 1e-9)
	assertSolarEclipseFloatClose(t, "Greatest.Latitude", path.Greatest.Latitude, global.GreatestLatitude, 1e-9)
	assertSolarEclipseFloatClose(t, "Greatest.WidthKM", path.Greatest.WidthKM, global.PathWidthKM, 1e-9)

	foundGreatest := false
	for _, point := range path.CenterLine {
		if math.Abs(point.JDE-global.GreatestEclipse) <= solarEclipsePathDuplicateTimeDays {
			foundGreatest = true
			break
		}
	}
	if !foundGreatest {
		t.Fatalf("center line should include greatest eclipse JDE %.12f", global.GreatestEclipse)
	}
}

func TestSolarEclipseCentralPathTargetSpacingRefinesSamples(t *testing.T) {
	seedJDE := JDECalc(2024, 4, 8)
	coarse := SolarEclipseCentralPath(seedJDE, SolarEclipsePathOptions{StepDays: 20.0 / 1440.0})
	refined := SolarEclipseCentralPath(seedJDE, SolarEclipsePathOptions{
		StepDays:        20.0 / 1440.0,
		TargetSpacingKM: 120,
	})

	if len(coarse.CenterLine) == 0 || len(refined.CenterLine) == 0 {
		t.Fatalf("expected path points: coarse=%d refined=%d", len(coarse.CenterLine), len(refined.CenterLine))
	}
	if len(refined.CenterLine) <= len(coarse.CenterLine) {
		t.Fatalf("target spacing should refine samples: coarse=%d refined=%d", len(coarse.CenterLine), len(refined.CenterLine))
	}
	for i := 1; i < len(refined.CenterLine); i++ {
		distanceKM := solarEclipsePathDistanceKM(refined.CenterLine[i-1], refined.CenterLine[i])
		if distanceKM > 120.1 {
			t.Fatalf("segment %d too long: got %.6f km want <= 120.1 km", i, distanceKM)
		}
	}
}

func TestSolarEclipseCentralPathPartialHasNoCenterLine(t *testing.T) {
	path := SolarEclipseCentralPath(JDECalc(2025, 3, 29), SolarEclipsePathOptions{})

	if path.Eclipse.Type != SolarEclipsePartial {
		t.Fatalf("unexpected eclipse type: got %s want %s", path.Eclipse.Type, SolarEclipsePartial)
	}
	if path.Eclipse.HasCentral {
		t.Fatalf("partial eclipse should not have central path")
	}
	if len(path.CenterLine) != 0 || len(path.NorthernLimit) != 0 || len(path.SouthernLimit) != 0 {
		t.Fatalf(
			"partial eclipse should not return central path points: center=%d north=%d south=%d",
			len(path.CenterLine),
			len(path.NorthernLimit),
			len(path.SouthernLimit),
		)
	}
}

func TestSolarEclipsePartialFootprintsIncludeGreatest(t *testing.T) {
	seedJDE := JDECalc(2024, 4, 8)
	global := SolarEclipse(seedJDE)
	footprints := SolarEclipsePartialFootprints(seedJDE, SolarEclipsePartialFootprintOptions{
		StepDays:       30.0 / 1440.0,
		BoundaryPoints: 72,
	})

	if footprints.Eclipse.Type != SolarEclipseTotal {
		t.Fatalf("unexpected eclipse type: got %s want %s", footprints.Eclipse.Type, SolarEclipseTotal)
	}
	if footprints.BoundaryPoints != 72 {
		t.Fatalf("boundary points mismatch: got %d want 72", footprints.BoundaryPoints)
	}
	if len(footprints.Footprints) == 0 {
		t.Fatalf("expected partial footprints")
	}

	foundGreatest := false
	for _, footprint := range footprints.Footprints {
		if math.Abs(footprint.JDE-global.GreatestEclipse) <= solarEclipsePathDuplicateTimeDays {
			foundGreatest = true
		}
		if len(footprint.Boundaries) == 0 {
			t.Fatalf("footprint at %.12f has no boundaries", footprint.JDE)
		}
		for _, boundary := range footprint.Boundaries {
			if len(boundary) == 0 {
				t.Fatalf("footprint at %.12f has an empty boundary segment", footprint.JDE)
			}
			for _, point := range boundary {
				if math.Abs(point.JDE-footprint.JDE) > 1e-12 {
					t.Fatalf("point JDE mismatch: got %.12f want %.12f", point.JDE, footprint.JDE)
				}
				if point.Longitude < -180 || point.Longitude > 180 {
					t.Fatalf("longitude out of range: %.9f", point.Longitude)
				}
				if point.Latitude < -90 || point.Latitude > 90 {
					t.Fatalf("latitude out of range: %.9f", point.Latitude)
				}
			}
		}
		assertSolarEclipseFootprintClosedFlag(t, footprint)
	}
	if !foundGreatest {
		t.Fatalf("partial footprints should include greatest eclipse JDE %.12f", global.GreatestEclipse)
	}
}

func TestSolarEclipsePartialFootprintsWorkForPartialOnlyEclipse(t *testing.T) {
	footprints := SolarEclipsePartialFootprints(JDECalc(2025, 3, 29), SolarEclipsePartialFootprintOptions{
		StepDays:       30.0 / 1440.0,
		BoundaryPoints: 72,
	})

	if footprints.Eclipse.Type != SolarEclipsePartial {
		t.Fatalf("unexpected eclipse type: got %s want %s", footprints.Eclipse.Type, SolarEclipsePartial)
	}
	if footprints.Eclipse.HasCentral {
		t.Fatalf("partial-only eclipse should not have central path")
	}
	if len(footprints.Footprints) == 0 {
		t.Fatalf("expected partial footprints for partial-only eclipse")
	}
}

func TestSolarEclipsePartialFootprintsNoEvent(t *testing.T) {
	footprints := SolarEclipsePartialFootprints(JDECalc(2023, 5, 15), SolarEclipsePartialFootprintOptions{})

	if footprints.Eclipse.Type != SolarEclipseNone {
		t.Fatalf("unexpected eclipse type: got %s want %s", footprints.Eclipse.Type, SolarEclipseNone)
	}
	if len(footprints.Footprints) != 0 {
		t.Fatalf("no eclipse should not return footprints: got %d", len(footprints.Footprints))
	}
}

func TestSolarEclipsePartialAreaCompatibilityWrapper(t *testing.T) {
	seedJDE := JDECalc(2024, 4, 8)
	options := SolarEclipsePartialAreaOptions{
		StepDays:       30.0 / 1440.0,
		BoundaryPoints: 72,
	}
	compat := SolarEclipsePartialArea(seedJDE, options)
	primary := SolarEclipsePartialFootprints(seedJDE, options)

	if compat.Eclipse.Type != primary.Eclipse.Type {
		t.Fatalf("compat type mismatch: got %s want %s", compat.Eclipse.Type, primary.Eclipse.Type)
	}
	if len(compat.Footprints) != len(primary.Footprints) {
		t.Fatalf("compat footprint count mismatch: got %d want %d", len(compat.Footprints), len(primary.Footprints))
	}
}

func assertSolarEclipseFootprintClosedFlag(t *testing.T, footprint SolarEclipsePartialFootprint) {
	t.Helper()
	if !footprint.Closed {
		return
	}
	if len(footprint.Boundaries) != 1 {
		t.Fatalf("closed footprint should have one boundary: got %d", len(footprint.Boundaries))
	}
	boundary := footprint.Boundaries[0]
	if len(boundary) < 2 {
		t.Fatalf("closed footprint boundary too short: got %d", len(boundary))
	}
	first := boundary[0]
	last := boundary[len(boundary)-1]
	if first.Longitude != last.Longitude || first.Latitude != last.Latitude {
		t.Fatalf("closed footprint boundary should repeat first point at end")
	}
}

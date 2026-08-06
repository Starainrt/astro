package basic

import (
	"math"
	"testing"
	"time"
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
	if footprints.Footprints[0].Closed {
		t.Fatal("grazing first footprint must remain open for horizon closure")
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

func TestSolarEclipseShadowContactsAgainstNASA2012Baseline(t *testing.T) {
	result := SolarEclipsePartialFootprints(
		solarEclipseUTToTTJDE(time.Date(2012, 5, 20, 0, 0, 0, 0, time.UTC)),
		SolarEclipsePartialFootprintOptions{
			StepDays:              30.0 / 1440.0,
			BoundaryPoints:        72,
			CentralShadowStepDays: 10.0 / 1440.0,
		},
	)
	baseline := []struct {
		name  string
		point SolarEclipsePathPoint
		want  time.Time
	}{
		{"P1", result.P1, time.Date(2012, 5, 20, 20, 56, 7, 0, time.UTC)},
		{"P4", result.P4, time.Date(2012, 5, 21, 2, 49, 21, 500000000, time.UTC)},
		{"U1", result.U1, time.Date(2012, 5, 20, 22, 6, 16, 600000000, time.UTC)},
		{"U2", result.U2, time.Date(2012, 5, 20, 22, 11, 46, 400000000, time.UTC)},
		{"U3", result.U3, time.Date(2012, 5, 21, 1, 33, 42, 800000000, time.UTC)},
		{"U4", result.U4, time.Date(2012, 5, 21, 1, 39, 11, 200000000, time.UTC)},
	}
	for _, contact := range baseline {
		if contact.point.JDE == 0 {
			t.Fatalf("%s contact is absent", contact.name)
		}
		assertLocalSolarEclipseJDEClose(
			t,
			contact.name,
			contact.point.JDE,
			solarEclipseUTToTTJDE(contact.want),
			3*time.Second,
		)
		if math.Abs(contact.point.SunAltitude) > 0.01 {
			t.Fatalf("%s Sun altitude = %.9f degrees, want horizon contact", contact.name, contact.point.SunAltitude)
		}
	}
	if result.P2.JDE != 0 || result.P3.JDE != 0 {
		t.Fatalf("2012 eclipse unexpectedly has P2/P3 contacts: P2=%+v P3=%+v", result.P2, result.P3)
	}
	if len(result.CentralShadowFootprints) == 0 {
		t.Fatal("expected sampled central-shadow footprints")
	}
	if math.Abs(result.CentralShadowStepDays-10.0/1440.0) > 1e-12 {
		t.Fatalf("central shadow step = %.12f days, want ten minutes", result.CentralShadowStepDays)
	}
	for _, footprint := range result.CentralShadowFootprints {
		if len(footprint.Boundaries) == 0 {
			t.Fatalf("central-shadow footprint at %.12f has no boundary", footprint.JDE)
		}
	}
}

func TestSolarEclipseShadowContactsIncludeP2P3WhenPenumbraEntersEarthDisk(t *testing.T) {
	result := SolarEclipsePartialFootprints(
		solarEclipseUTToTTJDE(time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)),
		SolarEclipsePartialFootprintOptions{StepDays: 30.0 / 1440.0, BoundaryPoints: 36},
	)
	for name, contacts := range map[string][]SolarEclipsePathPoint{
		"penumbral": {result.P1, result.P2, result.P3, result.P4},
		"central":   {result.U1, result.U2, result.U3, result.U4},
	} {
		for index, contact := range contacts {
			if contact.JDE == 0 {
				t.Fatalf("%s contact %d is absent", name, index)
			}
			if index > 0 && !(contacts[index-1].JDE < contact.JDE) {
				t.Fatalf("%s contacts out of order at %d: %.12f >= %.12f", name, index, contacts[index-1].JDE, contact.JDE)
			}
		}
	}
	if !(result.P1.JDE < result.U1.JDE && result.U4.JDE < result.P4.JDE) {
		t.Fatalf("central shadow contacts must lie inside partial phase: P1=%v U1=%v U4=%v P4=%v",
			result.P1.JDE, result.U1.JDE, result.U4.JDE, result.P4.JDE)
	}
	if result.CentralShadowFootprints != nil || result.CentralShadowStepDays != 0 {
		t.Fatal("central-shadow footprints must remain disabled by default")
	}
}

func TestSolarEclipsePartialBoundarySegmentsRemainClosedAcrossAntimeridian(t *testing.T) {
	samples := []solarEclipsePartialBoundarySample{
		{point: SolarEclipsePathPoint{Longitude: 170, Latitude: 20}, ok: true},
		{point: SolarEclipsePathPoint{Longitude: -170, Latitude: 25}, ok: true},
		{point: SolarEclipsePathPoint{Longitude: -160, Latitude: 10}, ok: true},
		{point: SolarEclipsePathPoint{Longitude: 160, Latitude: 5}, ok: true},
	}
	boundaries, closed := solarEclipsePartialBoundarySegments(samples)
	if !closed {
		t.Fatal("complete spherical boundary must remain closed after antimeridian splitting")
	}
	if len(boundaries) < 2 {
		t.Fatalf("expected antimeridian split, got %d boundary segment(s)", len(boundaries))
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
	if len(footprint.Boundaries) == 0 {
		t.Fatal("closed footprint has no boundaries")
	}
	if len(footprint.Boundaries) > 1 {
		return
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

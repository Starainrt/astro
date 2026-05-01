package eclipse

import (
	"math"
	"testing"
	"time"
)

func TestSolarEclipseCentralPathKeepsLocation(t *testing.T) {
	loc := time.FixedZone("UTC+08", 8*3600)
	path, ok := SolarEclipseCentralPath(
		time.Date(2024, 4, 8, 12, 0, 0, 0, loc),
		SolarEclipsePathOptions{Step: 5 * time.Minute},
	)
	if !ok {
		t.Fatalf("expected central path")
	}
	if path.Eclipse.Type != SolarEclipseTotal {
		t.Fatalf("unexpected eclipse type: got %s want %s", path.Eclipse.Type, SolarEclipseTotal)
	}
	if math.Abs(float64(path.Step-5*time.Minute)) > float64(time.Millisecond) {
		t.Fatalf("step mismatch: got %s want %s", path.Step, 5*time.Minute)
	}

	for _, item := range []struct {
		name string
		tm   time.Time
	}{
		{name: "Eclipse.GreatestEclipse", tm: path.Eclipse.GreatestEclipse},
		{name: "Greatest.Time", tm: path.Greatest.Time},
		{name: "CenterLine[0].Time", tm: path.CenterLine[0].Time},
		{name: "CenterLine[last].Time", tm: path.CenterLine[len(path.CenterLine)-1].Time},
	} {
		if item.tm.Location() != loc {
			t.Fatalf("%s location mismatch: got %q want %q", item.name, item.tm.Location(), loc)
		}
	}
}

func TestSolarEclipseCentralPathStepControlsDensity(t *testing.T) {
	date := time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC)
	coarse, ok := SolarEclipseCentralPath(date, SolarEclipsePathOptions{Step: 10 * time.Minute})
	if !ok {
		t.Fatalf("expected coarse central path")
	}
	fine, ok := SolarEclipseCentralPath(date, SolarEclipsePathOptions{Step: 2 * time.Minute})
	if !ok {
		t.Fatalf("expected fine central path")
	}
	if len(fine.CenterLine) <= len(coarse.CenterLine) {
		t.Fatalf("finer step should produce more points: coarse=%d fine=%d", len(coarse.CenterLine), len(fine.CenterLine))
	}
}

func TestSolarEclipseCentralPathReturnsFalseForPartialOnly(t *testing.T) {
	_, ok := SolarEclipseCentralPath(time.Date(2025, 3, 29, 0, 0, 0, 0, time.UTC), SolarEclipsePathOptions{})
	if ok {
		t.Fatalf("partial-only eclipse should not return a central path")
	}
}

func TestSolarEclipsePartialFootprintsKeepLocation(t *testing.T) {
	loc := time.FixedZone("UTC+08", 8*3600)
	footprints, ok := SolarEclipsePartialFootprints(
		time.Date(2024, 4, 8, 12, 0, 0, 0, loc),
		SolarEclipsePartialFootprintOptions{
			Step:           30 * time.Minute,
			BoundaryPoints: 72,
		},
	)
	if !ok {
		t.Fatalf("expected partial footprints")
	}
	if footprints.Eclipse.Type != SolarEclipseTotal {
		t.Fatalf("unexpected eclipse type: got %s want %s", footprints.Eclipse.Type, SolarEclipseTotal)
	}
	if math.Abs(float64(footprints.Step-30*time.Minute)) > float64(time.Millisecond) {
		t.Fatalf("step mismatch: got %s want %s", footprints.Step, 30*time.Minute)
	}
	if footprints.BoundaryPoints != 72 {
		t.Fatalf("boundary points mismatch: got %d want 72", footprints.BoundaryPoints)
	}

	for _, item := range []struct {
		name string
		tm   time.Time
	}{
		{name: "Eclipse.GreatestEclipse", tm: footprints.Eclipse.GreatestEclipse},
		{name: "Footprints[0].Time", tm: footprints.Footprints[0].Time},
		{name: "Footprints[last].Time", tm: footprints.Footprints[len(footprints.Footprints)-1].Time},
		{name: "Boundary point", tm: footprints.Footprints[0].Boundaries[0][0].Time},
	} {
		if item.tm.Location() != loc {
			t.Fatalf("%s location mismatch: got %q want %q", item.name, item.tm.Location(), loc)
		}
	}
}

func TestSolarEclipsePartialFootprintsWorkForPartialOnly(t *testing.T) {
	footprints, ok := SolarEclipsePartialFootprints(
		time.Date(2025, 3, 29, 0, 0, 0, 0, time.UTC),
		SolarEclipsePartialFootprintOptions{Step: 30 * time.Minute, BoundaryPoints: 72},
	)
	if !ok {
		t.Fatalf("expected partial footprints for partial-only eclipse")
	}
	if footprints.Eclipse.Type != SolarEclipsePartial {
		t.Fatalf("unexpected eclipse type: got %s want %s", footprints.Eclipse.Type, SolarEclipsePartial)
	}
}

func TestSolarEclipsePartialFootprintsReturnFalseForNoEvent(t *testing.T) {
	_, ok := SolarEclipsePartialFootprints(time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC), SolarEclipsePartialFootprintOptions{})
	if ok {
		t.Fatalf("no eclipse should not return partial footprints")
	}
}

func TestSolarEclipsePartialAreaCompatibilityWrapper(t *testing.T) {
	date := time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC)
	options := SolarEclipsePartialAreaOptions{Step: 30 * time.Minute, BoundaryPoints: 72}
	compat, compatOK := SolarEclipsePartialArea(date, options)
	primary, primaryOK := SolarEclipsePartialFootprints(date, options)

	if compatOK != primaryOK {
		t.Fatalf("compat ok mismatch: got %v want %v", compatOK, primaryOK)
	}
	if compat.Eclipse.Type != primary.Eclipse.Type {
		t.Fatalf("compat type mismatch: got %s want %s", compat.Eclipse.Type, primary.Eclipse.Type)
	}
	if len(compat.Footprints) != len(primary.Footprints) {
		t.Fatalf("compat footprint count mismatch: got %d want %d", len(compat.Footprints), len(primary.Footprints))
	}
}

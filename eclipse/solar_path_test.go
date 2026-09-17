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
			Step:              30 * time.Minute,
			BoundaryPoints:    72,
			CentralShadowStep: 10 * time.Minute,
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
	if footprints.CentralShadowStep != 10*time.Minute || len(footprints.CentralShadowFootprints) == 0 {
		t.Fatalf("central shadow sampling mismatch: step=%s footprints=%d", footprints.CentralShadowStep, len(footprints.CentralShadowFootprints))
	}

	for _, item := range []struct {
		name string
		tm   time.Time
	}{
		{name: "Eclipse.GreatestEclipse", tm: footprints.Eclipse.GreatestEclipse},
		{name: "Footprints[0].Time", tm: footprints.Footprints[0].Time},
		{name: "Footprints[last].Time", tm: footprints.Footprints[len(footprints.Footprints)-1].Time},
		{name: "Boundary point", tm: footprints.Footprints[0].Boundaries[0][0].Time},
		{name: "P1 contact", tm: footprints.P1.Time},
		{name: "P2 contact", tm: footprints.P2.Time},
		{name: "P3 contact", tm: footprints.P3.Time},
		{name: "P4 contact", tm: footprints.P4.Time},
		{name: "U1 contact", tm: footprints.U1.Time},
		{name: "U2 contact", tm: footprints.U2.Time},
		{name: "U3 contact", tm: footprints.U3.Time},
		{name: "U4 contact", tm: footprints.U4.Time},
		{name: "Central shadow", tm: footprints.CentralShadowFootprints[0].Time},
	} {
		if item.tm.Location() != loc {
			t.Fatalf("%s location mismatch: got %q want %q", item.name, item.tm.Location(), loc)
		}
	}
}

func TestSolarEclipsePartialFootprintsExposePartialBandContours(t *testing.T) {
	location := time.FixedZone("UTC+08", 8*3600)
	result, ok := SolarEclipsePartialFootprints(
		time.Date(2009, time.July, 22, 9, 0, 0, 0, location),
		SolarEclipsePartialFootprintOptions{Step: 10 * time.Minute, BoundaryPoints: 24},
	)
	if !ok || len(result.PartialBandContours) == 0 {
		t.Fatalf("partial-band contours=%d ok=%v", len(result.PartialBandContours), ok)
	}
	for segmentIndex, segment := range result.PartialBandContours {
		for pointIndex, point := range segment {
			if point.Time.Location() != location {
				t.Fatalf("segment %d point %d location=%v, want input location", segmentIndex, pointIndex, point.Time.Location())
			}
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
	if !footprints.U1.Time.IsZero() || !footprints.U4.Time.IsZero() || len(footprints.CentralShadowFootprints) != 0 {
		t.Fatal("partial-only eclipse must not return central-shadow contacts or footprints")
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

func TestSolarEclipsePartialFootprints20140429NonCentralContactsMatchNASA(t *testing.T) {
	partial, ok := SolarEclipsePartialFootprints(
		time.Date(2014, time.April, 29, 0, 0, 0, 0, time.UTC),
		SolarEclipsePartialFootprintOptions{Step: 10 * time.Minute, BoundaryPoints: 96},
	)
	if !ok || partial.Eclipse.Centrality != SolarEclipseNonCentral {
		t.Fatal("expected the 2014-04-29 non-central annular eclipse")
	}
	for name, sample := range map[string]struct {
		got  time.Time
		want time.Time
	}{
		"U1": {partial.U1.Time, time.Date(2014, time.April, 29, 5, 57, 38, 0, time.UTC)},
		"U4": {partial.U4.Time, time.Date(2014, time.April, 29, 6, 9, 34, 0, time.UTC)},
	} {
		if difference := sample.got.Sub(sample.want); difference < -5*time.Second || difference > 5*time.Second {
			t.Fatalf("%s=%s, want %s within five seconds", name, sample.got, sample.want)
		}
	}
	if !partial.U2.Time.IsZero() || !partial.U3.Time.IsZero() {
		t.Fatalf("non-central eclipse has internal contacts U2=%s U3=%s", partial.U2.Time, partial.U3.Time)
	}
	if len(partial.CentralBandSegments) != 1 {
		t.Fatalf("central-band regions=%d, want one continuous non-central band", len(partial.CentralBandSegments))
	}
}

func TestSolarEclipsePartialFootprintsRetryTinyGrazingFootprints(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(1859, 2, 3, 0, 0, 0, 0, time.UTC),
		time.Date(1935, 1, 5, 0, 0, 0, 0, time.UTC),
	} {
		partial, ok := SolarEclipsePartialFootprints(date, SolarEclipsePartialFootprintOptions{
			Step: 10 * time.Minute, BoundaryPoints: 24, CentralShadowStep: 10 * time.Minute,
			DisableRiseSet: true,
		})
		if !ok || len(partial.Footprints) == 0 {
			t.Fatalf("%s tiny grazing partial eclipse unavailable at low boundary resolution", date.Format("2006-01-02"))
		}
		if partial.BoundaryPoints <= 24 {
			t.Fatalf("%s did not report the finer fallback boundary resolution: %d", date.Format("2006-01-02"), partial.BoundaryPoints)
		}
	}
}

func TestSolarEclipseCentralPathRecoversFromCoarseSampling(t *testing.T) {
	path, ok := SolarEclipseCentralPath(
		time.Date(2119, 3, 11, 0, 0, 0, 0, time.UTC),
		SolarEclipsePathOptions{Step: time.Hour},
	)
	if !ok || len(path.CenterLine) < 2 {
		t.Fatalf("coarse central path has %d points, want at least two", len(path.CenterLine))
	}
	if path.Step >= time.Hour {
		t.Fatalf("fallback central path retained coarse metadata step %s", path.Step)
	}
	for index := 1; index < len(path.CenterLine); index++ {
		if !path.CenterLine[index].Time.After(path.CenterLine[index-1].Time) {
			t.Fatalf("center-line times are not strictly increasing at %d", index)
		}
	}
}

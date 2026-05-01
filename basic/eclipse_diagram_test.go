package basic

import (
	"math"
	"testing"
	"time"
)

func TestLunarEclipseDiagramIncludesContacts(t *testing.T) {
	diagram := LunarEclipseDiagram(JDECalc(2026, 3, 3), LunarEclipseDiagramOptions{StepDays: 10.0 / 1440.0})
	if diagram.Eclipse.Type != LunarEclipseTotal {
		t.Fatalf("unexpected eclipse type: got %s want %s", diagram.Eclipse.Type, LunarEclipseTotal)
	}
	if diagram.MoonRadius != 1 {
		t.Fatalf("moon radius mismatch: got %.9f want 1", diagram.MoonRadius)
	}
	if !(diagram.PenumbraRadius > diagram.UmbraRadius && diagram.UmbraRadius > diagram.MoonRadius) {
		t.Fatalf(
			"unexpected radii: penumbra=%.9f umbra=%.9f moon=%.9f",
			diagram.PenumbraRadius,
			diagram.UmbraRadius,
			diagram.MoonRadius,
		)
	}

	labels := map[string]bool{}
	for _, point := range diagram.Points {
		if point.Label != "" {
			labels[point.Label] = true
		}
	}
	for _, label := range []string{"P1", "U1", "U2", "Greatest", "U3", "U4", "P4"} {
		if !labels[label] {
			t.Fatalf("missing label %s in diagram points", label)
		}
	}
}

func TestLocalSolarEclipseDiagramIncludesContacts(t *testing.T) {
	diagram := LocalSolarEclipseDiagram(
		TD2UT(Date2JDE(time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC)), true),
		-96.7970,
		32.7767,
		0,
		LocalSolarEclipseDiagramOptions{StepDays: 5.0 / 1440.0},
	)
	if diagram.Eclipse.Type == SolarEclipseNone {
		t.Fatalf("expected local solar eclipse")
	}
	if len(diagram.Frames) == 0 {
		t.Fatalf("expected diagram frames")
	}

	labels := map[string]bool{}
	for _, frame := range diagram.Frames {
		if frame.SunRadius != 1 {
			t.Fatalf("sun radius mismatch: got %.9f want 1", frame.SunRadius)
		}
		if !(frame.MoonRadius > 0) {
			t.Fatalf("invalid moon radius: %.9f", frame.MoonRadius)
		}
		if math.IsNaN(frame.MoonX) || math.IsNaN(frame.MoonY) {
			t.Fatalf("invalid moon position: x=%f y=%f", frame.MoonX, frame.MoonY)
		}
		if frame.Label != "" {
			labels[frame.Label] = true
		}
	}
	for _, label := range []string{"C1", "Greatest", "C4"} {
		if !labels[label] {
			t.Fatalf("missing label %s in diagram frames", label)
		}
	}
}

func TestLocalSolarEclipseDiagramTimesKeepCoincidentLabels(t *testing.T) {
	eclipse := LocalSolarEclipseResult{
		HasPartial:      true,
		HasCentral:      true,
		PartialStart:    1,
		GreatestEclipse: 2,
		CentralStart:    2,
		CentralEnd:      2,
		PartialEnd:      3,
	}
	times, _ := localSolarEclipseDiagramTimes(eclipse, 0.5)
	var coincident localSolarEclipseDiagramTime
	found := false
	for _, item := range times {
		if item.jde == 2 {
			coincident = item
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing coincident event time")
	}
	want := []string{"C2", "Greatest", "C3"}
	if len(coincident.labels) != len(want) {
		t.Fatalf("coincident labels = %v, want %v", coincident.labels, want)
	}
	for i, label := range want {
		if coincident.labels[i] != label {
			t.Fatalf("coincident labels = %v, want %v", coincident.labels, want)
		}
	}
	if got := localSolarEclipseDiagramPrimaryLabel(coincident.labels); got != "Greatest" {
		t.Fatalf("primary label = %q, want %q", got, "Greatest")
	}
}

func TestLunarEclipseDiagramTimesKeepCoincidentLabels(t *testing.T) {
	eclipse := LunarEclipseResult{
		HasPenumbral:   true,
		HasPartial:     true,
		HasTotal:       true,
		PenumbralStart: 1,
		PartialStart:   1.5,
		TotalStart:     2,
		Maximum:        2,
		TotalEnd:       2,
		PartialEnd:     2.5,
		PenumbralEnd:   3,
	}
	times, _ := lunarEclipseDiagramTimes(eclipse, 0.5)
	var coincident lunarEclipseDiagramTime
	found := false
	for _, item := range times {
		if item.jde == 2 {
			coincident = item
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing coincident event time")
	}
	want := []string{"U2", "Greatest", "U3"}
	if len(coincident.labels) != len(want) {
		t.Fatalf("coincident labels = %v, want %v", coincident.labels, want)
	}
	for i, label := range want {
		if coincident.labels[i] != label {
			t.Fatalf("coincident labels = %v, want %v", coincident.labels, want)
		}
	}
	if got := lunarEclipseDiagramPrimaryLabel(coincident.labels); got != "Greatest" {
		t.Fatalf("primary label = %q, want %q", got, "Greatest")
	}
}

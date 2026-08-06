package basic

import (
	"math"
	"testing"
	"time"
)

func TestPlanetOccultationDiagramSaturnIncludesFiveDynamicDiskStages(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, location)
	end := start.Add(24 * time.Hour)
	events, err := FindPlanetOccultations(
		start, end, OccultationSaturn, 104.52219613, 55.25401991, 0, OccultationSearchOptions{},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultations() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("FindPlanetOccultations() returned %d events, want 1", len(events))
	}
	diagram := PlanetOccultationDiagram(events[0], PlanetOccultationDiagramOptions{})
	if len(diagram.Frames) < 5 {
		t.Fatalf("PlanetOccultationDiagram() returned %d frames, want at least 5", len(diagram.Frames))
	}

	frames := make(map[string]PlanetOccultationDiagramFrame)
	for _, frame := range diagram.Frames {
		for _, label := range frame.Labels {
			frames[label] = frame
		}
	}
	for _, label := range []string{"C1", "C2", "Greatest", "C3", "C4"} {
		if _, ok := frames[label]; !ok {
			t.Fatalf("diagram is missing %s", label)
		}
	}
	for _, label := range []string{"C1", "C4"} {
		frame := frames[label]
		residual := frame.SeparationArcsec - frame.MoonRadiusArcsec - frame.PlanetRadiusArcsec
		if math.Abs(residual) > 0.1 {
			t.Errorf("%s outer-contact residual = %.6f arcsec", label, residual)
		}
	}
	for _, label := range []string{"C2", "C3"} {
		frame := frames[label]
		residual := frame.SeparationArcsec - frame.MoonRadiusArcsec + frame.PlanetRadiusArcsec
		if math.Abs(residual) > 0.1 {
			t.Errorf("%s inner-contact residual = %.6f arcsec", label, residual)
		}
	}
	if !frames["Greatest"].FullyOcculted {
		t.Fatal("greatest frame does not report a fully occulted Saturn disk")
	}
}

func TestPlanetOccultationDiagramRejectsIncompleteEvent(t *testing.T) {
	result := PlanetOccultationDiagram(PlanetOccultationInfo{Planet: OccultationSaturn}, PlanetOccultationDiagramOptions{})
	if len(result.Frames) != 0 {
		t.Fatalf("invalid event returned %d frames, want none", len(result.Frames))
	}
}

func TestPlanetOccultationDiagramRejectsTargetGeometryMismatch(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	events, err := FindPlanetOccultations(
		time.Date(2025, time.February, 1, 0, 0, 0, 0, location),
		time.Date(2025, time.February, 2, 0, 0, 0, 0, location),
		OccultationSaturn, 104.52219613, 55.25401991, 0, OccultationSearchOptions{},
	)
	if err != nil || len(events) != 1 {
		t.Fatalf("FindPlanetOccultations() = %d events, %v; want one", len(events), err)
	}
	// TargetID 只是显示元数据，但改变建模行星必须 / TargetID is display metadata, but changing the modeled planet must
	// 使事件几何失效，而不是静默绘制另一个天体 / invalidate the event geometry rather than silently drawing another body.
	event := events[0]
	event.Planet = OccultationVenus
	event.TargetID = "Venus"
	result := PlanetOccultationDiagram(event, PlanetOccultationDiagramOptions{})
	if len(result.Frames) != 0 {
		t.Fatalf("diagram accepted mismatched planet and returned %d frames", len(result.Frames))
	}
}

func TestPlanetOccultationDiagramPartialEventHasOnlyExternalStages(t *testing.T) {
	start := time.Date(2024, time.August, 21, 1, 30, 0, 0, time.UTC)
	end := time.Date(2024, time.August, 21, 4, 0, 0, 0, time.UTC)
	events, err := FindPlanetOccultations(
		start, end, OccultationSaturn, -30.072, -6.5, 0, OccultationSearchOptions{},
	)
	if err != nil || len(events) != 1 {
		t.Fatalf("FindPlanetOccultations() = %d events, %v; want one", len(events), err)
	}
	if events[0].Type != OccultationPartial {
		t.Fatalf("event type = %q, want partial", events[0].Type)
	}
	diagram := PlanetOccultationDiagram(events[0], PlanetOccultationDiagramOptions{})
	labels := make(map[string]bool)
	for _, frame := range diagram.Frames {
		for _, label := range frame.Labels {
			labels[label] = true
		}
	}
	for _, label := range []string{"C1", "Greatest", "C4"} {
		if !labels[label] {
			t.Fatalf("partial diagram is missing %s", label)
		}
	}
	if labels["C2"] || labels["C3"] {
		t.Fatalf("partial diagram contains internal contacts: %+v", labels)
	}
}

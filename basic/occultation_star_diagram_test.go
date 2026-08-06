package basic

import (
	"math"
	"testing"
	"time"
)

func TestStarOccultationDiagramUsesSolvedLocalContacts(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	star := StarCoordinate{
		ID:                             "HR 4799",
		RA:                             189.1975,
		Dec:                            -5.831944444444,
		Epoch:                          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28,
		ProperMotionDecMasPerYear:      -18,
	}
	events, err := FindStarOccultations(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		star, 121.56601, 6.80706, 0, OccultationSearchOptions{},
	)
	if err != nil {
		t.Fatalf("FindStarOccultations() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("FindStarOccultations() returned %d events, want 1", len(events))
	}

	diagram := StarOccultationDiagram(events[0], star, StarOccultationDiagramOptions{StepDays: 2.0 / 1440})
	if len(diagram.Frames) < 3 {
		t.Fatalf("StarOccultationDiagram() frame count = %d, want at least 3", len(diagram.Frames))
	}
	if !starOccultationDiagramFrameHasLabel(diagram.Frames[0], "Immersion") {
		t.Fatalf("first frame labels = %v, want Immersion", diagram.Frames[0].Labels)
	}
	last := diagram.Frames[len(diagram.Frames)-1]
	if !starOccultationDiagramFrameHasLabel(last, "Emersion") {
		t.Fatalf("last frame labels = %v, want Emersion", last.Labels)
	}
	if residual := math.Abs(diagram.Frames[0].SeparationArcsec - diagram.Frames[0].MoonRadiusArcsec); residual > 0.1 {
		t.Fatalf("immersion limb residual = %.6f arcsec, want <= 0.1", residual)
	}
	if residual := math.Abs(last.SeparationArcsec - last.MoonRadiusArcsec); residual > 0.1 {
		t.Fatalf("emersion limb residual = %.6f arcsec, want <= 0.1", residual)
	}
	greatest, ok := starOccultationDiagramFrameByLabel(diagram.Frames, "Greatest")
	if !ok {
		t.Fatalf("diagram does not contain Greatest frame")
	}
	if !greatest.BehindMoon || greatest.SeparationArcsec >= greatest.MoonRadiusArcsec {
		t.Fatalf("greatest frame is not behind Moon: separation=%.6f radius=%.6f", greatest.SeparationArcsec, greatest.MoonRadiusArcsec)
	}

	wrongStar := star
	wrongStar.RA += 30
	if frames := StarOccultationDiagram(events[0], wrongStar, StarOccultationDiagramOptions{}).Frames; len(frames) != 0 {
		t.Fatalf("same-ID wrong coordinate produced %d frames, want none", len(frames))
	}
	infoWithoutID := events[0]
	infoWithoutID.TargetID = ""
	wrongStar.ID = ""
	if frames := StarOccultationDiagram(infoWithoutID, wrongStar, StarOccultationDiagramOptions{}).Frames; len(frames) != 0 {
		t.Fatalf("empty-ID wrong coordinate produced %d frames, want none", len(frames))
	}
}

func starOccultationDiagramFrameHasLabel(frame StarOccultationDiagramFrame, label string) bool {
	for _, current := range frame.Labels {
		if current == label {
			return true
		}
	}
	return false
}

func starOccultationDiagramFrameByLabel(frames []StarOccultationDiagramFrame, label string) (StarOccultationDiagramFrame, bool) {
	for _, frame := range frames {
		if starOccultationDiagramFrameHasLabel(frame, label) {
			return frame, true
		}
	}
	return StarOccultationDiagramFrame{}, false
}

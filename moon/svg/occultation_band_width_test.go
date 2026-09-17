package svg

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

// 本文件钉住星掩图的"掩带宽"口径与行星图、详细版一致：取南北限地面间距，
// 不得与 Greatest.WidthKM 互相换算。

func TestStarOccultationSVGUsesLimitSeparationBandWidth(t *testing.T) {
	path := sampleStarOccultationPath()
	path.GreatestLimitSeparationKM = 3123.4
	path.Greatest.WidthKM = 3456.7
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{Width: 1200, Height: 800, Location: time.UTC})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	if !strings.Contains(diagram, "3123.4 km") {
		t.Fatal("stellar SVG band width does not use GreatestLimitSeparationKM")
	}
	if strings.Contains(diagram, "3456.7 km") {
		t.Fatal("stellar SVG band width still uses Greatest.WidthKM")
	}
}

func TestStarOccultationSVGBandWidthFallsBackToGreatestWidth(t *testing.T) {
	path := sampleStarOccultationPath()
	path.GreatestLimitSeparationKM = 0
	path.Greatest.WidthKM = 3456.7
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{Width: 1200, Height: 800, Location: time.UTC})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	if !strings.Contains(diagram, "3456.7 km") {
		t.Fatal("stellar SVG band width lost the Greatest.WidthKM fallback for paths without limit separations")
	}
}

func TestOccultationBandWidthSourcesAreNotInterchangeable(t *testing.T) {
	starPath := occultationTestStarPath(t)
	if starPath.GreatestLimitSeparationKM <= 0 || starPath.Greatest.WidthKM <= 0 {
		t.Fatalf("fixture widths = %.3f / %.3f, want both populated",
			starPath.GreatestLimitSeparationKM, starPath.Greatest.WidthKM)
	}
	if math.Abs(starPath.GreatestLimitSeparationKM-starPath.Greatest.WidthKM) < 1 {
		t.Fatalf("fixture widths %.3f / %.3f are equal; the semantics check would be vacuous",
			starPath.GreatestLimitSeparationKM, starPath.Greatest.WidthKM)
	}
	diagram, err := StarOccultationPathSVG(starPath, StarOccultationSVGOptions{Width: 1200, Height: 800, Location: time.UTC})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	expected := occultationTestWidthText(starPath.GreatestLimitSeparationKM)
	if !strings.Contains(diagram, expected) {
		t.Fatalf("stellar SVG is missing the limit-separation width %q", expected)
	}
	if other := occultationTestWidthText(starPath.Greatest.WidthKM); strings.Contains(diagram, other) {
		t.Fatalf("stellar SVG reports the non-interchangeable Greatest.WidthKM %q", other)
	}

	planetPath := occultationTestSaturnPath(t)
	planetDiagram, err := PlanetOccultationPathSVG(planetPath, StarOccultationSVGOptions{Width: 1200, Height: 800, Location: time.UTC})
	if err != nil {
		t.Fatalf("PlanetOccultationPathSVG() error = %v", err)
	}
	if expected := occultationTestWidthText(planetPath.GreatestLimitSeparationKM); !strings.Contains(planetDiagram, expected) {
		t.Fatalf("planetary SVG is missing the limit-separation width %q", expected)
	}
}

func occultationTestWidthText(width float64) string {
	return fmt.Sprintf("%.1f km", width)
}

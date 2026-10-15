package svg

// 月缘剖面精修过的接触时刻不落在圆月面上，站心图仍须能出图。

import (
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/moon/limb"
)

func TestLocalPlanetOccultationSVGAcceptsLunarLimbContacts(t *testing.T) {
	events, err := moon.FindPlanetOccultations(
		time.Date(2034, time.November, 21, 0, 0, 0, 0, time.UTC),
		time.Date(2034, time.November, 23, 0, 0, 0, 0, time.UTC),
		moon.OccultationJupiter, -42.36358358336173, 21.762064240049245, 0,
		moon.OccultationSearchOptions{LunarLimb: limb.Default()},
	)
	if err != nil || len(events) != 1 {
		t.Fatalf("FindPlanetOccultations() = %d events, %v; want one", len(events), err)
	}
	if !events[0].LunarLimbApplied {
		t.Fatal("LunarLimbApplied = false, want true")
	}
	if _, err := LocalPlanetOccultationSVG(events[0], LocalPlanetOccultationSVGOptions{}); err != nil {
		t.Fatalf("LocalPlanetOccultationSVG() error = %v", err)
	}
}

func TestLocalStarOccultationSVGAcceptsLunarLimbContacts(t *testing.T) {
	star := moon.StarCoordinate{
		ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
		Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC), Frame: moon.CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -10, ProperMotionDecMasPerYear: -20, ParallaxMas: 24,
	}
	date := time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC)
	events, err := moon.FindStarOccultations(date, date.Add(24*time.Hour), star,
		24.493397847811025, -74.18160307712625, 0,
		moon.OccultationSearchOptions{LunarLimb: limb.Default()})
	if err != nil || len(events) != 1 {
		t.Fatalf("FindStarOccultations() = %d events, %v; want one", len(events), err)
	}
	if _, err := LocalStarOccultationSVG(events[0], star, LocalStarOccultationSVGOptions{}); err != nil {
		t.Fatalf("LocalStarOccultationSVG() error = %v", err)
	}
}

func TestOccultationSVGsDeclareLunarLimbConvention(t *testing.T) {
	observer := [2]float64{-42.36358358336173, 21.762064240049245}
	start := time.Date(2034, time.November, 21, 0, 0, 0, 0, time.UTC)
	end := time.Date(2034, time.November, 23, 0, 0, 0, 0, time.UTC)
	draw := func(applied bool) string {
		t.Helper()
		options := moon.OccultationSearchOptions{}
		if applied {
			options.LunarLimb = limb.Default()
		}
		events, err := moon.FindPlanetOccultations(start, end, moon.OccultationJupiter,
			observer[0], observer[1], 0, options)
		if err != nil || len(events) != 1 {
			t.Fatalf("FindPlanetOccultations() = %d events, %v; want one", len(events), err)
		}
		if events[0].LunarLimbApplied != applied {
			t.Fatalf("LunarLimbApplied = %v, want %v", events[0].LunarLimbApplied, applied)
		}
		diagram, err := LocalPlanetOccultationSVG(events[0], LocalPlanetOccultationSVGOptions{Location: time.UTC})
		if err != nil {
			t.Fatalf("LocalPlanetOccultationSVG() error = %v", err)
		}
		return diagram
	}
	if circular := draw(false); !strings.Contains(circular, "圆月面") {
		t.Fatal("circular chart does not declare the circular limb")
	}
	if corrected := draw(true); !strings.Contains(corrected, "月缘剖面") {
		t.Fatal("limb chart does not declare the limb profile")
	}
}

func TestOccultationPathSVGsDeclareLunarLimbConvention(t *testing.T) {
	start := time.Date(2034, time.November, 21, 0, 0, 0, 0, time.UTC)
	end := time.Date(2034, time.November, 23, 0, 0, 0, 0, time.UTC)
	for _, applied := range []bool{false, true} {
		options := moon.OccultationPathOptions{Step: 10 * time.Minute, TargetSpacingKM: 400}
		if applied {
			options.LunarLimb = limb.Default()
		}
		paths, err := moon.FindPlanetOccultationPaths(start, end, moon.OccultationJupiter, options)
		if err != nil || len(paths) != 1 {
			t.Fatalf("FindPlanetOccultationPaths() = %d paths, %v; want one", len(paths), err)
		}
		if paths[0].LunarLimbApplied != applied {
			t.Fatalf("LunarLimbApplied = %v, want %v", paths[0].LunarLimbApplied, applied)
		}
		want := "月面按圆月面计算。"
		if applied {
			want = "界线按月缘剖面修正"
		}
		global, err := PlanetOccultationPathSVG(paths[0], PlanetOccultationSVGOptions{Location: time.UTC})
		if err != nil {
			t.Fatalf("PlanetOccultationPathSVG() error = %v", err)
		}
		if !strings.Contains(global, want) {
			t.Fatalf("global path chart (limb=%v) missing %q", applied, want)
		}
		detailed, err := PlanetOccultationDetailedSVG(paths[0], OccultationDetailedSVGOptions{Location: time.UTC})
		if err != nil {
			t.Fatalf("PlanetOccultationDetailedSVG() error = %v", err)
		}
		if !strings.Contains(detailed, want) {
			t.Fatalf("detailed chart (limb=%v) missing %q", applied, want)
		}
	}
}

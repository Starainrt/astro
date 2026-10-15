package svg

import (
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/moon"
)

type constantLimb float64

func (r constantLimb) LimbRadiusKM(_, _, _ float64) (float64, bool) { return float64(r), true }

type failingLimb struct{}

func (failingLimb) LimbRadiusKM(_, _, _ float64) (float64, bool) { return 0, false }

func TestLocalOccultationSVGAcceptsCustomLimbAndRejectsFailedLimbFlag(t *testing.T) {
	start := time.Date(2034, time.November, 21, 0, 0, 0, 0, time.UTC)
	end := time.Date(2034, time.November, 23, 0, 0, 0, 0, time.UTC)
	find := func(profile basic.LunarLimb) moon.PlanetOccultationInfo {
		t.Helper()
		events, err := moon.FindPlanetOccultations(start, end, moon.OccultationJupiter,
			-42.36358358336173, 21.762064240049245, 0, moon.OccultationSearchOptions{LunarLimb: profile})
		if err != nil || len(events) != 1 {
			t.Fatalf("FindPlanetOccultations() = %d events, %v; want one", len(events), err)
		}
		return events[0]
	}
	circular := find(nil)

	custom := find(constantLimb(1750))
	if !custom.LunarLimbApplied {
		t.Fatal("constant 1750 km limb: LunarLimbApplied = false, want true")
	}
	if custom.ExternalImmersion.Equal(circular.ExternalImmersion) {
		t.Fatal("constant 1750 km limb did not move the contact")
	}
	diagram, err := LocalPlanetOccultationSVG(custom, LocalPlanetOccultationSVGOptions{Location: time.UTC})
	if err != nil {
		t.Fatalf("LocalPlanetOccultationSVG(custom limb) error = %v", err)
	}
	if !strings.Contains(diagram, "月缘剖面") {
		t.Fatal("custom limb chart does not declare the limb profile")
	}

	failed := find(failingLimb{})
	if failed.LunarLimbApplied {
		t.Fatal("failing limb: LunarLimbApplied = true, want false")
	}
	if !failed.ExternalImmersion.Equal(circular.ExternalImmersion) || !failed.InternalEmersion.Equal(circular.InternalEmersion) {
		t.Fatal("failing limb changed the circular contacts")
	}
	diagram, err = LocalPlanetOccultationSVG(failed, LocalPlanetOccultationSVGOptions{Location: time.UTC})
	if err != nil {
		t.Fatalf("LocalPlanetOccultationSVG(failed limb) error = %v", err)
	}
	if strings.Contains(diagram, "月缘剖面") || !strings.Contains(diagram, "圆月面") {
		t.Fatal("failed limb chart must declare the circular limb")
	}
}

func TestOccultationPathLimbFlagRequiresCorrectedLimit(t *testing.T) {
	start := time.Date(2034, time.November, 21, 0, 0, 0, 0, time.UTC)
	end := time.Date(2034, time.November, 23, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(start, end, moon.OccultationJupiter,
		moon.OccultationPathOptions{Step: 10 * time.Minute, TargetSpacingKM: 400, LunarLimb: failingLimb{}})
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() = %d paths, %v; want one", len(paths), err)
	}
	if paths[0].LunarLimbApplied {
		t.Fatal("failing limb: path LunarLimbApplied = true, want false")
	}
}

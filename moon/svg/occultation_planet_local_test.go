package svg

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/moon"
)

func TestFindLocalPlanetOccultationSVGsSaturnFiveContacts(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	diagrams, err := FindLocalPlanetOccultationSVGs(
		time.Date(2025, time.February, 1, 0, 0, 0, 0, location),
		time.Date(2025, time.February, 2, 0, 0, 0, 0, location),
		moon.OccultationSaturn,
		104.52219613, 55.25401991, 0,
		moon.OccultationSearchOptions{},
		LocalPlanetOccultationSVGOptions{Width: 920, Height: 700, Location: location},
	)
	if err != nil {
		t.Fatalf("FindLocalPlanetOccultationSVGs() error = %v", err)
	}
	if len(diagrams) != 1 {
		t.Fatalf("FindLocalPlanetOccultationSVGs() returned %d diagrams, want 1", len(diagrams))
	}
	diagram := diagrams[0]
	for _, want := range []string{
		`<svg`, `width="920"`, `height="700"`, "2025-02-01", "指定地点月掩土星",
		"站心行星轨迹", "圆盘真实比例", "C1 外切始", "C2 内切始", "掩甚", "C3 内切终", "C4 外切终",
		"C1-C4 接触阶段", "行星圆盘放大示意", "本地接触时刻", "白道", "土星环不参与接触计算",
		`class="local-planet-track"`, `class="lunar-path-line"`, `class="overview-moon"`,
		`class="overview-planet-disk"`, `<symbol id="le-moon"`, `href="#le-moon"`,
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("local planetary SVG missing %q", want)
		}
	}
	if strings.Contains(diagram, "全球掩带") || strings.Contains(diagram, "全球中心线") {
		t.Fatal("local planetary SVG unexpectedly contains global-path labels")
	}
	if got := strings.Count(diagram, `class="stage-moon"`); got != 5 {
		t.Fatalf("stage Moon count = %d, want 5", got)
	}
	if got := strings.Count(diagram, `class="stage-planet-disk"`); got != 5 {
		t.Fatalf("stage planet count = %d, want 5", got)
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("generated local planetary SVG is not valid XML: %v", err)
	}
}

func TestLocalPlanetOccultationSVGUsesGreatestMoonPhase(t *testing.T) {
	info := localSaturnOccultation(t)
	diagram, err := LocalPlanetOccultationSVG(info, LocalPlanetOccultationSVGOptions{
		Location: time.UTC,
	})
	if err != nil {
		t.Fatalf("LocalPlanetOccultationSVG() error = %v", err)
	}
	appearance := localOccultationMoonAppearanceAt(info.Greatest, info.Observer)
	want := fmt.Sprintf(`data-illumination="%.6f" data-bright-limb-position-angle="%.3f"`,
		appearance.illumination, appearance.brightLimbPA)
	if !strings.Contains(diagram, want) {
		t.Fatalf("local planetary SVG does not carry greatest-time Moon appearance %q", want)
	}
	if appearance.illumination > 1e-6 && appearance.illumination < 1-1e-6 &&
		!strings.Contains(diagram, `<clipPath id="local-planet-occultation-moon-overview">`) {
		t.Fatal("non-full Moon overview is missing its phase clip path")
	}
}

func TestLocalPlanetOccultationSVGStageCentersPreserveDisplayedTangency(t *testing.T) {
	event := localPlanetOccultationEventFrame{
		label: "C2",
		frame: moon.PlanetOccultationDiagramFrame{
			SeparationArcsec: 970,
			PositionAngleDeg: 90,
		},
	}
	x, y := localPlanetOccultationSVGStageCenter(event, 100, 100, 40, 6, 0.04)
	if math.Abs(math.Hypot(x-100, y-100)-(40-6)) > 1e-12 {
		t.Fatalf("C2 display distance = %.12f, want %.12f", math.Hypot(x-100, y-100), float64(40-6))
	}
	event.label = "C4"
	x, y = localPlanetOccultationSVGStageCenter(event, 100, 100, 40, 6, 0.04)
	if math.Abs(math.Hypot(x-100, y-100)-(40+6)) > 1e-12 {
		t.Fatalf("C4 display distance = %.12f, want %.12f", math.Hypot(x-100, y-100), float64(40+6))
	}
}

func TestLocalPlanetOccultationSVGEnglishAndCustomText(t *testing.T) {
	info := localSaturnOccultation(t)
	diagram, err := LocalPlanetOccultationSVG(info, LocalPlanetOccultationSVGOptions{
		Language:         "en",
		Location:         time.UTC,
		Title:            "Custom local planet title",
		SummaryText:      "Custom local planet summary",
		GreatestText:     "Custom local planet greatest",
		OverviewTitle:    "Custom local planet overview",
		PhasePanelsTitle: "Custom local planet stages",
		ContactsTitle:    "Custom local planet contacts",
		DirectionText:    "Custom local planet direction",
		FooterNote:       "Custom local planet footer",
	})
	if err != nil {
		t.Fatalf("LocalPlanetOccultationSVG() error = %v", err)
	}
	for _, want := range []string{
		"Custom local planet title", "Custom local planet summary", "Custom local planet greatest",
		"Custom local planet overview", "Custom local planet stages", "Custom local planet contacts",
		"Custom local planet direction", "Custom local planet footer",
		"C1 external ingress", "C2 internal ingress", "Greatest", "C3 internal egress", "C4 external egress",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("English local planetary SVG missing %q", want)
		}
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("English local planetary SVG is not valid XML: %v", err)
	}
}

func TestLocalPlanetOccultationSVGRejectsInvalidInput(t *testing.T) {
	info := localSaturnOccultation(t)
	_, err := LocalPlanetOccultationSVG(info, LocalPlanetOccultationSVGOptions{Width: 1})
	if !errors.Is(err, ErrInvalidLocalPlanetOccultationSVGOptions) {
		t.Fatalf("invalid canvas error = %v, want ErrInvalidLocalPlanetOccultationSVGOptions", err)
	}
	info.ContactsComplete = false
	_, err = LocalPlanetOccultationSVG(info, LocalPlanetOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidLocalPlanetOccultationInfo) {
		t.Fatalf("incomplete event error = %v, want ErrInvalidLocalPlanetOccultationInfo", err)
	}
}

func TestLocalPlanetOccultationSVGRejectsTargetGeometryMismatch(t *testing.T) {
	info := localSaturnOccultation(t)
	info.Planet = moon.OccultationVenus
	info.TargetID = "Venus"
	_, err := LocalPlanetOccultationSVG(info, LocalPlanetOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidLocalPlanetOccultationInfo) {
		t.Fatalf("mismatched planet error = %v, want ErrInvalidLocalPlanetOccultationInfo", err)
	}
}

func TestLocalPlanetOccultationSVGPartialEventUsesThreeStages(t *testing.T) {
	events, err := moon.FindPlanetOccultations(
		time.Date(2024, time.August, 21, 1, 30, 0, 0, time.UTC),
		time.Date(2024, time.August, 21, 4, 0, 0, 0, time.UTC),
		moon.OccultationSaturn, -30.072, -6.5, 0, moon.OccultationSearchOptions{},
	)
	if err != nil || len(events) != 1 {
		t.Fatalf("FindPlanetOccultations() = %d events, %v; want one", len(events), err)
	}
	diagram, err := LocalPlanetOccultationSVG(events[0], LocalPlanetOccultationSVGOptions{Language: "en"})
	if err != nil {
		t.Fatalf("LocalPlanetOccultationSVG() error = %v", err)
	}
	if got := strings.Count(diagram, `class="stage-planet-disk"`); got != 3 {
		t.Fatalf("partial event stage count = %d, want 3", got)
	}
	if strings.Contains(diagram, "C2 internal ingress") || strings.Contains(diagram, "C3 internal egress") {
		t.Fatal("partial event SVG contains internal contacts")
	}
}

func localSaturnOccultation(t *testing.T) moon.PlanetOccultationInfo {
	t.Helper()
	location := time.FixedZone("UTC+8", 8*3600)
	events, err := moon.FindPlanetOccultations(
		time.Date(2025, time.February, 1, 0, 0, 0, 0, location),
		time.Date(2025, time.February, 2, 0, 0, 0, 0, location),
		moon.OccultationSaturn, 104.52219613, 55.25401991, 0, moon.OccultationSearchOptions{},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultations() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("FindPlanetOccultations() returned %d events, want 1", len(events))
	}
	return events[0]
}

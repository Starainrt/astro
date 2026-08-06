package svg

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/moon"
)

func TestFindLocalStarOccultationSVGsHR4799(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	diagrams, err := FindLocalStarOccultationSVGs(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		hr4799StarCoordinate(),
		121.56601, 6.80706, 0,
		moon.OccultationSearchOptions{},
		LocalStarOccultationSVGOptions{Width: 720, Height: 560, Location: location},
	)
	if err != nil {
		t.Fatalf("FindLocalStarOccultationSVGs() error = %v", err)
	}
	if len(diagrams) != 1 {
		t.Fatalf("FindLocalStarOccultationSVGs() returned %d diagrams, want 1", len(diagrams))
	}
	diagram := diagrams[0]
	for _, want := range []string{
		`<svg`, `width="720"`, `height="560"`, "2025-06-05", "HR 4799",
		"指定地点月掩", "观测点", "站心恒星轨迹", "掩始、掩甚与掩终视圆",
		"本地接触时刻", "掩始", "掩甚", "掩终", "左东右西", "UTC+8",
		`class="local-star-track"`, `class="overview-moon"`, `class="overview-star"`,
		`class="lunar-path-line"`, `class="lunar-path-label"`, "白道",
		`<symbol id="le-moon"`, `href="#le-moon"`,
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("local SVG missing %q", want)
		}
	}
	if strings.Contains(diagram, "local-occultation-moon") || strings.Contains(diagram, "<radialGradient") {
		t.Fatalf("local SVG still contains the temporary synthetic Moon rendering")
	}
	if strings.Contains(diagram, "全球掩带") || strings.Contains(diagram, "全球中心线") {
		t.Fatalf("local SVG unexpectedly contains global-path labels")
	}
	if got := strings.Count(diagram, `class="stage-moon"`); got != 3 {
		t.Fatalf("stage Moon count = %d, want 3", got)
	}
	if got := strings.Count(diagram, `class="stage-star"`); got != 3 {
		t.Fatalf("stage star count = %d, want 3", got)
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("generated local SVG is not valid XML: %v", err)
	}
}

func TestLocalStarOccultationSVGEnglishAndCustomText(t *testing.T) {
	info := localHR4799Occultation(t)
	diagram, err := LocalStarOccultationSVG(info, hr4799StarCoordinate(), LocalStarOccultationSVGOptions{
		Language:         "en",
		Location:         time.UTC,
		Title:            "Custom local title",
		SummaryText:      "Custom local summary",
		GreatestText:     "Custom local greatest",
		OverviewTitle:    "Custom local overview",
		PhasePanelsTitle: "Custom local stages",
		ContactsTitle:    "Custom local contacts",
		DirectionText:    "Custom local direction",
		FooterNote:       "Custom local footer",
	})
	if err != nil {
		t.Fatalf("LocalStarOccultationSVG() error = %v", err)
	}
	for _, want := range []string{
		"Custom local title", "Custom local summary", "Custom local greatest",
		"Custom local overview", "Custom local stages", "Custom local contacts",
		"Custom local direction", "Custom local footer", "Immersion", "Greatest", "Emersion", "Lunar path",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("English local SVG missing %q", want)
		}
	}
}

func TestLocalStarOccultationSVGLunarPathUsesGreatestTangent(t *testing.T) {
	frames := []moon.StarOccultationDiagramFrame{
		{StarXArcsec: 100, StarYArcsec: 100},
		{StarXArcsec: -1, StarYArcsec: -2},
		{StarXArcsec: 0, StarYArcsec: 0, Label: "Greatest", Labels: []string{"Greatest"}},
		{StarXArcsec: 1, StarYArcsec: 2},
		{StarXArcsec: -100, StarYArcsec: 100},
	}
	unitX, unitY, ok := localStarOccultationSVGLunarPathDirection(frames)
	if !ok {
		t.Fatalf("localStarOccultationSVGLunarPathDirection() returned no direction")
	}
	if math.Abs(unitX-1/math.Sqrt(5)) > 1e-12 || math.Abs(unitY-2/math.Sqrt(5)) > 1e-12 {
		t.Fatalf("lunar path direction = %.12f %.12f, want greatest tangent", unitX, unitY)
	}
}

func TestFindLocalStarOccultationSVGsNoEvent(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	star := hr4799StarCoordinate()
	star.Dec = 80
	diagrams, err := FindLocalStarOccultationSVGs(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		star, 121.56601, 6.80706, 0,
		moon.OccultationSearchOptions{}, LocalStarOccultationSVGOptions{},
	)
	if err != nil {
		t.Fatalf("FindLocalStarOccultationSVGs() error = %v", err)
	}
	if len(diagrams) != 0 {
		t.Fatalf("FindLocalStarOccultationSVGs() returned %d diagrams, want none", len(diagrams))
	}
}

func TestLocalStarOccultationSVGRejectsInvalidInput(t *testing.T) {
	info := localHR4799Occultation(t)
	_, err := LocalStarOccultationSVG(info, hr4799StarCoordinate(), LocalStarOccultationSVGOptions{Width: 1})
	if !errors.Is(err, ErrInvalidLocalStarOccultationSVGOptions) {
		t.Fatalf("invalid canvas error = %v, want ErrInvalidLocalStarOccultationSVGOptions", err)
	}
	info.ContactsComplete = false
	_, err = LocalStarOccultationSVG(info, hr4799StarCoordinate(), LocalStarOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidLocalStarOccultationInfo) {
		t.Fatalf("incomplete event error = %v, want ErrInvalidLocalStarOccultationInfo", err)
	}

	info = localHR4799Occultation(t)
	otherStar := hr4799StarCoordinate()
	otherStar.ID = "another-star"
	_, err = LocalStarOccultationSVG(info, otherStar, LocalStarOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidLocalStarOccultationInfo) {
		t.Fatalf("mismatched target error = %v, want ErrInvalidLocalStarOccultationInfo", err)
	}

	info = localHR4799Occultation(t)
	wrongCoordinate := hr4799StarCoordinate()
	wrongCoordinate.RA += 30
	_, err = LocalStarOccultationSVG(info, wrongCoordinate, LocalStarOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidLocalStarOccultationInfo) {
		t.Fatalf("same-ID wrong coordinate error = %v, want ErrInvalidLocalStarOccultationInfo", err)
	}

	info.TargetID = ""
	wrongCoordinate.ID = ""
	_, err = LocalStarOccultationSVG(info, wrongCoordinate, LocalStarOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidLocalStarOccultationInfo) {
		t.Fatalf("empty-ID wrong coordinate error = %v, want ErrInvalidLocalStarOccultationInfo", err)
	}
}

func localHR4799Occultation(t *testing.T) moon.StarOccultationInfo {
	t.Helper()
	location := time.FixedZone("CST", 8*3600)
	events, err := moon.FindStarOccultations(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		hr4799StarCoordinate(), 121.56601, 6.80706, 0, moon.OccultationSearchOptions{},
	)
	if err != nil {
		t.Fatalf("FindStarOccultations() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("FindStarOccultations() returned %d events, want 1", len(events))
	}
	return events[0]
}

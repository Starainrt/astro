package svg

import (
	"encoding/xml"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/moon"
)

func TestFindPlanetOccultationSVGsSaturnFiniteDisk(t *testing.T) {
	diagrams, err := FindPlanetOccultationSVGs(
		time.Date(2024, time.August, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.August, 22, 0, 0, 0, 0, time.UTC),
		moon.OccultationSaturn,
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
		PlanetOccultationSVGOptions{Width: 720, Height: 520},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultationSVGs() error = %v", err)
	}
	if len(diagrams) != 1 {
		t.Fatalf("FindPlanetOccultationSVGs() returned %d diagrams, want 1", len(diagrams))
	}
	diagram := diagrams[0]
	for _, want := range []string{
		`<svg`, `width="720"`, `height="520"`, "2024-08-21", "土星",
		"全球掩带", "外掩始", "全掩始", "掩甚", "全掩终", "外掩终", "外切锥面", "内切锥面",
		"部分掩带", "全掩带", `class="occultation-band"`, `class="total-occultation-band"`,
		`fill-rule="nonzero"`, `class="center-line"`,
		`class="occultation-time-marker"`,
		`class="event-marker event-total-start"`, `class="event-marker event-greatest"`,
		`class="event-marker event-total-end"`, `class="event-marker-leader"`,
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("planetary SVG missing %q", want)
		}
	}
	for _, obsolete := range []string{
		`class="northern-limit"`, `class="southern-limit"`,
		`class="northern-total-limit"`, `class="southern-total-limit"`,
	} {
		if strings.Contains(diagram, obsolete) {
			t.Fatalf("planetary footprint sweep retains obsolete approximate line %q", obsolete)
		}
	}
	if strings.Contains(diagram, "月掩Saturn") {
		t.Fatal("Chinese planetary SVG title contains the English default target name")
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("generated planetary SVG is not valid XML: %v", err)
	}
	for _, pair := range [][2]string{{"外掩始", "全掩始"}, {"全掩终", "外掩终"}} {
		_, firstY := planetOccultationSVGTextPosition(t, diagram, pair[0])
		_, secondY := planetOccultationSVGTextPosition(t, diagram, pair[1])
		if math.Abs(firstY-secondY) < 13 {
			t.Fatalf("planetary SVG labels %q and %q overlap vertically at y %.3f / %.3f", pair[0], pair[1], firstY, secondY)
		}
	}
	if got := strings.Count(diagram, `class="event-marker-leader"`); got != 4 {
		t.Fatalf("planetary SVG close-contact leader count = %d, want 4", got)
	}
	if got := strings.Count(diagram, `fill="#087f8c" stroke="#ffffff" stroke-width="1.0"`); got < 2 {
		t.Fatalf("planetary SVG total-contact marker count = %d, want at least 2", got)
	}
}

func TestPlanetOccultationPathSVGWrapsFivePhaseSummaryAtRenderedFontSize(t *testing.T) {
	path := samplePlanetOccultationPath()
	options := normalizeStarOccultationSVGOptions(PlanetOccultationSVGOptions{
		Width:       1100,
		SummaryText: strings.Repeat("A", 130),
	})
	options = planetOccultationSVGDefaults(path, options)
	lines := starOccultationSVGHeaderLines(planetOccultationStarShape(path, path.TargetID), options)
	if len(lines) < 3 {
		t.Fatalf("planetary SVG header lines = %d, want wrapped summary plus greatest line", len(lines))
	}
	if width := starOccultationTextWidth(lines[0], 14); width > float64(options.Width)-80 {
		t.Fatalf("planetary SVG first header line width = %.1f, canvas allowance %.1f", width, float64(options.Width)-80)
	}
}

func TestPlanetOccultationPathSVGUsesLargeGlobalMapDefault(t *testing.T) {
	diagram, err := PlanetOccultationPathSVG(samplePlanetOccultationPath(), PlanetOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("PlanetOccultationPathSVG() error = %v", err)
	}
	for _, want := range []string{
		`width="1200" height="800"`,
		`preserveAspectRatio="xMidYMid meet"`,
		`style="max-width:100%;height:auto;display:block"`,
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("planetary SVG missing large responsive canvas marker %q", want)
		}
	}
}

func planetOccultationSVGTextPosition(t *testing.T, diagram, label string) (float64, float64) {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(diagram))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("decode planetary SVG: %v", err)
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "text" {
			continue
		}
		var text string
		if err := decoder.DecodeElement(&text, &start); err != nil {
			t.Fatalf("decode planetary SVG text: %v", err)
		}
		if text != label {
			continue
		}
		values := map[string]float64{}
		for _, attribute := range start.Attr {
			if attribute.Name.Local != "x" && attribute.Name.Local != "y" {
				continue
			}
			value, parseErr := strconv.ParseFloat(attribute.Value, 64)
			if parseErr != nil {
				t.Fatalf("parse %s coordinate %q: %v", attribute.Name.Local, attribute.Value, parseErr)
			}
			values[attribute.Name.Local] = value
		}
		return values["x"], values["y"]
	}
	t.Fatalf("planetary SVG label %q not found", label)
	return 0, 0
}

func TestPlanetOccultationPathSVGRejectsInvalidCanvasWithPlanetError(t *testing.T) {
	path := samplePlanetOccultationPath()
	_, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{Width: 1})
	if !errors.Is(err, ErrInvalidPlanetOccultationSVGOptions) {
		t.Fatalf("PlanetOccultationPathSVG() error = %v, want ErrInvalidPlanetOccultationSVGOptions", err)
	}
}

func TestPlanetOccultationPathSVGEnglishAndCustomText(t *testing.T) {
	diagram, err := PlanetOccultationPathSVG(samplePlanetOccultationPath(), PlanetOccultationSVGOptions{
		Language:     "en",
		Location:     time.UTC,
		Title:        "Custom planetary occultation",
		SummaryText:  "Custom summary",
		GreatestText: "Custom greatest",
		FooterNote:   "Custom footer",
	})
	if err != nil {
		t.Fatalf("PlanetOccultationPathSVG() error = %v", err)
	}
	for _, want := range []string{
		"Custom planetary occultation", "Custom summary", "Custom greatest", "Custom footer",
		"Partial begins", "Total begins", "Greatest", "Total ends", "Partial ends",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("English planetary SVG missing %q", want)
		}
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("English planetary SVG is not valid XML: %v", err)
	}
}

func TestPlanetOccultationPathSVGRejectsInvalidPath(t *testing.T) {
	_, err := PlanetOccultationPathSVG(moon.PlanetOccultationPath{Planet: moon.OccultationSaturn}, PlanetOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidPlanetOccultationPath) {
		t.Fatalf("PlanetOccultationPathSVG() error = %v, want ErrInvalidPlanetOccultationPath", err)
	}
	path := samplePlanetOccultationPath()
	path.TotalComplete = false
	_, err = PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidPlanetOccultationPath) {
		t.Fatalf("incomplete total band error = %v, want ErrInvalidPlanetOccultationPath", err)
	}
}

func TestPlanetOccultationPathSVGSupportsPartialOnlyGlobalBand(t *testing.T) {
	path := samplePlanetOccultationPath()
	path.HasTotalBand = false
	path.TotalStart = moon.OccultationPathPoint{}
	path.TotalEnd = moon.OccultationPathPoint{}
	path.TotalComplete = false
	path.NorthernTotalLimit = nil
	path.SouthernTotalLimit = nil
	path.GreatestTotalWidthKM = 0
	diagram, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("PlanetOccultationPathSVG() partial-only error = %v", err)
	}
	if !strings.Contains(diagram, "部分掩带") || strings.Contains(diagram, `class="total-occultation-band"`) ||
		strings.Contains(diagram, "全掩始") || strings.Contains(diagram, "全掩终") {
		t.Fatal("partial-only planetary SVG contains inconsistent total-band content")
	}
	if got := strings.Count(diagram, `class="event-marker `); got != 3 {
		t.Fatalf("partial-only global marker count = %d, want 3", got)
	}
}

func samplePlanetOccultationPath() moon.PlanetOccultationPath {
	starPath := sampleStarOccultationPath()
	totalStart := starOccultationInterpolatePathPoint(
		starPath.CenterLine[0], starPath.CenterLine[1], 0.5, 167.5,
	)
	totalEnd := starOccultationInterpolatePathPoint(
		starPath.CenterLine[2], starPath.CenterLine[3], 0.5, -167.5,
	)
	northernTotal := []moon.OccultationPathPoint{totalStart}
	southernTotal := []moon.OccultationPathPoint{totalStart}
	for _, point := range starPath.CenterLine[1:3] {
		north, south := point, point
		north.Latitude += 10
		south.Latitude -= 10
		north.WidthKM = 0
		south.WidthKM = 0
		northernTotal = append(northernTotal, north)
		southernTotal = append(southernTotal, south)
	}
	northernTotal = append(northernTotal, totalEnd)
	southernTotal = append(southernTotal, totalEnd)
	return moon.PlanetOccultationPath{
		Planet:               moon.OccultationSaturn,
		TargetID:             "Saturn",
		Start:                starPath.Start,
		Greatest:             starPath.Greatest,
		End:                  starPath.End,
		Complete:             true,
		CenterLine:           starPath.CenterLine,
		NorthernLimit:        starPath.NorthernLimit,
		SouthernLimit:        starPath.SouthernLimit,
		HasTotalBand:         true,
		TotalStart:           totalStart,
		TotalEnd:             totalEnd,
		TotalComplete:        true,
		NorthernTotalLimit:   northernTotal,
		SouthernTotalLimit:   southernTotal,
		GreatestTotalWidthKM: 3100,
		Step:                 starPath.Step,
		TargetSpacingKM:      starPath.TargetSpacingKM,
	}
}

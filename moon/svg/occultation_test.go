package svg

import (
	"encoding/xml"
	"errors"
	"io"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/svgmap"
	"github.com/starainrt/astro/moon"
)

func TestFindStarOccultationSVGsHR4799(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	diagrams, err := FindStarOccultationSVGs(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		hr4799StarCoordinate(),
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
		StarOccultationSVGOptions{Width: 720, Height: 520},
	)
	if err != nil {
		t.Fatalf("FindStarOccultationSVGs() error = %v", err)
	}
	if len(diagrams) != 1 {
		t.Fatalf("FindStarOccultationSVGs() returned %d diagrams, want 1", len(diagrams))
	}
	diagram := diagrams[0]
	for _, want := range []string{
		`<svg`, `width="720"`, `height="520"`, "2025-06-05", "HR 4799",
		"全球掩带", "掩始", "掩甚", "掩终", "掩带宽", "UTC+8",
		`class="occultation-band"`, `class="center-line"`, `class="northern-limit"`,
		`class="southern-limit"`, `class="event-marker event-greatest"`, `class="land"`,
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("SVG missing %q", want)
		}
	}
	if got := strings.Count(diagram, `class="occultation-band"`); got != 2 {
		t.Fatalf("HR 4799 occultation-band segment count = %d, want two antimeridian-clipped fragments", got)
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("generated SVG is not valid XML: %v", err)
	}
}

func TestPolarOccultationLayoutSeparatesLegendAndFooter(t *testing.T) {
	layout := starOccultationSVGLayoutFor(
		StarOccultationSVGOptions{Width: 900, Height: 760},
		110,
		svgmap.ProjectionNorthPolar,
	)
	legendY := layout.mapY + layout.mapHeight + 30
	if gap := layout.footerY - legendY; gap < 24 {
		t.Fatalf("polar legend/footer gap = %.1f px, want at least 24 px", gap)
	}
}

func TestStarOccultationPathSVGEnglishAndCustomText(t *testing.T) {
	path := sampleStarOccultationPath()
	path.TargetID = "Alpha < Beta & Gamma"
	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{
		Language:      "en",
		Location:      time.UTC,
		Title:         "Custom <occultation> & title",
		SummaryText:   "Custom summary",
		GreatestText:  "Custom greatest",
		MapTitle:      "Custom map",
		ContactsTitle: "Custom events",
		FooterNote:    "Custom footer",
	})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	for _, want := range []string{
		"Custom &lt;occultation&gt; &amp; title",
		"Custom summary",
		"Custom greatest",
		"Custom map",
		"Custom events",
		"Custom footer",
		"Start",
		"Greatest",
		"End",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("SVG missing %q", want)
		}
	}
	if strings.Contains(diagram, "Custom <occultation>") {
		t.Fatal("custom title was not XML escaped")
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("generated SVG is not valid XML: %v", err)
	}
}

func TestStarOccultationPathSVGIncludesAlignedTimeLabels(t *testing.T) {
	diagram, err := StarOccultationPathSVG(sampleStarOccultationPath(), StarOccultationSVGOptions{
		Location: time.UTC, TimeLabelStep: 30 * time.Minute,
	})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	for _, want := range []string{`class="occultation-time-marker"`, `>10:30</text>`, `>12:30</text>`} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("stellar occultation SVG missing time marker %q", want)
		}
	}
	markers := occultationTimeMarkerPoints(
		sampleStarOccultationPath().CenterLine,
		30*time.Minute,
		time.UTC,
		[]time.Time{sampleStarOccultationPath().Start.Time, sampleStarOccultationPath().End.Time},
	)
	foundGreatestTime := false
	for _, marker := range markers {
		if marker.Time.Equal(sampleStarOccultationPath().Greatest.Time) {
			foundGreatestTime = true
			break
		}
	}
	if !foundGreatestTime {
		t.Fatal("aligned time at greatest was discarded instead of being placed below the event label")
	}
}

func TestStarOccultationPathSVGCanDisableTimeLabels(t *testing.T) {
	diagram, err := StarOccultationPathSVG(sampleStarOccultationPath(), StarOccultationSVGOptions{TimeLabelStep: -1})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	if strings.Contains(diagram, `class="occultation-time-marker"`) {
		t.Fatal("disabled occultation time labels were rendered")
	}
}

func TestStarOccultationPathSVGSplitsAntimeridian(t *testing.T) {
	diagram, err := StarOccultationPathSVG(sampleStarOccultationPath(), StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	if got := strings.Count(diagram, `class="center-line"`); got != 2 {
		t.Fatalf("center-line segment count = %d, want 2", got)
	}
	if got := strings.Count(diagram, `class="occultation-band"`); got != 2 {
		t.Fatalf("occultation-band segment count = %d, want 2", got)
	}
	segments := starOccultationPathSegments(sampleStarOccultationPath().CenterLine)
	if len(segments) != 2 {
		t.Fatalf("path segment count = %d, want 2", len(segments))
	}
	if segments[0][len(segments[0])-1].Longitude != 180 || segments[1][0].Longitude != -180 {
		t.Fatalf("antimeridian interpolation = %.3f / %.3f, want +180 / -180",
			segments[0][len(segments[0])-1].Longitude, segments[1][0].Longitude)
	}
	if !segments[0][len(segments[0])-1].Time.Equal(segments[1][0].Time) {
		t.Fatal("antimeridian split points do not share the interpolated time")
	}
}

func TestStarOccultationPathSVGUsesDetailedPhysicalLand(t *testing.T) {
	diagram, err := StarOccultationPathSVG(sampleStarOccultationPath(), StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	for _, want := range []string{
		`class="land-layer"`,
		`class="land"`,
		`fill="#d8d9d2"`,
		`stroke="#a6aaa4"`,
		`fill-rule="evenodd"`,
		`vector-effect="non-scaling-stroke"`,
		"Natural Earth 1:50m",
		"不含行政边界",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("SVG missing detailed-land marker %q", want)
		}
	}
	if got := strings.Count(diagram, `class="land"`); got != 1 {
		t.Fatalf("land path count = %d, want one compact path", got)
	}
	if strings.Contains(diagram, `fill="#000`) || strings.Contains(diagram, `fill="black"`) {
		t.Fatal("land path uses a black fill")
	}
}

func TestStarOccultationPathSVGUsesNorthPolarProjection(t *testing.T) {
	path := sampleStarOccultationPath()
	path.Greatest.Latitude = 72
	for index := range path.CenterLine {
		path.CenterLine[index].Latitude = 62 + float64(index)*4
	}
	for index := range path.NorthernLimit {
		path.NorthernLimit[index].Latitude = 68 + float64(index)*3
		path.SouthernLimit[index].Latitude = 58 + float64(index)*3
	}
	path.Start.Latitude = path.NorthernLimit[0].Latitude
	path.End.Latitude = path.NorthernLimit[len(path.NorthernLimit)-1].Latitude

	diagram, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("StarOccultationPathSVG() error = %v", err)
	}
	for _, want := range []string{`<circle class="map-ocean"`, `<circle class="map-frame"`, "北极方位等距投影"} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("north-polar SVG missing %q", want)
		}
	}
	if err := validateXML(diagram); err != nil {
		t.Fatalf("north-polar SVG is not valid XML: %v", err)
	}
}

func TestStarOccultationBandSegmentsPreserveAsymmetricAntimeridianCrossing(t *testing.T) {
	start := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	northern := []moon.OccultationPathPoint{
		{Time: start, Longitude: 170, Latitude: 20},
		{Time: start.Add(time.Hour), Longitude: -170, Latitude: 10},
	}
	southern := []moon.OccultationPathPoint{
		{Time: start, Longitude: 150, Latitude: 0},
		{Time: start.Add(time.Hour), Longitude: 160, Latitude: -10},
	}
	segments := starOccultationBandSegments(northern, southern)
	if len(segments) != 2 {
		t.Fatalf("asymmetric antimeridian band segment count = %d, want 2", len(segments))
	}
	hasWest, hasEast := false, false
	for _, segment := range segments {
		for _, point := range segment {
			hasWest = hasWest || point.longitude < -179
			hasEast = hasEast || point.longitude > 179
		}
	}
	if !hasWest || !hasEast {
		t.Fatalf("split band does not cover both map edges: west=%v east=%v", hasWest, hasEast)
	}
}

func TestFindStarOccultationSVGsNoEvent(t *testing.T) {
	diagrams, err := FindStarOccultationSVGs(
		time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC),
		moon.StarCoordinate{
			ID:    "polar-star",
			RA:    0,
			Dec:   89,
			Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
			Frame: moon.CoordinateFrameICRS,
		},
		moon.OccultationPathOptions{},
		StarOccultationSVGOptions{},
	)
	if err != nil {
		t.Fatalf("FindStarOccultationSVGs() error = %v", err)
	}
	if len(diagrams) != 0 {
		t.Fatalf("FindStarOccultationSVGs() returned %d diagrams, want none", len(diagrams))
	}
}

func TestStarOccultationPathSVGRejectsInvalidPath(t *testing.T) {
	path := sampleStarOccultationPath()
	path.Greatest.Longitude = math.NaN()
	_, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidStarOccultationPath) {
		t.Fatalf("StarOccultationPathSVG() error = %v, want ErrInvalidStarOccultationPath", err)
	}
}

func TestStarOccultationPathSVGRejectsMisalignedLimits(t *testing.T) {
	path := sampleStarOccultationPath()
	path.SouthernLimit[1].Time = path.SouthernLimit[1].Time.Add(time.Second)
	_, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidStarOccultationPath) {
		t.Fatalf("StarOccultationPathSVG() error = %v, want ErrInvalidStarOccultationPath", err)
	}
}

func TestStarOccultationPathSVGRejectsCanvasTooSmall(t *testing.T) {
	_, err := StarOccultationPathSVG(sampleStarOccultationPath(), StarOccultationSVGOptions{Width: 1, Height: 1})
	if !errors.Is(err, ErrInvalidStarOccultationSVGOptions) {
		t.Fatalf("StarOccultationPathSVG() error = %v, want ErrInvalidStarOccultationSVGOptions", err)
	}
}

func hr4799StarCoordinate() moon.StarCoordinate {
	return moon.StarCoordinate{
		ID:                             "HR 4799",
		RA:                             189.1975,
		Dec:                            -5.831944444444,
		Epoch:                          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          moon.CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28,
		ProperMotionDecMasPerYear:      -18,
	}
}

func sampleStarOccultationPath() moon.StarOccultationPath {
	start := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	center := []moon.OccultationPathPoint{
		{Time: start, Longitude: 160, Latitude: 18, MoonAltitude: 5, WidthKM: 3200},
		{Time: start.Add(time.Hour), Longitude: 175, Latitude: 10, MoonAltitude: 35, WidthKM: 3300},
		{Time: start.Add(2 * time.Hour), Longitude: -175, Latitude: 1, MoonAltitude: 50, WidthKM: 3400},
		{Time: start.Add(3 * time.Hour), Longitude: -160, Latitude: -8, MoonAltitude: 12, WidthKM: 3300},
	}
	northern := make([]moon.OccultationPathPoint, len(center))
	southern := make([]moon.OccultationPathPoint, len(center))
	for index, point := range center {
		northern[index] = point
		northern[index].Latitude += 12
		northern[index].WidthKM = 0
		southern[index] = point
		southern[index].Latitude -= 12
		southern[index].WidthKM = 0
	}
	return moon.StarOccultationPath{
		TargetID:      "synthetic-star",
		Start:         center[0],
		Greatest:      center[1],
		End:           center[3],
		Complete:      true,
		CenterLine:    center,
		NorthernLimit: northern,
		SouthernLimit: southern,
		Step:          time.Hour,
	}
}

func validateXML(value string) error {
	decoder := xml.NewDecoder(strings.NewReader(value))
	for {
		if _, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

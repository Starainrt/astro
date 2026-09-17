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
		`class="occultation-rise-set-boundary occultation-start-rise"`, `stroke="#d97706"`, "初掩/掩甚/终掩月升月落线",
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
	bandIndex := strings.Index(diagram, `class="occultation-band"`)
	curveIndex := strings.Index(diagram, `class="occultation-rise-set-boundary occultation-start-rise"`)
	if bandIndex < 0 || curveIndex < 0 || curveIndex < bandIndex {
		t.Fatalf("planetary SVG draws rise/set curves before the filled band: band=%d curve=%d", bandIndex, curveIndex)
	}
	curveEnd := -1
	if curveIndex >= 0 {
		curveEnd = strings.Index(diagram[curveIndex:], "/>")
	}
	if curveEnd < 0 || strings.Contains(diagram[curveIndex:curveIndex+curveEnd], "stroke-dasharray") {
		t.Fatal("planetary rise/set phase line is rendered with a gap-producing dash pattern")
	}
}

func TestPlanetOccultationLegendOmitsDisabledRiseSetCurves(t *testing.T) {
	path := samplePlanetOccultationPath()
	path.RiseSetCurves = nil
	diagram, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("PlanetOccultationPathSVG: %v", err)
	}
	if strings.Contains(diagram, "初掩/掩甚/终掩月升月落线") || strings.Contains(diagram, "Rise/set phase lines") {
		t.Fatal("planetary SVG legend claims disabled rise/set curves are present")
	}
}

func TestPlanetOccultationSVGWithoutFootprintsUsesLimitBands(t *testing.T) {
	path := samplePlanetOccultationPath()
	path.PartialFootprints = nil
	path.TotalFootprints = nil
	diagram, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{})
	if err != nil {
		t.Fatalf("PlanetOccultationPathSVG: %v", err)
	}
	for _, className := range []string{
		`class="occultation-band"`, `class="total-occultation-band"`,
		`class="northern-limit"`, `class="southern-limit"`,
		`class="northern-total-limit"`, `class="southern-total-limit"`,
	} {
		if !strings.Contains(diagram, className) {
			t.Fatalf("planetary SVG without footprints is missing %s", className)
		}
	}
}

func TestPlanetOccultationSVG20240725UsesCompactBandFootprints(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			DisableRiseSet: true, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if len(path.PartialFootprints) != 0 || len(path.TotalFootprints) != 0 ||
		len(path.PartialBandFootprints) == 0 || len(path.TotalBandFootprints) == 0 {
		t.Fatalf("dense/compact footprint counts partial=%d/%d total=%d/%d",
			len(path.PartialFootprints), len(path.PartialBandFootprints),
			len(path.TotalFootprints), len(path.TotalBandFootprints))
	}
	diagram, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{
		Width: 1200, Height: 800, Projection: MapProjectionEquirectangular,
	})
	if err != nil {
		t.Fatalf("PlanetOccultationPathSVG: %v", err)
	}
	for _, className := range []string{
		`class="northern-limit"`, `class="southern-limit"`,
		`class="northern-total-limit"`, `class="southern-total-limit"`,
	} {
		if strings.Contains(diagram, className) {
			t.Fatalf("compact planetary SVG still overlays discontinuous auxiliary line %s", className)
		}
	}
	x, y := planetOccultationSVGMapPoint(
		t, diagram, MapProjectionEquirectangular, path.Greatest.Longitude, path.Greatest.Latitude,
	)
	for _, className := range []string{`class="occultation-band"`, `class="total-occultation-band"`} {
		pathData := planetOccultationSVGPathData(t, diagram, strings.TrimSuffix(strings.TrimPrefix(className, `class="`), `"`))
		if !strings.Contains(pathData, "Z") {
			t.Fatalf("compact planetary SVG %s is not explicitly closed", className)
		}
		if !planetOccultationSVGPathContains(pathData, x, y) {
			t.Fatalf("compact planetary SVG %s excludes greatest at %.3f, %.3f", className, x, y)
		}
	}
}

func TestPlanetOccultationSVGMars20250729MergesTotalRiseSetBoundaryBand(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationMars,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if len(path.TotalBandFootprints) < 20 {
		t.Fatalf("total compact footprint count=%d, want sparse support samples", len(path.TotalBandFootprints))
	}
	if len(path.TotalRiseSetCurves) == 0 {
		t.Fatal("total compact band is missing inner-contact rise/set curves")
	}
	diagram, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{
		Width: 1200, Height: 800, Projection: MapProjectionEquirectangular,
	})
	if err != nil {
		t.Fatalf("PlanetOccultationPathSVG: %v", err)
	}
	pathData := planetOccultationSVGPathData(t, diagram, "total-occultation-band")
	subpaths := strings.Count(pathData, "M ")
	if subpaths >= len(path.TotalBandFootprints)/2 {
		t.Fatalf("Mars 2025-07-29 total band has %d subpaths for %d compact footprints, want merged boundary geometry",
			subpaths, len(path.TotalBandFootprints))
	}
	if !strings.Contains(pathData, "Z") {
		t.Fatal("Mars 2025-07-29 total boundary band is not explicitly closed")
	}
}

func TestPlanetOccultationPathSVGJupiter20230517PolarBandContainsPole(t *testing.T) {
	start := time.Date(2023, time.May, 17, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationJupiter,
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one Jupiter path", len(paths), err)
	}
	for _, projection := range []struct {
		name                string
		value               MapProjection
		longitude, latitude float64
	}{
		{name: "north-polar", value: MapProjectionNorthPolar, longitude: 0, latitude: 90},
		{name: "equirectangular", value: MapProjectionEquirectangular, longitude: 0, latitude: 89},
	} {
		diagram, renderErr := PlanetOccultationPathSVG(paths[0], PlanetOccultationSVGOptions{
			Width: 1200, Height: 800, Projection: projection.value,
		})
		if renderErr != nil {
			t.Fatalf("%s PlanetOccultationPathSVG() error = %v", projection.name, renderErr)
		}
		x, y := planetOccultationSVGMapPoint(t, diagram, projection.value, projection.longitude, projection.latitude)
		for _, className := range []string{"occultation-band", "total-occultation-band"} {
			pathData := planetOccultationSVGPathData(t, diagram, className)
			if !planetOccultationSVGPathContains(pathData, x, y) {
				t.Fatalf("%s %s does not contain %.1f°E %.1f°N at %.3f, %.3f", projection.name, className,
					projection.longitude, projection.latitude, x, y)
			}
		}
	}
}

func TestPlanetOccultationPathSVGJupiter20230517UsesFineFootprintSweep(t *testing.T) {
	start := time.Date(2023, time.May, 17, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationJupiter,
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one Jupiter path", len(paths), err)
	}
	path := paths[0]
	for _, series := range []struct {
		name       string
		footprints []moon.PlanetOccultationFootprint
	}{
		{name: "partial", footprints: path.PartialFootprints},
		{name: "total", footprints: path.TotalFootprints},
	} {
		if len(series.footprints) < 100 {
			t.Fatalf("%s footprint count = %d, want fine sweep sampling", series.name, len(series.footprints))
		}
		for index := 1; index < len(series.footprints); index++ {
			gap := series.footprints[index].Time.Sub(series.footprints[index-1].Time)
			if gap > time.Minute+time.Millisecond {
				t.Fatalf("%s footprint gap at %d = %s, want at most one minute", series.name, index, gap)
			}
		}
	}
}

func planetOccultationSVGMapCenter(t *testing.T, diagram string) (float64, float64) {
	t.Helper()
	attributes := planetOccultationSVGElementAttributes(t, diagram, "circle", "map-ocean")
	return planetOccultationSVGFloatAttribute(t, attributes, "cx"),
		planetOccultationSVGFloatAttribute(t, attributes, "cy")
}

func planetOccultationSVGMapPoint(
	t *testing.T, diagram string, projection MapProjection, longitude, latitude float64,
) (float64, float64) {
	t.Helper()
	if projection == MapProjectionNorthPolar || projection == MapProjectionSouthPolar {
		return planetOccultationSVGMapCenter(t, diagram)
	}
	attributes := planetOccultationSVGElementAttributes(t, diagram, "rect", "map-ocean")
	x := planetOccultationSVGFloatAttribute(t, attributes, "x")
	y := planetOccultationSVGFloatAttribute(t, attributes, "y")
	width := planetOccultationSVGFloatAttribute(t, attributes, "width")
	height := planetOccultationSVGFloatAttribute(t, attributes, "height")
	return x + (longitude+180)/360*width, y + (90-latitude)/180*height
}

func planetOccultationSVGPathData(t testing.TB, diagram, className string) string {
	t.Helper()
	attributes := planetOccultationSVGElementAttributes(t, diagram, "path", className)
	value, ok := attributes["d"]
	if !ok {
		t.Fatalf("SVG path %q has no d attribute", className)
	}
	return value
}

func planetOccultationSVGElementAttributes(
	t testing.TB,
	diagram, elementName, className string,
) map[string]string {
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
		if !ok || start.Name.Local != elementName {
			continue
		}
		attributes := make(map[string]string, len(start.Attr))
		for _, attribute := range start.Attr {
			attributes[attribute.Name.Local] = attribute.Value
		}
		if attributes["class"] == className {
			return attributes
		}
	}
	t.Fatalf("SVG element %s.%s not found", elementName, className)
	return nil
}

func planetOccultationSVGFloatAttribute(t testing.TB, attributes map[string]string, name string) float64 {
	t.Helper()
	value, ok := attributes[name]
	if !ok {
		t.Fatalf("SVG element has no %s attribute", name)
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		t.Fatalf("parse SVG %s attribute %q: %v", name, value, err)
	}
	return parsed
}

func planetOccultationSVGPathContains(pathData string, x, y float64) bool {
	fields := strings.Fields(pathData)
	polygon := make([][2]float64, 0)
	for index := 0; index < len(fields); {
		switch fields[index] {
		case "M", "L":
			if index+2 >= len(fields) {
				return false
			}
			pointX, xErr := strconv.ParseFloat(fields[index+1], 64)
			pointY, yErr := strconv.ParseFloat(fields[index+2], 64)
			if xErr != nil || yErr != nil {
				return false
			}
			polygon = append(polygon, [2]float64{pointX, pointY})
			index += 3
		case "Z":
			if planetOccultationSVGPolygonContains(polygon, x, y) {
				return true
			}
			polygon = polygon[:0]
			index++
		default:
			return false
		}
	}
	return false
}

func planetOccultationSVGPolygonContains(polygon [][2]float64, x, y float64) bool {
	inside := false
	for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
		a, b := polygon[previous], polygon[current]
		if (a[1] > y) == (b[1] > y) {
			continue
		}
		if a[0]+(y-a[1])*(b[0]-a[0])/(b[1]-a[1]) > x {
			inside = !inside
		}
	}
	return inside
}

func TestPlanetOccultationPathSVGWrapsFivePhaseSummaryAtRenderedFontSize(t *testing.T) {
	path := samplePlanetOccultationPath()
	options := normalizeStarOccultationSVGOptions(PlanetOccultationSVGOptions{
		Width:       1100,
		SummaryText: strings.Repeat("A", 130),
	})
	options = planetOccultationSVGDefaults(path, options)
	lines := starOccultationSVGHeaderLines(planetOccultationStarShape(path, path.TargetID), options, starOccultationSVGTextFlags(options))
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

func TestPlanetOccultationPathSVGRejectsMalformedRiseSetCurve(t *testing.T) {
	path := samplePlanetOccultationPath()
	curve := sampleOccultationRiseSetCurve(path.Start.Time)
	curve.Direction = moon.RiseSetDirection("bogus")
	path.RiseSetCurves = []moon.OccultationRiseSetCurve{curve}
	_, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{})
	if !errors.Is(err, ErrInvalidPlanetOccultationPath) {
		t.Fatalf("PlanetOccultationPathSVG() error = %v, want ErrInvalidPlanetOccultationPath", err)
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

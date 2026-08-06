package svg

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/svgmap"
)

func TestSolarEclipseMapSVGTotalIncludesPartialAndCentralRegions(t *testing.T) {
	diagram, ok := SolarEclipseMapSVG(
		time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{Width: 900, Height: 620, Location: time.UTC, PartialStep: 10 * time.Minute},
	)
	if !ok {
		t.Fatal("expected total solar-eclipse map")
	}
	for _, want := range []string{
		"日全食全球见食图", "偏食始", "偏食终", "偏食可见区", "全食带", "中心线",
		"全球见食范围与中心食带", "全球阶段", "中心食始", "中心食终", "食带宽",
		"太阳沙罗 139", "食甚点太阳高度", "中心食持续", "P2", "P3", "U1", "U4",
		`class="partial-eclipse-region"`, `class="central-eclipse-band"`, `class="solar-center-line"`,
		`class="northern-central-limit"`, `class="southern-central-limit"`, `class="solar-greatest-marker"`,
		`class="solar-global-events"`, `class="solar-time-marker"`, `>17:00</text>`,
		`class="solar-greatest-terminator"`, `class="solar-penumbral-outline"`,
		`class="solar-central-shadow-outline"`, `class="solar-shadow-contact solar-contact-p1"`,
		`class="solar-axis-contact"`, `class="solar-subsolar-marker"`,
		`class="land"`, "不含行政边界",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("total solar-eclipse map missing %q", want)
		}
	}
	if err := validateEclipseMapXML(diagram); err != nil {
		t.Fatalf("total solar-eclipse map is not valid XML: %v", err)
	}
}

func TestSolarEclipseMapSVGPartialOnlyUsesPolarProjection(t *testing.T) {
	diagram, ok := SolarEclipseMapSVG(
		time.Date(2025, 3, 29, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{PartialStep: 10 * time.Minute},
	)
	if !ok {
		t.Fatal("expected partial solar-eclipse map")
	}
	for _, want := range []string{"日偏食全球见食图", `class="partial-eclipse-region"`, `<circle class="map-ocean"`, "北极方位等距投影"} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("partial solar-eclipse map missing %q", want)
		}
	}
	if strings.Contains(diagram, `class="central-eclipse-band"`) || strings.Contains(diagram, `class="solar-center-line"`) {
		t.Fatal("partial-only map contains a central path")
	}
	if !strings.Contains(diagram, `class="solar-global-events"`) || !strings.Contains(diagram, "全球阶段") {
		t.Fatal("partial-only map is missing global phase information")
	}
	if !strings.Contains(diagram, "全球见食范围") || strings.Contains(diagram, "全球见食范围与中心食带") {
		t.Fatal("partial-only map claims to contain a central path")
	}
	if strings.Contains(diagram, `class="solar-time-marker"`) {
		t.Fatal("partial-only map contains center-line time markers")
	}
	if err := validateEclipseMapXML(diagram); err != nil {
		t.Fatalf("partial solar-eclipse map is not valid XML: %v", err)
	}
}

func TestSolarEclipseMapSVGCanDisableTimeLabels(t *testing.T) {
	diagram, ok := SolarEclipseMapSVG(
		time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{Location: time.UTC, TimeLabelStep: -1},
	)
	if !ok {
		t.Fatal("expected total solar-eclipse map")
	}
	if strings.Contains(diagram, `class="solar-time-marker"`) {
		t.Fatal("disabled solar time labels were rendered")
	}
}

func TestSolarEclipseMapSVGCanDisableSampledShadowOutlines(t *testing.T) {
	diagram, ok := SolarEclipseMapSVG(
		time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{
			Location:             time.UTC,
			PenumbralOutlineStep: -1,
			CentralShadowStep:    -1,
		},
	)
	if !ok {
		t.Fatal("expected total solar-eclipse map")
	}
	if strings.Contains(diagram, `class="solar-penumbral-outline"`) ||
		strings.Contains(diagram, `class="solar-central-shadow-outline"`) {
		t.Fatal("disabled sampled shadow outlines were rendered")
	}
	if strings.Contains(diagram, "半影时刻线") || strings.Contains(diagram, "本影轮廓") {
		t.Fatal("disabled sampled shadow outlines remain in the legend")
	}
	if !strings.Contains(diagram, `class="solar-greatest-terminator"`) ||
		!strings.Contains(diagram, `class="solar-shadow-contact solar-contact-u1"`) {
		t.Fatal("disabling sampled outlines removed required contact geometry")
	}
}

func TestSolarEclipseMapSVGExplainsSampledShadowLines(t *testing.T) {
	cst := time.FixedZone("UTC+8", 8*60*60)
	date := time.Date(2035, 9, 2, 12, 0, 0, 0, cst)
	options := SolarEclipseMapSVGOptions{
		Width: 1200, Height: 800, Location: cst,
		Projection: EclipseMapProjectionEquirectangular,
	}
	normalized := normalizeSolarEclipseMapSVGOptions(date, options)
	if normalized.PenumbralOutlineStep != time.Hour {
		t.Fatalf("default penumbral outline step = %s, want 1h", normalized.PenumbralOutlineStep)
	}
	explicit := normalizeSolarEclipseMapSVGOptions(date, SolarEclipseMapSVGOptions{PenumbralOutlineStep: 30 * time.Minute})
	if explicit.PenumbralOutlineStep != 30*time.Minute {
		t.Fatalf("explicit penumbral outline step = %s, want 30m", explicit.PenumbralOutlineStep)
	}
	diagram, ok := SolarEclipseMapSVG(date, options)
	if !ok {
		t.Fatal("expected 2035 total solar-eclipse map")
	}
	for _, want := range []string{
		"半影时刻线（60 分钟）", "食甚晨昏圈", "本影轮廓（10 分钟）", "P/U 影锥接触",
		`class="solar-penumbral-time-label"`, `class="solar-map-legend"`,
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("solar-eclipse map does not explain %q", want)
		}
	}
}

func TestSolarEclipseMapSVGAntarcticEventUsesSouthPolarProjection(t *testing.T) {
	diagram, ok := SolarEclipseMapSVG(
		time.Date(2021, 12, 4, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{PartialStep: 10 * time.Minute},
	)
	if !ok {
		t.Fatal("expected Antarctic total solar-eclipse map")
	}
	for _, want := range []string{`<circle class="map-ocean"`, `<circle class="map-frame"`, "南极方位等距投影", `class="central-eclipse-band"`} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("Antarctic solar-eclipse map missing %q", want)
		}
	}
	if err := validateEclipseMapXML(diagram); err != nil {
		t.Fatalf("Antarctic solar-eclipse map is not valid XML: %v", err)
	}
}

func TestSolarEclipseMapSVGAnnularLabelsCentralBand(t *testing.T) {
	diagram, ok := SolarEclipseMapSVG(
		time.Date(2023, 10, 14, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{PartialStep: 15 * time.Minute},
	)
	if !ok {
		t.Fatal("expected annular solar-eclipse map")
	}
	if !strings.Contains(diagram, "日环食全球见食图") || !strings.Contains(diagram, "环食带") {
		t.Fatal("annular map does not distinguish the annular path")
	}
	if !strings.Contains(diagram, "反本影轮廓（10 分钟）") {
		t.Fatal("annular map does not explain the antumbral outlines")
	}
}

func TestSolarEclipseMapSVG2012DoesNotFillAntimeridianSpikes(t *testing.T) {
	cst := time.FixedZone("UTC+8", 8*60*60)
	date := time.Date(2012, 5, 21, 12, 0, 0, 0, cst)
	options := SolarEclipseMapSVGOptions{
		Width:       1200,
		Height:      800,
		Location:    cst,
		Projection:  EclipseMapProjectionEquirectangular,
		PartialStep: 2 * time.Minute,
	}
	diagram, ok := SolarEclipseMapSVG(date, options)
	if !ok {
		t.Fatal("expected 2012 annular solar-eclipse map")
	}
	frame := solarEclipseMapLayoutFor(
		normalizeSolarEclipseMapSVGOptions(date, options),
		svgmap.ProjectionEquirectangular,
	).frame

	for _, point := range []svgmap.GeoPoint{
		{Longitude: -170, Latitude: -50},
		{Longitude: 170, Latitude: -50},
	} {
		if solarPartialRegionContainsGeoPoint(t, diagram, frame, point) {
			t.Fatalf("known invisible point %#v is inside the rendered partial-eclipse region", point)
		}
	}
	visible := svgmap.GeoPoint{Longitude: 0, Latitude: 85}
	if !solarPartialRegionContainsGeoPoint(t, diagram, frame, visible) {
		t.Fatalf("known visible Arctic point %#v is outside the rendered partial-eclipse region", visible)
	}
}

func TestSolarEclipseMapSVG2012SupportsNorthPolarProjection(t *testing.T) {
	diagram, ok := SolarEclipseMapSVG(
		time.Date(2012, 5, 21, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{
			Projection:  EclipseMapProjectionNorthPolar,
			PartialStep: 10 * time.Minute,
		},
	)
	if !ok {
		t.Fatal("expected 2012 annular solar-eclipse north-polar map")
	}
	for _, want := range []string{
		`<circle class="map-ocean"`, `<circle class="map-frame"`,
		"北极方位等距投影", `class="central-eclipse-band"`,
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("2012 north-polar solar-eclipse map missing %q", want)
		}
	}
	if err := validateEclipseMapXML(diagram); err != nil {
		t.Fatalf("2012 north-polar solar-eclipse map is not valid XML: %v", err)
	}
}

func TestSolarEclipseMapSVGRejectsNoEventAndInvalidProjection(t *testing.T) {
	if _, ok := SolarEclipseMapSVG(time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC), SolarEclipseMapSVGOptions{}); ok {
		t.Fatal("unexpected solar-eclipse map for a no-event date")
	}
	if _, ok := SolarEclipseMapSVG(time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC), SolarEclipseMapSVGOptions{Projection: "invalid"}); ok {
		t.Fatal("invalid projection was accepted")
	}
}

func solarPartialRegionContainsGeoPoint(
	t *testing.T,
	diagram string,
	frame svgmap.Frame,
	point svgmap.GeoPoint,
) bool {
	t.Helper()
	const prefix = `<path class="partial-eclipse-region" d="`
	start := strings.Index(diagram, prefix)
	if start < 0 {
		t.Fatal("partial-eclipse SVG path is missing")
	}
	value := diagram[start+len(prefix):]
	end := strings.IndexByte(value, '"')
	if end < 0 {
		t.Fatal("partial-eclipse SVG path is malformed")
	}
	polygons := solarSVGPathPolygons(t, value[:end])
	x, y, ok := frame.Project(point.Longitude, point.Latitude)
	if !ok {
		return false
	}
	for _, polygon := range polygons {
		if solarSVGPointInPolygon(x, y, polygon) {
			return true
		}
	}
	return false
}

func solarSVGPathPolygons(t *testing.T, path string) [][][2]float64 {
	t.Helper()
	fields := strings.Fields(path)
	polygons := make([][][2]float64, 0)
	var current [][2]float64
	for index := 0; index < len(fields); {
		switch fields[index] {
		case "M":
			if len(current) > 0 {
				polygons = append(polygons, current)
			}
			current = nil
			index++
		case "L":
			index++
		case "Z":
			if len(current) > 0 {
				polygons = append(polygons, current)
				current = nil
			}
			index++
		default:
			if index+1 >= len(fields) {
				t.Fatalf("incomplete SVG coordinate at token %d", index)
			}
			x, err := strconv.ParseFloat(fields[index], 64)
			if err != nil {
				t.Fatalf("invalid SVG x coordinate %q: %v", fields[index], err)
			}
			y, err := strconv.ParseFloat(fields[index+1], 64)
			if err != nil {
				t.Fatalf("invalid SVG y coordinate %q: %v", fields[index+1], err)
			}
			current = append(current, [2]float64{x, y})
			index += 2
		}
	}
	if len(current) > 0 {
		polygons = append(polygons, current)
	}
	return polygons
}

func solarSVGPointInPolygon(x, y float64, polygon [][2]float64) bool {
	inside := false
	for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
		a, b := polygon[current], polygon[previous]
		if (a[1] > y) != (b[1] > y) && x < (b[0]-a[0])*(y-a[1])/(b[1]-a[1])+a[0] {
			inside = !inside
		}
	}
	return inside
}

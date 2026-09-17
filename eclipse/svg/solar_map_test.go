package svg

import (
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
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
		"全球见食范围与中心食带", "中心食始", "中心食终", "中心食带宽",
		"沙罗序列 139", "食甚点太阳高度", "中心食带宽", "图中时刻为", "中心食持续", "P2", "P3", "U1", "U4",
		`class="partial-eclipse-region"`, `class="central-eclipse-band"`, `class="solar-center-line"`,
		`class="northern-central-limit"`, `class="southern-central-limit"`, `class="solar-greatest-marker"`,
		`class="solar-time-marker"`, `>17:00</text>`,
		`class="solar-rise-set-boundary solar-start-rise"`, `stroke="#d97706"`,
		`class="solar-shadow-contact solar-contact-p1"`, `class="solar-magnitude-contour"`,
		`class="solar-axis-contact"`, `class="solar-subsolar-marker"`, `font-size="7"`, `opacity="0.72">太阳直射点</text>`,
		`class="land"`, "不含行政边界",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("total solar-eclipse map missing %q", want)
		}
	}
	// 瞬时半影/本影轮廓默认不画，否则会把地球盖住。
	for _, unwanted := range []string{`class="solar-penumbral-outline"`, `class="solar-central-shadow-outline"`} {
		if strings.Contains(diagram, unwanted) {
			t.Fatalf("default map must not draw %s", unwanted)
		}
	}
	if err := validateEclipseMapXML(diagram); err != nil {
		t.Fatalf("total solar-eclipse map is not valid XML: %v", err)
	}
}

func TestSolarEclipsePartialBandPolygonsContainSampledFootprints(t *testing.T) {
	for _, test := range []struct {
		name       string
		date       time.Time
		step       time.Duration
		boundaries int
	}{
		{name: "2009 antimeridian", date: time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC), step: 2 * time.Minute, boundaries: 96},
		{name: "2010 polar fold", date: time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC), step: 10 * time.Minute, boundaries: 24},
		{name: "2014 noncentral", date: time.Date(2014, time.April, 29, 0, 0, 0, 0, time.UTC), step: 10 * time.Minute, boundaries: 24},
	} {
		t.Run(test.name, func(t *testing.T) {
			partial, ok := eclipsecore.SolarEclipsePartialFootprints(test.date, eclipsecore.SolarEclipsePartialFootprintOptions{
				Step: test.step, BoundaryPoints: test.boundaries,
			})
			if !ok {
				t.Fatal("expected solar eclipse")
			}
			polygons, ok := solarEclipsePartialBandPolygons(partial)
			if !ok || len(polygons) == 0 {
				t.Fatal("expected authoritative partial-band polygons")
			}
			paths := make([][]geodata.GeoPoint, 0, len(partial.Footprints)*2)
			for _, footprint := range partial.Footprints {
				for _, boundary := range footprint.Boundaries {
					path := make([]geodata.GeoPoint, len(boundary))
					for index, point := range boundary {
						path[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
					}
					paths = append(paths, path)
				}
			}
			if miss := geodata.SphericalPolygonsPathMissDistanceKM(polygons, paths, false); miss > 5 {
				t.Fatalf("sampled penumbral boundary protrudes %.2f km outside the authoritative band", miss)
			}
		})
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
	if !strings.Contains(diagram, `class="solar-detailed-panel"`) || !strings.Contains(diagram, "食甚点的地方情况") {
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
	if !strings.Contains(diagram, `class="solar-rise-set-boundary solar-greatest-rise"`) ||
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
	if normalized.PenumbralOutlineStep != 0 {
		t.Fatalf("default penumbral outline step = %s, want it disabled", normalized.PenumbralOutlineStep)
	}
	explicit := normalizeSolarEclipseMapSVGOptions(date, SolarEclipseMapSVGOptions{PenumbralOutlineStep: 30 * time.Minute})
	if explicit.PenumbralOutlineStep != 30*time.Minute {
		t.Fatalf("explicit penumbral outline step = %s, want 30m", explicit.PenumbralOutlineStep)
	}
	// 显式请求时才画瞬时轮廓，图例也要给出采样间隔。
	options.PenumbralOutlineStep = time.Hour
	options.CentralShadowStep = 10 * time.Minute
	diagram, ok := SolarEclipseMapSVG(date, options)
	if !ok {
		t.Fatal("expected 2035 total solar-eclipse map")
	}
	for _, want := range []string{
		"半影时刻线（60 分钟）", "初亏/食甚/复圆日升日落线", "本影轮廓（10 分钟）", "P/U 影锥接触",
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
		SolarEclipseMapSVGOptions{PartialStep: 15 * time.Minute, CentralShadowStep: 10 * time.Minute},
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

func TestSolarEclipseMapSVG20100115TrimsExternalContactsFromTwoLimitSides(t *testing.T) {
	path, ok := eclipsecore.SolarEclipseCentralPath(
		time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC),
		eclipsecore.SolarEclipsePathOptions{Step: 2 * time.Minute, TargetSpacingKM: 700},
	)
	if !ok {
		t.Fatal("expected 2010 annular central path")
	}
	northern, southern, ok := solarEclipseTwoLimitPresentationLimits(path)
	if !ok {
		t.Fatal("expected physical-endpoint two-limit presentation geometry")
	}
	if len(northern) >= len(path.NorthernLimit) || len(southern) >= len(path.SouthernLimit) {
		t.Fatal("external-contact samples were not trimmed from the rendered limits")
	}
	start := path.CenterLine[0].Time
	end := path.CenterLine[len(path.CenterLine)-1].Time
	for name, points := range map[string][]eclipsecore.SolarEclipsePathPoint{
		"north": northern,
		"south": southern,
	} {
		if !points[0].Time.After(start) || !points[len(points)-1].Time.Before(end) {
			t.Fatalf("%s rendered limit extends outside the axis-contact interval", name)
		}
	}
	partial, ok := eclipsecore.SolarEclipsePartialFootprints(
		time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC),
		eclipsecore.SolarEclipsePartialFootprintOptions{Step: 10 * time.Minute, BoundaryPoints: 180},
	)
	if !ok {
		t.Fatal("expected 2010 annular partial footprint model")
	}
	polygons, ok := solarEclipseTwoLimitBandPolygons(
		path, northern, southern, partial.CentralBandFootprints,
	)
	if !ok || len(polygons) != 1 || len(polygons[0]) <= len(northern)+len(southern) {
		t.Fatalf("physical endpoint sweep polygons=%d ok=%v, want one enriched central-band outline", len(polygons), ok)
	}
	minimumEndTurn := 180.0
	for index := 1; index+1 < len(polygons[0]); index++ {
		point := polygons[0][index]
		if point.Longitude < 120 || point.Longitude > 123 || point.Latitude < 36 || point.Latitude > 39 {
			continue
		}
		previous, next := polygons[0][index-1], polygons[0][index+1]
		incoming := math.Atan2(point.Latitude-previous.Latitude, point.Longitude-previous.Longitude)
		outgoing := math.Atan2(next.Latitude-point.Latitude, next.Longitude-point.Longitude)
		minimumEndTurn = math.Min(minimumEndTurn, math.Remainder((outgoing-incoming)*180/math.Pi, 360))
	}
	if minimumEndTurn < -30 {
		t.Fatalf("2010 eastern SVG central-band cap turns inward by %.1f degrees", minimumEndTurn)
	}
	diagram, ok := SolarEclipseMapSVG(
		time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{PartialStep: 10 * time.Minute},
	)
	if !ok {
		t.Fatal("expected 2010 annular SVG")
	}
	if strings.Contains(diagram, `<g class="solar-central-shadow-sweep"`) {
		t.Fatal("central eclipse renders a second endpoint sweep over the canonical central band")
	}
}

func TestSolarEclipseMapSVGNonCentralAnnularUsesContinuousShadowSweep(t *testing.T) {
	diagram, ok := SolarEclipseMapSVG(
		time.Date(2014, 4, 29, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{
			Width: 1200, Height: 800, Location: time.UTC,
			Projection:  EclipseMapProjectionEquirectangular,
			PartialStep: 10 * time.Minute, CentralShadowStep: 10 * time.Minute,
		},
	)
	if !ok {
		t.Fatal("expected 2014 non-central annular solar-eclipse map")
	}
	start := strings.Index(diagram, `<g class="solar-central-shadow-sweep"`)
	if start < 0 {
		t.Fatal("non-central annular map is missing the continuous central-shadow sweep")
	}
	end := strings.Index(diagram[start:], `</g>`)
	if end < 0 {
		t.Fatal("central-shadow sweep group is not closed")
	}
	group := diagram[start : start+end]
	if got := strings.Count(group, `<path d="`); got != 1 {
		t.Fatalf("non-central annular sweep rendered %d disconnected paths, want one continuous path", got)
	}
	if strings.Contains(diagram, `class="central-eclipse-band"`) || strings.Contains(diagram, `class="solar-center-line"`) {
		t.Fatal("non-central annular map contains a central path")
	}
	if strings.Contains(diagram, `class="solar-central-shadow-outline"`) {
		t.Fatal("non-central annular map renders diagnostic open shadow arcs as interior boundary lines")
	}
	for _, want := range []string{"环食带", "P/U 影锥接触"} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("non-central annular legend is missing %q", want)
		}
	}
	if err := validateEclipseMapXML(diagram); err != nil {
		t.Fatalf("non-central annular map is not valid XML: %v", err)
	}
}

func TestSolarEclipseMapSVG20430409NonCentralTotalUsesCriticalEnvelope(t *testing.T) {
	diagram, ok := SolarEclipseMapSVG(
		time.Date(2043, 4, 9, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{
			Width: 1200, Height: 800, Location: time.UTC,
			Projection:  EclipseMapProjectionEquirectangular,
			PartialStep: 2 * time.Minute, CentralShadowStep: 2 * time.Minute,
		},
	)
	if !ok {
		t.Fatal("expected 2043 non-central total solar-eclipse map")
	}
	start := strings.Index(diagram, `<g class="solar-central-shadow-sweep" data-source="besselian-critical-envelope"`)
	if start < 0 {
		t.Fatal("2043 map is missing the critical-envelope central band")
	}
	end := strings.Index(diagram[start:], `</g>`)
	if end < 0 {
		t.Fatal("2043 critical-envelope group is not closed")
	}
	if got := strings.Count(diagram[start:start+end], `<path d="`); got != 1 {
		t.Fatalf("2043 critical envelope rendered %d disconnected paths, want one", got)
	}
	if strings.Contains(diagram, `class="solar-center-line"`) ||
		strings.Contains(diagram, `class="solar-central-shadow-outline"`) {
		t.Fatal("2043 non-central map contains a center line or diagnostic interior shadow arcs")
	}
	if err := validateEclipseMapXML(diagram); err != nil {
		t.Fatalf("2043 non-central total map is not valid XML: %v", err)
	}
}

func TestSolarEclipseMapSVG19500318UsesOpenSweepFallback(t *testing.T) {
	diagram, ok := SolarEclipseMapSVG(
		time.Date(1950, time.March, 18, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{
			Width: 1200, Height: 800, Location: time.UTC,
			Projection:  EclipseMapProjectionEquirectangular,
			PartialStep: 5 * time.Minute,
		},
	)
	if !ok {
		t.Fatal("expected 1950 non-central annular solar-eclipse map")
	}
	if !strings.Contains(diagram, `<g class="solar-central-shadow-sweep"`) {
		t.Fatal("1950 map is missing the validated open central-shadow sweep")
	}
	if strings.Contains(diagram, `class="central-eclipse-band"`) || strings.Contains(diagram, `class="solar-center-line"`) {
		t.Fatal("1950 non-central map contains a central path")
	}
	if err := validateEclipseMapXML(diagram); err != nil {
		t.Fatalf("1950 non-central map is not valid XML: %v", err)
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
	// 等经纬图现在按食甚经度居中，测试必须用与实际渲染相同的中心。
	centre, ok := eclipsecore.SolarEclipseOnDateNASABulletinSplitK(date)
	if !ok {
		t.Fatal("expected 2012 annular solar-eclipse info")
	}
	frame := solarEclipseMapLayoutFor(
		normalizeSolarEclipseMapSVGOptions(date, options),
		svgmap.ProjectionEquirectangular,
		svgmap.GeoPoint{Longitude: centre.GreatestLongitude, Latitude: centre.GreatestLatitude},
	).frame

	for _, point := range []svgmap.GeoPoint{
		{Longitude: -170, Latitude: -50},
		{Longitude: 170, Latitude: -50},
	} {
		if solarPartialRegionContainsGeoPoint(t, diagram, frame, point) {
			t.Fatalf("known invisible point %#v is inside the rendered partial-eclipse region", point)
		}
	}
	// 正例取确实落在偏食区内的点：0°E/85°N 的上边界只到约 70.7°N，它在界外。
	for _, visible := range []svgmap.GeoPoint{
		{Longitude: 140, Latitude: 60},
		{Longitude: centre.GreatestLongitude, Latitude: centre.GreatestLatitude},
	} {
		if !solarPartialRegionContainsGeoPoint(t, diagram, frame, visible) {
			t.Fatalf("known visible point %#v is outside the rendered partial-eclipse region", visible)
		}
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
	const prefix = `<path class="partial-eclipse-region"`
	start := strings.Index(diagram, prefix)
	if start < 0 {
		t.Fatal("partial-eclipse SVG path is missing")
	}
	value := diagram[start+len(prefix):]
	const marker = ` d="`
	dataStart := strings.Index(value, marker)
	if dataStart < 0 {
		t.Fatal("partial-eclipse SVG path carries no geometry")
	}
	value = value[dataStart+len(marker):]
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

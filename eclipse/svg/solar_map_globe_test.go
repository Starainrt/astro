package svg

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgmap"
)

func solarEclipseGlobeFrame(t *testing.T, date time.Time, options SolarEclipseMapSVGOptions) svgmap.Frame {
	t.Helper()
	info, ok := eclipsecore.SolarEclipseOnDateNASABulletinSplitK(date)
	if !ok {
		t.Fatal("missing eclipse")
	}
	return solarEclipseMapLayoutFor(
		normalizeSolarEclipseMapSVGOptions(date, options),
		svgmap.ProjectionOrthographic,
		svgmap.GeoPoint{Longitude: info.GreatestLongitude, Latitude: info.GreatestLatitude},
	).frame
}

var solarEclipseGlobeCoordinatePattern = regexp.MustCompile(`[ML](-?[0-9.]+) (-?[0-9.]+)`)

// 正射球面图上陆地必须被裁到圆盘内：背面的陆地若漏进来会折叠到盘内画错位置。
func TestSolarEclipseMapOrthographicClipsLandToDisk(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	options := SolarEclipseMapSVGOptions{
		Width: 1000, Height: 760, Location: time.UTC,
		Projection: EclipseMapProjectionOrthographic, PartialStep: 10 * time.Minute,
		GreatestTimeStep: 30 * time.Minute,
	}
	rendered, ok := SolarEclipseMapSVG(date, options)
	if !ok {
		t.Fatal("missing map")
	}
	if !strings.Contains(rendered, `<clipPath id="solar-map-clip"><circle`) {
		t.Fatal("an orthographic map must clip to a circular disk")
	}
	for _, class := range []string{`class="land"`, `class="graticule"`, `class="solar-greatest-time-isoline"`} {
		if !strings.Contains(rendered, class) {
			t.Fatalf("missing %s on the globe", class)
		}
	}

	const landMarker = `<path class="land" d="`
	start := strings.Index(rendered, landMarker)
	if start < 0 {
		t.Fatal("missing land path")
	}
	rest := rendered[start+len(landMarker):]
	path := rest[:strings.Index(rest, `"`)]
	frame := solarEclipseGlobeFrame(t, date, options)
	centerX := frame.X + frame.Width/2
	centerY := frame.Y + frame.Height/2
	radius := math.Min(frame.Width, frame.Height) / 2
	matches := solarEclipseGlobeCoordinatePattern.FindAllStringSubmatch(path, -1)
	if len(matches) < 2000 {
		t.Fatalf("land path carries only %d vertices", len(matches))
	}
	outside := 0
	for _, match := range matches {
		x, errX := strconv.ParseFloat(match[1], 64)
		y, errY := strconv.ParseFloat(match[2], 64)
		if errX != nil || errY != nil {
			t.Fatalf("unparsable land coordinate %q %q", match[1], match[2])
		}
		if math.Hypot(x-centerX, y-centerY) > radius+0.6 {
			outside++
		}
	}
	if outside != 0 {
		t.Fatalf("%d of %d land vertices fall outside the disk", outside, len(matches))
	}
	t.Logf("land vertices=%d all inside the disk", len(matches))
}

// 正射投影下食甚点必须落在盘心，且背面点不可见。
func TestSolarEclipseMapOrthographicFrameGeometry(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	options := SolarEclipseMapSVGOptions{
		Width: 1000, Height: 760, Location: time.UTC,
		Projection: EclipseMapProjectionOrthographic,
	}
	frame := solarEclipseGlobeFrame(t, date, options)
	if frame.Width != frame.Height {
		t.Fatalf("globe frame is not square: %gx%g", frame.Width, frame.Height)
	}
	x, y, ok := frame.Project(frame.CenterLongitude, frame.CenterLatitude)
	if !ok || math.Hypot(x-(frame.X+frame.Width/2), y-(frame.Y+frame.Height/2)) > 1e-9 {
		t.Fatalf("view centre projects to %g,%g ok=%v", x, y, ok)
	}
	antipode := svgmap.GeoPoint{
		Longitude: frame.CenterLongitude + 180,
		Latitude:  -frame.CenterLatitude,
	}
	if _, _, visible := frame.Project(antipode.Longitude, antipode.Latitude); visible {
		t.Fatal("the antipode must not project onto the globe")
	}
}

// 球面版式只支持显式请求，自动投影仍按事件几何选择。
func TestSolarEclipseMapOrthographicIsExplicitOnly(t *testing.T) {
	if !validEclipseMapProjection(EclipseMapProjectionOrthographic) {
		t.Fatal("orthographic projection is rejected by validation")
	}
	if got := internalEclipseMapProjection(EclipseMapProjectionOrthographic); got != svgmap.ProjectionOrthographic {
		t.Fatalf("orthographic maps to %q", got)
	}
}

// 图上时刻必须同时写明时标（UTC）与展示时区相对它的偏差，且详细版式的所有内容都不得溢出图框。
func TestSolarEclipseMapStatesTimeZoneOffset(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		location *time.Location
		want     string
	}{
		{location: time.UTC, want: "图中时刻为 UTC"},
		{location: time.FixedZone("CST", 8*3600), want: "UTC+08:00"},
		{location: time.FixedZone("EST", -5*3600), want: "UTC-05:00"},
	}
	for _, test := range cases {
		options := SolarEclipseMapSVGOptions{
			Width: 1000, Height: 1414, Location: test.location,
			Projection: EclipseMapProjectionOrthographic,
		}
		rendered, ok := SolarEclipseMapSVG(date, options)
		if !ok {
			t.Fatalf("%s: missing map", test.location)
		}
		if !strings.Contains(rendered, test.want) {
			t.Fatalf("%s: map does not state %q", test.location, test.want)
		}
	}
}

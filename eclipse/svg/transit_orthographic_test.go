package svg

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestTransitMapVisibilityCenter(tester *testing.T) {
	for _, test := range []struct {
		name      string
		columns   []basic.PlanetTransitVisibilityColumn
		longitude float64
		latitude  float64
		valid     bool
	}{
		{name: "empty"},
		{name: "no-visible-interval", columns: []basic.PlanetTransitVisibilityColumn{{Longitude: -180}, {Longitude: 180}}},
		{name: "antimeridian", columns: []basic.PlanetTransitVisibilityColumn{
			{Longitude: -180, Intervals: [][2]float64{{-10, 10}}},
			{Longitude: -178, Intervals: [][2]float64{{-10, 10}}},
			{Longitude: -176}, {Longitude: 176},
			{Longitude: 178, Intervals: [][2]float64{{-10, 10}}},
			{Longitude: 180, Intervals: [][2]float64{{-10, 10}}},
		}, longitude: 180, valid: true},
		{name: "north", columns: []basic.PlanetTransitVisibilityColumn{
			{Longitude: -180, Intervals: [][2]float64{{70, 90}}},
			{Longitude: -90, Intervals: [][2]float64{{70, 90}}},
			{Longitude: 0, Intervals: [][2]float64{{70, 90}}},
			{Longitude: 90, Intervals: [][2]float64{{70, 90}}},
			{Longitude: 180, Intervals: [][2]float64{{70, 90}}},
		}, latitude: 90, valid: true},
		{name: "south", columns: []basic.PlanetTransitVisibilityColumn{
			{Longitude: -180, Intervals: [][2]float64{{-90, -70}}},
			{Longitude: -90, Intervals: [][2]float64{{-90, -70}}},
			{Longitude: 0, Intervals: [][2]float64{{-90, -70}}},
			{Longitude: 90, Intervals: [][2]float64{{-90, -70}}},
			{Longitude: 180, Intervals: [][2]float64{{-90, -70}}},
		}, latitude: -90, valid: true},
	} {
		tester.Run(test.name, func(tester *testing.T) {
			center, valid := transitMapVisibilityCenter(test.columns)
			if valid != test.valid {
				tester.Fatalf("valid=%v, want %v", valid, test.valid)
			}
			if valid && (math.Abs(center.Latitude-test.latitude) > 1e-8 ||
				math.Abs(test.latitude) < 90 && math.Abs(math.Remainder(center.Longitude-test.longitude, 360)) > 1e-8) {
				tester.Fatalf("center=%+v, want longitude=%v latitude=%v", center, test.longitude, test.latitude)
			}
		})
	}
}

func TestTransitMapOrthographicPartialCenter(tester *testing.T) {
	for _, test := range []struct {
		planet    TransitPlanet
		year      int
		month     time.Month
		day       int
		longitude float64
		latitude  float64
	}{
		{TransitVenus, 2854, time.December, 14, -135, -60},
		{TransitMercury, 1937, time.May, 11, 0, -60},
	} {
		tester.Run(fmt.Sprintf("%d-%d", test.planet, test.year), func(tester *testing.T) {
			date := time.Date(test.year, test.month, test.day, 0, 0, 0, 0, time.UTC)
			transit, valid := closestPlanetTransit(test.planet, date)
			if !valid || transit.HasExternal {
				tester.Fatal("expected a transit without geocentric contacts")
			}
			center := transitMapOrthographicCenter(transit)
			local := basic.PlanetTransitLocal(transit, basic.Observer{Longitude: center.Longitude, Latitude: center.Latitude})
			if !local.Valid || !local.Visible {
				tester.Fatalf("view centre must lie in the visible footprint: %+v", center)
			}
			for _, projection := range []EclipseMapProjection{EclipseMapProjectionEquirectangular, EclipseMapProjectionSouthPolar, EclipseMapProjectionOrthographic} {
				chart, valid := TransitMapSVG(test.planet, date, TransitMapSVGOptions{Projection: projection, Language: "en"})
				if !valid || strings.Contains(chart, "NaN") || strings.Contains(chart, "Inf") {
					tester.Fatalf("%s: invalid chart", projection)
				}
				paths := regexp.MustCompile(`<path class="transit-partial-region" d="([^"]+)"[^>]+>`).FindAllStringSubmatch(chart, -1)
				if len(paths) != 1 || !strings.Contains(paths[0][0], `fill-rule="nonzero"`) {
					tester.Fatalf("%s: footprint must use a single nonzero fill, got %d paths", projection, len(paths))
				}
				frame := eclipseMapFrame(960, 640, internalEclipseMapProjection(projection), 142, 122)
				if projection == EclipseMapProjectionOrthographic {
					frame.CenterLongitude, frame.CenterLatitude = center.Longitude, center.Latitude
					if !strings.Contains(chart, fmt.Sprintf(`data-center-longitude="%.6f" data-center-latitude="%.6f"`, center.Longitude, center.Latitude)) ||
						!strings.Contains(chart, "centred on the topocentric visible footprint") {
						tester.Fatal("orthographic view centre or explanation missing")
					}
				}
				x, y, projectable := frame.Project(test.longitude, test.latitude)
				if !projectable || !lunarPenumbralTestInside(lunarPenumbralTestRings(tester, paths[0][1]), x, y) {
					tester.Fatalf("%s: known visible site is outside the painted footprint", projection)
				}
			}
		})
	}
}

func TestTransitMapOrthographicCenterFallback(tester *testing.T) {
	for _, test := range []struct {
		planet TransitPlanet
		date   time.Time
	}{
		{TransitMercury, time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC)},
		{TransitVenus, time.Date(2012, 6, 6, 0, 0, 0, 0, time.UTC)},
	} {
		transit, valid := closestPlanetTransit(test.planet, test.date)
		if !valid || !transit.HasExternal {
			tester.Fatal("expected an ordinary geocentric transit")
		}
		want := solarEclipseSubsolarPoint(transitMapContactsOf(transit).greatest)
		if got := transitMapOrthographicCenter(transit); got != want {
			tester.Fatalf("ordinary transit view changed: got %+v want %+v", got, want)
		}
		transit.Valid, transit.HasExternal = false, false
		if got := transitMapOrthographicCenter(transit); got != want {
			tester.Fatalf("empty visibility must fall back to the subsolar point: %+v", got)
		}
	}
}

func TestTransitMapFooterAlignment(tester *testing.T) {
	footer := regexp.MustCompile(`<text x="([0-9.]+)" y="[0-9.]+" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="11" text-anchor="(\w+)">`)
	date := time.Date(2854, 12, 14, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		projection EclipseMapProjection
		x, anchor  string
	}{
		{EclipseMapProjectionOrthographic, "480.000", "middle"},
		{"", "104.000", "start"},
	} {
		chart, ok := TransitMapSVG(TransitVenus, date, TransitMapSVGOptions{Projection: test.projection, Language: "en"})
		if !ok {
			tester.Fatalf("%q: render failed", test.projection)
		}
		lines := footer.FindAllStringSubmatch(chart, -1)
		if len(lines) < 2 {
			tester.Fatalf("%q: footer lines=%d", test.projection, len(lines))
		}
		for _, line := range lines {
			if line[1] != test.x || line[2] != test.anchor {
				tester.Fatalf("%q: footer at x=%s anchor=%s, want %s %s", test.projection, line[1], line[2], test.x, test.anchor)
			}
		}
		if !strings.Contains(chart, `class="transit-partial-region"`) || strings.Contains(chart, "#747b7d") || !strings.Contains(chart, `fill="#f7dd93"`) {
			tester.Fatalf("%q: partial region fill changed", test.projection)
		}
	}
}

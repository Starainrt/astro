package svg

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/svgmap"
)

func TestTransitMapSVGRendersRegionsAndIsochrones(t *testing.T) {
	cases := []struct {
		name   string
		planet TransitPlanet
		date   time.Time
	}{
		{"venus-2012", TransitVenus, time.Date(2012, 6, 6, 0, 0, 0, 0, time.UTC)},
		{"mercury-2019", TransitMercury, time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		for _, projection := range []EclipseMapProjection{"", EclipseMapProjectionOrthographic, EclipseMapProjectionNorthPolar} {
			chart, ok := TransitMapSVG(tc.planet, tc.date, TransitMapSVGOptions{Location: time.UTC, Projection: projection, Isochrones: TransitIsochroneAll})
			if !ok {
				t.Fatalf("%s %q: render failed", tc.name, projection)
			}
			for _, want := range []string{"transit-entire-region", "transit-sunrise-region", "transit-sunset-region",
				"c1-horizon", "c4-horizon", `class="c1-isochrone"`, `class="c2-isochrone"`, `class="c3-isochrone"`,
				`class="c4-isochrone"`, "c1-isochrone-label"} {
				if !strings.Contains(chart, want) {
					t.Fatalf("%s %q: missing %s", tc.name, projection, want)
				}
			}
			if strings.Contains(chart, "NaN") || strings.Contains(chart, "Inf") {
				t.Fatalf("%s %q: non-finite coordinate", tc.name, projection)
			}
		}
	}
}

func TestTransitMapSVGOptions(t *testing.T) {
	date := time.Date(2012, 6, 6, 0, 0, 0, 0, time.UTC)
	if _, ok := TransitMapSVG(3, date, TransitMapSVGOptions{}); ok {
		t.Fatal("unsupported planet must fail")
	}
	if _, ok := TransitMapSVG(TransitVenus, date, TransitMapSVGOptions{
		TimeScale: astro.TimeScaleUT1, Location: time.FixedZone("CST", 8*3600)}); ok {
		t.Fatal("UT1 with non-UTC location must fail")
	}
	chart, ok := TransitMapSVG(TransitVenus, date, TransitMapSVGOptions{Location: time.UTC, Language: "en"})
	if !ok || strings.Contains(chart, "isochrone") || !strings.Contains(chart, "Transit of Venus") {
		t.Fatal("isochrones must be off by default and English labels kept")
	}
	chart, ok = TransitMapSVG(TransitVenus, date, TransitMapSVGOptions{Location: time.UTC, Isochrones: TransitIsochroneC2, IsochroneStep: 5 * time.Minute})
	if !ok || !strings.Contains(chart, `class="c2-isochrone"`) || strings.Contains(chart, `class="c1-isochrone"`) ||
		!strings.Contains(chart, "二触等时线") || strings.Contains(chart, "一触等时线") {
		t.Fatal("only the selected contact must be drawn and listed in the legend")
	}
}

// 太阳在视地平上的区域由图框减去对日点夜冠得到，边界距直射点 90°+地平折射。
// 曾直接画 90° 半球环，拆分日界线时分不清内外而整片填反。
func TestTransitMapSunUpBoundaryIncludesRefraction(t *testing.T) {
	transit := basic.ClosestVenusTransit(basic.Date2JD(time.Date(2012, 6, 6, 0, 0, 0, 0, time.UTC)))
	contacts := transitMapContactsOf(transit)
	radius := 90 + basic.RefractionFromApparentAltitude(0, transitMapRefractionPressureHPa, transitMapRefractionTemperatureC)
	frame := eclipseMapFrame(960, 640, svgmap.ProjectionEquirectangular, 142, 92)
	for _, at := range []time.Time{contacts.start, contacts.greatest, contacts.end} {
		path, _ := transitMapSunUpPath(at, frame)
		if !strings.HasPrefix(path, fmt.Sprintf("M %.3f %.3f H", frame.X, frame.Y)) {
			t.Fatalf("%v: sun-up path must start with the frame rectangle", at)
		}
		subsolar := solarEclipseSubsolarPoint(at)
		_, boundary := transitMapSunUpPath(at, frame)
		for _, point := range boundary {
			distance := transitMapTestAngularDistance(point, subsolar)
			if math.Abs(distance-radius) > 1e-6 {
				t.Fatalf("%v: boundary point %.4f deg from subsolar point, want %.4f", at, distance, radius)
			}
		}
	}
}

func transitMapTestAngularDistance(a, b svgmap.GeoPoint) float64 {
	r := math.Pi / 180
	cosine := math.Sin(a.Latitude*r)*math.Sin(b.Latitude*r) + math.Cos(a.Latitude*r)*math.Cos(b.Latitude*r)*math.Cos((a.Longitude-b.Longitude)*r)
	return math.Acos(math.Max(-1, math.Min(1, cosine))) / r
}

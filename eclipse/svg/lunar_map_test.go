package svg

import (
	"encoding/xml"
	"errors"
	"io"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/svgmap"
)

func TestLunarEclipseMapSVGVisibilityRegions(t *testing.T) {
	diagram, ok := LunarEclipseMapSVG(
		time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		LunarEclipseMapSVGOptions{Width: 900, Height: 620, Location: time.UTC},
	)
	if !ok {
		t.Fatal("expected lunar-eclipse visibility map")
	}
	for _, want := range []string{
		"月全食全球可见图", "P1", "P4", "食甚", "全程可见", "带食月出", "带食月落", "不可见",
		`class="entire-eclipse-region"`, "moonrise-region", "moonset-region",
		`class="p1-horizon"`, `class="p4-horizon"`, `class="land"`, "不含行政边界",
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("lunar-eclipse map missing %q", want)
		}
	}
	if err := validateEclipseMapXML(diagram); err != nil {
		t.Fatalf("lunar-eclipse map is not valid XML: %v", err)
	}
}

func TestLunarEclipseMapSVGUsesExclusiveVisibilityLayers(t *testing.T) {
	cst := time.FixedZone("CST", 8*60*60)
	diagram, ok := LunarEclipseMapSVG(
		time.Date(2029, 1, 1, 0, 0, 0, 0, cst),
		LunarEclipseMapSVGOptions{Width: 1200, Height: 800, Location: cst},
	)
	if !ok {
		t.Fatal("expected lunar-eclipse visibility map")
	}

	for _, want := range []string{
		`mask id="lunar-not-visible-mask"`,
		`mask id="lunar-not-p4-mask"`,
		`mask id="lunar-not-p1-mask"`,
		`class="eclipse-unavailable-region" mask="url(#lunar-not-visible-mask)"`,
		`class="visible-at-p1 moonset-region" mask="url(#lunar-not-p4-mask)"`,
		`class="visible-at-p4 moonrise-region" mask="url(#lunar-not-p1-mask)"`,
	} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("lunar-eclipse map does not render exclusive visibility regions: missing %q", want)
		}
	}
	if strings.Contains(diagram, `class="entire-eclipse-region"`) && strings.Contains(diagram, `fill-opacity="0.70"`) {
		t.Fatal("entire-eclipse region still uses the opaque stacked-overlay style")
	}
	if !strings.Contains(diagram, `clipPath id="lunar-visible-maximum"`) {
		t.Fatal("entire-eclipse region is not constrained by greatest-eclipse visibility")
	}
}

func TestLunarEclipseMapSVGSupportsForcedPolarProjection(t *testing.T) {
	diagram, ok := LunarEclipseMapSVG(
		time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		LunarEclipseMapSVGOptions{Projection: EclipseMapProjectionSouthPolar},
	)
	if !ok {
		t.Fatal("expected forced south-polar lunar-eclipse map")
	}
	for _, want := range []string{`<circle class="map-ocean"`, `<circle class="map-frame"`, "南极方位等距投影"} {
		if !strings.Contains(diagram, want) {
			t.Fatalf("south-polar lunar-eclipse map missing %q", want)
		}
	}
	if err := validateEclipseMapXML(diagram); err != nil {
		t.Fatalf("south-polar lunar-eclipse map is not valid XML: %v", err)
	}
}

func TestLunarEclipseVisibilityPathMatchesTopocentricHorizon(t *testing.T) {
	for _, at := range []time.Time{
		time.Date(2026, 3, 3, 9, 0, 0, 0, time.UTC),
		time.Date(2025, 9, 7, 16, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 15, 12, 0, 0, 0, time.UTC),
	} {
		for _, projection := range []svgmap.Projection{svgmap.ProjectionEquirectangular, svgmap.ProjectionNorthPolar, svgmap.ProjectionSouthPolar} {
			frame := svgmap.Frame{Width: 36000, Height: 18000, Projection: projection}
			path, boundary := lunarEclipseVisibilityPath(at, frame)
			if strings.Contains(path, "NaN") || strings.Contains(path, "Inf") {
				t.Fatal("nonfinite path")
			}
			rings := svgClosedPathRings(t, `<path d="`+path+`"/>`)
			if len(rings) == 0 || len(rings) > 3 {
				t.Fatalf("projection=%s subpaths=%d", projection, len(rings))
			}
			jd := basic.Date2JDE(at)
			for _, point := range boundary {
				if altitude := basic.HMoonHeight(jd, point.Longitude, point.Latitude, 0); math.Abs(altitude) > 1e-9 {
					t.Fatalf("horizon altitude=%g", altitude)
				}
			}
			for lon := -175.; lon < 180; lon += 20 {
				for lat := -85.; lat < 90; lat += 10 {
					x, y, ok := frame.Project(lon, lat)
					if !ok {
						continue
					}
					altitude := basic.HMoonHeight(jd, lon, lat, 0)
					if math.Abs(altitude) < 0.05 {
						continue
					}
					inside := false
					for _, ring := range rings {
						inside = inside || pointInLunarVisibilityPolygon(x, y, ring)
					}
					if inside != (altitude > 0) {
						t.Fatalf("projection=%s at=%s site=(%v,%v) inside=%v altitude=%v", projection, at, lon, lat, inside, altitude)
					}
				}
			}
		}
	}
}

func TestLunarEclipseMapSVGRejectsNoEventAndInvalidProjection(t *testing.T) {
	if _, ok := LunarEclipseMapSVG(time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC), LunarEclipseMapSVGOptions{}); ok {
		t.Fatal("unexpected lunar-eclipse map for a no-event date")
	}
	if _, ok := LunarEclipseMapSVG(time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC), LunarEclipseMapSVGOptions{Projection: "invalid"}); ok {
		t.Fatal("invalid projection was accepted")
	}
}

func validateEclipseMapXML(value string) error {
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

func pointInLunarVisibilityPolygon(x, y float64, polygon [][2]float64) bool {
	inside := false
	for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
		a, b := polygon[current], polygon[previous]
		crosses := (a[1] > y) != (b[1] > y)
		if crosses && x < (b[0]-a[0])*(y-a[1])/(b[1]-a[1])+a[0] {
			inside = !inside
		}
	}
	return inside
}

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

func TestLunarEclipseVisibilityPolygonsContainOnlyVisibleHemisphere(t *testing.T) {
	tests := []struct {
		name       string
		projection svgmap.Projection
		center     svgmap.GeoPoint
		visible    svgmap.GeoPoint
		hidden     svgmap.GeoPoint
	}{
		{
			name:       "equirectangular across antimeridian",
			projection: svgmap.ProjectionEquirectangular,
			center:     svgmap.GeoPoint{Longitude: 170, Latitude: 12},
			visible:    svgmap.GeoPoint{Longitude: 170, Latitude: 12},
			hidden:     svgmap.GeoPoint{Longitude: -10, Latitude: -12},
		},
		{
			name:       "north polar center inside projection",
			projection: svgmap.ProjectionNorthPolar,
			center:     svgmap.GeoPoint{Longitude: 30, Latitude: 20},
			visible:    svgmap.GeoPoint{Longitude: 30, Latitude: 80},
			hidden:     svgmap.GeoPoint{Longitude: -150, Latitude: 10},
		},
		{
			name:       "north polar center outside projection",
			projection: svgmap.ProjectionNorthPolar,
			center:     svgmap.GeoPoint{Longitude: 30, Latitude: -20},
			visible:    svgmap.GeoPoint{Longitude: 30, Latitude: 10},
			hidden:     svgmap.GeoPoint{Longitude: 30, Latitude: 90},
		},
		{
			name:       "south polar center inside projection",
			projection: svgmap.ProjectionSouthPolar,
			center:     svgmap.GeoPoint{Longitude: -45, Latitude: -20},
			visible:    svgmap.GeoPoint{Longitude: -45, Latitude: -80},
			hidden:     svgmap.GeoPoint{Longitude: 135, Latitude: -10},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			frame := svgmap.Frame{Width: 360, Height: 360, Projection: test.projection}
			polygons := lunarEclipseVisibilityPolygons(test.center, test.projection, 360)
			if len(polygons) == 0 {
				t.Fatal("visibility polygon is empty")
			}
			if !projectedPointInLunarVisibility(frame, polygons, test.visible) {
				t.Fatalf("visible point %#v is outside the rendered region", test.visible)
			}
			if projectedPointInLunarVisibility(frame, polygons, test.hidden) {
				t.Fatalf("hidden point %#v is inside the rendered region", test.hidden)
			}
		})
	}
}

func TestLunarEclipseVisibilityPathDoesNotUseTriangleFan(t *testing.T) {
	for _, projection := range []svgmap.Projection{
		svgmap.ProjectionEquirectangular,
		svgmap.ProjectionNorthPolar,
		svgmap.ProjectionSouthPolar,
	} {
		frame := svgmap.Frame{Width: 720, Height: 360, Projection: projection}
		if projection != svgmap.ProjectionEquirectangular {
			frame.Width = 360
		}
		path := lunarEclipseVisibilityPathForCenter(
			svgmap.GeoPoint{Longitude: 170, Latitude: 12}, frame,
		)
		if subpaths := strings.Count(path, "M "); subpaths != 1 {
			t.Fatalf("%s visibility path has %d subpaths, want one continuous outline", projection, subpaths)
		}
		if strings.Contains(path, "NaN") || strings.Contains(path, "Inf") {
			t.Fatalf("%s visibility path contains a non-finite coordinate", projection)
		}
	}
}

func TestLunarEclipseVisibilityPolygonsMatchSphericalHorizon(t *testing.T) {
	tests := []struct {
		projection svgmap.Projection
		center     svgmap.GeoPoint
	}{
		{svgmap.ProjectionEquirectangular, svgmap.GeoPoint{Longitude: 170, Latitude: 18}},
		{svgmap.ProjectionEquirectangular, svgmap.GeoPoint{Longitude: -170, Latitude: -18}},
		{svgmap.ProjectionEquirectangular, svgmap.GeoPoint{Longitude: 170, Latitude: 0}},
		{svgmap.ProjectionNorthPolar, svgmap.GeoPoint{Longitude: 35, Latitude: 18}},
		{svgmap.ProjectionNorthPolar, svgmap.GeoPoint{Longitude: 35, Latitude: -18}},
		{svgmap.ProjectionSouthPolar, svgmap.GeoPoint{Longitude: -70, Latitude: -18}},
		{svgmap.ProjectionSouthPolar, svgmap.GeoPoint{Longitude: -70, Latitude: 18}},
	}
	for _, test := range tests {
		frame := svgmap.Frame{Width: 720, Height: 360, Projection: test.projection}
		if test.projection != svgmap.ProjectionEquirectangular {
			frame.Width = 360
		}
		polygons := lunarEclipseVisibilityPolygons(test.center, test.projection, 360)
		for latitude := -75.0; latitude <= 75; latitude += 15 {
			if test.projection == svgmap.ProjectionNorthPolar && latitude <= 0 {
				continue
			}
			if test.projection == svgmap.ProjectionSouthPolar && latitude >= 0 {
				continue
			}
			for longitude := -165.0; longitude <= 165; longitude += 30 {
				point := svgmap.GeoPoint{Longitude: longitude, Latitude: latitude}
				dot := lunarVisibilityDot(test.center, point)
				if math.Abs(dot) < 0.02 {
					continue
				}
				got := projectedPointInLunarVisibility(frame, polygons, point)
				if got != (dot > 0) {
					t.Fatalf("%s center=%#v point=%#v inside=%v dot=%.6f",
						test.projection, test.center, point, got, dot)
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

func projectedPointInLunarVisibility(frame svgmap.Frame, polygons [][]svgmap.GeoPoint, point svgmap.GeoPoint) bool {
	x, y, ok := frame.Project(point.Longitude, point.Latitude)
	if !ok {
		return false
	}
	for _, polygon := range polygons {
		projected := make([][2]float64, 0, len(polygon))
		for _, vertex := range polygon {
			px, py, projectedOK := frame.Project(vertex.Longitude, vertex.Latitude)
			if projectedOK {
				projected = append(projected, [2]float64{px, py})
			}
		}
		if pointInLunarVisibilityPolygon(x, y, projected) {
			return true
		}
	}
	return false
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

func lunarVisibilityDot(center, point svgmap.GeoPoint) float64 {
	centerLongitude := center.Longitude * math.Pi / 180
	centerLatitude := center.Latitude * math.Pi / 180
	longitude := point.Longitude * math.Pi / 180
	latitude := point.Latitude * math.Pi / 180
	return math.Sin(centerLatitude)*math.Sin(latitude) +
		math.Cos(centerLatitude)*math.Cos(latitude)*math.Cos(longitude-centerLongitude)
}

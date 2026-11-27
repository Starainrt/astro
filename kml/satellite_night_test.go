package kml

import (
	"encoding/json"
	"encoding/xml"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/satellite"
)

func TestSatelliteKMLNightPolarOctantConsistency(t *testing.T) {
	for _, pole := range []float64{-90, 90} {
		ring := [][]float64{{-180, pole}, {0, pole}, {0, 0}, {-180, 0}, {-180, pole}}
		got, err := satelliteNightGeometry([]element{{kind: "Polygon", rings: [][][]float64{ring}}})
		if err != nil || got.MultiGeometry == nil || len(got.MultiGeometry.Polygons) != 2 {
			t.Fatalf("pole=%g: %#v, %v", pole, got, err)
		}
		for _, polygon := range got.MultiGeometry.Polygons {
			positions := strings.Fields(polygon.OuterBoundaryIs.Ring.Coordinates)
			if len(positions) != 5 || positions[0] != positions[4] {
				t.Fatalf("pole=%g: expected a closed octant, got %v", pole, positions)
			}
			poleVertices := 0
			for _, position := range positions[:4] {
				if strings.HasSuffix(position, ","+strconv.FormatFloat(pole, 'f', 0, 64)) {
					poleVertices++
				}
			}
			if poleVertices != 2 {
				t.Fatalf("pole=%g: lost an octant meridian: %v", pole, positions)
			}
		}
	}
}

func TestSatelliteKMLNightOnSphere(t *testing.T) {
	raw, err := os.ReadFile("../satellite/testdata/skyfield_coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Groups []struct{ Elements satellite.Elements }
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		hour           int
		threshold      float64
		solarLongitude float64
	}{
		{9, -6, 41.1}, {14, 30, -33.889208}, {0, 30, 176}, {0, 10, 176}, {14, 85, -33.889208},
		{14, 90, -33.889208}, {14, -90, -33.889208},
		{14, math.Nextafter(-90, 0), -33.889208}, {14, math.Nextafter(90, 0), -33.889208},
		{14, -89.99999999, -33.889208}, {14, 89.99999999, -33.889208},
		{14, -89.999999, -33.889208}, {14, 89.999999, -33.889208},
	} {
		at := time.Date(2026, 11, 14, test.hour, 0, 0, 0, time.UTC)
		coverage, err := satellite.CoverageAt(fixture.Groups[4].Elements, at, satellite.CoverageOptions{TwilightSunAltitude: &test.threshold})
		if err != nil {
			t.Fatal(err)
		}
		root := coverageKML(t, coverage)
		items := placemarksOfRole(root, "night-region")
		switch test.threshold {
		case -90, math.Nextafter(-90, 0), -89.99999999:
			if len(items) != 0 || coverage.NightIsGlobal || len(coverage.Night) != 0 {
				t.Fatalf("threshold %g: zero-area night must not create a KML night layer", test.threshold)
			}
		case 90, math.Nextafter(90, 0), 89.99999999:
			if len(items) == 0 || !coverage.NightIsGlobal || len(coverage.Night) != 0 {
				t.Fatalf("threshold %g: missing global night", test.threshold)
			}
		case -89.999999:
			if len(items) != 1 || items[0].MultiGeometry == nil || len(items[0].MultiGeometry.Polygons) != 1 {
				t.Fatal("tiny night must be one polygon, not a world-sized complement")
			}
		}
		var probes []geodata.GeoPoint
		for _, latitude := range []float64{-90, -82, -55, -30, -18.32, 12, 68, 82, 90} {
			for _, longitude := range []float64{-179.9, -161, -101, -45, -33.89, -22, 18, 101, 161, 179.9} {
				probes = append(probes, geodata.GeoPoint{Longitude: longitude, Latitude: latitude})
			}
		}
		night := kmlCoverageMembership(t, root, "night-region", probes)
		for index, point := range probes {
			rad := math.Pi / 180
			alt := math.Asin(math.Sin(point.Latitude*rad)*math.Sin(-18.32*rad)+math.Cos(point.Latitude*rad)*math.Cos(-18.32*rad)*math.Cos((point.Longitude-test.solarLongitude)*rad)) / rad
			if math.Abs(alt-test.threshold) < 2 {
				continue
			}
			if night[index] != (alt < test.threshold) {
				t.Fatalf("%s threshold %g, %v: night=%t, solar altitude=%g", at, test.threshold, point, night[index], alt)
			}
		}
		if test.threshold == 85 || test.threshold == 89.999999 {
			var holes int
			for _, placemark := range placemarksOfRole(root, "night-region") {
				for _, shape := range placemark.MultiGeometry.Polygons {
					holes += len(shape.InnerBoundaries)
				}
			}
			if holes != 1 {
				t.Fatalf("night geometry has %d enclosed daylight holes, want 1", holes)
			}
		}
		if test.hour == 14 && test.threshold == 30 {
			visible := kmlCoverageMembership(t, root, "satellite-visible", []geodata.GeoPoint{{Longitude: -33.889208, Latitude: -18.320628}, {Longitude: -101, Latitude: -10}})
			if visible[0] || !visible[1] {
				t.Fatalf("KML visual coverage filled=%v, want [false true]", visible)
			}
		}
	}
}

func coverageKML(t *testing.T, coverage satellite.Coverage) kmlRoot {
	t.Helper()
	data, err := geojson.MarshalSatelliteCoverage(coverage)
	if err != nil {
		t.Fatal(err)
	}
	out, err := FromGeoJSON(data, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var root kmlRoot
	if err = xml.Unmarshal(out, &root); err != nil {
		t.Fatal(err)
	}
	return root
}

func kmlCoverageMembership(t *testing.T, root kmlRoot, role string, points []geodata.GeoPoint) []bool {
	t.Helper()
	positions := func(ring boundary) []geodata.GeoPoint {
		var vertices []geodata.GeoPoint
		for _, value := range strings.Fields(ring.Ring.Coordinates) {
			fields := strings.Split(value, ",")
			longitude, lonErr := strconv.ParseFloat(fields[0], 64)
			latitude, latErr := strconv.ParseFloat(fields[1], 64)
			if lonErr != nil || latErr != nil {
				t.Fatalf("invalid position %q", value)
			}
			vertices = append(vertices, geodata.GeoPoint{Longitude: longitude, Latitude: latitude})
		}
		return vertices
	}
	result := make([]bool, len(points))
	for _, placemark := range placemarksOfRole(root, role) {
		if placemark.MultiGeometry == nil {
			continue
		}
		for _, shape := range placemark.MultiGeometry.Polygons {
			outer := positions(*shape.OuterBoundaryIs)
			var distinct []geodata.GeoPoint
			for _, point := range outer {
				seen := false
				for _, vertex := range distinct {
					seen = seen || geodata.SameGeoPoint(vertex, point) || (math.Abs(point.Latitude) == 90 && vertex.Latitude == point.Latitude)
				}
				if !seen {
					distinct = append(distinct, point)
				}
			}
			if len(distinct) < 3 {
				t.Fatalf("%s has a degenerate spherical outer ring: %v", role, outer)
			}
			var holes [][]geodata.GeoPoint
			for _, hole := range shape.InnerBoundaries {
				holes = append(holes, positions(hole))
			}
			inside := geodata.SphericalPolygonsContainPoints([][]geodata.GeoPoint{outer}, points)
			excluded := geodata.SphericalPolygonsContainPoints(holes, points)
			for index := range result {
				result[index] = result[index] || (inside[index] && !excluded[index])
			}
		}
	}
	return result
}

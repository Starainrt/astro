package svg

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/satellite"
)

func TestCoverageSVGDaylightHole(t *testing.T) {
	data, err := os.ReadFile("../testdata/skyfield_coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Groups []struct{ Elements satellite.Elements }
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 11, 14, 14, 0, 0, 0, time.UTC)
	threshold := 30.0
	sweep, err := satellite.CoverageBetween(fixture.Groups[4].Elements, at, at.Add(10*time.Second), 10*time.Second, satellite.CoverageOptions{TwilightSunAltitude: &threshold})
	if err != nil {
		t.Fatal(err)
	}
	for _, projection := range []MapProjection{MapProjectionEquirectangular, MapProjectionOrthographic} {
		value, err := CoverageSVG(sweep, CoverageSVGOptions{Projection: projection, Highlight: at})
		if err != nil {
			t.Fatal(err)
		}
		paths := map[string]string{}
		var frame map[string]string
		for _, node := range parseDiagram(t, value) {
			paths[node.Attr["data-role"]] = node.Attr["d"]
			if node.Attr["data-role"] == "map" {
				frame = node.Attr
			}
		}
		var dayChecks, nightChecks int
		for row := 1; row < 20; row++ {
			for column := 1; column < 20; column++ {
				if projection == MapProjectionOrthographic && math.Hypot(float64(column)/20-.5, float64(row)/20-.5) > .48 {
					continue
				}
				x := number(frame["data-x"]) + float64(column)/20*number(frame["data-width"])
				y := number(frame["data-y"]) + float64(row)/20*number(frame["data-height"])
				if !coveragePathContains(paths["coverage-sweep"], x, y) {
					continue
				}
				lon, lat := inverseMap(frame, x, y)
				rad := math.Pi / 180
				alt := math.Asin(math.Sin(lat*rad)*math.Sin(-18.320628*rad)+math.Cos(lat*rad)*math.Cos(-18.320628*rad)*math.Cos((lon+33.889208)*rad)) / rad
				if math.Abs(alt-threshold) < 2 {
					continue
				}
				night := alt < threshold
				if night {
					nightChecks++
				} else {
					dayChecks++
				}
				for _, role := range []string{"satellite-visible", "visible-sweep"} {
					if got := coveragePathContains(paths[role], x, y); got != night {
						t.Fatalf("%s %s fills=%t at %g, %g (night=%t)", projection, role, got, lon, lat, night)
					}
				}
			}
		}
		if dayChecks == 0 || nightChecks == 0 {
			t.Fatalf("%s: need both daylight and night samples", projection)
		}
	}
}

func coveragePathContains(path string, x, y float64) bool {
	inside := false
	for _, ring := range strings.Split(path, "Z") {
		fields := strings.Fields(ring)
		if len(fields) < 6 {
			continue
		}
		previousX, previousY := number(strings.TrimLeft(fields[len(fields)-2], "ML")), number(fields[len(fields)-1])
		for index := 0; index < len(fields); index += 2 {
			currentX, currentY := number(strings.TrimLeft(fields[index], "ML")), number(fields[index+1])
			if (currentY > y) != (previousY > y) && x < currentX+(previousX-currentX)*(y-currentY)/(previousY-currentY) {
				inside = !inside
			}
			previousX, previousY = currentX, currentY
		}
	}
	return inside
}

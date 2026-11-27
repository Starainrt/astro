package satellite

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func coverageGEOFixture(t *testing.T) Elements {
	t.Helper()
	data, err := os.ReadFile("testdata/skyfield_coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Groups []struct {
			Label    string
			Elements Elements
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, group := range fixture.Groups {
		if group.Label == "geo-tle" {
			return group.Elements
		}
	}
	t.Fatal("missing geo-tle fixture")
	return Elements{}
}

func TestCoverageSweptDaylightHole(t *testing.T) {
	at := time.Date(2026, 11, 14, 14, 0, 0, 0, time.UTC)
	elements := coverageGEOFixture(t)
	day := GeoPoint{Longitude: -33.889208, Latitude: -18.320628}
	night := GeoPoint{Longitude: -101, Latitude: -10}
	threshold := 30.0
	for _, options := range []CoverageOptions{{TwilightSunAltitude: &threshold}, {}} {
		sweep, err := CoverageBetween(elements, at, at.Add(10*time.Second), 10*time.Second, options)
		if err != nil {
			t.Fatal(err)
		}
		inRegions := func(regions []Region, point GeoPoint) bool {
			for _, region := range regions {
				inside := coverageReferenceChartContains(region.Outer, point)
				for _, hole := range region.Holes {
					inside = inside && !coverageReferenceChartContains(hole, point)
				}
				if inside {
					return true
				}
			}
			return false
		}
		if !inRegions(sweep.Swept, day) || inRegions(sweep.VisibleSwept, day) || inRegions(sweep.VisibleSwept, night) != (options.TwilightSunAltitude != nil) {
			t.Fatalf("wrong daylight or twilight membership for %+v", options)
		}
		if options.TwilightSunAltitude != nil && (len(sweep.VisibleSwept) != 1 || len(sweep.VisibleSwept[0].Holes) != 1) {
			t.Fatal("swept region lost its daylight hole")
		}
		instant := sweep.Instants[0]
		if options.TwilightSunAltitude != nil && (len(instant.Visible) != 0 || inRegions(instant.VisibleRegions, day) || !inRegions(instant.VisibleRegions, night)) {
			t.Fatal("wrong instantaneous region or legacy projection")
		}
		look, err := LookAt(elements, Observer{Longitude: day.Longitude, Latitude: day.Latitude}, at, ObserveOptions{NoRefraction: true})
		if err != nil || !look.SunCenterVisible || look.Elevation < 69 || look.Elevation > 70 {
			t.Fatalf("unexpected geometry at the daylight site: %+v, %v", look, err)
		}
	}
}

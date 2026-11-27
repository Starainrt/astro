package geojson_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/satellite"
)

func TestSatelliteGeoJSONDaylightHole(t *testing.T) {
	data, err := os.ReadFile("../satellite/testdata/skyfield_coverage.json")
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
	check := func(role string, data []byte, err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		var got satelliteJSONCollection
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		for _, feature := range got.Features {
			if feature.Properties["role"] != role {
				continue
			}
			var rings [][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &rings); err != nil || len(rings) != 2 {
				t.Fatalf("%s: expected an outer ring and hole, got %d rings: %v", role, len(rings), err)
			}
			if !visualContains(rings[:1], -33.889208, -18.320628) || visualContains(rings, -33.889208, -18.320628) || !visualContains(rings, -101, -10) {
				t.Fatalf("%s: wrong day or night membership", role)
			}
			return
		}
		t.Fatalf("missing %s", role)
	}
	inst, err := geojson.MarshalSatelliteCoverage(sweep.Instants[0])
	check("satellite-visible", inst, err)
	interval, err := geojson.MarshalSatelliteCoverageSweep(sweep, geojson.SatellitePathOptions{})
	check("visible-sweep", interval, err)
}

func visualContains(rings [][][]float64, lon, lat float64) bool {
	inside := false
	for _, ring := range rings {
		for index, from := range ring[:len(ring)-1] {
			to := ring[index+1]
			if (from[1] > lat) != (to[1] > lat) && lon < from[0]+(to[0]-from[0])*(lat-from[1])/(to[1]-from[1]) {
				inside = !inside
			}
		}
	}
	return inside
}

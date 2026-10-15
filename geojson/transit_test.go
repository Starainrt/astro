package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/venus"
)

type transitCollection struct {
	TimeScale string `json:"time_scale"`
	Features  []struct {
		Properties map[string]interface{} `json:"properties"`
		Geometry   struct {
			Type        string          `json:"type"`
			Coordinates json.RawMessage `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

func decodeTransit(t *testing.T, data []byte) (transitCollection, map[string]int) {
	t.Helper()
	var collection transitCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatal(err)
	}
	roles := map[string]int{}
	for _, item := range collection.Features {
		if item.Properties["event"] != "planet-transit" {
			t.Fatalf("unexpected event %v", item.Properties["event"])
		}
		roles[item.Properties["role"].(string)]++
	}
	return collection, roles
}

// 2032-11-13 水星凌日：等时线四个接触俱全，区域与地平线成对输出，凌甚点落在太阳直射点。
func TestMarshalMercuryTransit(t *testing.T) {
	info := mercury.ClosestTransit(time.Date(2032, 11, 13, 0, 0, 0, 0, time.UTC))
	data, err := geojson.MarshalMercuryTransit(info, geojson.PlanetTransitOptions{
		IsochroneStep: 20 * time.Second,
		TimeMarkers:   &geojson.TimeMarkerOptions{Step: 30 * time.Minute},
	})
	if err != nil {
		t.Fatal(err)
	}
	collection, roles := decodeTransit(t, data)
	for _, role := range []string{
		"visible-at-c1", "c1-horizon", "visible-at-c4", "c4-horizon",
		"transit-entire", "transit-at-sunrise", "transit-at-sunset",
		"c1-isochrone", "c2-isochrone", "c3-isochrone", "c4-isochrone", "greatest",
	} {
		if roles[role] != 1 {
			t.Fatalf("role %s count = %d, want 1", role, roles[role])
		}
	}
	if roles["time-marker"] < 8 {
		t.Fatalf("time markers = %d, want one every 30 min over ~4.4 h", roles["time-marker"])
	}
	if collection.TimeScale != "" {
		t.Fatal("UTC export must not carry time_scale")
	}
	for _, item := range collection.Features {
		role := item.Properties["role"]
		switch role {
		case "c1-isochrone", "c2-isochrone", "c3-isochrone", "c4-isochrone":
			var lines [][][]float64
			if err := json.Unmarshal(item.Geometry.Coordinates, &lines); err != nil {
				t.Fatal(err)
			}
			times := item.Properties["times"].([]interface{})
			if len(times) != len(lines) || len(lines) == 0 {
				t.Fatalf("%s: times %d do not align with %d lines", role, len(times), len(lines))
			}
			for index, line := range lines {
				if len(times[index].([]interface{})) != len(line) {
					t.Fatalf("%s line %d: times length mismatch", role, index)
				}
			}
		case "greatest":
			if item.Properties["internal_start"] == nil || item.Properties["minimum_separation_arcsec"] == nil {
				t.Fatal("greatest point must carry contact metadata")
			}
		}
	}
}

// 默认不出等时线；SkipRoles 跳过全部分区后只剩地平线、可见区与凌甚点。
func TestMarshalVenusTransitDefaultsAndSkip(t *testing.T) {
	info := venus.ClosestTransit(time.Date(2012, 6, 6, 0, 0, 0, 0, time.UTC))
	data, err := geojson.MarshalVenusTransit(info, geojson.PlanetTransitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, roles := decodeTransit(t, data)
	if roles["c1-isochrone"] != 0 || roles["transit-entire"] != 1 {
		t.Fatalf("default roles = %v", roles)
	}
	data, err = geojson.MarshalVenusTransit(info, geojson.PlanetTransitOptions{
		SkipRoles: []string{"transit-entire", "transit-at-sunrise", "transit-at-sunset"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, roles = decodeTransit(t, data)
	if len(roles) != 5 {
		t.Fatalf("skipped region roles = %v", roles)
	}
}

func TestMarshalPlanetTransitValidation(t *testing.T) {
	info := mercury.ClosestTransit(time.Date(2019, 11, 11, 0, 0, 0, 0, time.UTC))
	if _, err := geojson.MarshalMercuryTransit(mercury.TransitInfo{}, geojson.PlanetTransitOptions{}); err == nil {
		t.Fatal("invalid transit must be rejected")
	}
	if _, err := geojson.MarshalMercuryTransit(info, geojson.PlanetTransitOptions{IsochroneStep: time.Millisecond}); err == nil {
		t.Fatal("sub-second isochrone step must be rejected")
	}
	swapped := info
	swapped.Start, swapped.End = info.End, info.Start
	if _, err := geojson.MarshalMercuryTransit(swapped, geojson.PlanetTransitOptions{}); err == nil {
		t.Fatal("unordered contacts must be rejected")
	}
	inconsistent := info
	inconsistent.HasInternal = false
	if _, err := geojson.MarshalMercuryTransit(inconsistent, geojson.PlanetTransitOptions{IsochroneStep: time.Minute}); err == nil {
		t.Fatal("internal contact times without HasInternal must be rejected")
	}
	if _, err := geojson.MarshalMercuryTransit(info, geojson.PlanetTransitOptions{
		TimeMarkers: &geojson.TimeMarkerOptions{TimeScale: astro.TimeScaleUT1, Location: time.FixedZone("CST", 8*3600)},
	}); err == nil {
		t.Fatal("UT1 with a non-UTC location must be rejected")
	}
	data, err := geojson.MarshalMercuryTransit(info, geojson.PlanetTransitOptions{
		TimeMarkers: &geojson.TimeMarkerOptions{TimeScale: astro.TimeScaleUT1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if collection, _ := decodeTransit(t, data); collection.TimeScale != "UT1" {
		t.Fatal("UT1 export must carry time_scale")
	}
}

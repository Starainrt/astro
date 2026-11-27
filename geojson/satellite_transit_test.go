package geojson_test

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/satellite"
)

func satelliteJSONPath() satellite.TransitPath {
	at := time.Date(2026, 11, 14, 0, 0, 0, 123456789, time.UTC)
	line := []satellite.TransitPathPoint{{Time: at, Longitude: 179, Latitude: 10}, {Time: at.Add(time.Second), Longitude: 180, Latitude: 11}, {Time: at.Add(2 * time.Second), Longitude: -179, Latitude: 12}}
	left := append([]satellite.TransitPathPoint(nil), line...)
	right := append([]satellite.TransitPathPoint(nil), line...)
	for i := range line {
		left[i].Latitude += .1
		right[i].Latitude -= .1
	}
	return satellite.TransitPath{Elements: satellite.Elements{Name: "ISS", CatalogNumber: 25544, Epoch: at}, Event: satellite.TransitEvent{Target: satellite.TransitSun, Start: line[0].Time, Greatest: line[1].Time, End: line[2].Time, StartPoint: line[0], GreatestPoint: line[1], EndPoint: line[2], Complete: true}, CenterLine: line, LeftLimit: left, RightLimit: right, StartHorizon: []satellite.GeoPoint{{Longitude: 179, Latitude: 10}, {Longitude: 180, Latitude: 20}, {Longitude: -179, Latitude: 10}}, EndHorizon: []satellite.GeoPoint{{Longitude: 179, Latitude: 12}, {Longitude: 180, Latitude: 22}, {Longitude: -179, Latitude: 12}}}
}

type satelliteJSONCollection struct {
	Type       string
	Properties map[string]interface{}
	Features   []struct {
		Properties map[string]interface{}
		Geometry   struct {
			Type        string
			Coordinates json.RawMessage
		}
	}
}

func TestSatelliteTransitJSONSynthetic(t *testing.T) {
	path := satelliteJSONPath()
	obs := satellite.Observer{Longitude: 178, Latitude: 11}
	nearby := satellite.NearbyTransit{Event: path.Event, Nearest: path.CenterLine[1], DistanceKM: -4, NearestDistanceKM: 12}
	data, err := geojson.MarshalSatelliteTransitPath(path, geojson.SatellitePathOptions{TimeMarkerStep: time.Second, Observer: &obs, Nearby: &nearby})
	if err != nil {
		t.Fatal(err)
	}
	var got satelliteJSONCollection
	if err = json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "FeatureCollection" || len(got.Properties) != 8 || got.Properties["event"] != "satellite-transit" || got.Properties["target"] != "sun" || got.Properties["norad"] != float64(25544) {
		t.Fatalf("properties: %+v", got)
	}
	roles := map[string]int{}
	for _, f := range got.Features {
		role := f.Properties["role"].(string)
		roles[role]++
		switch role {
		case "center-line":
			var lines [][][]float64
			if err = json.Unmarshal(f.Geometry.Coordinates, &lines); err != nil {
				t.Fatal(err)
			}
			times := f.Properties["times"].([]interface{})
			for i, line := range lines {
				for j, p := range line {
					if math.Abs(p[0]) > 180 || math.Abs(p[1]) > 90 {
						t.Fatal(p)
					}
					at, err := time.Parse(time.RFC3339Nano, times[i].([]interface{})[j].(string))
					if err != nil {
						t.Fatal(err)
					}
					fraction := at.Sub(path.Event.Start).Seconds()
					if math.Abs(p[1]-(10+fraction)) > 1e-9 {
						t.Fatal("time/coordinate", p, at)
					}
				}
			}
		case "transit-band":
			var polys [][][][]float64
			if err = json.Unmarshal(f.Geometry.Coordinates, &polys); err != nil {
				t.Fatal(err)
			}
			if len(polys) != 2 {
				t.Fatal("seam", len(polys))
			}
			for _, poly := range polys {
				r := poly[0]
				if r[0][0] != r[len(r)-1][0] || r[0][1] != r[len(r)-1][1] {
					t.Fatal("open")
				}
				for j := 1; j < len(r); j++ {
					if math.Abs(r[j][0]-r[j-1][0]) > 180 {
						t.Fatal("seam edge")
					}
				}
			}
		case "nearest-connector":
			if f.Properties["distance_km"] != float64(12) || f.Properties["side"] != "right" {
				t.Fatal(f.Properties)
			}
		}
	}
	want := map[string]int{"transit-band": 1, "center-line": 1, "left-limit": 1, "right-limit": 1, "target-horizon": 2, "start": 1, "greatest": 1, "end": 1, "time-marker": 1, "observer": 1, "nearest-point": 1, "nearest-connector": 1}
	if len(roles) != len(want) {
		t.Fatal(roles)
	}
	for k, v := range want {
		if roles[k] != v {
			t.Fatal(k, roles)
		}
	}
	again, err := geojson.NewWithInstance(nil).MarshalSatelliteTransitPath(path, geojson.SatellitePathOptions{TimeMarkerStep: time.Second, Observer: &obs, Nearby: &nearby})
	if err != nil || string(again) != string(data) {
		t.Fatal("module", err)
	}
}
func TestSatelliteTransitJSONInvalidSynthetic(t *testing.T) {
	for _, p := range []satellite.TransitPath{{}, func() satellite.TransitPath {
		p := satelliteJSONPath()
		p.CenterLine[1].Longitude = math.NaN()
		return p
	}(), func() satellite.TransitPath {
		p := satelliteJSONPath()
		p.CenterLine[1].Time = p.CenterLine[0].Time
		return p
	}()} {
		if _, err := geojson.MarshalSatelliteTransitPath(p, geojson.SatellitePathOptions{}); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Fatal(err)
		}
	}
}

func TestSatelliteTransitShadowJSONSynthetic(t *testing.T) {
	at := satelliteJSONPath().Event.Start
	shadow := satellite.TransitShadow{Time: at, Target: satellite.TransitMoon}
	data, err := geojson.MarshalSatelliteTransitShadow(shadow)
	if err != nil {
		t.Fatal(err)
	}
	var got satelliteJSONCollection
	if err = json.Unmarshal(data, &got); err != nil || len(got.Features) != 0 || got.Properties["target"] != "moon" {
		t.Fatal(string(data), err)
	}
	shadow.Valid = true
	shadow.Point = satellite.TransitPathPoint{Time: at, Longitude: 179, Latitude: 10}
	shadow.SubPoint = satellite.GroundPoint{Longitude: 178, Latitude: 9}
	shadow.Footprint = []satellite.GeoPoint{{Longitude: 179, Latitude: 10}, {Longitude: -179, Latitude: 10}, {Longitude: -179, Latitude: 11}, {Longitude: 179, Latitude: 11}, {Longitude: 179, Latitude: 10}}
	data, err = geojson.MarshalSatelliteTransitShadow(shadow)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &got); err != nil || len(got.Features) != 3 {
		t.Fatal(string(data), err)
	}
	want := []string{"current-transit-footprint", "current-transit-point", "subsatellite-point"}
	for i, f := range got.Features {
		if f.Properties["role"] != want[i] || f.Properties["time"] != at.UTC().Format(time.RFC3339Nano) {
			t.Fatal(f)
		}
	}
	data2, err := geojson.NewWithInstance(nil).MarshalSatelliteTransitShadow(shadow)
	if err != nil || string(data2) != string(data) {
		t.Fatal("module", err)
	}
}
func TestSatelliteTransitJSONFixtureConsistency(t *testing.T) {
	var fixture struct {
		Groups []struct {
			Elements   satellite.Elements
			Target     satellite.TransitTarget
			Minimum    float64
			Start, End time.Time
		}
	}
	b, err := os.ReadFile("../satellite/testdata/skyfield_transit_events.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, g := range fixture.Groups {
		events, err := satellite.TransitEvents(g.Elements, g.Start, g.End, satellite.TransitSearchOptions{Targets: []satellite.TransitTarget{g.Target}, MinTargetAltitude: g.Minimum})
		if err != nil || len(events) == 0 {
			t.Fatal(err)
		}
		path, err := satellite.TransitEventPath(g.Elements, events[0], 60*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		data, err := geojson.MarshalSatelliteTransitPath(path, geojson.SatellitePathOptions{})
		if err != nil {
			t.Fatal(g.Elements.Name, err)
		}
		var got satelliteJSONCollection
		if err = json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		for _, f := range got.Features {
			if f.Properties["role"] != "center-line" {
				continue
			}
			var lines [][][]float64
			json.Unmarshal(f.Geometry.Coordinates, &lines)
			for _, p := range path.CenterLine {
				found := false
				for _, l := range lines {
					for _, q := range l {
						if math.Abs(math.Remainder(p.Longitude-q[0], 360)) < 1e-10 && p.Latitude == q[1] {
							found = true
						}
					}
				}
				if !found {
					t.Fatalf("missing input coordinate %+v", p)
				}
			}
		}
	}
	t.Logf("%d fixture paths independently decoded; original centre coordinates retained", len(fixture.Groups))
}
func TestSatelliteTransitJSONOptionsSynthetic(t *testing.T) {
	p := satelliteJSONPath()
	obs := satellite.Observer{Longitude: 1, Latitude: 2}
	n := satellite.NearbyTransit{Event: p.Event, Nearest: p.CenterLine[1]}
	for _, opt := range []geojson.SatellitePathOptions{{TimeMarkerStep: -1}, {Nearby: &n}, {Observer: &satellite.Observer{Latitude: 91}}, {Observer: &obs, Nearby: &satellite.NearbyTransit{}}} {
		if _, err := geojson.MarshalSatelliteTransitPath(p, opt); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Fatal(opt, err)
		}
	}
	for _, distance := range []float64{-1, 0, 1} {
		n.DistanceKM = distance
		data, err := geojson.MarshalSatelliteTransitPath(p, geojson.SatellitePathOptions{Observer: &obs, Nearby: &n})
		if err != nil {
			t.Fatal(err)
		}
		var got satelliteJSONCollection
		json.Unmarshal(data, &got)
		for _, f := range got.Features {
			if f.Properties["role"] == "nearest-point" {
				want := "none"
				if distance < 0 {
					want = "right"
				}
				if distance > 0 {
					want = "left"
				}
				if !reflect.DeepEqual(f.Properties["side"], want) {
					t.Fatal(f.Properties)
				}
			}
		}
	}
}

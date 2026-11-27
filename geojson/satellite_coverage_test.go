package geojson_test

import (
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/satellite"
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"
	"time"
)

func TestSatelliteCoverageSynthetic(t *testing.T) {
	ring := []satellite.GeoPoint{{Longitude: 179, Latitude: 10}, {Longitude: -179, Latitude: 10}, {Longitude: -179, Latitude: 12}, {Longitude: 179, Latitude: 12}, {Longitude: 179, Latitude: 10}}
	c := satellite.Coverage{Time: time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC), Footprint: ring, Night: ring, Visible: [][]satellite.GeoPoint{ring}}
	data, err := geojson.MarshalSatelliteCoverage(c)
	if err != nil {
		t.Fatal(err)
	}
	var got satelliteJSONCollection
	if err = json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	roles := map[string]int{}
	for _, f := range got.Features {
		roles[f.Properties["role"].(string)]++
	}
	for _, r := range []string{"subsatellite-point", "satellite-footprint", "night-region", "satellite-visible"} {
		if roles[r] == 0 {
			t.Fatal(r)
		}
	}
}

func coverageJSONSweep() satellite.CoverageSweep {
	at := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	box := func(a, b, c, d float64) []satellite.GeoPoint {
		if a > 180 {
			a -= 360
		}
		if c > 180 {
			c -= 360
		}
		return []satellite.GeoPoint{{Longitude: a, Latitude: b}, {Longitude: c, Latitude: b}, {Longitude: c, Latitude: d}, {Longitude: a, Latitude: d}, {Longitude: a, Latitude: b}}
	}
	s := satellite.CoverageSweep{Start: at, End: at.Add(3 * time.Minute), Swept: []satellite.Region{{Outer: box(-20, -20, 20, 20), Holes: [][]satellite.GeoPoint{box(-5, -5, 5, 5)}}}, VisibleSwept: []satellite.Region{{Outer: box(5, 5, 15, 15)}}}
	for i := 0; i < 4; i++ {
		p := satellite.GroundPoint{Time: at.Add(time.Duration(i) * time.Minute), Longitude: 179 + float64(i), Latitude: 10 + float64(i), HeightKM: 400, SpeedKMS: 7}
		if p.Longitude > 180 {
			p.Longitude -= 360
		}
		s.Track = append(s.Track, p)
		s.TrackIllumination = append(s.TrackIllumination, satellite.Illumination(i%3))
		s.Instants = append(s.Instants, satellite.Coverage{Time: p.Time, SubPoint: p, Footprint: box(p.Longitude-.1, p.Latitude-.1, p.Longitude+.1, p.Latitude+.1)})
	}
	return s
}

func TestSatelliteTimeMarkerLimit(t *testing.T) {
	sweep := coverageJSONSweep()
	path := satelliteJSONPath()
	for _, fixture := range []struct {
		name    string
		span    time.Duration
		marshal func(geojson.SatellitePathOptions) ([]byte, error)
	}{
		{"ground track", sweep.End.Sub(sweep.Start), func(options geojson.SatellitePathOptions) ([]byte, error) {
			return geojson.MarshalSatelliteGroundTrack(sweep.Track, options)
		}},
		{"coverage sweep", sweep.End.Sub(sweep.Start), func(options geojson.SatellitePathOptions) ([]byte, error) {
			return geojson.MarshalSatelliteCoverageSweep(sweep, options)
		}},
		{"transit path", path.Event.End.Sub(path.Event.Start), func(options geojson.SatellitePathOptions) ([]byte, error) {
			return geojson.MarshalSatelliteTransitPath(path, options)
		}},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			step := fixture.span / 1441
			if _, err := fixture.marshal(geojson.SatellitePathOptions{TimeMarkerStep: step}); !errors.Is(err, satellite.ErrInvalidInput) {
				t.Fatalf("over the limit: %v", err)
			}
			raw, err := fixture.marshal(geojson.SatellitePathOptions{TimeMarkerStep: step + time.Nanosecond})
			if err != nil {
				t.Fatal(err)
			}
			var collection satelliteJSONCollection
			if err = json.Unmarshal(raw, &collection); err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, feature := range collection.Features {
				if feature.Properties["role"] == "time-marker" {
					count++
				}
			}
			if count != 1440 {
				t.Fatalf("marker count %d, want 1440", count)
			}
		})
	}
	track := append([]satellite.GroundPoint(nil), sweep.Track[:2]...)
	track[1].Time = track[0].Time.Add(1441 * time.Nanosecond)
	if _, err := geojson.MarshalSatelliteGroundTrack(track, geojson.SatellitePathOptions{TimeMarkerStep: time.Nanosecond}); err != nil {
		t.Fatalf("endpoint does not count as an interior marker: %v", err)
	}
	track[1].Time = track[1].Time.Add(time.Nanosecond)
	if _, err := geojson.MarshalSatelliteGroundTrack(track, geojson.SatellitePathOptions{TimeMarkerStep: time.Nanosecond}); !errors.Is(err, satellite.ErrInvalidInput) {
		t.Fatalf("one extra nanosecond adds an interior marker: %v", err)
	}
}

func TestSatelliteCoverageSweepConsistency(t *testing.T) {
	s := coverageJSONSweep()
	s.Elements = satellite.Elements{Name: "SECRET", CatalogNumber: 25544}
	obs := satellite.Observer{Longitude: 12, Latitude: 14}
	opt := geojson.SatellitePathOptions{TimeMarkerStep: time.Minute, Observer: &obs}
	data, err := geojson.MarshalSatelliteCoverageSweep(s, opt)
	if err != nil {
		t.Fatal(err)
	}
	var got satelliteJSONCollection
	if err = json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	roles := map[string]int{}
	illum := map[string]bool{}
	for _, f := range got.Features {
		role := f.Properties["role"].(string)
		roles[role]++
		if role == "current-satellite-footprint" {
			if _, ok := f.Properties["time"]; !ok {
				t.Fatal("missing time")
			}
		}
		if role == "ground-track" {
			illum[f.Properties["illumination"].(string)] = true
			if _, ok := f.Properties["times"]; !ok {
				t.Fatal("missing times")
			}
		}
		if role == "coverage-sweep" {
			var rings [][][]float64
			if err = json.Unmarshal(f.Geometry.Coordinates, &rings); err != nil || len(rings) != 2 {
				t.Fatal("holes", err, len(rings))
			}
			if rings[0][0][0] != -20 || rings[1][0][0] != -5 {
				t.Fatal(rings)
			}
		}
	}
	for _, r := range []string{"coverage-sweep", "visible-sweep", "ground-track", "time-marker", "observer", "current-satellite-footprint"} {
		if roles[r] == 0 {
			t.Fatal(r)
		}
	}
	if len(illum) != 3 || roles["current-satellite-footprint"] != len(s.Instants) {
		t.Fatal(illum, roles)
	}
	s.Elements = satellite.Elements{}
	other, err := geojson.MarshalSatelliteCoverageSweep(s, opt)
	if err != nil || string(data) != string(other) {
		t.Fatal("elements leaked", err)
	}
	track, err := geojson.MarshalSatelliteGroundTrack(s.Track, opt)
	if err != nil || len(track) == 0 {
		t.Fatal(err)
	}
}

func TestSatelliteCoverageInstantConsistency(t *testing.T) {
	s := coverageJSONSweep()
	s.Track = s.Track[:1]
	s.Instants = s.Instants[:1]
	s.TrackIllumination = s.TrackIllumination[:1]
	s.End = s.Start
	if _, err := geojson.MarshalSatelliteGroundTrack(s.Track, geojson.SatellitePathOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := geojson.MarshalSatelliteCoverageSweep(s, geojson.SatellitePathOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestSatelliteNightRegionTwilightMembership(t *testing.T) {
	data, err := os.ReadFile("../satellite/testdata/skyfield_coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Groups []struct{ Elements satellite.Elements }
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, example := range []struct {
		hour                    int
		threshold, sunLongitude float64
		hole                    bool
	}{{0, -6, 176, false}, {0, 0, 176, false}, {0, 6, 176, false}, {0, 30, 176, false}, {12, 30, -4, true}} {
		instant := time.Date(2026, 11, 14, example.hour, 0, 0, 0, time.UTC)
		coverage, err := satellite.CoverageAt(fixture.Groups[0].Elements, instant, satellite.CoverageOptions{TwilightSunAltitude: &example.threshold})
		if err != nil {
			t.Fatal(err)
		}
		raw, err := geojson.MarshalSatelliteCoverage(coverage)
		if err != nil {
			t.Fatal(err)
		}
		var collection satelliteJSONCollection
		if err = json.Unmarshal(raw, &collection); err != nil {
			t.Fatal(err)
		}
		var polygons [][][][]float64
		for _, feature := range collection.Features {
			if feature.Properties["role"] == "night-region" {
				if err := json.Unmarshal(feature.Geometry.Coordinates, &polygons); err != nil {
					t.Fatal(err)
				}
			}
		}
		hole := false
		for _, polygon := range polygons {
			hole = hole || len(polygon) > 1
			for index, ring := range polygon {
				area := 0.0
				for vertex, from := range ring[:len(ring)-1] {
					to := ring[vertex+1]
					if math.Abs(to[0]-from[0]) > 180 && !(from[1] == to[1] && math.Abs(from[1]) == 90) {
						t.Fatalf("uncut edge from %v to %v", from, to)
					}
					area += from[0]*to[1] - from[1]*to[0]
				}
				if (area > 0) != (index == 0) {
					t.Fatal("wrong ring winding")
				}
			}
		}
		if hole != example.hole {
			t.Errorf("%s threshold %g: hole=%t, want %t", instant, example.threshold, hole, example.hole)
		}
		for latitude := -80.0; latitude <= 80; latitude += 10 {
			for longitude := -178.0; longitude < 180; longitude += 10 {
				radians := math.Pi / 180
				altitude := math.Asin(math.Sin(latitude*radians)*math.Sin(-18*radians)+math.Cos(latitude*radians)*math.Cos(-18*radians)*math.Cos((longitude-example.sunLongitude)*radians)) / radians
				if math.Abs(altitude-example.threshold) < 2 {
					continue
				}
				if got := coverageJSONContains(polygons, longitude, latitude); got != (altitude < example.threshold) {
					t.Fatalf("%s threshold %g: filled=%t at (%.0f, %.0f), Sun alt %.2f", instant, example.threshold, got, longitude, latitude, altitude)
				}
			}
		}
	}
}

func coverageJSONContains(polygons [][][][]float64, longitude, latitude float64) bool {
	for _, polygon := range polygons {
		inside := false
		for _, ring := range polygon {
			for index, from := range ring[:len(ring)-1] {
				to := ring[index+1]
				if (from[1] > latitude) != (to[1] > latitude) && longitude < from[0]+(to[0]-from[0])*(latitude-from[1])/(to[1]-from[1]) {
					inside = !inside
				}
			}
		}
		if inside {
			return true
		}
	}
	return false
}

func TestSatelliteCoverageCoordinatesConsistency(t *testing.T) {
	data, err := os.ReadFile("../satellite/testdata/skyfield_coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Groups []struct {
			Elements satellite.Elements
			Rows     []struct{ Time time.Time }
		}
	}
	if err = json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	maximum := 0.0
	vertices := 0
	for _, g := range source.Groups {
		for _, row := range g.Rows[:2] {
			c, err := satellite.CoverageAt(g.Elements, row.Time, satellite.CoverageOptions{})
			if err != nil {
				t.Fatal(err)
			}
			raw, err := geojson.MarshalSatelliteCoverage(c)
			if err != nil {
				t.Fatal(err)
			}
			var collection satelliteJSONCollection
			if err = json.Unmarshal(raw, &collection); err != nil {
				t.Fatal(err)
			}
			var coords [][2]float64
			var scan func(interface{})
			scan = func(v interface{}) {
				a, ok := v.([]interface{})
				if !ok {
					return
				}
				if len(a) == 2 {
					x, xok := a[0].(float64)
					y, yok := a[1].(float64)
					if xok && yok {
						coords = append(coords, [2]float64{x, y})
						return
					}
				}
				for _, x := range a {
					scan(x)
				}
			}
			for _, f := range collection.Features {
				if f.Properties["role"] != "satellite-footprint" {
					continue
				}
				var v interface{}
				if err = json.Unmarshal(f.Geometry.Coordinates, &v); err != nil {
					t.Fatal(err)
				}
				scan(v)
			}
			for _, p := range c.Footprint {
				best := math.Inf(1)
				for _, q := range coords {
					best = math.Min(best, math.Hypot(math.Remainder(p.Longitude-q[0], 360), p.Latitude-q[1]))
				}
				maximum = math.Max(maximum, best)
				vertices++
				if best > 1e-9 {
					t.Fatal("lost coordinate", p, best)
				}
			}
		}
	}
	t.Logf("encoded boundary vertices=%d maximum coordinate error=%g degrees", vertices, maximum)
}

func TestSatelliteCoverageExportInvalidSynthetic(t *testing.T) {
	s := coverageJSONSweep()
	for _, opt := range []geojson.SatellitePathOptions{{TimeMarkerStep: -1}, {Nearby: &satellite.NearbyTransit{}}, {Observer: &satellite.Observer{Latitude: 91}}} {
		if _, err := geojson.MarshalSatelliteCoverageSweep(s, opt); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Fatal(opt, err)
		}
	}
	for _, q := range []satellite.CoverageSweep{{}, func() satellite.CoverageSweep { v := s; v.TrackIllumination = nil; return v }(), func() satellite.CoverageSweep {
		v := s
		v.Track = append([]satellite.GroundPoint(nil), s.Track...)
		v.Track[0].HeightKM = math.NaN()
		return v
	}()} {
		if _, err := geojson.MarshalSatelliteCoverageSweep(q, geojson.SatellitePathOptions{}); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Fatal(err)
		}
	}
	if _, err := geojson.MarshalSatelliteCoverage(satellite.Coverage{}); !errors.Is(err, satellite.ErrInvalidInput) {
		t.Fatal(err)
	}
}
func TestSatelliteCoverageGeneratorConsistency(t *testing.T) {
	s := coverageJSONSweep()
	opt := geojson.SatellitePathOptions{}
	a, err := geojson.MarshalSatelliteCoverageSweep(s, opt)
	if err != nil {
		t.Fatal(err)
	}
	b, err := geojson.NewWithInstance(ephem.NewModelInstance(ephem.Semianalytic)).MarshalSatelliteCoverageSweep(s, opt)
	if err != nil || !bytes.Equal(a, b) {
		t.Fatal(err)
	}
}

package kml

import (
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/satellite"
	"encoding/json"
	"encoding/xml"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSatelliteCoverageLayersSynthetic(t *testing.T) {
	at := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	ring := []satellite.GeoPoint{{Longitude: -20, Latitude: -20}, {Longitude: 20, Latitude: -20}, {Longitude: 20, Latitude: 20}, {Longitude: -20, Latitude: 20}, {Longitude: -20, Latitude: -20}}
	hole := []satellite.GeoPoint{{Longitude: -5, Latitude: -5}, {Longitude: -5, Latitude: 5}, {Longitude: 5, Latitude: 5}, {Longitude: 5, Latitude: -5}, {Longitude: -5, Latitude: -5}}
	s := satellite.CoverageSweep{Start: at, End: at.Add(3 * time.Minute), Swept: []satellite.Region{{Outer: ring, Holes: [][]satellite.GeoPoint{hole}}}, VisibleSwept: []satellite.Region{{Outer: ring}}}
	for i := 0; i < 4; i++ {
		p := satellite.GroundPoint{Time: at.Add(time.Duration(i) * time.Minute), Longitude: float64(i), Latitude: 10}
		s.Track = append(s.Track, p)
		s.TrackIllumination = append(s.TrackIllumination, satellite.Illumination(i%3))
		s.Instants = append(s.Instants, satellite.Coverage{Time: p.Time, Footprint: ring})
	}
	obs := satellite.Observer{Longitude: 10, Latitude: 10}
	data, err := geojson.MarshalSatelliteCoverageSweep(s, geojson.SatellitePathOptions{Observer: &obs})
	if err != nil {
		t.Fatal(err)
	}
	for _, language := range []string{"zh", "en"} {
		out, err := FromGeoJSON(data, Options{Language: language})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), "innerBoundaryIs") {
			t.Fatal("missing hole")
		}
		var got kmlRoot
		if err = xml.Unmarshal(out, &got); err != nil {
			t.Fatal(err)
		}
		for _, role := range []string{"coverage-sweep", "visible-sweep", "current-satellite-footprint", "ground-track", "observer"} {
			if len(placemarksOfRole(got, role)) == 0 {
				t.Fatal(role)
			}
		}
		for _, p := range placemarksOfRole(got, "current-satellite-footprint") {
			if p.TimeStamp == nil {
				t.Fatal("missing timestamp")
			}
		}
		tracks := placemarksOfRole(got, "ground-track")
		if len(tracks) != 3 {
			t.Fatalf("%d illumination segments, want 3", len(tracks))
		}
		for index, illumination := range []string{"sunlit", "penumbra", "umbra"} {
			want := timeSpan{Begin: s.Track[index].Time.Format(time.RFC3339Nano), End: s.Track[index+1].Time.Format(time.RFC3339Nano)}
			found := false
			for _, track := range tracks {
				if track.StyleURL == "#style-satellite-"+illumination+"-ground-track" {
					found = track.TimeSpan != nil && *track.TimeSpan == want
				}
			}
			if !found {
				t.Errorf("missing %s segment with interval %v", illumination, want)
			}
		}
		names := map[string][2]string{"coverage-sweep": {"扫掠可见区", "Coverage sweep"}, "visible-sweep": {"扫掠目视区", "Visual coverage sweep"}, "current-satellite-footprint": {"当前卫星可见圈", "Current satellite footprint"}, "ground-track": {"卫星地面轨迹", "Satellite ground track"}, "observer": {"观测点", "Observer"}}
		for _, pair := range names {
			name := pair[0]
			if language == "en" {
				name = pair[1]
			}
			found := false
			for _, f := range got.Document.Folders {
				found = found || f.Name == name || strings.HasSuffix(f.Name, " · "+name)
			}
			if !found {
				t.Fatal(language, name)
			}
		}
		for _, tc := range []struct{ id, color string }{{"style-satellite-coverage-coverage-sweep", "ffff8040"}, {"style-satellite-coverage-visible-sweep", "ff889b32"}, {"style-satellite-sunlit-ground-track", ColorYellow}, {"style-satellite-penumbra-ground-track", ColorOrange}, {"style-satellite-umbra-ground-track", "ff505050"}} {
			if styleByID(t, got, tc.id).LineStyle.Color != tc.color {
				t.Fatal(tc.id, styleByID(t, got, tc.id).LineStyle.Color, tc.color)
			}
		}
	}
}

func TestSatelliteCoverageDegenerateFootprint(t *testing.T) {
	at := time.Date(2026, 11, 14, 19, 27, 20, 0, time.UTC)
	instant := satellite.Coverage{Time: at, SubPoint: satellite.GroundPoint{Time: at, Longitude: 25, Latitude: 15}}
	sweep := satellite.CoverageSweep{Start: at, End: at, Track: []satellite.GroundPoint{instant.SubPoint}, TrackIllumination: []satellite.Illumination{satellite.IlluminationSunlit}, Instants: []satellite.Coverage{instant}}
	data, err := geojson.MarshalSatelliteCoverageSweep(sweep, geojson.SatellitePathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range [][]byte{data, mustMarshalDegenerateCoverage(t, instant)} {
		output, err := FromGeoJSON(input, Options{})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(output), "<Point>") || strings.Contains(string(output), "<Polygon>") {
			t.Fatal("coverage without area must contain a point and no polygon")
		}
	}
}

func mustMarshalDegenerateCoverage(t *testing.T, instant satellite.Coverage) []byte {
	t.Helper()
	data, err := geojson.MarshalSatelliteCoverage(instant)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestSatelliteGroundTrackKMLRejectsSingleVertexLine(t *testing.T) {
	raw := []byte(`{"type":"FeatureCollection","features":[{"type":"Feature","properties":{"event":"satellite-coverage","role":"ground-track","times":[["2026-11-14T00:00:00Z"]]},"geometry":{"type":"LineString","coordinates":[[12,34]]}}]}`)
	for _, noTimes := range []bool{false, true} {
		if _, err := FromGeoJSON(raw, Options{NoTimes: noTimes}); err == nil || !strings.Contains(err.Error(), "at least two positions") {
			t.Fatalf("NoTimes=%v: %v", noTimes, err)
		}
	}
}

func TestSatelliteGroundTrackKMLCoincidentSegmentSynthetic(t *testing.T) {
	at := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	track := []satellite.GroundPoint{
		{Time: at, Longitude: 180, Latitude: 34},
		{Time: at.Add(time.Minute), Longitude: -180, Latitude: 34},
		{Time: at.Add(2 * time.Minute), Longitude: -179, Latitude: 34},
	}
	raw, err := geojson.MarshalSatelliteGroundTrack(track, geojson.SatellitePathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, noTimes := range []bool{false, true} {
		out, err := FromGeoJSON(raw, Options{NoTimes: noTimes})
		if err != nil {
			t.Fatal(err)
		}
		var got kmlRoot
		if err := xml.Unmarshal(out, &got); err != nil {
			t.Fatal(err)
		}
		items := placemarksOfRole(got, "ground-track")
		if len(items) != 2 || items[0].Point == nil || items[0].Point.Coordinates != "180,34" || items[1].LineString == nil || items[1].LineString.Coordinates != "-180,34 -179,34" {
			t.Fatalf("expected a point and the remaining line: %+v", items)
		}
		if noTimes {
			if items[0].TimeSpan != nil || items[1].TimeSpan != nil || strings.Contains(string(out), `name="times"`) {
				t.Fatal("static output contains times")
			}
		} else {
			want := timeSpan{Begin: at.Format(time.RFC3339Nano), End: at.Add(time.Minute).Format(time.RFC3339Nano)}
			if items[0].TimeSpan == nil || *items[0].TimeSpan != want || !strings.Contains(string(out), `[&#34;`+want.Begin+`&#34;,&#34;`+want.End+`&#34;]`) {
				t.Fatalf("missing stationary interval or sample times: %+v", items[0])
			}
		}
	}
}

func TestSatelliteGroundTrackKMLSingleInstant(t *testing.T) {
	at := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	raw, err := geojson.MarshalSatelliteGroundTrack([]satellite.GroundPoint{{Time: at, Longitude: 12, Latitude: 34}}, geojson.SatellitePathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	out, err := FromGeoJSON(raw, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var got kmlRoot
	if err = xml.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	items := placemarksOfRole(got, "ground-track")
	if len(items) != 1 || items[0].Point == nil || items[0].LineString != nil || items[0].TimeStamp == nil || items[0].TimeStamp.When != at.Format(time.RFC3339Nano) {
		t.Fatalf("single instant must be a timestamped point: %+v", items)
	}
}

func TestSatelliteGroundTrackKMLVertexTimes(t *testing.T) {
	start := time.Date(2026, 11, 14, 0, 0, 0, 123456789, time.UTC)
	times := func(offsets ...time.Duration) []string {
		result := make([]string, len(offsets))
		for index, offset := range offsets {
			result[index] = start.Add(offset).Format(time.RFC3339Nano)
		}
		return result
	}
	for _, fixture := range []struct {
		name       string
		longitudes []float64
		want       [][]string
	}{
		{"continuous", []float64{178, 179, 180}, [][]string{times(0, time.Minute, 2*time.Minute)}},
		{"seam split", []float64{178, 179, -179}, [][]string{times(0, time.Minute, 90*time.Second), times(90*time.Second, 2*time.Minute)}},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			track := make([]satellite.GroundPoint, len(fixture.longitudes))
			for index, longitude := range fixture.longitudes {
				track[index] = satellite.GroundPoint{Time: start.Add(time.Duration(index) * time.Minute), Longitude: longitude, Latitude: 10}
			}
			raw, err := geojson.MarshalSatelliteGroundTrack(track, geojson.SatellitePathOptions{})
			if err != nil {
				t.Fatal(err)
			}
			for _, noTimes := range []bool{false, true} {
				out, err := FromGeoJSON(raw, Options{NoTimes: noTimes})
				if err != nil {
					t.Fatal(err)
				}
				var got kmlRoot
				if err = xml.Unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if noTimes && (got.Document.TimeSpan != nil || strings.Contains(string(out), `name="times"`)) {
					t.Fatal("static output contains times")
				}
				wantSpan := timeSpan{Begin: times(0)[0], End: times(2 * time.Minute)[0]}
				if !noTimes && (got.Document.TimeSpan == nil || *got.Document.TimeSpan != wantSpan) {
					t.Fatalf("document span %v, want %v", got.Document.TimeSpan, wantSpan)
				}
				items := placemarksOfRole(got, "ground-track")
				if len(items) != len(fixture.want) {
					t.Fatalf("%d placemarks, want %d", len(items), len(fixture.want))
				}
				for index, item := range items {
					want := fixture.want[index]
					if item.TimeStamp != nil || item.LineString == nil || len(strings.Fields(item.LineString.Coordinates)) != len(want) {
						t.Fatalf("segment %d: %+v", index, item)
					}
					if noTimes {
						if item.TimeSpan != nil {
							t.Fatal("static placemark has an interval")
						}
						continue
					}
					if item.TimeSpan == nil || *item.TimeSpan != (timeSpan{Begin: want[0], End: want[len(want)-1]}) {
						t.Fatalf("segment %d interval: %v, want %v", index, item.TimeSpan, want)
					}
					var vertexTimes []string
					for _, entry := range item.ExtendedData.Data {
						if entry.Name == "times" {
							if err := json.Unmarshal([]byte(entry.Value), &vertexTimes); err != nil {
								t.Fatal(err)
							}
						}
					}
					if !reflect.DeepEqual(vertexTimes, want) {
						t.Fatalf("segment %d times %v, want %v", index, vertexTimes, want)
					}
				}
			}
		})
	}
}

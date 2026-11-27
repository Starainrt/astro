package kml

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/satellite"
)

func satelliteKMLPath() satellite.TransitPath {
	at := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	a := satellite.TransitPathPoint{Time: at, Longitude: 10, Latitude: 20}
	b := satellite.TransitPathPoint{Time: at.Add(time.Second), Longitude: 11, Latitude: 21}
	c := satellite.TransitPathPoint{Time: at.Add(2 * time.Second), Longitude: 12, Latitude: 22}
	left := []satellite.TransitPathPoint{a, b, c}
	right := append([]satellite.TransitPathPoint(nil), left...)
	for i := range left {
		left[i].Latitude += .1
		right[i].Latitude -= .1
	}
	return satellite.TransitPath{Elements: satellite.Elements{Name: "ISS", CatalogNumber: 25544, Epoch: at}, Event: satellite.TransitEvent{Target: satellite.TransitSun, Start: a.Time, Greatest: b.Time, End: c.Time, StartPoint: a, GreatestPoint: b, EndPoint: c, Complete: true}, CenterLine: []satellite.TransitPathPoint{a, b, c}, LeftLimit: left, RightLimit: right, StartHorizon: []satellite.GeoPoint{{Longitude: 10, Latitude: 20}, {Longitude: 20, Latitude: 30}}, EndHorizon: []satellite.GeoPoint{{Longitude: 11, Latitude: 21}, {Longitude: 21, Latitude: 31}}}
}
func TestSatelliteLayersSynthetic(t *testing.T) {
	path := satelliteKMLPath()
	obs := satellite.Observer{Longitude: 11, Latitude: 20}
	near := satellite.NearbyTransit{Event: path.Event, Nearest: path.CenterLine[1], DistanceKM: 1, NearestDistanceKM: 2}
	data, err := geojson.MarshalSatelliteTransitPath(path, geojson.SatellitePathOptions{TimeMarkerStep: time.Second, Observer: &obs, Nearby: &near})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string][2]string{"transit-band": {"凌现象带", "Transit band"}, "center-line": {"中心线", "Center line"}, "left-limit": {"左限", "Left limit"}, "right-limit": {"右限", "Right limit"}, "target-horizon": {"目标地平线", "Target horizon"}, "start": {"凌始", "Transit start"}, "greatest": {"凌甚", "Greatest transit"}, "end": {"凌终", "Transit end"}, "time-marker": {"时间标记", "Time markers"}, "observer": {"观测点", "Observer"}, "nearest-point": {"最近点", "Nearest point"}, "nearest-connector": {"最近点连线", "Nearest-point connector"}}
	colors := map[string]string{"transit-band": ColorYellow, "center-line": ColorRed, "left-limit": ColorOrange, "right-limit": ColorOrange, "target-horizon": ColorGray}
	for i, lang := range []string{"zh", "en"} {
		out, err := FromGeoJSON(data, Options{Language: lang})
		if err != nil {
			t.Fatal(err)
		}
		var got kmlRoot
		if err = xml.Unmarshal(out, &got); err != nil {
			t.Fatal(err)
		}
		wantLegend := [2]string{"红：凌现象中心线", "red: satellite transit center line"}[i]
		if !strings.Contains(got.Document.Description, wantLegend) {
			t.Fatalf("%s satellite legend: %q", lang, got.Document.Description)
		}
		wantDoc := [2]string{"卫星凌现象", "Satellite transit"}[i]
		if !strings.HasPrefix(got.Document.Name, wantDoc) {
			t.Fatalf("%s document name: %q", lang, got.Document.Name)
		}
		for role, pair := range names {
			items := placemarksOfRole(got, role)
			if len(items) == 0 {
				t.Fatal(role)
			}
			found := false
			for _, f := range got.Document.Folders {
				if f.Name == pair[i] {
					found = true
				}
			}
			if !found {
				t.Fatal(lang, role, pair[i])
			}
			if c, ok := colors[role]; ok {
				style := styleByID(t, got, items[0].StyleURL[1:])
				if style.LineStyle.Color != c {
					t.Fatal(role, style.LineStyle.Color, c)
				}
			}
		}
		for _, pm := range placemarksOfRole(got, "center-line") {
			want := timeSpan{Begin: path.Event.Start.Format(time.RFC3339Nano), End: path.Event.End.Format(time.RFC3339Nano)}
			if pm.TimeStamp != nil || pm.TimeSpan == nil || *pm.TimeSpan != want {
				t.Fatalf("center-line interval: %v %v", pm.TimeStamp, pm.TimeSpan)
			}
		}
	}
	shadow := satellite.TransitShadow{Time: path.Event.Start, Target: satellite.TransitSun, Valid: true, Point: path.Event.StartPoint, SubPoint: satellite.GroundPoint{Longitude: 10, Latitude: 20}, Footprint: []satellite.GeoPoint{{Longitude: 10, Latitude: 20}, {Longitude: 11, Latitude: 20}, {Longitude: 11, Latitude: 21}, {Longitude: 10, Latitude: 20}}}
	data, err = geojson.MarshalSatelliteTransitShadow(shadow)
	if err != nil {
		t.Fatal(err)
	}
	for i, lang := range []string{"zh", "en"} {
		out, err := FromGeoJSON(data, Options{Language: lang})
		if err != nil {
			t.Fatal(err)
		}
		var got kmlRoot
		xml.Unmarshal(out, &got)
		for role, pair := range map[string][2]string{"current-transit-footprint": {"当前凌现象足迹", "Current transit footprint"}, "current-transit-point": {"当前中心点", "Current transit centre"}, "subsatellite-point": {"星下点", "Subsatellite point"}} {
			items := placemarksOfRole(got, role)
			if len(items) == 0 || items[0].TimeStamp == nil {
				t.Fatal(role)
			}
			found := false
			for _, f := range got.Document.Folders {
				if f.Name == pair[i] {
					found = true
				}
			}
			if !found {
				t.Fatal(role, lang)
			}
			if role == "current-transit-footprint" {
				st := styleByID(t, got, items[0].StyleURL[1:])
				if st.LineStyle.Color != "ff00008b" {
					t.Fatal(st.LineStyle)
				}
			}
		}
	}
}
func TestSatelliteExistingKMLConsistency(t *testing.T) {
	data := []byte(`{"type":"FeatureCollection","features":[{"type":"Feature","properties":{"event":"solar-eclipse","role":"center-line","times":[["2026-11-14T00:00:00Z","2026-11-14T00:00:01Z"]]},"geometry":{"type":"MultiLineString","coordinates":[[[10,20],[11,21]]]}},{"type":"Feature","properties":{"event":"lunar-occultation","role":"occultation-band"},"geometry":{"type":"Polygon","coordinates":[[[10,20],[11,20],[11,21],[10,20]]]}}]}`)
	for _, row := range []struct {
		lang, hash string
		size       int
	}{{"zh", "11b88a3322591e4a7868fee6807aac47", 2170}, {"en", "de4d60102fd1a1371b5b0dbb9afe5706", 2206}} {
		got, err := FromGeoJSON(data, Options{Language: row.lang})
		if err != nil || len(got) != row.size || fmt.Sprintf("%x", md5.Sum(got)) != row.hash {
			t.Fatal(row.lang, err, len(got), fmt.Sprintf("%x", md5.Sum(got)))
		}
	}
}

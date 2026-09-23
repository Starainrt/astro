package geojson_test

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

// 掩星 GeoJSON 与日月食共用同一份时标契约：UT1 必须写 time_scale 成员、换掉全部时刻，
// 且几何逐字节不变；UT1 配非 UTC 时区必须报错而不是静默按 UTC 导出。
// Occultation GeoJSON shares the eclipse time-scale contract.
func TestOccultationGeoJSONTimeScale(t *testing.T) {
	antares := moon.StarCoordinate{
		ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
		Epoch:                          time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          moon.CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -10,
		ProperMotionDecMasPerYear:      -20,
		ParallaxMas:                    24,
	}
	starStart := time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC)
	starPaths, err := moon.FindStarOccultationPaths(starStart, starStart.Add(24*time.Hour), antares,
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200})
	if err != nil || len(starPaths) != 1 {
		t.Fatalf("star paths=%d err=%v", len(starPaths), err)
	}
	planetStart := time.Date(2024, time.September, 5, 0, 0, 0, 0, time.UTC)
	planetPaths, err := moon.FindPlanetOccultationPaths(planetStart, planetStart.AddDate(0, 0, 1),
		moon.OccultationVenus, moon.OccultationPathOptions{})
	if err != nil || len(planetPaths) != 1 {
		t.Fatalf("planet paths=%d err=%v", len(planetPaths), err)
	}
	utc := time.UTC
	cst := time.FixedZone("CST", 8*3600)
	for _, testCase := range []struct {
		name    string
		marshal func(geojson.TimeMarkerOptions) ([]byte, error)
	}{
		{"star", func(options geojson.TimeMarkerOptions) ([]byte, error) {
			return geojson.MarshalStarOccultationWithTimeMarkers(starPaths[0], options)
		}},
		{"planet", func(options geojson.TimeMarkerOptions) ([]byte, error) {
			return geojson.MarshalPlanetOccultationWithTimeMarkers(planetPaths[0], options)
		}},
	} {
		utcDoc, err := testCase.marshal(geojson.TimeMarkerOptions{Step: time.Hour})
		if err != nil {
			t.Fatalf("%s: UTC export: %v", testCase.name, err)
		}
		ut1Doc, err := testCase.marshal(geojson.TimeMarkerOptions{Step: time.Hour, TimeScale: astro.TimeScaleUT1})
		if err != nil {
			t.Fatalf("%s: UT1 export: %v", testCase.name, err)
		}
		if _, err := testCase.marshal(geojson.TimeMarkerOptions{
			Step: time.Hour, TimeScale: astro.TimeScaleUT1, Location: cst,
		}); err == nil {
			t.Errorf("%s: UT1 with a non-UTC location must fail", testCase.name)
		}
		var utcHeader, ut1Header struct {
			TimeScale string `json:"time_scale"`
		}
		if err := json.Unmarshal(utcDoc, &utcHeader); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(ut1Doc, &ut1Header); err != nil {
			t.Fatal(err)
		}
		if utcHeader.TimeScale != "" {
			t.Errorf("%s: UTC export carries time_scale=%q", testCase.name, utcHeader.TimeScale)
		}
		if ut1Header.TimeScale != "UT1" {
			t.Errorf("%s: UT1 export time_scale=%q", testCase.name, ut1Header.TimeScale)
		}
		if !reflect.DeepEqual(decodeGeometryOnly(t, utcDoc), decodeGeometryOnly(t, ut1Doc)) {
			t.Errorf("%s: switching to UT1 changed geometry", testCase.name)
		}
		utcGreatest := featureTimeProperty(t, utcDoc, "greatest")
		ut1Greatest := featureTimeProperty(t, ut1Doc, "greatest")
		shift := ut1Greatest.Sub(utcGreatest).Seconds()
		if math.Abs(shift-astro.DUT1(utcGreatest)) > 1e-3 {
			t.Errorf("%s: greatest time shifted %.6f s, want DUT1 %.6f s",
				testCase.name, shift, astro.DUT1(utcGreatest))
		}
	}
	_ = utc
}

func featureTimeProperty(t *testing.T, data []byte, role string) time.Time {
	t.Helper()
	var collection struct {
		Features []struct {
			Properties map[string]interface{} `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatal(err)
	}
	for _, feature := range collection.Features {
		if feature.Properties["role"] != role {
			continue
		}
		text, _ := feature.Properties["time"].(string)
		parsed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			t.Fatalf("role %s time=%q: %v", role, text, err)
		}
		return parsed
	}
	t.Fatalf("role %s not found", role)
	return time.Time{}
}

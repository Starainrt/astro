package geojson_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

// 切换输出时标只能改时刻文字，不能改几何：UT1 读数一旦被当成民用时刻喂给几何，
// 月下点、地平闭合弧与食甚点都会平移。
// Switching the output time scale may only rewrite instants, never geometry.
func TestUT1ExportKeepsGeometry(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	for _, testCase := range []struct {
		name  string
		build func(t *testing.T) ([]byte, []byte)
	}{
		{"solar partial footprints", func(t *testing.T) ([]byte, []byte) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 20 * time.Minute, BoundaryPoints: 36,
			})
			if !ok {
				t.Fatal("expected solar partial footprints")
			}
			utc, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, nil, geojson.TimeMarkerOptions{Step: time.Hour})
			if err != nil {
				t.Fatal(err)
			}
			ut1, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, nil, geojson.TimeMarkerOptions{
				Step: time.Hour, TimeScale: astro.TimeScaleUT1,
			})
			if err != nil {
				t.Fatal(err)
			}
			return utc, ut1
		}},
		// 1995-10-24 是 central_two_limits 事件：极限带几何按时刻插值，最容易再把 UT1 读数当民用时刻。
		{"solar central two limits", func(t *testing.T) ([]byte, []byte) {
			date := time.Date(1995, time.October, 24, 0, 0, 0, 0, time.UTC)
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 20 * time.Minute, BoundaryPoints: 36,
			})
			if !ok {
				t.Fatal("expected 1995-10-24 partial footprints")
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 10 * time.Minute})
			if !ok {
				t.Fatal("expected 1995-10-24 central path")
			}
			utc, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, &central, geojson.TimeMarkerOptions{Step: time.Hour})
			if err != nil {
				t.Fatal(err)
			}
			ut1, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, &central, geojson.TimeMarkerOptions{
				Step: time.Hour, TimeScale: astro.TimeScaleUT1,
			})
			if err != nil {
				t.Fatal(err)
			}
			return utc, ut1
		}},
		{"lunar eclipse", func(t *testing.T) ([]byte, []byte) {
			info, ok := eclipse.LunarEclipseOnDate(time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC))
			if !ok {
				t.Fatal("expected lunar eclipse")
			}
			utc, err := geojson.MarshalLunarEclipseWithTimeMarkers(info, 24, geojson.TimeMarkerOptions{Step: time.Hour})
			if err != nil {
				t.Fatal(err)
			}
			ut1, err := geojson.MarshalLunarEclipseWithTimeMarkers(info, 24, geojson.TimeMarkerOptions{
				Step: time.Hour, TimeScale: astro.TimeScaleUT1,
			})
			if err != nil {
				t.Fatal(err)
			}
			return utc, ut1
		}},
	} {
		utc, ut1 := testCase.build(t)
		utcValue := decodeGeometryOnly(t, utc)
		ut1Value := decodeGeometryOnly(t, ut1)
		if !reflect.DeepEqual(utcValue, ut1Value) {
			t.Errorf("%s: UT1 export changed geometry", testCase.name)
		}
		var collection struct {
			TimeScale string `json:"time_scale"`
		}
		if err := json.Unmarshal(ut1, &collection); err != nil {
			t.Fatal(err)
		}
		if collection.TimeScale != "UT1" {
			t.Errorf("%s: UT1 export time_scale=%q", testCase.name, collection.TimeScale)
		}
		if testCase.name == "lunar eclipse" {
			utcTime := featureTimeProperty(t, utc, "visible-at-p1")
			ut1Time := featureTimeProperty(t, ut1, "visible-at-p1")
			if utcTime.Equal(ut1Time) {
				t.Errorf("lunar eclipse P1 time was not converted to UT1: %s", utcTime)
			}
			if delta := ut1Time.Sub(utcTime); delta < -time.Second || delta > time.Second {
				t.Errorf("lunar eclipse P1 UTC/UT1 delta=%s, want sub-second", delta)
			}
		}
	}
}

// decodeGeometryOnly 去掉只随时标变化的成员，保留其余结构用于逐项比较。
func decodeGeometryOnly(t *testing.T, data []byte) interface{} {
	t.Helper()
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	return stripTimeScaleMembers(value)
}

func stripTimeScaleMembers(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		// 时间标记的取点跟着所标时刻走（UT1 整点是另一个物理时刻），几何不变性对它不适用。
		if properties, ok := typed["properties"].(map[string]interface{}); ok {
			if role, _ := properties["role"].(string); role == "time-marker" {
				return nil
			}
		}
		for _, key := range []string{"time", "times", "label", "time_scale"} {
			delete(typed, key)
		}
		for key, item := range typed {
			// 时间属性名随结构而变（penumbral_start、partial_end_on_earth 之类），按值判定更稳。
			if text, ok := item.(string); ok {
				if _, err := time.Parse(time.RFC3339Nano, text); err == nil {
					delete(typed, key)
					continue
				}
			}
			typed[key] = stripTimeScaleMembers(item)
		}
		return typed
	case []interface{}:
		for index, item := range typed {
			typed[index] = stripTimeScaleMembers(item)
		}
		return typed
	}
	return value
}

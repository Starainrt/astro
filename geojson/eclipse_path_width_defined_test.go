package geojson

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
)

// 带宽契约：单侧极限事件的 path_width_km 与"非中心食"一样是 0，消费者只能靠
// path_width_defined 区分，该键必须与 path_width_km 并列出现在同一份 properties 里。

func TestSolarEclipseMetadataPathWidthDefined(t *testing.T) {
	testCases := []struct {
		name    string
		date    time.Time
		defined bool
		widthKM float64
	}{
		{name: "-1404-01-07 single-sided limit", date: time.Date(-1404, time.January, 7, 0, 0, 0, 0, time.UTC)},
		{name: "2003-05-31 single-sided limit", date: time.Date(2003, time.May, 31, 0, 0, 0, 0, time.UTC)},
		{name: "1874-10-10 single-sided limit", date: time.Date(1874, time.October, 10, 0, 0, 0, 0, time.UTC)},
		{name: "2024-04-08 total", date: time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC), defined: true, widthKM: 198.6161},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, ok := eclipsecore.SolarEclipseOnDate(tc.date)
			if !ok {
				t.Fatalf("no solar eclipse on %v", tc.date)
			}
			properties := solarEclipseMetadata(info)
			defined, isBool := properties["path_width_defined"].(bool)
			if !isBool {
				t.Fatalf("path_width_defined=%#v, want a bool", properties["path_width_defined"])
			}
			if defined != tc.defined {
				t.Fatalf("path_width_defined=%v want %v", defined, tc.defined)
			}
			if defined != info.PathWidthDefined {
				t.Fatalf("path_width_defined=%v disagrees with eclipse.PathWidthDefined=%v", defined, info.PathWidthDefined)
			}
			width, isFloat := properties["path_width_km"].(float64)
			if !isFloat {
				t.Fatalf("path_width_km=%#v, want a number", properties["path_width_km"])
			}
			if !tc.defined {
				if width != 0 {
					t.Fatalf("path_width_km=%.6f with an undefined width, want 0", width)
				}
				return
			}
			if math.Abs(width-tc.widthKM) > 0.01 {
				t.Fatalf("path_width_km=%.6f want %.6f", width, tc.widthKM)
			}
		})
	}
}

func TestMarshalSolarEclipseCarriesPathWidthDefined(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipsecore.SolarEclipsePartialFootprints(date, eclipsecore.SolarEclipsePartialFootprintOptions{
		Step: 20 * time.Minute, BoundaryPoints: 36,
	})
	if !ok {
		t.Fatalf("no partial footprints on %v", date)
	}
	central, hasCentral := eclipsecore.SolarEclipseCentralPath(date, eclipsecore.SolarEclipsePathOptions{Step: 5 * time.Minute})
	var centralPath *eclipsecore.SolarEclipsePath
	if hasCentral {
		centralPath = &central
	}
	data, err := MarshalSolarEclipse(partial, centralPath)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	var collection featureCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(collection.Features) == 0 {
		t.Fatal("no features exported")
	}
	// 带宽两个键只在食甚点要素上出现；其余要素沿用最小 properties。
	seen := 0
	for index, item := range collection.Features {
		if item.Properties["role"] != "greatest" {
			continue
		}
		defined, ok := item.Properties["path_width_defined"].(bool)
		if !ok {
			t.Fatalf("greatest feature has no boolean path_width_defined: %#v", item.Properties["path_width_defined"])
		}
		width, ok := item.Properties["path_width_km"].(float64)
		if !ok {
			t.Fatalf("greatest feature has no numeric path_width_km: %#v", item.Properties["path_width_km"])
		}
		if !defined || math.Abs(width-198.6161) > 0.01 {
			t.Fatalf("greatest feature %d path_width_km=%.6f path_width_defined=%v, want 198.6161/true", index, width, defined)
		}
		seen++
	}
	if seen != 1 {
		t.Fatalf("greatest features with width properties = %d, want exactly one", seen)
	}
}

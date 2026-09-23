package geojson_test

import (
	"bytes"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

// 契约：SolarEclipseOptions.SkipRoles 只裁掉列出的 role、默认输出不变、全部裁光时报错；TimeMarkers 与旧入口等价。

func solarEclipseOptionsFixture(t *testing.T) (eclipse.SolarEclipsePartialFootprintsInfo, eclipse.SolarEclipsePath) {
	t.Helper()
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 24,
	})
	if !ok {
		t.Fatal("expected the 2024-04-08 partial phase")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: time.Minute, TargetSpacingKM: 500,
	})
	if !ok {
		t.Fatal("expected the 2024-04-08 central path")
	}
	return partial, central
}

func solarRoleCounts(collection decodedCollection) map[string]int {
	counts := map[string]int{}
	for _, feature := range collection.Features {
		role, _ := feature.Properties["role"].(string)
		counts[role]++
	}
	return counts
}

func TestMarshalSolarEclipseSkipRoles(t *testing.T) {
	partial, central := solarEclipseOptionsFixture(t)
	baseline, err := geojson.MarshalSolarEclipseWithOptions(partial, &central, geojson.SolarEclipseOptions{})
	if err != nil {
		t.Fatalf("MarshalSolarEclipseWithOptions: %v", err)
	}
	legacy, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	if !bytes.Equal(baseline, legacy) {
		t.Fatal("零值 SolarEclipseOptions 应等价于 MarshalSolarEclipse")
	}
	full := solarRoleCounts(decodeCollection(t, baseline))
	if full["partial-footprint"] == 0 {
		t.Fatalf("默认输出应含瞬时半影轮廓：%v", full)
	}

	skip := []string{"partial-footprint", "partial-band"}
	data, err := geojson.MarshalSolarEclipseWithOptions(partial, &central, geojson.SolarEclipseOptions{SkipRoles: skip})
	if err != nil {
		t.Fatalf("MarshalSolarEclipseWithOptions: %v", err)
	}
	filtered := solarRoleCounts(decodeCollection(t, data))
	want := map[string]int{}
	for role, count := range full {
		want[role] = count
	}
	for _, role := range skip {
		if filtered[role] != 0 {
			t.Fatalf("被跳过的 %s 仍输出 %d 个", role, filtered[role])
		}
		delete(want, role)
	}
	if !reflect.DeepEqual(filtered, want) {
		t.Fatalf("跳过 %v 后其余图层被改动：got %v want %v", skip, filtered, want)
	}

	all := make([]string, 0, len(full))
	for role := range full {
		all = append(all, role)
	}
	sort.Strings(all)
	if _, err := geojson.MarshalSolarEclipseWithOptions(partial, nil, geojson.SolarEclipseOptions{SkipRoles: all}); err == nil {
		t.Fatal("全部 role 都跳过时应返回错误")
	}
}

func TestMarshalSolarEclipseOptionsTimeMarkers(t *testing.T) {
	partial, central := solarEclipseOptionsFixture(t)
	markers := geojson.TimeMarkerOptions{Step: 30 * time.Minute}

	legacy, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, &central, markers)
	if err != nil {
		t.Fatalf("MarshalSolarEclipseWithTimeMarkers: %v", err)
	}
	current, err := geojson.MarshalSolarEclipseWithOptions(partial, &central, geojson.SolarEclipseOptions{TimeMarkers: &markers})
	if err != nil {
		t.Fatalf("MarshalSolarEclipseWithOptions: %v", err)
	}
	if !bytes.Equal(legacy, current) {
		t.Fatal("TimeMarkers 字段与 MarshalSolarEclipseWithTimeMarkers 不等价")
	}

	both, err := geojson.MarshalSolarEclipseWithOptions(partial, &central, geojson.SolarEclipseOptions{
		TimeMarkers: &markers,
		SkipRoles:   []string{"partial-footprint"},
	})
	if err != nil {
		t.Fatalf("MarshalSolarEclipseWithOptions: %v", err)
	}
	counts := solarRoleCounts(decodeCollection(t, both))
	if counts["time-marker"] == 0 {
		t.Fatalf("时间标记被误删：%v", counts)
	}
	if counts["partial-footprint"] != 0 {
		t.Fatalf("瞬时半影轮廓未丢弃：%v", counts)
	}
}

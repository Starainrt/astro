package svg

import (
	"strings"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
)

// 核心判定当天无日食时，三个日食图入口都必须返回 false，与 LunarEclipseMapSVG 的日期门一致；
// 偏食足迹是“取最近一次”语义，缺这道门会把邻近日期的图当成当天的图交出去。
func TestSolarEclipseMapSVGRequiresEclipseOnDate(t *testing.T) {
	type entry struct {
		name   string
		render func(time.Time, SolarEclipseMapSVGOptions) (string, bool)
		global func(time.Time) (eclipsecore.SolarEclipseInfo, bool)
	}
	entries := []entry{
		{
			name:   "default",
			render: SolarEclipseMapSVG,
			global: eclipsecore.SolarEclipseOnDateNASABulletinSplitK,
		},
		{
			name:   "nasaSplitK",
			render: SolarEclipseMapSVGNASABulletinSplitK,
			global: eclipsecore.SolarEclipseOnDateNASABulletinSplitK,
		},
		{
			name:   "iauSingleK",
			render: SolarEclipseMapSVGIAUSingleK,
			global: eclipsecore.SolarEclipseOnDateIAUSingleK,
		},
	}
	// 2024-04-08 是本组日期里唯一有日食的一天，其余都在最近一次日食前后半个朔望月内。
	for _, test := range []struct {
		day  string
		want bool
	}{
		{day: "2024-04-08", want: true},
		{day: "2024-04-15"},
		{day: "2024-04-20"},
		{day: "2024-04-23"},
		{day: "2024-04-24"},
	} {
		date, err := time.Parse("2006-01-02", test.day)
		if err != nil {
			t.Fatalf("bad fixture date %q: %v", test.day, err)
		}
		for _, entry := range entries {
			_, coreOK := entry.global(date)
			if coreOK != test.want {
				t.Fatalf("%s %s: core eclipse gate = %v, want %v", entry.name, test.day, coreOK, test.want)
			}
			rendered, ok := entry.render(date, SolarEclipseMapSVGOptions{
				Width: 900, Height: 620, Location: time.UTC, PartialStep: 10 * time.Minute,
			})
			if ok != test.want {
				t.Fatalf("%s %s: map gate = %v, want %v (core says %v)", entry.name, test.day, ok, test.want, coreOK)
			}
			if test.want && rendered == "" {
				t.Fatalf("%s %s: gate passed but the map is empty", entry.name, test.day)
			}
			if !test.want && rendered != "" {
				t.Fatalf("%s %s: rejected date still returned %d bytes", entry.name, test.day, len(rendered))
			}
		}
	}
}

// 过了日期门的日期必须仍带完整等时线层：等时线自己还要查一次核心，门不能把它挡掉。
func TestSolarEclipseMapOnDateKeepsIsochrones(t *testing.T) {
	rendered, ok := SolarEclipseMapSVG(
		time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{
			Width: 900, Height: 620, Location: time.UTC,
			PartialStep: 10 * time.Minute, GreatestTimeStep: 30 * time.Minute,
		},
	)
	if !ok {
		t.Fatal("2024-04-08 is an eclipse date but the map was rejected")
	}
	if got := strings.Count(rendered, `class="solar-greatest-time-isoline"`); got != 7 {
		t.Fatalf("isochrone paths = %d, want the full 7-branch fan", got)
	}
}

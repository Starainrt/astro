package geojson_test

// 开启月缘时偏食足迹与食甚等时线必须落在月缘修正后的 P1–P4 内，整条管线（足迹 → 中心带 → 编码）不得被校验拒收。

import (
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon/limb"
)

func TestMarshalSolarEclipseWithLunarLimb(t *testing.T) {
	module := eclipse.NewWithInstance(nil).WithLunarLimb(limb.Default())
	for _, date := range []time.Time{
		time.Date(2028, 7, 22, 0, 0, 0, 0, time.UTC),
		time.Date(2029, 6, 12, 0, 0, 0, 0, time.UTC),
		time.Date(2032, 5, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2040, 11, 4, 0, 0, 0, 0, time.UTC),
	} {
		label := date.Format("2006-01-02")
		partial, ok := module.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{})
		if !ok || len(partial.Footprints) == 0 {
			t.Fatalf("%s: footprints unavailable", label)
		}
		if !partial.Eclipse.LunarLimbApplied {
			t.Fatalf("%s: limb not applied to the global contacts", label)
		}
		begin, end := partial.Eclipse.PartialBeginOnEarth, partial.Eclipse.PartialEndOnEarth
		for index, footprint := range partial.Footprints {
			if footprint.Time.Before(begin) || footprint.Time.After(end) {
				t.Fatalf("%s: footprint %d at %v outside limb P1–P4 %v–%v", label, index, footprint.Time, begin, end)
			}
		}
		for index, contour := range partial.GreatestTimeContours {
			if contour.Time.Before(begin) || contour.Time.After(end) {
				t.Errorf("%s: greatest-time contour %d at %v outside limb P1–P4", label, index, contour.Time)
			}
		}
		var central *eclipse.SolarEclipsePath
		if path, ok := module.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{}); ok {
			central = &path
		}
		if _, err := geojson.MarshalSolarEclipse(partial, central); err != nil {
			t.Errorf("%s: %v", label, err)
		}
	}
}

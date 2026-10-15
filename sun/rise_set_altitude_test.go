// 被测契约：默认选项与 RiseTime/SetTime 一致；夏令时切换日结果落在当天；实例方法使用绑定实例。
package sun

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

func TestRiseSetMatchesRiseTimeSetTime(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	for i := 0; i < 30; i++ {
		date := time.Date(2025, 1, 1, 12, 0, 0, 0, loc).AddDate(0, 0, i*12)
		got, err := RiseSet(date, 116.4, 39.9, 50, basic.AltitudeCrossingOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkEvent(t, date, "rise", got.Rise, got.HasRise, func() (time.Time, error) { return RiseTime(date, 116.4, 39.9, 50, true) })
		checkEvent(t, date, "set", got.Set, got.HasSet, func() (time.Time, error) { return SetTime(date, 116.4, 39.9, 50, true) })
	}
}

func TestRiseSetDaylightSavingDays(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip(err)
	}
	for _, date := range []time.Time{
		time.Date(2023, 3, 12, 12, 0, 0, 0, loc),
		time.Date(2023, 11, 5, 12, 0, 0, 0, loc),
	} {
		got, err := RiseSet(date, -74, 40.7, 0, basic.AltitudeCrossingOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkEvent(t, date, "rise", got.Rise, got.HasRise, func() (time.Time, error) { return RiseTime(date, -74, 40.7, 0, true) })
		checkEvent(t, date, "set", got.Set, got.HasSet, func() (time.Time, error) { return SetTime(date, -74, 40.7, 0, true) })
		if got.Rise.Location() != loc || got.Rise.YearDay() != date.YearDay() || got.Set.YearDay() != date.YearDay() {
			t.Errorf("%s: events %v %v outside civil day", date.Format("2006-01-02"), got.Rise, got.Set)
		}
	}
}

func TestRiseSetUsesModuleInstance(t *testing.T) {
	date := time.Date(2025, 3, 20, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	inst := ephem.NewModelInstance(ephem.Legacy)
	got, err := NewWithInstance(inst).RiseSet(date, 116.4, 39.9, 0, basic.AltitudeCrossingOptions{Limb: basic.AltitudeLimbCenter})
	if err != nil {
		t.Fatal(err)
	}
	want, err := basic.SkyAltitudeCrossingOnDate(basic.SkySun, date, 116.4, 39.9, 0, 0, basic.AltitudeCrossingOptions{Limb: basic.AltitudeLimbCenter}, inst)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Rise.Equal(want.Rise) || !got.Set.Equal(want.Set) {
		t.Errorf("instance result %+v, want %+v", got, want)
	}
}

func checkEvent(t *testing.T, date time.Time, kind string, got time.Time, has bool, ref func() (time.Time, error)) {
	t.Helper()
	want, err := ref()
	if has != (err == nil) {
		t.Errorf("%s %s: has=%v, reference err=%v", date.Format("2006-01-02"), kind, has, err)
		return
	}
	if has {
		if d := got.Sub(want); d > 100*time.Millisecond || d < -100*time.Millisecond {
			t.Errorf("%s %s: %v, reference %v", date.Format("2006-01-02"), kind, got, want)
		}
	}
}

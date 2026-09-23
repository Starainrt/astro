package astro_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/jupiter"
	litemoon "github.com/starainrt/astro/lite/moon"
	litesun "github.com/starainrt/astro/lite/sun"
	"github.com/starainrt/astro/mars"
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/neptune"
	"github.com/starainrt/astro/orbit"
	"github.com/starainrt/astro/saturn"
	"github.com/starainrt/astro/star"
	"github.com/starainrt/astro/sun"
	"github.com/starainrt/astro/uranus"
	"github.com/starainrt/astro/venus"
)

type civilEventCase struct {
	name     string
	event    func(time.Time) (time.Time, error)
	altitude func(time.Time) float64
}

func civilEventCases(lon, lat float64) []civilEventCase {
	elements := orbit.Elements{EpochJD: 2461000.5, A: 2.765615651508659, E: 0.07957631994408416,
		I: 10.58788658206854, Omega: 80.24963090816965, W: 73.29975464616518, M0: 231.5397330043706}
	return []civilEventCase{
		{"sun/RiseTime", func(d time.Time) (time.Time, error) { return sun.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return sun.Altitude(d, lon, lat) }},
		{"sun/RiseTimeN", func(d time.Time) (time.Time, error) { return sun.RiseTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return sun.AltitudeN(d, lon, lat, -1) }},
		{"sun/SetTime", func(d time.Time) (time.Time, error) { return sun.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return sun.Altitude(d, lon, lat) }},
		{"sun/SetTimeN", func(d time.Time) (time.Time, error) { return sun.SetTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return sun.AltitudeN(d, lon, lat, -1) }},
		{"sun/CulminationTime", func(d time.Time) (time.Time, error) { return sun.CulminationTime(d, lon), nil }, nil},
		{"sun/CulminationTimeN", func(d time.Time) (time.Time, error) { return sun.CulminationTimeN(d, lon, -1), nil }, nil},
		{"moon/RiseTime", func(d time.Time) (time.Time, error) { return moon.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return moon.Altitude(d, lon, lat) }},
		{"moon/SetTime", func(d time.Time) (time.Time, error) { return moon.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return moon.Altitude(d, lon, lat) }},
		{"moon/CulminationTime", func(d time.Time) (time.Time, error) { return moon.CulminationTime(d, lon, lat), nil }, nil},
		{"mercury/RiseTime", func(d time.Time) (time.Time, error) { return mercury.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return mercury.Altitude(d, lon, lat) }},
		{"mercury/RiseTimeN", func(d time.Time) (time.Time, error) { return mercury.RiseTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return mercury.AltitudeN(d, lon, lat, -1) }},
		{"mercury/SetTime", func(d time.Time) (time.Time, error) { return mercury.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return mercury.Altitude(d, lon, lat) }},
		{"mercury/SetTimeN", func(d time.Time) (time.Time, error) { return mercury.SetTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return mercury.AltitudeN(d, lon, lat, -1) }},
		{"mercury/CulminationTime", func(d time.Time) (time.Time, error) { return mercury.CulminationTime(d, lon), nil }, nil},
		{"mercury/CulminationTimeN", func(d time.Time) (time.Time, error) { return mercury.CulminationTimeN(d, lon, -1), nil }, nil},
		{"venus/RiseTime", func(d time.Time) (time.Time, error) { return venus.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return venus.Altitude(d, lon, lat) }},
		{"venus/RiseTimeN", func(d time.Time) (time.Time, error) { return venus.RiseTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return venus.AltitudeN(d, lon, lat, -1) }},
		{"venus/SetTime", func(d time.Time) (time.Time, error) { return venus.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return venus.Altitude(d, lon, lat) }},
		{"venus/SetTimeN", func(d time.Time) (time.Time, error) { return venus.SetTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return venus.AltitudeN(d, lon, lat, -1) }},
		{"venus/CulminationTime", func(d time.Time) (time.Time, error) { return venus.CulminationTime(d, lon), nil }, nil},
		{"venus/CulminationTimeN", func(d time.Time) (time.Time, error) { return venus.CulminationTimeN(d, lon, -1), nil }, nil},
		{"mars/RiseTime", func(d time.Time) (time.Time, error) { return mars.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return mars.Altitude(d, lon, lat) }},
		{"mars/RiseTimeN", func(d time.Time) (time.Time, error) { return mars.RiseTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return mars.AltitudeN(d, lon, lat, -1) }},
		{"mars/SetTime", func(d time.Time) (time.Time, error) { return mars.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return mars.Altitude(d, lon, lat) }},
		{"mars/SetTimeN", func(d time.Time) (time.Time, error) { return mars.SetTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return mars.AltitudeN(d, lon, lat, -1) }},
		{"mars/CulminationTime", func(d time.Time) (time.Time, error) { return mars.CulminationTime(d, lon), nil }, nil},
		{"mars/CulminationTimeN", func(d time.Time) (time.Time, error) { return mars.CulminationTimeN(d, lon, -1), nil }, nil},
		{"jupiter/RiseTime", func(d time.Time) (time.Time, error) { return jupiter.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return jupiter.Altitude(d, lon, lat) }},
		{"jupiter/RiseTimeN", func(d time.Time) (time.Time, error) { return jupiter.RiseTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return jupiter.AltitudeN(d, lon, lat, -1) }},
		{"jupiter/SetTime", func(d time.Time) (time.Time, error) { return jupiter.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return jupiter.Altitude(d, lon, lat) }},
		{"jupiter/SetTimeN", func(d time.Time) (time.Time, error) { return jupiter.SetTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return jupiter.AltitudeN(d, lon, lat, -1) }},
		{"jupiter/CulminationTime", func(d time.Time) (time.Time, error) { return jupiter.CulminationTime(d, lon), nil }, nil},
		{"jupiter/CulminationTimeN", func(d time.Time) (time.Time, error) { return jupiter.CulminationTimeN(d, lon, -1), nil }, nil},
		{"saturn/RiseTime", func(d time.Time) (time.Time, error) { return saturn.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return saturn.Altitude(d, lon, lat) }},
		{"saturn/RiseTimeN", func(d time.Time) (time.Time, error) { return saturn.RiseTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return saturn.AltitudeN(d, lon, lat, -1) }},
		{"saturn/SetTime", func(d time.Time) (time.Time, error) { return saturn.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return saturn.Altitude(d, lon, lat) }},
		{"saturn/SetTimeN", func(d time.Time) (time.Time, error) { return saturn.SetTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return saturn.AltitudeN(d, lon, lat, -1) }},
		{"saturn/CulminationTime", func(d time.Time) (time.Time, error) { return saturn.CulminationTime(d, lon), nil }, nil},
		{"saturn/CulminationTimeN", func(d time.Time) (time.Time, error) { return saturn.CulminationTimeN(d, lon, -1), nil }, nil},
		{"uranus/RiseTime", func(d time.Time) (time.Time, error) { return uranus.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return uranus.Altitude(d, lon, lat) }},
		{"uranus/RiseTimeN", func(d time.Time) (time.Time, error) { return uranus.RiseTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return uranus.AltitudeN(d, lon, lat, -1) }},
		{"uranus/SetTime", func(d time.Time) (time.Time, error) { return uranus.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return uranus.Altitude(d, lon, lat) }},
		{"uranus/SetTimeN", func(d time.Time) (time.Time, error) { return uranus.SetTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return uranus.AltitudeN(d, lon, lat, -1) }},
		{"uranus/CulminationTime", func(d time.Time) (time.Time, error) { return uranus.CulminationTime(d, lon), nil }, nil},
		{"uranus/CulminationTimeN", func(d time.Time) (time.Time, error) { return uranus.CulminationTimeN(d, lon, -1), nil }, nil},
		{"neptune/RiseTime", func(d time.Time) (time.Time, error) { return neptune.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return neptune.Altitude(d, lon, lat) }},
		{"neptune/RiseTimeN", func(d time.Time) (time.Time, error) { return neptune.RiseTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return neptune.AltitudeN(d, lon, lat, -1) }},
		{"neptune/SetTime", func(d time.Time) (time.Time, error) { return neptune.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return neptune.Altitude(d, lon, lat) }},
		{"neptune/SetTimeN", func(d time.Time) (time.Time, error) { return neptune.SetTimeN(d, lon, lat, 0, false, -1) }, func(d time.Time) float64 { return neptune.AltitudeN(d, lon, lat, -1) }},
		{"neptune/CulminationTime", func(d time.Time) (time.Time, error) { return neptune.CulminationTime(d, lon), nil }, nil},
		{"neptune/CulminationTimeN", func(d time.Time) (time.Time, error) { return neptune.CulminationTimeN(d, lon, -1), nil }, nil},
		{"litesun/RiseTime", func(d time.Time) (time.Time, error) { return litesun.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return litesun.Altitude(d, lon, lat) }},
		{"litesun/SetTime", func(d time.Time) (time.Time, error) { return litesun.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return litesun.Altitude(d, lon, lat) }},
		{"litemoon/RiseTime", func(d time.Time) (time.Time, error) { return litemoon.RiseTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return litemoon.Altitude(d, lon, lat) }},
		{"litemoon/SetTime", func(d time.Time) (time.Time, error) { return litemoon.SetTime(d, lon, lat, 0, false) }, func(d time.Time) float64 { return litemoon.Altitude(d, lon, lat) }},
		{"star/RiseTime", func(d time.Time) (time.Time, error) { return star.RiseTime(d, 100, -10, lon, lat, 0, false) }, func(d time.Time) float64 { return star.Altitude(d, 100, -10, lon, lat) }},
		{"orbit/RiseTime", func(d time.Time) (time.Time, error) { return orbit.RiseTime(d, elements, lon, lat, 0, false) }, func(d time.Time) float64 { return orbit.Altitude(d, elements, lon, lat, 0) }},
		{"star/SetTime", func(d time.Time) (time.Time, error) { return star.SetTime(d, 100, -10, lon, lat, 0, false) }, func(d time.Time) float64 { return star.Altitude(d, 100, -10, lon, lat) }},
		{"orbit/SetTime", func(d time.Time) (time.Time, error) { return orbit.SetTime(d, elements, lon, lat, 0, false) }, func(d time.Time) float64 { return orbit.Altitude(d, elements, lon, lat, 0) }},
		{"star/CulminationTime", func(d time.Time) (time.Time, error) { return star.CulminationTime(d, 100, lon), nil }, nil},
		{"orbit/CulminationTime", func(d time.Time) (time.Time, error) { return orbit.CulminationTime(d, elements, lon, lat, 0), nil }, nil},
		{"sun/MorningTwilight", func(d time.Time) (time.Time, error) { return sun.MorningTwilight(d, lon, lat, -6) }, func(d time.Time) float64 { return sun.Altitude(d, lon, lat) + 6 }},
		{"sun/MorningTwilightN", func(d time.Time) (time.Time, error) { return sun.MorningTwilightN(d, lon, lat, -6, -1) }, func(d time.Time) float64 { return sun.AltitudeN(d, lon, lat, -1) + 6 }},
		{"sun/EveningTwilight", func(d time.Time) (time.Time, error) { return sun.EveningTwilight(d, lon, lat, -6) }, func(d time.Time) float64 { return sun.Altitude(d, lon, lat) + 6 }},
		{"sun/EveningTwilightN", func(d time.Time) (time.Time, error) { return sun.EveningTwilightN(d, lon, lat, -6, -1) }, func(d time.Time) float64 { return sun.AltitudeN(d, lon, lat, -1) + 6 }},
	}
}

func TestCivilEventsIgnoreTimeOfDay(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	base := time.Date(2026, 2, 17, 0, 0, 0, 0, loc)
	for _, tc := range civilEventCases(108.93, 34.27) {
		t.Run(tc.name, func(t *testing.T) {
			want, err := tc.event(base)
			if err != nil {
				t.Fatal(err)
			}
			for _, clock := range [][3]int{{11, 59, 59}, {12, 0, 0}, {12, 30, 0}, {12, 59, 59}, {13, 0, 0}, {23, 59, 59}} {
				input := time.Date(2026, 2, 17, clock[0], clock[1], clock[2], 0, loc)
				got, err := tc.event(input)
				if err != nil || !got.Equal(want) || got.Location() != loc {
					t.Errorf("input %s: got %s, %v; want %s", input, got, err, want)
				}
			}
		})
	}
}

func TestCivilEventsAcrossDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, md := range [][2]int{{3, 8}, {11, 1}} {
		base := time.Date(2026, time.Month(md[0]), md[1], 0, 0, 0, 0, loc)
		for _, tc := range civilEventCases(-74.006, 40.7128) {
			t.Run(base.Format("2006-01-02")+"/"+tc.name, func(t *testing.T) {
				want, wantErr := tc.event(base)
				if wantErr == nil && want.Location() != loc {
					t.Fatalf("lost location: %s", want)
				}
				if wantErr == nil && tc.altitude != nil {
					if residual := tc.altitude(want); math.IsNaN(residual) || math.Abs(residual) > 0.03 {
						t.Errorf("event %s residual altitude %.6f degrees", want, residual)
					}
				}
				for _, hour := range []int{1, 3, 10, 12, 13, 23} {
					input := time.Date(base.Year(), base.Month(), base.Day(), hour, 30, 0, 0, loc)
					got, err := tc.event(input)
					if err != wantErr || !got.Equal(want) {
						t.Errorf("input %s: got %s, %v; want %s, %v", input, got, err, want, wantErr)
					}
				}
			})
		}
	}
}

func TestSunriseDSTMatchesFixedOffsetInstant(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, md := range [][2]int{{3, 8}, {11, 1}} {
		date := time.Date(2026, time.Month(md[0]), md[1], 10, 0, 0, 0, loc)
		got, err := sun.RiseTime(date, -74.006, 40.7128, 0, true)
		if err != nil {
			t.Fatal(err)
		}
		_, offset := got.Zone()
		fixed := time.FixedZone("fixed", offset)
		want, err := sun.RiseTime(time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, fixed), -74.006, 40.7128, 0, true)
		if err != nil || math.Abs(got.Sub(want).Seconds()) > 0.001 {
			t.Errorf("%s: got %s, want %s, err=%v", date, got, want, err)
		}
	}
}

func TestMoonriseInLastHourOfLongCivilDay(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	date := time.Date(2026, 11, 1, 12, 30, 0, 0, loc)
	for _, tc := range []struct {
		name string
		rise func(time.Time, float64, float64, float64, bool) (time.Time, error)
	}{
		{"moon", moon.RiseTime}, {"lite/moon", litemoon.RiseTime},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.rise(date, -74.006, 40.7128, 0, false)
			if err != nil {
				t.Fatal(err)
			}
			if got.Day() != 1 || got.Hour() != 23 || got.Location() != loc {
				t.Fatalf("expected late local moonrise, got %s", got)
			}
			fixed := time.FixedZone("EST", -5*3600)
			want, err := tc.rise(time.Date(2026, 11, 1, 0, 0, 0, 0, fixed), -74.006, 40.7128, 0, false)
			if err != nil || !got.Equal(want) {
				t.Fatalf("got %s, want %s, err=%v", got, want, err)
			}
		})
	}
}

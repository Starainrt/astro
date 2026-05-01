package basic_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestBasicSunObservationNFullMatchesDefault(t *testing.T) {
	date := time.Date(2026, 4, 26, 9, 30, 45, 123456789, time.FixedZone("CST", 8*3600))
	ttJD := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	jde := basic.Date2JDE(date)
	lon := 116.391
	lat := 39.907
	tz := 8.0
	height := 45.0
	twilightAngle := -6.0

	assertSame := func(name string, got, want float64) {
		t.Helper()
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("%s full-n mismatch", name)
		}
	}
	assertSameEvent := func(name string, got float64, gotErr error, want float64, wantErr error) {
		t.Helper()
		switch {
		case gotErr == nil && wantErr == nil:
			assertSame(name, got, want)
		case gotErr == nil || wantErr == nil:
			t.Fatalf("%s full-n mismatch", name)
		case gotErr != wantErr:
			t.Fatalf("%s full-n mismatch", name)
		}
	}

	assertSame("SunTime", basic.SunTime(ttJD), basic.SunTimeN(ttJD, -1))
	assertSame("CulminationTime", basic.CulminationTime(jde, lon, tz), basic.CulminationTimeN(jde, lon, tz, -1))
	assertSame("SunTimeAngle", basic.SunTimeAngle(jde, lon, lat, tz), basic.SunTimeAngleN(jde, lon, lat, tz, -1))
	assertSame("SunHeight", basic.SunHeight(jde, lon, lat, tz), basic.SunHeightN(jde, lon, lat, tz, -1))
	assertSame("SunAzimuth", basic.SunAzimuth(jde, lon, lat, tz), basic.SunAzimuthN(jde, lon, lat, tz, -1))
	rise1, riseErr1 := basic.GetSunRiseTime(jde, lon, lat, tz, 1, height)
	rise2, riseErr2 := basic.GetSunRiseTimeN(jde, lon, lat, tz, 1, height, -1)
	assertSameEvent("GetSunRiseTime", rise1, riseErr1, rise2, riseErr2)
	set1, setErr1 := basic.GetSunSetTime(jde, lon, lat, tz, 1, height)
	set2, setErr2 := basic.GetSunSetTimeN(jde, lon, lat, tz, 1, height, -1)
	assertSameEvent("GetSunSetTime", set1, setErr1, set2, setErr2)
	morning1, morningErr1 := basic.MorningTwilight(jde, lon, lat, tz, twilightAngle)
	morning2, morningErr2 := basic.MorningTwilightN(jde, lon, lat, tz, twilightAngle, -1)
	assertSameEvent("MorningTwilight", morning1, morningErr1, morning2, morningErr2)
	evening1, eveningErr1 := basic.EveningTwilight(jde, lon, lat, tz, twilightAngle)
	evening2, eveningErr2 := basic.EveningTwilightN(jde, lon, lat, tz, twilightAngle, -1)
	assertSameEvent("EveningTwilight", evening1, eveningErr1, evening2, eveningErr2)
}

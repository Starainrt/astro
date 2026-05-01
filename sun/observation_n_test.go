package sun

import (
	"math"
	"testing"
	"time"
)

func TestObservationNFullMatchesDefault(t *testing.T) {
	date := time.Date(2026, 4, 26, 9, 30, 45, 123456789, time.FixedZone("CST", 8*3600))
	lon := 116.391
	lat := 39.907
	height := 45.0
	twilightAngle := -6.0

	assertSame := func(name string, got, want float64) {
		t.Helper()
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("%s full-n mismatch", name)
		}
	}
	assertTimeSame := func(name string, got, want time.Time) {
		t.Helper()
		if got.UnixNano() != want.UnixNano() || got.Location().String() != want.Location().String() {
			t.Fatalf("%s full-n mismatch", name)
		}
	}
	assertErrSame := func(name string, got, want error) {
		t.Helper()
		switch {
		case got == nil && want == nil:
			return
		case got == nil || want == nil:
			t.Fatalf("%s full-n mismatch", name)
		case got.Error() != want.Error():
			t.Fatalf("%s full-n mismatch", name)
		}
	}

	assertSame("HourAngle", HourAngle(date, lon, lat), HourAngleN(date, lon, lat, -1))
	assertSame("ParallacticAngle", ParallacticAngle(date, lon, lat), ParallacticAngleN(date, lon, lat, -1))
	assertSame("Azimuth", Azimuth(date, lon, lat), AzimuthN(date, lon, lat, -1))
	assertSame("Altitude", Altitude(date, lon, lat), AltitudeN(date, lon, lat, -1))
	assertSame("Zenith", Zenith(date, lon, lat), ZenithN(date, lon, lat, -1))
	if !math.IsNaN(Altitude(date, lon, lat)) && math.Abs((Altitude(date, lon, lat)+Zenith(date, lon, lat))-90) > 1e-12 {
		t.Fatal("sun altitude + zenith should equal 90 degrees")
	}
	assertTimeSame("CulminationTime", CulminationTime(date, lon), CulminationTimeN(date, lon, -1))
	assertTimeSame("ApparentSolarTime", ApparentSolarTime(date, lon), ApparentSolarTimeN(date, lon, -1))

	rise1, err1 := RiseTime(date, lon, lat, height, true)
	rise2, err2 := RiseTimeN(date, lon, lat, height, true, -1)
	assertTimeSame("RiseTime", rise1, rise2)
	assertErrSame("RiseTime.err", err1, err2)

	set1, err1 := SetTime(date, lon, lat, height, true)
	set2, err2 := SetTimeN(date, lon, lat, height, true, -1)
	assertTimeSame("SetTime", set1, set2)
	assertErrSame("SetTime.err", err1, err2)

	down1, err1 := DownTime(date, lon, lat, height, true)
	down2, err2 := DownTimeN(date, lon, lat, height, true, -1)
	assertTimeSame("DownTime", down1, down2)
	assertErrSame("DownTime.err", err1, err2)

	morning1, err1 := MorningTwilight(date, lon, lat, twilightAngle)
	morning2, err2 := MorningTwilightN(date, lon, lat, twilightAngle, -1)
	assertTimeSame("MorningTwilight", morning1, morning2)
	assertErrSame("MorningTwilight.err", err1, err2)

	evening1, err1 := EveningTwilight(date, lon, lat, twilightAngle)
	evening2, err2 := EveningTwilightN(date, lon, lat, twilightAngle, -1)
	assertTimeSame("EveningTwilight", evening1, evening2)
	assertErrSame("EveningTwilight.err", err1, err2)
}

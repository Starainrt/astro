package uranus

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

	assertSame := func(name string, got, want float64) {
		t.Helper()
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("%s full-n mismatch", name)
		}
	}
	assertSamePair := func(name string, got1, got2, want1, want2 float64) {
		t.Helper()
		assertSame(name+".1", got1, want1)
		assertSame(name+".2", got2, want2)
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

	floatChecks := []struct {
		name string
		got  func() float64
		want func() float64
	}{
		{"ApparentLo", func() float64 { return ApparentLo(date) }, func() float64 { return ApparentLoN(date, -1) }},
		{"ApparentBo", func() float64 { return ApparentBo(date) }, func() float64 { return ApparentBoN(date, -1) }},
		{"ApparentRa", func() float64 { return ApparentRa(date) }, func() float64 { return ApparentRaN(date, -1) }},
		{"ApparentDec", func() float64 { return ApparentDec(date) }, func() float64 { return ApparentDecN(date, -1) }},
		{"ApparentMagnitude", func() float64 { return ApparentMagnitude(date) }, func() float64 { return ApparentMagnitudeN(date, -1) }},
		{"PhaseAngle", func() float64 { return PhaseAngle(date) }, func() float64 { return PhaseAngleN(date, -1) }},
		{"IlluminatedFraction", func() float64 { return IlluminatedFraction(date) }, func() float64 { return IlluminatedFractionN(date, -1) }},
		{"Phase", func() float64 { return Phase(date) }, func() float64 { return PhaseN(date, -1) }},
		{"BrightLimbPositionAngle", func() float64 { return BrightLimbPositionAngle(date) }, func() float64 { return BrightLimbPositionAngleN(date, -1) }},
		{"EarthDistance", func() float64 { return EarthDistance(date) }, func() float64 { return EarthDistanceN(date, -1) }},
		{"SunDistance", func() float64 { return SunDistance(date) }, func() float64 { return SunDistanceN(date, -1) }},
		{"Altitude", func() float64 { return Altitude(date, lon, lat) }, func() float64 { return AltitudeN(date, lon, lat, -1) }},
		{"Zenith", func() float64 { return Zenith(date, lon, lat) }, func() float64 { return ZenithN(date, lon, lat, -1) }},
		{"Azimuth", func() float64 { return Azimuth(date, lon, lat) }, func() float64 { return AzimuthN(date, lon, lat, -1) }},
		{"HourAngle", func() float64 { return HourAngle(date, lon) }, func() float64 { return HourAngleN(date, lon, -1) }},
		{"ParallacticAngle", func() float64 { return ParallacticAngle(date, lon, lat) }, func() float64 { return ParallacticAngleN(date, lon, lat, -1) }},
	}
	for _, tc := range floatChecks {
		assertSame(tc.name, tc.got(), tc.want())
	}

	if math.Abs((Altitude(date, lon, lat)+Zenith(date, lon, lat))-90) > 1e-12 {
		t.Fatal("altitude + zenith should equal 90 degrees")
	}

	gotRa, gotDec := ApparentRaDec(date)
	wantRa, wantDec := ApparentRaDecN(date, -1)
	assertSamePair("ApparentRaDec", gotRa, gotDec, wantRa, wantDec)

	assertTimeSame("CulminationTime", CulminationTime(date, lon), CulminationTimeN(date, lon, -1))

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
}

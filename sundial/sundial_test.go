package sundial

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/sun"
)

func TestTrueSolarTimeMatchesSunPackage(t *testing.T) {
	date := time.Date(2026, 6, 21, 15, 30, 45, 123000000, time.FixedZone("CST", 8*3600))
	lon := 116.3913

	got := TrueSolarTime(date, lon)
	want := sun.ApparentSolarTime(date, lon)
	if !got.Equal(want) {
		t.Fatalf("true solar time mismatch: got %s want %s", got, want)
	}
}

func TestMeanSolarTimeMatchesSyntheticLongitudeZone(t *testing.T) {
	date := time.Date(2026, 6, 21, 15, 30, 45, 123000000, time.FixedZone("CST", 8*3600))
	lon := 116.3913

	got := MeanSolarTime(date, lon)
	want := date.In(time.FixedZone("LTZ", int(lon*3600.0/15.0)))
	if math.Abs(got.Sub(want).Seconds()) > 1 {
		t.Fatalf("mean solar time mismatch: got %s want %s", got, want)
	}
}

func TestEquationTimeMatchesTrueMinusMeanSolarTime(t *testing.T) {
	date := time.Date(2026, 6, 21, 15, 30, 45, 123000000, time.FixedZone("CST", 8*3600))
	lon := 116.3913

	apparent := clockHours(TrueSolarTime(date, lon))
	mean := clockHours(MeanSolarTime(date, lon))
	diff := apparent - mean
	want := sun.EquationTime(date)
	if math.Abs(diff-want) > 1e-9 {
		t.Fatalf("equation-time mismatch: got %.12f want %.12f", diff, want)
	}
}

func TestHourAngleMatchesTrueSolarClockFace(t *testing.T) {
	date := time.Date(2026, 6, 21, 15, 30, 45, 123000000, time.FixedZone("CST", 8*3600))
	lon := 116.3913

	hourAngle := HourAngle(date, lon)
	trueSolar := TrueSolarTime(date, lon)
	trueHours := float64(trueSolar.Hour()) +
		float64(trueSolar.Minute())/60 +
		float64(trueSolar.Second())/3600 +
		float64(trueSolar.Nanosecond())/3.6e12
	want := normalizeSigned180((trueHours - 12) * 15)

	if math.Abs(hourAngle-want) > 0.02 {
		t.Fatalf("hour angle mismatch: got %.6f want %.6f", hourAngle, want)
	}
}

func TestMeanSolarHourAngleMatchesInstantChain(t *testing.T) {
	date := time.Date(2026, 6, 21, 15, 30, 45, 123000000, time.FixedZone("CST", 8*3600))
	lon := 116.3913

	meanDate := MeanSolarTime(date, lon)
	meanSolarHours := clockHours(meanDate)
	got := MeanSolarHourAngle(meanDate, meanSolarHours)
	want := HourAngle(meanDate, longitudeFromTimeZone(meanDate))
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("mean-solar hour angle mismatch: got %.12f want %.12f", got, want)
	}
}

func TestZoneTimeHourAngleRebuildsTargetClockInstant(t *testing.T) {
	date := time.Date(2026, 6, 21, 15, 30, 45, 123000000, time.FixedZone("CST", 8*3600))
	lon := 116.3913
	zoneTimeHours := 9.5

	got := ZoneTimeHourAngle(date, lon, zoneTimeHours)
	want := HourAngle(dateWithClockHours(date, zoneTimeHours), lon)
	if math.Abs(got-want) > 1e-7 {
		t.Fatalf("zone-time hour angle mismatch: got %.12f want %.12f", got, want)
	}
}

func TestHorizontalHourLineAngleKnownValues(t *testing.T) {
	if math.Abs(HorizontalHourLineAngle(0, -75)) > 1e-12 {
		t.Fatalf("equatorial horizontal sundial should keep all hour lines on the noon line")
	}

	got := HorizontalHourLineAngle(45, 45)
	want := math.Atan(math.Sin(45*math.Pi/180)*math.Tan(45*math.Pi/180)) * 180 / math.Pi
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("known-value mismatch: got %.12f want %.12f", got, want)
	}

	if math.Abs(HorizontalHourLineAngle(45, -45)+got) > 1e-12 {
		t.Fatalf("morning/afternoon symmetry mismatch")
	}
}

func TestHorizontalHourLineAngleAtMatchesHourAngleChain(t *testing.T) {
	date := time.Date(2026, 6, 21, 15, 30, 45, 123000000, time.FixedZone("CST", 8*3600))
	lon := 116.3913
	lat := 39.9042

	got := HorizontalHourLineAngleAt(date, lon, lat)
	want := HorizontalHourLineAngle(lat, HourAngle(date, lon))
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("hour-line chain mismatch: got %.12f want %.12f", got, want)
	}
}

func TestInvalidHourLineInputReturnsNaN(t *testing.T) {
	if !math.IsNaN(HorizontalHourLineAngle(math.NaN(), 30)) {
		t.Fatalf("NaN latitude should produce NaN result")
	}
	if !math.IsNaN(HorizontalHourLineAngle(30, math.Inf(1))) {
		t.Fatalf("Inf hour angle should produce NaN result")
	}
	if !math.IsNaN(MeanSolarHourAngle(time.Now(), math.NaN())) {
		t.Fatalf("NaN mean-solar time should produce NaN result")
	}
	if !math.IsNaN(ZoneTimeHourAngle(time.Now(), math.Inf(1), 12)) {
		t.Fatalf("Inf longitude should produce NaN result")
	}
}

package astro_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/mars"
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/neptune"
	"github.com/starainrt/astro/saturn"
	"github.com/starainrt/astro/star"
	"github.com/starainrt/astro/sun"
	"github.com/starainrt/astro/uranus"
	"github.com/starainrt/astro/venus"
)

func nearlyEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-12
}

func TestPlanetAbsoluteQuantitiesIgnoreInputTimezone(t *testing.T) {
	utc := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	cst := time.FixedZone("CST", 8*3600)
	local := utc.In(cst)

	scalars := []struct {
		name string
		fn   func(time.Time) float64
	}{
		{"mercury.ApparentLo", mercury.ApparentLo},
		{"mercury.ApparentBo", mercury.ApparentBo},
		{"mercury.ApparentRa", mercury.ApparentRa},
		{"mercury.ApparentDec", mercury.ApparentDec},
		{"mercury.ApparentMagnitude", mercury.ApparentMagnitude},
		{"mercury.EarthDistance", mercury.EarthDistance},
		{"mercury.SunDistance", mercury.SunDistance},
		{"venus.ApparentLo", venus.ApparentLo},
		{"venus.ApparentBo", venus.ApparentBo},
		{"venus.ApparentRa", venus.ApparentRa},
		{"venus.ApparentDec", venus.ApparentDec},
		{"venus.ApparentMagnitude", venus.ApparentMagnitude},
		{"venus.EarthDistance", venus.EarthDistance},
		{"venus.SunDistance", venus.SunDistance},
		{"mars.ApparentLo", mars.ApparentLo},
		{"mars.ApparentBo", mars.ApparentBo},
		{"mars.ApparentRa", mars.ApparentRa},
		{"mars.ApparentDec", mars.ApparentDec},
		{"mars.ApparentMagnitude", mars.ApparentMagnitude},
		{"mars.EarthDistance", mars.EarthDistance},
		{"mars.SunDistance", mars.SunDistance},
		{"jupiter.ApparentLo", jupiter.ApparentLo},
		{"jupiter.ApparentBo", jupiter.ApparentBo},
		{"jupiter.ApparentRa", jupiter.ApparentRa},
		{"jupiter.ApparentDec", jupiter.ApparentDec},
		{"jupiter.ApparentMagnitude", jupiter.ApparentMagnitude},
		{"jupiter.EarthDistance", jupiter.EarthDistance},
		{"jupiter.SunDistance", jupiter.SunDistance},
		{"saturn.ApparentLo", saturn.ApparentLo},
		{"saturn.ApparentBo", saturn.ApparentBo},
		{"saturn.ApparentRa", saturn.ApparentRa},
		{"saturn.ApparentDec", saturn.ApparentDec},
		{"saturn.ApparentMagnitude", saturn.ApparentMagnitude},
		{"saturn.EarthDistance", saturn.EarthDistance},
		{"saturn.SunDistance", saturn.SunDistance},
		{"uranus.ApparentLo", uranus.ApparentLo},
		{"uranus.ApparentBo", uranus.ApparentBo},
		{"uranus.ApparentRa", uranus.ApparentRa},
		{"uranus.ApparentDec", uranus.ApparentDec},
		{"uranus.ApparentMagnitude", uranus.ApparentMagnitude},
		{"uranus.EarthDistance", uranus.EarthDistance},
		{"uranus.SunDistance", uranus.SunDistance},
		{"neptune.ApparentLo", neptune.ApparentLo},
		{"neptune.ApparentBo", neptune.ApparentBo},
		{"neptune.ApparentRa", neptune.ApparentRa},
		{"neptune.ApparentDec", neptune.ApparentDec},
		{"neptune.ApparentMagnitude", neptune.ApparentMagnitude},
		{"neptune.EarthDistance", neptune.EarthDistance},
		{"neptune.SunDistance", neptune.SunDistance},
	}

	for _, tc := range scalars {
		if !nearlyEqual(tc.fn(utc), tc.fn(local)) {
			t.Fatalf("%s should depend on absolute time only", tc.name)
		}
	}

	pairs := []struct {
		name string
		fn   func(time.Time) (float64, float64)
	}{
		{"mercury.ApparentRaDec", mercury.ApparentRaDec},
		{"venus.ApparentRaDec", venus.ApparentRaDec},
		{"mars.ApparentRaDec", mars.ApparentRaDec},
		{"jupiter.ApparentRaDec", jupiter.ApparentRaDec},
		{"saturn.ApparentRaDec", saturn.ApparentRaDec},
		{"uranus.ApparentRaDec", uranus.ApparentRaDec},
		{"neptune.ApparentRaDec", neptune.ApparentRaDec},
	}

	for _, tc := range pairs {
		leftA, leftB := tc.fn(utc)
		rightA, rightB := tc.fn(local)
		if !nearlyEqual(leftA, rightA) || !nearlyEqual(leftB, rightB) {
			t.Fatalf("%s should depend on absolute time only", tc.name)
		}
	}
}

func TestJDECalcRejectsGregorianGap(t *testing.T) {
	cases := []float64{5, 6.5, 10, 14.25}
	for _, day := range cases {
		got := basic.JDECalc(1582, 10, day)
		if !math.IsNaN(got) {
			t.Fatalf("1582-10-%v should be rejected, got %.15f", day, got)
		}
	}

	before := basic.JDECalc(1582, 10, 4)
	after := basic.JDECalc(1582, 10, 15)
	if math.IsNaN(before) || math.IsNaN(after) {
		t.Fatal("boundary dates around Gregorian reform should remain valid")
	}
	if !nearlyEqual(after-before, 1) {
		t.Fatalf("1582-10-15 should remain the civil day after 1582-10-04")
	}
}

func TestCalendarAddPreservesOriginalTimezone(t *testing.T) {
	oldLocal := time.Local
	time.Local = time.UTC
	defer func() {
		time.Local = oldLocal
	}()

	tz := time.FixedZone("CST", 8*3600)
	start := time.Date(1985, 1, 21, 9, 30, 0, 0, tz)

	lunar, err := calendar.SolarToLunar(start)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := calendar.SolarToLunar(lunar.Time().Add(36 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	shifted := lunar.Add(36 * time.Hour).Time()
	if delta := shifted.Sub(expected.Time()); delta < -time.Millisecond || delta > time.Millisecond {
		t.Fatalf("calendar.Time.Add should not depend on time.Local: got %v want %v", shifted, expected.Time())
	}
}

func TestObservationZenithMatchesIndependentFormula(t *testing.T) {
	places := []struct {
		name     string
		lon, lat float64
	}{
		{"beijing", 116.391, 39.907},
		{"sydney", 151.2093, -33.8688},
		{"tromso", 18.9553, 69.6492},
	}
	dates := []time.Time{
		time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC),
		time.Date(2026, 4, 26, 9, 30, 45, 123456789, time.FixedZone("CST", 8*3600)),
		time.Date(2100, 6, 15, 3, 4, 5, 0, time.UTC),
	}
	starRa := 6.752477
	starDec := -16.716116

	// 容差取实测最大偏差的 5 倍以上；太阳还含视位置的入口差异，月光低精度级数与高精度级数本身不同源。
	const (
		sunTolerance     = 2e-3
		moonTolerance    = 5e-8
		moonLowTolerance = 3e-3
		starTolerance    = 1e-12
		planetTolerance  = 5e-9
	)

	for _, date := range dates {
		for _, place := range places {
			jde := basic.Date2JDE(date)
			_, loc := date.Zone()
			timezone := float64(loc) / 3600.0
			tt := basic.TD2UT(jde-timezone/24, true)

			checks := []struct {
				name    string
				tol     float64
				zenith  float64
				witness float64
			}{
				{"sun", sunTolerance,
					sun.Zenith(date, place.lon, place.lat),
					zenithFromHourAngle(sun.HourAngle(date, place.lon, place.lat), basic.HSunApparentDec(tt), place.lat)},
				{"moon", moonTolerance,
					moon.Zenith(date, place.lon, place.lat),
					zenithFromHourAngle(moon.HourAngle(date, place.lon, place.lat), moon.ApparentDec(date, place.lon, place.lat), place.lat)},
				{"moon-low-precision-series", moonLowTolerance,
					moon.Zenith(date, place.lon, place.lat),
					90 - basic.MoonHeight(jde, place.lon, place.lat, timezone)},
				{"star", starTolerance,
					star.Zenith(date, starRa, starDec, place.lon, place.lat),
					zenithFromHourAngle(star.HourAngle(date, starRa, place.lon), starDec, place.lat)},
				{"mercury", planetTolerance,
					mercury.Zenith(date, place.lon, place.lat),
					zenithFromHourAngle(mercury.HourAngle(date, place.lon), mercury.ApparentDec(date), place.lat)},
				{"venus", planetTolerance,
					venus.Zenith(date, place.lon, place.lat),
					zenithFromHourAngle(venus.HourAngle(date, place.lon), venus.ApparentDec(date), place.lat)},
				{"mars", planetTolerance,
					mars.Zenith(date, place.lon, place.lat),
					zenithFromHourAngle(mars.HourAngle(date, place.lon), mars.ApparentDec(date), place.lat)},
				{"jupiter", planetTolerance,
					jupiter.Zenith(date, place.lon, place.lat),
					zenithFromHourAngle(jupiter.HourAngle(date, place.lon), jupiter.ApparentDec(date), place.lat)},
				{"saturn", planetTolerance,
					saturn.Zenith(date, place.lon, place.lat),
					zenithFromHourAngle(saturn.HourAngle(date, place.lon), saturn.ApparentDec(date), place.lat)},
				{"uranus", planetTolerance,
					uranus.Zenith(date, place.lon, place.lat),
					zenithFromHourAngle(uranus.HourAngle(date, place.lon), uranus.ApparentDec(date), place.lat)},
				{"neptune", planetTolerance,
					neptune.Zenith(date, place.lon, place.lat),
					zenithFromHourAngle(neptune.HourAngle(date, place.lon), neptune.ApparentDec(date), place.lat)},
			}

			for _, tc := range checks {
				if delta := math.Abs(tc.zenith - tc.witness); delta > tc.tol {
					t.Fatalf("%s %s at %s: zenith %.9f, independent formula %.9f, delta %.3g > %.3g",
						place.name, tc.name, date.Format(time.RFC3339), tc.zenith, tc.witness, delta, tc.tol)
				}
			}
		}
	}
}

// zenithFromHourAngle 由时角与赤纬按 cos z = sinφ·sinδ + cosφ·cosδ·cos H 独立求天顶距，单位度。
func zenithFromHourAngle(hourAngle, dec, lat float64) float64 {
	rad := math.Pi / 180
	sinZenith := math.Sin(lat*rad)*math.Sin(dec*rad) + math.Cos(dec*rad)*math.Cos(lat*rad)*math.Cos(hourAngle*rad)
	if sinZenith > 1 {
		sinZenith = 1
	}
	if sinZenith < -1 {
		sinZenith = -1
	}
	return math.Acos(sinZenith) / rad
}

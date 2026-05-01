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

func TestObservationZenithSemantics(t *testing.T) {
	date := time.Date(2026, 4, 26, 9, 30, 45, 123456789, time.FixedZone("CST", 8*3600))
	lon := 116.391
	lat := 39.907
	ra := 6.752477
	dec := -16.716116

	checks := []struct {
		name     string
		altitude func() float64
		zenith   func() float64
	}{
		{"sun", func() float64 { return sun.Altitude(date, lon, lat) }, func() float64 { return sun.Zenith(date, lon, lat) }},
		{"moon", func() float64 { return moon.Altitude(date, lon, lat) }, func() float64 { return moon.Zenith(date, lon, lat) }},
		{"star", func() float64 { return star.Altitude(date, ra, dec, lon, lat) }, func() float64 { return star.Zenith(date, ra, dec, lon, lat) }},
		{"mercury", func() float64 { return mercury.Altitude(date, lon, lat) }, func() float64 { return mercury.Zenith(date, lon, lat) }},
		{"venus", func() float64 { return venus.Altitude(date, lon, lat) }, func() float64 { return venus.Zenith(date, lon, lat) }},
		{"mars", func() float64 { return mars.Altitude(date, lon, lat) }, func() float64 { return mars.Zenith(date, lon, lat) }},
		{"jupiter", func() float64 { return jupiter.Altitude(date, lon, lat) }, func() float64 { return jupiter.Zenith(date, lon, lat) }},
		{"saturn", func() float64 { return saturn.Altitude(date, lon, lat) }, func() float64 { return saturn.Zenith(date, lon, lat) }},
		{"uranus", func() float64 { return uranus.Altitude(date, lon, lat) }, func() float64 { return uranus.Zenith(date, lon, lat) }},
		{"neptune", func() float64 { return neptune.Altitude(date, lon, lat) }, func() float64 { return neptune.Zenith(date, lon, lat) }},
	}

	for _, tc := range checks {
		if !nearlyEqual(tc.zenith(), 90-tc.altitude()) {
			t.Fatalf("%s zenith should equal 90-altitude", tc.name)
		}
	}
}

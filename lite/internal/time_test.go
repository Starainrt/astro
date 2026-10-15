package internal

import (
	"math"
	"testing"
	"time"
)

func TestDateJulianRoundTrip(t *testing.T) {
	cases := []struct {
		date time.Time
		jd   float64
	}{
		{time.Date(2026, 10, 6, 8, 30, 15, 123000000, time.UTC), 2461319.8543417014},
		{time.Date(1582, 10, 4, 0, 0, 0, 0, time.UTC), 2299159.5},
		{time.Date(1582, 10, 15, 0, 0, 0, 0, time.UTC), 2299160.5},
	}
	for _, tc := range cases {
		if got := Date2JD(tc.date); math.Abs(got-tc.jd) > 1e-9 {
			t.Fatalf("Date2JD(%s) = %.12f, want %.12f", tc.date, got, tc.jd)
		}
	}
}

func TestJD2DateByZone(t *testing.T) {
	zone := time.FixedZone("CST", 8*3600)
	got := JD2DateByZone(2461320.5, zone, false)
	want := time.Date(2026, 10, 7, 8, 0, 0, 0, zone)
	if !got.Equal(want) {
		t.Fatalf("JD2DateByZone = %s, want %s", got, want)
	}
}

func TestUTC2UT1IgnoresDUT1(t *testing.T) {
	jd := 2461320.5
	if got := UTC2UT1(jd); got != jd {
		t.Fatalf("UTC2UT1 = %.12f, want %.12f", got, jd)
	}
}

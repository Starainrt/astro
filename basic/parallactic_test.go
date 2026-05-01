package basic

import (
	"math"
	"testing"
	"time"
)

func TestParallacticAngleByHourAngleKnownCases(t *testing.T) {
	cases := []struct {
		name      string
		hourAngle float64
		dec       float64
		lat       float64
		want      float64
	}{
		{name: "meridian", hourAngle: 0, dec: 10, lat: 45, want: 0},
		{name: "equator west", hourAngle: 30, dec: 0, lat: 0, want: 90},
		{name: "equator east", hourAngle: -30, dec: 0, lat: 0, want: -90},
	}

	for _, tc := range cases {
		got := ParallacticAngleByHourAngle(tc.hourAngle, tc.dec, tc.lat)
		if math.Abs(got-tc.want) > 1e-12 {
			t.Fatalf("%s mismatch: got %.15f want %.15f", tc.name, got, tc.want)
		}
	}
}

func TestStarParallacticAngleMatchesHourAngleForm(t *testing.T) {
	date := time.Date(2026, 4, 29, 21, 15, 0, 0, time.FixedZone("CST", 8*3600))
	jde := Date2JDE(date)
	_, offsetSeconds := date.Zone()
	timezone := float64(offsetSeconds) / 3600.0
	ra := 101.28715533
	dec := -16.71611586
	lon := 115.0
	lat := 40.0

	got := StarParallacticAngle(jde, ra, dec, lon, lat, timezone)
	want := ParallacticAngleByHourAngle(StarHourAngle(jde, ra, lon, timezone), dec, lat)
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("star parallactic angle mismatch: got %.15f want %.15f", got, want)
	}
}

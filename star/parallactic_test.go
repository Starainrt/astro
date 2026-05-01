package star

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestParallacticAngleMatchesHourAngleForm(t *testing.T) {
	date := time.Date(2026, 4, 29, 21, 15, 0, 0, time.FixedZone("CST", 8*3600))
	ra := 101.28715533
	dec := -16.71611586
	lon := 115.0
	lat := 40.0

	got := ParallacticAngle(date, ra, dec, lon, lat)
	want := basic.ParallacticAngleByHourAngle(HourAngle(date, ra, lon), dec, lat)
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("parallactic angle mismatch: got %.15f want %.15f", got, want)
	}
}

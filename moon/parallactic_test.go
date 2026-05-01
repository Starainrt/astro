package moon

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestParallacticAngleMatchesHourAngleForm(t *testing.T) {
	date := time.Date(2026, 4, 29, 21, 15, 0, 0, time.FixedZone("CST", 8*3600))
	lon := 116.391
	lat := 39.907

	got := ParallacticAngle(date, lon, lat)
	want := basic.ParallacticAngleByHourAngle(HourAngle(date, lon, lat), ApparentDec(date, lon, lat), lat)
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("parallactic angle mismatch: got %.15f want %.15f", got, want)
	}
}

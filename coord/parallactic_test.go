package coord

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestParallacticAngleWrappers(t *testing.T) {
	date := time.Date(2026, 4, 29, 13, 15, 0, 0, time.UTC)
	ra := 101.28715533
	dec := -16.71611586
	observerLon := 115.0
	observerLat := 40.0

	got := ParallacticAngle(date, ra, dec, observerLon, observerLat)
	want := basic.ParallacticAngleByHourAngle(HourAngle(date, ra, observerLon), dec, observerLat)
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("parallactic angle mismatch: got %.15f want %.15f", got, want)
	}

	if direct := ParallacticAngleByHourAngle(30, 0, 0); math.Abs(direct-90) > 1e-12 {
		t.Fatalf("direct formula mismatch: got %.15f want %.15f", direct, 90.0)
	}
}

package star

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestApparentAltitudeWrappers(t *testing.T) {
	date := time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)
	pressureHPa := 1010.0
	temperatureC := 10.0

	trueAltitude := Altitude(date, 101.28715533, -16.71611586, 115, 40)
	got := ApparentAltitude(date, 101.28715533, -16.71611586, 115, 40, pressureHPa, temperatureC)
	want := basic.ApparentAltitude(trueAltitude, pressureHPa, temperatureC)
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("ApparentAltitude mismatch: got %.18f want %.18f", got, want)
	}

	gotZenith := ApparentZenith(date, 101.28715533, -16.71611586, 115, 40, pressureHPa, temperatureC)
	if math.Abs(gotZenith-(90-got)) > 1e-12 {
		t.Fatalf("ApparentZenith mismatch: got %.18f want %.18f", gotZenith, 90-got)
	}
}

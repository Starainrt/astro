package saturn

import (
	"math"
	"testing"
	"time"
)

func TestRingNFullMatchesDefault(t *testing.T) {
	date := time.Date(2026, 4, 26, 9, 30, 45, 123456789, time.FixedZone("CST", 8*3600))
	got := Ring(date)
	want := RingN(date, -1)

	assertSameRingFloat(t, "EarthLatitude", got.EarthLatitude, want.EarthLatitude)
	assertSameRingFloat(t, "SunLatitude", got.SunLatitude, want.SunLatitude)
	assertSameRingFloat(t, "PositionAngle", got.PositionAngle, want.PositionAngle)
	assertSameRingFloat(t, "DeltaU", got.DeltaU, want.DeltaU)
	assertSameRingFloat(t, "MajorAxis", got.MajorAxis, want.MajorAxis)
	assertSameRingFloat(t, "MinorAxis", got.MinorAxis, want.MinorAxis)
}

func assertSameRingFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("%s full-n mismatch: got %.18f want %.18f", name, got, want)
	}
}

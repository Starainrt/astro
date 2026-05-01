package moon

import (
	"math"
	"testing"
	"time"
)

func TestPhysicalPreservesInstantAcrossTimezones(t *testing.T) {
	utc := time.Date(2026, 4, 28, 9, 30, 45, 123000000, time.UTC)
	shanghai := utc.In(time.FixedZone("UTC+8", 8*3600))
	got := Physical(shanghai)
	want := Physical(utc)
	valuesGot := []float64{got.OpticalLongitude, got.OpticalLatitude, got.PhysicalLongitude, got.PhysicalLatitude, got.LibrationLongitude, got.LibrationLatitude, got.PositionAngle}
	valuesWant := []float64{want.OpticalLongitude, want.OpticalLatitude, want.PhysicalLongitude, want.PhysicalLatitude, want.LibrationLongitude, want.LibrationLatitude, want.PositionAngle}
	for i := range valuesGot {
		if math.Float64bits(valuesGot[i]) != math.Float64bits(valuesWant[i]) {
			t.Fatalf("timezone instant mismatch at index %d: got %.18f want %.18f", i, valuesGot[i], valuesWant[i])
		}
	}
}

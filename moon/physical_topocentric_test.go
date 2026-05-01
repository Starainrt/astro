package moon

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestTopocentricPhysicalWrapperMatchesBasic(t *testing.T) {
	date := time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)
	jd := observationTT(date)
	observerLon := 121.4737
	observerLat := 31.2304
	height := 4.0

	got := TopocentricPhysical(date, observerLon, observerLat, height)
	want := basic.MoonTopocentricPhysicalN(jd, observerLon, observerLat, height, -1)

	assertMoonPhysicalSameFloat(t, "OpticalLongitude", got.OpticalLongitude, want.OpticalLongitude)
	assertMoonPhysicalSameFloat(t, "OpticalLatitude", got.OpticalLatitude, want.OpticalLatitude)
	assertMoonPhysicalSameFloat(t, "PhysicalLongitude", got.PhysicalLongitude, want.PhysicalLongitude)
	assertMoonPhysicalSameFloat(t, "PhysicalLatitude", got.PhysicalLatitude, want.PhysicalLatitude)
	assertMoonPhysicalSameFloat(t, "LibrationLongitude", got.LibrationLongitude, want.LibrationLongitude)
	assertMoonPhysicalSameFloat(t, "LibrationLatitude", got.LibrationLatitude, want.LibrationLatitude)
	assertMoonPhysicalSameFloat(t, "PositionAngle", got.PositionAngle, want.PositionAngle)
}

func TestTopocentricPhysicalPreservesInstantAcrossTimezones(t *testing.T) {
	utc := time.Date(2026, 4, 28, 9, 30, 45, 123000000, time.UTC)
	shanghai := utc.In(time.FixedZone("UTC+8", 8*3600))
	observerLon := 121.4737
	observerLat := 31.2304
	height := 4.0

	got := TopocentricPhysical(shanghai, observerLon, observerLat, height)
	want := TopocentricPhysical(utc, observerLon, observerLat, height)
	valuesGot := []float64{got.OpticalLongitude, got.OpticalLatitude, got.PhysicalLongitude, got.PhysicalLatitude, got.LibrationLongitude, got.LibrationLatitude, got.PositionAngle}
	valuesWant := []float64{want.OpticalLongitude, want.OpticalLatitude, want.PhysicalLongitude, want.PhysicalLatitude, want.LibrationLongitude, want.LibrationLatitude, want.PositionAngle}
	for i := range valuesGot {
		if math.Float64bits(valuesGot[i]) != math.Float64bits(valuesWant[i]) {
			t.Fatalf("timezone instant mismatch at index %d: got %.18f want %.18f", i, valuesGot[i], valuesWant[i])
		}
	}
}

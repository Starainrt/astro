package moon

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestPhysicalWrapperMatchesBasic(t *testing.T) {
	date := time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)
	jde := basic.Date2JDE(date.UTC())

	got := Physical(date)
	want := basic.MoonPhysicalN(basic.TD2UT(jde, true), -1)

	assertMoonPhysicalSameFloat(t, "OpticalLongitude", got.OpticalLongitude, want.OpticalLongitude)
	assertMoonPhysicalSameFloat(t, "OpticalLatitude", got.OpticalLatitude, want.OpticalLatitude)
	assertMoonPhysicalSameFloat(t, "PhysicalLongitude", got.PhysicalLongitude, want.PhysicalLongitude)
	assertMoonPhysicalSameFloat(t, "PhysicalLatitude", got.PhysicalLatitude, want.PhysicalLatitude)
	assertMoonPhysicalSameFloat(t, "LibrationLongitude", got.LibrationLongitude, want.LibrationLongitude)
	assertMoonPhysicalSameFloat(t, "LibrationLatitude", got.LibrationLatitude, want.LibrationLatitude)
	assertMoonPhysicalSameFloat(t, "PositionAngle", got.PositionAngle, want.PositionAngle)
}

func assertMoonPhysicalSameFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("%s mismatch: got %.18f want %.18f", name, got, want)
	}
}

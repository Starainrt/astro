package neptune

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
	gotN := PhysicalN(date, -1)
	want := basic.NeptunePhysicalN(basic.TD2UT(jde, true), -1)

	assertSamePhysicalFloat(t, "SubEarthLongitude", got.SubEarthLongitude, want.SubEarthLongitude)
	assertSamePhysicalFloat(t, "SubEarthLatitude", got.SubEarthLatitude, want.SubEarthLatitude)
	assertSamePhysicalFloat(t, "SubSolarLongitude", got.SubSolarLongitude, want.SubSolarLongitude)
	assertSamePhysicalFloat(t, "SubSolarLatitude", got.SubSolarLatitude, want.SubSolarLatitude)
	assertSamePhysicalFloat(t, "NorthPolePositionAngle", got.NorthPolePositionAngle, want.NorthPolePositionAngle)

	assertSamePhysicalFloat(t, "PhysicalN.SubEarthLongitude", got.SubEarthLongitude, gotN.SubEarthLongitude)
	assertSamePhysicalFloat(t, "PhysicalN.SubEarthLatitude", got.SubEarthLatitude, gotN.SubEarthLatitude)
	assertSamePhysicalFloat(t, "PhysicalN.SubSolarLongitude", got.SubSolarLongitude, gotN.SubSolarLongitude)
	assertSamePhysicalFloat(t, "PhysicalN.SubSolarLatitude", got.SubSolarLatitude, gotN.SubSolarLatitude)
	assertSamePhysicalFloat(t, "PhysicalN.NorthPolePositionAngle", got.NorthPolePositionAngle, gotN.NorthPolePositionAngle)
}

func assertSamePhysicalFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("%s mismatch: got %.18f want %.18f", name, got, want)
	}
}

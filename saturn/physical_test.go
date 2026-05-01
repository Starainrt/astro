package saturn

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
	want := basic.SaturnPhysicalN(basic.TD2UT(jde, true), -1)

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

func TestPhysicalSystemIIIAliasesPhysical(t *testing.T) {
	date := time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)
	got := Physical(date)
	alias := PhysicalSystemIII(date)
	aliasN := PhysicalSystemIIIN(date, -1)

	assertSamePhysicalFloat(t, "PhysicalSystemIII.SubEarthLongitude", alias.SubEarthLongitude, got.SubEarthLongitude)
	assertSamePhysicalFloat(t, "PhysicalSystemIII.SubEarthLatitude", alias.SubEarthLatitude, got.SubEarthLatitude)
	assertSamePhysicalFloat(t, "PhysicalSystemIII.SubSolarLongitude", alias.SubSolarLongitude, got.SubSolarLongitude)
	assertSamePhysicalFloat(t, "PhysicalSystemIII.SubSolarLatitude", alias.SubSolarLatitude, got.SubSolarLatitude)
	assertSamePhysicalFloat(t, "PhysicalSystemIII.NorthPolePositionAngle", alias.NorthPolePositionAngle, got.NorthPolePositionAngle)

	assertSamePhysicalFloat(t, "PhysicalSystemIIIN.SubEarthLongitude", aliasN.SubEarthLongitude, got.SubEarthLongitude)
	assertSamePhysicalFloat(t, "PhysicalSystemIIIN.SubEarthLatitude", aliasN.SubEarthLatitude, got.SubEarthLatitude)
	assertSamePhysicalFloat(t, "PhysicalSystemIIIN.SubSolarLongitude", aliasN.SubSolarLongitude, got.SubSolarLongitude)
	assertSamePhysicalFloat(t, "PhysicalSystemIIIN.SubSolarLatitude", aliasN.SubSolarLatitude, got.SubSolarLatitude)
	assertSamePhysicalFloat(t, "PhysicalSystemIIIN.NorthPolePositionAngle", aliasN.NorthPolePositionAngle, got.NorthPolePositionAngle)
}

func assertSamePhysicalFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("%s mismatch: got %.18f want %.18f", name, got, want)
	}
}

package jupiter

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
	want := basic.JupiterPhysicalN(basic.TD2UT(jde, true), -1)
	wantMeridians := basic.JupiterCentralMeridiansN(basic.TD2UT(jde, true), -1)
	gotMeridians := CentralMeridians(date)
	gotMeridiansN := CentralMeridiansN(date, -1)

	assertSamePhysicalFloat(t, "SubEarthLongitude", got.SubEarthLongitude, want.SubEarthLongitude)
	assertSamePhysicalFloat(t, "SubEarthLatitude", got.SubEarthLatitude, want.SubEarthLatitude)
	assertSamePhysicalFloat(t, "SubSolarLongitude", got.SubSolarLongitude, want.SubSolarLongitude)
	assertSamePhysicalFloat(t, "SubSolarLatitude", got.SubSolarLatitude, want.SubSolarLatitude)
	assertSamePhysicalFloat(t, "NorthPolePositionAngle", got.NorthPolePositionAngle, want.NorthPolePositionAngle)
	wantDS, wantDE := basic.JupiterDSDEN(basic.TD2UT(jde, true), -1)
	assertSamePhysicalFloat(t, "DS", got.DS, wantDS)
	assertSamePhysicalFloat(t, "DE", got.DE, wantDE)
	assertSamePhysicalFloat(t, "CentralMeridianSystemI", got.CentralMeridianSystemI, wantMeridians.SystemI)
	assertSamePhysicalFloat(t, "CentralMeridianSystemII", got.CentralMeridianSystemII, wantMeridians.SystemII)
	assertSamePhysicalFloat(t, "CentralMeridianSystemIII", got.CentralMeridianSystemIII, wantMeridians.SystemIII)

	assertSamePhysicalFloat(t, "PhysicalN.SubEarthLongitude", got.SubEarthLongitude, gotN.SubEarthLongitude)
	assertSamePhysicalFloat(t, "PhysicalN.SubEarthLatitude", got.SubEarthLatitude, gotN.SubEarthLatitude)
	assertSamePhysicalFloat(t, "PhysicalN.SubSolarLongitude", got.SubSolarLongitude, gotN.SubSolarLongitude)
	assertSamePhysicalFloat(t, "PhysicalN.SubSolarLatitude", got.SubSolarLatitude, gotN.SubSolarLatitude)
	assertSamePhysicalFloat(t, "PhysicalN.NorthPolePositionAngle", got.NorthPolePositionAngle, gotN.NorthPolePositionAngle)
	assertSamePhysicalFloat(t, "PhysicalN.DS", got.DS, gotN.DS)
	assertSamePhysicalFloat(t, "PhysicalN.DE", got.DE, gotN.DE)
	assertSamePhysicalFloat(t, "PhysicalN.CentralMeridianSystemI", got.CentralMeridianSystemI, gotN.CentralMeridianSystemI)
	assertSamePhysicalFloat(t, "PhysicalN.CentralMeridianSystemII", got.CentralMeridianSystemII, gotN.CentralMeridianSystemII)
	assertSamePhysicalFloat(t, "PhysicalN.CentralMeridianSystemIII", got.CentralMeridianSystemIII, gotN.CentralMeridianSystemIII)
	assertSamePhysicalFloat(t, "CentralMeridians.SystemI", gotMeridians.SystemI, wantMeridians.SystemI)
	assertSamePhysicalFloat(t, "CentralMeridians.SystemII", gotMeridians.SystemII, wantMeridians.SystemII)
	assertSamePhysicalFloat(t, "CentralMeridians.SystemIII", gotMeridians.SystemIII, wantMeridians.SystemIII)
	assertSamePhysicalFloat(t, "CentralMeridiansN.SystemI", gotMeridians.SystemI, gotMeridiansN.SystemI)
	assertSamePhysicalFloat(t, "CentralMeridiansN.SystemII", gotMeridians.SystemII, gotMeridiansN.SystemII)
	assertSamePhysicalFloat(t, "CentralMeridiansN.SystemIII", gotMeridians.SystemIII, gotMeridiansN.SystemIII)
}

func assertSamePhysicalFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("%s mismatch: got %.18f want %.18f", name, got, want)
	}
}

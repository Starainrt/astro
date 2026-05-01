package sun

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
	want := basic.SunPhysicalN(basic.TD2UT(jde, true), -1)

	assertSamePhysicalFloat(t, "P", got.P, want.P)
	assertSamePhysicalFloat(t, "B0", got.B0, want.B0)
	assertSamePhysicalFloat(t, "L0", got.L0, want.L0)

	assertSamePhysicalFloat(t, "PhysicalN.P", got.P, gotN.P)
	assertSamePhysicalFloat(t, "PhysicalN.B0", got.B0, gotN.B0)
	assertSamePhysicalFloat(t, "PhysicalN.L0", got.L0, gotN.L0)
}

func assertSamePhysicalFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("%s mismatch: got %.18f want %.18f", name, got, want)
	}
}

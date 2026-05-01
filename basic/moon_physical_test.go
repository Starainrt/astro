package basic

import (
	"math"
	"testing"
	"time"
)

func TestMoonPhysicalMeeusExample(t *testing.T) {
	info := MoonPhysical(2448724.5)

	assertClose(t, "OpticalLongitude", info.OpticalLongitude, -1.206, 0.01)
	assertClose(t, "OpticalLatitude", info.OpticalLatitude, 4.194, 0.01)
	assertClose(t, "PhysicalLongitude", info.PhysicalLongitude, -0.025, 0.01)
	assertClose(t, "PhysicalLatitude", info.PhysicalLatitude, 0.006, 0.01)
	assertClose(t, "LibrationLongitude", info.LibrationLongitude, -1.23, 0.02)
	assertClose(t, "LibrationLatitude", info.LibrationLatitude, 4.20, 0.02)
	assertClose(t, "PositionAngle", info.PositionAngle, 15.08, 0.02)
}

func TestMoonPhysicalNFullMatchesDefault(t *testing.T) {
	jd := 2461163.896354167

	got := MoonPhysical(jd)
	gotN := MoonPhysicalN(jd, -1)

	assertSameFloat(t, "OpticalLongitude", got.OpticalLongitude, gotN.OpticalLongitude)
	assertSameFloat(t, "OpticalLatitude", got.OpticalLatitude, gotN.OpticalLatitude)
	assertSameFloat(t, "PhysicalLongitude", got.PhysicalLongitude, gotN.PhysicalLongitude)
	assertSameFloat(t, "PhysicalLatitude", got.PhysicalLatitude, gotN.PhysicalLatitude)
	assertSameFloat(t, "LibrationLongitude", got.LibrationLongitude, gotN.LibrationLongitude)
	assertSameFloat(t, "LibrationLatitude", got.LibrationLatitude, gotN.LibrationLatitude)
	assertSameFloat(t, "PositionAngle", got.PositionAngle, gotN.PositionAngle)
}

func TestMoonPhysicalSampleSweepFiniteAndInRange(t *testing.T) {
	dates := []time.Time{
		time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1969, 7, 20, 20, 17, 40, 0, time.UTC),
		time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC),
		time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	for _, date := range dates {
		jd := TD2UT(Date2JDE(date.UTC()), true)
		info := MoonPhysical(jd)
		prefix := date.Format(time.RFC3339)

		assertFiniteSymmetric(t, prefix+".OpticalLongitude", info.OpticalLongitude, 180)
		assertFiniteSymmetric(t, prefix+".OpticalLatitude", info.OpticalLatitude, 90)
		assertFiniteSymmetric(t, prefix+".PhysicalLongitude", info.PhysicalLongitude, 180)
		assertFiniteSymmetric(t, prefix+".PhysicalLatitude", info.PhysicalLatitude, 90)
		assertFiniteSymmetric(t, prefix+".LibrationLongitude", info.LibrationLongitude, 180)
		assertFiniteSymmetric(t, prefix+".LibrationLatitude", info.LibrationLatitude, 90)
		assertFiniteSymmetric(t, prefix+".PositionAngle", info.PositionAngle, 90)
	}
}

func assertFiniteSymmetric(t *testing.T, name string, got, limit float64) {
	t.Helper()
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("%s is not finite: %.18f", name, got)
	}
	if got < -limit || got > limit {
		t.Fatalf("%s out of range: %.18f not in [-%.18f, %.18f]", name, got, limit, limit)
	}
}

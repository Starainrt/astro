package tools

import (
	"math"
	"testing"
)

// 参考值由同一组 IAU 定义值独立推出，不使用 9 位凑整常数。
const (
	testPcPerLy = LightYearKilometers / AstronomicalUnitKilometers / parsecAU
	testAuPerLy = LightYearKilometers / AstronomicalUnitKilometers
)

func assertDistanceClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.15f want %.15f", name, got, want)
	}
}

func TestDistanceToParsecsUnits(t *testing.T) {
	assertDistanceClose(t, "1 pc", DistanceToParsecs(1, DistanceParsec), 1, 1e-15)
	assertDistanceClose(t, "1 ly in pc", DistanceToParsecs(1, DistanceLightYear), testPcPerLy, 1e-15)
	assertDistanceClose(t, "1 ly in pc 参考 0.3066", testPcPerLy, 0.306601394, 1e-9)
	assertDistanceClose(t, "1 pc in ly 参考 3.261563777", 1/testPcPerLy, 3.261563777, 1e-9)
	assertDistanceClose(t, "1 ly in au 参考 63241.077084", DistanceToParsecs(1, DistanceLightYear)*parsecAU, 63241.077084, 1e-6)
	assertDistanceClose(t, "1 pc in au 参考 206264.806247096", DistanceToParsecs(1, DistanceParsec)*parsecAU, 206264.806247096, 1e-9)
	assertDistanceClose(t, "648000/pi au in pc", DistanceToParsecs(parsecAU, DistanceAU), 1, 1e-14)
}

func TestDistanceToParsecsRoundTrip(t *testing.T) {
	for _, pc := range []float64{0.0001, 1, 2.637, 5.7803, 1000} {
		assertDistanceClose(t, "pc round trip", DistanceToParsecs(pc, DistanceParsec), pc, 1e-15)
		ly := pc / testPcPerLy
		assertDistanceClose(t, "pc->ly->pc", DistanceToParsecs(ly, DistanceLightYear), pc, 1e-12)
		assertDistanceClose(t, "au round trip", DistanceToParsecs(pc*parsecAU, DistanceAU), pc, 1e-12)
	}
}

func TestDistanceToParsecsRejectsInvalidInput(t *testing.T) {
	for _, value := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if got := DistanceToParsecs(value, DistanceParsec); !math.IsNaN(got) {
			t.Fatalf("DistanceToParsecs(%v, parsec) = %v, want NaN", value, got)
		}
	}
	if got := DistanceToParsecs(1, DistanceUnit(200)); !math.IsNaN(got) {
		t.Fatalf("DistanceToParsecs(1, unknown unit) = %v, want NaN", got)
	}
}

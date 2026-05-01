package basic

import (
	"math"
	"testing"
	"time"
)

func TestMoonBrightLimbPositionAngleMeeusExample(t *testing.T) {
	assertPlanetPhaseClose(t, "MoonBrightLimbPositionAngle", MoonBrightLimbPositionAngle(2448724.5), 285.0, 0.1)
}

func TestMoonBrightLimbPositionAngleNFullMatchesDefault(t *testing.T) {
	jd := TD2UT(Date2JDE(time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)), true)

	got := MoonBrightLimbPositionAngle(jd)
	gotN := MoonBrightLimbPositionAngleN(jd, -1)
	if math.Float64bits(got) != math.Float64bits(gotN) {
		t.Fatalf("MoonBrightLimbPositionAngle full-n mismatch: got %.18f want %.18f", got, gotN)
	}
}

func TestMoonTopocentricBrightLimbPositionAngleSampleFiniteAndInRange(t *testing.T) {
	jd := TD2UT(Date2JDE(time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)), true)
	got := MoonTopocentricBrightLimbPositionAngle(jd, 121.4737, 31.2304, 4)

	assertFiniteRange(t, "MoonTopocentricBrightLimbPositionAngle", got, 0, 360, true)
	if angleDiffAbs(got, MoonBrightLimbPositionAngle(jd)) == 0 {
		t.Fatalf("expected topocentric bright limb angle to differ from geocentric value")
	}
}

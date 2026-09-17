package basic

import (
	"math"
	"testing"
)

func TestMoonHorizonMatchesTopocentricAltitude(t *testing.T) {
	for _, jd := range []float64{JDECalc(2026, 3, 3), JDECalc(2025, 9, 7), JDECalc(2024, 12, 15)} {
		points := MoonHorizon(jd, 360)
		if len(points) != 360 {
			t.Fatalf("horizon points=%d", len(points))
		}
		for _, point := range points {
			if altitude := HMoonHeight(jd, point[0], point[1], 0); math.Abs(altitude) > 1e-9 {
				t.Fatalf("JD=%v point=%v altitude=%g, want zero", jd, point, altitude)
			}
		}
	}
	if MoonHorizon(math.NaN(), 360) != nil {
		t.Fatal("invalid JD accepted")
	}
}

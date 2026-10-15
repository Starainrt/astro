package refraction

import (
	"math"
	"testing"
)

func TestRefractionRoundTrip(t *testing.T) {
	for _, alt := range []float64{-1, 0, 0.5, 5, 30, 80} {
		apparent := ApparentAltitude(alt, StandardPressureHPa, StandardTemperatureC)
		if back := TrueAltitude(apparent, StandardPressureHPa, StandardTemperatureC); math.Abs(back-alt) > 1e-9 {
			t.Errorf("alt %v: round trip %v", alt, back)
		}
	}
	if r := RefractionFromTrueAltitude(0, 1010, 10) * 60; r < 28 || r > 36 {
		t.Errorf("horizon refraction %.2f arcmin", r)
	}
	if !math.IsNaN(RefractionFromTrueAltitude(0, 0, 10)) {
		t.Error("zero pressure must be NaN")
	}
}

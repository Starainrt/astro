package formula

import (
	"math"
	"testing"
)

func TestExtinctionSchaefer(t *testing.T) {
	// 海平面、20 °C、湿度 50%、北纬 30°、春分点：逐项手算。
	got := ExtinctionSchaefer(0, 20, 50, 30, 0)
	aerosol := 0.1 * math.Pow(1-0.32/math.Log(0.5), 1.33)
	lat := math.Pi / 6
	ozone := 0.031 * (3 + 0.4*(lat-math.Cos(3*lat))) / 3
	water := 0.031 * 0.94 * 0.5 * math.Exp(20.0/15)
	assertFormulaClose(t, "Rayleigh", got.Rayleigh, 0.1066, 1e-15)
	assertFormulaClose(t, "Aerosol", got.Aerosol, aerosol, 1e-15)
	assertFormulaClose(t, "Ozone", got.Ozone, ozone, 1e-15)
	assertFormulaClose(t, "Water", got.Water, water, 1e-15)
	assertFormulaClose(t, "Total", got.Total, 0.1066+aerosol+ozone+water, 1e-15)
	if got.Total < 0.2 || got.Total > 0.5 {
		t.Fatalf("sea-level humid extinction out of typical range: %.3f", got.Total)
	}

	// 高山干燥站点消光明显更小。
	high := ExtinctionSchaefer(2500, 0, 20, 30, 0)
	if high.Total >= got.Total*0.6 {
		t.Fatalf("high dry site %.3f should be well below sea level %.3f", high.Total, got.Total)
	}

	for _, rh := range []float64{0, 100, -1, math.NaN()} {
		if !math.IsNaN(ExtinctionSchaefer(0, 20, rh, 30, 0).Total) {
			t.Fatalf("humidity %v should be rejected", rh)
		}
	}
	if !math.IsNaN(ExtinctionSchaefer(0, 20, 50, 91, 0).Total) {
		t.Fatal("latitude out of range should be rejected")
	}
}

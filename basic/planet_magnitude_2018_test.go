// 被测契约：2018 模型在中小相位角下与旧模型同量级，且不产生 NaN；大相位角两者本就分歧，只查有限。
package basic

import (
	"math"
	"testing"
)

func TestPlanetMagnitude2018CloseToClassic(t *testing.T) {
	classic := []func(float64) float64{nil, MercuryMag, VenusMag, MarsMag, JupiterMag, SaturnMag, UranusMag, NeptuneMag}
	limits := []float64{0, 0.6, 0.3, 0.3, 0.1, 0.5, 0.1, 0.2}
	for jd := 2451545.0; jd < 2462502; jd += 37.3 {
		for i := 1; i <= 7; i++ {
			got := planetMagnitude2018(i, jd, nil)
			if math.IsNaN(got) {
				t.Fatalf("index %d jd %.1f: NaN", i, jd)
			}
			if magnitudeTestPhase(i, jd) > 100 {
				continue
			}
			if d := math.Abs(got - classic[i](jd)); d > limits[i] {
				t.Errorf("index %d jd %.1f: 2018 %.3f vs classic %.2f", i, jd, got, classic[i](jd))
			}
		}
	}
}

func TestNeptuneMagnitude2018UnitEpochs(t *testing.T) {
	for _, c := range []struct{ jd, unit float64 }{
		{2444239.5 - 3653, -6.89}, {2451545.0 + 3653, -7.00},
	} {
		r, delta := NeptuneR(c.jd), EarthNeptuneAway(c.jd)
		got := planetMagnitude2018(7, c.jd, nil) - 5*math.Log10(r*delta)
		if math.Abs(got-c.unit) > 0.02 {
			t.Errorf("jd %.1f: unit magnitude %.3f, want about %.2f", c.jd, got, c.unit)
		}
	}
}

func magnitudeTestPhase(index int, jd float64) float64 {
	r, delta, earth := planetRWithInstance(index, jd, nil), planetEarthAwayWithInstance(index, jd, nil), planetRWithInstance(-1, jd, nil)
	return math.Acos((r*r+delta*delta-earth*earth)/(2*r*delta)) * 180 / math.Pi
}

// 输入取自模型作者随源码发布的测试行（日心距、地心距、相位角、子地/子日西经、日心黄经），期望值保留三位小数。
func TestMarsMagnitude2018RotationAndSeason(t *testing.T) {
	for _, c := range [][7]float64{
		{1.381191244505, 0.37274381097911, 4.8948, 329.27, 330.76, 334.4996, -2.862},
		{1.664150453905, 2.58995164518460, 11.5877, 76.84, 64.66, 147.3485, 1.788},
		{1.591952180003, 3.85882552272013, 167.9000, 29.50, 221.11, 212.8886, 8.977},
	} {
		if got := marsMagnitude2018(c[0], c[1], c[2], c[3], c[4], c[5]); math.Abs(got-c[6]) > 0.001 {
			t.Errorf("got %.4f want %.3f", got, c[6])
		}
	}
	if d := marsStirling(&marsRotationCorrection, 359.999) - marsStirling(&marsRotationCorrection, 0); math.Abs(d) > 1e-3 {
		t.Errorf("rotation correction not continuous across 0°: %.4f", d)
	}
}

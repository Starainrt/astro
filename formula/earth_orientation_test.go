package formula

import (
	"math"
	"testing"
)

func TestTIOLocator(t *testing.T) {
	// 2002-09-29 0h TT（MJD 52541）：s′ = −6.216698469981019e-12 rad。
	got := TIOLocatorArcsec(2452541.5) * arcsecToRadian
	assertFormulaClose(t, "TIOLocator", got, -0.6216698469981019309e-11, 1e-24)
	if !math.IsNaN(TIOLocatorArcsec(math.NaN())) {
		t.Fatal("expected NaN")
	}
}

func TestPolarMotionMatrix(t *testing.T) {
	const r2as = 1 / arcsecToRadian
	w := PolarMotionMatrix(2.55060238e-7*r2as, 1.860359247e-6*r2as, -0.1367174580728891460e-10*r2as)
	want := [3][3]float64{
		{0.9999999999999674721, -0.1367174580728846989e-10, 0.2550602379999972345e-6},
		{0.1414624947957029801e-10, 0.9999999999982695317, -0.1860359246998866389e-5},
		{-0.2550602379741215021e-6, 0.1860359247002414021e-5, 0.9999999999982370039},
	}
	for i := range want {
		for j := range want[i] {
			assertFormulaClose(t, "PolarMotionMatrix", w[i][j], want[i][j], 1e-16)
		}
	}
}

func TestCelestialPoleOffsetMatrix(t *testing.T) {
	// 改正后的极方向必须被映射到 z 轴：Δ·(dX, dY, 1) ≈ (0, 0, 1)。
	d := CelestialPoleOffsetMatrix(0.3, -0.2)
	x, y := 0.3*arcsecToRadian/1000, -0.2*arcsecToRadian/1000
	p := [3]float64{x, y, 1}
	for i, want := range [3]float64{0, 0, 1} {
		got := d[i][0]*p[0] + d[i][1]*p[1] + d[i][2]*p[2]
		assertFormulaClose(t, "CelestialPoleOffsetMatrix", got, want, 1e-20)
	}
	if !math.IsNaN(CelestialPoleOffsetMatrix(math.Inf(1), 0)[0][0]) {
		t.Fatal("expected NaN matrix")
	}
}

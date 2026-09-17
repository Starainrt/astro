package planet

import (
	"math"
	"testing"
)

func TestPlanetViewsMatchRawCuts(t *testing.T) {
	views := planetViews()
	for bodyIndex, raw := range planetRawData {
		view := views[bodyIndex]
		if math.Float64bits(view.scale) != math.Float64bits(raw[0]) {
			t.Fatalf("body=%d scale mismatch", bodyIndex)
		}
		for zn := 0; zn < 3; zn++ {
			pn := zn*6 + 1
			for order := 0; order < 6; order++ {
				start := int(raw[pn+order])
				end := int(raw[pn+order+1])
				if len(view.coords[zn].orders[order]) != end-start {
					t.Fatalf("body=%d zn=%d order=%d got=%d want=%d", bodyIndex, zn, order, len(view.coords[zn].orders[order]), end-start)
				}
			}
		}
	}
}

func TestBuildPlanetViewsRejectsInvalidCuts(t *testing.T) {
	_, err := buildPlanetViews([][]float64{{
		10000000000,
		20, 21, 20, 20, 20, 20, 20,
		20, 20, 20, 20, 20, 20,
		20, 20, 20, 20, 20, 20,
	}})
	if err == nil {
		t.Fatal("expected invalid cut error")
	}
}

func TestWherePlanetDomainGuards(t *testing.T) {
	jds := []float64{2415020.123456789, 2451545.0, 2469808.7654321}
	for _, jd := range jds {
		for xt := -1; xt <= 7; xt++ {
			for zn := 0; zn <= 2; zn++ {
				for _, n := range []int{-1, 0, 3, 100} {
					got := WherePlanetN(xt, zn, jd, n)
					if math.IsNaN(got) || math.IsInf(got, 0) {
						t.Fatalf("in-domain WherePlanetN(%d, %d, %v, %d) = %v", xt, zn, jd, n, got)
					}
				}
				got := WherePlanet(xt, zn, jd)
				if math.IsNaN(got) || math.IsInf(got, 0) {
					t.Fatalf("in-domain WherePlanet(%d, %d, %v) = %v", xt, zn, jd, got)
				}
			}
		}
		for _, xt := range []int{-2, -100, 8, 9, 100} {
			for zn := 0; zn <= 2; zn++ {
				assertWherePlanetNaN(t, xt, zn, jd)
			}
		}
		for _, zn := range []int{-1, -100, 3, 100} {
			for xt := -1; xt <= 7; xt++ {
				assertWherePlanetNaN(t, xt, zn, jd)
			}
		}
	}
}

func assertWherePlanetNaN(t *testing.T, xt, zn int, jd float64) {
	t.Helper()
	if got := WherePlanet(xt, zn, jd); !math.IsNaN(got) {
		t.Fatalf("WherePlanet(%d, %d, %v) = %v, want NaN", xt, zn, jd, got)
	}
	for _, n := range []int{-1, 0, 6} {
		if got := WherePlanetN(xt, zn, jd, n); !math.IsNaN(got) {
			t.Fatalf("WherePlanetN(%d, %d, %v, %d) = %v, want NaN", xt, zn, jd, n, got)
		}
	}
}

func TestWherePlanetGoldenValues(t *testing.T) {
	jd := 2451545.0
	cases := []struct {
		xt, zn int
		want   float64
	}{
		{-1, 0, 100.37782037193487},
		{-1, 1, -0.00022991463883506049},
		{-1, 2, 0.98332764743480028},
		{0, 0, 280.37782037193489},
		{0, 1, 0.00022991463883506049},
		{0, 2, 0.98332764743480028},
		{4, 0, 36.29462069188007},
		{4, 1, -1.1745385942079292},
		{4, 2, 4.9653814383525443},
	}
	for _, tc := range cases {
		got := WherePlanet(tc.xt, tc.zn, jd)
		if math.Abs(got-tc.want) > 1e-12 {
			t.Fatalf("WherePlanet(%d, %d, %v) = %.17g, want %.17g", tc.xt, tc.zn, jd, got, tc.want)
		}
	}
}

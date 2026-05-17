package planet

import (
	"math"
	"testing"
)

func TestWherePlanetNFullMatchesDefault(t *testing.T) {
	jds := []float64{
		2415020.123456789,
		2451545.0,
		2469808.7654321,
	}
	for _, jd := range jds {
		for xt := -1; xt < 8; xt++ {
			for zn := 0; zn < 3; zn++ {
				got := WherePlanet(xt, zn, jd)
				gotN := WherePlanetN(xt, zn, jd, -1)
				if math.Float64bits(got) != math.Float64bits(gotN) {
					t.Fatalf("jd=%f xt=%d zn=%d full mismatch: got=%v gotN=%v", jd, xt, zn, got, gotN)
				}
			}
		}
	}
}

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

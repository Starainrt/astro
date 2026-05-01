package formula

import (
	"math"
	"testing"
)

func TestAirmassModels(t *testing.T) {
	assertFormulaClose(t, "AirmassPlaneParallel(90)", AirmassPlaneParallel(90), 1, 1e-15)
	assertFormulaClose(t, "AirmassPlaneParallel(30)", AirmassPlaneParallel(30), 2, 1e-15)
	if !math.IsInf(AirmassPlaneParallel(0), 1) {
		t.Fatal("expected plane-parallel airmass at horizon to be +Inf")
	}

	assertFormulaClose(t, "AirmassPlaneParallelByZenithDistance(0)", AirmassPlaneParallelByZenithDistance(0), 1, 1e-15)
	assertFormulaClose(t, "AirmassPlaneParallelByZenithDistance(60)", AirmassPlaneParallelByZenithDistance(60), 2, 1e-12)
	if !math.IsInf(AirmassPlaneParallelByZenithDistance(90), 1) {
		t.Fatal("expected sec(z) at z=90 to be +Inf")
	}

	assertFormulaClose(t, "AirmassKastenYoung(90)", AirmassKastenYoung(90), 0.9997119918558381, 1e-15)
	assertFormulaClose(t, "AirmassKastenYoung(30)", AirmassKastenYoung(30), 1.9942928525292503, 1e-15)
	assertFormulaClose(t, "AirmassKastenYoung(0)", AirmassKastenYoung(0), 37.91960837783633, 1e-12)

	assertFormulaClose(t, "AirmassPickering(90)", AirmassPickering(90), 1.000000196171337, 1e-15)
	assertFormulaClose(t, "AirmassPickering(30)", AirmassPickering(30), 1.9931538464145713, 1e-15)
	assertFormulaClose(t, "AirmassPickering(0)", AirmassPickering(0), 38.749398755780355, 1e-12)
}

func TestAirmassInvalidInput(t *testing.T) {
	for name, value := range map[string]float64{
		"plane-parallel negative altitude": AirmassPlaneParallel(-1),
		"plane-parallel >90 altitude":      AirmassPlaneParallel(91),
		"plane-parallel zenith <0":         AirmassPlaneParallelByZenithDistance(-1),
		"plane-parallel zenith >90":        AirmassPlaneParallelByZenithDistance(91),
		"kasten-young negative altitude":   AirmassKastenYoung(-1),
		"pickering >90 altitude":           AirmassPickering(91),
	} {
		if !math.IsNaN(value) {
			t.Fatalf("%s should be NaN, got %.15f", name, value)
		}
	}
}

package basic

import (
	"math"
	"testing"

	. "github.com/starainrt/astro/tools"
)

func TestEclipticToEquatorialCachedObliquityMatchesRepeatedEvaluation(t *testing.T) {
	for _, test := range []struct {
		jde float64
		lo  float64
		bo  float64
	}{
		{jde: 2451545.0, lo: 0, bo: 0},
		{jde: 2460832.0, lo: 193.25, bo: -5.75},
		{jde: 2378496.5, lo: 359.999, bo: 87.5},
	} {
		legacyDec := ArcSin(
			Sin(test.bo)*Cos(TrueObliquity(test.jde)) +
				Cos(test.bo)*Sin(TrueObliquity(test.jde))*Sin(test.lo),
		)
		legacyRA := math.Atan2(
			Sin(test.lo)*Cos(TrueObliquity(test.jde))-Tan(test.bo)*Sin(TrueObliquity(test.jde)),
			Cos(test.lo),
		) * 180 / math.Pi
		if legacyRA < 0 {
			legacyRA += 360
		}

		ra, dec := LoBoToRaDec(test.jde, test.lo, test.bo)
		if ra != legacyRA || dec != legacyDec {
			t.Fatalf("LoBoToRaDec(%v, %v, %v) = %.15g %.15g, want %.15g %.15g",
				test.jde, test.lo, test.bo, ra, dec, legacyRA, legacyDec)
		}
		if got := LoToRa(test.jde, test.lo, test.bo); got != legacyRA {
			t.Fatalf("LoToRa(%v, %v, %v) = %.15g, want %.15g", test.jde, test.lo, test.bo, got, legacyRA)
		}
		if got := BoToDec(test.jde, test.lo, test.bo); got != legacyDec {
			t.Fatalf("BoToDec(%v, %v, %v) = %.15g, want %.15g", test.jde, test.lo, test.bo, got, legacyDec)
		}
	}
}

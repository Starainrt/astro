package basic

import (
	"math"
	"testing"
)

// 本文件用改动前的“七个候选全精修”实现做差分对照与 A/B 基准。

func moonMaximumDeclinationReferenceSearch(jd float64, coeffs moonMaxDeclinationCoefficients, direction int, includeCurrent bool) DeclinationEvent {
	if !finite(jd) {
		return DeclinationEvent{}
	}
	cfg := moonMaximumDeclinationSearchConfig(coeffs)
	centerK := moonMaximumDeclinationOffset(jd, coeffs)
	found := false
	bestDistance := math.Inf(1)
	var best DeclinationEvent
	for offset := -moonMaxDeclinationSearchSpan; offset <= moonMaxDeclinationSearchSpan; offset++ {
		event := moonMaximumDeclinationEvent(centerK+offset, coeffs, cfg)
		delta := event.JD - jd
		if !moonMaximumDeclinationMatchesDirection(delta, direction, includeCurrent) {
			continue
		}
		distance := math.Abs(delta)
		if !found || distance < bestDistance || (distance == bestDistance && moonMaximumDeclinationEarlier(event, best)) {
			best = event
			bestDistance = distance
			found = true
		}
	}
	return best
}

func moonMaximumDeclinationReferenceClosest(jd float64, coeffs moonMaxDeclinationCoefficients) DeclinationEvent {
	if !finite(jd) {
		return DeclinationEvent{}
	}
	last := moonMaximumDeclinationReferenceSearch(jd, coeffs, -1, true)
	next := moonMaximumDeclinationReferenceSearch(jd, coeffs, 1, false)
	lastDistance := math.Abs(jd - last.JD)
	nextDistance := math.Abs(next.JD - jd)
	if lastDistance <= nextDistance {
		return last
	}
	return next
}

func TestMoonMaximumDeclinationPruningMatchesFullSearch(t *testing.T) {
	for _, coeffs := range []moonMaxDeclinationCoefficients{moonMaxDeclinationNorthCoefficients, moonMaxDeclinationSouthCoefficients} {
		for jd := JDCalc(2024, 1, 1); jd <= JDCalc(2025, 6, 1); jd += 2.5 {
			for _, direction := range []struct {
				name           string
				dir            int
				includeCurrent bool
				fn             func(float64) DeclinationEvent
			}{
				{"Last", -1, true, func(v float64) DeclinationEvent {
					return moonMaximumDeclinationSearch(v, coeffs, -1, true)
				}},
				{"Next", 1, false, func(v float64) DeclinationEvent {
					return moonMaximumDeclinationSearch(v, coeffs, 1, false)
				}},
			} {
				want := moonMaximumDeclinationReferenceSearch(jd, coeffs, direction.dir, direction.includeCurrent)
				got := direction.fn(jd)
				if got != want {
					t.Fatalf("k=%v %s(%v) = %+v, want %+v", coeffs.sign, direction.name, jd, got, want)
				}
			}
			wantClosest := moonMaximumDeclinationReferenceClosest(jd, coeffs)
			if got := moonClosestMaximumDeclination(jd, coeffs); got != wantClosest {
				t.Fatalf("k=%v Closest(%v) = %+v, want %+v", coeffs.sign, jd, got, wantClosest)
			}
		}
	}
}

func BenchmarkMoonMaximumDeclinationFullSearchReference(b *testing.B) {
	jd := JDCalc(2025, 3, 1)
	b.Run("Next", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = moonMaximumDeclinationReferenceSearch(jd, moonMaxDeclinationNorthCoefficients, 1, false)
		}
	})
	b.Run("Closest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = moonMaximumDeclinationReferenceClosest(jd, moonMaxDeclinationNorthCoefficients)
		}
	})
}

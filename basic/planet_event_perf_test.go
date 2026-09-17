package basic

import (
	"math"
	"testing"
)

// 本切片（行星/月球事件层）的性能守护基准；改动热路径时必须给出改前→改后。

func BenchmarkPlanetEventGreatestElongationAnySide(b *testing.B) {
	jd := JDECalc(2025, 3, 1)
	b.Run("Mercury", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NextMercuryGreatestElongation(jd)
		}
	})
	b.Run("Venus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NextVenusGreatestElongation(jd)
		}
	})
}

func BenchmarkPlanetEventOuterStationNonFiniteQuery(b *testing.B) {
	for _, tc := range []struct {
		name string
		fn   func(float64) float64
	}{
		{"Mars", NextMarsProgradeToRetrograde},
		{"Jupiter", NextJupiterProgradeToRetrograde},
		{"Saturn", NextSaturnProgradeToRetrograde},
		{"Uranus", NextUranusProgradeToRetrograde},
		{"Neptune", NextNeptuneProgradeToRetrograde},
	} {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tc.fn(math.NaN())
			}
		})
	}
}

func BenchmarkPlanetEventMoonMaximumDeclination(b *testing.B) {
	jd := JDECalc(2025, 3, 1)
	b.Run("Next", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NextMoonMaximumNorthDeclination(jd)
		}
	})
	b.Run("Closest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ClosestMoonMaximumNorthDeclination(jd)
		}
	})
}

func BenchmarkPlanetEventLunarEclipse(b *testing.B) {
	jd := 2458860.0
	for i := 0; i < b.N; i++ {
		_ = LunarEclipse(jd)
	}
}

func BenchmarkPlanetEventJupiterGalileanCallisto(b *testing.B) {
	b.Run("NextTransit", func(b *testing.B) {
		jd := 2463143.9646
		for i := 0; i < b.N; i++ {
			_ = NextJupiterGalileanPhenomenonEvent(jd, 4, JupiterGalileanTransit)
		}
	})
	b.Run("ClosestOccultation", func(b *testing.B) {
		jd := 2463400.0
		for i := 0; i < b.N; i++ {
			_ = ClosestJupiterGalileanPhenomenonEvent(jd, 4, JupiterGalileanOccultation)
		}
	})
}

func BenchmarkPlanetEventJupiterSixEventsSameInstant(b *testing.B) {
	jd := JDECalc(2025, 3, 1)
	for i := 0; i < b.N; i++ {
		_ = NextJupiterConjunction(jd)
		_ = NextJupiterOpposition(jd)
		_ = NextJupiterEasternQuadrature(jd)
		_ = NextJupiterWesternQuadrature(jd)
		_ = NextJupiterProgradeToRetrograde(jd)
		_ = NextJupiterRetrogradeToPrograde(jd)
	}
}

func BenchmarkPlanetEventLunarEclipseDiagram(b *testing.B) {
	jd := 2458860.0
	options := LunarEclipseDiagramOptions{StepDays: 1.0 / 86400.0}
	for i := 0; i < b.N; i++ {
		_ = LunarEclipseDiagram(jd, options)
	}
}

// 差分对照基准：改动前“两侧都算再取极值”的聚合方式。
func BenchmarkPlanetEventGreatestElongationBothSidesReference(b *testing.B) {
	mercuryJD, venusJD := JDECalc(2025, 3, 1), JDECalc(2025, 3, 1)
	b.Run("Mercury", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = earliestFiniteEventUT(NextMercuryGreatestElongationEast(mercuryJD), NextMercuryGreatestElongationWest(mercuryJD))
		}
	})
	b.Run("Venus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = earliestFiniteEventUT(NextVenusGreatestElongationEast(venusJD), NextVenusGreatestElongationWest(venusJD))
		}
	})
}

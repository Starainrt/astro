package basic

import "testing"

func benchmarkOuterPlanetEventFamily(b *testing.B, plan outerPlanetEventPlan, cases []outerPlanetEventCase) {
	samples := plan.samples()
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, sample := range samples {
			jd := outerPlanetEventSampleTTJD(sample)
			for _, event := range cases {
				sink += event.fn(jd)
			}
		}
	}
	_ = sink
}

func BenchmarkOuterPlanetEventFamilies(b *testing.B) {
	for _, plan := range outerPlanetEventPlans() {
		plan := plan
		b.Run(plan.planet+"PhaseFamily", func(b *testing.B) {
			benchmarkOuterPlanetEventFamily(b, plan, plan.phaseCases)
		})
		b.Run(plan.planet+"RetrogradeFamily", func(b *testing.B) {
			benchmarkOuterPlanetEventFamily(b, plan, plan.retroCases)
		})
	}
}

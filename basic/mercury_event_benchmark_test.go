package basic

import "testing"

func BenchmarkMercuryConjunctionFamily(b *testing.B) {
	cases := mercuryEventCases()[:6]
	samples := mercuryEventSamples()
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, sample := range samples {
			jd := mercuryEventSampleTTJD(sample)
			for _, event := range cases {
				sink += event.fn(jd)
			}
		}
	}
	_ = sink
}

func BenchmarkMercuryRetrogradeFamily(b *testing.B) {
	cases := mercuryEventCases()[6:12]
	samples := mercuryEventSamples()
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, sample := range samples {
			jd := mercuryEventSampleTTJD(sample)
			for _, event := range cases {
				sink += event.fn(jd)
			}
		}
	}
	_ = sink
}

func BenchmarkMercuryGreatestElongationFamily(b *testing.B) {
	cases := mercuryEventCases()[12:]
	samples := mercuryEventSamples()
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, sample := range samples {
			jd := mercuryEventSampleTTJD(sample)
			for _, event := range cases {
				sink += event.fn(jd)
			}
		}
	}
	_ = sink
}

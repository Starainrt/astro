package basic

import "testing"

func BenchmarkMarsPhaseFamily(b *testing.B) {
	cases := marsEventCases()[:8]
	samples := marsEventSamples()
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, sample := range samples {
			jd := marsEventSampleTTJD(sample)
			for _, event := range cases {
				sink += event.fn(jd)
			}
		}
	}
	_ = sink
}

func BenchmarkMarsRetrogradeFamily(b *testing.B) {
	cases := marsEventCases()[8:]
	samples := marsEventSamples()
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, sample := range samples {
			jd := marsEventSampleTTJD(sample)
			for _, event := range cases {
				sink += event.fn(jd)
			}
		}
	}
	_ = sink
}

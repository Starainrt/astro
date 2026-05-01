package basic

import "testing"

func BenchmarkVenusConjunctionFamily(b *testing.B) {
	cases := venusEventCases()[:6]
	samples := venusEventSamples()
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, sample := range samples {
			jd := venusEventSampleTTJD(sample)
			for _, event := range cases {
				sink += event.fn(jd)
			}
		}
	}
	_ = sink
}

func BenchmarkVenusRetrogradeFamily(b *testing.B) {
	cases := venusEventCases()[6:12]
	samples := venusEventSamples()
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, sample := range samples {
			jd := venusEventSampleTTJD(sample)
			for _, event := range cases {
				sink += event.fn(jd)
			}
		}
	}
	_ = sink
}

func BenchmarkVenusGreatestElongationFamily(b *testing.B) {
	cases := venusEventCases()[12:]
	samples := venusEventSamples()
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, sample := range samples {
			jd := venusEventSampleTTJD(sample)
			for _, event := range cases {
				sink += event.fn(jd)
			}
		}
	}
	_ = sink
}

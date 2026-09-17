package planet

import "testing"

func BenchmarkWherePlanet(b *testing.B) {
	const jd = 2451545.0
	var sink float64
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink += WherePlanet(4, 0, jd)
	}
	_ = sink
}

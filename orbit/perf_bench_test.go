package orbit

import (
	"testing"
	"time"
)

var (
	benchmarkParallacticSink float64
	benchmarkPositionSink    EquatorialPosition
)

func benchmarkObservationInputs() (time.Time, Elements) {
	return time.Date(2025, 11, 21, 20, 0, 0, 0, time.FixedZone("CST", 8*3600)), sampleObservationElements()
}

func BenchmarkParallacticAngle(b *testing.B) {
	date, elements := benchmarkObservationInputs()
	b.ReportAllocs()
	b.ResetTimer()
	sink := 0.0
	for i := 0; i < b.N; i++ {
		sink = ParallacticAngle(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	}
	benchmarkParallacticSink = sink
}

func BenchmarkParallacticAngleLegacy(b *testing.B) {
	date, elements := benchmarkObservationInputs()
	b.ReportAllocs()
	b.ResetTimer()
	sink := 0.0
	for i := 0; i < b.N; i++ {
		sink = duplicatedSolveParallacticAngle(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	}
	benchmarkParallacticSink = sink
}

func BenchmarkApparentTopocentricEquatorial(b *testing.B) {
	date, elements := benchmarkObservationInputs()
	b.ReportAllocs()
	b.ResetTimer()
	sink := EquatorialPosition{}
	for i := 0; i < b.N; i++ {
		sink = ApparentTopocentricEquatorial(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	}
	benchmarkPositionSink = sink
}

func BenchmarkHourAngle(b *testing.B) {
	date, elements := benchmarkObservationInputs()
	b.ReportAllocs()
	b.ResetTimer()
	sink := 0.0
	for i := 0; i < b.N; i++ {
		sink = HourAngle(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	}
	benchmarkParallacticSink = sink
}

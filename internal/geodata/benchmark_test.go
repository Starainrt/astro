package geodata

import (
	"math"
	"testing"
)

func benchmarkRingGrid(count, points int, centerLatitude, centerLongitude, latitudeStep float64) [][]GeoPoint {
	rings := make([][]GeoPoint, count)
	for index := range rings {
		rings[index] = SphericalCircle(GeoPoint{
			Longitude: centerLongitude + 1.2*float64(index%10),
			Latitude:  centerLatitude + latitudeStep*float64(index/10),
		}, 2, points)
	}
	return rings
}

func BenchmarkUnionPolygons100x64Lat0(b *testing.B) {
	rings := benchmarkRingGrid(100, 64, 0, 0, 1.2)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if _, err := UnionPolygons(rings); err != nil {
			b.Fatalf("UnionPolygons: %v", err)
		}
	}
}

func BenchmarkUnionPolygons100x64Lat80(b *testing.B) {
	rings := benchmarkRingGrid(100, 64, 80, 0, 1.2)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if _, err := UnionPolygons(rings); err != nil {
			b.Fatalf("UnionPolygons: %v", err)
		}
	}
}

func BenchmarkUnionPolygons100x64Lat80Cap(b *testing.B) {
	rings := benchmarkRingGrid(99, 64, 78, 0, 1)
	rings = append(rings, SphericalCircle(GeoPoint{Longitude: 0, Latitude: 87}, 5, 64))
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if _, err := UnionPolygons(rings); err != nil {
			b.Fatalf("UnionPolygons: %v", err)
		}
	}
}

func BenchmarkSweepBridgeTouchingPolygons120x128(b *testing.B) {
	rings := make([][]GeoPoint, 0, 120)
	for row := 0; row < 10; row++ {
		for col := 0; col < 12; col++ {
			rings = append(rings, SphericalCircle(GeoPoint{
				Longitude: 15 * float64(col),
				Latitude:  10 + 5*float64(row),
			}, 0.5, 128))
		}
	}
	// 最近的一对相距约 32 米：大于 0.01 公里的桥接门限，且排在扫描顺序最后。
	rings[119] = SphericalCircle(GeoPoint{Longitude: 150.0005, Latitude: 55}, 0.5, 128)
	for iteration := 0; iteration < b.N; iteration++ {
		input := make([][]GeoPoint, len(rings))
		for index, ring := range rings {
			input[index] = append([]GeoPoint(nil), ring...)
		}
		result, err := sweepBridgeTouchingPolygons(input)
		if err != nil {
			b.Fatalf("sweepBridgeTouchingPolygons: %v", err)
		}
		if len(result) != len(input) {
			b.Fatalf("ring count=%d, want %d", len(result), len(input))
		}
	}
}

func benchmarkCycleGridLines(size int) [][]GeoPoint {
	lines := make([][]GeoPoint, 0, 2*size*(size-1))
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			x, y := float64(col), float64(row)
			if col < size-1 {
				lines = append(lines, []GeoPoint{{x, y}, {x + 1, y}})
			}
			if row < size-1 {
				lines = append(lines, []GeoPoint{{x, y}, {x, y + 1}})
			}
		}
	}
	return lines
}

func benchmarkWheelLines(spokes int) [][]GeoPoint {
	lines := make([][]GeoPoint, 0, 2*spokes)
	ring := make([]GeoPoint, spokes)
	for index := range ring {
		angle := 2 * math.Pi * float64(index) / float64(spokes)
		ring[index] = GeoPoint{Longitude: 10 * math.Cos(angle), Latitude: 10 * math.Sin(angle)}
	}
	for index := range ring {
		lines = append(lines, []GeoPoint{{Longitude: 0, Latitude: 0}, ring[index]})
		lines = append(lines, []GeoPoint{ring[index], ring[(index+1)%spokes]})
	}
	return lines
}

func BenchmarkEnumerateVisibleLineworkCyclesWheel(b *testing.B) {
	lines := benchmarkWheelLines(720)
	chart := polygonUnionChart{xAxis: geoVector3{x: 1}, yAxis: geoVector3{y: 1}, zAxis: geoVector3{z: 1}}
	nodes, edges, _ := buildVisibleLineworkGraph(lines, chart, 1)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		enumerateVisibleLineworkCycles(nodes, edges)
	}
}

func BenchmarkEnumerateVisibleLineworkCycles30(b *testing.B) {
	lines := benchmarkCycleGridLines(30)
	chart := polygonUnionChart{xAxis: geoVector3{x: 1}, yAxis: geoVector3{y: 1}, zAxis: geoVector3{z: 1}}
	nodes, edges, _ := buildVisibleLineworkGraph(lines, chart, 1)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		enumerateVisibleLineworkCycles(nodes, edges)
	}
}

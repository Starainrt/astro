package coord

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

var benchmarkEclipticSink Ecliptic

func benchmarkEclipticInputs() (time.Time, float64, float64, float64, float64, float64, float64) {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)), 139.7, 4.9, 121.4737, 31.2304, 0.00257, 20
}

func BenchmarkTopocentricEcliptic(b *testing.B) {
	date, lon, lat, observerLon, observerLat, distanceAU, height := benchmarkEclipticInputs()
	b.ReportAllocs()
	b.ResetTimer()
	sink := Ecliptic{}
	for i := 0; i < b.N; i++ {
		sink = TopocentricEcliptic(date, lon, lat, observerLon, observerLat, distanceAU, height)
	}
	benchmarkEclipticSink = sink
}

func BenchmarkTopocentricEclipticLegacy(b *testing.B) {
	date, lon, lat, observerLon, observerLat, distanceAU, height := benchmarkEclipticInputs()
	b.ReportAllocs()
	b.ResetTimer()
	sink := Ecliptic{}
	for i := 0; i < b.N; i++ {
		sink = Ecliptic{
			Lon: basic.TopocentricLo(lon, lat, observerLat, observerLon, jdeUTC(date), distanceAU, height),
			Lat: basic.TopocentricBo(lon, lat, observerLat, observerLon, jdeUTC(date), distanceAU, height),
		}
	}
	benchmarkEclipticSink = sink
}

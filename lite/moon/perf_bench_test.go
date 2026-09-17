package moon

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	lite "github.com/starainrt/astro/lite/internal"
)

var benchmarkHorizontalSink float64

func benchmarkHorizontalInputs() (time.Time, float64, float64) {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)), 121.4737, 31.2304
}

func BenchmarkMoonHourAngle(b *testing.B) {
	date, lon, lat := benchmarkHorizontalInputs()
	b.ReportAllocs()
	b.ResetTimer()
	sink := 0.0
	for i := 0; i < b.N; i++ {
		sink = HourAngle(date, lon, lat)
	}
	benchmarkHorizontalSink = sink
}

func BenchmarkMoonAzimuth(b *testing.B) {
	date, lon, lat := benchmarkHorizontalInputs()
	b.ReportAllocs()
	b.ResetTimer()
	sink := 0.0
	for i := 0; i < b.N; i++ {
		sink = Azimuth(date, lon, lat)
	}
	benchmarkHorizontalSink = sink
}

func BenchmarkMoonAltitude(b *testing.B) {
	date, lon, lat := benchmarkHorizontalInputs()
	b.ReportAllocs()
	b.ResetTimer()
	sink := 0.0
	for i := 0; i < b.N; i++ {
		sink = Altitude(date, lon, lat)
	}
	benchmarkHorizontalSink = sink
}

func BenchmarkMoonZenith(b *testing.B) {
	date, lon, lat := benchmarkHorizontalInputs()
	b.ReportAllocs()
	b.ResetTimer()
	sink := 0.0
	for i := 0; i < b.N; i++ {
		sink = Zenith(date, lon, lat)
	}
	benchmarkHorizontalSink = sink
}

func BenchmarkMoonHorizontalLegacy(b *testing.B) {
	date, lon, lat := benchmarkHorizontalInputs()
	jd := basic.Date2JDE(date.UTC())
	entries := []struct {
		name string
		pick func(altitude, azimuth, hourAngle float64) float64
	}{
		{"HourAngle", func(_, _, hourAngle float64) float64 { return hourAngle }},
		{"Azimuth", func(_, azimuth, _ float64) float64 { return azimuth }},
		{"Altitude", func(altitude, _, _ float64) float64 { return altitude }},
		{"Zenith", func(altitude, _, _ float64) float64 { return 90 - altitude }},
	}
	for _, entry := range entries {
		b.Run(entry.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			sink := 0.0
			for i := 0; i < b.N; i++ {
				sink = entry.pick(lite.HorizontalCoordinates(ApparentRa(date, lon, lat), ApparentDec(date, lon, lat), jd, lon, lat))
			}
			benchmarkHorizontalSink = sink
		})
	}
}

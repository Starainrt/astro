package svg

import (
	"testing"
	"time"
)

// 一秒步长请求应被夹到两分钟，耗时与默认请求同量级。
func BenchmarkSolarEclipseMapSVGPartialStepOneSecond(b *testing.B) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	options := SolarEclipseMapSVGOptions{
		Width: 900, Height: 620, Location: time.UTC, PartialStep: time.Second,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := SolarEclipseMapSVG(date, options); !ok {
			b.Fatal("missing map")
		}
	}
}

func BenchmarkSolarEclipseMapSVG(b *testing.B) {
	for _, tc := range []struct {
		name string
		date time.Time
	}{
		{name: "2010Annular", date: time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC)},
		{name: "2009Total", date: time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			options := SolarEclipseMapSVGOptions{
				Width: 1200, Height: 800, Location: time.UTC,
				PartialStep: 2 * time.Minute, CentralStep: 2 * time.Minute,
				GreatestTimeStep: solarEclipseMapDefaultGreatestTimeStep,
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, ok := SolarEclipseMapSVG(tc.date, options); !ok {
					b.Fatal("missing map")
				}
			}
		})
	}
}

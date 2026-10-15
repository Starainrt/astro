package eclipse

import (
	"testing"
	"time"
)

func BenchmarkSarosLookup(b *testing.B) {
	for _, tc := range []struct {
		name string
		date time.Time
	}{
		{name: "NASA", date: time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC)},
		{name: "PrecomputedFuture", date: time.Date(3300, 1, 1, 12, 0, 0, 0, time.UTC)},
		{name: "PrecomputedAncient", date: time.Date(-2800, 1, 1, 12, 0, 0, 0, time.UTC)},
		{name: "LiveAncient", date: time.Date(-3200, 1, 1, 12, 0, 0, 0, time.UTC)},
		{name: "LiveFuture", date: time.Date(7000, 1, 1, 12, 0, 0, 0, time.UTC)},
	} {
		b.Run("Solar/"+tc.name, func(b *testing.B) {
			info := ClosestSolarEclipse(tc.date)
			tt := solarEclipseTimeToTTJDE(info.GreatestEclipse)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if got, ok := solarSarosInfo(tt); !ok || got != info.Saros {
					b.Fatal("inconsistent solar Saros metadata")
				}
			}
		})
		b.Run("Lunar/"+tc.name, func(b *testing.B) {
			info := ClosestLunarEclipse(tc.date)
			tt := timeToTTJDE(info.Maximum)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if got, ok := lunarSarosInfo(tt); !ok || got != info.Saros {
					b.Fatal("inconsistent lunar Saros metadata")
				}
			}
		})
	}
}

func BenchmarkSolarEclipsePrecomputedSaros(b *testing.B) {
	date := time.Date(3306, 11, 28, 12, 0, 0, 0, time.UTC)
	for i := 0; i < b.N; i++ {
		info, ok := SolarEclipseOnDate(date)
		if !ok || !info.HasSaros || info.Saros.Series != 202 {
			b.Fatal("missing precomputed Saros metadata")
		}
	}
}

func BenchmarkSolarEclipseCandidatesSaros(b *testing.B) {
	start := time.Date(6001, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(6002, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, include := range []bool{false, true} {
		name := "WithoutSaros"
		if include {
			name = "WithSaros"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if got := SolarEclipseCandidates(start, end, SolarEclipseCandidateOptions{IncludeSaros: include}); len(got) == 0 {
					b.Fatal("missing candidates")
				}
			}
		})
	}
}

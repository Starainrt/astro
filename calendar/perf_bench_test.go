package calendar

import (
	"testing"
	"time"
)

func benchTime(tb testing.TB, y, m, d int) time.Time {
	tb.Helper()
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, getCst())
}

func BenchmarkSolarToLunarHanQing(b *testing.B) {
	date := benchTime(b, 1000, 7, 7)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := SolarToLunar(date); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSolarToLunarModern(b *testing.B) {
	date := benchTime(b, 2025, 1, 1)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := SolarToLunar(date); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLunarToSolarText(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := LunarToSolar("元丰六年十月十二"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLunarToSolarByYMD(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := LunarToSolarByYMD(1083, 10, 12, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTimeAdd(b *testing.B) {
	t, err := LunarToSolarByYMD(2024, 1, 1, false)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = t.Add(time.Hour)
	}
}

func BenchmarkCalendricalJieQi(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := CalendricalJieQi(1000, JQ_冬至); err != nil {
			b.Fatal(err)
		}
	}
}

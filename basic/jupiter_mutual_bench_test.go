package basic

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkJupiterGalileanMutualNextSeasonal(benchmark *testing.B) {
	for _, phenomenon := range []JupiterGalileanMutualPhenomenonType{JupiterGalileanMutualOccultation, JupiterGalileanMutualEclipse} {
		for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
			benchmark.Run(fmt.Sprintf("%s/%d-%d", phenomenon, pair[0], pair[1]), func(benchmark *testing.B) {
				benchmark.StopTimer()
				var seeds []float64
				seed := JDCalc(2026, 1, 1)
				first := true
				for {
					event := NextJupiterGalileanMutualEvent(seed, pair[0], pair[1], phenomenon)
					if !event.Valid || event.Greatest >= JDCalc(2028, 1, 1) {
						break
					}
					if !first {
						seeds = append(seeds, seed)
					}
					first = false
					seed = event.End + 0.1/86400
				}
				if len(seeds) == 0 {
					benchmark.Fatal("empty seasonal chain")
				}
				benchmark.ReportAllocs()
				benchmark.ResetTimer()
				benchmark.StartTimer()
				started := time.Now()
				maximum := time.Duration(0)
				var result JupiterGalileanMutualEvent
				for iteration := 0; iteration < benchmark.N; iteration++ {
					for _, seed := range seeds {
						callStarted := time.Now()
						result = NextJupiterGalileanMutualEvent(seed, pair[0], pair[1], phenomenon)
						if elapsed := time.Since(callStarted); elapsed > maximum {
							maximum = elapsed
						}
						if !result.Valid {
							benchmark.Fatal("missing seasonal event")
						}
					}
				}
				elapsed := time.Since(started)
				benchmark.StopTimer()
				benchmark.ReportMetric(float64(elapsed.Nanoseconds())/float64(benchmark.N*len(seeds)), "ns/Next")
				benchmark.ReportMetric(float64(maximum.Nanoseconds()), "max-ns/Next")
				benchmark.ReportMetric(float64(len(seeds)), "calls/op")
				_ = result
			})
		}
	}
}

func BenchmarkJupiterGalileanMutualNextLongInterval(benchmark *testing.B) {
	for _, phenomenon := range []JupiterGalileanMutualPhenomenonType{JupiterGalileanMutualOccultation, JupiterGalileanMutualEclipse} {
		for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
			for _, date := range []struct {
				name string
				seed float64
			}{{"YearStart", JDCalc(2026, 1, 1)}, {"OffSeason", JDCalc(2030, 6, 1)}} {
				benchmark.Run(fmt.Sprintf("%s/%d-%d/%s", phenomenon, pair[0], pair[1], date.name), func(benchmark *testing.B) {
					benchmark.ReportAllocs()
					var result JupiterGalileanMutualEvent
					for iteration := 0; iteration < benchmark.N; iteration++ {
						result = NextJupiterGalileanMutualEvent(date.seed, pair[0], pair[1], phenomenon)
					}
					benchmark.StopTimer()
					if !result.Valid || result.Greatest-date.seed <= 30 {
						benchmark.Fatal("long-interval fixture invalid")
					}
				})
			}
		}
	}
	for _, fixture := range []struct {
		name       string
		seed       float64
		first      int
		second     int
		phenomenon JupiterGalileanMutualPhenomenonType
	}{
		{"Tail/Occultation/1-2", JDCalc(2026, 10, 3), 1, 2, JupiterGalileanMutualOccultation},
		{"Tail/Eclipse/1-2", JDCalc(2026, 12, 6), 1, 2, JupiterGalileanMutualEclipse},
		{"Tail/Occultation/2-3", JDCalc(2027, 8, 10), 2, 3, JupiterGalileanMutualOccultation},
		{"Tail/Eclipse/2-3", JDCalc(2027, 8, 10), 2, 3, JupiterGalileanMutualEclipse},
	} {
		benchmark.Run(fixture.name, func(benchmark *testing.B) {
			benchmark.StopTimer()
			last := NextJupiterGalileanMutualEvent(fixture.seed, fixture.first, fixture.second, fixture.phenomenon)
			if !last.Valid || last.Greatest-fixture.seed >= 1 {
				benchmark.Fatal("season-tail fixture invalid")
			}
			seed := last.End + 0.1/86400
			benchmark.ReportAllocs()
			benchmark.ResetTimer()
			benchmark.StartTimer()
			var result JupiterGalileanMutualEvent
			for iteration := 0; iteration < benchmark.N; iteration++ {
				result = NextJupiterGalileanMutualEvent(seed, fixture.first, fixture.second, fixture.phenomenon)
			}
			benchmark.StopTimer()
			if !result.Valid || result.Greatest-seed <= 30 {
				benchmark.Fatal("season-tail next event invalid")
			}
		})
	}
}

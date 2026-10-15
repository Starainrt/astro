package basic

// 月缘路径基准：与同一调用的圆月面版本成对，改动扫描步长或剪枝时对照两者倍率。

import (
	"testing"
	"time"

	"github.com/starainrt/astro/moon/limb"
)

func benchmarkLocalSolarEclipseLimb(b *testing.B, profile LunarLimb) {
	seed := Date2JD(time.Date(2035, 9, 2, 0, 0, 0, 0, time.UTC))
	options := SolarEclipseOptions{RadiusModel: SolarEclipseModelNASABulletinSplitK, LunarLimb: profile}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		LocalSolarEclipseWithOptionsWithInstance(seed, 116.2310, 39.70, 0, options, nil)
	}
}

func BenchmarkLocalSolarEclipseCircular20350902(b *testing.B) { benchmarkLocalSolarEclipseLimb(b, nil) }

func BenchmarkLocalSolarEclipseLunarLimb20350902(b *testing.B) {
	benchmarkLocalSolarEclipseLimb(b, limb.Default())
}

func benchmarkStarOccultationLimb(b *testing.B, profile LunarLimb) {
	start := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	options := OccultationSearchOptions{LunarLimb: profile}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := FindStarOccultationsWithInstance(start, start.Add(36*time.Hour), limbTestAntares, 41.5907, -55.07036, 0, options, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStarOccultationCircularAntares(b *testing.B) { benchmarkStarOccultationLimb(b, nil) }

func BenchmarkStarOccultationLunarLimbAntares(b *testing.B) {
	benchmarkStarOccultationLimb(b, limb.Default())
}

func benchmarkPlanetOccultationLimb(b *testing.B, profile LunarLimb) {
	start := time.Date(2024, time.August, 21, 1, 30, 0, 0, time.UTC)
	end := time.Date(2024, time.August, 21, 4, 0, 0, 0, time.UTC)
	options := OccultationSearchOptions{LunarLimb: profile}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := FindPlanetOccultationsWithInstance(start, end, OccultationSaturn, -30.072, 43.562592, 0, options, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPlanetOccultationCircularSaturnGraze(b *testing.B) {
	benchmarkPlanetOccultationLimb(b, nil)
}

func BenchmarkPlanetOccultationLunarLimbSaturnGraze(b *testing.B) {
	benchmarkPlanetOccultationLimb(b, limb.Default())
}

func BenchmarkStarOccultationLimbProfileAntares(b *testing.B) {
	info := StarOccultationInfo{Greatest: time.Date(2026, 2, 11, 3, 47, 15, 0, time.UTC), Observer: Observer{Longitude: 41.5907, Latitude: -55.07036}}
	profile := limb.Default()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := StarOccultationLimbProfileWithInstance(info, limbTestAntares, profile, StarOccultationLimbProfileOptions{}, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkSolarEclipseCentralPathLimb(b *testing.B, profile LunarLimb) {
	seed := Date2JD(time.Date(2035, 9, 2, 0, 0, 0, 0, time.UTC))
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		SolarEclipseCentralPathNASABulletinSplitKWithInstance(seed, SolarEclipsePathOptions{LunarLimb: profile}, nil)
	}
}

func BenchmarkSolarEclipseCentralPathCircular20350902(b *testing.B) {
	benchmarkSolarEclipseCentralPathLimb(b, nil)
}

func BenchmarkSolarEclipseCentralPathLunarLimb20350902(b *testing.B) {
	benchmarkSolarEclipseCentralPathLimb(b, limb.Default())
}

func benchmarkSolarEclipsePartialFootprintsLimb(b *testing.B, profile LunarLimb) {
	seed := Date2JD(time.Date(2035, 9, 2, 0, 0, 0, 0, time.UTC))
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		SolarEclipsePartialFootprintsNASABulletinSplitKWithInstance(seed, SolarEclipsePartialFootprintOptions{LunarLimb: profile}, nil)
	}
}

func BenchmarkSolarEclipsePartialFootprintsCircular20350902(b *testing.B) {
	benchmarkSolarEclipsePartialFootprintsLimb(b, nil)
}

func BenchmarkSolarEclipsePartialFootprintsLunarLimb20350902(b *testing.B) {
	benchmarkSolarEclipsePartialFootprintsLimb(b, limb.Default())
}

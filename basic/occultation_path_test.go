package basic

import (
	"testing"
	"time"
)

// 稠密星历表必须被接受并覆盖完整的路径搜索窗口，否则路径样本会静默退回精确星历。
func TestDenseOccultationTableCoversPathSearchWindow(t *testing.T) {
	center := occultationTimeToTT(time.Date(2025, time.January, 5, 12, 0, 0, 0, time.UTC))
	const toleranceDays = 1e-9
	for _, planet := range []OccultationPlanet{
		OccultationMercury, OccultationVenus, OccultationMars, OccultationJupiter,
		OccultationSaturn, OccultationUranus, OccultationNeptune,
	} {
		config, ok := planetOccultationConfigFor(planet)
		if !ok {
			t.Fatalf("%s occultation config is unavailable", planet)
		}
		cache := newPlanetOccultationEventCache(config)
		cache.preparePathEphemeris(center, OccultationPathAlgorithmOptimized)
		if !cache.local.dense {
			t.Fatalf("%s dense path table was rejected", planet)
		}
		nodes := cache.local.nodes
		if len(nodes) < 2 {
			t.Fatalf("%s dense path table has %d nodes", planet, len(nodes))
		}
		if first := center - occultationPathSearchSpanDays; nodes[0].tt > first+toleranceDays {
			t.Fatalf("%s dense table starts at %.9f, want at most %.9f", planet, nodes[0].tt, first)
		}
		if last := center + occultationPathSearchSpanDays; nodes[len(nodes)-1].tt < last-toleranceDays {
			t.Fatalf("%s dense table ends at %.9f, want at least %.9f", planet, nodes[len(nodes)-1].tt, last)
		}
		// 节点时刻是 JD 浮点相加的结果，相邻差只近似等于步长（约 1e-10 天）。
		if step := nodes[1].tt - nodes[0].tt; step < occultationDenseEphemerisStepDays-1e-8 ||
			step > occultationDenseEphemerisStepDays+1e-8 {
			t.Fatalf("%s dense table step=%g, want %g", planet, step, occultationDenseEphemerisStepDays)
		}
	}
}

func BenchmarkOccultationPlanetPathsMars20250729(b *testing.B) {
	zone := time.FixedZone("UTC+8", 8*3600)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	options := OccultationPathOptions{
		Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
		DisableFootprints: true, IncludeFootprintTimeline: true, FootprintTimelineStep: 5 * time.Minute,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		paths, err := FindPlanetOccultationPaths(start, start.Add(24*time.Hour), OccultationMars, options)
		if err != nil || len(paths) != 1 {
			b.Fatalf("paths=%d err=%v", len(paths), err)
		}
	}
}

func BenchmarkOccultationPlanetPathsSaturnDefault(b *testing.B) {
	start := time.Date(2024, time.August, 21, 0, 0, 0, 0, time.UTC)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		paths, err := FindPlanetOccultationPaths(start, start.Add(24*time.Hour), OccultationSaturn, OccultationPathOptions{})
		if err != nil || len(paths) != 1 {
			b.Fatalf("paths=%d err=%v", len(paths), err)
		}
	}
}

func BenchmarkOccultationPathLimitsAndWidth(b *testing.B) {
	config, _ := planetOccultationConfigFor(OccultationSaturn)
	cache := newPlanetOccultationEventCache(config)
	tt := occultationTimeToTT(time.Date(2025, time.January, 5, 4, 0, 0, 0, time.UTC))
	cache.preparePathEphemeris(tt, OccultationPathAlgorithmOptimized)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, _, _, _ = occultationPathLimitsAndWidthForFrame(tt, cache.outerFrameAt)
	}
}

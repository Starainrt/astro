package svg

import (
	"testing"
	"time"

	"github.com/starainrt/astro/moon"
)

func BenchmarkStarOccultationSVGHR4799(b *testing.B) {
	path := starOccultationSimplificationPath(b)
	for _, tolerance := range []float64{0, occultationGeometryTolerancePixels} {
		b.Run(occultationBenchmarkToleranceName(tolerance), func(b *testing.B) {
			occultationBenchmarkRender(b, tolerance, func() (string, error) {
				return StarOccultationPathSVG(path, StarOccultationSVGOptions{Width: 720, Height: 520})
			}, "occultation-band")
		})
	}
}

func BenchmarkPlanetOccultationSVGSaturn20240725(b *testing.B) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
	)
	if err != nil || len(paths) != 1 {
		b.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	for _, tolerance := range []float64{0, occultationGeometryTolerancePixels} {
		b.Run(occultationBenchmarkToleranceName(tolerance), func(b *testing.B) {
			occultationBenchmarkRender(b, tolerance, func() (string, error) {
				return PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{Width: 1200, Height: 800})
			}, "occultation-band")
		})
	}
}

func occultationBenchmarkToleranceName(tolerance float64) string {
	if tolerance <= 0 {
		return "unsimplified"
	}
	return "simplified"
}

func occultationBenchmarkRender(
	b *testing.B,
	tolerance float64,
	render func() (string, error),
	className string,
) {
	b.Helper()
	previous := occultationGeometryTolerancePixels
	occultationGeometryTolerancePixels = tolerance
	defer func() { occultationGeometryTolerancePixels = previous }()
	diagram, err := render()
	if err != nil {
		b.Fatalf("render: %v", err)
	}
	b.ReportMetric(float64(len(diagram)), "doc-bytes")
	b.ReportMetric(float64(occultationPathVertexCount(planetOccultationSVGPathData(b, diagram, className))), "band-vertices")
	b.ResetTimer()
	for attempt := 0; attempt < b.N; attempt++ {
		if _, err := render(); err != nil {
			b.Fatalf("render: %v", err)
		}
	}
}

// BenchmarkStarOccultationSimplificationSpeedup 在同一次基准里测原始与简化两档容差，
// 把渲染耗时对比从单元测试移到基准，并直接报告加速比。
func BenchmarkStarOccultationSimplificationSpeedup(b *testing.B) {
	path := starOccultationSimplificationPath(b)
	render := func() (string, error) {
		return StarOccultationPathSVG(path, StarOccultationSVGOptions{Width: 720, Height: 520})
	}
	raw := occultationBenchmarkDuration(b, 0, render)
	simplified := occultationBenchmarkDuration(b, occultationGeometryTolerancePixels, render)
	b.ReportMetric(float64(raw.Nanoseconds()), "raw-ns/op")
	b.ReportMetric(float64(simplified.Nanoseconds()), "simplified-ns/op")
	b.ReportMetric(float64(raw)/float64(simplified), "speedup")
}

// BenchmarkPlanetOccultationSimplificationSpeedup 是行星路径的同一对比。
func BenchmarkPlanetOccultationSimplificationSpeedup(b *testing.B) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
	)
	if err != nil || len(paths) != 1 {
		b.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	render := func() (string, error) {
		return PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{Width: 1200, Height: 800})
	}
	raw := occultationBenchmarkDuration(b, 0, render)
	simplified := occultationBenchmarkDuration(b, occultationGeometryTolerancePixels, render)
	b.ReportMetric(float64(raw.Nanoseconds()), "raw-ns/op")
	b.ReportMetric(float64(simplified.Nanoseconds()), "simplified-ns/op")
	b.ReportMetric(float64(raw)/float64(simplified), "speedup")
}

func occultationBenchmarkDuration(b *testing.B, tolerance float64, render func() (string, error)) time.Duration {
	b.Helper()
	previous := occultationGeometryTolerancePixels
	occultationGeometryTolerancePixels = tolerance
	defer func() { occultationGeometryTolerancePixels = previous }()
	begin := time.Now()
	for attempt := 0; attempt < b.N; attempt++ {
		if _, err := render(); err != nil {
			b.Fatalf("render: %v", err)
		}
	}
	return time.Since(begin) / time.Duration(b.N)
}

// BenchmarkPlanetOccultationSVGFootprintBands 对比 DisableFootprints 两档：
// 逐条绘制密集瞬时足迹 vs 绘制紧凑带的合并结果（命中只读缓存）。
func BenchmarkPlanetOccultationSVGFootprintBands(b *testing.B) {
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	for _, disable := range []bool{false, true} {
		name := "dense-footprints"
		if disable {
			name = "compact-band"
		}
		b.Run(name, func(b *testing.B) {
			paths, err := moon.FindPlanetOccultationPaths(
				start, start.Add(24*time.Hour), moon.OccultationSaturn,
				moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200, DisableFootprints: disable},
			)
			if err != nil || len(paths) != 1 {
				b.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
			}
			path := paths[0]
			render := func() (string, error) {
				return PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{Width: 1200, Height: 800})
			}
			if _, err := render(); err != nil {
				b.Fatalf("warm-up render: %v", err)
			}
			b.ResetTimer()
			for attempt := 0; attempt < b.N; attempt++ {
				if _, err := render(); err != nil {
					b.Fatalf("render: %v", err)
				}
			}
		})
	}
}

// BenchmarkPlanetOccultationSVGCompactBandCold 是紧凑带合并缓存未命中时的成本，
// 用于说明首次渲染仍要付出一次合并代价。
func BenchmarkPlanetOccultationSVGCompactBandCold(b *testing.B) {
	start := time.Date(2024, time.July, 25, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationSaturn,
		moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200, DisableFootprints: true},
	)
	if err != nil || len(paths) != 1 {
		b.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	b.ResetTimer()
	for attempt := 0; attempt < b.N; attempt++ {
		occultationBandCache.Lock()
		occultationBandCache.entries = make(map[occultationBandCacheKey]occultationBandPolygons, occultationBandCacheCapacity)
		occultationBandCache.order = nil
		occultationBandCache.Unlock()
		if _, err := PlanetOccultationPathSVG(path, PlanetOccultationSVGOptions{Width: 1200, Height: 800}); err != nil {
			b.Fatalf("render: %v", err)
		}
	}
}

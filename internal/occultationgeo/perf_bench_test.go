package occultationgeo

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

var (
	saturn20250105Once  sync.Once
	saturn20250105Paths []basic.PlanetOccultationPath
)

func saturn20250105Path(tb testing.TB) basic.PlanetOccultationPath {
	tb.Helper()
	saturn20250105Once.Do(func() {
		zone := time.FixedZone("UTC+8", 8*60*60)
		start := time.Date(2025, time.January, 5, 0, 0, 0, 0, zone)
		paths, err := basic.FindPlanetOccultationPaths(
			start, start.Add(24*time.Hour), basic.OccultationSaturn,
			basic.OccultationPathOptions{
				Step: 20 * time.Minute, TargetSpacingKM: 900,
				DisableFootprints: true, IncludeFootprintTimeline: true,
				FootprintTimelineStep: 5 * time.Minute, RiseSetStep: time.Minute,
			},
		)
		if err != nil || len(paths) != 1 {
			tb.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v", len(paths), err)
		}
		saturn20250105Paths = paths
	})
	return saturn20250105Paths[0]
}

func BenchmarkVisibleBandPolygonsSaturn20250105(b *testing.B) {
	path := saturn20250105Path(b)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if _, _, err := VisibleBandPolygonsFromContours(
			path.PartialBandFootprints, path.PartialBandContours,
			path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
		); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVisibleBandPolygonsAnalyticSaturn20250105(b *testing.B) {
	path := saturn20250105Path(b)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if _, _, err := VisibleBandPolygonsFromAnalyticContours(
			path.PartialBandFootprints, path.PartialBandContours, path.PartialVisibilityContours,
			path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
		); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFootprintSweepPolygonsSaturn20250105(b *testing.B) {
	path := saturn20250105Path(b)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if _, err := footprintSweepPolygons(
			path.PartialBandFootprints, path.NorthernLimit, path.SouthernLimit,
		); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFootprintOpenSweepSaturn20250105(b *testing.B) {
	path := saturn20250105Path(b)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if _, err := footprintOpenSweepPolygons(path.PartialBandFootprints); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFootprintVisibleUnionSaturn20250105(b *testing.B) {
	path := saturn20250105Path(b)
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if polygons := footprintVisibleUnionPolygons(path.PartialBandFootprints); len(polygons) == 0 {
			b.Fatal("empty visible union")
		}
	}
}

func occultationBenchChainPolygons(count, points int) [][]geodata.GeoPoint {
	polygons := make([][]geodata.GeoPoint, 0, count)
	for index := 0; index < count; index++ {
		center := geodata.GeoPoint{
			Longitude: -0.5 + float64(index)*1.5,
			Latitude:  40,
		}
		ring := make([]geodata.GeoPoint, 0, points+1)
		for step := 0; step <= points; step++ {
			angle := 2 * math.Pi * float64(step) / float64(points)
			ring = append(ring, geodata.GeoPoint{
				Longitude: center.Longitude + 0.5*math.Cos(angle),
				Latitude:  center.Latitude + 0.5*math.Sin(angle),
			})
		}
		polygons = append(polygons, ring)
	}
	return polygons
}

func BenchmarkMergeTouchingVisiblePolygonsChain(b *testing.B) {
	b.Run("count=12/points=16", func(b *testing.B) {
		for iteration := 0; iteration < b.N; iteration++ {
			polygons := occultationBenchChainPolygons(12, 16)
			if merged := mergeTouchingVisiblePolygons(polygons); len(merged) != 1 {
				b.Fatalf("merged=%d, want one face", len(merged))
			}
		}
	})
	b.Run("count=24/points=16", func(b *testing.B) {
		for iteration := 0; iteration < b.N; iteration++ {
			polygons := occultationBenchChainPolygons(24, 16)
			if merged := mergeTouchingVisiblePolygons(polygons); len(merged) != 1 {
				b.Fatalf("merged=%d, want one face", len(merged))
			}
		}
	})
}

func occultationBenchConstrainInput(points int) (parent, child [][]geodata.GeoPoint) {
	ring := make([]geodata.GeoPoint, 0, points+1)
	childRing := make([]geodata.GeoPoint, 0, points+1)
	for step := 0; step <= points; step++ {
		angle := 2 * math.Pi * float64(step) / float64(points)
		longitude := -70 + 20*math.Cos(angle)
		latitude := -40 + 8*math.Sin(angle)
		ring = append(ring, geodata.GeoPoint{Longitude: longitude, Latitude: latitude})
		childRing = append(childRing, geodata.GeoPoint{
			Longitude: longitude + 0.02*math.Cos(angle),
			Latitude:  latitude + 0.02*math.Sin(angle),
		})
	}
	return [][]geodata.GeoPoint{ring}, [][]geodata.GeoPoint{childRing}
}

func BenchmarkConstrainPolygonsWithinSmallBreach(b *testing.B) {
	b.Run("points=512", func(b *testing.B) {
		parent, child := occultationBenchConstrainInput(512)
		b.ResetTimer()
		for iteration := 0; iteration < b.N; iteration++ {
			if repaired := ConstrainPolygonsWithin(parent, child); len(repaired) == 0 {
				b.Fatal("empty repair")
			}
		}
	})
}

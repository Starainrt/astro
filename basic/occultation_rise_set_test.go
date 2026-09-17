package basic

import (
	"reflect"
	"testing"
	"time"
)

// 折点补根的判据只能统计查询窗口内部的未成对端点：窗口边界上的裁剪端点没有任何折点
// 可以配对，把它们计入会让“未成对端点数严格下降”永远不成立。
func TestOccultationRiseSetFoldRecoveryCountsInteriorEndpoints(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*3600)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	options := OccultationPathOptions{
		Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
		DisableFootprints: true,
	}
	paths, err := FindPlanetOccultationPaths(start, start.Add(24*time.Hour), OccultationMars, options)
	if err != nil || len(paths) != 1 || !paths[0].HasTotalBand {
		t.Fatalf("paths=%d err=%v, want one total-band event", len(paths), err)
	}
	path := paths[0]
	config, _ := planetOccultationConfigFor(OccultationMars)
	cache := newPlanetOccultationEventCache(config)
	greatestTT := occultationTimeToTT(path.Greatest.Time)
	cache.preparePathEphemeris(greatestTT, OccultationPathAlgorithmOptimized)
	firstTT := occultationTimeToTT(path.TotalStart.Time)
	lastTT := occultationTimeToTT(path.TotalEnd.Time)
	curves, base, recovered := occultationRiseSetCurvesWithRecoveryReport(
		firstTT, lastTT, greatestTT, options, start.Location(), cache.totalRiseSetCache,
	)
	if recovered == nil {
		t.Fatal("fold recovery was never attempted on a graph with interior unpaired endpoints")
	}
	baseInterior := occultationRiseSetUnclosedEndpointCount(base, true, firstTT, lastTT)
	recoveredInterior := occultationRiseSetUnclosedEndpointCount(recovered, true, firstTT, lastTT)
	if baseInterior == 0 || recoveredInterior >= baseInterior {
		t.Fatalf("interior unpaired endpoints base=%d recovered=%d, want a strict decrease",
			baseInterior, recoveredInterior)
	}
	// 全局计数被窗口裁剪端点抬平：按它判定会丢弃这次重建。
	baseGlobal := occultationRiseSetUnclosedEndpointCount(base, false, firstTT, lastTT)
	recoveredGlobal := occultationRiseSetUnclosedEndpointCount(recovered, false, firstTT, lastTT)
	if recoveredGlobal < baseGlobal {
		t.Fatalf("global unpaired endpoints base=%d recovered=%d, sample no longer shows a window-truncated endpoint",
			baseGlobal, recoveredGlobal)
	}
	if !reflect.DeepEqual(curves, recovered) {
		t.Fatal("interior-endpoint test did not select the rebuilt phase graph")
	}
}

func TestOccultationRiseSetWindowClippedEndpointIgnoredByInteriorCount(t *testing.T) {
	firstTT := 2460000.0
	lastTT := firstTT + 0.25
	point := func(tt float64) OccultationPathPoint {
		return OccultationPathPoint{Time: occultationTTToLocation(tt, time.UTC)}
	}
	curves := []OccultationRiseSetCurve{
		{
			Phase: RiseSetPhaseStart, Direction: RiseSetDirectionRise,
			Segments: [][]OccultationPathPoint{
				{point(lastTT), point(lastTT + 0.01)},
				{point(firstTT + 0.02), point(firstTT + 0.03)},
			},
		},
	}
	if got := occultationRiseSetUnclosedEndpointCount(curves, false, firstTT, lastTT); got != 4 {
		t.Fatalf("global unpaired endpoints=%d, want 4", got)
	}
	if got := occultationRiseSetUnclosedEndpointCount(curves, true, firstTT, lastTT); got != 2 {
		t.Fatalf("interior unpaired endpoints=%d, want 2", got)
	}
}

// stateAt 的分配预算：同一上下文内换站点只复用已分配的站点表，不再额外分配。
func BenchmarkOccultationRiseSetStateAt(b *testing.B) {
	context := newOccultationRiseSetContext(
		occultationTimeToTT(time.Date(2025, time.July, 29, 12, 0, 0, 0, time.UTC)),
		100, 10, 384400, 200, -15, 149597870.7*2, 3389.5,
	)
	coordinates := make([][2]float64, 0, 64)
	for index := 0; index < 64; index++ {
		coordinates = append(coordinates, [2]float64{float64(index) * 1.7, float64(index%30) - 15})
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		for _, coordinate := range coordinates {
			_ = context.stateAt(coordinate[0], coordinate[1])
		}
	}
}

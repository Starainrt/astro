package basic

import (
	"testing"
	"time"
)

// TestOccultationPathFrameBoundaryCacheMatchesUncachedSearch 固定 frame 级边界记忆化的语义：
// 记忆化只是复用同一 frame 上纯函数的计算结果，因此切点、可见 θ 区间与按中心角的 θ 区间
// 都必须与直接调用未缓存实现逐位一致，重复调用也必须稳定。
// TestOccultationPathFrameBoundaryCacheMatchesUncachedSearch pins the frame-level boundary memo
// semantics: memoization only reuses pure results for the same frame, so the tangency, the
// visible theta intervals and the per-center-angle interval must stay bit-identical to the
// uncached implementations, and repeated calls must be stable.
func TestOccultationPathFrameBoundaryCacheMatchesUncachedSearch(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2024, time.August, 21, 2, 41, 36, 0, time.UTC))
	cache := newPlanetOccultationEventCache(config)

	cachedFrame, cachedOK := cache.outerFrameAt(tt)
	directFrame, directOK := planetOccultationPathFrameAt(tt, config)
	if cachedOK != directOK || !cachedOK {
		t.Fatalf("frame availability cached=%v direct=%v, want both true", cachedOK, directOK)
	}
	if !occultationPathFrameGeometryEqual(cachedFrame, directFrame) {
		t.Fatal("cached frame geometry differs from a direct frame")
	}
	if cachedFrame.boundary == nil {
		t.Fatal("cached frame is missing its boundary memo")
	}

	wantPoint, wantTheta, wantOK := occultationPathBoundaryTangentUncached(directFrame)
	gotPoint, gotTheta, gotOK := occultationPathBoundaryTangent(cachedFrame)
	if gotOK != wantOK || gotPoint != wantPoint || gotTheta != wantTheta {
		t.Fatalf("memoized tangency = (%.15g, %.15g, %v), want (%.15g, %.15g, %v)",
			gotPoint, gotTheta, gotOK, wantPoint, wantTheta, wantOK)
	}
	againPoint, againTheta, againOK := occultationPathBoundaryTangent(cachedFrame)
	if againOK != gotOK || againPoint != gotPoint || againTheta != gotTheta {
		t.Fatal("repeated tangency query is not stable")
	}

	wantIntervals := occultationPathBoundaryThetaIntervalsUncached(directFrame)
	gotIntervals := occultationPathBoundaryThetaIntervals(cachedFrame)
	if len(gotIntervals) != len(wantIntervals) {
		t.Fatalf("memoized theta intervals = %v, want %v", gotIntervals, wantIntervals)
	}
	for index := range gotIntervals {
		if gotIntervals[index] != wantIntervals[index] {
			t.Fatalf("memoized theta interval %d = %v, want %v", index, gotIntervals[index], wantIntervals[index])
		}
	}
	if againIntervals := occultationPathBoundaryThetaIntervals(cachedFrame); len(againIntervals) != len(gotIntervals) {
		t.Fatal("repeated theta-interval query is not stable")
	}

	cachedLeft, cachedRight, cachedIntervalOK := occultationPathBoundaryThetaInterval(cachedFrame, wantTheta)
	directLeft, directRight, directIntervalOK := occultationPathBoundaryThetaInterval(directFrame, wantTheta)
	if cachedIntervalOK != directIntervalOK || cachedLeft != directLeft || cachedRight != directRight {
		t.Fatalf("memoized interval = (%.15g, %.15g, %v), want (%.15g, %.15g, %v)",
			cachedLeft, cachedRight, cachedIntervalOK, directLeft, directRight, directIntervalOK)
	}
	repeatLeft, repeatRight, repeatOK := occultationPathBoundaryThetaInterval(cachedFrame, wantTheta)
	if repeatOK != cachedIntervalOK || repeatLeft != cachedLeft || repeatRight != cachedRight {
		t.Fatal("repeated per-angle interval query is not stable")
	}
}

// TestOccultationPathZeroFrameComputesBoundaryWithoutCache 保证未挂记忆化的零值 frame 仍然
// 直接计算，不会被 nil 缓存跳过。
// TestOccultationPathZeroFrameComputesBoundaryWithoutCache keeps a frame without a memo working:
// a nil boundary cache must fall back to direct computation, not skip it.
func TestOccultationPathZeroFrameComputesBoundaryWithoutCache(t *testing.T) {
	frame := occultationPathFrame{
		moon:         occultationPathVector{x: 384000},
		axis:         occultationPathVector{x: -1},
		first:        occultationPathVector{y: 1},
		second:       occultationPathVector{z: 1},
		moonRadius:   0.0045,
		targetRadius: 1e-5,
	}
	if frame.boundary != nil {
		t.Fatal("fixture frame unexpectedly carries a boundary memo")
	}
	point, theta, ok := occultationPathBoundaryTangent(frame)
	wantPoint, wantTheta, wantOK := occultationPathBoundaryTangentUncached(frame)
	if ok != wantOK || point != wantPoint || theta != wantTheta {
		t.Fatalf("uncached tangency = (%.15g, %.15g, %v), want (%.15g, %.15g, %v)",
			point, theta, ok, wantPoint, wantTheta, wantOK)
	}
}

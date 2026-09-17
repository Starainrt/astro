package occultationgeo

import (
	"errors"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

// 分裂轮廓场景的接触线网互不连续（离扫掠边界上千公里），无法充当见证；该分支的
// 权威性来自"单分量 + 同一批瞬时可见面并集"的构造契约，此处把这条契约钉住。
func TestSplitContourBandUsesSingleComponentSweepContract(t *testing.T) {
	path := saturn20250105Path(t)
	if len(path.PartialBandContours) <= 2 {
		t.Fatalf("contours=%d, want the split-contour branch input", len(path.PartialBandContours))
	}
	swept, err := footprintSweepPolygons(
		path.PartialBandFootprints, path.NorthernLimit, path.SouthernLimit,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(swept) != 1 {
		t.Fatalf("sweep components=%d, want the single-component contract", len(swept))
	}
	if occultationSweepWitnessedByContactContours(swept, path.PartialBandContours) {
		t.Fatal("fragmented contact contours unexpectedly witness the sweep; revisit the split-contour branch")
	}
	polygons, authoritative, err := VisibleBandPolygonsFromContours(
		path.PartialBandFootprints, path.PartialBandContours,
		path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
	)
	if err != nil || !authoritative || len(polygons) != 1 {
		t.Fatalf("split-contour band polygons=%d authoritative=%v err=%v", len(polygons), authoritative, err)
	}
}

func TestHorizonConnectorSegmentsFiltersBranchesOutsideLimits(t *testing.T) {
	footprints, curves := horizonConnectorTestOpeningFootprint()
	if got := HorizonConnectorSegments(footprints, curves); len(got) != 1 {
		t.Fatalf("connectors without limits=%d, want one baseline closure", len(got))
	}
	nearLimit := []basic.OccultationPathPoint{
		{Longitude: 0, Latitude: 0},
		{Longitude: 0, Latitude: 3},
	}
	if got := HorizonConnectorSegments(footprints, curves, nearLimit); len(got) != 1 {
		t.Fatalf("connectors with an attached limit=%d, want the closure kept", len(got))
	}
	farLimit := []basic.OccultationPathPoint{
		{Longitude: 60, Latitude: 0},
		{Longitude: 60, Latitude: 3},
	}
	if got := HorizonConnectorSegments(footprints, curves, farLimit); len(got) != 0 {
		t.Fatalf("connectors with a detached limit=%d, want the branch filtered out", len(got))
	}
}

func TestHorizonConnectorSegmentsKeepsSamePhaseBranchOpeningWithLimits(t *testing.T) {
	eventTime := horizonConnectorTestEventTime()
	contactStart := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 0}
	contactEnd := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 1}
	footprint := basic.OccultationFootprint{
		Boundaries: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 2, Latitude: 0.5},
			contactEnd,
		}},
		Polygons: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 2, Latitude: 0.5},
			contactEnd,
			{Time: eventTime, Longitude: 1, Latitude: 1.2},
			{Time: eventTime, Longitude: 1, Latitude: -0.2},
			contactStart,
		}},
	}
	for _, test := range []struct {
		name      string
		second    basic.OccultationPathPoint
		wantCount int
	}{
		{name: "separated-branch", second: basic.OccultationPathPoint{Time: eventTime, Longitude: 0.3, Latitude: 0.5}, wantCount: 1},
		{name: "same-phase-fold", second: basic.OccultationPathPoint{Time: eventTime, Longitude: 0.02, Latitude: 0.5}, wantCount: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			curves := []basic.OccultationRiseSetCurve{{
				Phase: basic.RiseSetPhaseStart, Direction: basic.RiseSetDirectionRise,
				Segments: [][]basic.OccultationPathPoint{
					{{Time: eventTime, Longitude: 0, Latitude: 0.5}, {Time: eventTime, Longitude: -2, Latitude: 0.5}},
					{test.second, {Time: eventTime, Longitude: -2, Latitude: 0.6}},
				},
			}}
			if got := HorizonConnectorSegments([]basic.OccultationFootprint{footprint}, curves); len(got) != test.wantCount {
				t.Fatalf("connectors=%d, want %d", len(got), test.wantCount)
			}
		})
	}
}

// 通用并集失败后恢复阶梯必须真正重试，而不是直接返回未合并的面。
func TestOccultationRetryOpenSweepUnionRecoversUnion(t *testing.T) {
	footprint := basic.OccultationFootprint{
		Closed: true,
		Polygons: [][]basic.OccultationPathPoint{{
			{Longitude: 0, Latitude: 0}, {Longitude: 1, Latitude: 0},
			{Longitude: 1, Latitude: 1}, {Longitude: 0, Latitude: 1}, {Longitude: 0, Latitude: 0},
		}},
	}
	broken := [][]geodata.GeoPoint{{{Longitude: 0, Latitude: 0}, {Longitude: 1, Latitude: 1}}}
	if _, err := geodata.UnionPolygons(broken); err == nil {
		t.Skip("primary union no longer fails on a two-point ring")
	}
	original := errors.New("primary union failed")
	merged, err := occultationRetryOpenSweepUnion(
		broken, []basic.OccultationFootprint{footprint}, nil, nil, original,
	)
	if err != nil {
		t.Fatalf("recovery ladder returned the primary error: %v", err)
	}
	if len(merged) != 1 || len(merged[0]) < 4 {
		t.Fatalf("recovered polygons=%d, want the closed instantaneous face", len(merged))
	}
	if _, err := occultationRetryOpenSweepUnion(broken, nil, nil, nil, original); !errors.Is(err, original) {
		t.Fatalf("ladder without usable faces returned %v, want the primary error", err)
	}
}

// 逐边修复已把残差压小时必须保留细化结果，不能退回未修复的 child。
func TestConstrainPolygonsWithinKeepsPartialEdgeRepair(t *testing.T) {
	// 凹槽父带：子环顶点都在父带内，只有跨越凹槽的长边落在外面，顶点吸附无从下手。
	parent := [][]geodata.GeoPoint{{
		{Longitude: -2, Latitude: -3}, {Longitude: 2, Latitude: -3},
		{Longitude: 2, Latitude: 2}, {Longitude: 1.2, Latitude: 2},
		{Longitude: 1.2, Latitude: -0.2}, {Longitude: -1.2, Latitude: -0.2},
		{Longitude: -1.2, Latitude: 2}, {Longitude: -2, Latitude: 2},
		{Longitude: -2, Latitude: -3},
	}}
	child := [][]geodata.GeoPoint{{
		{Longitude: -1.5, Latitude: 0.2}, {Longitude: 1.5, Latitude: 0.2},
		{Longitude: 1.5, Latitude: 0.4}, {Longitude: -1.5, Latitude: 0.4},
		{Longitude: -1.5, Latitude: 0.2},
	}}
	initial := geodata.SphericalPolygonsPathMissDistanceKM(parent, child, true)
	if initial <= 10 {
		t.Fatalf("test child misses the parent by only %.3f km, want an edge-only breach", initial)
	}
	repaired := ConstrainPolygonsWithin(parent, differentialCopyPolygons(child))
	repairedMiss := geodata.SphericalPolygonsPathMissDistanceKM(parent, repaired, true)
	if repairedMiss >= initial*0.9 {
		t.Fatalf("edge repair did not substantially reduce the breach: %.3f km -> %.3f km", initial, repairedMiss)
	}
	if len(repaired) != 1 || len(repaired[0]) <= len(child[0]) {
		t.Fatalf("repaired rings=%d points=%d, want the densified ring", len(repaired), len(repaired[0]))
	}
	discarded := constrainPolygonsWithinReference(parent, differentialCopyPolygons(child))
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(parent, discarded, true); miss != initial {
		t.Fatalf("reference implementation unexpectedly repaired the child to %.3f km", miss)
	}
}

func horizonConnectorTestEventTime() time.Time {
	return time.Date(2025, time.July, 29, 0, 0, 0, 0, time.UTC)
}

func horizonConnectorTestOpeningFootprint() ([]basic.OccultationFootprint, []basic.OccultationRiseSetCurve) {
	eventTime := horizonConnectorTestEventTime()
	contactStart := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 0}
	contactEnd := basic.OccultationPathPoint{Time: eventTime, Longitude: 0, Latitude: 3}
	branchStart := basic.OccultationPathPoint{Time: eventTime, Longitude: -0.05, Latitude: 0}
	branchEnd := basic.OccultationPathPoint{Time: eventTime, Longitude: -0.05, Latitude: 3}
	footprints := []basic.OccultationFootprint{{
		Boundaries: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 2, Latitude: 1.5},
			contactEnd,
		}},
		Polygons: [][]basic.OccultationPathPoint{{
			contactStart,
			{Time: eventTime, Longitude: 2, Latitude: 1.5},
			contactEnd,
			{Time: eventTime, Longitude: 1, Latitude: 3},
			{Time: eventTime, Longitude: 1, Latitude: 0},
			contactStart,
		}},
	}}
	curves := []basic.OccultationRiseSetCurve{{
		Phase: basic.RiseSetPhaseStart, Direction: basic.RiseSetDirectionRise,
		Segments: [][]basic.OccultationPathPoint{
			{branchStart, {Time: eventTime.Add(time.Minute), Longitude: -1, Latitude: -1}},
			{branchEnd, {Time: eventTime.Add(time.Minute), Longitude: -1, Latitude: 4}},
		},
	}}
	return footprints, curves
}

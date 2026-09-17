package geodata

import (
	"math"
	"testing"
)

func TestVisibleLineworkFacesSeparateTouchingLoops(t *testing.T) {
	lines := [][]GeoPoint{
		{{0, 0}, {-2, 0}, {-2, 2}, {0, 2}, {0, 0}},
		{{0, 0}, {2, 0}, {2, -2}, {0, -2}, {0, 0}},
		{{0, 0}, {1, 1}},
	}
	chart := polygonUnionChart{xAxis: geoVector3{x: 1}, yAxis: geoVector3{y: 1}, zAxis: geoVector3{z: 1}}
	nodes, edges, _ := buildVisibleLineworkGraph(lines, chart, 1)
	cycles := enumerateVisibleLineworkCycles(nodes, edges)
	if len(cycles) != 2 {
		t.Fatalf("got %d cycles, want two lobes without the dangling bridge", len(cycles))
	}
	for _, ring := range cycles {
		if len(ring) != 4 || math.Abs(math.Abs(visibleLineworkSignedArea(ring))-4) > 1e-10 {
			t.Fatalf("face contains a repeated articulation or bridge: %+v", ring)
		}
	}
}

func TestVisibleLineworkFaceTraversalIsBounded(t *testing.T) {
	var lines [][]GeoPoint
	for row := 0; row < 12; row++ {
		for col := 0; col < 12; col++ {
			x, y := float64(col), float64(row)
			if col < 11 {
				lines = append(lines, []GeoPoint{{x, y}, {x + 1, y}})
			}
			if row < 11 {
				lines = append(lines, []GeoPoint{{x, y}, {x, y + 1}})
			}
		}
	}
	chart := polygonUnionChart{xAxis: geoVector3{x: 1}, yAxis: geoVector3{y: 1}, zAxis: geoVector3{z: 1}}
	nodes, edges, _ := buildVisibleLineworkGraph(lines, chart, 1)
	cycles := enumerateVisibleLineworkCycles(nodes, edges)
	if len(cycles) != 122 {
		t.Fatalf("got %d cycles, want 121 cells and one exterior", len(cycles))
	}
}

func TestVisibleLineworkKeepsDistinctClosedJunctions(t *testing.T) {
	var lines [][]GeoPoint
	for _, longitude := range []float64{0, 1.001} {
		ring := []GeoPoint{{longitude, 0}, {longitude + 1, 0}, {longitude + 1, 1}, {longitude, 1}}
		for i, point := range ring {
			lines = append(lines, []GeoPoint{point, ring[(i+1)%len(ring)]})
		}
	}
	chart := polygonUnionChart{xAxis: geoVector3{x: 1}, yAxis: geoVector3{y: 1}, zAxis: geoVector3{z: 1}}
	nodes, edges, _ := buildVisibleLineworkGraph(lines, chart, 25)
	if len(nodes) != 8 || len(edges) != 8 {
		t.Fatalf("independent closed junctions merged: nodes=%d edges=%d", len(nodes), len(edges))
	}
}

func TestVisibleLineworkProjectionPreservesSphericalArc(t *testing.T) {
	chart := polygonUnionChart{xAxis: geoVector3{x: 1}, yAxis: geoVector3{y: 1}, zAxis: geoVector3{z: 1}}
	start, end := GeoPoint{-20, 60}, GeoPoint{20, 60}
	line := projectVisibleLine(chart, []GeoPoint{start, end})
	for i := 0; i <= 100; i++ {
		point := chart.project(sphericalInterpolate(start, end, float64(i)/100))
		minimum := math.Inf(1)
		for j := 1; j < len(line); j++ {
			minimum = math.Min(minimum, visibleLineworkPointSegmentDistanceSquared(point, line[j-1], line[j]))
		}
		if minimum > 0.003*0.003 {
			t.Fatalf("projected spherical arc leaves line by %.6f degrees", math.Sqrt(minimum))
		}
	}
}

func TestVisibleLineworkPolygonsSelectsFilledFaces(t *testing.T) {
	lines := [][]GeoPoint{
		{{Longitude: 0, Latitude: 0}, {Longitude: 4, Latitude: 0}},
		{{Longitude: 4, Latitude: 0}, {Longitude: 4, Latitude: 4}},
		{{Longitude: 4, Latitude: 4}, {Longitude: 0, Latitude: 4}},
		{{Longitude: 0, Latitude: 4}, {Longitude: 0, Latitude: 0}},
		{{Longitude: 2, Latitude: 0}, {Longitude: 2, Latitude: 4}},
	}
	fill := [][]GeoPoint{{
		{Longitude: 0, Latitude: 0},
		{Longitude: 4, Latitude: 0},
		{Longitude: 4, Latitude: 4},
		{Longitude: 0, Latitude: 4},
	}}
	polygons, err := VisibleLineworkPolygons(lines, fill, fill, 1)
	if err != nil {
		t.Fatalf("VisibleLineworkPolygons: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("polygon count=%d, want the two selected faces merged", len(polygons))
	}
	for _, point := range []GeoPoint{{Longitude: 1, Latitude: 2}, {Longitude: 3, Latitude: 2}} {
		if !sphericalPolygonContainsOrTouches(polygons[0], point) {
			t.Fatalf("merged visible linework misses %+v", point)
		}
	}
}

func TestVisibleLineworkBoundaryBridgePreservesReverseDirection(t *testing.T) {
	line := []GeoPoint{
		{Longitude: 0, Latitude: 0},
		{Longitude: 1, Latitude: 0},
		{Longitude: 2, Latitude: 0},
		{Longitude: 3, Latitude: 0},
	}
	start, end := line[len(line)-1], line[0]
	bridge, ok := visibleLineworkBoundaryBridge(start, end, [][]GeoPoint{line}, 120)
	if !ok {
		t.Fatal("visibleLineworkBoundaryBridge() returned no bridge")
	}
	if !SameGeoPoint(bridge[0], start) || !SameGeoPoint(bridge[len(bridge)-1], end) {
		t.Fatalf("bridge endpoints=%+v -> %+v, want %+v -> %+v", bridge[0], bridge[len(bridge)-1], start, end)
	}
	if len(bridge) != len(line) {
		t.Fatalf("bridge points=%d, want %d", len(bridge), len(line))
	}
	for index, point := range bridge {
		if !SameGeoPoint(point, line[len(line)-1-index]) {
			t.Fatalf("bridge[%d]=%+v, want reverse source point %+v", index, point, line[len(line)-1-index])
		}
	}
}

func TestVisibleLineworkPolygonsSnapsNearbyJunctions(t *testing.T) {
	lines := [][]GeoPoint{
		{{Longitude: 0, Latitude: 0}, {Longitude: 4, Latitude: 0}},
		{{Longitude: 4.001, Latitude: 0}, {Longitude: 4, Latitude: 4}},
		{{Longitude: 4, Latitude: 4}, {Longitude: 0, Latitude: 4}},
		{{Longitude: 0, Latitude: 4}, {Longitude: 0, Latitude: 0.001}},
	}
	fill := [][]GeoPoint{{
		{Longitude: 0, Latitude: 0},
		{Longitude: 4, Latitude: 0},
		{Longitude: 4, Latitude: 4},
		{Longitude: 0, Latitude: 4},
	}}
	if _, err := VisibleLineworkPolygons(lines, fill, fill, 1); err != nil {
		t.Fatalf("VisibleLineworkPolygons did not snap sub-kilometer junctions: %v", err)
	}
}

func TestVisibleLineworkPolygonsKeepsSnapAndFillAuditTolerancesIndependent(t *testing.T) {
	lines := [][]GeoPoint{
		{{Longitude: 0, Latitude: 0}, {Longitude: 4, Latitude: 0}},
		{{Longitude: 4, Latitude: 0}, {Longitude: 4, Latitude: 4}},
		{{Longitude: 4, Latitude: 4}, {Longitude: 0, Latitude: 4}},
		{{Longitude: 0, Latitude: 4}, {Longitude: 0, Latitude: 0}},
	}
	fill := [][]GeoPoint{{
		{Longitude: -0.02, Latitude: 0},
		{Longitude: 4, Latitude: 0},
		{Longitude: 4, Latitude: 4},
		{Longitude: -0.02, Latitude: 4},
	}}
	coverage := [][]GeoPoint{{{Longitude: 2, Latitude: 2}}}
	if _, err := VisibleLineworkPolygons(lines, fill, coverage, 1); err == nil {
		t.Fatal("default fill audit unexpectedly accepted a 2 km source residual")
	}
	polygons, err := VisibleLineworkPolygonsWithAuditTolerance(lines, fill, coverage, 1, 3)
	if err != nil {
		t.Fatalf("independent fill audit rejected a bounded source residual: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("polygon count=%d, want one without increasing graph snap distance", len(polygons))
	}
}

func TestVisibleLineworkPolygonIndexMatchesPointInPolygon(t *testing.T) {
	polygon := []GeoPoint{
		{Longitude: -2, Latitude: -2},
		{Longitude: 2, Latitude: -2},
		{Longitude: 2, Latitude: 2},
		{Longitude: -2, Latitude: 2},
	}
	index := newVisibleLineworkPolygonIndex(polygon)
	for longitude := -3.0; longitude <= 3; longitude += 0.25 {
		for latitude := -3.0; latitude <= 3; latitude += 0.25 {
			point := GeoPoint{Longitude: longitude, Latitude: latitude}
			want := visibleLineworkContainsWithin(polygon, point, 0.05)
			got := visibleLineworkContainsWithinIndexed(&index, point, 0.05)
			if got != want {
				t.Fatalf("indexed containment for %+v=%v, want %v", point, got, want)
			}
		}
	}
}

func TestSplitVisibleLineworkIntersectionsCreatesSharedGraphNode(t *testing.T) {
	lines := [][]GeoPoint{
		{{Longitude: 0, Latitude: 0}, {Longitude: 10, Latitude: 10}},
		{{Longitude: 0, Latitude: 10}, {Longitude: 10, Latitude: 0}},
	}
	split := splitVisibleLineworkIntersections(lines)
	if len(split) != 4 {
		t.Fatalf("split line count=%d, want four half-edges: %#v", len(split), split)
	}
	junctionCount := 0
	for _, line := range split {
		for _, point := range []GeoPoint{line[0], line[len(line)-1]} {
			if math.Abs(point.Longitude-5) <= 1e-9 && math.Abs(point.Latitude-5) <= 1e-9 {
				junctionCount++
			}
		}
	}
	if junctionCount != 4 {
		t.Fatalf("intersection endpoint count=%d, want four", junctionCount)
	}
}

func TestVisibleLineworkArcIndexMatchesSourceEdgeDistance(t *testing.T) {
	line := []GeoPoint{
		{Longitude: 179, Latitude: 70},
		{Longitude: -179, Latitude: 71},
	}
	index := newVisibleLineworkArcIndex([][]GeoPoint{line})
	query := sphericalInterpolate(line[0], line[1], 0.5)
	query.Latitude += 0.05
	want := sphericalPointArcDistanceKM(query, line[0], line[1])

	for _, endpoints := range [][2]GeoPoint{
		{line[0], line[1]},
		{line[1], line[0]},
		{
			{Longitude: line[0].Longitude + 2e-9, Latitude: line[0].Latitude},
			{Longitude: line[1].Longitude, Latitude: line[1].Latitude - 2e-9},
		},
	} {
		got, matched := index.edgePointDistanceKM(endpoints[0], endpoints[1], query)
		if !matched {
			t.Fatalf("indexed source edge was not matched: %+v", endpoints)
		}
		if math.Abs(got-want) > 1e-6 {
			t.Fatalf("indexed edge distance=%.12f km, want %.12f km", got, want)
		}
	}

	if _, matched := index.edgePointDistanceKM(
		line[0], GeoPoint{Longitude: -178, Latitude: 71}, query,
	); matched {
		t.Fatal("non-source edge unexpectedly matched the source index")
	}
}

func TestVisibleLineworkBoundarySourceMissDistanceDoesNotHideLongChord(t *testing.T) {
	boundaryLines := [][]GeoPoint{
		{{Longitude: 0, Latitude: 0}, {Longitude: 0.5, Latitude: 0}},
		{{Longitude: 1.5, Latitude: 0}, {Longitude: 2, Latitude: 0}},
	}
	// Both endpoints are source vertices, but the connecting edge is not a
	// source arc. Its midpoint is about 55 km from either source segment.
	polygons := [][]GeoPoint{{
		{Longitude: 0, Latitude: 0},
		{Longitude: 2, Latitude: 0},
	}}
	if miss := visibleLineworkBoundarySourceMissDistanceKM(polygons, boundaryLines); miss < 50 {
		t.Fatalf("source miss distance=%.1f km, want midpoint deviation to be retained", miss)
	}
}

func TestVisibleLineworkNodesTJunctionEndpoint(t *testing.T) {
	// 端点落在另一段内部（T 型交点）：参数解会被 1e-9 的分数容差拒绝，而图幅旋转
	// 带来的误差约 2.4e-6 度，于是两个面会被并成一个。这里用一个 8e-7 度（约 9 厘米）
	// 的偏离复现该量级。
	// An endpoint on another segment's interior: the parametric solve rejects it with its
	// 1e-9 fraction tolerance while chart rotation carries about 2.4e-6 degrees of error, so
	// the two faces used to merge. The offset below (8e-7 degrees, about 9 cm) reproduces
	// that magnitude.
	horizontal := []GeoPoint{{Longitude: -1, Latitude: 0}, {Longitude: 1, Latitude: 0}}
	// 端点离水平段 8e-7 度（约 9 厘米）：参数解把它算成 −8e-7 的分数，被 1e-9 容差拒绝。
	// The endpoint sits 8e-7 degrees (about 9 cm) off the horizontal line, which the
	// parametric solve turns into a fraction of -8e-7 and rejects against its 1e-9
	// tolerance.
	vertical := []GeoPoint{{Longitude: 8e-7, Latitude: 8e-7}, {Longitude: 8e-7, Latitude: 1}}
	// 水平段在 T 点被切成两段，竖直段保持一条 ⇒ 2 条线变 3 条线。
	// The horizontal line is cut at the T point while the vertical one stays whole, so two
	// input lines become three.
	split := splitVisibleLineworkIntersections([][]GeoPoint{horizontal, vertical})
	if len(split) != 3 {
		t.Fatalf("lines=%d (%v), want the horizontal line split at the T junction", len(split), split)
	}
	outerEnds := 0
	for _, line := range split {
		if len(line) != 2 {
			continue
		}
		for _, point := range line {
			// 水平段被切成两半后，每半各保留一个 lon=±1、lat=0 的外端。
			// After the cut each half keeps exactly one outer end at lon=±1, lat=0.
			if math.Abs(math.Abs(point.Longitude)-1) < 1e-12 && math.Abs(point.Latitude) < 1e-12 {
				outerEnds++
			}
		}
	}
	if outerEnds != 2 {
		t.Fatalf("horizontal halves keeping an outer end = %d, want 2 (%v)", outerEnds, split)
	}
	// 真分离（起点离水平段约 1.1 公里）不得被打成节点。
	// A genuine separation (the endpoint sits about 1.1 km off the horizontal line) must
	// not be noded.
	far := []GeoPoint{{Longitude: 0.01, Latitude: 0.01}, {Longitude: 0.01, Latitude: 1}}
	splitFar := splitVisibleLineworkIntersections([][]GeoPoint{horizontal, far})
	if len(splitFar) != 2 {
		t.Fatalf("a 1 km separation was noded: %v", splitFar)
	}
}

func TestVisibleLineworkPolygonsRejectsInfiniteTolerances(t *testing.T) {
	lines := [][]GeoPoint{
		{{Longitude: 0, Latitude: 0}, {Longitude: 4, Latitude: 0}},
		{{Longitude: 4, Latitude: 0}, {Longitude: 4, Latitude: 4}},
		{{Longitude: 4, Latitude: 4}, {Longitude: 0, Latitude: 4}},
		{{Longitude: 0, Latitude: 4}, {Longitude: 0, Latitude: 0}},
	}
	fill := [][]GeoPoint{{
		{Longitude: 0, Latitude: 0},
		{Longitude: 4, Latitude: 0},
		{Longitude: 4, Latitude: 4},
		{Longitude: 0, Latitude: 4},
	}}
	if _, err := VisibleLineworkPolygons(lines, fill, fill, math.Inf(1)); err == nil {
		t.Fatal("infinite snap distance was accepted")
	}
	if _, err := VisibleLineworkPolygonsWithAuditTolerance(lines, fill, fill, 1, math.Inf(1)); err == nil {
		t.Fatal("infinite fill audit tolerance was accepted")
	}
}

func TestVisibleLineworkContainsWithinIndexedHandlesNilIndex(t *testing.T) {
	if visibleLineworkContainsWithinIndexed(nil, GeoPoint{Longitude: 1, Latitude: 1}, 0.5) {
		t.Fatal("nil index reported containment")
	}
}

func TestVisibleLineworkCyclesDeduplicateReversedFaces(t *testing.T) {
	lines := [][]GeoPoint{
		{{Longitude: 0, Latitude: 0}, {Longitude: 2, Latitude: 0}},
		{{Longitude: 2, Latitude: 0}, {Longitude: 2, Latitude: 2}},
		{{Longitude: 2, Latitude: 2}, {Longitude: 0, Latitude: 2}},
		{{Longitude: 0, Latitude: 2}, {Longitude: 0, Latitude: 0}},
	}
	chart := polygonUnionChart{xAxis: geoVector3{x: 1}, yAxis: geoVector3{y: 1}, zAxis: geoVector3{z: 1}}
	nodes, edges, _ := buildVisibleLineworkGraph(lines, chart, 1)
	if cycles := enumerateVisibleLineworkCycles(nodes, edges); len(cycles) != 1 {
		t.Fatalf("got %d cycles, want one square counted once for both directions", len(cycles))
	}
}

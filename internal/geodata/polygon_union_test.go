package geodata

import (
	"math"
	"strings"
	"testing"
)

func TestPolygonUnionLatitudeIndexMatchesExactPointLocation(t *testing.T) {
	rings, err := polygonUnionRings([][]GeoPoint{{
		{Longitude: -7, Latitude: -2},
		{Longitude: -1, Latitude: -6},
		{Longitude: 5, Latitude: -3},
		{Longitude: 8, Latitude: 2},
		{Longitude: 3, Latitude: 7},
		{Longitude: -4, Latitude: 5},
	}})
	if err != nil {
		t.Fatalf("polygonUnionRings: %v", err)
	}
	ring := rings[0]
	for latitudeIndex := -80; latitudeIndex <= 80; latitudeIndex++ {
		for longitudeIndex := -100; longitudeIndex <= 100; longitudeIndex++ {
			point := polygonUnionPoint{
				x: float64(longitudeIndex) / 10,
				y: float64(latitudeIndex) / 10,
			}
			got := polygonUnionPointLocationOnEdges(point, ring.locationEdges)
			want := polygonUnionPointLocation(point, ring.points)
			if got != want {
				t.Fatalf("indexed point location at %.1f, %.1f = %d, want %d", point.x, point.y, got, want)
			}
			if got := polygonUnionPointLocationInRing(point, ring); got != want {
				t.Fatalf("binned point location at %.1f, %.1f = %d, want %d", point.x, point.y, got, want)
			}
		}
	}
	for _, point := range ring.points {
		if got := polygonUnionPointLocationOnEdges(point, ring.locationEdges); got != 0 {
			t.Fatalf("indexed point location at boundary %+v = %d, want 0", point, got)
		}
		if got := polygonUnionPointLocationInRing(point, ring); got != 0 {
			t.Fatalf("binned point location at boundary %+v = %d, want 0", point, got)
		}
	}
}

func TestUnionPolygonsMergesOverlappingRingsWithoutInternalEdges(t *testing.T) {
	polygons, err := UnionPolygons([][]GeoPoint{
		{
			{Longitude: 0, Latitude: 0},
			{Longitude: 3, Latitude: 0},
			{Longitude: 3, Latitude: 2},
			{Longitude: 0, Latitude: 2},
		},
		{
			{Longitude: 2, Latitude: 1},
			{Longitude: 4, Latitude: 1},
			{Longitude: 4, Latitude: 3},
			{Longitude: 2, Latitude: 3},
		},
	})
	if err != nil {
		t.Fatalf("UnionPolygons: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("polygon count=%d, want one", len(polygons))
	}
	for _, point := range []GeoPoint{
		{Longitude: 1, Latitude: 1},
		{Longitude: 2.5, Latitude: 2.5},
		{Longitude: 3.5, Latitude: 1.5},
	} {
		if !sweepPolygonContainsOrTouches(polygons[0], point) {
			t.Fatalf("union does not contain %+v", point)
		}
	}
	if sweepPolygonContainsOrTouches(polygons[0], GeoPoint{Longitude: 1, Latitude: 2.5}) {
		t.Fatal("union contains a point outside both inputs")
	}

	// 并集边界恰为两环的外轮廓：被另一环吞掉的角点 (3,2) 与 (2,1) 不得作为顶点留下，
	// 否则说明内部边被当成边界保留。
	wantVertices := []GeoPoint{
		{Longitude: 0, Latitude: 0},
		{Longitude: 3, Latitude: 0},
		{Longitude: 3, Latitude: 1},
		{Longitude: 4, Latitude: 1},
		{Longitude: 4, Latitude: 3},
		{Longitude: 2, Latitude: 3},
		{Longitude: 2, Latitude: 2},
		{Longitude: 0, Latitude: 2},
	}
	for _, want := range wantVertices {
		if !ringHasVertex(polygons[0], want) {
			t.Fatalf("union boundary lost the outline vertex %+v: %+v", want, polygons[0])
		}
	}
	for _, swallowed := range []GeoPoint{
		{Longitude: 3, Latitude: 2},
		{Longitude: 2, Latitude: 1},
	} {
		if ringHasVertex(polygons[0], swallowed) {
			t.Fatalf("union kept the swallowed corner %+v as a boundary vertex (internal edge): %+v", swallowed, polygons[0])
		}
	}
}

// ringHasVertex 判断环上是否存在与目标重合（1e-9 度）的顶点；环首尾重复只算一次。
func ringHasVertex(ring []GeoPoint, want GeoPoint) bool {
	for _, point := range ring {
		if math.Abs(point.Longitude-want.Longitude) <= 1e-9 && math.Abs(point.Latitude-want.Latitude) <= 1e-9 {
			return true
		}
	}
	return false
}

func TestUnionPolygonsPreservesDisjointAndContainedRings(t *testing.T) {
	polygons, err := UnionPolygons([][]GeoPoint{
		{
			{Longitude: 0, Latitude: 0},
			{Longitude: 4, Latitude: 0},
			{Longitude: 4, Latitude: 4},
			{Longitude: 0, Latitude: 4},
		},
		{
			{Longitude: 1, Latitude: 1},
			{Longitude: 2, Latitude: 1},
			{Longitude: 2, Latitude: 2},
			{Longitude: 1, Latitude: 2},
		},
		{
			{Longitude: 10, Latitude: 0},
			{Longitude: 11, Latitude: 0},
			{Longitude: 11, Latitude: 1},
			{Longitude: 10, Latitude: 1},
		},
	})
	if err != nil {
		t.Fatalf("UnionPolygons: %v", err)
	}
	if len(polygons) != 2 {
		t.Fatalf("polygon count=%d, want containing and disjoint rings", len(polygons))
	}
}

func TestPolygonUnionContainmentChecksInputEdgeMidpoints(t *testing.T) {
	result := [][]GeoPoint{{
		{Longitude: 0, Latitude: 0},
		{Longitude: 4, Latitude: 0},
		{Longitude: 4, Latitude: 4},
		{Longitude: 3, Latitude: 4},
		{Longitude: 3, Latitude: 1},
		{Longitude: 1, Latitude: 1},
		{Longitude: 1, Latitude: 4},
		{Longitude: 0, Latitude: 4},
	}}
	input := [][]GeoPoint{{
		{Longitude: 0.5, Latitude: 3.5},
		{Longitude: 3.5, Latitude: 3.5},
		{Longitude: 0.5, Latitude: 3},
	}}
	for _, point := range input[0] {
		if !sphericalPolygonContainsOrTouches(result[0], point) {
			t.Fatalf("fixture input vertex is outside result: %+v", point)
		}
	}
	if polygonUnionContainsInputs(result, input) {
		t.Fatal("containment accepted an input edge crossing the result's open notch")
	}
}

func TestUnionPolygonsMergesAcrossAntimeridian(t *testing.T) {
	polygons, err := UnionPolygons([][]GeoPoint{
		{
			{Longitude: 170, Latitude: -5},
			{Longitude: -175, Latitude: -5},
			{Longitude: -175, Latitude: 5},
			{Longitude: 170, Latitude: 5},
		},
		{
			{Longitude: 175, Latitude: 0},
			{Longitude: -170, Latitude: 0},
			{Longitude: -170, Latitude: 10},
			{Longitude: 175, Latitude: 10},
		},
	})
	if err != nil {
		t.Fatalf("UnionPolygons across antimeridian: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("antimeridian polygon count=%d, want one", len(polygons))
	}
	for _, point := range []GeoPoint{
		{Longitude: 172, Latitude: 0},
		{Longitude: 179, Latitude: 3},
		{Longitude: -172, Latitude: 7},
	} {
		if !sphericalPolygonContainsOrTouches(polygons[0], point) {
			t.Fatalf("antimeridian union does not contain %+v", point)
		}
	}
}

func TestUnionPolygonsMergesOverlappingPolarCaps(t *testing.T) {
	first := SphericalCircle(GeoPoint{Longitude: 0, Latitude: 86}, 7, 72)
	second := SphericalCircle(GeoPoint{Longitude: 120, Latitude: 86}, 7, 72)
	polygons, err := UnionPolygons([][]GeoPoint{first, second})
	if err != nil {
		t.Fatalf("UnionPolygons at north pole: %v", err)
	}
	if len(polygons) != 1 {
		t.Fatalf("polar polygon count=%d, want one", len(polygons))
	}
	for _, point := range []GeoPoint{
		{Longitude: 0, Latitude: 86},
		{Longitude: 120, Latitude: 86},
		{Longitude: 0, Latitude: 90},
	} {
		if !sphericalPolygonContainsOrTouches(polygons[0], point) {
			t.Fatalf("polar union does not contain %+v", point)
		}
	}
}

func TestUnionPolygonsPreservesPoleWindingBoundary(t *testing.T) {
	// The minimum lies just after the ring's start. A planar closure across
	// 360 degrees cuts off that minimum as a separate thin polygon.
	source := []GeoPoint{
		{Longitude: 31.55, Latitude: 3.79},
		{Longitude: 31.78, Latitude: 3.76},
		{Longitude: 32.02, Latitude: 3.74},
		{Longitude: 32.51, Latitude: 3.77},
		{Longitude: 33.02, Latitude: 3.86},
		{Longitude: 60, Latitude: 20},
		{Longitude: 120, Latitude: 50},
		{Longitude: 179, Latitude: 70},
		{Longitude: -120, Latitude: 75},
		{Longitude: -60, Latitude: 55},
		{Longitude: 0, Latitude: 15},
		{Longitude: 30, Latitude: 4.14},
		{Longitude: 31.45, Latitude: 3.81},
	}
	for _, south := range []bool{false, true} {
		for _, reverse := range []bool{false, true} {
			ring := append([]GeoPoint(nil), source...)
			if south {
				for index := range ring {
					ring[index].Latitude = -ring[index].Latitude
				}
			}
			if reverse {
				reverseGeoPoints(ring)
			}
			if _, err := unionPolygonsPlanar([][]GeoPoint{ring}); err == nil {
				t.Errorf("south=%v reverse=%v: planar union accepted a pole-winding ring", south, reverse)
			}
			polygons, err := UnionPolygons([][]GeoPoint{ring})
			if err != nil || len(polygons) != 1 {
				t.Fatalf("south=%v reverse=%v: rings=%d err=%v, want one intact ring", south, reverse, len(polygons), err)
			}
			if len(openGeoRing(polygons[0])) != len(ring) {
				t.Fatalf("union changed a single pole-winding boundary: points=%d want=%d", len(openGeoRing(polygons[0])), len(ring))
			}
			line := append(append([]GeoPoint(nil), ring...), ring[0])
			if miss := visibleLineworkBoundarySourceMissDistanceKM(polygons, [][]GeoPoint{line}); miss > 1e-3 {
				t.Fatalf("pole-winding union leaves source boundary by %.6f km", miss)
			}
		}
	}
}

func TestUnionPolygonsRejectsHoles(t *testing.T) {
	// 4 根条围成方环：并集的外边界面积 4.0、内边界（洞）1.0，真实并集面积 3.0。
	// 环列表表达不了孔洞，因此必须报错，而不是把内边界也当成实体面返回 4+1=5 的超集。
	// Four bars forming a square frame: the union's outer boundary has area 4.0 and the
	// inner one (a hole) 1.0, so the true union area is 3.0. A ring list cannot express
	// the hole, so the union must fail instead of returning the inner boundary as a solid
	// face and reporting the superset 4+1=5.
	ring := func(x0, y0, x1, y1 float64) []GeoPoint {
		return []GeoPoint{
			{Longitude: x0, Latitude: y0},
			{Longitude: x1, Latitude: y0},
			{Longitude: x1, Latitude: y1},
			{Longitude: x0, Latitude: y1},
		}
	}
	bars := [][]GeoPoint{
		ring(-1, 0.5, 1, 1),
		ring(-1, -1, 1, -0.5),
		ring(-1, -0.5, -0.5, 0.5),
		ring(0.5, -0.5, 1, 0.5),
	}
	if _, err := UnionPolygons(bars); err == nil {
		t.Fatal("UnionPolygons accepted a union with a hole instead of reporting it")
	}

	// 正对照：两个分离的方块仍是两个环，不应报错。
	// Positive control: two disjoint squares stay two rings and must not fail.
	disjoint := [][]GeoPoint{ring(0, 0, 1, 1), ring(3, 0, 4, 1)}
	result, err := UnionPolygons(disjoint)
	if err != nil {
		t.Fatalf("disjoint union failed: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("disjoint union rings=%d, want 2", len(result))
	}
}

func TestPolygonUnionWindingPredictsPlanarFailure(t *testing.T) {
	rings := map[string][]GeoPoint{
		"square":     {{0, 0}, {4, 0}, {4, 4}, {0, 4}},
		"equatorial": SphericalCircle(GeoPoint{Longitude: 0, Latitude: 0}, 30, 64),
		"north cap":  SphericalCircle(GeoPoint{Longitude: 0, Latitude: 86}, 7, 64),
		"south cap":  SphericalCircle(GeoPoint{Longitude: 120, Latitude: -86}, 7, 64),
	}
	for name, ring := range rings {
		_, err := unionPolygonsPlanar([][]GeoPoint{ring})
		reported := err != nil && strings.Contains(err.Error(), "winds around a chart pole")
		if winding := polygonUnionRingWindsAroundPole(ring); winding != reported {
			t.Fatalf("%s: winding=%v planar error=%v", name, winding, err)
		}
	}
}

func TestUnionPolygonsWindingInputMatchesSphericalChart(t *testing.T) {
	cap := SphericalCircle(GeoPoint{Longitude: 40, Latitude: 86}, 8, 72)
	want, err := unionPolygonsSphericalChart([][]GeoPoint{cap})
	if err != nil {
		t.Fatalf("unionPolygonsSphericalChart: %v", err)
	}
	got, err := UnionPolygons([][]GeoPoint{cap})
	if err != nil {
		t.Fatalf("UnionPolygons: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("ring count=%d, want %d", len(got), len(want))
	}
	for index := range want {
		if len(got[index]) != len(want[index]) {
			t.Fatalf("ring %d points=%d, want %d", index, len(got[index]), len(want[index]))
		}
		for pointIndex := range want[index] {
			if !SameGeoPoint(got[index][pointIndex], want[index][pointIndex]) {
				t.Fatalf("ring %d point %d=%+v, want %+v",
					index, pointIndex, got[index][pointIndex], want[index][pointIndex])
			}
		}
	}
}

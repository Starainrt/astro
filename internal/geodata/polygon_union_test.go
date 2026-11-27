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

func TestUnionPolygonsSyntheticNarrowOverlap(t *testing.T) {
	inputs := [][]GeoPoint{
		{{Longitude: -14.210843959468804, Latitude: 47.58100849077532}, {Longitude: -14.212350248757758, Latitude: 47.58138906174322}, {Longitude: -14.213847997196353, Latitude: 47.58178485642994}},
		{{Longitude: -14.209918479187706, Latitude: 47.58055494679322}, {Longitude: -14.211357855904737, Latitude: 47.581039875645246}, {Longitude: -14.212784532456112, Latitude: 47.581541709815575}},
	}
	got, err := UnionPolygons(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("regions=%d, want 1", len(got))
	}
	for _, input := range inputs {
		for _, p := range input {
			if !sweepPolygonContainsOrTouches(got[0], p) {
				t.Fatalf("lost vertex: %+v", p)
			}
		}
	}
}

func TestUnionRegionsSyntheticHolesAndIsland(t *testing.T) {
	box := func(x0, y0, x1, y1 float64) []GeoPoint {
		return []GeoPoint{{Longitude: x0, Latitude: y0}, {Longitude: x1, Latitude: y0}, {Longitude: x1, Latitude: y1}, {Longitude: x0, Latitude: y1}}
	}
	inputs := [][]GeoPoint{box(-1, .5, 1, 1), box(-1, -1, 1, -.5), box(-1, -.5, -.5, .5), box(.5, -.5, 1, .5), box(-.1, -.1, .1, .1)}
	regions, err := UnionRegions(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(regions) != 2 || len(regions[0].Holes) != 1 || len(regions[1].Holes) != 0 {
		t.Fatalf("unexpected topology: %+v", regions)
	}
	for _, r := range regions {
		if polygonUnionTestArea(r.Outer) <= 0 {
			t.Fatal("outer winding")
		}
		for _, h := range r.Holes {
			if polygonUnionTestArea(h) >= 0 {
				t.Fatal("hole winding")
			}
		}
	}
	if !sweepPolygonContainsOrTouches(regions[0].Holes[0], GeoPoint{Longitude: .3, Latitude: .3}) {
		t.Fatal("missing hole witness")
	}
}

func TestUnionRegionsBatchesSyntheticHolesAndSeam(t *testing.T) {
	box := func(x0, y0, x1, y1 float64) []GeoPoint {
		return []GeoPoint{{Longitude: x0, Latitude: y0}, {Longitude: x1, Latitude: y0}, {Longitude: x1, Latitude: y1}, {Longitude: x0, Latitude: y1}}
	}
	inputs := [][]GeoPoint{
		box(175, 3, 179, 4), box(175, -4, 179, -3), box(175, -3, 176, 3), box(178, -3, 179, 3),
		box(176.5, -1, 177.5, 1), box(178.5, -1, -179, 1), box(-179.5, -1, -178, 1), box(176.5, -1, 177.5, 1),
	}
	regions, err := polygonUnionRegionBatches(inputs, 16)
	if err != nil {
		t.Fatal(err)
	}
	want, err := UnionRegions(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(regions) != 3 || len(regions[0].Holes) != 1 || len(regions[1].Holes) != 0 || len(regions[2].Holes) != 0 {
		t.Fatalf("unexpected topology: %+v", regions)
	}
	for i, region := range regions {
		if polygonUnionTestArea(region.Outer) != polygonUnionTestArea(want[i].Outer) {
			t.Fatalf("region %d outer area differs: %+v vs %+v", i, region, want[i])
		}
		for j, hole := range region.Holes {
			if polygonUnionTestArea(hole) != polygonUnionTestArea(want[i].Holes[j]) {
				t.Fatalf("region %d hole %d area differs", i, j)
			}
		}
	}
}

func polygonUnionTestArea(ring []GeoPoint) float64 {
	area := 0.0
	for i, a := range ring {
		b := ring[(i+1)%len(ring)]
		area += a.Longitude*b.Latitude - b.Longitude*a.Latitude
	}
	return area / 2
}

func TestUnionRegionsSyntheticCanonicalSeamAndPoles(t *testing.T) {
	cases := []struct {
		name string
		ring []GeoPoint
		pole GeoPoint
	}{
		{"seam", []GeoPoint{{Longitude: 170, Latitude: -10}, {Longitude: -170, Latitude: -10}, {Longitude: -170, Latitude: 10}, {Longitude: 170, Latitude: 10}}, GeoPoint{Longitude: 180, Latitude: 0}},
		{"exact", []GeoPoint{{Longitude: -180, Latitude: -10}, {Longitude: -170, Latitude: -10}, {Longitude: -170, Latitude: 10}, {Longitude: -180, Latitude: 10}}, GeoPoint{Longitude: -175, Latitude: 0}},
		{"north", SphericalCircle(GeoPoint{Longitude: 0, Latitude: 86}, 7, 72), GeoPoint{Longitude: 0, Latitude: 90}},
		{"south", SphericalCircle(GeoPoint{Longitude: 120, Latitude: -86}, 7, 72), GeoPoint{Longitude: 120, Latitude: -90}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for variant := 0; variant < 3; variant++ {
				input := append([]GeoPoint(nil), tc.ring...)
				if variant == 1 {
					for i, j := 0, len(input)-1; i < j; i, j = i+1, j-1 {
						input[i], input[j] = input[j], input[i]
					}
				}
				if variant == 2 {
					shift := 17 % len(input)
					input = append(input[shift:], input[:shift]...)
				}
				regions, err := UnionRegions([][]GeoPoint{input})
				if err != nil {
					t.Fatal(err)
				}
				want := 1
				if tc.name == "seam" {
					want = 2
				}
				if len(regions) != want {
					t.Fatalf("variant%d regions=%d want%d", variant, len(regions), want)
				}
				for _, p := range regions[0].Outer {
					if p.Longitude < -180 || p.Longitude > 180 || math.Abs(p.Latitude) > 90 {
						t.Fatalf("noncanonical %+v", p)
					}
				}
				covered := false
				for _, r := range regions {
					covered = covered || sweepPolygonContainsOrTouches(r.Outer, tc.pole)
				}
				if !covered {
					t.Fatalf("variant%d lost witness %+v", variant, tc.pole)
				}
			}
		})
	}
}

func TestUnionRegionsSyntheticPhysicalPolarCaps(t *testing.T) {
	for _, sign := range []float64{-1, 1} {
		ring := make([]GeoPoint, 360)
		for i := range ring {
			ring[i] = GeoPoint{Longitude: float64(i) - 180, Latitude: 80 * sign}
		}
		for variant := 0; variant < 3; variant++ {
			if variant == 1 {
				for i, j := 0, len(ring)-1; i < j; i, j = i+1, j-1 {
					ring[i], ring[j] = ring[j], ring[i]
				}
			}
			if variant == 2 {
				ring = append(ring[57:], ring[:57]...)
			}
			result, err := UnionRegions([][]GeoPoint{ring})
			if err != nil {
				t.Fatal(err)
			}
			if len(result) != 1 || len(result[0].Holes) != 0 {
				t.Fatalf("polar topology %+v", result)
			}
			for y := -89.0; y <= 89; y += 2 {
				for x := -179.0; x < 180; x += 2 {
					got := sweepPolygonContainsOrTouches(result[0].Outer, GeoPoint{Longitude: x, Latitude: y})
					if got != (sign*y >= 80) {
						t.Fatalf("sign%g variant%d (%g,%g) inside=%v", sign, variant, x, y, got)
					}
				}
			}
		}
	}
}

func TestUnionRegionsExternalPolarFootprint(t *testing.T) {
	ring := []GeoPoint{
		{Longitude: 66.781889636237281, Latitude: 45.160308185808184},
		{Longitude: 53.469036457868185, Latitude: 44.658003047549322},
		{Longitude: 40.518701422318202, Latitude: 43.19110706493047},
		{Longitude: 28.197257917974525, Latitude: 40.868139075177133},
		{Longitude: 16.630766935582244, Latitude: 37.838911911022919},
		{Longitude: 5.8189995080220456, Latitude: 34.266465633517626},
		{Longitude: -4.3218409478872104, Latitude: 30.308052035412747},
		{Longitude: -13.916844118358256, Latitude: 26.106574138601587},
		{Longitude: -23.1029954624791, Latitude: 21.789288443347626},
		{Longitude: -32.012653430994078, Latitude: 17.470143916545986},
		{Longitude: -40.765544061696147, Latitude: 13.253196432980589},
		{Longitude: -49.465106436816257, Latitude: 9.2357376256778068},
		{Longitude: -58.196426407711485, Latitude: 5.5104971069985131},
		{Longitude: -67.024275128876482, Latitude: 2.1665183042687546},
		{Longitude: -75.990839586325421, Latitude: -0.71147717498158647},
		{Longitude: -85.113551859185549, Latitude: -3.0452854300510679},
		{Longitude: -94.384047132435924, Latitude: -4.7668426339490191},
		{Longitude: -103.76960391933653, Latitude: -5.8227248146618882},
		{Longitude: -113.2181103637627, Latitude: -6.1786368736718513},
		{Longitude: -122.66661680818889, Latitude: -5.8227248146618882},
		{Longitude: -132.05217359508947, Latitude: -4.7668426339490226},
		{Longitude: -141.32266886833986, Latitude: -3.0452854300510679},
		{Longitude: -150.44538114119999, Latitude: -0.71147717498158969},
		{Longitude: -159.41194559864891, Latitude: 2.1665183042687546},
		{Longitude: -168.23979431981388, Latitude: 5.5104971069984972},
		{Longitude: -176.97111429070918, Latitude: 9.2357376256778085},
		{Longitude: 174.32932333417074, Latitude: 13.253196432980589},
		{Longitude: 165.57643270346864, Latitude: 17.470143916545972},
		{Longitude: 156.66677473495372, Latitude: 21.789288443347615},
		{Longitude: 147.48062339083282, Latitude: 26.106574138601587},
		{Longitude: 137.88562022036183, Latitude: 30.308052035412747},
		{Longitude: 127.74477976445252, Latitude: 34.266465633517626},
		{Longitude: 116.93301233689238, Latitude: 37.838911911022919},
		{Longitude: 105.36652135450009, Latitude: 40.868139075177112},
		{Longitude: 93.04507785015636, Latitude: 43.19110706493047},
		{Longitude: 80.094742814606434, Latitude: 44.658003047549322},
	}
	regions, err := UnionRegions([][]GeoPoint{ring})
	if err != nil || len(regions) != 1 || len(regions[0].Holes) != 0 {
		t.Fatal(regions, err)
	}
	if !sweepPolygonContainsOrTouches(regions[0].Outer, GeoPoint{Longitude: 0, Latitude: 90}) || sweepPolygonContainsOrTouches(regions[0].Outer, GeoPoint{Longitude: 0, Latitude: -90}) {
		t.Fatal("wrong pole")
	}
	for _, p := range ring {
		if !sweepPolygonContainsOrTouches(regions[0].Outer, p) {
			t.Fatalf("lost %+v", p)
		}
	}
	t.Logf("polar footprint vertices=%d, regions=%d, holes=%d", len(ring), len(regions), len(regions[0].Holes))
}

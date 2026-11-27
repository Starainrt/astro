package geodata

import "testing"

func TestSubtractPolygonAcrossSeam(t *testing.T) {
	outer := []GeoPoint{{160, -20}, {-160, -20}, {-160, 20}, {160, 20}}
	day := []GeoPoint{{175, -5}, {-175, -5}, {-175, 5}, {175, 5}}
	regions, err := SubtractPolygon(outer, day)
	if err != nil || len(regions) != 2 {
		t.Fatalf("seam-split regions: %v, %v", regions, err)
	}
	for _, region := range regions {
		if len(region.Holes) != 0 {
			t.Fatal("a hole crossing the seam should open onto the chart boundary")
		}
	}
	check := func(regions []Region, point GeoPoint, want bool) {
		t.Helper()
		inside := false
		for _, region := range regions {
			inside = inside || sweepPolygonContainsOrTouches(region.Outer, point)
		}
		if inside != want {
			t.Errorf("membership at %v = %t, want %t", point, inside, want)
		}
	}
	for _, lon := range []float64{-178, 178} {
		check(regions, GeoPoint{lon, 0}, false)
		check(regions, GeoPoint{lon, 10}, true)
	}
	fill, err := UnionRegions([][]GeoPoint{day})
	if err != nil {
		t.Fatal(err)
	}
	union, err := UnionRegionSets([][]Region{regions, fill})
	if err != nil {
		t.Fatal(err)
	}
	check(union, GeoPoint{178, 0}, true)
}

func TestSubtractPolygonDegenerateFirstEdge(t *testing.T) {
	outer := []GeoPoint{{-180, 10}, {-180, 10}, {-170, 10}, {-170, 20}, {-180, 20}}
	disjoint := []GeoPoint{{10, 10}, {20, 10}, {20, 20}, {10, 20}}
	regions, err := SubtractPolygon(outer, disjoint)
	if err != nil || len(regions) != 1 || !sweepPolygonContainsOrTouches(regions[0].Outer, GeoPoint{-175, 15}) {
		t.Fatalf("degenerate first edge dropped ring: %v, %v", regions, err)
	}
}

func TestClassifyTouchingHoleVertices(t *testing.T) {
	outer := []GeoPoint{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	diamond := []GeoPoint{{5, 0}, {10, 5}, {5, 10}, {0, 5}}
	regions, err := SubtractPolygon(outer, diamond)
	if err != nil || len(regions) != 4 {
		t.Fatalf("diamond touching outer edges: %d regions, err=%v", len(regions), err)
	}
	outerBox := []polygonUnionPoint{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	innerDiamond := []polygonUnionPoint{{5, 0}, {0, 5}, {5, 10}, {10, 5}}
	classified := polygonUnionClassifyRegions([]polygonUnionJoinedRing{
		{chart: outerBox, points: outer, area: 100},
		{chart: innerDiamond, points: diamond, area: 50},
	})
	if len(classified) != 1 || len(classified[0].Holes) != 1 {
		t.Fatalf("touching hole misclassified: %+v", classified)
	}
}

func TestIntersectPolygon(t *testing.T) {
	a := []GeoPoint{{0, 0}, {20, 0}, {20, 20}, {0, 20}}
	b := []GeoPoint{{10, 10}, {30, 10}, {30, 30}, {10, 30}}
	regions, err := IntersectPolygon(a, b)
	if err != nil || len(regions) != 1 || !sweepPolygonContainsOrTouches(regions[0].Outer, GeoPoint{15, 15}) || sweepPolygonContainsOrTouches(regions[0].Outer, GeoPoint{5, 5}) {
		t.Fatalf("intersect failed: %v, %v", regions, err)
	}
}

package geodata

import (
	"math"
	"testing"
)

func TestJoinPolylineSegmentsUsesBothResultEndpoints(t *testing.T) {
	segments := [][]GeoPoint{
		{
			{Longitude: -179, Latitude: 16},
			{Longitude: -41, Latitude: 64},
		},
		{
			{Longitude: 51, Latitude: 60},
			{Longitude: 179, Latitude: 16},
		},
	}
	joined := JoinPolylineSegments(segments)
	if len(joined) != 4 {
		t.Fatalf("joined point count = %d, want 4", len(joined))
	}
	if absoluteLongitude(joined[0].Longitude) > 90 || absoluteLongitude(joined[len(joined)-1].Longitude) > 90 {
		t.Fatalf("joined open endpoints are on the antimeridian: first=%+v last=%+v", joined[0], joined[len(joined)-1])
	}
	if angularDistanceDegrees(joined[1], joined[2]) > 3 {
		t.Fatalf("nearest antimeridian endpoints were not joined: %+v -> %+v", joined[1], joined[2])
	}
}

func TestSphericalPolygonsContainPointsMatchesIndividualPathChecks(t *testing.T) {
	polygons := [][]GeoPoint{{
		{Longitude: -20, Latitude: -10},
		{Longitude: 20, Latitude: -10},
		{Longitude: 20, Latitude: 10},
		{Longitude: -20, Latitude: 10},
	}}
	points := []GeoPoint{
		{Longitude: 0, Latitude: 0},
		{Longitude: 20, Latitude: 0},
		{Longitude: 40, Latitude: 0},
	}
	got := SphericalPolygonsContainPoints(polygons, points)
	if len(got) != len(points) {
		t.Fatalf("result count=%d, want %d", len(got), len(points))
	}
	for index, point := range points {
		want := SphericalPolygonsContainPaths(polygons, [][]GeoPoint{{point}}, false)
		if got[index] != want {
			t.Fatalf("point %d result=%v, want %v", index, got[index], want)
		}
	}
}

func absoluteLongitude(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func TestSphericalContainmentRejectsAntipodalInterior(t *testing.T) {
	for _, center := range []GeoPoint{{}, {Longitude: 179, Latitude: 70}, {Latitude: -90}} {
		polygon := SphericalCircle(center, 12, 48)
		antipode := GeoPoint{Longitude: normalizeLongitude(center.Longitude + 180), Latitude: -center.Latitude}
		for _, reversed := range []bool{false, true} {
			if reversed {
				for first, last := 0, len(polygon)-1; first < last; first, last = first+1, last-1 {
					polygon[first], polygon[last] = polygon[last], polygon[first]
				}
			}
			if !sphericalPolygonContainsOrTouches(polygon, center) || sphericalPolygonContainsOrTouches(polygon, antipode) {
				t.Errorf("center=%+v reversed=%v: direct containment must distinguish the antipode", center, reversed)
			}
			points := []GeoPoint{center, antipode, polygon[0], {
				Longitude: normalizeLongitude(polygon[0].Longitude + 180), Latitude: -polygon[0].Latitude,
			}}
			got := SphericalPolygonsContainPoints([][]GeoPoint{polygon}, points)
			for index, want := range []bool{true, false, true, false} {
				if got[index] != want {
					t.Errorf("center=%+v reversed=%v point=%+v contains=%v want=%v", center, reversed, points[index], got[index], want)
				}
			}
		}
	}
	// These components cannot share a gnomonic chart. The spherical fallback
	// must also distinguish each component from its antipodal image.
	polygons := [][]GeoPoint{
		SphericalCircle(GeoPoint{}, 12, 48),
		SphericalCircle(GeoPoint{Longitude: 160}, 12, 48),
	}
	got := SphericalPolygonsContainPoints(polygons, []GeoPoint{{}, {Longitude: 160}, {Longitude: 180}, {Longitude: -20}})
	for index, want := range []bool{true, true, false, false} {
		if got[index] != want {
			t.Errorf("disjoint components: point %d contains=%v want=%v", index, got[index], want)
		}
	}
}

func TestSphericalCircleRejectsDegenerateRadii(t *testing.T) {
	center := GeoPoint{Longitude: 10, Latitude: 20}
	for _, radius := range []float64{0, -1, 180, 181, math.Inf(1), math.NaN()} {
		if circle := SphericalCircle(center, radius, 32); len(circle) != 0 {
			t.Fatalf("radius %v produced %d points, want none", radius, len(circle))
		}
	}
	if circle := SphericalCircle(center, 1, 32); len(circle) != 32 {
		t.Fatalf("valid radius produced %d points, want 32", len(circle))
	}
}

func TestJoinPolylineSegmentsDropsSharedJunctionVertex(t *testing.T) {
	first := []GeoPoint{{Longitude: 0, Latitude: 0}, {Longitude: 1, Latitude: 1}}
	second := []GeoPoint{{Longitude: 1, Latitude: 1}, {Longitude: 2, Latitude: 0}}
	for _, segments := range [][][]GeoPoint{{first, second}, {second, first}} {
		joined := JoinPolylineSegments(segments)
		if len(joined) != 3 {
			t.Fatalf("joined=%v, want three points", joined)
		}
		for index := 1; index < len(joined); index++ {
			if SameGeoPoint(joined[index-1], joined[index]) {
				t.Fatalf("joined=%v keeps the shared junction twice", joined)
			}
		}
	}
}

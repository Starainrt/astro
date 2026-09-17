package geodata

import (
	"math"
	"testing"
)

func TestPolylineSegmentsTreatsExactAntimeridianAsOneMeridian(t *testing.T) {
	segments := PolylineSegments([]GeoPoint{
		{Longitude: -180, Latitude: 10},
		{Longitude: 180, Latitude: 20},
	}, ClipView{Projection: ProjectionEquirectangular})
	if len(segments) != 1 || len(segments[0]) != 2 {
		t.Fatalf("exact-antimeridian line segments = %#v", segments)
	}
	if segments[0][0].Longitude != segments[0][1].Longitude {
		t.Fatalf("exact-antimeridian line spans %.1f degrees",
			math.Abs(segments[0][1].Longitude-segments[0][0].Longitude))
	}
}

func TestPolylineSegmentsInterpolatesAntimeridianOnSphere(t *testing.T) {
	segments := PolylineSegments([]GeoPoint{
		{Longitude: 170, Latitude: 80},
		{Longitude: -170, Latitude: 80},
	}, ClipView{Projection: ProjectionEquirectangular})
	if len(segments) != 2 {
		t.Fatalf("high-latitude antimeridian line segments = %d, want 2", len(segments))
	}
	firstCrossing := segments[0][len(segments[0])-1]
	secondCrossing := segments[1][0]
	if firstCrossing.Longitude != 180 || secondCrossing.Longitude != -180 {
		t.Fatalf("unexpected antimeridian crossings: %+v %+v", firstCrossing, secondCrossing)
	}
	if math.Abs(firstCrossing.Latitude-secondCrossing.Latitude) > 1e-10 {
		t.Fatalf("split crossings disagree in latitude: %.12f != %.12f",
			firstCrossing.Latitude, secondCrossing.Latitude)
	}
	// The shorter great-circle arc bends poleward. Linear longitude/latitude
	// interpolation would incorrectly leave the crossing at exactly 80 degrees.
	if firstCrossing.Latitude <= 80.1 || firstCrossing.Latitude >= 81 {
		t.Fatalf("great-circle crossing latitude = %.9f, want (80.1, 81)", firstCrossing.Latitude)
	}
}

func TestPolarPolylineClippingUsesSphericalEquatorIntersection(t *testing.T) {
	segments := PolylineSegments([]GeoPoint{
		{Longitude: 0, Latitude: -10},
		{Longitude: 90, Latitude: 80},
	}, ClipView{Projection: ProjectionNorthPolar})
	if len(segments) != 1 || len(segments[0]) != 2 {
		t.Fatalf("north-polar clipped line = %#v", segments)
	}
	crossing := segments[0][0]
	if math.Abs(crossing.Latitude) > 1e-10 {
		t.Fatalf("equator crossing latitude = %.12f", crossing.Latitude)
	}
	// The great circle crosses near 2 degrees, not at the 45 degree
	// arithmetic interpolation used by planar clipping.
	if crossing.Longitude < 1 || crossing.Longitude > 3 {
		t.Fatalf("spherical equator crossing longitude = %.9f, want (1, 3)", crossing.Longitude)
	}
}

func TestPolygonFragmentsDropsExactAntimeridianZeroAreaDuplicate(t *testing.T) {
	fragments := PolygonFragments([]GeoPoint{
		{Longitude: -180, Latitude: 15},
		{Longitude: 180, Latitude: 14},
		{Longitude: 150, Latitude: 13},
		{Longitude: 150, Latitude: -12},
		{Longitude: 180, Latitude: -11},
		{Longitude: -180, Latitude: -10},
	}, ClipView{Projection: ProjectionEquirectangular})
	if len(fragments) != 1 {
		t.Fatalf("exact-antimeridian polygon produced %d fragments, want 1", len(fragments))
	}
	if math.Abs(signedPolygonArea(fragments[0])) < 1e-12 {
		t.Fatal("exact-antimeridian polygon fragment has zero area")
	}
}

func TestPolygonFragmentsClosesPoleEnclosingRingsAtEquirectangularMapEdge(t *testing.T) {
	for _, test := range []struct {
		name             string
		boundaryLatitude float64
		insideLatitude   float64
		outsideLatitude  float64
	}{
		{name: "north", boundaryLatitude: 70, insideLatitude: 89, outsideLatitude: -89},
		{name: "south", boundaryLatitude: -70, insideLatitude: -89, outsideLatitude: 89},
	} {
		t.Run(test.name, func(t *testing.T) {
			polygon := []GeoPoint{
				{Longitude: -135, Latitude: test.boundaryLatitude},
				{Longitude: -45, Latitude: test.boundaryLatitude},
				{Longitude: 45, Latitude: test.boundaryLatitude},
				{Longitude: 135, Latitude: test.boundaryLatitude},
				{Longitude: -135, Latitude: test.boundaryLatitude},
			}
			fragments := PolygonFragments(polygon, ClipView{Projection: ProjectionEquirectangular})
			if len(fragments) == 0 {
				t.Fatal("pole-enclosing polygon produced no fragments")
			}
			for _, longitude := range []float64{-150, -90, 0, 90, 150} {
				if !polygonFragmentsContain(fragments, longitude, test.insideLatitude) {
					t.Fatalf("pole cap does not contain %.0f, %.0f", longitude, test.insideLatitude)
				}
				if polygonFragmentsContain(fragments, longitude, test.outsideLatitude) {
					t.Fatalf("pole cap incorrectly contains %.0f, %.0f", longitude, test.outsideLatitude)
				}
			}
		})
	}
}

func TestPolygonFragmentsPoleClosurePreservesAntimeridianSeam(t *testing.T) {
	for _, test := range []struct {
		name         string
		poleLatitude float64
		boundary     float64
	}{
		{name: "north", poleLatitude: 90, boundary: 50},
		{name: "south", poleLatitude: -90, boundary: -50},
	} {
		t.Run(test.name, func(t *testing.T) {
			polygon := []GeoPoint{
				{Longitude: 150, Latitude: test.boundary},
				{Longitude: 180, Latitude: 5},
				{Longitude: 120, Latitude: -5},
				{Longitude: 30, Latitude: 15},
				{Longitude: -60, Latitude: 25},
				{Longitude: -150, Latitude: 35},
			}
			if test.poleLatitude < 0 {
				polygon = mirrorGeoPointLatitudes(polygon)
			}
			fragments := PolygonFragments(polygon, ClipView{Projection: ProjectionEquirectangular})
			if len(fragments) == 0 || len(fragments) > 2 {
				t.Fatalf("pole polygon fragments=%d, want one or two map-edge fragments", len(fragments))
			}
			for _, longitude := range []float64{149, 151} {
				latitude := 60.0
				if test.poleLatitude < 0 {
					latitude = -60
				}
				if !polygonFragmentsContain(fragments, longitude, latitude) {
					t.Fatalf("pole closure lost %.0f, %.0f beside its internal seam", longitude, latitude)
				}
			}
		})
	}
}

func TestPolygonFragmentsPoleClosureAnchorsInteriorStartAtAntimeridian(t *testing.T) {
	// This ring winds around the north pole but starts in the Bering Sea. The
	// pole closure must use the map seam, not a meridian through that start
	// vertex.
	fragments := PolygonFragments([]GeoPoint{
		{Longitude: -172, Latitude: 50},
		{Longitude: -100, Latitude: 70},
		{Longitude: 0, Latitude: 80},
		{Longitude: 100, Latitude: 70},
		{Longitude: 188, Latitude: 50},
	}, ClipView{Projection: ProjectionEquirectangular})
	if len(fragments) == 0 {
		t.Fatal("pole-winding ring produced no fragments")
	}
	for _, fragment := range fragments {
		for index, first := range fragment {
			second := fragment[(index+1)%len(fragment)]
			if math.Abs(first.Longitude-second.Longitude) <= 1e-9 &&
				math.Abs(first.Latitude-second.Latitude) > 20 &&
				math.Abs(first.Longitude) < 179.999 {
				t.Fatalf("interior pole closure at %.3f: %+v -> %+v", first.Longitude, first, second)
			}
		}
	}
}

func TestPolygonFragmentsPoleClosureDoesNotDependOnVertexDensity(t *testing.T) {
	base := []GeoPoint{
		{Longitude: -135, Latitude: 70},
		{Longitude: -45, Latitude: 70},
		{Longitude: 45, Latitude: -60},
		{Longitude: 135, Latitude: -60},
	}
	dense := []GeoPoint{
		{Longitude: -135, Latitude: 70},
		{Longitude: -45, Latitude: 70},
		{Longitude: 45, Latitude: -60},
		{Longitude: 55, Latitude: -60},
		{Longitude: 65, Latitude: -60},
		{Longitude: 75, Latitude: -60},
		{Longitude: 85, Latitude: -60},
		{Longitude: 95, Latitude: -60},
		{Longitude: 105, Latitude: -60},
		{Longitude: 115, Latitude: -60},
		{Longitude: 125, Latitude: -60},
		{Longitude: 135, Latitude: -60},
	}
	for _, test := range []struct {
		name  string
		base  []GeoPoint
		dense []GeoPoint
	}{
		{name: "north", base: base, dense: dense},
		{name: "south", base: mirrorGeoPointLatitudes(base), dense: mirrorGeoPointLatitudes(dense)},
	} {
		t.Run(test.name, func(t *testing.T) {
			baseFragments := PolygonFragments(test.base, ClipView{Projection: ProjectionEquirectangular})
			denseFragments := PolygonFragments(test.dense, ClipView{Projection: ProjectionEquirectangular})
			for _, point := range []GeoPoint{
				{Longitude: 0, Latitude: 89},
				{Longitude: 0, Latitude: -89},
			} {
				baseContains := polygonFragmentsContain(baseFragments, point.Longitude, point.Latitude)
				denseContains := polygonFragmentsContain(denseFragments, point.Longitude, point.Latitude)
				if denseContains != baseContains {
					t.Fatalf("densifying a boundary changed containment at %.0f, %.0f: base=%v dense=%v",
						point.Longitude, point.Latitude, baseContains, denseContains)
				}
			}
		})
	}
}

func TestPolygonFragmentsPolarIndentationPreservesInterior(t *testing.T) {
	// A polar boundary crosses the map seam three times. Closing through the
	// lowest crossing fills the excluded indentation on the east side.
	polygon := []GeoPoint{
		{Longitude: -172, Latitude: 50}, {Longitude: -179, Latitude: 54},
		{Longitude: 175, Latitude: 60}, {Longitude: 173, Latitude: 65},
		{Longitude: 175, Latitude: 70}, {Longitude: -178, Latitude: 75},
		{Longitude: -140, Latitude: 79}, {Longitude: -70, Latitude: 68},
		{Longitude: 0, Latitude: 15}, {Longitude: 90, Latitude: 0},
		{Longitude: 140, Latitude: -23}, {Longitude: 175, Latitude: 30},
		{Longitude: -179, Latitude: 40},
	}
	for _, sign := range []float64{1, -1} {
		points := append([]GeoPoint(nil), polygon...)
		for i := range points {
			points[i].Latitude *= sign
		}
		for offset := range points {
			rotated := append(append([]GeoPoint(nil), points[offset:]...), points[:offset]...)
			for _, reverse := range []bool{false, true} {
				if reverse {
					for i, j := 0, len(rotated)-1; i < j; i, j = i+1, j-1 {
						rotated[i], rotated[j] = rotated[j], rotated[i]
					}
				}
				fragments := PolygonFragments(rotated, ClipView{Projection: ProjectionEquirectangular})
				for _, point := range []GeoPoint{
					{Longitude: 179, Latitude: sign * 65}, {Longitude: 175, Latitude: sign * 65},
					{Longitude: 170, Latitude: sign * 65}, {Longitude: -179, Latitude: sign * 65},
					{Longitude: 179, Latitude: sign * 85}, {Longitude: 0, Latitude: sign * 85},
				} {
					want := SphericalPolygonsContainPoints([][]GeoPoint{rotated}, []GeoPoint{point})[0]
					if got := polygonFragmentsContain(fragments, point.Longitude, point.Latitude); got != want {
						t.Fatalf("sign=%g offset=%d reversed=%v point=%+v: split=%v sphere=%v", sign, offset, reverse, point, got, want)
					}
				}
			}
		}
	}
}

func mirrorGeoPointLatitudes(points []GeoPoint) []GeoPoint {
	mirrored := make([]GeoPoint, len(points))
	for index, point := range points {
		mirrored[index] = GeoPoint{Longitude: point.Longitude, Latitude: -point.Latitude}
	}
	return mirrored
}

func polygonFragmentsContain(fragments [][]GeoPoint, longitude, latitude float64) bool {
	for _, polygon := range fragments {
		inside := false
		for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
			a, b := polygon[previous], polygon[current]
			if (a.Latitude > latitude) == (b.Latitude > latitude) {
				continue
			}
			intersection := a.Longitude +
				(latitude-a.Latitude)*(b.Longitude-a.Longitude)/(b.Latitude-a.Latitude)
			if intersection > longitude {
				inside = !inside
			}
		}
		if inside {
			return true
		}
	}
	return false
}

func TestClosePoleEnclosingPolygonRequiresSingleWinding(t *testing.T) {
	open := []GeoPoint{
		{Longitude: 0, Latitude: 0},
		{Longitude: 1, Latitude: 0},
		{Longitude: 1, Latitude: 1},
	}
	if closed := closePoleEnclosingPolygon(open); len(closed) != len(open) {
		t.Fatalf("a ring without pole winding gained %d closure points", len(closed)-len(open))
	}
	winding := []GeoPoint{
		{Longitude: -135, Latitude: 70},
		{Longitude: -45, Latitude: 70},
		{Longitude: 45, Latitude: 70},
		{Longitude: 135, Latitude: 70},
	}
	if closed := closePoleEnclosingPolygon(winding); len(closed) <= len(winding) {
		t.Fatalf("a single-winding ring kept %d points, want a pole closure", len(closed))
	}
}

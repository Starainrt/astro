package geodata

import (
	"math"
	"testing"
)

func TestPolylineSegmentsTreatsExactAntimeridianAsOneMeridian(t *testing.T) {
	segments := PolylineSegments([]GeoPoint{
		{Longitude: -180, Latitude: 10},
		{Longitude: 180, Latitude: 20},
	}, ProjectionEquirectangular)
	if len(segments) != 1 || len(segments[0]) != 2 {
		t.Fatalf("exact-antimeridian line segments = %#v", segments)
	}
	if segments[0][0].Longitude != segments[0][1].Longitude {
		t.Fatalf("exact-antimeridian line spans %.1f degrees",
			math.Abs(segments[0][1].Longitude-segments[0][0].Longitude))
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
	}, ProjectionEquirectangular)
	if len(fragments) != 1 {
		t.Fatalf("exact-antimeridian polygon produced %d fragments, want 1", len(fragments))
	}
	if math.Abs(signedPolygonArea(fragments[0])) < 1e-12 {
		t.Fatal("exact-antimeridian polygon fragment has zero area")
	}
}

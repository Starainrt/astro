package svgmap

import (
	"math"
	"strings"
	"testing"
)

func TestPolarProjectionPlacesPoleAtCenterAndEquatorOnFrame(t *testing.T) {
	frame := Frame{X: 10, Y: 20, Width: 300, Height: 300, Projection: ProjectionNorthPolar}
	x, y, ok := frame.Project(123, 90)
	if !ok || math.Abs(x-160) > 1e-9 || math.Abs(y-170) > 1e-9 {
		t.Fatalf("north pole = %.3f %.3f %v, want map center", x, y, ok)
	}
	x, y, ok = frame.Project(0, 0)
	if !ok || math.Abs(x-160) > 1e-9 || math.Abs(y-20) > 1e-9 {
		t.Fatalf("prime-meridian equator = %.3f %.3f %v, want top edge", x, y, ok)
	}
	if _, _, ok := frame.Project(0, -1); ok {
		t.Fatal("north-polar projection accepted a southern-hemisphere point")
	}
}

func TestProjectionClippingDoesNotSplitPolarAntimeridian(t *testing.T) {
	points := []GeoPoint{{Longitude: 170, Latitude: 70}, {Longitude: -170, Latitude: 70}}
	segments := PolylineSegments(points, ProjectionNorthPolar)
	if len(segments) != 1 || len(segments[0]) != 2 {
		t.Fatalf("polar antimeridian segments = %#v, want one continuous segment", segments)
	}
	segments = PolylineSegments(points, ProjectionEquirectangular)
	if len(segments) != 2 {
		t.Fatalf("equirectangular antimeridian segment count = %d, want 2", len(segments))
	}
}

func TestNaturalEarthAssetsAndPolarClip(t *testing.T) {
	if len(equirectangularLandPath) < 150000 || len(northPolarLandPath) < 100000 || len(southPolarLandPath) < 40000 {
		t.Fatalf("unexpected embedded map sizes: world=%d north=%d south=%d",
			len(equirectangularLandPath), len(northPolarLandPath), len(southPolarLandPath))
	}
	frame := Frame{Width: 300, Height: 300, Projection: ProjectionSouthPolar}
	var builder strings.Builder
	frame.WriteOcean(&builder)
	frame.WriteLand(&builder, "clip")
	frame.WriteFrame(&builder)
	for _, want := range []string{`class="map-ocean"`, `class="land"`, `class="map-frame"`, `<circle`} {
		if !strings.Contains(builder.String(), want) {
			t.Fatalf("polar map output missing %q", want)
		}
	}
}

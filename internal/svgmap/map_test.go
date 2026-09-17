package svgmap

import (
	"fmt"
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
	// 事件经线朝下，中心经线在赤道上落在圆盘最下方。
	x, y, ok = frame.Project(0, 0)
	if !ok || math.Abs(x-160) > 1e-9 || math.Abs(y-320) > 1e-9 {
		t.Fatalf("prime-meridian equator = %.3f %.3f %v, want bottom edge", x, y, ok)
	}
	if _, _, ok := frame.Project(0, -1); ok {
		t.Fatal("north-polar projection accepted a southern-hemisphere point")
	}
}

func TestProjectionClippingDoesNotSplitPolarAntimeridian(t *testing.T) {
	points := []GeoPoint{{Longitude: 170, Latitude: 70}, {Longitude: -170, Latitude: 70}}
	segments := PolylineSegments(points, ClipView{Projection: ProjectionNorthPolar})
	if len(segments) != 1 || len(segments[0]) != 2 {
		t.Fatalf("polar antimeridian segments = %#v, want one continuous segment", segments)
	}
	segments = PolylineSegments(points, ClipView{Projection: ProjectionEquirectangular})
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

// 极地投影必须与“从极点上方俯视”一致，并把事件经线转到可读的一侧：
// 北极图的极点在圆心，事件经线落在正下方，于是事件附近是北在上、东在右。
// 用 +sin(λ−C) 会得到镜像图，东京会跑到北京西边。
func TestPolarProjectionKeepsTrueEastWest(t *testing.T) {
	frame := Frame{Width: 200, Height: 200, Projection: ProjectionNorthPolar}
	centreY := frame.Height / 2

	// 极点落在圆心。
	_, poleY, _ := frame.Project(0, 90)
	if math.Abs(poleY-centreY) > 1e-9 {
		t.Fatalf("pole projects to y=%.3f, want centre %.3f", poleY, centreY)
	}
	// 事件经线朝下：赤道上中心经线的点落在圆盘最下方。
	_, meridianY, _ := frame.Project(0, 0)
	if meridianY <= centreY {
		t.Fatalf("centre meridian projects to y=%.1f, want below centre %.1f", meridianY, centreY)
	}
	// 事件附近东在右：同纬度上偏东的点横坐标更大。
	west, _, _ := frame.Project(-10, 45)
	east, _, _ := frame.Project(10, 45)
	if east <= west {
		t.Fatalf("10E projects to x=%.1f, want right of 10W at %.1f", east, west)
	}
}

func TestCenteredEquirectangularLandCoversWholeFrame(t *testing.T) {
	for _, center := range []float64{-179.5, -104.145, -90, -1, 1, 90, 104.145, 179.5} {
		frame := Frame{
			X: 57, Y: 184, Width: 360, Height: 180,
			Projection: ProjectionEquirectangular, CenterLongitude: center,
		}
		var builder strings.Builder
		frame.WriteLand(&builder, "clip")
		offsets := centeredLandTileOffsets(t, builder.String())
		if len(offsets) == 0 {
			t.Fatalf("center %.3f drew no land tiles", center)
		}
		left, right := math.Inf(1), math.Inf(-1)
		for _, offset := range offsets {
			left = math.Min(left, offset)
			right = math.Max(right, offset+frame.Width)
		}
		if left > frame.X+1e-6 || right < frame.X+frame.Width-1e-6 {
			t.Fatalf("center %.3f covers longitude band [%.3f,%.3f], frame needs [%.3f,%.3f]",
				center, left, right, frame.X, frame.X+frame.Width)
		}
	}
}

func centeredLandTileOffsets(t *testing.T, document string) []float64 {
	t.Helper()
	offsets := make([]float64, 0, 2)
	for _, piece := range strings.Split(document, `<use href="#land-equirectangular-tile" transform="translate(`)[1:] {
		var x, y float64
		if _, err := fmt.Sscanf(piece, "%f %f", &x, &y); err != nil {
			t.Fatalf("parse land tile translate: %v", err)
		}
		offsets = append(offsets, x)
	}
	return offsets
}

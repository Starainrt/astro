package svgmap

import (
	"math"
	"testing"
)

// 居中经线对面的接缝同时是画面左右边缘：贴右边缘的分段必须落在窗口右端，
// 折回左端会让闭合边横穿整幅图。
func TestEquirectangularSeamKeepsBothWindowEdges(t *testing.T) {
	frame := Frame{X: 52, Y: 168, Width: 610.8, Height: 305.4,
		Projection: ProjectionEquirectangular, CenterLongitude: 176.271}
	left := frame.CenterLongitude - 180
	right := frame.CenterLongitude + 180
	if x, _, _ := frame.Project(left, 0); math.Abs(x-frame.X) > 1e-9 {
		t.Fatalf("seam at the left edge maps to x=%.6f, want %.6f", x, frame.X)
	}
	if x, _, _ := frame.Project(right, 0); math.Abs(x-(frame.X+frame.Width)) > 1e-9 {
		t.Fatalf("the same seam in the adjacent world maps to x=%.6f, want %.6f", x, frame.X+frame.Width)
	}
	if x, _, _ := frame.Project(right-0.5, 0); x >= frame.X+frame.Width {
		t.Fatalf("a point just west of the seam maps to x=%.6f, want inside the frame", x)
	}

	ring := make([]GeoPoint, 0, 40)
	for longitude := -10.0; longitude <= 5; longitude += 5 {
		ring = append(ring, GeoPoint{Longitude: longitude, Latitude: 0})
	}
	for latitude := 5.0; latitude <= 20; latitude += 5 {
		ring = append(ring, GeoPoint{Longitude: 5, Latitude: latitude})
	}
	for longitude := 0.0; longitude >= -10; longitude -= 5 {
		ring = append(ring, GeoPoint{Longitude: longitude, Latitude: 20})
	}
	for latitude := 15.0; latitude >= 5; latitude -= 5 {
		ring = append(ring, GeoPoint{Longitude: -10, Latitude: latitude})
	}
	fragments := PolygonFragments(ring, frame.Clip())
	if len(fragments) != 2 {
		t.Fatalf("straddling ring fragments = %d, want 2", len(fragments))
	}
	for index, fragment := range fragments {
		minX, maxX := math.Inf(1), math.Inf(-1)
		for _, point := range fragment {
			x, _, ok := frame.Project(point.Longitude, point.Latitude)
			if !ok {
				t.Fatalf("fragment %d dropped a point", index)
			}
			minX, maxX = math.Min(minX, x), math.Max(maxX, x)
		}
		if maxX-minX > frame.Width/2 {
			t.Fatalf("fragment %d spans %.3f px, want one side of the seam only", index, maxX-minX)
		}
	}
}

// 绕极点一圈的环靠地图上边缘闭合，闭合边两端必须分贴左右边缘。
func TestEquirectangularPoleClosureUsesMapEdges(t *testing.T) {
	frame := Frame{X: 52, Y: 168, Width: 610.8, Height: 305.4,
		Projection: ProjectionEquirectangular, CenterLongitude: 176.271}
	ring := make([]GeoPoint, 0, 36)
	for longitude := -180.0; longitude < 180; longitude += 10 {
		ring = append(ring, GeoPoint{Longitude: longitude, Latitude: 70})
	}
	fragments := PolygonFragments(ring, frame.Clip())
	if len(fragments) != 1 {
		t.Fatalf("cap ring fragments = %d, want 1", len(fragments))
	}
	left, right, top := false, false, false
	for _, point := range fragments[0] {
		if math.Abs(point.Latitude) < 89.999 {
			continue
		}
		x, y, _ := frame.Project(point.Longitude, point.Latitude)
		if math.Abs(y-frame.Y) > 1e-6 {
			t.Fatalf("pole closure at y=%.3f, want the map top edge %.3f", y, frame.Y)
		}
		switch {
		case math.Abs(x-frame.X) < 1e-6:
			left = true
		case math.Abs(x-(frame.X+frame.Width)) < 1e-6:
			right = true
		}
		top = true
	}
	if !top || !left || !right {
		t.Fatalf("pole closure left=%v right=%v top=%v, want the cap closed along both edges", left, right, top)
	}
	if fragments := PolygonFragments(ring, ClipView{Projection: ProjectionEquirectangular}); len(fragments) != 1 {
		t.Fatalf("uncentred cap fragments = %d, want 1", len(fragments))
	}
}

package geodata

import (
	"math"
	"runtime"
	"testing"
)

func orthographicTestContainsRing(ring []GeoPoint, point GeoPoint) bool {
	return sphericalPolygonContainsOrTouches(ring, point)
}

// 采样点若紧贴环边界，容差内外的判定会摇摆；这类点不参与比对。
func orthographicTestBorderline(ring []GeoPoint, point GeoPoint) bool {
	base := orthographicTestContainsRing(ring, point)
	for _, bearing := range []float64{0, 90, 180, 270} {
		offset := InterpolateGreatCircle(point, GeoPoint{
			Longitude: NormalizeLongitude(point.Longitude + 0.05*math.Cos(bearing*rad)),
			Latitude:  point.Latitude + 0.05*math.Sin(bearing*rad),
		}, 0.001)
		if orthographicTestContainsRing(ring, offset) != base {
			return true
		}
	}
	return false
}

func orthographicTestPlanarContains(polygon [][2]float64, x, y float64) bool {
	inside := false
	for index, current := range polygon {
		previous := polygon[(index+len(polygon)-1)%len(polygon)]
		if (current[1] > y) != (previous[1] > y) {
			span := previous[1] - current[1]
			if span != 0 {
				if current[0]+(y-current[1])/span*(previous[0]-current[0]) < x {
					inside = !inside
				}
			}
		}
	}
	return inside
}

func orthographicTestFillMatches(t *testing.T, name string, ring []GeoPoint, center GeoPoint) {
	t.Helper()
	fragments := polygonFragmentsOrthographic(ring, center)
	projected := make([][][2]float64, 0, len(fragments))
	for _, fragment := range fragments {
		polygon := make([][2]float64, 0, len(fragment))
		for _, point := range fragment {
			x, y, ok := OrthographicDiskPoint(point, center)
			if !ok {
				t.Fatalf("%s: fragment point %v is on the back hemisphere", name, point)
			}
			polygon = append(polygon, [2]float64{x, y})
		}
		projected = append(projected, polygon)
	}
	checked, mismatches := 0, 0
	state := uint64(20260916)
	for sample := 0; sample < 4000; sample++ {
		state = state*6364136223846793005 + 1442695040888963407
		latitude := float64(int64(state>>11)%18000)/100 - 90
		state = state*6364136223846793005 + 1442695040888963407
		longitude := float64(int64(state>>11)%36000)/100 - 180
		point := GeoPoint{Longitude: longitude, Latitude: latitude}
		x, y, visible := OrthographicDiskPoint(point, center)
		if !visible {
			continue
		}
		// 视界闭合弧是折线，紧贴视界的一薄层（约 1°）落在弦与圆弧之间，判定本就有歧义；
		// 在 600 像素的球面图上这一层不足 0.1 像素，不参与比对。
		if orthographicDepth(point, center) < 0.02 {
			continue
		}
		if orthographicTestBorderline(ring, point) {
			continue
		}
		expected := orthographicTestContainsRing(ring, point)
		got := false
		for _, polygon := range projected {
			if orthographicTestPlanarContains(polygon, x, y) {
				got = !got
			}
		}
		checked++
		if expected != got {
			mismatches++
			if mismatches <= 3 {
				t.Errorf("%s: point %.3f,%.3f expected %v got %v", name, longitude, latitude, expected, got)
			}
		}
	}
	if checked < 200 {
		t.Fatalf("%s: only %d samples usable", name, checked)
	}
	t.Logf("%s: %d samples, %d mismatches, fragments=%d", name, checked, mismatches, len(fragments))
	if mismatches != 0 {
		t.Fatalf("%s: %d/%d samples disagree with spherical containment", name, mismatches, checked)
	}
}

// 裁剪后的填充必须与"球面包含且位于可见半球"完全一致，这同时验证了视界闭合弧的取侧。
func TestOrthographicPolygonFragmentsFillMatchesContainment(t *testing.T) {
	center := GeoPoint{Longitude: 144.1, Latitude: 24.2}
	for _, fixture := range []struct {
		name string
		ring []GeoPoint
	}{
		{name: "visible", ring: SphericalCircle(GeoPoint{Longitude: 140, Latitude: 30}, 25, 180)},
		{name: "straddling", ring: SphericalCircle(GeoPoint{Longitude: 60, Latitude: 40}, 35, 180)},
		{name: "behind", ring: SphericalCircle(GeoPoint{Longitude: -40, Latitude: -30}, 20, 180)},
		{name: "encircling", ring: SphericalCircle(center, 100, 240)},
		{name: "antipodal-cap", ring: SphericalCircle(GeoPoint{Longitude: -35.9, Latitude: -24.2}, 20, 180)},
		{name: "limb-hugging", ring: SphericalCircle(GeoPoint{Longitude: 100, Latitude: 60}, 60, 240)},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			orthographicTestFillMatches(t, fixture.name, fixture.ring, center)
		})
	}
}

// 折线裁剪只保留可见段，且两端恰好落在视界上。
func TestOrthographicPolylineSegmentsEndOnLimb(t *testing.T) {
	center := GeoPoint{Longitude: 0, Latitude: 0}
	points := []GeoPoint{
		{Longitude: -120, Latitude: 10},
		{Longitude: 0, Latitude: 0},
		{Longitude: 120, Latitude: -10},
	}
	segments := clipPolylineOrthographic(points, center)
	if len(segments) != 1 {
		t.Fatalf("expected one visible segment, got %d", len(segments))
	}
	segment := segments[0]
	first, last := segment[0], segment[len(segment)-1]
	if depth := math.Abs(orthographicDepth(first, center)); depth > 1e-9 {
		t.Fatalf("segment start is not on the limb: depth=%g", depth)
	}
	if depth := math.Abs(orthographicDepth(last, center)); depth > 1e-9 {
		t.Fatalf("segment end is not on the limb: depth=%g", depth)
	}
	for _, point := range segment {
		if orthographicDepth(point, center) < -1e-12 {
			t.Fatalf("segment keeps a back-hemisphere point: %v", point)
		}
	}
}

// 视点自身投影到盘心，与之相距 90° 的点落在盘边。
func TestOrthographicDiskPointReferenceCases(t *testing.T) {
	center := GeoPoint{Longitude: 144.1, Latitude: 24.2}
	if x, y, ok := OrthographicDiskPoint(center, center); !ok || math.Hypot(x, y) > 1e-12 {
		t.Fatalf("center projects to %g,%g ok=%v", x, y, ok)
	}
	// 视界是与视点相距 90° 的大圆；同纬度加 90° 经度并不等于 90° 球面距离。
	centerVector := geoPointVector(center)
	axis := geoVector3{x: 0, y: 0, z: 1}
	if math.Abs(geoVectorDot(axis, centerVector)) > 0.9 {
		axis = geoVector3{x: 1, y: 0, z: 0}
	}
	east, ok := geoVectorNormalize(geoVectorAdd(axis, geoVectorScale(centerVector, -geoVectorDot(axis, centerVector))))
	if !ok {
		t.Fatal("degenerate tangent basis")
	}
	onLimbPoints := []GeoPoint{geoVectorPoint(east), geoVectorPoint(geoVectorScale(east, -1))}
	north := geoVectorCross(east, centerVector)
	onLimbPoints = append(onLimbPoints, geoVectorPoint(north), geoVectorPoint(geoVectorScale(north, -1)))
	for _, onLimb := range onLimbPoints {
		x, y, ok := OrthographicDiskPoint(onLimb, center)
		if !ok || math.Abs(math.Hypot(x, y)-1) > 1e-9 {
			t.Fatalf("%v should sit on the limb, got %g,%g ok=%v", onLimb, x, y, ok)
		}
	}
	if _, _, ok := OrthographicDiskPoint(GeoPoint{Longitude: -35.9, Latitude: -24.2}, center); ok {
		t.Fatal("the antipode must not be visible")
	}
}

// 反复穿越视界的环会产生大量短可见段；单段预分配若按环长给容量，分配量会退化成 O(段数×环长)。
func TestOrthographicFragmentsBoundAllocationForManyShortRuns(t *testing.T) {
	const cycles = 500
	ring := make([]GeoPoint, 0, 4*cycles)
	for index := 0; index < cycles; index++ {
		ring = append(ring,
			GeoPoint{Longitude: 80, Latitude: 0},
			GeoPoint{Longitude: 80, Latitude: 0.5},
			GeoPoint{Longitude: 100, Latitude: 0},
			GeoPoint{Longitude: 100, Latitude: 0.5},
		)
	}
	center := GeoPoint{}
	if fragments := polygonFragmentsOrthographic(ring, center); len(fragments) < cycles/2 {
		t.Fatalf("synthetic ring produced %d fragments, want at least %d", len(fragments), cycles/2)
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	fragments := polygonFragmentsOrthographic(ring, center)
	runtime.ReadMemStats(&after)
	allocated := after.TotalAlloc - before.TotalAlloc
	limit := uint64(12 << 20)
	if allocated > limit {
		t.Fatalf("fragments with %d runs allocated %d bytes, limit %d", len(fragments), allocated, limit)
	}
	t.Logf("%d runs allocated %d bytes", len(fragments), allocated)
}

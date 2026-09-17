package geojson

import (
	"math"
	"testing"

	eclipsecore "github.com/starainrt/astro/eclipse"
)

// solarCentralBandPointSegmentKM 用于判断解析限线与导出带边界的偏离量（超过
// solarCentralBandSnapToleranceKM 就会改用环导出边）。跨换日线的限线必须先把经度差
// 归约到 ±180°，否则会被判成数万公里之外。
// solarCentralBandPointSegmentKM measures how far an analytic limit line deviates from the
// exported band boundary (beyond solarCentralBandSnapToleranceKM the ring edge wins). A
// limit line crossing the antimeridian must wrap its longitude difference to ±180°, or it
// reads as tens of thousands of kilometres away.
func TestSolarCentralBandPointSegmentWrapsAtAntimeridian(t *testing.T) {
	point := eclipsecore.SolarEclipsePathPoint{Longitude: 179.99, Latitude: 10}
	first := eclipsecore.SolarEclipsePathPoint{Longitude: 179.95, Latitude: 10}
	second := eclipsecore.SolarEclipsePathPoint{Longitude: -179.95, Latitude: 10}

	// 该点落在这一段的内部（180° 处是段内点），距离必须是 0。
	// The point lies inside the segment (which passes 180°), so the distance is zero.
	if got := solarCentralBandPointSegmentKM(point, first, second); math.Abs(got) > 0.5 {
		t.Fatalf("point on an antimeridian-crossing segment measured %v km away", got)
	}
	// 反向跨越（从 +179.95 到 -179.95 之外）同样要归约。
	// Wrapping must also hold when the segment runs the other way.
	if got := solarCentralBandPointSegmentKM(point, second, first); math.Abs(got) > 0.5 {
		t.Fatalf("reversed segment measured %v km away", got)
	}
	// 普通（非换日线）段的距离不受影响：与水平段相差 1° 纬度 ≈ 111.32 km。
	// An ordinary segment keeps its plain latitudinal distance: 1° of latitude off a
	// horizontal segment is 111.32 km.
	far := eclipsecore.SolarEclipsePathPoint{Longitude: 179.99, Latitude: 11}
	if got := solarCentralBandPointSegmentKM(far, first, second); math.Abs(got-111.32) > 0.5 {
		t.Fatalf("plain segment distance = %v km, want about 111.32 km", got)
	}
}

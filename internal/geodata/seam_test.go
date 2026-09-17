package geodata

import (
	"math"
	"testing"
)

// 居中后的等经纬接缝必须落在视图中心对面；旋转前后地理点集合应当一致。
func TestPolygonFragmentsCentredSeamKeepsGeometry(t *testing.T) {
	// 密集采样，避免矩形边在大圆上鼓向极点——库按大圆处理边，稀疏顶点会放大这个效应。
	ring := make([]GeoPoint, 0, 240)
	for longitude := -60.0; longitude <= 60; longitude += 2 {
		ring = append(ring, GeoPoint{Longitude: longitude, Latitude: -30})
	}
	for longitude := 60.0; longitude >= -60; longitude -= 2 {
		ring = append(ring, GeoPoint{Longitude: longitude, Latitude: 30})
	}
	plain := PolygonFragments(ring, ClipView{Projection: ProjectionEquirectangular})
	if len(plain) != 1 {
		t.Fatalf("uncentred: got %d fragments, want 1", len(plain))
	}
	// 中心放在 176.269°E 时接缝落到 -3.73°，这个环横跨接缝，应当被切成两段。
	centred := PolygonFragments(ring, ClipView{
		Projection: ProjectionEquirectangular,
		Center:     GeoPoint{Longitude: 176.269},
	})
	if len(centred) != 2 {
		t.Fatalf("centred: got %d fragments, want 2", len(centred))
	}
	for _, fragment := range centred {
		if len(fragment) < 3 {
			t.Fatalf("fragment has %d points", len(fragment))
		}
		for _, point := range fragment {
			if point.Longitude < -60-1e-6 || point.Longitude > 60+1e-6 {
				t.Fatalf("fragment longitude %.3f escaped the ring", point.Longitude)
			}
			if math.Abs(point.Latitude) > 30+0.5 {
				t.Fatalf("fragment latitude %.3f escaped the ring", point.Latitude)
			}
		}
	}
	// 两段拼起来的经度跨度和应等于原环的 120°。
	total := 0.0
	for _, fragment := range centred {
		minLon, maxLon := fragment[0].Longitude, fragment[0].Longitude
		for _, point := range fragment[1:] {
			minLon = math.Min(minLon, point.Longitude)
			maxLon = math.Max(maxLon, point.Longitude)
		}
		total += maxLon - minLon
	}
	if math.Abs(total-120) > 4 {
		t.Fatalf("fragment longitude span %.3f, want about 120", total)
	}
}

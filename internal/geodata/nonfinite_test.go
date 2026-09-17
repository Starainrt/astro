package geodata

import (
	"math"
	"testing"
)

// 裁剪入口对非有限坐标必须快速失败而不能死循环；超出 ±180 的有限经度按同一子午线归一化。
func TestTopologyRejectsNonFiniteCoordinates(t *testing.T) {
	cases := []struct {
		name   string
		points []GeoPoint
	}{
		{"huge-longitude", []GeoPoint{{0, 0}, {1e300, 10}, {20, 20}}},
		{"positive-infinity-longitude", []GeoPoint{{0, 0}, {math.Inf(1), 10}, {20, 20}}},
		{"negative-infinity-longitude", []GeoPoint{{0, 0}, {math.Inf(-1), 10}, {20, 20}}},
		{"nan-longitude", []GeoPoint{{0, 0}, {math.NaN(), 10}, {20, 20}}},
		{"nan-latitude", []GeoPoint{{0, 0}, {10, math.NaN()}, {20, 20}}},
		{"infinity-latitude", []GeoPoint{{0, 0}, {10, math.Inf(1)}, {20, 20}}},
	}
	views := []ClipView{
		{},
		{Projection: ProjectionEquirectangular},
		{Projection: ProjectionEquirectangular, Center: GeoPoint{Longitude: 100}},
		{Projection: ProjectionNorthPolar},
		{Projection: ProjectionSouthPolar},
		{Projection: ProjectionOrthographic, Center: GeoPoint{Longitude: 104, Latitude: -1.5}},
	}
	for _, testCase := range cases {
		wantsRejection := testCase.name != "huge-longitude"
		for _, view := range views {
			fragments := PolygonFragments(testCase.points, view)
			segments := PolylineSegments(testCase.points, view)
			if wantsRejection {
				if len(fragments) != 0 || len(segments) != 0 {
					t.Fatalf("%s %s: fragments=%d segments=%d, want none",
						testCase.name, view.Projection, len(fragments), len(segments))
				}
				continue
			}
			checkFinite := func(points []GeoPoint) {
				for _, point := range points {
					if math.IsNaN(point.Longitude) || math.IsNaN(point.Latitude) ||
						math.IsInf(point.Longitude, 0) || math.IsInf(point.Latitude, 0) ||
						math.Abs(point.Longitude) > 180 {
						t.Fatalf("%s %s: non-finite or unnormalized point %+v", testCase.name, view.Projection, point)
					}
				}
			}
			for _, fragment := range fragments {
				checkFinite(fragment)
			}
			for _, segment := range segments {
				checkFinite(segment)
			}
		}
	}
}

func TestTopologyNormalizesOutOfRangeLongitudeToSameMeridian(t *testing.T) {
	views := []ClipView{
		{},
		{Projection: ProjectionEquirectangular, Center: GeoPoint{Longitude: 30}},
		{Projection: ProjectionNorthPolar},
	}
	cases := [][2][]GeoPoint{
		{
			{{0, 0}, {190, 10}, {20, 20}},
			{{0, 0}, {-170, 10}, {20, 20}},
		},
		{
			{{350, -5}, {190, 10}, {181, 20}},
			{{-10, -5}, {-170, 10}, {-179, 20}},
		},
	}
	for _, pair := range cases {
		for _, view := range views {
			got := PolygonFragments(pair[0], view)
			want := PolygonFragments(pair[1], view)
			if len(got) != len(want) {
				t.Fatalf("%s: fragment count %d, want %d", view.Projection, len(got), len(want))
			}
			for index := range got {
				if len(got[index]) != len(want[index]) {
					t.Fatalf("%s: fragment %d length %d, want %d", view.Projection, index, len(got[index]), len(want[index]))
				}
				for point := range got[index] {
					if got[index][point] != want[index][point] {
						t.Fatalf("%s: fragment %d point %d = %+v, want %+v",
							view.Projection, index, point, got[index][point], want[index][point])
					}
				}
			}
		}
	}
}

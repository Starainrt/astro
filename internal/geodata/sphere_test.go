package geodata

import "testing"

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

func absoluteLongitude(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

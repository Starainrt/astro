package geojson_test

import (
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
)

// 3288-11-16 in the local display falls on 3288-11-15 UTC. It is an
// extremely shallow non-central eclipse: the sampled footprints are open,
// while the rise/set phase tracks still form part of the visible envelope.
func TestMarshalSolarEclipse32881115FallbackCoversPhaseTracks(t *testing.T) {
	path, ok := eclipse.SolarEclipsePartialFootprints(
		time.Date(3288, time.November, 15, 0, 0, 0, 0, time.UTC),
		eclipse.SolarEclipsePartialFootprintOptions{
			Step: 5 * time.Minute, BoundaryPoints: 96,
			CentralShadowStep: 5 * time.Minute, RiseSetStep: 2 * time.Minute,
		},
	)
	if !ok {
		t.Fatal("expected 3288-11-15 UTC solar eclipse")
	}
	if len(path.PartialBandContours) == 0 || len(path.RiseSetCurves) == 0 {
		t.Fatalf("unexpected topology: contours=%d curves=%d", len(path.PartialBandContours), len(path.RiseSetCurves))
	}
	data, err := geojson.MarshalSolarEclipse(path, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "partial-band")
	rings := geoJSONMultiPolygonOuterRings(t, band)
	lines := make([][]geodata.GeoPoint, 0, len(path.RiseSetCurves)*2)
	for _, curve := range path.RiseSetCurves {
		for _, segment := range curve.Segments {
			line := make([]geodata.GeoPoint, len(segment))
			for index, point := range segment {
				line[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			lines = append(lines, line)
		}
	}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(rings, lines, false); miss > 2 {
		t.Fatalf("fallback partial-band misses phase tracks by %.1f km", miss)
	}
}

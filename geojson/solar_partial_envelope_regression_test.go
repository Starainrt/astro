package geojson_test

import (
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
)

func TestSolarPartialEnvelopeCoversHorizonCurves(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(43, 3, 29, 0, 0, 0, 0, time.UTC),
		time.Date(1848, 9, 27, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 4, 20, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 3, 29, 0, 0, 0, 0, time.UTC),
	} {
		t.Run(date.Format("2006-01-02"), func(t *testing.T) {
			path, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 5 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 5 * time.Minute, RiseSetStep: 2 * time.Minute,
			})
			if !ok {
				t.Fatal("missing eclipse")
			}
			data, err := geojson.MarshalSolarEclipse(path, nil)
			if err != nil {
				t.Fatal(err)
			}
			collection := decodeCollection(t, data)
			band := featureWithRole(t, collection, "partial-band")
			if source := band.Properties["source"]; source != "zero-magnitude-envelope+horizon-boundary" {
				t.Fatalf("incomplete envelope fallback: %v", source)
			}
			assertClosedMultiPolygon(t, band)
			rings := geoJSONMultiPolygonOuterRings(t, band)
			var lines [][]geodata.GeoPoint
			for _, curve := range path.RiseSetCurves {
				for _, segment := range curve.Segments {
					line := make([]geodata.GeoPoint, len(segment))
					for i, point := range segment {
						line[i] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
					}
					lines = append(lines, line)
				}
			}
			if miss := geodata.SphericalPolygonsPathMissDistanceKM(rings, lines, false); miss > 2 {
				t.Fatalf("phase curves miss envelope by %.3f km", miss)
			}
		})
	}
}

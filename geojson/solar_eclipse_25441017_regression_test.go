package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
)

func TestMarshalSolarEclipse25441017PreservesHybridEnvelopeComponents(t *testing.T) {
	date := time.Date(2544, 10, 17, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step:              5 * time.Minute,
		BoundaryPoints:    96,
		CentralShadowStep: 5 * time.Minute,
		RiseSetStep:       5 * time.Minute,
		MagnitudeValues:   []float64{0.2, 0.4, 0.6, 0.8, 1},
	})
	if !ok || partial.Eclipse.Type != eclipse.SolarEclipseHybrid || len(partial.CentralBandSegments) != 3 {
		t.Fatalf("unexpected hybrid envelope: ok=%v type=%s segments=%d", ok, partial.Eclipse.Type, len(partial.CentralBandSegments))
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 5 * time.Minute, TargetSpacingKM: 700,
	})
	if !ok {
		t.Fatal("missing hybrid central path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatal(err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "central-band")
	if source := band.Properties["source"]; source != "besselian-critical-envelope-components" && source != "besselian-critical-envelope" {
		t.Fatalf("central band source=%v, want critical envelope", source)
	}
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatal(err)
	}
	if len(polygons) < 3 {
		t.Fatalf("central band polygons=%d, want at least three physical components", len(polygons))
	}
	for polygonIndex, polygon := range polygons {
		if len(polygon) != 1 {
			t.Fatalf("central band polygon %d rings=%d, want one", polygonIndex, len(polygon))
		}
		assertSolarCentralBandRingSimpleAndSampled(t, polygon[0], 250)
	}
	var sourceRings [][]geodata.GeoPoint
	for _, segment := range partial.CentralBandSegments {
		var ring []geodata.GeoPoint
		for _, point := range segment {
			ring = append(ring, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		}
		sourceRings = append(sourceRings, ring)
	}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(geoJSONMultiPolygonOuterRings(t, band), sourceRings, true); miss > 1 {
		t.Fatalf("hybrid components leave the exported band by %.3f km", miss)
	}
}

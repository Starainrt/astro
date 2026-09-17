package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func TestMarshalSolarEclipse20611013TangentCentralBandAvoidsAxisEndpointCap(t *testing.T) {
	date := time.Date(2061, 10, 13, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
		MagnitudeValues: []float64{0.2, 0.4, 0.6, 0.8, 1},
	})
	if !ok || partial.Eclipse.Type != eclipse.SolarEclipseAnnular ||
		partial.Eclipse.Centrality != eclipse.SolarEclipseCentralTwoLimits {
		t.Fatalf("unexpected event: ok=%v type=%s centrality=%s", ok, partial.Eclipse.Type, partial.Eclipse.Centrality)
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 2 * time.Minute, TargetSpacingKM: 700,
	})
	if !ok {
		t.Fatal("expected central path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "central-band")
	if source := band.Properties["source"]; source != "besselian-critical-envelope" {
		t.Fatalf("central-band source=%v, want a validated continuous central band", source)
	}
	var polygons [][][][]float64
	if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
		t.Fatalf("decode central band: %v", err)
	}
	if len(polygons) != 1 || len(polygons[0]) == 0 {
		t.Fatalf("central-band polygons=%d, want one exterior ring", len(polygons))
	}
	ring := polygons[0][0]
	if len(ring) < 20 {
		t.Fatalf("central-band ring points=%d, want a spatially refined envelope", len(ring))
	}
	assertSolarCentralBandRingSimpleAndSampled(t, ring, 250)
}

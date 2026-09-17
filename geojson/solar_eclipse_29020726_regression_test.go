package geojson_test

import (
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func TestMarshalSolarEclipse29020726DoesNotFillNighttimeCanada(t *testing.T) {
	date := time.Date(2902, time.July, 26, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 2 * time.Minute,
		MagnitudeValues: []float64{0.2, 0.4, 0.6, 0.8, 1.0},
	})
	if !ok {
		t.Fatal("expected solar eclipse footprints")
	}
	data, err := geojson.MarshalSolarEclipse(partial, nil)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "partial-band")
	if band.Geometry.Type != "MultiPolygon" {
		t.Fatalf("partial-band geometry=%q, want MultiPolygon", band.Geometry.Type)
	}
	for _, site := range []struct {
		name       string
		longitude  float64
		latitude   float64
		wantInside bool
	}{
		{name: "Edmonton", longitude: -113.4909, latitude: 53.5461},
		{name: "Churchill", longitude: -94.1650, latitude: 58.7684},
		{name: "Iqaluit", longitude: -68.5170, latitude: 63.7467},
		{name: "Resolute", longitude: -94.8297, latitude: 74.6973, wantInside: true},
		{name: "Alert", longitude: -62.3481, latitude: 82.5018, wantInside: true},
		{name: "Bering east of envelope", longitude: 179, latitude: 65},
		{name: "Bering inside excluded pocket", longitude: 175, latitude: 65},
		{name: "Bering west of envelope", longitude: 170, latitude: 65, wantInside: true},
		{name: "Bering west world", longitude: -179, latitude: 65},
		{name: "Arctic east world", longitude: 179.9, latitude: 80, wantInside: true},
	} {
		if got := geometryContainsPoint(t, band.Geometry, site.longitude, site.latitude); got != site.wantInside {
			t.Errorf("partial-band contains %s=%v, want %v", site.name, got, site.wantInside)
		}
	}
}

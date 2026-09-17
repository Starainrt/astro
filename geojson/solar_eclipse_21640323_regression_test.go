package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func TestMarshalSolarEclipse21640323ContainsVisibleAnnularEnds(t *testing.T) {
	date := time.Date(2164, 3, 23, 0, 0, 0, 0, time.UTC)
	sites := [][2]float64{{-110.6, 48.75}, {-110.6, 48.5}, {-110.7, 48.76}, {-120.575, 48.75}, {-120.575, 48.745}}
	for _, site := range sites {
		local, ok := eclipse.LocalSolarEclipseOnDate(date, site[0], site[1], 0)
		if !ok || local.Type != eclipse.SolarEclipseAnnular || local.SunAltitude <= 0 {
			t.Fatalf("reference site %v is not visible annular greatest: %+v", site, local)
		}
	}
	for _, step := range []time.Duration{time.Minute, 2 * time.Minute, 5 * time.Minute} {
		t.Run(step.String(), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: step, BoundaryPoints: 96, CentralShadowStep: step,
				MagnitudeValues: []float64{0.2, 0.4, 0.6, 0.8, 1},
			})
			if !ok || partial.Eclipse.Type != eclipse.SolarEclipseHybrid {
				t.Fatal("expected hybrid eclipse")
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: step, TargetSpacingKM: 700})
			if !ok {
				t.Fatal("expected central path")
			}
			data, err := geojson.MarshalSolarEclipse(partial, &central)
			if err != nil {
				t.Fatal(err)
			}
			band := featureWithRole(t, decodeCollection(t, data), "central-band")
			for _, site := range sites {
				if !geometryContainsPoint(t, band.Geometry, site[0], site[1]) {
					t.Errorf("central band omits visible annular site %v", site)
				}
			}
			var polygons [][][][]float64
			if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
				t.Fatal(err)
			}
			for _, polygon := range polygons {
				assertSolarCentralBandRingSimpleAndSampled(t, polygon[0], 250)
			}
			if len(partial.CentralBandHorizonClosures) != 2 {
				t.Fatalf("horizon closures=%d, want sunrise and sunset", len(partial.CentralBandHorizonClosures))
			}
			for _, closure := range partial.CentralBandHorizonClosures {
				assertSolarPathMaximumEdgeKM(t, closure, 12)
				for _, point := range closure {
					if distance := geoJSONMultiPolygonBoundaryDistanceKM(polygons, []float64{point.Longitude, point.Latitude}); distance > 0.02 {
						t.Fatalf("central band misses horizon closure by %.3f km", distance)
					}
				}
			}
		})
	}
}

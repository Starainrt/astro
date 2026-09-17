package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func TestSolarPolarProjectedPhaseCoverage(t *testing.T) {
	for _, date := range [][3]int{{163, 3, 22}, {564, 9, 21}, {936, 9, 18}, {1327, 9, 16}, {3429, 9, 21}, {3820, 9, 20}, {3932, 3, 20}, {4008, 3, 20}, {4026, 9, 24}} {
		day := time.Date(date[0], time.Month(date[1]), date[2], 0, 0, 0, 0, time.UTC)
		t.Run(day.Format("2006-01-02"), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(day, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 5 * time.Minute, CentralShadowStep: 5 * time.Minute, RiseSetStep: 2 * time.Minute, BoundaryPoints: 96,
			})
			if !ok {
				t.Fatal("missing eclipse")
			}
			data, err := geojson.MarshalSolarEclipse(partial, nil)
			if err != nil {
				t.Fatal(err)
			}
			band := featureWithRole(t, decodeCollection(t, data), "partial-band")
			var polygons [][][][]float64
			if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
				t.Fatal(err)
			}
			for _, curve := range partial.RiseSetCurves {
				for _, segment := range curve.Segments {
					for _, p := range segment {
						if !geoJSONMultiPolygonContains(polygons, p.Longitude, p.Latitude) {
							if miss := geoJSONMultiPolygonBoundaryDistanceKM(polygons, []float64{p.Longitude, p.Latitude}); miss > 2 {
								t.Fatalf("projected phase point %.8f/%.8f misses by %.3f km", p.Longitude, p.Latitude, miss)
							}
						}
					}
				}
			}
		})
	}
}

func TestSolarProjectedPolarAndPartialRingsDoNotCross(t *testing.T) {
	for _, date := range [][3]int{{33, 9, 12}, {767, 4, 3}, {1527, 11, 23}, {2459, 5, 3}, {2981, 10, 19}, {3284, 2, 8}} {
		day := time.Date(date[0], time.Month(date[1]), date[2], 0, 0, 0, 0, time.UTC)
		t.Run(day.Format("2006-01-02"), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(day, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 5 * time.Minute, CentralShadowStep: 5 * time.Minute, RiseSetStep: 2 * time.Minute, BoundaryPoints: 96,
			})
			if !ok {
				t.Fatal("missing eclipse")
			}
			central, hasCentral := eclipse.SolarEclipseCentralPath(day, eclipse.SolarEclipsePathOptions{Step: time.Minute, TargetSpacingKM: 500})
			var centralPath *eclipse.SolarEclipsePath
			if hasCentral {
				centralPath = &central
			}
			data, err := geojson.MarshalSolarEclipse(partial, centralPath)
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range decodeCollection(t, data).Features {
				if f.Properties["role"] != "partial-band" && f.Properties["role"] != "central-band" {
					continue
				}
				var polygons [][][][]float64
				if err := json.Unmarshal(f.Geometry.Coordinates, &polygons); err != nil {
					t.Fatal(err)
				}
				for pi, polygon := range polygons {
					for _, ring := range polygon {
						for i := 0; i+1 < len(ring); i++ {
							for j := i + 2; j+1 < len(ring); j++ {
								if i == 0 && j+1 == len(ring)-1 || math.Abs(ring[i][0]) == 180 && math.Abs(ring[i+1][0]) == 180 || math.Abs(ring[j][0]) == 180 && math.Abs(ring[j+1][0]) == 180 {
									continue
								}
								if geoJSONSegmentsCross(ring[i], ring[i+1], ring[j], ring[j+1]) {
									t.Fatalf("%s component %d crosses at %d/%d: %v %v / %v %v", f.Properties["role"], pi, i, j, ring[i], ring[i+1], ring[j], ring[j+1])
								}
							}
						}
					}
				}
			}
		})
	}
}

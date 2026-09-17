package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func TestSolarHybridThinComponentsDoNotCross(t *testing.T) {
	for _, date := range [][3]int{{881, 3, 4}, {985, 7, 20}, {1703, 1, 17}, {2386, 4, 29}, {3405, 7, 18}, {3667, 1, 7}} {
		day := time.Date(date[0], time.Month(date[1]), date[2], 0, 0, 0, 0, time.UTC)
		t.Run(day.Format("2006-01-02"), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(day, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 5 * time.Minute, CentralShadowStep: 5 * time.Minute, RiseSetStep: 2 * time.Minute, BoundaryPoints: 96,
			})
			if !ok {
				t.Fatal("missing solar eclipse")
			}
			data, err := geojson.MarshalSolarEclipse(partial, nil)
			if err != nil {
				t.Fatal(err)
			}
			band := featureWithRole(t, decodeCollection(t, data), "central-band")
			var polygons [][][][]float64
			if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
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
								t.Fatalf("component %d crosses itself at edges %d/%d", pi, i, j)
							}
						}
					}
				}
			}
		})
	}
}

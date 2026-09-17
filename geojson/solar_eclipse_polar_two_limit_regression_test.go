package geojson_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
)

func TestMarshalSolarEclipsePolarTwoLimitBandUsesSimpleFaces(t *testing.T) {
	for _, sample := range []struct {
		date        time.Time
		centralStep time.Duration
	}{
		{time.Date(767, time.April, 3, 0, 0, 0, 0, time.UTC), 5 * time.Minute},
		{time.Date(767, time.April, 3, 0, 0, 0, 0, time.UTC), time.Minute},
		{time.Date(2981, time.October, 19, 0, 0, 0, 0, time.UTC), 5 * time.Minute},
		{time.Date(2981, time.October, 19, 0, 0, 0, 0, time.UTC), time.Minute},
	} {
		t.Run(sample.date.Format("2006-01-02")+"/"+sample.centralStep.String(), func(t *testing.T) {
			date := sample.date
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 5 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 5 * time.Minute,
				MagnitudeValues: []float64{0.2, 0.4, 0.6, 0.8, 1},
			})
			if !ok {
				t.Fatal("expected solar eclipse footprints")
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
				Step: sample.centralStep, TargetSpacingKM: 500,
			})
			if !ok {
				t.Fatal("expected central path")
			}
			data, err := geojson.MarshalSolarEclipse(partial, &central)
			if err != nil {
				t.Fatalf("MarshalSolarEclipse: %v", err)
			}
			collection := decodeCollection(t, data)
			band := featureWithRole(t, collection, "central-band")
			assertClosedMultiPolygon(t, band)
			var polygons [][][][]float64
			if err := json.Unmarshal(band.Geometry.Coordinates, &polygons); err != nil {
				t.Fatalf("decode central-band: %v", err)
			}
			if len(polygons) == 0 {
				t.Fatal("polar central-band has no polygon")
			}
			for polygonIndex, polygon := range polygons {
				if len(polygon) == 0 {
					t.Fatalf("polygon %d has no exterior ring", polygonIndex)
				}
				ring := polygon[0]
				for first := 0; first+1 < len(ring); first++ {
					for second := first + 2; second+1 < len(ring); second++ {
						if first == 0 && second+1 == len(ring)-1 {
							continue
						}
						if polarAntimeridianFragmentEdge(ring[first], ring[first+1]) ||
							polarAntimeridianFragmentEdge(ring[second], ring[second+1]) {
							continue
						}
						if geoJSONSegmentsCross(ring[first], ring[first+1], ring[second], ring[second+1]) {
							t.Fatalf("polygon %d self-intersects between edges %d and %d", polygonIndex, first, second)
						}
					}
				}
			}
			rings := geoJSONMultiPolygonOuterRings(t, band)
			paths := make([][]geodata.GeoPoint, 0, len(central.CenterLine))
			for _, series := range [][]eclipse.SolarEclipsePathPoint{
				central.NorthernLimit, central.SouthernLimit, central.CenterLine,
			} {
				for _, point := range series {
					// Cross-section limits near the limb can have local greatest
					// below the horizon; they are not visible-band witnesses.
					local, ok := eclipse.GeometricLocalSolarEclipseOnDate(date, point.Longitude, point.Latitude, 0)
					if ok && local.VisibleAtGreatest && local.Type != eclipse.SolarEclipsePartial {
						paths = append(paths, []geodata.GeoPoint{{Longitude: point.Longitude, Latitude: point.Latitude}})
					}
				}
			}
			if len(paths) == 0 {
				t.Fatal("missing independently verified visible witnesses")
			}
			if miss := geodata.SphericalPolygonsPathMissDistanceKM(rings, paths, false); miss > 2 {
				t.Fatalf("central-band misses source path by %.3f km", miss)
			}
		})
	}
}

func polarAntimeridianFragmentEdge(first, second []float64) bool {
	return len(first) >= 2 && len(second) >= 2 &&
		(math.Abs(first[0]) == 180 || math.Abs(second[0]) == 180)
}

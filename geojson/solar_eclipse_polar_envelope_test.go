package geojson_test

import (
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
)

func TestSolarPolarBandPreservesContinuousEnvelope(t *testing.T) {
	for _, day := range [][3]int{
		{2003, 11, 23}, {2021, 12, 4}, {2039, 12, 15}, {2061, 10, 13}, {2981, 10, 19},
	} {
		date := time.Date(day[0], time.Month(day[1]), day[2], 0, 0, 0, 0, time.UTC)
		t.Run(date.Format("2006-01-02"), func(t *testing.T) {
			partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
				Step: 2 * time.Minute, BoundaryPoints: 96, DisableRiseSet: true,
			})
			if !ok || len(partial.CentralBandSegments) != 1 {
				t.Fatal("missing continuous envelope")
			}
			central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{Step: 2 * time.Minute, TargetSpacingKM: 700})
			if !ok {
				t.Fatal("missing central path")
			}
			for _, path := range []*eclipse.SolarEclipsePath{nil, &central} {
				data, err := geojson.MarshalSolarEclipse(partial, path)
				if err != nil {
					t.Fatal(err)
				}
				band := featureWithRole(t, decodeCollection(t, data), "central-band")
				source, _ := band.Properties["source"].(string)
				if !strings.Contains(source, "envelope") {
					t.Fatalf("continuous envelope replaced by %q", source)
				}
				rings := geoJSONMultiPolygonOuterRings(t, band)
				var centers []geodata.GeoPoint
				for _, point := range central.CenterLine {
					centers = append(centers, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
				}
				if miss := geodata.SphericalPolygonsPathMissDistanceKM(rings, [][]geodata.GeoPoint{centers}, false); miss > 2 {
					t.Fatalf("central band omits center line by %.3f km", miss)
				}
				assertClosedMultiPolygon(t, band)
			}
		})
	}
}

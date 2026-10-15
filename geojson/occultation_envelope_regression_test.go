package geojson_test

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/moon"
)

func TestOccultationDistantEnvelopesRemainComplete(t *testing.T) {
	if testing.Short() {
		t.Skip("slow regression")
	}
	if err := basic.LoadStarData(); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name             string
		year, month, day int
		planet           moon.OccultationPlanet
		hr               int
	}{
		{"mercury", 26, 4, 4, moon.OccultationMercury, 0},
		{"aldebaran", 26, 7, 26, "", 1457},
		{"regulus", 4026, 12, 14, "", 3982},
		{"saturn", 4026, 12, 29, moon.OccultationSaturn, 0},
		{"mars-fold", 1426, 7, 1, moon.OccultationMars, 0},
		{"venus-tangent", 1227, 5, 19, moon.OccultationVenus, 0},
		{"jupiter-horizon", 1226, 12, 23, moon.OccultationJupiter, 0},
		{"jupiter-time", 3627, 9, 20, moon.OccultationJupiter, 0},
	} {
		for _, algorithm := range []moon.OccultationPathAlgorithm{moon.OccultationPathAlgorithmExact, moon.OccultationPathAlgorithmOptimized} {
			t.Run(test.name+"/"+string(algorithm), func(t *testing.T) {
				day := time.Date(test.year, time.Month(test.month), test.day, 0, 0, 0, 0, time.UTC)
				options := moon.OccultationPathOptions{Algorithm: algorithm, Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true, RiseSetStep: time.Minute}
				var data []byte
				var err error
				role := "partial-band"
				if test.hr == 0 {
					paths, findErr := moon.FindPlanetOccultationPaths(day.Add(-12*time.Hour), day.Add(36*time.Hour), test.planet, options)
					if findErr != nil || len(paths) != 1 {
						t.Fatalf("paths=%d err=%v", len(paths), findErr)
					}
					data, err = geojson.MarshalPlanetOccultation(paths[0])
				} else {
					catalog, catalogErr := basic.StarDataByHR(test.hr)
					if catalogErr != nil {
						t.Fatal(catalogErr)
					}
					star, starErr := moon.StarCoordinateFromStarData(catalog)
					if starErr != nil {
						t.Fatal(starErr)
					}
					paths, findErr := moon.FindStarOccultationPaths(day.Add(-12*time.Hour), day.Add(36*time.Hour), star, options)
					if findErr != nil || len(paths) != 1 {
						t.Fatalf("paths=%d err=%v", len(paths), findErr)
					}
					data, err = geojson.MarshalStarOccultation(paths[0])
					role = "occultation-band"
				}
				if err != nil {
					t.Fatal(err)
				}
				collection := decodeCollection(t, data)
				assertCollectionCoordinates(t, collection)
				band := featureWithRole(t, collection, role)
				assertClosedMultiPolygon(t, band)
				if test.hr == 0 {
					assertOccultationP2PhaseCurvesInsidePartialBand(t, collection, 2)
					partial := geoJSONMultiPolygonOuterRings(t, band)
					total := geoJSONMultiPolygonOuterRings(t, featureWithRole(t, collection, "total-band"))
					if miss := geodata.SphericalPolygonsPathMissDistanceKM(partial, total, true); miss > 2 {
						t.Fatalf("total extends %.3f km beyond partial", miss)
					}
				}
			})
		}
	}
}

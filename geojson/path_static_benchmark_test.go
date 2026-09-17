package geojson_test

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

func BenchmarkStaticMap(b *testing.B) {
	if err := basic.LoadStarData(); err != nil {
		b.Fatal(err)
	}
	for _, test := range []struct {
		name             string
		year, month, day int
		planet           moon.OccultationPlanet
		hr               int
	}{
		{"saturn", 2025, 1, 5, moon.OccultationSaturn, 0},
		{"mars", 2025, 1, 14, moon.OccultationMars, 0},
		{"venus", 2025, 9, 19, moon.OccultationVenus, 0},
		{"star", 2025, 6, 5, "", 4799},
		{"partial", 2025, 3, 29, "", 0},
		{"hybrid", 2023, 4, 20, "", 0},
		{"annular2012", 2012, 5, 21, "", 0},
		{"annular2056", 2056, 7, 13, "", 0},
	} {
		b.Run(test.name, func(b *testing.B) {
			zone := time.FixedZone("UTC+8", 8*3600)
			day := time.Date(test.year, time.Month(test.month), test.day, 0, 0, 0, 0, zone)
			var star moon.StarCoordinate
			if test.hr > 0 {
				data, err := basic.StarDataByHR(test.hr)
				if err != nil {
					b.Fatal(err)
				}
				star, err = moon.StarCoordinateFromStarData(data)
				if err != nil {
					b.Fatal(err)
				}
			}
			options := moon.OccultationPathOptions{Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true, RiseSetStep: time.Minute}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var data []byte
				var err error
				if test.planet != "" {
					paths, findErr := moon.FindPlanetOccultationPaths(day, day.Add(24*time.Hour), test.planet, options)
					if findErr != nil || len(paths) != 1 {
						b.Fatalf("paths=%d err=%v", len(paths), findErr)
					}
					data, err = geojson.MarshalPlanetOccultation(paths[0])
				} else if test.hr > 0 {
					paths, findErr := moon.FindStarOccultationPaths(day, day.Add(24*time.Hour), star, options)
					if findErr != nil || len(paths) != 1 {
						b.Fatalf("paths=%d err=%v", len(paths), findErr)
					}
					data, err = geojson.MarshalStarOccultation(paths[0])
				} else {
					path, ok := eclipse.SolarEclipsePartialFootprints(day, eclipse.SolarEclipsePartialFootprintOptions{Step: 5 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 5 * time.Minute, RiseSetStep: 2 * time.Minute})
					if !ok {
						b.Fatal("missing eclipse")
					}
					data, err = geojson.MarshalSolarEclipse(path, nil)
				}
				if err != nil || len(data) == 0 {
					b.Fatalf("bytes=%d err=%v", len(data), err)
				}
			}
		})
	}
}

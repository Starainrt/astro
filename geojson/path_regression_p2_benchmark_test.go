package geojson_test

import (
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

func BenchmarkP2Mars20250729PlanetOccultationFind(b *testing.B) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	options := moon.OccultationPathOptions{
		Step:                     20 * time.Minute,
		TargetSpacingKM:          900,
		RiseSetStep:              time.Minute,
		DisableFootprints:        true,
		IncludeFootprintTimeline: true,
		FootprintTimelineStep:    5 * time.Minute,
	}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		paths, err := moon.FindPlanetOccultationPaths(
			start, start.Add(24*time.Hour), moon.OccultationMars, options,
		)
		if err != nil || len(paths) != 1 {
			b.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v", len(paths), err)
		}
	}
}

func BenchmarkP2Mars20250729PlanetOccultationGeoJSONMarshal(b *testing.B) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationMars,
		moon.OccultationPathOptions{
			Step:                     20 * time.Minute,
			TargetSpacingKM:          900,
			RiseSetStep:              time.Minute,
			DisableFootprints:        true,
			IncludeFootprintTimeline: true,
			FootprintTimelineStep:    5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		b.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v", len(paths), err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		data, err := geojson.MarshalPlanetOccultation(paths[0])
		if err != nil || len(data) == 0 {
			b.Fatalf("MarshalPlanetOccultation() bytes=%d err=%v", len(data), err)
		}
	}
}

func BenchmarkP2SolarEclipseGeoJSONMarshal(b *testing.B) {
	partial, ok := eclipse.SolarEclipsePartialFootprints(
		time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC),
		eclipse.SolarEclipsePartialFootprintOptions{
			Step: 10 * time.Minute, BoundaryPoints: 72, CentralShadowStep: 5 * time.Minute,
			DisableRiseSet: true,
		},
	)
	if !ok {
		b.Fatal("expected 2024-04-08 eclipse")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		data, err := geojson.MarshalSolarEclipse(partial, nil)
		if err != nil || len(data) == 0 {
			b.Fatalf("MarshalSolarEclipse() bytes=%d err=%v", len(data), err)
		}
	}
}

func BenchmarkP2StarOccultationGeoJSONMarshal(b *testing.B) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	star := moon.StarCoordinate{
		ID: "HR 4799", RA: 189.1975, Dec: -5.831944444444,
		Epoch:                          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          moon.CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28, ProperMotionDecMasPerYear: -18,
	}
	paths, err := moon.FindStarOccultationPaths(
		time.Date(2025, time.June, 5, 0, 0, 0, 0, zone),
		time.Date(2025, time.June, 6, 0, 0, 0, 0, zone), star,
		moon.OccultationPathOptions{Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true},
	)
	if err != nil || len(paths) != 1 {
		b.Fatalf("FindStarOccultationPaths() paths=%d err=%v", len(paths), err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		data, err := geojson.MarshalStarOccultation(paths[0])
		if err != nil || len(data) == 0 {
			b.Fatalf("MarshalStarOccultation() bytes=%d err=%v", len(data), err)
		}
	}
}

func BenchmarkP2PlanetOccultationGeoJSONMarshal(b *testing.B) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	paths, err := moon.FindPlanetOccultationPaths(
		time.Date(2025, time.January, 5, 0, 0, 0, 0, zone),
		time.Date(2025, time.January, 6, 0, 0, 0, 0, zone), moon.OccultationSaturn,
		moon.OccultationPathOptions{Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true},
	)
	if err != nil || len(paths) != 1 {
		b.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v", len(paths), err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		data, err := geojson.MarshalPlanetOccultation(paths[0])
		if err != nil || len(data) == 0 {
			b.Fatalf("MarshalPlanetOccultation() bytes=%d err=%v", len(data), err)
		}
	}
}

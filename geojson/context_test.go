package geojson

import (
	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/moon"
	"testing"
	"time"
)

type contextAPI interface {
	MarshalLunarEclipse(info eclipsecore.LunarEclipseInfo, boundaryPoints int) ([]byte, error)
	MarshalLunarEclipseWithOptions(
		info eclipsecore.LunarEclipseInfo,
		boundaryPoints int,
		options LunarEclipseOptions,
	) ([]byte, error)
	MarshalLunarEclipseWithTimeMarkers(
		info eclipsecore.LunarEclipseInfo,
		boundaryPoints int,
		options TimeMarkerOptions,
	) ([]byte, error)
	MarshalPlanetOccultation(path moon.PlanetOccultationPath) ([]byte, error)
	MarshalPlanetOccultationFootprints(instant moon.PlanetOccultationInstant) ([]byte, error)
	MarshalPlanetOccultationWithTimeMarkers(
		path moon.PlanetOccultationPath,
		options TimeMarkerOptions,
	) ([]byte, error)
	MarshalSolarEclipse(
		partial eclipsecore.SolarEclipsePartialFootprintsInfo,
		central *eclipsecore.SolarEclipsePath,
	) ([]byte, error)
	MarshalSolarEclipseShadowInstant(
		instant eclipsecore.SolarEclipseShadowInstant,
	) ([]byte, error)
	MarshalSolarEclipseWithOptions(
		partial eclipsecore.SolarEclipsePartialFootprintsInfo,
		central *eclipsecore.SolarEclipsePath,
		options SolarEclipseOptions,
	) ([]byte, error)
	MarshalSolarEclipseWithTimeMarkers(
		partial eclipsecore.SolarEclipsePartialFootprintsInfo,
		central *eclipsecore.SolarEclipsePath,
		options TimeMarkerOptions,
	) ([]byte, error)
	MarshalStarOccultation(path moon.StarOccultationPath) ([]byte, error)
	MarshalStarOccultationFootprint(instant moon.StarOccultationInstant) ([]byte, error)
	MarshalStarOccultationWithTimeMarkers(
		path moon.StarOccultationPath,
		options TimeMarkerOptions,
	) ([]byte, error)
}

var _ contextAPI = (*Generator)(nil)

func TestContextModuleIsolation(t *testing.T) {
	previous := basic.GetEphemerisModel()
	t.Cleanup(func() { basic.SetEphemerisModel(previous) })
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		basic.SetEphemerisModel(basic.EphemerisModel(model))
		info, ok := eclipsecore.LunarEclipseOnDateWithInstance(time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC), ephem.NewModelInstance(model))
		if !ok {
			t.Fatal("known lunar eclipse not found")
		}
		want, wantErr := MarshalLunarEclipse(info, 12)
		module := NewWithInstance(ephem.NewModelInstance(model))
		basic.SetEphemerisModel(basic.EphemerisModel(1 - int(model)))
		got, gotErr := module.MarshalLunarEclipse(info, 12)
		if gotErr != nil || wantErr != nil || string(got) != string(want) {
			t.Fatalf("model %d: GeoJSON escaped instance: %v/%v", model, gotErr, wantErr)
		}
	}
}

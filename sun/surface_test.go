package sun_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/sun"
)

func TestMarsSurfacePublicConsistencyPaths(t *testing.T) {
	previous := basic.GetEphemerisModel()
	defer basic.SetEphemerisModel(previous)
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		if err := astro.SetModel(model); err != nil {
			t.Fatal(err)
		}
		inst := ephem.NewModelInstance(ephem.Model(model))
		module := sun.NewWithInstance(inst)
		context, err := astro.NewContext(model)
		if err != nil {
			t.Fatal(err)
		}
		paths := [][2]func(time.Time, float64, float64) (time.Time, error){{sun.NextRiseOnMars, sun.NextSetOnMars}, {module.NextRiseOnMars, module.NextSetOnMars}, {context.Sun().NextRiseOnMars, context.Sun().NextSetOnMars}}
		observer := basic.BodyObserver{Body: basic.SkyMars, EastLongitude: 137.4, Latitude: -4.6}
		for _, zone := range []*time.Location{time.UTC, time.FixedZone("east", 8*3600), time.FixedZone("west", -10*3600)} {
			after := time.Date(2026, 11, 14, 0, 0, 0, 0, zone)
			for i, kind := range []basic.SurfaceEventKind{basic.SurfaceRise, basic.SurfaceSet} {
				event, err := basic.NextSurfaceEventWithInstance(basic.SkySun, observer, kind, basic.UTC2TT(basic.Date2JD(after.UTC())), basic.SurfaceOptions{}, inst)
				if err != nil {
					t.Fatal(err)
				}
				eventTime := basic.JD2DateByZone(basic.TT2UTC(event.JDE), zone, false)
				for _, seed := range []time.Time{after, eventTime.Add(-time.Minute), eventTime, eventTime.Add(time.Minute)} {
					core, err := basic.NextSurfaceEventWithInstance(basic.SkySun, observer, kind, basic.UTC2TT(basic.Date2JD(seed.UTC())), basic.SurfaceOptions{}, inst)
					if err != nil {
						t.Fatal(err)
					}
					want := basic.JD2DateByZone(basic.TT2UTC(core.JDE), zone, false)
					for _, path := range paths {
						got, err := path[i](seed, observer.EastLongitude, observer.Latitude)
						if err != nil || got != want || got.Location() != zone {
							t.Fatalf("model=%d kind=%d seed=%v got=%v want=%v err=%v", model, kind, seed, got, want, err)
						}
					}
				}
			}
		}
	}
}

type marsSurfaceErrorEngine struct{}

func (marsSurfaceErrorEngine) State(astro.Body, astro.Body, float64) ([3]float64, [3]float64, error) {
	return [3]float64{}, [3]float64{}, basic.ErrInvalidObservationInput
}
func (marsSurfaceErrorEngine) Coverage(astro.Body) (float64, float64, bool) {
	return 2400000, 2500000, true
}

func TestMarsSurfacePublicSyntheticInvalid(t *testing.T) {
	date := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	module := sun.NewWithInstance(nil)
	for _, query := range []func(time.Time, float64, float64) (time.Time, error){sun.NextRiseOnMars, sun.NextSetOnMars, module.NextRiseOnMars, module.NextSetOnMars} {
		for _, coords := range [][2]float64{{math.NaN(), 0}, {math.Inf(1), 0}, {0, math.NaN()}, {0, 91}} {
			got, err := query(date, coords[0], coords[1])
			if !got.IsZero() || !errors.Is(err, basic.ErrInvalidObservationInput) {
				t.Fatalf("coords=%v result=%v err=%v", coords, got, err)
			}
		}
	}
	engine := marsSurfaceErrorEngine{}
	bound := sun.NewWithInstance(ephem.NewInstance(engine))
	context, err := astro.NewContext(astro.ModelCustom, engine)
	if err != nil {
		t.Fatal(err)
	}
	coreInst := ephem.NewInstance(engine)
	_, expectedErr := basic.NextSurfaceEventWithInstance(basic.SkySun, basic.BodyObserver{Body: basic.SkyMars}, basic.SurfaceRise, basic.UTC2TT(basic.Date2JD(date.UTC())), basic.SurfaceOptions{}, coreInst)
	if expectedErr == nil {
		t.Fatal("error engine returned success")
	}
	for _, query := range []func(time.Time, float64, float64) (time.Time, error){bound.NextRiseOnMars, bound.NextSetOnMars, context.Sun().NextRiseOnMars, context.Sun().NextSetOnMars} {
		got, err := query(date, 0, 0)
		if !got.IsZero() || !errors.Is(err, expectedErr) {
			t.Fatalf("engine error result=%v err=%v expected=%v", got, err, expectedErr)
		}
	}
}

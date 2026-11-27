package mars_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/mars"
)

type marsTimeAPI struct {
	sol, coordinated, longitude func(time.Time) float64
	mean, trueTime              func(time.Time, float64) float64
}

func TestMarsTimePublicConsistencyPaths(t *testing.T) {
	previous := basic.GetEphemerisModel()
	defer basic.SetEphemerisModel(previous)
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		if err := astro.SetModel(model); err != nil {
			t.Fatal(err)
		}
		inst := ephem.NewModelInstance(ephem.Model(model))
		module := mars.NewWithInstance(inst)
		context, err := astro.NewContext(model)
		if err != nil {
			t.Fatal(err)
		}
		paths := []marsTimeAPI{{mars.SolDate, mars.CoordinatedTime, mars.SolarLongitude, mars.LocalMeanSolarTime, mars.LocalTrueSolarTime}, {module.SolDate, module.CoordinatedTime, module.SolarLongitude, module.LocalMeanSolarTime, module.LocalTrueSolarTime}, {context.Mars().SolDate, context.Mars().CoordinatedTime, context.Mars().SolarLongitude, context.Mars().LocalMeanSolarTime, context.Mars().LocalTrueSolarTime}}
		for _, date := range []time.Time{time.Date(1965, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(2000, 1, 6, 0, 0, 0, 0, time.UTC), time.Date(2026, 11, 14, 6, 7, 8, 900000000, time.UTC), time.Date(2105, 1, 1, 0, 0, 0, 0, time.UTC)} {
			jd := basic.UTC2TT(basic.Date2JD(date.UTC()))
			sol, err := basic.MarsSolDate(jd)
			if err != nil {
				t.Fatal(err)
			}
			coordinated, err := basic.MarsCoordinatedTime(jd)
			if err != nil {
				t.Fatal(err)
			}
			longitude, err := basic.MarsSolarLongitudeWithInstance(jd, inst)
			if err != nil {
				t.Fatal(err)
			}
			for _, zone := range []*time.Location{time.UTC, time.FixedZone("east", 14*3600), time.FixedZone("west", -12*3600)} {
				input := date.In(zone)
				for _, path := range paths {
					for _, pair := range [][2]float64{{path.sol(input), sol}, {path.coordinated(input), coordinated}, {path.longitude(input), longitude}} {
						if math.IsNaN(pair[0]) || math.IsInf(pair[0], 0) || math.Float64bits(pair[0]) != math.Float64bits(pair[1]) {
							t.Fatalf("model=%d date=%v bits %.17g/%.17g", model, input, pair[0], pair[1])
						}
					}
					for _, east := range []float64{-184.702, 0, 137.4, 720} {
						mean, err := basic.MarsLocalMeanSolarTime(jd, east)
						if err != nil {
							t.Fatal(err)
						}
						trueTime, err := basic.MarsLocalTrueSolarTimeWithInstance(jd, east, inst)
						if err != nil {
							t.Fatal(err)
						}
						for _, pair := range [][2]float64{{path.mean(input, east), mean}, {path.trueTime(input, east), trueTime}} {
							if math.IsNaN(pair[0]) || math.IsInf(pair[0], 0) || math.Float64bits(pair[0]) != math.Float64bits(pair[1]) {
								t.Fatalf("east=%g bits %.17g/%.17g", east, pair[0], pair[1])
							}
						}
					}
				}
			}
		}
		date := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
		frozen := module.LocalTrueSolarTime(date, 137.4)
		other := astro.ModelLegacy
		if model == other {
			other = astro.ModelSemianalytic
		}
		if err := astro.SetModel(other); err != nil {
			t.Fatal(err)
		}
		for _, path := range paths[1:] {
			if math.Float64bits(path.trueTime(date, 137.4)) != math.Float64bits(frozen) {
				t.Fatal("bound model changed with global")
			}
		}
	}
}

type marsTimeErrorEngine struct{}

func (marsTimeErrorEngine) State(astro.Body, astro.Body, float64) ([3]float64, [3]float64, error) {
	return [3]float64{}, [3]float64{}, basic.ErrInvalidObservationInput
}
func (marsTimeErrorEngine) Coverage(astro.Body) (float64, float64, bool) {
	return 2400000, 2500000, true
}

func TestMarsTimePublicSyntheticInvalid(t *testing.T) {
	date := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	module := mars.NewWithInstance(nil)
	for _, x := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		for _, query := range []func(time.Time, float64) float64{mars.LocalMeanSolarTime, mars.LocalTrueSolarTime, module.LocalMeanSolarTime, module.LocalTrueSolarTime} {
			if !math.IsNaN(query(date, x)) {
				t.Fatalf("invalid longitude %g did not return NaN", x)
			}
		}
	}
	engine := marsTimeErrorEngine{}
	context, err := astro.NewContext(astro.ModelCustom, engine)
	if err != nil {
		t.Fatal(err)
	}
	bound := mars.NewWithInstance(ephem.NewInstance(engine))
	for _, query := range []func(time.Time) float64{bound.SolarLongitude, context.Mars().SolarLongitude} {
		if !math.IsNaN(query(date)) {
			t.Fatal("engine error did not return NaN")
		}
	}
	for _, query := range []func(time.Time, float64) float64{bound.LocalTrueSolarTime, context.Mars().LocalTrueSolarTime} {
		if !math.IsNaN(query(date, 0)) {
			t.Fatal("engine error did not return NaN")
		}
	}
}

package astro_test

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/mars"
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/neptune"
	"github.com/starainrt/astro/saturn"
	"github.com/starainrt/astro/sun"
	"github.com/starainrt/astro/uranus"
	"github.com/starainrt/astro/venus"
)

func mustContext(tb testing.TB, model astro.Model, engines ...astro.Engine) *astro.Context {
	tb.Helper()
	context, err := astro.NewContext(model, engines...)
	if err != nil {
		tb.Fatalf("NewContext(%v): %v", model, err)
	}
	return context
}

func mustInstance(tb testing.TB, model astro.EphemerisModel) *astro.Instance {
	tb.Helper()
	inst, err := astro.NewModelInstance(model)
	if err != nil {
		tb.Fatalf("NewModelInstance(%v): %v", model, err)
	}
	return inst
}

type contextScalarCase struct {
	name   string
	global func(time.Time) float64
	bound  func(*astro.Context, time.Time) float64
}

func contextScalars() []contextScalarCase {
	return []contextScalarCase{
		{"moon", moon.ApparentLo, func(context *astro.Context, date time.Time) float64 { return context.Moon().ApparentLo(date) }},
		{"sun", sun.ApparentLo, func(context *astro.Context, date time.Time) float64 { return context.Sun().ApparentLo(date) }},
		{"mercury", mercury.ApparentLo, func(context *astro.Context, date time.Time) float64 { return context.Mercury().ApparentLo(date) }},
		{"venus", venus.ApparentLo, func(context *astro.Context, date time.Time) float64 { return context.Venus().ApparentLo(date) }},
		{"mars", mars.ApparentLo, func(context *astro.Context, date time.Time) float64 { return context.Mars().ApparentLo(date) }},
		{"jupiter", jupiter.ApparentLo, func(context *astro.Context, date time.Time) float64 { return context.Jupiter().ApparentLo(date) }},
		{"saturn", saturn.ApparentLo, func(context *astro.Context, date time.Time) float64 { return context.Saturn().ApparentLo(date) }},
		{"uranus", uranus.ApparentLo, func(context *astro.Context, date time.Time) float64 { return context.Uranus().ApparentLo(date) }},
		{"neptune", neptune.ApparentLo, func(context *astro.Context, date time.Time) float64 { return context.Neptune().ApparentLo(date) }},
		{"sun-distance", sun.EarthDistance, func(context *astro.Context, date time.Time) float64 { return context.Sun().EarthDistance(date) }},
		{"moon-distance", moon.EarthDistance, func(context *astro.Context, date time.Time) float64 { return context.Moon().EarthDistance(date) }},
		{"venus-distance", venus.SunDistance, func(context *astro.Context, date time.Time) float64 { return context.Venus().SunDistance(date) }},
	}
}

func TestContextIsolationAndGlobalModelSwitch(t *testing.T) {
	previous := astro.GetModel()
	t.Cleanup(func() { astro.SetModel(previous) })
	date := time.Date(2024, 10, 6, 12, 0, 0, 0, time.UTC)
	legacy := mustContext(t, astro.ModelLegacy)
	modern := mustContext(t, astro.ModelSemianalytic)
	for _, test := range contextScalars() {
		t.Run(test.name, func(t *testing.T) {
			astro.SetModel(astro.ModelLegacy)
			wantLegacy := test.global(date)
			gotLegacy := test.bound(legacy, date)
			astro.SetModel(astro.ModelSemianalytic)
			wantModern := test.global(date)
			gotModern := test.bound(modern, date)
			if math.IsNaN(gotLegacy) || math.IsNaN(gotModern) || math.Abs(gotLegacy-wantLegacy) > 1e-9 || math.Abs(gotModern-wantModern) > 1e-9 {
				t.Fatalf("legacy %.12f/%.12f; modern %.12f/%.12f", gotLegacy, wantLegacy, gotModern, wantModern)
			}
			if math.Abs(gotLegacy-gotModern) < 1e-12 {
				t.Fatalf("model switch did not change %s: %.12f", test.name, gotLegacy)
			}
			if test.bound(legacy, date) != gotLegacy || test.bound(modern, date) != gotModern {
				t.Fatal("explicit context changed with the global model")
			}
		})
	}
	if legacy.Model() != astro.ModelLegacy || modern.Model() != astro.ModelSemianalytic || astro.GetModel() != astro.ModelSemianalytic {
		t.Fatal("incorrect model identity")
	}
	if !basic.SetEphemerisModel(basic.EphemerisLegacy) || astro.GetModel() != astro.ModelLegacy {
		t.Fatal("existing basic model setter diverged from the root facade")
	}
}

func TestContextRisePhaseAndEclipseIsolation(t *testing.T) {
	previous := astro.GetModel()
	t.Cleanup(func() { astro.SetModel(previous) })
	zone := time.FixedZone("UTC+8", 8*3600)
	date := time.Date(2024, 4, 8, 0, 0, 0, 0, zone)
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		context := mustContext(t, model)
		astro.SetModel(model)
		wantRise, wantErr := moon.RiseTime(date, 116.4, 39.9, 0, true)
		wantSunrise, wantSunErr := sun.RiseTime(date, 116.4, 39.9, 0, true)
		wantPhase := moon.NextNewMoon(date)
		solarDate := time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)
		wantSolar, wantSolarOK := eclipse.SolarEclipseOnDate(solarDate)
		wantLunar, wantLunarOK := eclipse.LunarEclipseOnDateWithInstance(time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC), mustInstance(t, model))
		if !wantSolarOK || !wantLunarOK {
			t.Fatal("known solar or lunar eclipse was not found")
		}
		wantJSON, wantJSONErr := geojson.MarshalLunarEclipse(wantLunar, 24)
		astro.SetModel(astro.Model(1 - int(model)))
		gotRise, gotErr := context.Moon().RiseTime(date, 116.4, 39.9, 0, true)
		gotSunrise, gotSunErr := context.Sun().RiseTime(date, 116.4, 39.9, 0, true)
		if gotRise != wantRise || gotErr != wantErr || gotSunrise != wantSunrise || gotSunErr != wantSunErr || context.Moon().NextNewMoon(date) != wantPhase {
			t.Fatalf("model %d: rise or phase escaped context", model)
		}
		gotSolar, gotSolarOK := context.Eclipse().SolarEclipseOnDate(solarDate)
		gotLunar, gotLunarOK := context.Eclipse().LunarEclipseOnDate(time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC))
		if gotSolarOK != wantSolarOK || !reflect.DeepEqual(gotSolar, wantSolar) || gotLunarOK != wantLunarOK || !reflect.DeepEqual(gotLunar, wantLunar) {
			t.Fatalf("model %d: eclipse escaped context\nsolar got=%+v\nsolar want=%+v\nlunar got=%+v\nlunar want=%+v", model, gotSolar, wantSolar, gotLunar, wantLunar)
		}
		gotJSON, gotJSONErr := context.GeoJSON().MarshalLunarEclipse(wantLunar, 24)
		if (gotJSONErr == nil) != (wantJSONErr == nil) || !bytes.Equal(gotJSON, wantJSON) {
			t.Fatalf("model %d: lunar GeoJSON differs: %v/%v", model, gotJSONErr, wantJSONErr)
		}
	}
}

func TestContextConcurrentReadAndLazyModules(t *testing.T) {
	previous := astro.GetModel()
	t.Cleanup(func() { astro.SetModel(previous) })
	context := mustContext(t, astro.ModelSemianalytic)
	date := time.Date(2024, 10, 6, 12, 0, 0, 0, time.UTC)
	want := mustContext(t, astro.ModelSemianalytic).Moon().ApparentLo(date)
	var wait sync.WaitGroup
	for worker := 0; worker < 12; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for iteration := 0; iteration < 12; iteration++ {
				if context.Moon() != context.Moon() || context.Sun() != context.Sun() || context.Mercury() != context.Mercury() || context.Venus() != context.Venus() || context.Mars() != context.Mars() || context.Jupiter() != context.Jupiter() || context.Saturn() != context.Saturn() || context.Uranus() != context.Uranus() || context.Neptune() != context.Neptune() || context.Eclipse() != context.Eclipse() || context.GeoJSON() != context.GeoJSON() {
					t.Error("lazy module identity changed")
				}
				if context.Moon().ApparentLo(date) != want {
					t.Error("concurrent context calculation changed")
				}
			}
		}()
	}
	for iteration := 0; iteration < 12; iteration++ {
		astro.SetModel(astro.Model(iteration % 2))
	}
	wait.Wait()
}

func TestContextSolarShadowCacheAndGeoJSONIsolation(t *testing.T) {
	previous := astro.GetModel()
	t.Cleanup(func() { astro.SetModel(previous) })
	date := time.Date(2024, 4, 8, 18, 17, 0, 0, time.UTC)
	options := eclipse.SolarEclipseShadowSolverOptions{BoundaryPoints: 24, TargetSpacingKM: -1}
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		astro.SetModel(model)
		wantSolver := eclipse.NewSolarEclipseShadowSolver(options)
		want, wantOK := wantSolver.ShadowAt(date)
		if !wantOK || want.Empty() {
			t.Fatal("known solar eclipse has no shadow")
		}
		wantJSON, wantErr := geojson.MarshalSolarEclipseShadowInstant(want)
		context := mustContext(t, model)
		solver := context.Eclipse().NewSolarEclipseShadowSolver(options)
		before, beforeOK := solver.ShadowAt(date)
		astro.SetModel(astro.Model(1 - int(model)))
		after, afterOK := solver.ShadowAt(date)
		if !beforeOK || !afterOK || !reflect.DeepEqual(before, want) || !reflect.DeepEqual(after, want) {
			t.Fatalf("model %d: shadow solver cache escaped instance", model)
		}
		gotJSON, gotErr := context.GeoJSON().MarshalSolarEclipseShadowInstant(after)
		if wantErr != nil || gotErr != nil || !bytes.Equal(gotJSON, wantJSON) {
			t.Fatalf("model %d: solar GeoJSON escaped context: %v/%v", model, gotErr, wantErr)
		}
	}
}

func TestContextPackageCoordinateAllocations(t *testing.T) {
	previous := astro.GetModel()
	t.Cleanup(func() { astro.SetModel(previous) })
	date := time.Date(2024, 10, 6, 12, 0, 0, 0, time.UTC)
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		astro.SetModel(model)
		context := mustContext(t, model)
		for _, test := range contextScalars() {
			if count := testing.AllocsPerRun(10, func() { test.global(date) }); count != 0 {
				t.Errorf("model %d %s package allocations=%g", model, test.name, count)
			}
			if count := testing.AllocsPerRun(10, func() { test.bound(context, date) }); count != 0 {
				t.Errorf("model %d %s warm context allocations=%g", model, test.name, count)
			}
		}
	}
}

func TestContextPackageEventAllocations(t *testing.T) {
	previous := astro.GetModel()
	t.Cleanup(func() { astro.SetModel(previous) })
	astro.SetModel(astro.ModelLegacy)
	date := time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name    string
		compute func()
		maximum float64
	}{
		{"moon-rise", func() { moon.RiseTime(date, 116.4, 39.9, 0, true) }, 0},
		{"sun-rise", func() { sun.RiseTime(date, 116.4, 39.9, 0, true) }, 0},
		{"moon-phase", func() { moon.NextNewMoon(date) }, 0},
		{"venus-event", func() { venus.NextConjunction(date) }, 0},
		{"solar-eclipse", func() { eclipse.SolarEclipseOnDate(date) }, 27},
	} {
		if allocations := testing.AllocsPerRun(5, test.compute); allocations > test.maximum {
			t.Errorf("%s allocations=%g, maximum=%g", test.name, allocations, test.maximum)
		}
	}
}

type contextRejectingEngine struct{}

type contextRotatedEngine struct {
	astro.Engine
	apparent astro.ApparentEngine
}

func (engine contextRotatedEngine) Apparent(target astro.Body, jdTT float64) ([3]float64, float64, float64, error) {
	direction, distance, delay, err := engine.apparent.Apparent(target, jdTT)
	if err != nil {
		return direction, distance, delay, err
	}
	sine, cosine := math.Sincos(0.01)
	direction = [3]float64{cosine*direction[0] - sine*direction[1], sine*direction[0] + cosine*direction[1], direction[2]}
	return direction, distance, delay, nil
}

func TestContextCustomApparentEngineIsolation(t *testing.T) {
	previous := astro.GetModel()
	t.Cleanup(func() { astro.SetModel(previous) })
	engine := mustInstance(t, astro.ModelSemianalytic).Engine()
	apparent, ok := engine.(astro.ApparentEngine)
	if !ok {
		t.Fatal("built-in semi-analytic engine lacks apparent positions")
	}
	customEngine := contextRotatedEngine{Engine: engine, apparent: apparent}
	context := mustContext(t, astro.ModelCustom, customEngine)
	instance := astro.NewInstance(customEngine)
	date := time.Date(2024, 10, 6, 12, 0, 0, 0, time.UTC)
	jde := basic.UTC2TT(basic.Date2JD(date))
	want := basic.HMoonApparentLoWithInstance(jde, instance)
	ordinary := mustContext(t, astro.ModelSemianalytic).Moon().ApparentLo(date)
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		astro.SetModel(model)
		if got := context.Moon().ApparentLo(date); got != want || math.Abs(got-ordinary) < 0.01 {
			t.Fatalf("custom engine ignored or escaped context: got %.12f, want %.12f, ordinary %.12f", got, want, ordinary)
		}
	}
}

func (*contextRejectingEngine) State(target, center astro.Body, jdTDB float64) ([3]float64, [3]float64, error) {
	return [3]float64{}, [3]float64{}, astro.ErrEphemerisOutOfRange
}

func (*contextRejectingEngine) Coverage(target astro.Body) (float64, float64, bool) {
	return 0, 0, false
}

func TestContextCustomEngineAndInvalidConstruction(t *testing.T) {
	date := time.Date(2024, 10, 6, 12, 0, 0, 0, time.UTC)
	context := mustContext(t, astro.ModelCustom, &contextRejectingEngine{})
	if context.Model() != astro.ModelCustom || !math.IsNaN(context.Moon().ApparentLo(date)) || !math.IsNaN(context.Sun().ApparentLo(date)) || !math.IsNaN(context.Venus().SunDistance(date)) {
		t.Fatal("failed custom engine silently fell back to a built-in model")
	}
	var typedNil *contextRejectingEngine
	contextErr := func(model astro.Model, engines ...astro.Engine) func() error {
		return func() error {
			context, err := astro.NewContext(model, engines...)
			if context != nil {
				t.Error("failed construction returned a context")
			}
			return err
		}
	}
	stackErr := func(engines ...astro.Engine) func() error {
		return func() error {
			context, err := astro.NewStackContext(engines...)
			if context != nil {
				t.Error("failed construction returned a context")
			}
			return err
		}
	}
	for _, test := range []struct {
		name      string
		construct func() error
		want      error
	}{
		{"unknown", contextErr(astro.Model(99)), astro.ErrUnknownModel},
		{"missing-engine", contextErr(astro.ModelCustom), astro.ErrInvalidEngine},
		{"nil-engine", contextErr(astro.ModelCustom, nil), astro.ErrInvalidEngine},
		{"typed-nil-engine", contextErr(astro.ModelCustom, typedNil), astro.ErrInvalidEngine},
		{"extra-engine", contextErr(astro.ModelLegacy, &contextRejectingEngine{}), astro.ErrInvalidEngine},
		{"stack-nil-engine", stackErr(&contextRejectingEngine{}, nil), astro.ErrInvalidEngine},
		{"global-custom", func() error { return astro.SetModel(astro.ModelCustom) }, astro.ErrUnknownModel},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.construct(); !errors.Is(err, test.want) {
				t.Errorf("err = %v, want %v", err, test.want)
			}
		})
	}
	before := astro.GetModel()
	if astro.SetModel(astro.ModelCustom) == nil || astro.GetModel() != before {
		t.Error("rejected SetModel changed the global model")
	}
}

func TestZeroContextHasNoEngine(t *testing.T) {
	var zero astro.Context
	var nilContext *astro.Context
	date := time.Date(2024, 10, 6, 12, 0, 0, 0, time.UTC)
	for name, context := range map[string]*astro.Context{"zero": &zero, "nil": nilContext} {
		if context.Model() != astro.ModelCustom || !math.IsNaN(context.Moon().ApparentLo(date)) || !math.IsNaN(context.Sun().ApparentLo(date)) ||
			!math.IsNaN(context.Venus().SunDistance(date)) {
			t.Fatalf("%s context must have no engine", name)
		}
		if context.Mercury() == nil || context.Mars() == nil || context.Jupiter() == nil || context.Saturn() == nil || context.Uranus() == nil ||
			context.Neptune() == nil || context.Eclipse() == nil || context.GeoJSON() == nil || context.Phenomena() == nil {
			t.Fatalf("%s context accessor returned nil", name)
		}
		if _, err := context.Sun().RiseTime(date, 116.4, 39.9, 0, true); !errors.Is(err, basic.ErrInvalidObservationInput) {
			t.Fatalf("%s context rise/set err = %v", name, err)
		}
		if limited := context.WithLunarLimb(nil); limited == nil || limited.Model() != astro.ModelCustom {
			t.Fatalf("%s context WithLunarLimb lost the empty instance", name)
		}
	}
}

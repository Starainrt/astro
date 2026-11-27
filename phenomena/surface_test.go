package phenomena_test

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/phenomena"
)

type surfacePublicAPI struct {
	position     func(phenomena.Body, phenomena.BodyObserver, time.Time) (phenomena.SurfacePosition, error)
	next         [3]func(phenomena.Body, phenomena.BodyObserver, time.Time, phenomena.SurfaceOptions) (phenomena.SurfaceEvent, error)
	events       func(phenomena.Body, phenomena.BodyObserver, time.Time, time.Time, phenomena.SurfaceOptions) ([]phenomena.SurfaceEvent, phenomena.AltitudeState, error)
	starPosition func(phenomena.StarCoordinate, phenomena.BodyObserver, time.Time) (phenomena.SurfacePosition, error)
	starNext     [3]func(phenomena.StarCoordinate, phenomena.BodyObserver, time.Time) (phenomena.SurfaceEvent, error)
	starEvents   func(phenomena.StarCoordinate, phenomena.BodyObserver, time.Time, time.Time) ([]phenomena.SurfaceEvent, phenomena.AltitudeState, error)
	local        func(phenomena.BodyObserver, time.Time) (phenomena.SurfaceLocalTime, error)
}

func surfacePublicPaths(t *testing.T, model astro.Model) (*ephem.Instance, []surfacePublicAPI) {
	t.Helper()
	if err := astro.SetModel(model); err != nil {
		t.Fatal(err)
	}
	instance := ephem.NewModelInstance(ephem.Model(model))
	context, err := astro.NewContext(model)
	if err != nil {
		t.Fatal(err)
	}
	modules := []*phenomena.Phenomena{phenomena.NewWithInstance(instance), context.Phenomena()}
	paths := []surfacePublicAPI{{
		position: phenomena.SurfacePositionAt,
		next:     [3]func(phenomena.Body, phenomena.BodyObserver, time.Time, phenomena.SurfaceOptions) (phenomena.SurfaceEvent, error){phenomena.NextSurfaceRise, phenomena.NextSurfaceSet, phenomena.NextSurfaceCulmination},
		events:   phenomena.SurfaceEvents, starPosition: phenomena.StarSurfacePositionAt,
		starNext:   [3]func(phenomena.StarCoordinate, phenomena.BodyObserver, time.Time) (phenomena.SurfaceEvent, error){phenomena.NextStarSurfaceRise, phenomena.NextStarSurfaceSet, phenomena.NextStarSurfaceCulmination},
		starEvents: phenomena.StarSurfaceEvents, local: phenomena.SurfaceLocalTimeAt,
	}}
	for _, m := range modules {
		paths = append(paths, surfacePublicAPI{m.SurfacePositionAt, [3]func(phenomena.Body, phenomena.BodyObserver, time.Time, phenomena.SurfaceOptions) (phenomena.SurfaceEvent, error){m.NextSurfaceRise, m.NextSurfaceSet, m.NextSurfaceCulmination}, m.SurfaceEvents, m.StarSurfacePositionAt, [3]func(phenomena.StarCoordinate, phenomena.BodyObserver, time.Time) (phenomena.SurfaceEvent, error){m.NextStarSurfaceRise, m.NextStarSurfaceSet, m.NextStarSurfaceCulmination}, m.StarSurfaceEvents, m.SurfaceLocalTimeAt})
	}
	return instance, paths
}

func assertSurfacePublicBits(t *testing.T, got, want interface{}) {
	t.Helper()
	var compare func(reflect.Value, reflect.Value)
	compare = func(a, b reflect.Value) {
		if a.Type() != b.Type() {
			t.Fatalf("types %v/%v", a.Type(), b.Type())
		}
		if a.Type() == reflect.TypeOf(time.Time{}) {
			if a.Interface().(time.Time) != b.Interface().(time.Time) {
				t.Fatalf("time %v/%v", a, b)
			}
			return
		}
		switch a.Kind() {
		case reflect.Float64:
			x, y := a.Float(), b.Float()
			if math.IsNaN(x) || math.IsInf(x, 0) || math.Float64bits(x) != math.Float64bits(y) {
				t.Fatalf("bits %.17g / %.17g", x, y)
			}
		case reflect.Struct:
			for i := 0; i < a.NumField(); i++ {
				compare(a.Field(i), b.Field(i))
			}
		case reflect.Slice:
			if a.IsNil() != b.IsNil() || a.Len() != b.Len() {
				t.Fatalf("slice mismatch %v/%v", a, b)
			}
			for i := 0; i < a.Len(); i++ {
				compare(a.Index(i), b.Index(i))
			}
		default:
			if !reflect.DeepEqual(a.Interface(), b.Interface()) {
				t.Fatalf("values %v/%v", a, b)
			}
		}
	}
	compare(reflect.ValueOf(got), reflect.ValueOf(want))
}

func checkSurfacePublicEvent(t *testing.T, event phenomena.SurfaceEvent, core basic.SurfaceEvent, observer phenomena.BodyObserver, loc *time.Location, instance *ephem.Instance) {
	t.Helper()
	expectedTime := basic.JD2DateByZone(basic.TT2UTC(core.JDE), loc, false)
	if event.Time != expectedTime || event.Time.Location() != loc || event.Kind != core.Kind {
		t.Fatalf("event metadata %+v core=%+v expected=%v", event, core, expectedTime)
	}
	assertSurfacePublicBits(t, event.Position, core.Position)
	local, err := basic.SurfaceLocalTimeAtWithInstance(observer, basic.UTC2TT(basic.Date2JD(event.Time.UTC())), instance)
	if err != nil {
		t.Fatal(err)
	}
	assertSurfacePublicBits(t, event.Local, local)
	roundTrip := math.Abs(basic.UTC2TT(basic.Date2JD(event.Time.UTC()))-core.JDE) * 86400
	if roundTrip > 0.001 {
		t.Fatalf("TT roundtrip %.9g s", roundTrip)
	}
}

func TestSurfacePublicConsistencyPositions(t *testing.T) {
	previous := basic.GetEphemerisModel()
	defer basic.SetEphemerisModel(previous)
	star, err := phenomena.StarCoordinateByName("Sirius")
	if err != nil {
		t.Fatal(err)
	}
	date := time.Date(2026, 11, 14, 6, 7, 8, 900000000, time.UTC)
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		inst, paths := surfacePublicPaths(t, model)
		for body := phenomena.Mercury; body <= phenomena.Neptune; body++ {
			observer := phenomena.BodyObserver{Body: body, EastLongitude: 30, Latitude: -25, Height: 2}
			for _, zone := range []*time.Location{time.UTC, time.FixedZone("east", 14*3600), time.FixedZone("west", -12*3600)} {
				input := date.In(zone)
				jd := basic.UTC2TT(basic.Date2JD(input.UTC()))
				local, err := basic.SurfaceLocalTimeAtWithInstance(observer, jd, inst)
				if err != nil {
					t.Fatal(err)
				}
				for _, path := range paths {
					got, err := path.local(observer, input)
					if err != nil {
						t.Fatal(err)
					}
					assertSurfacePublicBits(t, got, local)
				}
				for _, target := range []phenomena.Body{phenomena.Sun, phenomena.Earth, phenomena.Mercury, phenomena.Venus, phenomena.Mars, phenomena.Jupiter, phenomena.Saturn, phenomena.Uranus, phenomena.Neptune} {
					if target == body {
						continue
					}
					want, err := basic.SurfacePositionAtWithInstance(target, observer, jd, inst)
					if err != nil {
						t.Fatal(err)
					}
					for _, path := range paths {
						got, err := path.position(target, observer, input)
						if err != nil {
							t.Fatal(err)
						}
						assertSurfacePublicBits(t, got, want)
					}
				}
				want, err := basic.StarSurfacePositionAtWithInstance(star, observer, jd, inst)
				if err != nil {
					t.Fatal(err)
				}
				for _, path := range paths {
					got, err := path.starPosition(star, observer, input)
					if err != nil {
						t.Fatal(err)
					}
					assertSurfacePublicBits(t, got, want)
				}
			}
		}
		observer := phenomena.BodyObserver{Body: phenomena.Mars, EastLongitude: 137.4, Latitude: -4.6}
		frozen, err := paths[1].position(phenomena.Earth, observer, date)
		if err != nil {
			t.Fatal(err)
		}
		other := astro.ModelLegacy
		if model == other {
			other = astro.ModelSemianalytic
		}
		if err := astro.SetModel(other); err != nil {
			t.Fatal(err)
		}
		for _, path := range paths[1:] {
			got, err := path.position(phenomena.Earth, observer, date)
			if err != nil {
				t.Fatal(err)
			}
			assertSurfacePublicBits(t, got, frozen)
		}
	}
}

func TestSurfacePublicConsistencyEvents(t *testing.T) {
	previous := basic.GetEphemerisModel()
	defer basic.SetEphemerisModel(previous)
	star, err := phenomena.StarCoordinateByName("Sirius")
	if err != nil {
		t.Fatal(err)
	}
	for _, model := range []astro.Model{astro.ModelLegacy, astro.ModelSemianalytic} {
		inst, paths := surfacePublicPaths(t, model)
		observer := phenomena.BodyObserver{Body: phenomena.Mars, EastLongitude: 137.4, Latitude: -4.6}
		for _, zone := range []*time.Location{time.UTC, time.FixedZone("east", 8*3600), time.FixedZone("west", -10*3600)} {
			start := time.Date(2026, 11, 14, 0, 0, 0, 0, zone)
			end := start.Add(48 * time.Hour)
			a, b := basic.UTC2TT(basic.Date2JD(start.UTC())), basic.UTC2TT(basic.Date2JD(end.UTC()))
			options := phenomena.SurfaceOptions{Limb: phenomena.LimbCenter, TargetAltitude: -6}
			want, state, err := basic.SurfaceEventsWithInstance(phenomena.Earth, observer, a, b, options, inst)
			if err != nil || len(want) == 0 {
				t.Fatalf("events %d %v", len(want), err)
			}
			for _, path := range paths {
				got, s, err := path.events(phenomena.Earth, observer, start, end, options)
				if err != nil || s != state || len(got) != len(want) {
					t.Fatalf("events %d/%d state=%v/%v err=%v", len(got), len(want), s, state, err)
				}
				for i, event := range got {
					checkSurfacePublicEvent(t, event, want[i], observer, zone, inst)
					local, err := path.local(observer, event.Time)
					if err != nil {
						t.Fatal(err)
					}
					assertSurfacePublicBits(t, event.Local, local)
				}
			}
			swant, sstate, err := basic.StarSurfaceEventsWithInstance(star, observer, a, b, basic.SurfaceOptions{}, inst)
			if err != nil || len(swant) == 0 {
				t.Fatalf("star events %d %v", len(swant), err)
			}
			for _, path := range paths {
				got, s, err := path.starEvents(star, observer, start, end)
				if err != nil || s != sstate || len(got) != len(swant) {
					t.Fatalf("star events %d/%d state=%v/%v err=%v", len(got), len(swant), s, sstate, err)
				}
				for i, event := range got {
					checkSurfacePublicEvent(t, event, swant[i], observer, zone, inst)
				}
			}
			for i, kind := range []basic.SurfaceEventKind{basic.SurfaceRise, basic.SurfaceSet, basic.SurfaceUpperCulmination} {
				core, err := basic.NextSurfaceEventWithInstance(phenomena.Sun, observer, kind, a, options, inst)
				if err != nil {
					t.Fatal(err)
				}
				starCore, err := basic.NextStarSurfaceEventWithInstance(star, observer, kind, a, basic.SurfaceOptions{}, inst)
				if err != nil {
					t.Fatal(err)
				}
				for _, seed := range []time.Time{start, basic.JD2DateByZone(basic.TT2UTC(core.JDE), zone, false).Add(-time.Minute), basic.JD2DateByZone(basic.TT2UTC(core.JDE), zone, false), basic.JD2DateByZone(basic.TT2UTC(core.JDE), zone, false).Add(time.Minute)} {
					seedJD := basic.UTC2TT(basic.Date2JD(seed.UTC()))
					expected, err := basic.NextSurfaceEventWithInstance(phenomena.Sun, observer, kind, seedJD, options, inst)
					if err != nil {
						t.Fatal(err)
					}
					for _, path := range paths {
						got, err := path.next[i](phenomena.Sun, observer, seed, options)
						if err != nil {
							t.Fatal(err)
						}
						checkSurfacePublicEvent(t, got, expected, observer, zone, inst)
					}
				}
				for _, offset := range []time.Duration{-time.Minute, 0, time.Minute} {
					seed := basic.JD2DateByZone(basic.TT2UTC(starCore.JDE), zone, false).Add(offset)
					expected, err := basic.NextStarSurfaceEventWithInstance(star, observer, kind, basic.UTC2TT(basic.Date2JD(seed.UTC())), basic.SurfaceOptions{}, inst)
					if err != nil {
						t.Fatal(err)
					}
					for _, path := range paths {
						got, err := path.starNext[i](star, observer, seed)
						if err != nil {
							t.Fatal(err)
						}
						checkSurfacePublicEvent(t, got, expected, observer, zone, inst)
					}
				}
			}
		}
	}
}

func TestSurfacePublicConsistencyWrappedLongitude(t *testing.T) {
	date := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
	for _, pair := range [][2]float64{{280, 1e18}, {80, -1e18}} {
		base := phenomena.BodyObserver{Body: phenomena.Mars, EastLongitude: pair[0]}
		wrapped := phenomena.BodyObserver{Body: phenomena.Mars, EastLongitude: pair[1]}
		position, err := phenomena.SurfacePositionAt(phenomena.Sun, base, date)
		wrappedPosition, wrappedErr := phenomena.SurfacePositionAt(phenomena.Sun, wrapped, date)
		if err != nil || wrappedErr != nil || wrappedPosition != position {
			t.Errorf("longitudes %v: positions %+v / %+v, errors %v / %v", pair, position, wrappedPosition, err, wrappedErr)
		}
		event, err := phenomena.NextSurfaceCulmination(phenomena.Sun, base, date, phenomena.SurfaceOptions{})
		wrappedEvent, wrappedErr := phenomena.NextSurfaceCulmination(phenomena.Sun, wrapped, date, phenomena.SurfaceOptions{})
		if err != nil || wrappedErr != nil || wrappedEvent != event {
			t.Errorf("longitudes %v: culminations %s / %s, errors %v / %v", pair, event.Time, wrappedEvent.Time, err, wrappedErr)
		}
		if math.Abs(wrappedEvent.Local.LocalTrueSolarTime-12) > 1e-6 {
			t.Errorf("longitude %g: LTST at upper culmination = %.12f", pair[1], wrappedEvent.Local.LocalTrueSolarTime)
		}
	}
}

func TestSurfacePublicSyntheticInvalid(t *testing.T) {
	previous := basic.GetEphemerisModel()
	defer basic.SetEphemerisModel(previous)
	date := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	_, paths := surfacePublicPaths(t, astro.ModelSemianalytic)
	observer := phenomena.BodyObserver{Body: phenomena.Mars}
	star, err := phenomena.StarCoordinateByName("Sirius")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		for _, sample := range []struct {
			target   phenomena.Body
			observer phenomena.BodyObserver
		}{{phenomena.Moon, observer}, {phenomena.Sun, phenomena.BodyObserver{Body: phenomena.Moon}}, {phenomena.Earth, phenomena.BodyObserver{}}, {phenomena.Mars, observer}} {
			got, err := path.position(sample.target, sample.observer, date)
			if !errors.Is(err, basic.ErrInvalidObservationInput) || got != (phenomena.SurfacePosition{}) {
				t.Fatalf("invalid position %+v %v", got, err)
			}
			list, _, err := path.events(sample.target, sample.observer, date, date.Add(time.Hour), phenomena.SurfaceOptions{})
			if !errors.Is(err, basic.ErrInvalidObservationInput) || list != nil {
				t.Fatalf("invalid events %+v %v", list, err)
			}
			for _, next := range path.next {
				got, err := next(sample.target, sample.observer, date, phenomena.SurfaceOptions{})
				if !errors.Is(err, basic.ErrInvalidObservationInput) || got != (phenomena.SurfaceEvent{}) {
					t.Fatalf("invalid event %+v %v", got, err)
				}
			}
		}
		invalid := phenomena.BodyObserver{Body: phenomena.Moon}
		local, err := path.local(invalid, date)
		if !errors.Is(err, basic.ErrInvalidObservationInput) || local != (phenomena.SurfaceLocalTime{}) {
			t.Fatalf("invalid local %+v %v", local, err)
		}
		position, err := path.starPosition(star, invalid, date)
		if !errors.Is(err, basic.ErrInvalidObservationInput) || position != (phenomena.SurfacePosition{}) {
			t.Fatalf("invalid star position %+v %v", position, err)
		}
		for _, next := range path.starNext {
			got, err := next(star, invalid, date)
			if !errors.Is(err, basic.ErrInvalidObservationInput) || got != (phenomena.SurfaceEvent{}) {
				t.Fatalf("invalid star event %+v %v", got, err)
			}
		}
		events, _, err := path.starEvents(star, invalid, date, date.Add(time.Hour))
		if !errors.Is(err, basic.ErrInvalidObservationInput) || events != nil {
			t.Fatalf("invalid star list %+v %v", events, err)
		}
		events, _, err = path.events(phenomena.Sun, observer, date.Add(time.Hour), date, phenomena.SurfaceOptions{})
		if !errors.Is(err, basic.ErrInvalidObservationInput) || events != nil {
			t.Fatalf("reversed list %+v %v", events, err)
		}
	}
}

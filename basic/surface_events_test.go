package basic

import (
	"github.com/starainrt/astro/internal/ephem"
	"encoding/json"
	"errors"
	"math"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"
)

type surfaceFixtureObserver struct {
	Body          string
	EastLongitude float64 `json:"east_longitude"`
	Latitude      float64
	Height        float64 `json:"height_km"`
}

func TestMercuryNextRiseSolarDayConsistency(t *testing.T) {
	observer := BodyObserver{Body: SkyMercury}
	instance := ephem.NewModelInstance(ephem.Semianalytic)
	after := UTC2TT(Date2JD(time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)))
	events, _, err := SurfaceEventsWithInstance(SkySun, observer, after, after+200, SurfaceOptions{}, instance)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range events {
		if want.Kind != SurfaceRise {
			continue
		}
		if want.JDE-after <= 1.1*orbitalPeriodDays[SkyMercury] {
			t.Fatalf("fixture did not exercise the old horizon: %.3f days", want.JDE-after)
		}
		got, err := NextSurfaceEventWithInstance(SkySun, observer, SurfaceRise, after, SurfaceOptions{}, instance)
		if err != nil || got != want {
			t.Fatalf("next=%+v, batch=%+v, err=%v", got, want, err)
		}
		return
	}
	t.Fatal("no batch sunrise")
}

type surfaceFixtureGroup struct {
	Label     string
	Observer  surfaceFixtureObserver
	Target    int `json:"target_naif"`
	Samples   [][3]float64
	Crossings []struct {
		Kind    string
		JD      float64    `json:"jd_tt"`
		Bracket [2]float64 `json:"sample_jd_tt_bracket"`
	}
}

func surfaceFixture(t *testing.T) (groups []surfaceFixtureGroup) {
	t.Helper()
	raw, err := os.ReadFile("testdata/surface_horizons.json")
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		Requests []surfaceFixtureGroup
		Curves   []surfaceFixtureGroup
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	for _, g := range data.Requests {
		if len(g.Samples) > 0 {
			groups = append(groups, g)
		}
	}
	return append(groups, data.Curves...)
}
func surfaceFixtureBodies(g surfaceFixtureGroup) (SkyBody, BodyObserver) {
	bodies := map[string]SkyBody{"Mercury": SkyMercury, "Venus": SkyVenus, "Mars": SkyMars, "Jupiter": SkyJupiter, "Saturn": SkySaturn, "Uranus": SkyUranus, "Neptune": SkyNeptune}
	targets := map[int]SkyBody{10: SkySun, 399: SkyEarth, 599: SkyJupiter, 699: SkySaturn}
	return targets[g.Target], BodyObserver{Body: bodies[g.Observer.Body], EastLongitude: g.Observer.EastLongitude, Latitude: g.Observer.Latitude, Height: g.Observer.Height}
}
func surfaceAngularDifference(a, b Vector3) float64 {
	return math.Atan2(vectorMagnitude(pxp(a, b)), vectorDot(a, b)) * 180 / math.Pi * 3600
}
func surfaceHorizontalVector(alt, az float64) Vector3 {
	a, z := alt*math.Pi/180, az*math.Pi/180
	return Vector3{math.Cos(a) * math.Cos(z), math.Cos(a) * math.Sin(z), math.Sin(a)}
}
func TestSurfacePositionExternal(t *testing.T) {
	const innerLimit, outerLimit, legacyLimit = 5.0, 10.0, 30.0
	groups := surfaceFixture(t)
	for _, model := range []ephem.Model{ephem.Semianalytic, ephem.Legacy} {
		inst := ephem.NewModelInstance(model)
		maximum := 0.0
		count := 0
		for _, g := range groups {
			target, observer := surfaceFixtureBodies(g)
			for _, sample := range g.Samples {
				got, err := SurfacePositionAtWithInstance(target, observer, sample[0], inst)
				if err != nil {
					t.Fatal(err)
				}
				delta := surfaceAngularDifference(surfaceHorizontalVector(got.Altitude, got.Azimuth), surfaceHorizontalVector(sample[2], sample[1]))
				limit := legacyLimit
				if model == ephem.Semianalytic {
					limit = innerLimit
					if observer.Body >= SkyJupiter {
						limit = outerLimit
					}
				}
				if delta > limit {
					t.Fatalf("%s jd=%g model=%d difference=%g arcsec", g.Label, sample[0], model, delta)
				}
				maximum = math.Max(maximum, delta)
				count++

			}
		}
		if count != 4161 {
			t.Fatalf("sample count=%d", count)
		}
		t.Logf("model=%d samples=%d maximum angular difference=%.9f arcsec", model, count, maximum)
	}
}
func TestSurfacePositionConsistencyInstances(t *testing.T) {
	previous := GetEphemerisModel()
	defer SetEphemerisModel(previous)
	for _, model := range []EphemerisModel{EphemerisLegacy, EphemerisSemianalytic} {
		SetEphemerisModel(model)
		inst := ephem.NewModelInstance(ephem.Model(model))
		for body := SkyMercury; body <= SkyNeptune; body++ {
			observer := BodyObserver{Body: body, EastLongitude: 31, Latitude: 25, Height: 1}
			a, e := SurfacePositionAt(SkySun, observer, 2461319.5)
			b, f := SurfacePositionAtWithInstance(SkySun, observer, 2461319.5, inst)
			if e != nil || f != nil || a != b {
				t.Fatalf("body=%d got=%+v expected=%+v err=%v/%v", body, a, b, e, f)
			}
			for _, v := range []float64{a.Altitude, a.Azimuth, a.HourAngle, a.BodyRA, a.BodyDec, a.DistanceAU, a.AngularRadius} {
				if !isFiniteFloat(v) {
					t.Fatal(a)
				}
			}
			if a.HourAngle <= -180 || a.HourAngle > 180 || a.Azimuth < 0 || a.Azimuth >= 360 || a.BodyRA < 0 || a.BodyRA >= 360 {
				t.Fatal(a)
			}
			star := StarCoordinate{RA: 50, Dec: -15, Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameICRS}
			c, e := StarSurfacePositionAt(star, observer, 2461319.5)
			d, f := StarSurfacePositionAtWithInstance(star, observer, 2461319.5, inst)
			if e != nil || f != nil || c != d || c.DistanceAU != 0 || c.AngularRadius != 0 {
				t.Fatalf("star=%+v/%+v %v/%v", c, d, e, f)
			}
		}
	}
}
func TestSurfacePositionSyntheticProjection(t *testing.T) {
	for _, rate := range []float64{300, -300} {
		frame := bodyObserverFrame{up: Vector3{1, 0, 0}, east: Vector3{0, 1, 0}, north: Vector3{0, 0, 1}, orientation: BodyOrientation{PoleRA: 0, PoleDec: 90, PrimeMeridianAngle: 40, RotationRate: rate}}
		p := surfacePosition(surfaceDirection{direction: normalizeVector(Vector3{1, 1, 1}), distanceKM: 149597870.7, frame: frame}, BodyObserver{EastLongitude: 20}, true)
		if math.Abs(p.Altitude-35.264389682754654) > 1e-12 || math.Abs(p.Azimuth-45) > 1e-12 || math.Abs(p.BodyRA-315) > 1e-12 || math.Abs(p.BodyDec-35.264389682754654) > 1e-12 || math.Abs(p.AngularRadius-695700/149597870.7*180/math.Pi) > 1e-14 {
			t.Fatal(p)
		}
		expected := angleDifference(60-p.BodyRA, 0) * math.Copysign(1, rate)
		if math.Abs(angleDifference(p.HourAngle, expected)) > 1e-12 {
			t.Fatal(p)
		}
	}
	for _, body := range []SkyBody{SkyVenus, SkyUranus} {
		observer := BodyObserver{Body: body}
		a, e := SurfacePositionAtWithInstance(SkySun, observer, 2461319.5, ephem.NewModelInstance(ephem.Semianalytic))
		b, f := SurfacePositionAtWithInstance(SkySun, observer, 2461319.50001, ephem.NewModelInstance(ephem.Semianalytic))
		if e != nil || f != nil || angleDifference(b.HourAngle, a.HourAngle) <= 0 {
			t.Fatalf("retrograde H=%g/%g err=%v/%v", a.HourAngle, b.HourAngle, e, f)
		}
	}
}
func TestSurfacePositionSyntheticInvalid(t *testing.T) {
	for _, sample := range []struct {
		target   SkyBody
		observer BodyObserver
		jd       float64
	}{
		{SkySun, BodyObserver{Body: SkyEarth}, 2451545}, {SkyMars, BodyObserver{Body: SkyMars}, 2451545}, {SkyMoon, BodyObserver{Body: SkyMars}, 2451545}, {SkySun, BodyObserver{Body: SkyMars}, math.NaN()},
	} {
		p, e := SurfacePositionAt(sample.target, sample.observer, sample.jd)
		if e == nil || p != (SurfacePosition{}) {
			t.Fatalf("%+v %v", p, e)
		}
	}
	p, e := StarSurfacePositionAt(StarCoordinate{}, BodyObserver{Body: SkyMars}, 2451545)
	if e == nil || p != (SurfacePosition{}) {
		t.Fatal(p, e)
	}
}

func TestSurfacePositionConsistencySubsolar(t *testing.T) {
	const limit = 0.02
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	functions := map[SkyBody]func(float64, int, *Instance) PlanetPhysicalInfo{SkyMercury: MercuryPhysicalNWithInstance, SkyVenus: VenusPhysicalNWithInstance, SkyMars: MarsPhysicalNWithInstance, SkyJupiter: JupiterPhysicalNWithInstance, SkySaturn: SaturnPhysicalNWithInstance, SkyUranus: UranusPhysicalNWithInstance, SkyNeptune: NeptunePhysicalNWithInstance}
	maximum := 0.0
	for body, physical := range functions {
		for _, jd := range []float64{2451545, 2461319.5} {
			got, err := SurfacePositionAtWithInstance(SkySun, BodyObserver{Body: body}, jd, inst)
			if err != nil {
				t.Fatal(err)
			}
			want, err := PlanetodeticToPlanetocentricLatitude(body, physical(jd, -1, inst).SubSolarLatitude)
			if err != nil {
				t.Fatal(err)
			}
			delta := math.Abs(got.BodyDec - want)
			maximum = math.Max(maximum, delta)
			if delta > limit {
				t.Fatalf("body=%d jd=%g BodyDec=%g subsolar=%g delta=%g", body, jd, got.BodyDec, want, delta)
			}
		}
	}
	t.Logf("subsolar declination maximum difference=%.9f degrees", maximum)
}

func TestSurfaceSearchSynthetic(t *testing.T) {
	const start = 2451545.0
	config := surfaceSearchConfig{rotationDays: 1, solarDays: 1, steps: 48}
	sinusoid := func(jd float64) (SurfacePosition, error) {
		angle := 2 * math.Pi * (jd - start)
		return SurfacePosition{Altitude: 30 * math.Sin(angle), HourAngle: angle * 180 / math.Pi}, nil
	}
	events, state, err := surfaceSearch(sinusoid, start+0.01, start+0.99, config)
	if err != nil || state != AltitudeCrosses || len(events) != 2 {
		t.Fatalf("events=%+v state=%d err=%v", events, state, err)
	}
	if events[0].Kind != SurfaceSet || events[1].Kind != SurfaceLowerCulmination {
		t.Fatalf("events=%+v", events)
	}
	for _, e := range events {
		if math.Abs(e.JDE-(start+0.5))*86400 > 0.1 {
			t.Fatal(e)
		}
	}
	for _, offset := range []float64{0, 0.0002} {
		sample := func(jd float64) (SurfacePosition, error) {
			x := (jd - start) - 0.503
			return SurfacePosition{Altitude: x*x + offset, HourAngle: 30}, nil
		}
		for _, jd := range []float64{start + 0.5, start + 25.0/48} {
			p, err := sample(jd)
			if err != nil || p.Altitude <= 0 {
				t.Fatalf("hidden-root premise: jd=%g altitude=%g err=%v", jd, p.Altitude, err)
			}
		}
		ev, st, e := surfaceSearch(sample, start+0.45, start+0.55, config)
		if e != nil {
			t.Fatal(e)
		}
		if math.Abs(offset) <= altitudeTangentDegrees {
			if st != AltitudeTangent {
				t.Fatalf("offset=%g events=%+v state=%d", offset, ev, st)
			}
		}
	}
	sample := func(jd float64) (SurfacePosition, error) {
		x := (jd - start) - 0.5104
		return SurfacePosition{Altitude: 100*x*x - 0.0009, HourAngle: 30}, nil
	}
	for _, jd := range []float64{start + 0.5, start + 25.0/48} {
		p, err := sample(jd)
		if err != nil || p.Altitude <= 0 {
			t.Fatalf("hidden-root premise: jd=%g altitude=%g err=%v", jd, p.Altitude, err)
		}
	}
	ev, st, e := surfaceSearch(sample, start+0.45, start+0.55, config)
	if e != nil || st != AltitudeCrosses || len(ev) != 2 || ev[0].Kind != SurfaceSet || ev[1].Kind != SurfaceRise {
		t.Fatalf("hidden pair=%+v state=%d err=%v", ev, st, e)
	}
	for i, event := range ev {
		expected := start + 0.5104 + float64(2*i-1)*math.Sqrt(0.0009/100)
		if math.Abs(event.JDE-expected)*86400 > 0.1 {
			t.Fatal(event, expected)
		}
	}
	fail := func(float64) (SurfacePosition, error) { return SurfacePosition{}, ErrInvalidObservationInput }
	if _, _, err := surfaceSearch(fail, start, start+1, config); err == nil {
		t.Fatal("missing sample error")
	}
}
func TestSurfaceSearchConsistencySampling(t *testing.T) {
	const timeLimit = 0.1
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	for _, body := range []SkyBody{SkyMercury, SkyVenus, SkyMars, SkyJupiter, SkyUranus} {
		observer := BodyObserver{Body: body, EastLongitude: 30, Latitude: 25}
		start, duration := 2461041.5, 3.0
		if body <= SkyVenus {
			duration = 400
		}
		if body == SkyMercury {
			observer.EastLongitude = 90.090184137733
			observer.Latitude = 0
			start = 2461080.9
			duration = 20
		}
		if body == SkyUranus {
			duration = 370
		}
		sample := func(jd float64) (SurfacePosition, error) {
			return SurfacePositionAtWithInstance(SkySun, observer, jd, inst)
		}
		config := surfaceConfigForBody(observer, SurfaceOptions{Limb: AltitudeLimbCenter})
		a, sa, err := surfaceSearch(sample, start, start+duration, config)
		if err != nil {
			t.Fatal(err)
		}
		config.envelope = false
		b, sb, err := surfaceSearch(sample, start, start+duration, config)
		if err != nil {
			t.Fatal(err)
		}
		config.steps *= 4
		c, sc, err := surfaceSearch(sample, start, start+duration, config)
		if err != nil {
			t.Fatal(err)
		}
		maximumSeconds := 0.0
		for _, other := range [][]SurfaceEvent{b, c} {
			if len(a) != len(other) {
				t.Fatalf("body=%d counts=%d/%d", body, len(a), len(other))
			}
			for i, e := range a {
				maximumSeconds = math.Max(maximumSeconds, math.Abs(e.JDE-other[i].JDE)*86400)
				if e.Kind != other[i].Kind || math.Abs(e.JDE-other[i].JDE)*86400 > timeLimit {
					t.Fatalf("body=%d event=%+v dense=%+v", body, e, other[i])
				}
			}
		}
		if sa != sb || sa != sc {
			t.Fatalf("body=%d state=%d/%d/%d", body, sa, sb, sc)
		}
		rises := 0
		for _, e := range a {
			if e.Kind <= SurfaceSet {
				rises++
			}
		}
		if body == SkyMercury && rises != 3 {
			t.Fatalf("reversing solar crossings=%d", rises)
		}
		if body == SkyUranus && (sa != AltitudeAlwaysAbove || rises != 0) {
			t.Fatalf("polar state=%d crossings=%d", sa, rises)
		}
		t.Logf("body=%d %g days events=%d crossings=%d; envelope and step/4 maximum=%.9f seconds", body, duration, len(a), rises, maximumSeconds)
	}
}

func TestSurfaceSearchSyntheticEnvelope(t *testing.T) {
	for _, alt := range []float64{-30, 30} {
		calls := 0
		sample := func(jd float64) (SurfacePosition, error) {
			calls++
			return SurfacePosition{Altitude: alt, BodyDec: alt, HourAngle: 360 * (jd - 2451545)}, nil
		}
		c := surfaceSearchConfig{rotationDays: 1, solarDays: 1, latitude: 90, steps: 48, envelope: true, onlyKind: true, kind: SurfaceRise}
		events, state, err := surfaceSearch(sample, 2451545.1, 2451548.1, c)
		want := AltitudeAlwaysAbove
		if alt < 0 {
			want = AltitudeAlwaysBelow
		}
		if err != nil || len(events) > 0 || state != want || calls > 20 {
			t.Fatalf("alt=%g state=%d events=%+v calls=%d err=%v", alt, state, events, calls, err)
		}
		c.onlyKind = false
		events, state, err = surfaceSearch(sample, 2451545.1, 2451548.1, c)
		if err != nil || len(events) != 6 || state != want {
			t.Fatalf("culminations: %d state=%d err=%v", len(events), state, err)
		}
	}
}

func TestSurfacePositionConsistencyCenters(t *testing.T) {
	const centerLimit = 0.02
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	centerMax := 0.0
	for _, g := range surfaceFixture(t) {
		target, observer := surfaceFixtureBodies(g)
		for _, sample := range g.Samples {
			frame, e := bodyObserverFrameAt(observer, sample[0])
			if e != nil {
				t.Fatal(e)
			}
			id, _ := surfaceBodyID(observer.Body)
			tid, _ := surfaceBodyID(target)
			ssb, _, _, e := ephem.ApparentFrom(inst.Engine(), ephem.SSB, id, [3]float64(frame.positionKM), [3]float64(surfaceSpinVelocity(frame)), tid, sample[0])
			if e != nil {
				t.Fatal(e)
			}
			helio, e := surfaceDirectionAt(target, observer, sample[0], inst)
			if e != nil {
				t.Fatal(e)
			}
			d := surfaceAngularDifference(Vector3(ssb), helio.direction)
			centerMax = math.Max(centerMax, d)
			if d > centerLimit {
				t.Fatalf("center difference=%g arcsec", d)
			}
		}
	}
	t.Logf("Sun/SSB maximum angular difference=%.9f arcsec", centerMax)
}

func TestSurfaceEventsExternal(t *testing.T) {
	const fastLimit, slowLimit = 10.0, 300.0
	for _, model := range []ephem.Model{ephem.Semianalytic, ephem.Legacy} {
		inst := ephem.NewModelInstance(model)
		fastMax, slowMax := 0.0, 0.0
		count := 0
		for _, g := range surfaceFixture(t) {
			if len(g.Crossings) == 0 {
				continue
			}
			target, observer := surfaceFixtureBodies(g)
			for _, crossing := range g.Crossings {
				kind := SurfaceRise
				if crossing.Kind == "set" {
					kind = SurfaceSet
				}
				options := SurfaceOptions{Limb: AltitudeLimbCenter}
				got, err := NextSurfaceEventWithInstance(target, observer, kind, crossing.JD-0.1, options, inst)
				if err != nil {
					t.Fatal(err)
				}
				delta := math.Abs(got.JDE-crossing.JD) * 86400
				limit := fastLimit
				if observer.Body == SkyMercury || observer.Body == SkyVenus {
					limit = slowLimit
					slowMax = math.Max(slowMax, delta)
				} else {
					fastMax = math.Max(fastMax, delta)
				}
				if delta > limit {
					t.Fatalf("%s model=%d event=%+v delta=%g s", g.Label, model, got, delta)
				}
				events, _, err := SurfaceEventsWithInstance(target, observer, crossing.JD-0.1, crossing.JD+0.1, options, inst)
				if err != nil {
					t.Fatal(err)
				}
				matched := 0
				for _, e := range events {
					if e.Kind == kind {
						matched++
						if e != got {
							t.Fatalf("Next/batch mismatch %s: %+v / %+v", g.Label, got, e)
						}
					}
				}
				if matched != 1 {
					t.Fatalf("%s matching event count=%d", g.Label, matched)
				}
				t.Logf("%s model=%d delta=%+.9f seconds", g.Label, model, (got.JDE-crossing.JD)*86400)
				count++
			}
		}
		if count != 17 {
			t.Fatal(count)
		}
		t.Logf("model=%d 17 center crossings max fast=%.9f / %.0f s, slow=%.9f / %.0f s", model, fastMax, fastLimit, slowMax, slowLimit)
	}
}

type finiteSurfaceEngine struct{ start, end float64 }

func (engine finiteSurfaceEngine) Coverage(ephem.Body) (float64, float64, bool) {
	return engine.start, engine.end, true
}

func (engine finiteSurfaceEngine) State(target, center ephem.Body, jd float64) ([3]float64, [3]float64, error) {
	if jd < engine.start || jd > engine.end {
		return [3]float64{}, [3]float64{}, ephem.ErrOutOfRange
	}
	if target == center {
		return [3]float64{}, [3]float64{}, nil
	}
	return [3]float64{225000000, 0, 0}, [3]float64{0, 2000000, 0}, nil
}

func TestSurfaceEventsFiniteEphemeris(t *testing.T) {
	const start, end = 2451545.02, 2451546.04
	observer := BodyObserver{Body: SkyMars}
	options := SurfaceOptions{}
	wide := ephem.NewInstance(finiteSurfaceEngine{2451540, 2451550})
	finite := ephem.NewInstance(finiteSurfaceEngine{2451545.01, 2451546.10})
	want, state, err := SurfaceEventsWithInstance(SkySun, observer, start, end, options, wide)
	if err != nil || len(want) != 4 {
		t.Fatal(want, state, err)
	}
	got, gotState, err := SurfaceEventsWithInstance(SkySun, observer, start, end, options, finite)
	if err != nil || gotState != state || !reflect.DeepEqual(got, want) {
		t.Fatalf("finite search: %+v %v %v, want %+v %v", got, gotState, err, want, state)
	}
	stack := ephem.NewInstance(ephem.NewStack(finiteSurfaceEngine{2451543.9, 2451544.25}, finiteSurfaceEngine{2451545.01, 2451550}))
	got, gotState, err = SurfaceEventsWithInstance(SkySun, observer, start, end, options, stack)
	if err != nil || gotState != state || !reflect.DeepEqual(got, want) {
		t.Fatalf("stack search: %+v %v %v, want %+v %v", got, gotState, err, want, state)
	}
	if _, _, err := SurfaceEventsWithInstance(SkySun, observer, 2451544.1, end, options, stack); !errors.Is(err, ephem.ErrOutOfRange) {
		t.Fatalf("gap inside query: %v", err)
	}
	for _, event := range want {
		last := ephem.NewInstance(finiteSurfaceEngine{2451545.01, event.JDE + 0.01})
		for _, seed := range []float64{event.JDE - 0.05, math.Nextafter(event.JDE, math.Inf(-1)), event.JDE} {
			got, err := NextSurfaceEventWithInstance(SkySun, observer, event.Kind, seed, options, last)
			if err != nil || got != event {
				t.Errorf("Next kind %d at seed %.12f near coverage end: %+v %v, want %+v", event.Kind, seed, got, err, event)
			}
		}
		after := math.Nextafter(event.JDE, math.Inf(1))
		if got, err := NextSurfaceEventWithInstance(SkySun, observer, event.Kind, after, options, last); !errors.Is(err, ephem.ErrOutOfRange) {
			t.Errorf("Next kind %d after event at %.12f: %+v %v", event.Kind, after, got, err)
		}
		got, _, err := SurfaceEventsWithInstance(SkySun, observer, event.JDE, event.JDE, options, finite)
		if err != nil || len(got) != 1 || got[0] != event {
			t.Errorf("batch at event %.12f: %+v %v, want %+v", event.JDE, got, err, event)
		}
	}
	if _, _, err := SurfaceEventsWithInstance(SkySun, observer, start, 2451546.11, options, finite); !errors.Is(err, ephem.ErrOutOfRange) {
		t.Fatalf("partly uncovered interval: %v", err)
	}
	star := StarCoordinate{RA: 75, Dec: 20, Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameICRS}
	wantStar, err := NextStarSurfaceEventWithInstance(star, observer, SurfaceRise, start, options, wide)
	if err != nil {
		t.Fatal(err)
	}
	last := ephem.NewInstance(finiteSurfaceEngine{2451545.01, wantStar.JDE + 0.01})
	gotStar, err := NextStarSurfaceEventWithInstance(star, observer, SurfaceRise, wantStar.JDE-0.05, options, last)
	if err != nil || gotStar.Kind != wantStar.Kind || math.Abs(gotStar.JDE-wantStar.JDE)*86400 > .1 {
		t.Fatalf("NextStar near coverage end: %+v %v, want %+v", gotStar, err, wantStar)
	}
}

func TestSurfaceEventsBoundedEndpoints(t *testing.T) {
	const start, end = 2451545.25, 2451545.75
	sample := func(jd float64) (SurfacePosition, error) {
		if jd < start || jd > end {
			return SurfacePosition{}, ephem.ErrOutOfRange
		}
		return SurfacePosition{Altitude: (jd - start) * (end - jd), HourAngle: 45}, nil
	}
	options := SurfaceOptions{Limb: AltitudeLimbCenter}
	got, _, err := surfaceEventsForSampler(sample, BodyObserver{Body: SkyMars}, start, end, options, nil)
	if err != nil || len(got) != 2 || got[0].Kind != SurfaceRise || got[0].JDE != start || got[1].Kind != SurfaceSet || got[1].JDE != end {
		t.Fatalf("endpoint events: %+v %v", got, err)
	}
	polar := func(jd float64) (SurfacePosition, error) {
		if jd < start || jd > end {
			return SurfacePosition{}, ephem.ErrOutOfRange
		}
		return SurfacePosition{Altitude: 70, BodyDec: 75, HourAngle: 360 * (jd - start)}, nil
	}
	got, _, err = surfaceEventsForSampler(polar, BodyObserver{Body: SkyMars, Latitude: 80}, start, end, options, nil)
	if err != nil || len(got) != 2 || got[0].Kind != SurfaceUpperCulmination || got[0].JDE != start || got[1].Kind != SurfaceLowerCulmination || got[1].JDE != end {
		t.Fatalf("endpoint culminations: %+v %v", got, err)
	}
}

func TestSurfaceEventsStackGapSynthetic(t *testing.T) {
	const start, end = 2451545.02, 2451546.04
	const gapStart, gapEnd = 2451545.330, 2451545.334
	observer := BodyObserver{Body: SkyMars}
	star := StarCoordinate{RA: 75, Dec: 20, Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameICRS}
	first := finiteSurfaceEngine{2451545.01, gapStart}
	last := finiteSurfaceEngine{gapEnd, 2451550}
	stack := ephem.NewInstance(ephem.NewStack(first, last))
	if events, _, err := SurfaceEventsWithInstance(SkySun, observer, start, end, SurfaceOptions{}, stack); !errors.Is(err, ephem.ErrOutOfRange) || len(events) != 0 {
		t.Errorf("solar batch across a sub-grid gap: %v, %v", events, err)
	}
	if events, _, err := StarSurfaceEventsWithInstance(star, observer, start, end, SurfaceOptions{}, stack); !errors.Is(err, ephem.ErrOutOfRange) || len(events) != 0 {
		t.Errorf("stellar batch across a sub-grid gap: %v, %v", events, err)
	}
	if _, err := NextSurfaceEventWithInstance(SkySun, observer, SurfaceRise, start, SurfaceOptions{}, stack); !errors.Is(err, ephem.ErrOutOfRange) {
		t.Errorf("next rise would cross the gap: %v", err)
	}
	if _, err := NextSurfaceEventWithInstance(SkySun, observer, SurfaceRise, (gapStart+gapEnd)/2, SurfaceOptions{}, stack); !errors.Is(err, ephem.ErrOutOfRange) {
		t.Errorf("next rise starting in the gap: %v", err)
	}
	resume := gapEnd + 0.001
	if rise, err := NextSurfaceEventWithInstance(SkySun, observer, SurfaceRise, resume, SurfaceOptions{}, stack); err != nil || rise.JDE <= resume || rise.Kind != SurfaceRise {
		t.Errorf("next rise after restarting beyond the gap: %+v, %v", rise, err)
	}
	want, err := NextSurfaceEventWithInstance(SkySun, observer, SurfaceSet, start, SurfaceOptions{}, ephem.NewInstance(first))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := NextSurfaceEventWithInstance(SkySun, observer, SurfaceSet, start, SurfaceOptions{}, stack); err != nil || got != want {
		t.Errorf("next set before the gap: %v, %v, want %v", got, err, want)
	}
	continuous := ephem.NewInstance(ephem.NewStack(first, last, finiteSurfaceEngine{gapStart, gapEnd}))
	if got, _, err := SurfaceEventsWithInstance(SkySun, observer, start, end, SurfaceOptions{}, continuous); err != nil || len(got) != 4 {
		t.Errorf("touching source boundaries: %v, %v", got, err)
	}
}

type surfaceGapEngine struct {
	finiteSurfaceEngine
	gapStart, gapEnd float64
}

func (engine surfaceGapEngine) State(target, center ephem.Body, jd float64) ([3]float64, [3]float64, error) {
	if center != ephem.Sun {
		return [3]float64{}, [3]float64{}, ephem.ErrUnsupportedBody
	}
	pos, vel, err := engine.finiteSurfaceEngine.State(target, center, jd)
	if target == ephem.Jupiter {
		if jd > engine.gapStart && jd < engine.gapEnd {
			return [3]float64{}, [3]float64{}, ephem.ErrOutOfRange
		}
		pos, vel = [3]float64{1025000000, 500000000, 0}, [3]float64{}
	}
	return pos, vel, err
}

func (engine surfaceGapEngine) AppendCoverageBoundaries(dst []float64, body ephem.Body) []float64 {
	if body == ephem.Jupiter {
		return append(dst, engine.start, engine.gapStart, engine.gapEnd, engine.end)
	}
	if body == ephem.Mars {
		return append(dst, engine.start, engine.end)
	}
	return dst
}

func TestSurfaceEventsRetardedGapSynthetic(t *testing.T) {
	observer := BodyObserver{Body: SkyMars}
	engine := surfaceGapEngine{finiteSurfaceEngine{2451540, 2451550}, 2451545.300, 2451545.301}
	const start, end = 2451545.33, 2451546.35
	for _, instance := range []*Instance{ephem.NewInstance(engine), ephem.NewInstance(ephem.NewStack(engine))} {
		for _, jd := range []float64{start, end} {
			if _, err := SurfacePositionAtWithInstance(SkyJupiter, observer, jd, instance); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := SurfacePositionAtWithInstance(SkyJupiter, observer, 2451545.336920, instance); !errors.Is(err, ephem.ErrOutOfRange) {
			t.Fatalf("expected missing retarded state, got %v", err)
		}
		if events, _, err := SurfaceEventsWithInstance(SkyJupiter, observer, start, end, SurfaceOptions{}, instance); len(events) != 0 || !errors.Is(err, ephem.ErrOutOfRange) {
			t.Errorf("batch crossed gap: %d events, %v", len(events), err)
		}
		if event, err := NextSurfaceEventWithInstance(SkyJupiter, observer, SurfaceUpperCulmination, start, SurfaceOptions{}, instance); !errors.Is(err, ephem.ErrOutOfRange) {
			t.Errorf("Next crossed gap: %+v, %v", event, err)
		}
		if _, _, err := SurfaceEventsWithInstance(SkyJupiter, observer, 2451545.325, 2451545.329, SurfaceOptions{}, instance); err != nil {
			t.Errorf("coverage needed only on each side of gap: %v", err)
		}
	}
}

func TestSurfaceEventsConsistencySeedsAndBatch(t *testing.T) {
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	for _, body := range []SkyBody{SkyVenus, SkyMars, SkyUranus} {
		observer := BodyObserver{Body: body, EastLongitude: 30}
		config := surfaceConfigForBody(observer, SurfaceOptions{})
		start := 2461041.5
		end := start + 2.1*config.solarDays
		events, _, err := SurfaceEventsWithInstance(SkySun, observer, start, end, SurfaceOptions{}, inst)
		if err != nil {
			t.Fatal(err)
		}
		assembled := []SurfaceEvent{}
		for kind := SurfaceRise; kind <= SurfaceLowerCulmination; kind++ {
			matching := []SurfaceEvent{}
			for _, e := range events {
				if e.Kind == kind {
					matching = append(matching, e)
				}
			}
			if len(matching) < 2 {
				t.Fatalf("body=%d kind=%d events=%+v", body, kind, events)
			}
			for i, seed := range []float64{matching[0].JDE - 1.0/1440, matching[0].JDE, matching[0].JDE + 1.0/1440} {
				got, err := NextSurfaceEventWithInstance(SkySun, observer, kind, seed, SurfaceOptions{}, inst)
				want := matching[0]
				if i == 2 {
					want = matching[1]
				}
				if err != nil || got != want {
					t.Fatalf("body=%d kind=%d seed=%d got=%+v expected=%+v err=%v", body, kind, i, got, want, err)
				}
			}
			seed := start
			for {
				e, err := NextSurfaceEventWithInstance(SkySun, observer, kind, seed, SurfaceOptions{}, inst)
				if err != nil {
					t.Fatal(err)
				}
				if e.JDE > end {
					break
				}
				assembled = append(assembled, e)
				seed = e.JDE + 0.1/86400
			}
		}
		sort.Slice(assembled, func(i, j int) bool {
			if assembled[i].JDE != assembled[j].JDE {
				return assembled[i].JDE < assembled[j].JDE
			}
			return assembled[i].Kind < assembled[j].Kind
		})
		if !reflect.DeepEqual(events, assembled) {
			t.Fatalf("body=%d batch=%+v assembled=%+v", body, events, assembled)
		}
		t.Logf("body=%d events=%d, all four kinds: three seed positions and Next assembly bitwise equal", body, len(events))
	}
}
func TestSurfaceEventsConsistencyInstancesAndStars(t *testing.T) {
	previous := GetEphemerisModel()
	defer SetEphemerisModel(previous)
	observer := BodyObserver{Body: SkyMars, EastLongitude: 42, Latitude: 25}
	star := StarCoordinate{RA: 75, Dec: 20, Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameICRS}
	for _, model := range []EphemerisModel{EphemerisLegacy, EphemerisSemianalytic} {
		SetEphemerisModel(model)
		inst := ephem.NewModelInstance(ephem.Model(model))
		a, sa, e := SurfaceEvents(SkySun, observer, 2461041.5, 2461044.5, SurfaceOptions{})
		b, sb, f := SurfaceEventsWithInstance(SkySun, observer, 2461041.5, 2461044.5, SurfaceOptions{}, inst)
		if e != nil || f != nil || sa != sb || !reflect.DeepEqual(a, b) {
			t.Fatalf("surface model=%d %v/%v", model, e, f)
		}
		c, sc, e := StarSurfaceEvents(star, observer, 2461041.5, 2461044.5, SurfaceOptions{Limb: AltitudeLimb(-99)})
		d, sd, f := StarSurfaceEventsWithInstance(star, observer, 2461041.5, 2461044.5, SurfaceOptions{}, inst)
		if e != nil || f != nil || sc != sd || !reflect.DeepEqual(c, d) {
			t.Fatalf("star model=%d %v/%v", model, e, f)
		}
		for kind := SurfaceRise; kind <= SurfaceLowerCulmination; kind++ {
			x, e := NextSurfaceEvent(SkySun, observer, kind, 2461041.5, SurfaceOptions{})
			y, f := NextSurfaceEventWithInstance(SkySun, observer, kind, 2461041.5, SurfaceOptions{}, inst)
			if e != nil || f != nil || x != y {
				t.Fatalf("next planet %v/%v", e, f)
			}
			x, e = NextStarSurfaceEvent(star, observer, kind, 2461041.5, SurfaceOptions{})
			y, f = NextStarSurfaceEventWithInstance(star, observer, kind, 2461041.5, SurfaceOptions{}, inst)
			if e != nil || f != nil || x != y {
				t.Fatalf("next star %v/%v", e, f)
			}
			matches := []SurfaceEvent{}
			for _, event := range d {
				if event.Kind == kind {
					matches = append(matches, event)
				}
			}
			if len(matches) < 2 {
				t.Fatal(matches)
			}
			for i, seed := range []float64{matches[0].JDE - 1.0/1440, matches[0].JDE, matches[0].JDE + 1.0/1440} {
				got, err := NextStarSurfaceEventWithInstance(star, observer, kind, seed, SurfaceOptions{}, inst)
				want := matches[0]
				if i == 2 {
					want = matches[1]
				}
				if err != nil || got != want {
					t.Fatalf("stellar kind=%d seed=%d got=%+v expected=%+v err=%v", kind, i, got, want, err)
				}
			}
		}
		for _, list := range [][]SurfaceEvent{a, c} {
			for _, e := range list {
				for _, v := range []float64{e.JDE, e.Position.Altitude, e.Position.Azimuth, e.Position.HourAngle, e.Position.BodyRA, e.Position.BodyDec, e.Position.DistanceAU, e.Position.AngularRadius} {
					if !isFiniteFloat(v) {
						t.Fatal(e)
					}
				}
			}
		}
	}
}
func TestSurfaceEventsSyntheticLimits(t *testing.T) {
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	observer := BodyObserver{Body: SkyMars, Latitude: 80}
	orientation, err := BodyOrientationAt(SkyMars, 2461041.5)
	if err != nil {
		t.Fatal(err)
	}
	star := StarCoordinate{RA: orientation.PoleRA, Dec: orientation.PoleDec, Epoch: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Frame: CoordinateFrameICRS}
	events, state, err := StarSurfaceEventsWithInstance(star, observer, 2461041.5, 2461044.5, SurfaceOptions{}, inst)
	if err != nil || state != AltitudeAlwaysAbove {
		t.Fatal(events, state, err)
	}
	event, err := NextStarSurfaceEventWithInstance(star, observer, SurfaceSet, 2461041.5, SurfaceOptions{}, inst)
	if !errors.Is(err, ErrSurfaceEventNotFound) || event != (SurfaceEvent{}) {
		t.Fatal(event, err)
	}
	for _, test := range []struct {
		start, end float64
		options    SurfaceOptions
	}{
		{2461041.5, 2461041.5 + 3652.501, SurfaceOptions{}}, {2461042.5, 2461041.5, SurfaceOptions{}}, {math.NaN(), 2461041.5, SurfaceOptions{}}, {2461041.5, 2461042.5, SurfaceOptions{TargetAltitude: 91}}, {2461041.5, 2461042.5, SurfaceOptions{Limb: AltitudeLimb(99)}},
	} {
		events, _, err := SurfaceEventsWithInstance(SkySun, observer, test.start, test.end, test.options, inst)
		if !errors.Is(err, ErrInvalidObservationInput) || events != nil {
			t.Fatal(events, err)
		}
	}
	event, err = NextSurfaceEventWithInstance(SkySun, observer, SurfaceEventKind(99), 2461041.5, SurfaceOptions{}, inst)
	if !errors.Is(err, ErrInvalidObservationInput) || event != (SurfaceEvent{}) {
		t.Fatal(event, err)
	}
	_, _, err = SurfaceEventsWithInstance(SkyMoon, observer, 2461041.5, 2461042.5, SurfaceOptions{}, inst)
	if err == nil {
		t.Fatal("unsupported target")
	}
}

func TestSurfaceSearchSyntheticShallowMinimum(t *testing.T) {
	const start = 2451545.0
	const center = 0.5104
	const depth = 0.0002
	const timeLimit = 0.1
	sample := func(jd float64) (SurfacePosition, error) {
		x := jd - start - center
		return SurfacePosition{Altitude: math.Min(80, 100*x*x-depth), HourAngle: 30}, nil
	}
	for _, jd := range []float64{start + 0.5, start + 25.0/48} {
		p, err := sample(jd)
		if err != nil || p.Altitude <= 0 {
			t.Fatalf("hidden-pair premise failed: jd=%g altitude=%g err=%v", jd, p.Altitude, err)
		}
	}
	config := surfaceSearchConfig{rotationDays: 1, solarDays: 1, steps: 48}
	coarse, coarseState, err := surfaceSearch(sample, start+0.45, start+0.55, config)
	if err != nil {
		t.Fatal(err)
	}
	config.steps *= 4
	dense, denseState, err := surfaceSearch(sample, start+0.45, start+0.55, config)
	if err != nil {
		t.Fatal(err)
	}
	for i, event := range dense {
		if i > 1 {
			t.Fatal(dense)
		}
		expected := start + center + float64(2*i-1)*math.Sqrt(depth/100)
		if math.Abs(event.JDE-expected)*86400 > timeLimit {
			t.Fatalf("dense root=%g expected=%g", event.JDE, expected)
		}
	}
	t.Logf("analytic minimum=-%.7f deg, two roots at center +/- %.9f days; coarse=%+v state=%d; dense=%+v state=%d", depth, math.Sqrt(depth/100), coarse, coarseState, dense, denseState)
	if len(coarse) != 2 || len(dense) != 2 || coarseState != AltitudeCrosses || denseState != AltitudeCrosses {
		t.Fatalf("two analytic crossings: coarse=%d/state=%d dense=%d/state=%d", len(coarse), coarseState, len(dense), denseState)
	}
}

func TestSurfaceSearchSyntheticExtremaDepths(t *testing.T) {
	const start = 2451545.0
	const center = 0.5104
	const curvature = 100.0
	const timeLimit = 0.1
	maximumSeconds := 0.0
	for _, steps := range []int{48, 96} {
		for _, shape := range []float64{1, -1} {
			for _, depth := range []float64{-3, -1, -0.72, -0.3, 0, 0.3, 0.72, 3} {
				sample := func(jd float64) (SurfacePosition, error) {
					x := jd - start - center
					return SurfacePosition{Altitude: shape * math.Min(80, curvature*x*x+depth/3600), HourAngle: 30}, nil
				}
				config := surfaceSearchConfig{rotationDays: 1, solarDays: 1, steps: steps}
				if steps == 96 {
					config.rotationDays = 4
				}
				coarse, coarseState, err := surfaceSearch(sample, start+0.45, start+0.55, config)
				if err != nil {
					t.Fatal(err)
				}
				config.steps *= 4
				dense, denseState, err := surfaceSearch(sample, start+0.45, start+0.55, config)
				if err != nil {
					t.Fatal(err)
				}
				expectedCount := 0
				expectedState := AltitudeAlwaysAbove
				if shape < 0 {
					expectedState = AltitudeAlwaysBelow
				}
				if depth < 0 {
					expectedCount = 2
					expectedState = AltitudeCrosses
				} else if depth < 1 {
					expectedState = AltitudeTangent
				}
				if len(coarse) != expectedCount || len(dense) != expectedCount || coarseState != expectedState || denseState != expectedState {
					t.Errorf("steps=%d shape=%g depth=%g arcsec: coarse=%d/state=%d dense=%d/state=%d expected=%d/state=%d", steps, shape, depth, len(coarse), coarseState, len(dense), denseState, expectedCount, expectedState)
					continue
				}
				for i, event := range coarse {
					expectedKind := SurfaceSet
					if i == 1 {
						expectedKind = SurfaceRise
					}
					if shape < 0 {
						expectedKind = SurfaceRise
						if i == 1 {
							expectedKind = SurfaceSet
						}
					}
					expectedJD := start + center + float64(2*i-1)*math.Sqrt(-depth/3600/curvature)
					for _, got := range []SurfaceEvent{event, dense[i]} {
						delta := math.Abs(got.JDE-expectedJD) * 86400
						maximumSeconds = math.Max(maximumSeconds, delta)
						if got.Kind != expectedKind || delta > timeLimit {
							t.Errorf("steps=%d shape=%g depth=%g root=%+v expected kind=%d JD=%.12f delta=%g s", steps, shape, depth, got, expectedKind, expectedJD, delta)
						}
					}
					delta := math.Abs(event.JDE-dense[i].JDE) * 86400
					maximumSeconds = math.Max(maximumSeconds, delta)
					if event.Kind != dense[i].Kind || delta > timeLimit {
						t.Errorf("steps=%d shape=%g depth=%g coarse=%+v dense=%+v", steps, shape, depth, event, dense[i])
					}
				}
			}
		}
	}
	t.Logf("32 shape/depth/mode cases; maximum analytic/grid time difference %.9f seconds", maximumSeconds)
}

func TestSurfaceSearchSyntheticTangentStrictBoundary(t *testing.T) {
	const start = 2451545.0
	for _, steps := range []int{48, 96, 192, 384} {
		for _, shape := range []float64{1, -1} {
			sample := func(jd float64) (SurfacePosition, error) {
				x := math.Max(0, math.Abs(jd-start-0.5104)-0.0001)
				return SurfacePosition{Altitude: shape * math.Min(80, 1.0/3600+100*x*x), HourAngle: 30}, nil
			}
			config := surfaceSearchConfig{rotationDays: 1, solarDays: 1, steps: steps}
			events, state, err := surfaceSearch(sample, start+0.45, start+0.55, config)
			expected := AltitudeAlwaysAbove
			if shape < 0 {
				expected = AltitudeAlwaysBelow
			}
			if err != nil || len(events) != 0 || state != expected {
				t.Errorf("steps=%d shape=%g exactly 1 arcsec: events=%+v state=%d expected=%d err=%v", steps, shape, events, state, expected, err)
			}
		}
	}
}

func TestSurfaceSearchConsistencySparseCulminations(t *testing.T) {
	const timeLimit = 0.1
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	for _, c := range []struct {
		name           string
		body, target   SkyBody
		latitude, days float64
	}{
		{"uranus-sun-25", SkyUranus, SkySun, 25, 3652.5},
		{"uranus-sun-80", SkyUranus, SkySun, 80, 3652.5},
		{"saturn-sun-70", SkySaturn, SkySun, 70, 3652.5},
		{"jupiter-earth-south", SkyJupiter, SkyEarth, -60, 365},
		{"mars-sun-85", SkyMars, SkySun, 85, 800},
		{"mars-jupiter-60", SkyMars, SkyJupiter, 60, 800},
	} {
		t.Run(c.name, func(t *testing.T) {
			observer := BodyObserver{Body: c.body, EastLongitude: 30, Latitude: c.latitude}
			calls := 0
			sample := func(jd float64) (SurfacePosition, error) {
				calls++
				return SurfacePositionAtWithInstance(c.target, observer, jd, inst)
			}
			config := surfaceConfigForBody(observer, SurfaceOptions{})
			start, end := 2461041.5, 2461041.5+c.days
			sparse, state, err := surfaceSearch(sample, start, end, config)
			if err != nil {
				t.Fatal(err)
			}
			sparseCalls := calls
			calls = 0
			config.disableSparse = true
			dense, denseState, err := surfaceSearch(sample, start, end, config)
			if err != nil {
				t.Fatal(err)
			}
			if len(sparse) != len(dense) || state != denseState {
				t.Fatalf("event set/state differs: sparse=%d/%d dense=%d/%d", len(sparse), state, len(dense), denseState)
			}
			maximum := 0.0
			for i, e := range sparse {
				delta := math.Abs(e.JDE-dense[i].JDE) * 86400
				maximum = math.Max(maximum, delta)
				if e.Kind != dense[i].Kind || !isFiniteFloat(delta) || delta > timeLimit {
					t.Fatalf("event %d: sparse=%+v dense=%+v delta=%g s", i, e, dense[i], delta)
				}
				position, err := sample(e.JDE)
				if err != nil || e.Position != position {
					t.Fatalf("event position mismatch: %+v sample=%+v err=%v", e, position, err)
				}
			}
			t.Logf("events=%d state=%d maximum sparse/dense=%.9f s / %.1f s; calls=%d/%d", len(sparse), state, maximum, timeLimit, sparseCalls, calls-len(sparse))
		})
	}
}

func TestSurfaceSearchSyntheticSparseGuardFallback(t *testing.T) {
	const epoch = 2451545.0
	for _, mode := range []string{"ra-drift", "hour-angle-jump"} {
		t.Run(mode, func(t *testing.T) {
			seen := make(map[uint64]bool)
			sample := func(jd float64) (SurfacePosition, error) {
				seen[math.Float64bits(jd)] = true
				x := jd - epoch
				ra := 0.0
				rate := 360.0
				if mode == "ra-drift" {
					ra = 20 * x
				} else {
					rate = 720
				}
				return SurfacePosition{Altitude: 30, BodyDec: 30, BodyRA: ra, HourAngle: math.Remainder(17+rate*x, 360)}, nil
			}
			config := surfaceSearchConfig{rotationDays: 1, solarDays: 1, latitude: 90, steps: 48, envelope: true}
			sparse, state, err := surfaceSearch(sample, epoch+0.1, epoch+0.9, config)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i <= 48; i++ {
				jd := epoch + float64(i)/48
				if !seen[math.Float64bits(jd)] {
					t.Fatalf("48-step fallback missing grid point %d: %.12f", i, jd)
				}
			}
			config.disableSparse = true
			dense, denseState, err := surfaceSearch(sample, epoch+0.1, epoch+0.9, config)
			if err != nil {
				t.Fatal(err)
			}
			if state != denseState || !reflect.DeepEqual(sparse, dense) {
				t.Fatalf("guard fallback differs: sparse=%+v/%d dense=%+v/%d", sparse, state, dense, denseState)
			}
			if len(sparse) < 1 {
				t.Fatal("fixture has no culminations")
			}
		})
	}
}

func TestSurfaceSearchSyntheticSparseAnalytic(t *testing.T) {
	const epoch = 2451545.0
	const timeLimit = 0.1
	for _, offset := range []float64{0, 17, 180, 359} {
		calls := 0
		sample := func(jd float64) (SurfacePosition, error) {
			calls++
			x := jd - epoch
			return SurfacePosition{Altitude: 30, BodyDec: 30, BodyRA: 359.5 + x, HourAngle: math.Remainder(offset+360*x, 360)}, nil
		}
		config := surfaceSearchConfig{rotationDays: 1, solarDays: 1, latitude: 90, steps: 48, envelope: true}
		events, state, err := surfaceSearch(sample, epoch+0.01, epoch+1.99, config)
		if err != nil || state != AltitudeAlwaysAbove {
			t.Fatal(state, err)
		}
		expected := []SurfaceEvent{}
		for turn := -1; turn <= 6; turn++ {
			jd := epoch + (float64(turn)*180-offset)/360
			if jd < epoch+0.01 || jd > epoch+1.99 {
				continue
			}
			kind := SurfaceUpperCulmination
			if turn%2 != 0 {
				kind = SurfaceLowerCulmination
			}
			expected = append(expected, SurfaceEvent{Kind: kind, JDE: jd})
		}
		if len(events) != len(expected) {
			t.Fatalf("offset=%g events=%+v expected=%+v", offset, events, expected)
		}
		for i, e := range events {
			if e.Kind != expected[i].Kind || math.Abs(e.JDE-expected[i].JDE)*86400 > timeLimit {
				t.Fatalf("offset=%g event=%+v expected=%+v", offset, e, expected[i])
			}
		}
		sparseCalls := calls
		calls = 0
		config.disableSparse = true
		dense, denseState, err := surfaceSearch(sample, epoch+0.01, epoch+1.99, config)
		if err != nil || denseState != state || len(dense) != len(events) {
			t.Fatalf("offset=%g dense=%+v state=%d err=%v", offset, dense, denseState, err)
		}
		for i, event := range events {
			if event.Kind != dense[i].Kind || math.Abs(event.JDE-dense[i].JDE)*86400 > timeLimit {
				t.Fatalf("offset=%g sparse=%+v dense=%+v", offset, event, dense[i])
			}
		}
		if sparseCalls >= calls {
			t.Fatalf("sparse sampling not active: sparse=%d dense=%d", sparseCalls, calls)
		}
		t.Logf("offset=%g sparse/dense calls=%d/%d", offset, sparseCalls, calls)
	}
}

func TestSurfaceEventsReviewRegressionsSynthetic(t *testing.T) {
	obs := BodyObserver{Body: SkyJupiter}
	if _, _, err := SurfaceEvents(SkySun, obs, 1e16, 1e16, SurfaceOptions{}); !errors.Is(err, ErrInvalidObservationInput) {
		t.Fatalf("extreme JD batch: %v", err)
	}
	if _, err := NextSurfaceEvent(SkySun, obs, SurfaceRise, 4e15, SurfaceOptions{}); !errors.Is(err, ErrInvalidObservationInput) {
		t.Fatalf("extreme JD next: %v", err)
	}
	const start = 2451545.0
	config := surfaceSearchConfig{rotationDays: 1, solarDays: 1, latitude: 0, steps: 48, envelope: false}
	gridTangent := func(jd float64) (SurfacePosition, error) {
		d := jd - (start + 0.5)
		return SurfacePosition{Altitude: d * d, HourAngle: 45}, nil
	}
	events, state, err := surfaceSearch(gridTangent, start+0.1, start+0.9, config)
	if err != nil || len(events) != 0 || state != AltitudeTangent {
		t.Fatalf("grid-aligned tangency: events=%+v state=%d err=%v", events, state, err)
	}
	for _, item := range []struct {
		slope float64
		kind  SurfaceEventKind
	}{
		{40, SurfaceRise},
		{-40, SurfaceSet},
	} {
		gridCrossing := func(jd float64) (SurfacePosition, error) {
			return SurfacePosition{Altitude: item.slope * (jd - (start + 0.5)), HourAngle: 45}, nil
		}
		events, state, err = surfaceSearch(gridCrossing, start+0.1, start+0.9, config)
		if err != nil || len(events) != 1 || events[0].Kind != item.kind || state != AltitudeCrosses || math.Abs(events[0].JDE-(start+0.5))*86400 > 0.1 {
			t.Fatalf("grid-aligned crossing slope=%g: events=%+v state=%d err=%v", item.slope, events, state, err)
		}
	}
	midSymmetric := func(jd float64) (SurfacePosition, error) {
		d := jd - (start + 24.5/48)
		return SurfacePosition{Altitude: 100*d*d - 0.0009, HourAngle: 45}, nil
	}
	events, state, err = surfaceSearch(midSymmetric, start+0.45, start+0.55, config)
	if err != nil || len(events) != 2 || events[0].Kind != SurfaceSet || events[1].Kind != SurfaceRise || state != AltitudeCrosses {
		t.Fatalf("midpoint-symmetric hidden crossing: events=%+v state=%d err=%v", events, state, err)
	}
	constZero := func(jd float64) (SurfacePosition, error) {
		return SurfacePosition{Altitude: 0, HourAngle: 45}, nil
	}
	events, state, err = surfaceSearch(constZero, start+0.1, start+0.9, config)
	if err != nil || len(events) != 0 || state != AltitudeTangent {
		t.Fatalf("constant zero residual: events=%+v state=%d err=%v", events, state, err)
	}
	// 窗口外存在窄于步长的空档且包含窗口外根，不应影响已完全覆盖的查询窗口。
	outOfWindowGap := func(jd float64) (SurfacePosition, error) {
		if jd > start-0.195 && jd < start-0.185 {
			return SurfacePosition{}, ephem.ErrOutOfRange
		}
		return SurfacePosition{Altitude: jd - (start - 0.19), HourAngle: 45}, nil
	}
	events, state, err = surfaceSearch(outOfWindowGap, start+0.02, start+0.06, config)
	if err != nil || len(events) != 0 || state != AltitudeAlwaysAbove {
		t.Fatalf("out-of-window sub-step gap: events=%+v state=%d err=%v", events, state, err)
	}
}

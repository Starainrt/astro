package ephem

import (
	"math"
	"testing"
)

type apparentFromTestEngine struct {
	positions, velocities map[Body][3]float64
	fail                  Body
}

func (e apparentFromTestEngine) State(target, center Body, jdTT float64) (p, v [3]float64, err error) {
	if center != SSB || target == e.fail {
		return p, v, ErrUnsupportedBody
	}
	p, ok := e.positions[target]
	if !ok {
		return p, v, ErrUnsupportedBody
	}
	v = e.velocities[target]
	for i := range p {
		p[i] += v[i] * (jdTT - 2451545)
	}
	return p, v, nil
}

func (e apparentFromTestEngine) Coverage(Body) (float64, float64, bool) {
	return 2400000, 2500000, true
}

func TestApparentFromConsistencyGeocentric(t *testing.T) {
	engine := apparentFromTestEngine{positions: map[Body][3]float64{Sun: {2e5, -1e5, 3e4}, Earth: {1e8, 2e7, 1e6}}, velocities: map[Body][3]float64{Earth: {-4e5, 2e6, 8e4}}, fail: Body(-1)}
	targets := []Body{Sun, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune}
	for _, target := range targets[1:] {
		engine.positions[target] = [3]float64{float64(target) * 8e7, -4e7, float64(target) * 3e6}
		engine.velocities[target] = [3]float64{3e5, float64(target) * 1e5, -3e4}
	}
	for _, target := range targets {
		for _, jdTT := range []float64{2451545, 2461319.5} {
			got, distance, tau, err := ApparentFrom(engine, SSB, Earth, [3]float64{}, [3]float64{}, target, jdTT)
			want, wantDistance, wantTau, wantErr := genericApparent(engine, target, jdTT)
			if err != nil || wantErr != nil || got != want || math.Float64bits(distance) != math.Float64bits(wantDistance) || math.Float64bits(tau) != math.Float64bits(wantTau) {
				t.Fatalf("target=%d jdTT=%g got=%v/%g/%g want=%v/%g/%g err=%v/%v", target, jdTT, got, distance, tau, want, wantDistance, wantTau, err, wantErr)
			}
		}
	}
}

func TestApparentFromSyntheticLightTimeAndVelocity(t *testing.T) {
	const c = 299792.458 * 86400
	engine := apparentFromTestEngine{positions: map[Body][3]float64{Mars: {1e8, 0, 0}, Sun: {2e8, 0, 0}}, velocities: map[Body][3]float64{Sun: {1e5, 0, 0}}, fail: Body(-1)}
	dir, distance, tau, err := ApparentFrom(engine, SSB, Mars, [3]float64{1000, 0, 0}, [3]float64{}, Sun, 2451545)
	wantDistance := 1e8 - 1000.0
	wantTau := wantDistance / (c + 1e5)
	if err != nil || math.Abs(tau-wantTau) > 1e-14 || distance != wantDistance || dir != ([3]float64{1, 0, 0}) {
		t.Fatalf("dir=%v distance=%.17g tau=%.17g want=%.17g err=%v", dir, distance, tau, wantTau, err)
	}
	for _, speed := range []float64{-1e6, 1e6} {
		dir, distance, _, err = ApparentFrom(engine, SSB, Mars, [3]float64{}, [3]float64{0, speed, 0}, Sun, 2451545)
		want := [3]float64{math.Sqrt(1 - speed*speed/(c*c)), speed / c, 0}
		for i := range dir {
			if err != nil || math.Abs(dir[i]-want[i]) > 1e-15 || distance != 1e8 {
				t.Fatalf("speed=%g dir=%v want=%v err=%v", speed, dir, want, err)
			}
		}
	}
	t.Logf("light-time residual = %.12g days", math.Abs(tau-wantTau))
}

func TestApparentFromConsistencyDeflection(t *testing.T) {
	engine := apparentFromTestEngine{positions: map[Body][3]float64{Sun: {}, Mars: {1e8, 0, 0}, Earth: {-2e8, 2e6, 0}, Moon: {-2e8, 2e6, 0}}, fail: Body(-1)}
	earth, _, _, err := ApparentFrom(engine, SSB, Mars, [3]float64{}, [3]float64{}, Earth, 2451545)
	moon, _, _, moonErr := ApparentFrom(engine, SSB, Mars, [3]float64{}, [3]float64{}, Moon, 2451545)
	geometric := genericScale([3]float64{-3e8, 2e6, 0}, 1/genericLen([3]float64{-3e8, 2e6, 0}))
	want := genericAberrate(genericDeflect(geometric, engine.positions[Earth], engine.positions[Mars]), [3]float64{})
	if err != nil || moonErr != nil || earth != want || earth == moon || moon != genericAberrate(geometric, [3]float64{}) {
		t.Fatalf("earth=%v moon=%v want=%v err=%v/%v", earth, moon, want, err, moonErr)
	}
}

func TestApparentFromSyntheticInvalid(t *testing.T) {
	engine := apparentFromTestEngine{positions: map[Body][3]float64{Mars: {1e8, 0, 0}, Sun: {}}, fail: Body(-1)}
	var typedNil *nilCheckEngine
	for _, sample := range []struct {
		engine           Engine
		observer, target Body
		offset, velocity [3]float64
		jd               float64
	}{
		{nil, Mars, Sun, [3]float64{}, [3]float64{}, 2451545}, {typedNil, Mars, Sun, [3]float64{}, [3]float64{}, 2451545},
		{engine, Mars, Mars, [3]float64{}, [3]float64{}, 2451545}, {engine, Earth, Earth, [3]float64{}, [3]float64{}, 2451545},
		{engine, Mars, Sun, [3]float64{math.NaN()}, [3]float64{}, 2451545}, {engine, Mars, Sun, [3]float64{}, [3]float64{math.Inf(1)}, 2451545},
		{engine, Mars, Sun, [3]float64{}, [3]float64{genericCKMPerDay}, 2451545}, {engine, Mars, Sun, [3]float64{}, [3]float64{}, math.NaN()},
		{engine, Mars, Jupiter, [3]float64{}, [3]float64{}, 2451545}, {engine, Jupiter, Sun, [3]float64{}, [3]float64{}, 2451545},
		{engine, Mars, Sun, [3]float64{-1e8}, [3]float64{}, 2451545},
	} {
		dir, distance, tau, err := ApparentFrom(sample.engine, SSB, sample.observer, sample.offset, sample.velocity, sample.target, sample.jd)
		if err == nil || dir != ([3]float64{}) || distance != 0 || tau != 0 {
			t.Fatalf("sample=%+v dir=%v distance=%g tau=%g err=%v", sample, dir, distance, tau, err)
		}
	}
}

func TestApparentFromSyntheticCenter(t *testing.T) {
	engine := apparentFromTestEngine{positions: map[Body][3]float64{Mars: {1e8, 0, 0}, Sun: {}}, fail: Body(-1)}
	dir, dist, tau, err := ApparentFrom(engine, Earth, Mars, [3]float64{}, [3]float64{}, Sun, 2451545)
	if err == nil || dir != ([3]float64{}) || dist != 0 || tau != 0 {
		t.Fatalf("invalid center: %v %g %g %v", dir, dist, tau, err)
	}
}

type apparentFromSolarTestEngine struct{ calls int }

func (e *apparentFromSolarTestEngine) State(target, center Body, jd float64) (p, v [3]float64, err error) {
	e.calls++
	if center != Sun {
		return p, v, ErrUnsupportedBody
	}
	switch target {
	case Sun:
	case Mars:
		p = [3]float64{1e8, 0, 0}
		v = [3]float64{0, 1e6, 0}
	default:
		return p, v, ErrUnsupportedBody
	}
	return p, v, nil
}
func (*apparentFromSolarTestEngine) Coverage(Body) (float64, float64, bool) {
	return 2400000, 2500000, true
}
func TestApparentFromSyntheticSolarCenter(t *testing.T) {
	engine := &apparentFromSolarTestEngine{}
	got, distance, tau, err := ApparentFrom(engine, Sun, Mars, [3]float64{1000}, [3]float64{0, 100, 0}, Sun, 2451545)
	beta := (1e6 + 100) / (299792.458 * 86400)
	want := [3]float64{-math.Sqrt(1 - beta*beta), beta, 0}
	if err != nil || distance != 1e8+1000 || tau != distance/(299792.458*86400) || engine.calls != 5 {
		t.Fatalf("got=%v/%g/%g calls=%d err=%v", got, distance, tau, engine.calls, err)
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > 1e-15 {
			t.Fatalf("got=%v expected=%v", got, want)
		}
	}
}

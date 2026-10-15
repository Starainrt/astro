package basic

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet"
)

type p3StateEngine struct {
	position [3]float64
	err      error
}

func (engine p3StateEngine) State(target, center ephem.Body, jde float64) ([3]float64, [3]float64, error) {
	return engine.position, [3]float64{}, engine.err
}

func (engine p3StateEngine) Coverage(target ephem.Body) (float64, float64, bool) {
	return 2400000, 2500000, true
}

type p3ApparentEngine struct {
	p3StateEngine
	direction [3]float64
	distance  float64
	tau       float64
	err       error
}

func (engine p3ApparentEngine) Apparent(target ephem.Body, jde float64) ([3]float64, float64, float64, error) {
	return engine.direction, engine.distance, engine.tau, engine.err
}

func TestP3BuiltInGeometryPreservesTruncation(t *testing.T) {
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		inst := ephem.NewModelInstance(model)
		for _, jde := range []float64{2451545, 2461319.5} {
			for _, terms := range []int{-1, 0, 16} {
				for _, sample := range []struct {
					name string
					got  float64
					want float64
				}{
					{"SunLo", HSunTrueLoNWithInstance(jde, terms, inst), planet.WherePlanetModelN(model, 0, 0, jde, terms)},
					{"SunBo", HSunTrueBoNWithInstance(jde, terms, inst), planet.WherePlanetModelN(model, 0, 1, jde, terms)},
					{"SunDistance", EarthAwayNWithInstance(jde, terms, inst), planet.WherePlanetModelN(model, 0, 2, jde, terms)},
					{"MoonLo", MoonCalcNewNWithInstance(0, jde, terms, inst), planet.WhereMoonModelN(model, 0, jde, terms)},
					{"MoonBo", MoonCalcNewNWithInstance(1, jde, terms, inst), planet.WhereMoonModelN(model, 1, jde, terms)},
					{"MoonDistance", HMoonAwayNWithInstance(jde, terms, inst), planet.WhereMoonModelN(model, 2, jde, terms)},
				} {
					if sample.got != sample.want {
						t.Errorf("%s model=%d jde=%v n=%d: got %.17g, want %.17g", sample.name, model, jde, terms, sample.got, sample.want)
					}
				}
			}
		}
	}
}

func TestP3CustomGeometryUsesMeanEcliptic(t *testing.T) {
	const jde = 2461319.5
	direction := [3]float64{1, 0, 0}
	inst := ephem.NewInstance(p3StateEngine{position: [3]float64{astronomicalUnitKM, 0, 0}})
	trueLo, wantBo := icrfToTrueEcliptic(direction, jde)
	wantLo := trueLo - Nutation2000Bi(jde)
	for _, got := range []float64{HSunTrueLoWithInstance(jde, inst), HMoonTrueLoWithInstance(jde, inst), MercuryLWithInstance(jde, inst)} {
		if math.Abs(angleDifference(got, wantLo)) > 1e-12 {
			t.Errorf("geometric longitude includes nutation: got %.17g, want %.17g", got, wantLo)
		}
	}
	if got := HSunTrueBoWithInstance(jde, inst); math.Abs(got-wantBo) > 1e-12 {
		t.Errorf("latitude: got %.17g, want %.17g", got, wantBo)
	}
	if got := EarthAwayNWithInstance(jde, 16, inst); got != 1 {
		t.Errorf("distance: got %v, want 1 AU", got)
	}
}

func TestP3CustomGeometryDoesNotFallback(t *testing.T) {
	for _, engine := range []ephem.Engine{
		nil,
		p3StateEngine{err: ephem.ErrOutOfRange},
		p3StateEngine{err: ephem.ErrUnsupportedBody},
		p3StateEngine{},
		p3StateEngine{position: [3]float64{math.NaN(), 0, 0}},
	} {
		inst := ephem.NewInstance(engine)
		for _, got := range []float64{
			HSunTrueLoWithInstance(2451545, inst), HSunTrueBoWithInstance(2451545, inst), EarthAwayNWithInstance(2451545, -1, inst),
			HMoonTrueLoWithInstance(2451545, inst), HMoonTrueBoWithInstance(2451545, inst), HMoonAwayWithInstance(2451545, inst),
		} {
			if !math.IsNaN(got) {
				t.Errorf("failed custom engine %T fell back to a built-in model: %v", engine, got)
			}
		}
	}
}

func TestP3CustomApparentFailureDoesNotFallback(t *testing.T) {
	state := p3StateEngine{position: [3]float64{astronomicalUnitKM, 0, 0}}
	for _, engine := range []p3ApparentEngine{
		{p3StateEngine: state, err: ephem.ErrOutOfRange},
		{p3StateEngine: state, direction: [3]float64{1, 0, 0}, distance: -1},
		{p3StateEngine: state, direction: [3]float64{math.NaN(), 0, 0}, distance: 1},
		{p3StateEngine: state, distance: 1},
		{p3StateEngine: state, direction: [3]float64{1, 0, 0}, distance: 1, tau: math.NaN()},
	} {
		inst := ephem.NewInstance(engine)
		for _, got := range []float64{HSunApparentLoWithInstance(2451545, inst), HMoonApparentLoWithInstance(2451545, inst), MercuryApparentLoWithInstance(2451545, inst)} {
			if !math.IsNaN(got) {
				t.Errorf("failed apparent engine fell back to geometry: %v", got)
			}
		}
	}
	inst := ephem.NewInstance(state)
	if got := HSunApparentLoWithInstance(2451545, inst); !isFiniteFloat(got) {
		t.Errorf("state-only engine must retain geometric fallback: %v", got)
	}
}

package orbit

import (
	"errors"
	"math"
	"testing"
)

// movingPerturbationEngine 提供随时间变化的太阳与行星位置 / supplies time-varying Sun and planet positions.
type movingPerturbationEngine struct{}

func (movingPerturbationEngine) State(target, center Body, jdTDB float64) (pos, vel [3]float64, err error) {
	if target == center {
		return pos, vel, nil
	}
	if center != 0 {
		return pos, vel, ErrPerturbationEngine
	}
	angle := func(period float64) float64 { return 2 * math.Pi * (jdTDB - 2451545) / period }
	switch target {
	case 10:
		phase := angle(4332.6)
		pos = [3]float64{-742000 * math.Cos(phase), -742000 * math.Sin(phase), 0}
	case 3:
		phase := angle(365.25)
		pos = [3]float64{149597870 * math.Cos(phase), 149597870 * math.Sin(phase), 0}
	case 5:
		phase := angle(4332.6)
		pos = [3]float64{778000000 * math.Cos(phase), 778000000 * math.Sin(phase), 0}
	default:
		return pos, vel, ErrPerturbationEngine
	}
	return pos, vel, nil
}

func (movingPerturbationEngine) Coverage(target Body) (start, end float64, ok bool) {
	return 0, 3000000, true
}

func TestEphemerisEngineMatchesFreshPropagationInAnyOrder(t *testing.T) {
	elements := Elements{EpochJD: 2451545, A: 1.5, E: 0.02, I: 3, Omega: 20, W: 30, M0: 10}
	options := PropagationOptions{Perturbers: []Body{5, 3}, Engine: movingPerturbationEngine{}, Step: 0.7, Tolerance: 1e-9}
	queries := []float64{2451545 + 400.3, 2451545 + 12.25, 2451545 - 90.1, 2451545 + 400.3, 2451545 + 33.6, 2451545 - 3.5, 2451545, 2451545 + 1000}
	engine := NewEphemerisEngine(Body(12345), elements, options)
	for _, jd := range queries {
		gotPos, gotVel, err := engine.State(Body(12345), 10, jd)
		if err != nil {
			t.Fatal(err)
		}
		wantPos, wantVel, err := StateWithOptions(elements, jd, options)
		if err != nil {
			t.Fatal(err)
		}
		if gotPos != wantPos || gotVel != wantVel {
			t.Fatalf("jd %.2f: engine %v %v, fresh %v %v", jd, gotPos, gotVel, wantPos, wantVel)
		}
	}
	// 逆序查询同样逐位一致 / reverse order also matches bit for bit.
	reverse := NewEphemerisEngine(Body(12345), elements, options)
	for index := len(queries) - 1; index >= 0; index-- {
		gotPos, _, err := reverse.State(Body(12345), 10, queries[index])
		if err != nil {
			t.Fatal(err)
		}
		wantPos, _, _ := StateWithOptions(elements, queries[index], options)
		if gotPos != wantPos {
			t.Fatalf("reverse jd %.2f: engine %v, fresh %v", queries[index], gotPos, wantPos)
		}
	}
}

func TestEphemerisEngineKeepsConstructionElements(t *testing.T) {
	elements := Elements{EpochJD: 2451545, A: 1.5, E: 0.02, I: 3, Omega: 20, W: 30, M0: 10}
	options := PropagationOptions{Perturbers: []Body{5}, Engine: movingPerturbationEngine{}, Step: 1}
	engine := NewEphemerisEngine(Body(12345), elements, options)
	if _, _, err := engine.State(Body(12345), 10, 2451545+50); err != nil {
		t.Fatal(err)
	}
	options.Perturbers[0] = 3
	got := engine.Elements()
	got.A = 9
	gotPos, _, err := engine.State(Body(12345), 10, 2451545+80)
	if err != nil {
		t.Fatal(err)
	}
	wantPos, _, _ := StateWithOptions(elements, 2451545+80, PropagationOptions{Perturbers: []Body{5}, Engine: movingPerturbationEngine{}, Step: 1})
	if gotPos != wantPos || engine.Elements() != elements || engine.Target() != Body(12345) {
		t.Fatalf("engine state changed after caller mutation: %v versus %v", gotPos, wantPos)
	}
}

func TestStateWithOptionsPerihelionFormStartsAtOsculatingEpoch(t *testing.T) {
	elements := Elements{EpochJD: 2451545 + 120, Q: 0.6, E: 0.9, I: 30, Omega: 40, W: 50, TpJD: 2451545}
	options := PropagationOptions{Perturbers: []Body{5}, Engine: movingPerturbationEngine{}, Step: 1, Tolerance: 1e-10}
	gotPos, gotVel, err := StateWithOptions(elements, elements.EpochJD, options)
	if err != nil {
		t.Fatal(err)
	}
	wantPos, wantVel, err := State(elements, elements.EpochJD)
	if err != nil {
		t.Fatal(err)
	}
	if gotPos != wantPos || gotVel != wantVel {
		t.Fatalf("osculating epoch state %v %v, want two-body %v %v", gotPos, gotVel, wantPos, wantVel)
	}
	// 未给 EpochJD 时仍以 TpJD 为起点 / without EpochJD the integration still starts at TpJD.
	noEpoch := elements
	noEpoch.EpochJD = 0
	gotPos, _, err = StateWithOptions(noEpoch, noEpoch.TpJD, options)
	if err != nil {
		t.Fatal(err)
	}
	wantPos, _, _ = State(noEpoch, noEpoch.TpJD)
	if gotPos != wantPos {
		t.Fatalf("TpJD start state %v, want %v", gotPos, wantPos)
	}
}

func TestStateWithOptionsRejectsUnsupportedPerturbers(t *testing.T) {
	elements := Elements{EpochJD: 2451545, A: 1.5, E: 0.02, I: 3, Omega: 20, W: 30, M0: 10}
	for _, perturbers := range [][]Body{{5, 9}, {2000001}, {3, 399}, {301, 3}} {
		_, _, err := StateWithOptions(elements, elements.EpochJD+5, PropagationOptions{Perturbers: perturbers, Engine: movingPerturbationEngine{}})
		if !errors.Is(err, ErrUnsupportedPerturber) {
			t.Fatalf("perturbers %v error = %v, want ErrUnsupportedPerturber", perturbers, err)
		}
		engine := NewEphemerisEngine(Body(12345), elements, PropagationOptions{Perturbers: perturbers, Engine: movingPerturbationEngine{}})
		if _, _, err := engine.State(Body(12345), 10, elements.EpochJD+5); !errors.Is(err, ErrUnsupportedPerturber) {
			t.Fatalf("engine perturbers %v error = %v, want ErrUnsupportedPerturber", perturbers, err)
		}
	}
	plain, _, _ := State(elements, elements.EpochJD+30)
	withEMB, _, err := StateWithOptions(elements, elements.EpochJD+30, PropagationOptions{Perturbers: []Body{3}, Engine: movingPerturbationEngine{}, Tolerance: 1e-10})
	if err != nil {
		t.Fatal(err)
	}
	if withEMB == plain {
		t.Fatal("EMB perturbation did not change the state")
	}
}

func TestStateWithOptionsRejectsMissingClassicalEpoch(t *testing.T) {
	elements := Elements{A: 1.5, E: 0.02, I: 3, Omega: 20, W: 30, M0: 10}
	_, _, err := StateWithOptions(elements, 2451545, PropagationOptions{Perturbers: []Body{3}, Engine: movingPerturbationEngine{}})
	if !errors.Is(err, ErrInvalidPropagationElements) {
		t.Fatalf("error = %v, want ErrInvalidPropagationElements", err)
	}
}

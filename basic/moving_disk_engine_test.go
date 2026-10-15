package basic

import (
	"math"
	"testing"
)

func TestMovingDiskEventEngineSampleTimesKeepAnchorsAndBudget(t *testing.T) {
	engine := movingDiskEventEngine{maxSampleCount: 7, reserveAnchorSlot: true, uniformOverflow: true}
	times, step := engine.sampleTimes(0, 1, 0.37, 0.01)
	if len(times) != 7 {
		t.Fatalf("sample count = %d, want 7", len(times))
	}
	if times[0] != 0 || times[len(times)-1] != 1 {
		t.Fatalf("sample endpoints = %.6f, %.6f, want 0 and 1", times[0], times[len(times)-1])
	}
	foundGreatest := false
	for index, value := range times {
		if index > 0 && value <= times[index-1] {
			t.Fatalf("sample times are not strictly increasing: %v", times)
		}
		if math.Abs(value-0.37) <= 1e-12 {
			foundGreatest = true
		}
	}
	if !foundGreatest {
		t.Fatalf("sample times omitted greatest: %v", times)
	}
	if step <= 0 {
		t.Fatalf("effective step = %v, want positive", step)
	}
}

func TestMovingDiskEventEngineWindowUsesCandidateAndExactPredicates(t *testing.T) {
	engine := movingDiskEventEngine{
		searchSpanDays:    1,
		rangeStepDays:     0.1,
		rootToleranceDays: 1e-9,
	}
	candidateCalls, exactCalls := 0, 0
	candidate := func(tt float64) bool {
		candidateCalls++
		return tt >= -0.35 && tt <= 0.42
	}
	exact := func(tt float64) bool {
		exactCalls++
		return tt >= -0.3 && tt <= 0.4
	}
	start, end, ok := engine.window(0, -1, 1, candidate, exact)
	if !ok {
		t.Fatal("window returned not found")
	}
	if math.Abs(start+0.3) > 2e-8 || math.Abs(end-0.4) > 2e-8 {
		t.Fatalf("window = %.12f..%.12f, want -0.3..0.4", start, end)
	}
	if candidateCalls == 0 || exactCalls == 0 {
		t.Fatalf("candidate/exact calls = %d/%d, want both", candidateCalls, exactCalls)
	}
}

func TestMovingDiskEventEngineGreatestFindsMinimumImpact(t *testing.T) {
	engine := movingDiskEventEngine{greatestSpanDays: 0.75}
	got := engine.greatest(0.1, -1, 1, func(tt float64) (float64, bool) {
		return (tt - 0.23) * (tt - 0.23), true
	}, 56)
	if math.Abs(got-0.23) > 1e-8 {
		t.Fatalf("greatest = %.12f, want 0.23", got)
	}
}

func TestMovingDiskContactStateUsesSolarOuterAndInnerRadii(t *testing.T) {
	state := movingDiskContactState{
		separation:           1.0,
		occultingOuterRadius: 0.6,
		occultingInnerRadius: 0.4,
		targetRadius:         0.5,
		valid:                true,
	}
	if got := state.externalContactGap(); math.Abs(got+0.1) > 1e-12 {
		t.Fatalf("external gap = %.12f, want -0.1", got)
	}
	if got := state.internalContactGap(); math.Abs(got-0.9) > 1e-12 {
		t.Fatalf("internal gap = %.12f, want 0.9", got)
	}
	if !movingDiskContactStateValid(1, 0.6, 0.4, 0.5) {
		t.Fatal("finite disk state was rejected")
	}
	if movingDiskContactStateValid(math.NaN(), 0.6, 0.4, 0.5) {
		t.Fatal("NaN disk state was accepted")
	}
}

func TestMovingDiskEventEngineContactRootWalksBothDirections(t *testing.T) {
	engine := movingDiskEventEngine{}
	metric := func(tt float64) (float64, bool) {
		return (tt + 0.3) * (tt - 0.4), true
	}
	for _, test := range []struct {
		name      string
		direction float64
		want      float64
	}{
		{name: "forward", direction: 1, want: 0.4},
		{name: "backward", direction: -1, want: -0.3},
	} {
		got, ok := engine.contactRoot(0, test.direction, 0.05, 1, 1e-9, metric, 64)
		if !ok || math.Abs(got-test.want) > 2e-8 {
			t.Errorf("%s contact root = %.12f, ok=%v, want %.3f", test.name, got, ok, test.want)
		}
	}
}

func TestMovingDiskContactEvaluatorCachesValidAndInvalidStates(t *testing.T) {
	calls := 0
	evaluator := newMovingDiskContactEvaluator(func(tt float64) (movingDiskContactState, bool) {
		calls++
		if tt < 0 {
			return movingDiskContactState{}, false
		}
		state := movingDiskContactState{
			separation:           1,
			occultingOuterRadius: 0.6,
			occultingInnerRadius: 0.6,
			targetRadius:         0.2,
			valid:                true,
		}
		return state, true
	})
	if _, ok := evaluator.gap(0.25, false); !ok {
		t.Fatal("first valid evaluation failed")
	}
	if _, ok := evaluator.gap(0.25, true); !ok {
		t.Fatal("cached valid evaluation failed")
	}
	if _, ok := evaluator.gap(-0.25, false); ok {
		t.Fatal("invalid evaluation was accepted")
	}
	if _, ok := evaluator.gap(-0.25, true); ok {
		t.Fatal("cached invalid evaluation was accepted")
	}
	if calls != 2 {
		t.Fatalf("evaluation calls = %d, want 2 after valid/invalid cache hits", calls)
	}
}

func TestMovingDiskContactEvaluatorPrimeAvoidsCallback(t *testing.T) {
	calls := 0
	evaluator := newMovingDiskContactEvaluator(func(tt float64) (movingDiskContactState, bool) {
		calls++
		return movingDiskContactState{}, false
	})
	state := movingDiskContactState{
		separation:           1,
		occultingOuterRadius: 0.6,
		occultingInnerRadius: 0.6,
		targetRadius:         0.2,
		valid:                true,
	}
	evaluator.prime(0.5, state, true)
	value, ok := evaluator.gap(0.5, false)
	if !ok || math.Abs(value-0.2) > 1e-12 {
		t.Fatalf("primed external gap = %.12f, ok=%v", value, ok)
	}
	if calls != 0 {
		t.Fatalf("callback calls = %d, want 0 for primed TT", calls)
	}
}

func TestMovingDiskContactCacheClearsAtBoundedCapacity(t *testing.T) {
	cache := &movingDiskContactCache{maxEntries: 2}
	state := movingDiskContactState{valid: true}
	cache.store(1, state, true)
	cache.store(2, state, true)
	cache.store(3, state, true)
	if len(cache.entries) != 1 {
		t.Fatalf("cache size = %d, want 1 after bounded clear", len(cache.entries))
	}
	if _, _, hit := cache.lookup(1); hit {
		t.Fatal("old entry survived bounded cache clear")
	}
	if _, _, hit := cache.lookup(3); !hit {
		t.Fatal("new entry missing after bounded cache clear")
	}
}

func TestMovingDiskEventEngineSampleTimesBoundsTinyStep(t *testing.T) {
	engine := movingDiskEventEngine{maxSampleCount: 9}
	for _, step := range []float64{1e-300, math.SmallestNonzeroFloat64} {
		times, _ := engine.sampleTimes(0, 1, 0.5, step)
		if len(times) > 9 {
			t.Fatalf("step %g: sample count = %d, want <= 9", step, len(times))
		}
	}
}

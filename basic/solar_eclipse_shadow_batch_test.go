package basic

import "testing"

// 批量入口是导出契约的一部分，必须与逐时刻入口给出同一条数值链。
func TestSolarEclipseShadowBatchEntriesMatchPerInstantCalls(t *testing.T) {
	footprints := SolarEclipsePartialFootprints(JDCalc(2009, 7, 22), SolarEclipsePartialFootprintOptions{
		StepDays: 6.0 / 1440.0, BoundaryPoints: 48, DisableRiseSetCurves: true,
		CentralShadowStepDays: 6.0 / 1440.0,
	})
	if len(footprints.CentralShadowFootprints) < 3 {
		t.Fatalf("central shadow samples=%d, want a full series", len(footprints.CentralShadowFootprints))
	}
	options := SolarEclipseShadowSolverOptions{}
	step := 6.0 / 1440.0
	start := footprints.CentralShadowFootprints[0].JDE
	end := start + 4*step
	batch := SolarEclipseShadowBetweenJDE(start, end, step, options)
	if len(batch) != 5 {
		t.Fatalf("batch length=%d, want 5", len(batch))
	}
	nonEmpty := 0
	for index, instant := range batch {
		direct, ok := SolarEclipseShadowAtJDE(start+float64(index)*step, options)
		if ok != !direct.Empty() {
			t.Fatalf("instant %d ok=%v, Empty=%v", index, ok, direct.Empty())
		}
		if instant.Empty() != direct.Empty() || len(instant.Boundaries) != len(direct.Boundaries) {
			t.Fatalf("instant %d batch=%d segments, direct=%d", index, len(instant.Boundaries), len(direct.Boundaries))
		}
		if !instant.Empty() {
			nonEmpty++
		}
	}
	if nonEmpty == 0 {
		t.Fatal("batch reported no shadow anywhere along the central path")
	}
	longitude, latitude := footprints.Eclipse.GreatestLongitude, footprints.Eclipse.GreatestLatitude
	solver := NewSolarEclipseShadowSolver(options)
	states := solver.StationStatesBetweenJDE(start, end, step, longitude, latitude, 0)
	if len(states) != len(batch) {
		t.Fatalf("station state batch length=%d, want %d", len(states), len(batch))
	}
	for index, state := range states {
		direct := SolarEclipseStationStateAtJDE(start+float64(index)*step, longitude, latitude, 0, options)
		if state != direct {
			t.Fatalf("station state %d = %+v, want %+v", index, state, direct)
		}
	}
	for _, badStep := range []float64{0, -step} {
		if got := SolarEclipseShadowBetweenJDE(start, end, badStep, options); got != nil {
			t.Fatalf("step %v produced %d instants, want nil", badStep, len(got))
		}
	}
	if got := solver.ShadowBetweenJDE(end, start, step); got != nil {
		t.Fatalf("reversed window produced %d instants, want nil", len(got))
	}
	if got := SolarEclipseShadowBetweenJDE(start, start+10, 1e-9, options); got != nil {
		t.Fatalf("oversized batch produced %d instants, want nil", len(got))
	}
}

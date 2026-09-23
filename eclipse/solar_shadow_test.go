package eclipse

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestSolarEclipseShadowSolverUTCAgreesWithTT(t *testing.T) {
	value := time.Date(2009, time.July, 22, 4, 19, 23, 378387689, time.UTC)
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	fromUTC, okUTC := solver.ShadowAt(value)
	if !okUTC {
		t.Fatal("expected an umbral footprint at the test instant")
	}
	// 显式 ΔT 的 TT 入口必须与 UTC 入口给出同一份几何。
	explicit := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{DeltaTSeconds: fromUTC.DeltaTSeconds})
	fromTT, okTT := explicit.ShadowAtJDE(fromUTC.JDE)
	if !okTT {
		t.Fatal("TT entry reported no umbra")
	}
	if len(fromTT.Boundaries) != len(fromUTC.Boundaries) {
		t.Fatalf("segments differ: TT=%d UTC=%d", len(fromTT.Boundaries), len(fromUTC.Boundaries))
	}
	for index := range fromUTC.Boundaries {
		if len(fromTT.Boundaries[index]) != len(fromUTC.Boundaries[index]) {
			t.Fatalf("segment %d length differs", index)
		}
		for pointIndex := range fromUTC.Boundaries[index] {
			a, b := fromTT.Boundaries[index][pointIndex], fromUTC.Boundaries[index][pointIndex]
			if a.Longitude != b.Longitude || a.Latitude != b.Latitude {
				t.Fatalf("point %d/%d differs: %.12f,%.12f vs %.12f,%.12f",
					index, pointIndex, a.Longitude, a.Latitude, b.Longitude, b.Latitude)
			}
		}
	}
	if fromTT.Topology.Signature() != fromUTC.Topology.Signature() {
		t.Fatalf("signature differs: %q vs %q", fromTT.Topology.Signature(), fromUTC.Topology.Signature())
	}
	if math.Abs(fromTT.DeltaTSeconds-fromUTC.DeltaTSeconds) > 1e-9 {
		t.Fatalf("ΔT differs: %.9f vs %.9f", fromTT.DeltaTSeconds, fromUTC.DeltaTSeconds)
	}
	if fromTT.Time.IsZero() {
		t.Fatal("TT entry must still report a usable instant instead of the zero time")
	}
	state := explicit.StationStateAtJDE(fromUTC.JDE, -96.8, 32.8, 0)
	if state.Time.IsZero() {
		t.Fatal("TT station entry must report a usable instant")
	}
}

func TestSolarEclipseShadowBetweenAlignsWithTimeline(t *testing.T) {
	// 覆盖 2009-07-22 的 U1..U4（本影在地球上的阶段）。
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	start := time.Date(2009, time.July, 22, 0, 50, 0, 0, time.UTC)
	end := time.Date(2009, time.July, 22, 4, 25, 0, 0, time.UTC)
	// 2 分钟步长才能踩到被地平线切断的那几分钟（该阶段只有约 3 分钟宽）。
	step := 2 * time.Minute
	instants := solver.ShadowBetween(start, end, step)
	want := int(end.Sub(start)/step) + 1
	if len(instants) != want {
		t.Fatalf("batch length=%d, want %d", len(instants), want)
	}
	empty, populated, closedSeen, horizonSeen := 0, 0, 0, 0
	for index, instant := range instants {
		if instant.Empty() {
			empty++
			continue
		}
		populated++
		if instant.Time.Before(start) {
			t.Fatalf("entry %d time %v before start", index, instant.Time)
		}
		if instant.Closed {
			closedSeen++
		} else {
			horizonSeen++
			if len(instant.HorizonEnds) != 2 {
				t.Fatalf("entry %d is horizon-cut but has %d grazing points", index, len(instant.HorizonEnds))
			}
		}
	}
	if empty == 0 || populated == 0 {
		t.Fatalf("batch should contain both empty and populated entries, got empty=%d populated=%d", empty, populated)
	}
	if closedSeen == 0 || horizonSeen == 0 {
		t.Fatalf("batch should contain both topology kinds, got closed=%d horizon=%d", closedSeen, horizonSeen)
	}
}

func TestSolarEclipseShadowBatchMatchesSingleCalls(t *testing.T) {
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	start := time.Date(2009, time.July, 22, 1, 0, 0, 0, time.UTC)
	step := 20 * time.Minute
	instants := solver.ShadowBetween(start, start.Add(2*time.Hour), step)
	if len(instants) == 0 {
		t.Fatal("empty batch")
	}
	for index, batch := range instants {
		single, _ := solver.ShadowAt(start.Add(time.Duration(index) * step))
		if single.Empty() != batch.Empty() {
			t.Fatalf("entry %d emptiness differs", index)
		}
		if single.Empty() {
			continue
		}
		if single.Topology.Signature() != batch.Topology.Signature() {
			t.Fatalf("entry %d signature %q vs %q", index, batch.Topology.Signature(), single.Topology.Signature())
		}
	}
}

func TestSolarEclipseStationStateExplicitDeltaTMovesStation(t *testing.T) {
	value := time.Date(2024, time.April, 8, 18, 42, 0, 0, time.UTC)
	base := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	shifted := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{DeltaTSeconds: base.stationDeltaT(value) + 100})
	first := base.StationStateAt(value, -96.8, 32.8, 0)
	second := shifted.StationStateAt(value, -96.8, 32.8, 0)
	if math.Abs(first.JDE-second.JDE) > 100.0/86400 {
		t.Fatalf("the TT instant must stay within one ΔT step, got %.9f vs %.9f", first.JDE, second.JDE)
	}
	if math.Abs(second.DeltaTSeconds-first.DeltaTSeconds-100) > 1e-9 {
		t.Fatalf("ΔT override not applied: %.6f vs %.6f", first.DeltaTSeconds, second.DeltaTSeconds)
	}
	// 站心角距对地面横移的响应量级：100 s ≈ 39 km ≈ 21 角秒；精确的
	// "位移 = DeltaTGroundShiftKM"由 basic 层的足迹测试覆盖，这里只验证量级与换算系数。
	if math.Abs(second.SeparationArcsec-first.SeparationArcsec) < 5 {
		t.Fatalf("explicit ΔT did not move the station geometry: %.3f vs %.3f arcsec",
			first.SeparationArcsec, second.SeparationArcsec)
	}
	if ground := basic.DeltaTGroundShiftKM(100, 32.8); ground < 38 || ground > 40 {
		t.Fatalf("DeltaTGroundShiftKM(100 s, 32.8°) = %.2f km, want about 39 km", ground)
	}
}

func (solver *SolarEclipseShadowSolver) stationDeltaT(value time.Time) float64 {
	state := solver.StationStateAt(value, 0, 0, 0)
	return state.DeltaTSeconds
}

func TestSolarEclipseShadowTTEntryReportsCivilTime(t *testing.T) {
	// 句柄 ΔT 只改地球自转相位：TT 入口回填的必须是民用时刻（TT2UTC），
	// 而不是 TT − ΔT —— 后者是 UT1，与民用时刻差一个 DUT1。
	override := 3666.18
	value := time.Date(2009, time.July, 22, 0, 53, 0, 0, time.UTC)
	model := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	jde := model.ttJDE(value)
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{DeltaTSeconds: override})
	instant, ok := solver.ShadowAtJDE(jde)
	if !ok {
		t.Fatal("expected a footprint at the shifted instant")
	}
	want := solarEclipseTTJDEToTime(jde, time.UTC)
	if !instant.Time.Equal(want) {
		t.Fatalf("TT entry time=%v, want the civil time %v", instant.Time.UTC(), want.UTC())
	}
	if instant.DeltaTSeconds != override {
		t.Fatalf("reported ΔT=%.6f, want %.2f", instant.DeltaTSeconds, override)
	}
	if shift := instant.Time.Sub(value); shift > time.Millisecond || shift < -time.Millisecond {
		t.Fatalf("TT entry time=%v, want the civil instant %v", instant.Time.UTC(), value)
	}
	state := solver.StationStateAtJDE(jde, -96.8, 32.8, 0)
	if !state.Time.Equal(want) {
		t.Fatalf("station TT entry time=%v, want the civil time %v", state.Time.UTC(), want.UTC())
	}
}

func TestSolarEclipseShadowBetweenSurvivesLongWindows(t *testing.T) {
	// time.Duration 是 int64 纳秒：start.Add(time.Duration(index)*step) 在跨度超过
	// 约 292 年时溢出，Time 会回绕倒退。30 天步长 × 300 年只有 3651 条，刚好越过
	// 292 年的边界，用很小的代价钉住这个回归。
	// time.Duration is int64 nanoseconds: start.Add(time.Duration(index)*step) overflows
	// past ~292 years and Time wraps backwards. A 30-day step over 300 years is only
	// 3651 entries yet crosses that boundary, so it pins the regression cheaply.
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	start := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(300, 0, 0)
	step := 30 * 24 * time.Hour
	instants := solver.ShadowBetween(start, end, step)
	// time.Time.Sub also saturates at the ~292-year duration limit, so the expected count
	// comes from Unix seconds (int64 seconds has no such limit here).
	want := int((end.Unix()-start.Unix())/int64(step.Seconds())) + 1
	if len(instants) != want {
		t.Fatalf("batch length=%d, want %d", len(instants), want)
	}
	previous := start.Add(-time.Second)
	for index, instant := range instants {
		if !instant.Time.After(previous) {
			t.Fatalf("entry %d time %v is not after %v (int64 nanosecond wrap)",
				index, instant.Time, previous)
		}
		if index == 0 && !instant.Time.Equal(start) {
			t.Fatalf("entry 0 time %v, want %v", instant.Time, start)
		}
		previous = instant.Time
	}
	if !previous.Before(end.Add(step)) {
		t.Fatalf("last entry %v overshoots the window", previous)
	}
}

func TestSolarEclipseStationStateAtJDERejectsNonFiniteTT(t *testing.T) {
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	for _, jdeTT := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		state := solver.StationStateAtJDE(jdeTT, 116.404, 39.915, 0)
		if state != (SolarEclipseStationState{}) {
			t.Fatalf("StationStateAtJDE(%v) = %+v, want the zero state", jdeTT, state)
		}
	}
	if state := solver.StationStateAtJDE(2460310.5, 116.404, 39.915, 0); state.JDE != 2460310.5 || state.Time.IsZero() {
		t.Fatalf("finite TT returned %+v", state)
	}
}

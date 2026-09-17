package basic

import (
	"math"
	"testing"
)

func TestSolarEclipseShadowInstantMatchesPackagedSamples(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDECalc(2009, 7, 22), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, CentralShadowStepDays: 2.0 / 1440.0,
		DisableRiseSetCurves: true,
	})
	if len(result.CentralShadowFootprints) < 3 {
		t.Fatalf("central shadow samples=%d, want a full series", len(result.CentralShadowFootprints))
	}
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	checked := map[string]bool{}
	for _, index := range []int{0, len(result.CentralShadowFootprints) / 2, len(result.CentralShadowFootprints) - 1} {
		sampled := result.CentralShadowFootprints[index]
		instant, ok := solver.ShadowAtJDE(sampled.JDE)
		if !ok {
			t.Fatalf("instant call reported no umbra at %v, packaged sample has %d segments",
				sampled.JDE, len(sampled.Boundaries))
		}
		if instant.Closed != sampled.Closed {
			t.Fatalf("closed mismatch at %v: instant=%v packaged=%v", sampled.JDE, instant.Closed, sampled.Closed)
		}
		if len(instant.Boundaries) != len(sampled.Boundaries) {
			t.Fatalf("segments mismatch at %v: instant=%d packaged=%d",
				sampled.JDE, len(instant.Boundaries), len(sampled.Boundaries))
		}
		for segmentIndex := range sampled.Boundaries {
			if len(instant.Boundaries[segmentIndex]) != len(sampled.Boundaries[segmentIndex]) {
				t.Fatalf("segment %d length mismatch at %v: instant=%d packaged=%d", segmentIndex, sampled.JDE,
					len(instant.Boundaries[segmentIndex]), len(sampled.Boundaries[segmentIndex]))
			}
			for pointIndex := range sampled.Boundaries[segmentIndex] {
				got := instant.Boundaries[segmentIndex][pointIndex]
				want := sampled.Boundaries[segmentIndex][pointIndex]
				if got.Longitude != want.Longitude || got.Latitude != want.Latitude {
					t.Fatalf("point mismatch at %v segment %d point %d: got %.12f,%.12f want %.12f,%.12f",
						sampled.JDE, segmentIndex, pointIndex,
						got.Longitude, got.Latitude, want.Longitude, want.Latitude)
				}
			}
		}
		checked[instant.Topology.Signature()] = true
	}
	if len(checked) < 2 {
		t.Fatalf("expected distinct topology signatures across the series, got %v", checked)
	}
}

func TestSolarEclipseShadowInstantEmptyOffPath(t *testing.T) {
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	// 2009-07-22 食甚前后 12 小时已经远离地球上的本影。
	instant, ok := solver.ShadowAtJDE(JDECalc(2009, 7, 22) + 0.6)
	if ok || !instant.Empty() {
		t.Fatalf("off-path instant reported ok=%v empty=%v", ok, instant.Empty())
	}
	if instant.Topology.Signature() != "empty" {
		t.Fatalf("empty topology signature=%q, want empty", instant.Topology.Signature())
	}
	if instant.DeltaTSeconds <= 0 {
		t.Fatalf("off-path instant must still report the ΔT used, got %v", instant.DeltaTSeconds)
	}
}

func TestSolarEclipseShadowTopologySignatureChangesAtHorizonCut(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDECalc(2009, 7, 22), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, CentralShadowStepDays: 2.0 / 1440.0,
		DisableRiseSetCurves: true,
	})
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	signatures := map[string]int{}
	for _, sampled := range result.CentralShadowFootprints {
		instant, ok := solver.ShadowAtJDE(sampled.JDE)
		if !ok {
			continue
		}
		signatures[instant.Topology.Signature()]++
	}
	closedSeen, horizonSeen := false, false
	for signature := range signatures {
		if len(signature) >= 7 && signature[:7] == "umbra-c" {
			closedSeen = true
		}
		if len(signature) >= 7 && signature[:7] == "umbra-h" {
			horizonSeen = true
		}
	}
	if !closedSeen || !horizonSeen {
		t.Fatalf("expected both self-closed and horizon-cut signatures, got %v", signatures)
	}
}

func TestSolarEclipseStationStateMatchesLocalEclipse(t *testing.T) {
	seed := JDECalc(2024, 4, 8)
	const lon, lat = -96.8, 32.8
	local := LocalSolarEclipse(seed, lon, lat, 0)
	if !local.HasTotal {
		t.Fatalf("expected a total eclipse at the test station, got type=%v magnitude=%.3f", local.Type, local.Magnitude)
	}
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	state := solver.StationStateAtJDE(local.GreatestEclipse, lon, lat, 0)
	if math.Abs(state.SeparationDeg-local.Separation) > 1e-9 {
		t.Fatalf("separation mismatch: station=%.12f local=%.12f", state.SeparationDeg, local.Separation)
	}
	if math.Abs(state.SunAltitudeDeg-local.SunAltitude) > 1e-9 || math.Abs(state.SunAzimuthDeg-local.SunAzimuth) > 1e-9 {
		t.Fatalf("sun position mismatch: station=(%.9f,%.9f) local=(%.9f,%.9f)",
			state.SunAltitudeDeg, state.SunAzimuthDeg, local.SunAltitude, local.SunAzimuth)
	}
	if math.Abs(state.Obscuration-local.Obscuration) > 1e-12 {
		t.Fatalf("obscuration mismatch: station=%.12f local=%.12f", state.Obscuration, local.Obscuration)
	}
	if !state.InCentralPhase || !state.HasTotalPhase || state.HasAnnularPhase {
		t.Fatalf("central phase flags wrong: %+v", state)
	}
	if !state.Visible {
		t.Fatal("expected the Sun above the horizon at greatest eclipse")
	}
}

func TestSolarEclipseShadowDeltaTMovesOnlyEarthRotation(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDECalc(2009, 7, 22), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, CentralShadowStepDays: 2.0 / 1440.0,
		DisableRiseSetCurves: true,
	})
	if len(result.CentralShadowFootprints) == 0 {
		t.Fatal("no central shadow samples")
	}
	jde := result.CentralShadowFootprints[len(result.CentralShadowFootprints)/2].JDE
	base := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	shifted := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{DeltaTSeconds: 100})
	first, okFirst := base.ShadowAtJDE(jde)
	second, okSecond := shifted.ShadowAtJDE(jde)
	if !okFirst || !okSecond {
		t.Fatalf("expected a footprint at %v (ok=%v/%v)", jde, okFirst, okSecond)
	}
	if first.JDE != second.JDE {
		t.Fatalf("geometry instant moved with ΔT: %v vs %v", first.JDE, second.JDE)
	}
	if math.Abs(second.DeltaTSeconds-100) > 1e-12 || first.DeltaTSeconds == second.DeltaTSeconds {
		t.Fatalf("ΔT not reported per solver: %.6f vs %.6f", first.DeltaTSeconds, second.DeltaTSeconds)
	}
	centroid := func(instant SolarEclipseShadowInstant) (float64, float64) {
		var sumX, sumY, sumZ, count float64
		for _, segment := range instant.Boundaries {
			for _, point := range segment {
				longitude, latitude := point.Longitude*rad, point.Latitude*rad
				sumX += math.Cos(latitude) * math.Cos(longitude)
				sumY += math.Cos(latitude) * math.Sin(longitude)
				sumZ += math.Sin(latitude)
				count++
			}
		}
		length := math.Sqrt(sumX*sumX + sumY*sumY + sumZ*sumZ)
		return math.Atan2(sumY/length, sumX/length) / rad, math.Asin(sumZ/length) / rad
	}
	firstLon, firstLat := centroid(first)
	secondLon, secondLat := centroid(second)
	shift := DeltaTGroundShiftKM(100-first.DeltaTSeconds, firstLat)
	measured := solarEclipsePathDistanceKM(
		SolarEclipsePathPoint{Longitude: firstLon, Latitude: firstLat},
		SolarEclipsePathPoint{Longitude: secondLon, Latitude: secondLat},
	)
	if math.Abs(measured-shift) > 0.05*shift {
		t.Fatalf("ΔT ground shift=%.1f km, want about %.1f km", measured, shift)
	}
}

func BenchmarkSolarEclipseShadowAtJDE(b *testing.B) {
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	result := SolarEclipsePartialFootprints(JDECalc(2009, 7, 22), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, CentralShadowStepDays: 2.0 / 1440.0,
		DisableRiseSetCurves: true,
	})
	jde := result.CentralShadowFootprints[len(result.CentralShadowFootprints)/2].JDE
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, ok := solver.ShadowAtJDE(jde + float64(index)*1e-9); !ok {
			b.Fatal("no footprint")
		}
	}
}

func BenchmarkSolarEclipseStationStateAtJDE(b *testing.B) {
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	jde := JDECalc(2024, 4, 8) + 0.78
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_ = solver.StationStateAtJDE(jde+float64(index)*1e-9, -96.8, 32.8, 0)
	}
}

func TestSolarEclipseShadowClampsBoundaryPoints(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDECalc(2009, 7, 22), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, CentralShadowStepDays: 2.0 / 1440.0,
		DisableRiseSetCurves: true,
	})
	if len(result.CentralShadowFootprints) == 0 {
		t.Fatal("no central shadow samples")
	}
	jde := result.CentralShadowFootprints[len(result.CentralShadowFootprints)/2].JDE
	// 个位数的边界点数会产出无意义的"足迹"（真实案例：1 个点也能 ok=true），必须被夹到下限。
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{BoundaryPoints: 1})
	instant, ok := solver.ShadowAtJDE(jde)
	if !ok {
		t.Fatal("expected a footprint")
	}
	points := 0
	for _, segment := range instant.Boundaries {
		points += len(segment)
	}
	if points < solarEclipsePartialFootprintMinBoundaryPoints {
		t.Fatalf("BoundaryPoints=1 produced %d vertices, want at least %d",
			points, solarEclipsePartialFootprintMinBoundaryPoints)
	}
	upper := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{BoundaryPoints: 100000})
	clamped, okClamped := upper.ShadowAtJDE(jde)
	if !okClamped {
		t.Fatal("expected a footprint with a clamped point count")
	}
	clampedPoints := 0
	for _, segment := range clamped.Boundaries {
		clampedPoints += len(segment)
	}
	if clampedPoints > solarEclipseShadowMaximumBoundaryPoints+2 {
		t.Fatalf("BoundaryPoints=100000 produced %d vertices, want at most %d",
			clampedPoints, solarEclipseShadowMaximumBoundaryPoints)
	}
}

func TestSolarEclipseShadowHandleSeesDeltaTChange(t *testing.T) {
	// 复用同一个句柄、中途覆盖进程级 ΔT：缓存里的贝塞尔轴带着由 ΔT 决定的 gst，而键
	// 只有 jd，旧实现会返回"新 ΔT + 旧几何"（实测边界漂移 0 km，新建句柄却差 58 km）。
	// Reusing one handle across a process-wide ΔT override used to replay the cached
	// Besselian axis, whose gst depends on ΔT, and report the new ΔT with the old
	// geometry (0 km drift where a fresh handle moved 58 km).
	original := GetDeltaTFn()
	defer SetDeltaTFn(original)

	SetDeltaTFn(DefaultDeltaTv2)
	// 取 2009-07-22 本影阶段内的一个真实时刻（与相邻测试同一取法）。
	// Use a real instant inside the 2009-07-22 umbral phase, derived like the neighbour test.
	samples := SolarEclipsePartialFootprints(JDECalc(2009, 7, 22), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, CentralShadowStepDays: 2.0 / 1440.0,
		DisableRiseSetCurves: true,
	})
	if len(samples.CentralShadowFootprints) == 0 {
		t.Fatal("no central shadow samples")
	}
	jde := samples.CentralShadowFootprints[len(samples.CentralShadowFootprints)/2].JDE
	reused := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	before, okBefore := reused.ShadowAtJDE(jde)
	if !okBefore {
		t.Fatalf("no footprint at %v with the process ΔT", jde)
	}

	SetDeltaTFn(func(float64, bool) float64 { return 200 })
	after, okAfter := reused.ShadowAtJDE(jde)
	if !okAfter {
		t.Fatalf("no footprint at %v after the ΔT override", jde)
	}
	fresh := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{})
	want, okWant := fresh.ShadowAtJDE(jde)
	if !okWant {
		t.Fatalf("no footprint at %v from a fresh handle", jde)
	}
	if math.Abs(after.DeltaTSeconds-200) > 1e-12 || math.Abs(want.DeltaTSeconds-200) > 1e-12 {
		t.Fatalf("ΔT not reported after the override: reused=%.6f fresh=%.6f",
			after.DeltaTSeconds, want.DeltaTSeconds)
	}

	maxDelta := func(first, second SolarEclipseShadowInstant) float64 {
		worst := 0.0
		for index, segment := range first.Boundaries {
			if index >= len(second.Boundaries) {
				return math.Inf(1)
			}
			if len(segment) != len(second.Boundaries[index]) {
				return math.Inf(1)
			}
			for pointIndex, point := range segment {
				other := second.Boundaries[index][pointIndex]
				worst = math.Max(worst, math.Abs(point.Longitude-other.Longitude))
				worst = math.Max(worst, math.Abs(point.Latitude-other.Latitude))
			}
		}
		return worst
	}
	if moved := maxDelta(before, after); moved < 1e-4 {
		t.Fatalf("reused handle did not move with ΔT (max %.9f deg)", moved)
	}
	if drift := maxDelta(after, want); drift > 1e-9 {
		t.Fatalf("reused handle differs from a fresh one by %.9f deg after the ΔT override", drift)
	}
}

func TestSolarEclipseShadowPenumbraMatchesPackagedSamples(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDECalc(2009, 7, 22), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, CentralShadowStepDays: 2.0 / 1440.0,
		DisableRiseSetCurves: true,
	})
	if len(result.Footprints) == 0 {
		t.Fatal("no partial footprints")
	}
	solver := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{
		Kind: SolarEclipseShadowPenumbra, BoundaryPoints: 96,
	})
	sampled := result.Footprints[len(result.Footprints)/2]
	instant, ok := solver.ShadowAtJDE(sampled.JDE)
	if !ok {
		t.Fatal("penumbra instant reported no footprint")
	}
	if instant.Kind != SolarEclipseShadowPenumbra {
		t.Fatalf("kind=%v, want penumbra", instant.Kind)
	}
	if instant.Closed != sampled.Closed || len(instant.Boundaries) != len(sampled.Boundaries) {
		t.Fatalf("shape mismatch: closed=%v/%v segments=%d/%d",
			instant.Closed, sampled.Closed, len(instant.Boundaries), len(sampled.Boundaries))
	}
	for index := range sampled.Boundaries {
		if len(instant.Boundaries[index]) != len(sampled.Boundaries[index]) {
			t.Fatalf("segment %d length %d, want %d", index, len(instant.Boundaries[index]), len(sampled.Boundaries[index]))
		}
		for pointIndex := range sampled.Boundaries[index] {
			got, want := instant.Boundaries[index][pointIndex], sampled.Boundaries[index][pointIndex]
			// 空间加密的递归顺序会让同一时刻相差约 1e-12 度，取 1e-9 度（亚毫米）即可。
			if math.Abs(got.Longitude-want.Longitude) > 1e-9 || math.Abs(got.Latitude-want.Latitude) > 1e-9 {
				t.Fatalf("point %d/%d = %.12f,%.12f want %.12f,%.12f",
					index, pointIndex, got.Longitude, got.Latitude, want.Longitude, want.Latitude)
			}
		}
	}
	if signature := instant.Topology.Signature(); len(signature) < 8 || signature[:8] != "penumbra" {
		t.Fatalf("signature=%q, want a penumbra prefix", signature)
	}
}

func TestSolarEclipseShadowKindChangesFootprint(t *testing.T) {
	result := SolarEclipsePartialFootprints(JDECalc(2009, 7, 22), SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, CentralShadowStepDays: 2.0 / 1440.0,
		DisableRiseSetCurves: true,
	})
	if len(result.CentralShadowFootprints) == 0 || len(result.Footprints) == 0 {
		t.Fatal("missing samples")
	}
	jde := result.CentralShadowFootprints[len(result.CentralShadowFootprints)/2].JDE
	umbra, okUmbra := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{}).ShadowAtJDE(jde)
	penumbra, okPenumbra := NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{
		Kind: SolarEclipseShadowPenumbra,
	}).ShadowAtJDE(jde)
	if !okUmbra || !okPenumbra {
		t.Fatalf("expected both footprints (umbra=%v penumbra=%v)", okUmbra, okPenumbra)
	}
	if umbra.Topology.Vertices == 0 || penumbra.Topology.Vertices == 0 {
		t.Fatal("empty topology")
	}
	umbraMin, umbraMax := longitudeRange(umbra)
	penumbraMin, penumbraMax := longitudeRange(penumbra)
	if penumbraMax-penumbraMin <= umbraMax-umbraMin {
		t.Fatalf("penumbra span %.1f should exceed the umbra span %.1f",
			penumbraMax-penumbraMin, umbraMax-umbraMin)
	}
}

func longitudeRange(instant SolarEclipseShadowInstant) (float64, float64) {
	minimum, maximum := 361.0, -361.0
	for _, segment := range instant.Boundaries {
		for _, point := range segment {
			if point.Longitude < minimum {
				minimum = point.Longitude
			}
			if point.Longitude > maximum {
				maximum = point.Longitude
			}
		}
	}
	return minimum, maximum
}

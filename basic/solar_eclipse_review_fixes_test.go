package basic

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/geodata"
)

// 采样并集的顶点必须保留自己的时间：平滑把顶点移了位置，时间场必须跟着同一个滤波器走，
// 不能整体回落到食甚时刻。
func TestSolarEclipseSampledBandVerticesKeepOwnTime(t *testing.T) {
	for _, date := range [][3]int{{4862, 9, 28}, {1552, 7, 21}, {1874, 10, 10}} {
		seed := JDCalc(date[0], date[1], float64(date[2]))
		result := SolarEclipsePartialFootprintsNASABulletinSplitK(seed, SolarEclipsePartialFootprintOptions{
			StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
		})
		if len(result.CentralBandSegments) == 0 {
			t.Fatalf("%04d-%02d-%02d exported no central band", date[0], date[1], date[2])
		}
		lowest, highest := math.Inf(1), math.Inf(-1)
		distinct := make(map[float64]bool)
		for _, segment := range result.CentralBandSegments {
			for _, point := range segment {
				lowest, highest = math.Min(lowest, point.JDE), math.Max(highest, point.JDE)
				distinct[point.JDE] = true
				if point.JDE < result.Eclipse.PartialBeginOnEarth-1 ||
					point.JDE > result.Eclipse.PartialEndOnEarth+1 {
					t.Fatalf("%04d-%02d-%02d band vertex time %.6f is outside the eclipse",
						date[0], date[1], date[2], point.JDE)
				}
			}
		}
		if len(distinct) < 2 || highest-lowest < 1.0/1440.0 {
			t.Fatalf("%04d-%02d-%02d band vertices collapsed onto one time: span=%.9f days, distinct=%d",
				date[0], date[1], date[2], highest-lowest, len(distinct))
		}
	}
}

// 扫掠兜底的采样标记必须跟着 centralBandSweepPolygons 的返回值走，不能丢。
func TestSolarEclipseCentralBandSweepKeepsSampledFlag(t *testing.T) {
	for _, date := range [][3]int{{2024, 4, 8}, {1136, 6, 1}, {4862, 9, 28}, {1552, 7, 21}} {
		seed := JDCalc(date[0], date[1], float64(date[2]))
		band := SolarEclipsePartialFootprintsNASABulletinSplitK(seed, SolarEclipsePartialFootprintOptions{
			StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
		})
		if band.U1.JDE == 0 || band.U4.JDE == 0 {
			t.Fatalf("%04d-%02d-%02d has no umbral contact pair", date[0], date[1], date[2])
		}
		solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
		_, _, sampled := solver.centralBandSweepPolygons(
			band.U1.JDE, band.U4.JDE, band.Eclipse.GreatestEclipse, band.RiseSetCurves,
		)
		probe := band
		probe.CentralBandSegments, probe.CentralBandSampled = nil, false
		solver.attachCentralBandSweep(
			&probe, band.U1.JDE, band.U4.JDE, band.Eclipse.GreatestEclipse, band.RiseSetCurves,
		)
		if probe.CentralBandSampled != sampled {
			t.Fatalf("%04d-%02d-%02d sampled flag=%v, sweep reports %v",
				date[0], date[1], date[2], probe.CentralBandSampled, sampled)
		}
	}
}

// 弧角沿环展开后可以超过 2π，包络角取的是 [0,2π) 支；映射到同一支之前，
// 展开段的样本会被两端同时丢弃。
func TestSolarEclipseArcBranchAngleMapsUnwrappedArcs(t *testing.T) {
	const twoPi = 2 * math.Pi
	if got, want := solarEclipseArcBranchAngle(0.1, 6.0), 0.1+twoPi; math.Abs(got-want) > 1e-12 {
		t.Fatalf("branch angle=%.15f, want %.15f", got, want)
	}
	for _, item := range []struct{ angle, reference float64 }{
		{0.1, 6.0}, {6.2, 0.05}, {0, 0}, {3, twoPi + 0.5}, {twoPi - 0.1, 0.2}, {0.0, twoPi - 0.05},
	} {
		got := solarEclipseArcBranchAngle(item.angle, item.reference)
		if math.Abs(math.Remainder(got-item.angle, twoPi)) > 1e-12 {
			t.Fatalf("angle %.6f mapped to %.6f is not on the same branch as reference %.6f",
				item.angle, got, item.reference)
		}
		if got < item.reference-math.Pi-1e-9 || got > item.reference+math.Pi+1e-9 {
			t.Fatalf("angle %.6f mapped to %.6f left the reference branch %.6f",
				item.angle, got, item.reference)
		}
	}
}

// 非 greatest 残差只允许算一次 stateAt；结果必须与"先算相位残差再单独算 stateAt"逐位一致。
func TestSolarEclipseRiseSetArcResidualMatchesPhaseResidual(t *testing.T) {
	seed := JDCalc(2024, 4, 8)
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	solver = solver.withLocalEphemeris()
	result := solarEclipse(seed, SolarEclipseModelNASABulletinSplitK)
	for offset := -3.0; offset <= 3.0; offset += 0.37 {
		jd := result.GreatestEclipse + offset/24.0
		evaluation := solver.magnitudeEvaluationAt(jd)
		for _, point := range [][2]float64{{10, 40}, {-70, -20}, {120, 65}, {0, 0}} {
			for _, greatest := range []bool{true, false} {
				got, gotOK := solarEclipseRiseSetArcResidualAt(evaluation, point[0], point[1], greatest)
				phase, phaseOK := solarEclipseRiseSetPhaseResidual(evaluation, point[0], point[1], greatest)
				state := evaluation.center.stateAt(point[0]*rad, point[1]*rad, 0)
				if gotOK != phaseOK || got[0] != phase || got[1] != state.sunAltitudeRad {
					t.Fatalf("residual(%.6f, %.3f, %.3f, greatest=%v)=%v/%v, want %v/%v",
						jd, point[0], point[1], greatest, got, gotOK,
						[2]float64{phase, state.sunAltitudeRad}, phaseOK)
				}
			}
		}
	}
}

// 事件局部插值星历与精确星历的差别必须远小于任何求解容差，这是所有
// "用插值星历迭代 + 精确星历复核"优化的前提。
func TestSolarEclipseCandidateEvaluationMatchesExact(t *testing.T) {
	seed := JDCalc(1136, 6, 1)
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	solver = solver.withLocalEphemeris()
	center := solver.newMoonJDE
	maximum := 0.0
	for step := -21600.0; step <= 21600.0; step += 300 {
		jd := center + step/86400.0
		exact := solver.magnitudeEvaluationAt(jd).center.stateAt(0.3, 0.6, 0)
		candidate := solver.magnitudeCandidateEvaluationAt(jd).center.stateAt(0.3, 0.6, 0)
		for _, delta := range []float64{
			candidate.separationRad - exact.separationRad,
			candidate.sunAltitudeRad - exact.sunAltitudeRad,
			candidate.sunRadiusRad - exact.sunRadiusRad,
			candidate.moonOuterRadiusRad - exact.moonOuterRadiusRad,
			candidate.moonInnerRadiusRad - exact.moonInnerRadiusRad,
		} {
			maximum = math.Max(maximum, math.Abs(delta))
		}
	}
	if maximum > 1e-9 {
		t.Fatalf("candidate ephemeris deviates by %.3e rad, above the 1e-9 bound", maximum)
	}
}

// 中心相时长的插值求解必须与完整站心解一致到远小于目录精度的量级。
func TestSolarEclipseCentralDurationMatchesExactLocalSolve(t *testing.T) {
	for _, date := range [][3]int{{2024, 4, 8}, {2009, 7, 22}, {2017, 8, 21}, {2020, 6, 21}, {1136, 6, 1}} {
		seed := JDCalc(date[0], date[1], float64(date[2]))
		result := SolarEclipseNASABulletinSplitK(seed)
		solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
		solver = solver.withLocalEphemeris()
		fast := solver.centralPhaseDurationDaysAt(
			result.GreatestEclipse, result.GreatestLongitude, result.GreatestLatitude,
		)
		local := LocalSolarEclipseNASABulletinSplitK(
			result.GreatestEclipse, result.GreatestLongitude, result.GreatestLatitude, 0,
		)
		exact := 0.0
		if local.CentralStart > 0 && local.CentralEnd > local.CentralStart {
			exact = local.CentralEnd - local.CentralStart
		}
		// 两个解各自只收敛到接触根容差，插值星历本身的贡献在 1e-12 日量级。
		if math.Abs(fast-exact) > 3*localSolarEclipseContactTolerance {
			t.Fatalf("%04d-%02d-%02d interpolated duration=%.12f days, exact=%.12f",
				date[0], date[1], date[2], fast, exact)
		}
		if math.Abs(fast-result.CentralDurationDays) > 3*localSolarEclipseContactTolerance {
			t.Fatalf("%04d-%02d-%02d published duration=%.12f, solver=%.12f",
				date[0], date[1], date[2], result.CentralDurationDays, fast)
		}
	}
}

// 路径与升落层必须与单时刻层使用同一个 ΔT：两层混用时食甚点会沿经度错开
// 0.4651·|ΔΔT|·cosφ 千米。
func TestSolarEclipsePathHonoursDeltaTOverride(t *testing.T) {
	const override = 200.0
	seed := JDCalc(2024, 4, 8)
	options := SolarEclipsePathOptions{StepDays: 2.0 / 1440.0, DeltaTSeconds: override}
	path := SolarEclipseCentralPathNASABulletinSplitK(seed, options)
	defaultPath := SolarEclipseCentralPathNASABulletinSplitK(seed, SolarEclipsePathOptions{StepDays: 2.0 / 1440.0})
	shift := math.Abs(path.Eclipse.GreatestLongitude - defaultPath.Eclipse.GreatestLongitude)
	if shift < 0.3 || shift > 0.9 {
		t.Fatalf("ΔT override moved the greatest point by %.4f degrees, want about 0.55", shift)
	}
	aligned, ok := SolarEclipseShadowAtJDE(path.Eclipse.GreatestEclipse, SolarEclipseShadowSolverOptions{
		DeltaTSeconds: override, Kind: SolarEclipseShadowUmbra, BoundaryPoints: 360,
	})
	if !ok || len(aligned.Boundaries) == 0 {
		t.Fatal("no umbral footprint at the greatest eclipse")
	}
	if distance := solarEclipseRingCenterDistanceKM(aligned.Boundaries, path.Greatest); distance > 5 {
		t.Fatalf("path and instantaneous umbra differ by %.2f km under one ΔT", distance)
	}
	mismatched, ok := SolarEclipseShadowAtJDE(path.Eclipse.GreatestEclipse, SolarEclipseShadowSolverOptions{
		Kind: SolarEclipseShadowUmbra, BoundaryPoints: 360,
	})
	if !ok || len(mismatched.Boundaries) == 0 {
		t.Fatal("no umbral footprint for the mismatched ΔT control")
	}
	if distance := solarEclipseRingCenterDistanceKM(mismatched.Boundaries, path.Greatest); distance < 20 {
		t.Fatalf("control distance %.2f km is too small to prove the override is honoured", distance)
	}
}

// solarEclipseRingCenterDistanceKM 取足迹各分段的平均位置到参考点的大圆距离。
func solarEclipseRingCenterDistanceKM(
	boundaries [][]SolarEclipsePathPoint,
	reference SolarEclipsePathPoint,
) float64 {
	points := 0
	longitude, latitude := 0.0, 0.0
	base := 0.0
	for _, boundary := range boundaries {
		for _, point := range boundary {
			if points == 0 {
				base = point.Longitude
			}
			longitude += base + math.Remainder(point.Longitude-base, 360)
			latitude += point.Latitude
			points++
		}
	}
	if points == 0 {
		return math.Inf(1)
	}
	return solarEclipsePathDistanceKM(SolarEclipsePathPoint{
		Longitude: geodata.NormalizeLongitude(longitude / float64(points)),
		Latitude:  latitude / float64(points),
	}, reference)
}

// 限界线必须来自与宽度同一份配对横截面；这里用独立的两次求解复算一遍作为对照。
func TestSolarEclipseLimitPairsMatchIndependentSolve(t *testing.T) {
	seed := JDCalc(2024, 4, 8)
	path := SolarEclipseCentralPathNASABulletinSplitK(seed, SolarEclipsePathOptions{StepDays: 2.0 / 1440.0})
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	referenceNorth, _ := solver.centralPathLimits(path.CenterLine)
	if len(referenceNorth) == 0 {
		t.Fatal("the independent limit solve returned nothing")
	}
	northern, southern, paired := solver.centralPathLimitPairs(path.CenterLine)
	// 限界线与宽度共用同一份配对结果的前提：同一输入的配对必须逐位可复现。
	againNorth, againSouth, againPaired := solver.centralPathLimitPairs(path.CenterLine)
	for index := range northern {
		if northern[index] != againNorth[index] || southern[index] != againSouth[index] || paired[index] != againPaired[index] {
			t.Fatalf("paired cross-section %d is not reproducible", index)
		}
	}
	// 端点由接触对单独恢复宽度（setSolarEclipseCentralContactWidths），只校验内部样本。
	for index := 1; index < len(path.CenterLine)-1; index++ {
		point := path.CenterLine[index]
		want := 0.0
		if paired[index] {
			want = solarEclipsePathDistanceKM(northern[index], southern[index])
			if want <= 0 || want > solarEclipsePathMaxPossibleWidthKM {
				want = 0
			}
		}
		if point.WidthKM != want {
			t.Fatalf("width at sample %d = %.6f km, paired cross-section gives %.6f km",
				index, point.WidthKM, want)
		}
	}
}

// 事件级状态缓存必须有界，且 ΔT 世代变化后旧条目必须视为未命中。
func TestSolarEclipseLocalStateContextCacheIsBounded(t *testing.T) {
	seed := JDCalc(2024, 4, 8)
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	base := solver.newMoonJDE
	for index := 0; index < 3*solarEclipseBesselGeometryCacheMaximumEntries; index++ {
		solver.localStateContextAt(base + float64(index)*0.01)
	}
	if len(solver.localStateContextCache) > solarEclipseBesselGeometryCacheMaximumEntries {
		t.Fatalf("local state cache holds %d entries, above the %d bound",
			len(solver.localStateContextCache), solarEclipseBesselGeometryCacheMaximumEntries)
	}
	jd := base + 0.25
	before := solver.localStateContextAt(jd)
	previous := GetDeltaTFn()
	defer SetDeltaTFn(previous)
	SetDeltaTFn(func(float64, bool) float64 { return 200 })
	after := solver.localStateContextAt(jd)
	if after.gst == before.gst {
		t.Fatalf("ΔT generation did not invalidate the cached context: gst %.12f", before.gst)
	}
	if after.generation != deltaTGenerationValue() {
		t.Fatalf("cached context generation=%d, want %d", after.generation, deltaTGenerationValue())
	}
}

// 升落层段数上限是质量门槛：超过就返回截断结果，调用方必须能通过降级标记看到它。
func TestSolarEclipseRiseSetTopologyDegradedContract(t *testing.T) {
	segments := make([][]SolarEclipsePathPoint, 17)
	for index := range segments {
		segments[index] = []SolarEclipsePathPoint{
			{JDE: float64(index) + 1}, {JDE: float64(index) + 1.5},
		}
	}
	overSegmented := []SolarEclipseRiseSetCurve{{
		Phase: RiseSetPhaseStart, Direction: RiseSetDirectionRise, Segments: segments,
	}}
	if solarEclipseRiseSetCurveTopologyComplete(overSegmented) {
		t.Fatal("a 17-segment curve must not pass the topology audit")
	}
	seed := JDCalc(2024, 4, 8)
	result := SolarEclipsePartialFootprintsNASABulletinSplitK(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
	})
	if len(result.RiseSetCurves) == 0 {
		t.Fatal("no rise/set curves were exported")
	}
	if want := !solarEclipseRiseSetCurveTopologyComplete(result.RiseSetCurves); result.RiseSetTopologyDegraded != want {
		t.Fatalf("degraded flag=%v, topology audit says %v", result.RiseSetTopologyDegraded, want)
	}
	disabled := SolarEclipsePartialFootprintsNASABulletinSplitK(seed, SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, DisableRiseSetCurves: true,
	})
	if disabled.RiseSetTopologyDegraded || len(disabled.RiseSetCurves) != 0 {
		t.Fatalf("disabled rise/set curves must not report degradation: flag=%v curves=%d",
			disabled.RiseSetTopologyDegraded, len(disabled.RiseSetCurves))
	}
}

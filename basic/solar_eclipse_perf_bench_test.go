package basic

import "testing"

// 性能基线：逐项隔离 §1.4 P2 热点，改动前后跑同一份基准做对照。

func solarEclipseBenchSeed(year, month, day int) float64 {
	return JDCalc(year, month, float64(day))
}

func solarEclipseBenchCenterLine(b *testing.B, seed float64) (SolarEclipsePathResult, float64, SolarEclipseRadiusModel) {
	b.Helper()
	model := SolarEclipseModelNASABulletinSplitK
	result := solarEclipse(seed, model)
	newMoonJDE := CalcMoonSHByJDE(seed, 0)
	solver := newSolarEclipseSolver(newMoonJDE, model)
	samplingOptions := SolarEclipsePathOptions{StepDays: 2.0 / 1440.0}
	samplingOptions.TargetSpacingKM = 0
	centerLine, _ := solver.centralPathPoints(
		result.CentralBeginOnEarth, result.CentralEndOnEarth, result.GreatestEclipse, samplingOptions,
	)
	centerLine = solver.attachCentralAxisContactPoints(
		centerLine, result.CentralBeginOnEarth, result.CentralEndOnEarth,
	)
	return SolarEclipsePathResult{CenterLine: centerLine}, newMoonJDE, model
}

func BenchmarkSolarEclipseCentralPath20240408(b *testing.B) {
	seed := solarEclipseBenchSeed(2024, 4, 8)
	options := SolarEclipsePathOptions{StepDays: 2.0 / 1440.0, TargetSpacingKM: 150}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		SolarEclipseCentralPathNASABulletinSplitK(seed, options)
	}
}

func BenchmarkSolarEclipseMaxCentralDuration20240408(b *testing.B) {
	path, newMoonJDE, model := solarEclipseBenchCenterLine(b, solarEclipseBenchSeed(2024, 4, 8))
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		solver := newSolarEclipseSolver(newMoonJDE, model)
		solver.centralPathMaxCentralDuration(path.CenterLine)
	}
}

func BenchmarkSolarEclipseLimitPairs20240408(b *testing.B) {
	path, newMoonJDE, model := solarEclipseBenchCenterLine(b, solarEclipseBenchSeed(2024, 4, 8))
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		solver := newSolarEclipseSolver(newMoonJDE, model)
		solver.centralPathLimitPairs(path.CenterLine)
	}
}

func BenchmarkSolarEclipseGlobal20240408(b *testing.B) {
	seed := solarEclipseBenchSeed(2024, 4, 8)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		SolarEclipseNASABulletinSplitK(seed)
	}
}

func BenchmarkSolarEclipseMagnitudeEvaluation(b *testing.B) {
	seed := solarEclipseBenchSeed(1136, 6, 1)
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	solver = solver.withLocalEphemeris()
	jd := solver.newMoonJDE
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		solver.magnitudeEvaluationAt(jd)
	}
}

func BenchmarkSolarEclipsePartialFootprints11360601(b *testing.B) {
	seed := solarEclipseBenchSeed(1136, 6, 1)
	options := SolarEclipsePartialFootprintOptions{
		StepDays: 2.0 / 1440.0, BoundaryPoints: 96, RiseSetStepDays: 2.0 / 1440.0,
	}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		SolarEclipsePartialFootprintsNASABulletinSplitK(seed, options)
	}
}

func BenchmarkSolarEclipseRiseSetArcResidual(b *testing.B) {
	seed := solarEclipseBenchSeed(2024, 4, 8)
	solver := newSolarEclipseSolver(CalcMoonSHByJDE(seed, 0), SolarEclipseModelNASABulletinSplitK)
	solver = solver.withLocalEphemeris()
	result := solarEclipse(seed, SolarEclipseModelNASABulletinSplitK)
	evaluation := solver.magnitudeEvaluationAt(result.GreatestEclipse)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		solarEclipseRiseSetArcResidualAt(evaluation, 10, 40, false)
	}
}

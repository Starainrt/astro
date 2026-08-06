package basic

import (
	"math"
	"sort"
)

const (
	solarEclipsePathDefaultStepDays   = 1.0 / 1440.0
	solarEclipsePathMinStepDays       = 1.0 / 86400.0
	solarEclipsePathMaxSampleCount    = 30000
	solarEclipsePathMaxAdaptiveDepth  = 20
	solarEclipsePathVelocityStepDays  = 1.0 / 1440.0
	solarEclipsePathDuplicateTimeDays = 1e-10

	solarEclipsePartialFootprintDefaultStepDays       = 5.0 / 1440.0
	solarEclipsePartialFootprintDefaultBoundaryPoints = 180
	solarEclipsePartialFootprintMinBoundaryPoints     = 12
	solarEclipsePartialFootprintMaxBoundaryPoints     = 1440
	solarEclipsePartialFootprintPointTolerance        = 1e-12
	solarEclipsePartialFootprintIterationLimit        = 10
	solarEclipsePartialFootprintTransitionIterations  = 48
	solarEclipseShadowContactSearchStepDays           = 10.0 / 1440.0
	solarEclipseShadowContactSearchSpanDays           = 0.75
	solarEclipseShadowContactToleranceDays            = 1e-9
)

type solarEclipseShadowKind uint8

const (
	solarEclipsePenumbralShadow solarEclipseShadowKind = iota
	solarEclipseCentralShadow
)

// SolarEclipsePathOptions 控制日食中心路径采样。
// SolarEclipsePathOptions controls central solar eclipse path sampling.
type SolarEclipsePathOptions struct {
	// StepDays 是基础时间采样步长，单位为日；<=0 时使用 1 分钟。
	// StepDays is the base time step in days; values <= 0 use one minute.
	StepDays float64
	// TargetSpacingKM 是相邻中心线点的最大目标地表距离；<=0 时不按距离加密。
	// TargetSpacingKM is the target maximum ground spacing between centerline points; values <= 0 disable spacing refinement.
	TargetSpacingKM float64
}

// SolarEclipsePathPoint 表示日食路径上的一个地理点。
// SolarEclipsePathPoint is one geographic point on a solar eclipse path.
type SolarEclipsePathPoint struct {
	// JDE 是力学时儒略日, TT Julian ephemeris day.
	JDE float64
	// Longitude 经度，东正西负, longitude in degrees, east positive.
	Longitude float64
	// Latitude 纬度，北正南负, latitude in degrees, north positive.
	Latitude float64
	// SunAltitude 太阳高度角，单位度, Sun altitude in degrees.
	SunAltitude float64
	// WidthKM 中心食带宽度，单位千米；仅中心线点有意义。
	// WidthKM is the central path width in kilometers; meaningful for centerline points.
	WidthKM float64
}

// SolarEclipsePathResult 表示一次中心日食的路径数据。
// SolarEclipsePathResult contains central solar eclipse path data.
type SolarEclipsePathResult struct {
	// Eclipse 是对应的全局日食结果, related global solar eclipse result.
	Eclipse SolarEclipseResult
	// Greatest 是食甚点/最佳观测点, greatest eclipse point.
	Greatest SolarEclipsePathPoint
	// CenterLine 是中心线, central line.
	CenterLine []SolarEclipsePathPoint
	// NorthernLimit 是中心食带北界近似线, approximate northern limit of the central path.
	NorthernLimit []SolarEclipsePathPoint
	// SouthernLimit 是中心食带南界近似线, approximate southern limit of the central path.
	SouthernLimit []SolarEclipsePathPoint
	// StepDays 是实际采用的基础时间采样步长，单位为日。
	// StepDays is the effective base time step in days.
	StepDays float64
	// TargetSpacingKM 是实际采用的目标空间采样距离，单位千米。
	// TargetSpacingKM is the effective target spacing in kilometers.
	TargetSpacingKM float64
}

// SolarEclipsePartialFootprintOptions 控制日食偏食半影足迹采样。
// SolarEclipsePartialFootprintOptions controls solar eclipse penumbral footprint sampling.
type SolarEclipsePartialFootprintOptions struct {
	// StepDays 是基础时间采样步长，单位为日；<=0 时使用 5 分钟。
	// StepDays is the base time step in days; values <= 0 use five minutes.
	StepDays float64
	// BoundaryPoints 是每个瞬时半影边界的角向采样点数；<=0 时使用 180。
	// BoundaryPoints is the angular sample count for each instantaneous penumbral boundary; values <= 0 use 180.
	BoundaryPoints int
	// CentralShadowStepDays 是本影/反本影瞬时足迹的时间步长，单位为日；<=0 时不计算。
	// CentralShadowStepDays is the umbral/antumbral footprint step in days; values <= 0 disable it.
	CentralShadowStepDays float64
}

// SolarEclipsePartialAreaOptions 是 SolarEclipsePartialFootprintOptions 的兼容别名。
// SolarEclipsePartialAreaOptions is a compatibility alias for SolarEclipsePartialFootprintOptions.
type SolarEclipsePartialAreaOptions = SolarEclipsePartialFootprintOptions

// SolarEclipsePartialFootprint 表示某一时刻的半影足迹边界。
// SolarEclipsePartialFootprint is the penumbral footprint boundary at one instant.
type SolarEclipsePartialFootprint struct {
	// JDE 是力学时儒略日, TT Julian ephemeris day.
	JDE float64
	// Boundaries 是半影边界分段；反经线或无效投影会拆成多段。
	// Boundaries are segmented penumbral boundary polylines, split at invalid projections or the antimeridian.
	Boundaries [][]SolarEclipsePathPoint
	// Closed 表示 Boundaries 是否构成一个闭合边界。
	// Closed indicates whether Boundaries form one closed boundary.
	Closed bool
}

// SolarEclipsePartialFootprintsResult 表示一次日食的偏食半影足迹序列。
// SolarEclipsePartialFootprintsResult contains penumbral footprint samples for a solar eclipse.
type SolarEclipsePartialFootprintsResult struct {
	// Eclipse 是对应的全局日食结果, related global solar eclipse result.
	Eclipse SolarEclipseResult
	// Footprints 是按时间采样的瞬时半影足迹, sampled instantaneous penumbral footprints.
	Footprints []SolarEclipsePartialFootprint
	// CentralShadowFootprints 是按时间采样的本影/反本影足迹。
	// CentralShadowFootprints are sampled umbral/antumbral footprints.
	CentralShadowFootprints []SolarEclipsePartialFootprint
	// P1-P4 是半影与地球的外切/内切接触点；不存在的内切点保持零值。
	// P1-P4 are external/internal penumbral contacts; absent internal contacts remain zero.
	P1 SolarEclipsePathPoint
	P2 SolarEclipsePathPoint
	P3 SolarEclipsePathPoint
	P4 SolarEclipsePathPoint
	// U1-U4 是本影/反本影与地球的外切/内切接触点；不存在时保持零值。
	// U1-U4 are external/internal umbral/antumbral contacts; absent contacts remain zero.
	U1 SolarEclipsePathPoint
	U2 SolarEclipsePathPoint
	U3 SolarEclipsePathPoint
	U4 SolarEclipsePathPoint
	// StepDays 是实际采用的基础时间采样步长，单位为日。
	// StepDays is the effective base time step in days.
	StepDays float64
	// BoundaryPoints 是实际采用的边界角向采样点数。
	// BoundaryPoints is the effective angular sample count for each boundary.
	BoundaryPoints int
	// CentralShadowStepDays 是本影/反本影足迹的实际采样步长；0 表示未计算。
	// CentralShadowStepDays is the effective umbral/antumbral footprint step; zero means disabled.
	CentralShadowStepDays float64
}

// SolarEclipsePartialAreaResult 是 SolarEclipsePartialFootprintsResult 的兼容别名。
// SolarEclipsePartialAreaResult is a compatibility alias for SolarEclipsePartialFootprintsResult.
type SolarEclipsePartialAreaResult = SolarEclipsePartialFootprintsResult

// SolarEclipseCentralPath 计算给定近朔时刻附近的日食中心路径，默认使用 NASA bulletin Split-K 模型。
// SolarEclipseCentralPath computes the central path near the given new-moon seed, using NASA bulletin Split-K by default.
func SolarEclipseCentralPath(seedJDE float64, options SolarEclipsePathOptions) SolarEclipsePathResult {
	return SolarEclipseCentralPathNASABulletinSplitK(seedJDE, options)
}

// SolarEclipseCentralPathIAUSingleK 计算日食中心路径，使用 IAU Single-K 模型。
// SolarEclipseCentralPathIAUSingleK computes the central path with the IAU Single-K model.
func SolarEclipseCentralPathIAUSingleK(seedJDE float64, options SolarEclipsePathOptions) SolarEclipsePathResult {
	return solarEclipseCentralPath(seedJDE, SolarEclipseModelIAUSingleK, options)
}

// SolarEclipseCentralPathNASABulletinSplitK 计算日食中心路径，使用 NASA bulletin Split-K 模型。
// SolarEclipseCentralPathNASABulletinSplitK computes the central path with the NASA bulletin Split-K model.
func SolarEclipseCentralPathNASABulletinSplitK(seedJDE float64, options SolarEclipsePathOptions) SolarEclipsePathResult {
	return solarEclipseCentralPath(seedJDE, SolarEclipseModelNASABulletinSplitK, options)
}

// SolarEclipsePartialFootprints 计算给定近朔时刻附近的日食偏食半影足迹序列，默认使用 NASA bulletin Split-K 模型。
// SolarEclipsePartialFootprints computes penumbral footprint samples near the given new-moon seed, using NASA bulletin Split-K by default.
func SolarEclipsePartialFootprints(seedJDE float64, options SolarEclipsePartialFootprintOptions) SolarEclipsePartialFootprintsResult {
	return SolarEclipsePartialFootprintsNASABulletinSplitK(seedJDE, options)
}

// SolarEclipsePartialFootprintsIAUSingleK 计算日食偏食半影足迹序列，使用 IAU Single-K 模型。
// SolarEclipsePartialFootprintsIAUSingleK computes penumbral footprint samples with the IAU Single-K model.
func SolarEclipsePartialFootprintsIAUSingleK(seedJDE float64, options SolarEclipsePartialFootprintOptions) SolarEclipsePartialFootprintsResult {
	return solarEclipsePartialFootprints(seedJDE, SolarEclipseModelIAUSingleK, options)
}

// SolarEclipsePartialFootprintsNASABulletinSplitK 计算日食偏食半影足迹序列，使用 NASA bulletin Split-K 模型。
// SolarEclipsePartialFootprintsNASABulletinSplitK computes penumbral footprint samples with the NASA bulletin Split-K model.
func SolarEclipsePartialFootprintsNASABulletinSplitK(seedJDE float64, options SolarEclipsePartialFootprintOptions) SolarEclipsePartialFootprintsResult {
	return solarEclipsePartialFootprints(seedJDE, SolarEclipseModelNASABulletinSplitK, options)
}

// SolarEclipsePartialArea 计算日食偏食半影足迹序列，是 SolarEclipsePartialFootprints 的兼容包装。
// SolarEclipsePartialArea computes penumbral footprint samples and is a compatibility wrapper for SolarEclipsePartialFootprints.
func SolarEclipsePartialArea(seedJDE float64, options SolarEclipsePartialAreaOptions) SolarEclipsePartialAreaResult {
	return SolarEclipsePartialFootprints(seedJDE, options)
}

// SolarEclipsePartialAreaIAUSingleK 计算日食偏食半影足迹序列，是 SolarEclipsePartialFootprintsIAUSingleK 的兼容包装。
// SolarEclipsePartialAreaIAUSingleK is a compatibility wrapper for SolarEclipsePartialFootprintsIAUSingleK.
func SolarEclipsePartialAreaIAUSingleK(seedJDE float64, options SolarEclipsePartialAreaOptions) SolarEclipsePartialAreaResult {
	return SolarEclipsePartialFootprintsIAUSingleK(seedJDE, options)
}

// SolarEclipsePartialAreaNASABulletinSplitK 计算日食偏食半影足迹序列，是 SolarEclipsePartialFootprintsNASABulletinSplitK 的兼容包装。
// SolarEclipsePartialAreaNASABulletinSplitK is a compatibility wrapper for SolarEclipsePartialFootprintsNASABulletinSplitK.
func SolarEclipsePartialAreaNASABulletinSplitK(seedJDE float64, options SolarEclipsePartialAreaOptions) SolarEclipsePartialAreaResult {
	return SolarEclipsePartialFootprintsNASABulletinSplitK(seedJDE, options)
}

func solarEclipseCentralPath(seedJDE float64, model SolarEclipseRadiusModel, options SolarEclipsePathOptions) SolarEclipsePathResult {
	options = normalizeSolarEclipsePathOptions(options)
	result := solarEclipse(seedJDE, model)
	path := SolarEclipsePathResult{
		Eclipse:         result,
		StepDays:        options.StepDays,
		TargetSpacingKM: options.TargetSpacingKM,
	}
	if !result.HasCentral {
		return path
	}

	newMoonJDE := CalcMoonSHByJDE(seedJDE, 0)
	solver := newSolarEclipseSolver(newMoonJDE, model)
	greatest, ok := solver.centralPathPointAt(result.GreatestEclipse)
	if !ok {
		greatest = SolarEclipsePathPoint{
			JDE:       result.GreatestEclipse,
			Longitude: result.GreatestLongitude,
			Latitude:  result.GreatestLatitude,
			WidthKM:   result.PathWidthKM,
			SunAltitude: solarEclipseSunAltitudeAtGreatest(
				result.GreatestEclipse,
				result.GreatestLongitude,
				result.GreatestLatitude,
				solver.besselAxisAt(result.GreatestEclipse).gst,
			) / rad,
		}
	}
	greatest.Longitude = result.GreatestLongitude
	greatest.Latitude = result.GreatestLatitude
	greatest.WidthKM = result.PathWidthKM
	path.Greatest = greatest

	centerLine, stepDays := solver.centralPathPoints(
		result.CentralBeginOnEarth,
		result.CentralEndOnEarth,
		result.GreatestEclipse,
		options,
	)
	path.StepDays = stepDays
	path.CenterLine = centerLine
	path.NorthernLimit, path.SouthernLimit = solver.centralPathLimits(centerLine)
	return path
}

func normalizeSolarEclipsePathOptions(options SolarEclipsePathOptions) SolarEclipsePathOptions {
	if options.StepDays <= 0 || math.IsNaN(options.StepDays) || math.IsInf(options.StepDays, 0) {
		options.StepDays = solarEclipsePathDefaultStepDays
	}
	if options.StepDays < solarEclipsePathMinStepDays {
		options.StepDays = solarEclipsePathMinStepDays
	}
	if options.TargetSpacingKM <= 0 || math.IsNaN(options.TargetSpacingKM) || math.IsInf(options.TargetSpacingKM, 0) {
		options.TargetSpacingKM = 0
	}
	return options
}

func solarEclipsePartialFootprints(
	seedJDE float64,
	model SolarEclipseRadiusModel,
	options SolarEclipsePartialFootprintOptions,
) SolarEclipsePartialFootprintsResult {
	options = normalizeSolarEclipsePartialFootprintOptions(options)
	result := solarEclipse(seedJDE, model)
	footprintsResult := SolarEclipsePartialFootprintsResult{
		Eclipse:               result,
		StepDays:              options.StepDays,
		BoundaryPoints:        options.BoundaryPoints,
		CentralShadowStepDays: options.CentralShadowStepDays,
	}
	if !result.HasPartial {
		return footprintsResult
	}

	newMoonJDE := CalcMoonSHByJDE(seedJDE, 0)
	solver := newSolarEclipseSolver(newMoonJDE, model)
	footprintsResult.P1, footprintsResult.P4, _ = solver.shadowContactPair(
		result.GreatestEclipse, solarEclipsePenumbralShadow, false,
	)
	footprintsResult.P2, footprintsResult.P3, _ = solver.shadowContactPair(
		result.GreatestEclipse, solarEclipsePenumbralShadow, true,
	)
	if result.Type != SolarEclipsePartial {
		footprintsResult.U1, footprintsResult.U4, _ = solver.shadowContactPair(
			result.GreatestEclipse, solarEclipseCentralShadow, false,
		)
		footprintsResult.U2, footprintsResult.U3, _ = solver.shadowContactPair(
			result.GreatestEclipse, solarEclipseCentralShadow, true,
		)
	}
	footprints, stepDays := solver.partialFootprints(
		result.PartialBeginOnEarth,
		result.PartialEndOnEarth,
		result.GreatestEclipse,
		options,
	)
	footprintsResult.StepDays = stepDays
	footprintsResult.Footprints = footprints
	if options.CentralShadowStepDays > 0 &&
		footprintsResult.U1.JDE != 0 && footprintsResult.U4.JDE != 0 {
		footprintsResult.CentralShadowFootprints, footprintsResult.CentralShadowStepDays = solver.shadowFootprints(
			footprintsResult.U1.JDE,
			footprintsResult.U4.JDE,
			result.GreatestEclipse,
			options.CentralShadowStepDays,
			options.BoundaryPoints,
			solarEclipseCentralShadow,
		)
	}
	return footprintsResult
}

func normalizeSolarEclipsePartialFootprintOptions(options SolarEclipsePartialFootprintOptions) SolarEclipsePartialFootprintOptions {
	if options.StepDays <= 0 || math.IsNaN(options.StepDays) || math.IsInf(options.StepDays, 0) {
		options.StepDays = solarEclipsePartialFootprintDefaultStepDays
	}
	if options.StepDays < solarEclipsePathMinStepDays {
		options.StepDays = solarEclipsePathMinStepDays
	}
	if options.BoundaryPoints <= 0 {
		options.BoundaryPoints = solarEclipsePartialFootprintDefaultBoundaryPoints
	}
	if options.BoundaryPoints < solarEclipsePartialFootprintMinBoundaryPoints {
		options.BoundaryPoints = solarEclipsePartialFootprintMinBoundaryPoints
	}
	if options.BoundaryPoints > solarEclipsePartialFootprintMaxBoundaryPoints {
		options.BoundaryPoints = solarEclipsePartialFootprintMaxBoundaryPoints
	}
	if options.CentralShadowStepDays <= 0 || math.IsNaN(options.CentralShadowStepDays) || math.IsInf(options.CentralShadowStepDays, 0) {
		options.CentralShadowStepDays = 0
	} else if options.CentralShadowStepDays < solarEclipsePathMinStepDays {
		options.CentralShadowStepDays = solarEclipsePathMinStepDays
	}
	return options
}

func (solver solarEclipseSolver) centralPathPoints(
	startJDE, endJDE, greatestJDE float64,
	options SolarEclipsePathOptions,
) ([]SolarEclipsePathPoint, float64) {
	if endJDE < startJDE {
		startJDE, endJDE = endJDE, startJDE
	}
	if startJDE == 0 || endJDE == 0 || endJDE <= startJDE {
		return nil, options.StepDays
	}

	stepDays := options.StepDays
	if sampleCount := int(math.Ceil((endJDE-startJDE)/stepDays)) + 1; sampleCount > solarEclipsePathMaxSampleCount {
		stepDays = (endJDE - startJDE) / float64(solarEclipsePathMaxSampleCount-1)
	}

	times := []float64{startJDE, greatestJDE, endJDE}
	for jd := startJDE + stepDays; jd < endJDE; jd += stepDays {
		times = append(times, jd)
	}
	sort.Float64s(times)
	times = uniqueSolarEclipsePathTimes(times)

	points := make([]SolarEclipsePathPoint, 0, len(times))
	for _, jd := range times {
		point, ok := solver.centralPathPointAt(jd)
		if ok {
			points = append(points, point)
		}
	}
	if options.TargetSpacingKM > 0 {
		points = solver.refineCentralPathSpacing(points, options.TargetSpacingKM)
	}
	return points, stepDays
}

func (solver solarEclipseSolver) partialFootprints(
	startJDE, endJDE, greatestJDE float64,
	options SolarEclipsePartialFootprintOptions,
) ([]SolarEclipsePartialFootprint, float64) {
	return solver.shadowFootprints(
		startJDE,
		endJDE,
		greatestJDE,
		options.StepDays,
		options.BoundaryPoints,
		solarEclipsePenumbralShadow,
	)
}

func (solver solarEclipseSolver) shadowFootprints(
	startJDE, endJDE, greatestJDE, requestedStepDays float64,
	boundaryPoints int,
	kind solarEclipseShadowKind,
) ([]SolarEclipsePartialFootprint, float64) {
	if endJDE < startJDE {
		startJDE, endJDE = endJDE, startJDE
	}
	if startJDE == 0 || endJDE == 0 || endJDE <= startJDE {
		return nil, requestedStepDays
	}

	stepDays := requestedStepDays
	if sampleCount := int(math.Ceil((endJDE-startJDE)/stepDays)) + 1; sampleCount > solarEclipsePathMaxSampleCount {
		stepDays = (endJDE - startJDE) / float64(solarEclipsePathMaxSampleCount-1)
	}

	times := []float64{startJDE, greatestJDE, endJDE}
	for jd := startJDE + stepDays; jd < endJDE; jd += stepDays {
		times = append(times, jd)
	}
	sort.Float64s(times)
	times = uniqueSolarEclipsePathTimes(times)

	footprints := make([]SolarEclipsePartialFootprint, 0, len(times))
	for _, jd := range times {
		footprint := solver.shadowFootprintAt(jd, boundaryPoints, kind)
		if len(footprint.Boundaries) > 0 {
			footprints = append(footprints, footprint)
		}
	}
	return footprints, stepDays
}

func uniqueSolarEclipsePathTimes(times []float64) []float64 {
	if len(times) < 2 {
		return times
	}

	unique := times[:1]
	for _, jd := range times[1:] {
		if math.Abs(jd-unique[len(unique)-1]) <= solarEclipsePathDuplicateTimeDays {
			continue
		}
		unique = append(unique, jd)
	}
	return unique
}

func (solver solarEclipseSolver) refineCentralPathSpacing(points []SolarEclipsePathPoint, targetSpacingKM float64) []SolarEclipsePathPoint {
	if len(points) < 2 || targetSpacingKM <= 0 {
		return points
	}

	refined := make([]SolarEclipsePathPoint, 0, len(points))
	refined = append(refined, points[0])
	for i := 1; i < len(points); i++ {
		refined = solver.appendRefinedCentralPathSegment(refined, points[i-1], points[i], targetSpacingKM, 0)
	}
	return refined
}

func (solver solarEclipseSolver) appendRefinedCentralPathSegment(
	points []SolarEclipsePathPoint,
	start, end SolarEclipsePathPoint,
	targetSpacingKM float64,
	depth int,
) []SolarEclipsePathPoint {
	if depth >= solarEclipsePathMaxAdaptiveDepth || solarEclipsePathDistanceKM(start, end) <= targetSpacingKM {
		return append(points, end)
	}

	midJDE := (start.JDE + end.JDE) / 2
	mid, ok := solver.centralPathPointAt(midJDE)
	if !ok {
		return append(points, end)
	}
	points = solver.appendRefinedCentralPathSegment(points, start, mid, targetSpacingKM, depth+1)
	return solver.appendRefinedCentralPathSegment(points, mid, end, targetSpacingKM, depth+1)
}

func (solver solarEclipseSolver) centralPathPointAt(jd float64) (SolarEclipsePathPoint, bool) {
	moon := solver.besselMoonAt(jd)
	axis := solver.besselAxisAt(jd)
	intersection := solarEclipseLineEar2(
		moon[0],
		moon[1],
		2,
		moon[0],
		moon[1],
		0,
		solarEclipseEarthPolarRatio,
		1,
		axis,
	)
	if !intersection.valid {
		return SolarEclipsePathPoint{}, false
	}

	longitude, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	sunAltitudeRad := solarEclipseSunAltitudeAtGreatest(jd, longitude, latitude, axis.gst)
	radii := solver.shadowRadiiAt(moon[2] - intersection.r2)

	widthKM := 0.0
	if math.Abs(math.Sin(sunAltitudeRad)) > 1e-12 {
		widthKM = math.Abs(2*radii.umbraRadius*solarEclipseEarthEquatorialRadiusKM) / math.Abs(math.Sin(sunAltitudeRad))
	}

	return SolarEclipsePathPoint{
		JDE:         jd,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: sunAltitudeRad / rad,
		WidthKM:     widthKM,
	}, true
}

func (solver solarEclipseSolver) centralPathLimits(centerLine []SolarEclipsePathPoint) ([]SolarEclipsePathPoint, []SolarEclipsePathPoint) {
	northern := make([]SolarEclipsePathPoint, 0, len(centerLine))
	southern := make([]SolarEclipsePathPoint, 0, len(centerLine))
	for _, center := range centerLine {
		north, south, ok := solver.centralPathLimitsAt(center)
		if !ok {
			continue
		}
		northern = append(northern, north)
		southern = append(southern, south)
	}
	return northern, southern
}

func (solver solarEclipseSolver) centralPathLimitsAt(center SolarEclipsePathPoint) (SolarEclipsePathPoint, SolarEclipsePathPoint, bool) {
	moon := solver.besselMoonAt(center.JDE)
	axis := solver.besselAxisAt(center.JDE)
	intersection := solarEclipseLineEar2(
		moon[0],
		moon[1],
		2,
		moon[0],
		moon[1],
		0,
		solarEclipseEarthPolarRatio,
		1,
		axis,
	)
	if !intersection.valid {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}

	radii := solver.shadowRadiiAt(moon[2] - intersection.r2)
	radius := radii.absUmbraRadius
	if radius <= 0 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}

	vx, vy, speed := solver.besselVelocityXYAt(center.JDE)
	if speed <= 0 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	perpX := -vy / speed
	perpY := vx / speed

	first, okFirst := solarEclipsePathPointFromBesselXY(center.JDE, moon[0]+radius*perpX, moon[1]+radius*perpY, axis)
	second, okSecond := solarEclipsePathPointFromBesselXY(center.JDE, moon[0]-radius*perpX, moon[1]-radius*perpY, axis)
	if !okFirst || !okSecond {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	first.WidthKM = center.WidthKM
	second.WidthKM = center.WidthKM
	if first.Latitude >= second.Latitude {
		return first, second, true
	}
	return second, first, true
}

func (solver solarEclipseSolver) besselVelocityXYAt(jd float64) (float64, float64, float64) {
	before := solver.besselMoonAt(jd - solarEclipsePathVelocityStepDays)
	after := solver.besselMoonAt(jd + solarEclipsePathVelocityStepDays)
	vx := (after[0] - before[0]) / (2 * solarEclipsePathVelocityStepDays)
	vy := (after[1] - before[1]) / (2 * solarEclipsePathVelocityStepDays)
	return vx, vy, math.Hypot(vx, vy)
}

func solarEclipsePathPointFromBesselXY(jd, x, y float64, axis solarEclipseAxis) (SolarEclipsePathPoint, bool) {
	longitude, latitude, ok := solarEclipseBesselXYToGeodetic(x, y, axis, true)
	if !ok {
		return SolarEclipsePathPoint{}, false
	}
	sunAltitudeRad := solarEclipseSunAltitudeAtGreatest(jd, longitude, latitude, axis.gst)
	return SolarEclipsePathPoint{
		JDE:         jd,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: sunAltitudeRad / rad,
	}, true
}

func (solver solarEclipseSolver) shadowContactPair(
	greatestJDE float64,
	kind solarEclipseShadowKind,
	internal bool,
) (SolarEclipsePathPoint, SolarEclipsePathPoint, bool) {
	middleResidual, ok := solver.shadowContactResidual(greatestJDE, kind, internal)
	if !ok || middleResidual > 0 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	firstJDE, firstOK := solver.shadowContactRoot(greatestJDE, -1, middleResidual, kind, internal)
	lastJDE, lastOK := solver.shadowContactRoot(greatestJDE, 1, middleResidual, kind, internal)
	if !firstOK || !lastOK {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	first, firstOK := solver.shadowContactPointAt(firstJDE, kind)
	last, lastOK := solver.shadowContactPointAt(lastJDE, kind)
	if !firstOK || !lastOK {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	return first, last, true
}

func (solver solarEclipseSolver) shadowContactRoot(
	greatestJDE float64,
	direction float64,
	middleResidual float64,
	kind solarEclipseShadowKind,
	internal bool,
) (float64, bool) {
	insideJDE := greatestJDE
	insideResidual := middleResidual
	for span := solarEclipseShadowContactSearchStepDays; span <= solarEclipseShadowContactSearchSpanDays; span += solarEclipseShadowContactSearchStepDays {
		outsideJDE := greatestJDE + direction*span
		outsideResidual, ok := solver.shadowContactResidual(outsideJDE, kind, internal)
		if !ok {
			continue
		}
		if outsideResidual >= 0 {
			leftJDE, rightJDE := outsideJDE, insideJDE
			leftResidual, rightResidual := outsideResidual, insideResidual
			if leftJDE > rightJDE {
				leftJDE, rightJDE = rightJDE, leftJDE
				leftResidual, rightResidual = rightResidual, leftResidual
			}
			for rightJDE-leftJDE > solarEclipseShadowContactToleranceDays {
				middleJDE := (leftJDE + rightJDE) / 2
				residual, valid := solver.shadowContactResidual(middleJDE, kind, internal)
				if !valid {
					return 0, false
				}
				if (residual >= 0) == (leftResidual >= 0) {
					leftJDE, leftResidual = middleJDE, residual
				} else {
					rightJDE, rightResidual = middleJDE, residual
				}
			}
			return (leftJDE + rightJDE) / 2, true
		}
		insideJDE, insideResidual = outsideJDE, outsideResidual
	}
	return 0, false
}

func (solver solarEclipseSolver) shadowContactResidual(
	jd float64,
	kind solarEclipseShadowKind,
	internal bool,
) (float64, bool) {
	moon := solver.besselMoonAt(jd)
	distanceSquared := moon[0]*moon[0] + moon[1]*moon[1]
	if distanceSquared <= 0 {
		return 0, false
	}
	radius := solver.shadowRadiusAt(moon[2], kind)
	if radius <= 0 {
		return 0, false
	}
	earthRadius := 1 - (1/solarEclipseEarthPolarRatioSquared-1)*moon[1]*moon[1]/distanceSquared/2
	limit := earthRadius + radius
	if internal {
		limit = earthRadius - radius
	}
	if limit <= 0 {
		return 0, false
	}
	return math.Sqrt(distanceSquared) - limit, true
}

func (solver solarEclipseSolver) shadowContactPointAt(
	jd float64,
	kind solarEclipseShadowKind,
) (SolarEclipsePathPoint, bool) {
	moon := solver.besselMoonAt(jd)
	distance := math.Hypot(moon[0], moon[1])
	if distance <= 0 || solver.shadowRadiusAt(moon[2], kind) <= 0 {
		return SolarEclipsePathPoint{}, false
	}
	axis := solver.besselAxisAt(jd)
	unitX, unitY := moon[0]/distance, moon[1]/distance
	insideScale, outsideScale := 0.0, 1.1
	var intersection solarEclipseLineIntersection
	for iteration := 0; iteration < 48; iteration++ {
		scale := (insideScale + outsideScale) / 2
		candidate := solarEclipseLineEar2(
			scale*unitX, scale*unitY, 2,
			scale*unitX, scale*unitY, 0,
			solarEclipseEarthPolarRatio, 1, axis,
		)
		if candidate.valid {
			insideScale = scale
			intersection = candidate
		} else {
			outsideScale = scale
		}
	}
	if !intersection.valid {
		return SolarEclipsePathPoint{}, false
	}
	longitude, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	sunAltitudeRad := solarEclipseSunAltitudeAtGreatest(jd, longitude, latitude, axis.gst)
	return SolarEclipsePathPoint{
		JDE:         jd,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: sunAltitudeRad / rad,
	}, true
}

func (solver solarEclipseSolver) shadowRadiusAt(moonBesselZ float64, kind solarEclipseShadowKind) float64 {
	radii := solver.shadowRadiiAt(moonBesselZ)
	if kind == solarEclipseCentralShadow {
		return radii.absUmbraRadius
	}
	return radii.penumbraRadius
}

func (solver solarEclipseSolver) partialFootprintAt(jd float64, boundaryPoints int) SolarEclipsePartialFootprint {
	return solver.shadowFootprintAt(jd, boundaryPoints, solarEclipsePenumbralShadow)
}

func (solver solarEclipseSolver) shadowFootprintAt(
	jd float64,
	boundaryPoints int,
	kind solarEclipseShadowKind,
) SolarEclipsePartialFootprint {
	moon := solver.besselMoonAt(jd)
	axis := solver.besselAxisAt(jd)
	samples := make([]solarEclipsePartialBoundarySample, boundaryPoints)
	for i := range samples {
		angle := 2 * math.Pi * float64(i) / float64(boundaryPoints)
		point, ok := solver.shadowFootprintPointAt(jd, moon, axis, angle, kind)
		samples[i] = solarEclipsePartialBoundarySample{
			point: point,
			ok:    ok,
			angle: angle,
		}
	}
	samples = solver.refineShadowFootprintTransitions(jd, moon, axis, samples, kind)

	boundaries, closed := solarEclipsePartialBoundarySegments(samples)
	return SolarEclipsePartialFootprint{
		JDE:        jd,
		Boundaries: boundaries,
		Closed:     closed,
	}
}

type solarEclipsePartialBoundarySample struct {
	point SolarEclipsePathPoint
	ok    bool
	angle float64
}

func (solver solarEclipseSolver) refineShadowFootprintTransitions(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	samples []solarEclipsePartialBoundarySample,
	kind solarEclipseShadowKind,
) []solarEclipsePartialBoundarySample {
	if len(samples) < 2 {
		return samples
	}
	result := make([]solarEclipsePartialBoundarySample, 0, len(samples)+4)
	for index, sample := range samples {
		result = append(result, sample)
		next := samples[(index+1)%len(samples)]
		if sample.ok == next.ok {
			continue
		}
		nextAngle := next.angle
		if index == len(samples)-1 {
			nextAngle += 2 * math.Pi
		}
		refined := solver.refineShadowFootprintTransition(jd, moon, axis, sample, next, nextAngle, kind)
		result = append(result, refined)
	}
	return result
}

func (solver solarEclipseSolver) refineShadowFootprintTransition(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	first, second solarEclipsePartialBoundarySample,
	secondAngle float64,
	kind solarEclipseShadowKind,
) solarEclipsePartialBoundarySample {
	leftAngle := first.angle
	rightAngle := secondAngle
	leftOK := first.ok
	best := first
	if second.ok {
		best = second
		best.angle = secondAngle
	}
	for iteration := 0; iteration < solarEclipsePartialFootprintTransitionIterations; iteration++ {
		middleAngle := (leftAngle + rightAngle) / 2
		evaluationAngle := math.Mod(middleAngle, 2*math.Pi)
		point, ok := solver.shadowFootprintPointAt(jd, moon, axis, evaluationAngle, kind)
		middle := solarEclipsePartialBoundarySample{point: point, ok: ok, angle: middleAngle}
		if ok {
			best = middle
		}
		if ok == leftOK {
			leftAngle = middleAngle
		} else {
			rightAngle = middleAngle
		}
		if rightAngle-leftAngle <= solarEclipsePartialFootprintPointTolerance {
			break
		}
	}
	best.angle = math.Mod(best.angle, 2*math.Pi)
	return best
}

func (solver solarEclipseSolver) shadowFootprintPointAt(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	angle float64,
	kind solarEclipseShadowKind,
) (SolarEclipsePathPoint, bool) {
	cosAngle := math.Cos(angle)
	sinAngle := math.Sin(angle)
	radius := solver.shadowRadiusAt(moon[2], kind)
	if radius <= 0 {
		return SolarEclipsePathPoint{}, false
	}

	var intersection solarEclipseLineIntersection
	for i := 0; i < solarEclipsePartialFootprintIterationLimit; i++ {
		x := moon[0] + radius*cosAngle
		y := moon[1] + radius*sinAngle
		intersection = solarEclipseLineEar2(
			x,
			y,
			2,
			x,
			y,
			0,
			solarEclipseEarthPolarRatio,
			1,
			axis,
		)
		if !intersection.valid {
			return SolarEclipsePathPoint{}, false
		}

		nextRadius := solver.shadowRadiusAt(moon[2]-intersection.r2, kind)
		if nextRadius <= 0 {
			return SolarEclipsePathPoint{}, false
		}
		if math.Abs(nextRadius-radius) <= solarEclipsePartialFootprintPointTolerance {
			radius = nextRadius
			break
		}
		radius = nextRadius
	}

	x := moon[0] + radius*cosAngle
	y := moon[1] + radius*sinAngle
	intersection = solarEclipseLineEar2(
		x,
		y,
		2,
		x,
		y,
		0,
		solarEclipseEarthPolarRatio,
		1,
		axis,
	)
	if !intersection.valid {
		return SolarEclipsePathPoint{}, false
	}

	longitude, latitude := solarEclipseIntersectionGeodetic(intersection, axis)
	sunAltitudeRad := solarEclipseSunAltitudeAtGreatest(jd, longitude, latitude, axis.gst)
	return SolarEclipsePathPoint{
		JDE:         jd,
		Longitude:   longitude,
		Latitude:    latitude,
		SunAltitude: sunAltitudeRad / rad,
	}, true
}

func solarEclipsePartialBoundarySegments(samples []solarEclipsePartialBoundarySample) ([][]SolarEclipsePathPoint, bool) {
	segments := make([][]SolarEclipsePathPoint, 0, 2)
	var current []SolarEclipsePathPoint
	allSamplesValid := len(samples) > 0
	for _, sample := range samples {
		if !sample.ok {
			allSamplesValid = false
			segments = appendSolarEclipsePartialSegment(segments, current)
			current = nil
			continue
		}
		if len(current) > 0 && solarEclipsePathCrossesAntimeridian(current[len(current)-1], sample.point) {
			segments = appendSolarEclipsePartialSegment(segments, current)
			current = nil
		}
		current = append(current, sample.point)
	}
	segments = appendSolarEclipsePartialSegment(segments, current)
	segments = mergeSolarEclipsePartialWrapSegment(segments, samples)

	if !allSamplesValid {
		return segments, false
	}
	totalPoints := 0
	for _, segment := range segments {
		totalPoints += len(segment)
	}
	if totalPoints < 3 {
		return segments, false
	}
	if len(segments) == 1 && !solarEclipsePathCrossesAntimeridian(segments[0][len(segments[0])-1], segments[0][0]) {
		segments[0] = append(segments[0], segments[0][0])
	}
	return segments, true
}

func appendSolarEclipsePartialSegment(
	segments [][]SolarEclipsePathPoint,
	segment []SolarEclipsePathPoint,
) [][]SolarEclipsePathPoint {
	if len(segment) == 0 {
		return segments
	}
	return append(segments, segment)
}

func mergeSolarEclipsePartialWrapSegment(
	segments [][]SolarEclipsePathPoint,
	samples []solarEclipsePartialBoundarySample,
) [][]SolarEclipsePathPoint {
	if len(segments) < 2 || len(samples) == 0 || !samples[0].ok || !samples[len(samples)-1].ok {
		return segments
	}
	first := segments[0]
	last := segments[len(segments)-1]
	if solarEclipsePathCrossesAntimeridian(last[len(last)-1], first[0]) {
		return segments
	}

	merged := make([]SolarEclipsePathPoint, 0, len(last)+len(first))
	merged = append(merged, last...)
	merged = append(merged, first...)
	result := make([][]SolarEclipsePathPoint, 0, len(segments)-1)
	result = append(result, merged)
	result = append(result, segments[1:len(segments)-1]...)
	return result
}

func solarEclipsePathCrossesAntimeridian(a, b SolarEclipsePathPoint) bool {
	return math.Abs(a.Longitude-b.Longitude) > 180
}

func solarEclipsePathDistanceKM(a, b SolarEclipsePathPoint) float64 {
	lat1 := a.Latitude * rad
	lat2 := b.Latitude * rad
	dlat := lat2 - lat1
	dlon := solarEclipseNormalizeSignedRadians((b.Longitude - a.Longitude) * rad)
	h := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	if h > 1 {
		h = 1
	}
	return 2 * solarEclipseEarthEquatorialRadiusKM * math.Asin(math.Sqrt(h))
}

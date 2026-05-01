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
	// StepDays 是实际采用的基础时间采样步长，单位为日。
	// StepDays is the effective base time step in days.
	StepDays float64
	// BoundaryPoints 是实际采用的边界角向采样点数。
	// BoundaryPoints is the effective angular sample count for each boundary.
	BoundaryPoints int
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
		Eclipse:        result,
		StepDays:       options.StepDays,
		BoundaryPoints: options.BoundaryPoints,
	}
	if !result.HasPartial {
		return footprintsResult
	}

	newMoonJDE := CalcMoonSHByJDE(seedJDE, 0)
	solver := newSolarEclipseSolver(newMoonJDE, model)
	footprints, stepDays := solver.partialFootprints(
		result.PartialBeginOnEarth,
		result.PartialEndOnEarth,
		result.GreatestEclipse,
		options,
	)
	footprintsResult.StepDays = stepDays
	footprintsResult.Footprints = footprints
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

	footprints := make([]SolarEclipsePartialFootprint, 0, len(times))
	for _, jd := range times {
		footprint := solver.partialFootprintAt(jd, options.BoundaryPoints)
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

func (solver solarEclipseSolver) partialFootprintAt(jd float64, boundaryPoints int) SolarEclipsePartialFootprint {
	moon := solver.besselMoonAt(jd)
	axis := solver.besselAxisAt(jd)
	samples := make([]solarEclipsePartialBoundarySample, boundaryPoints)
	for i := range samples {
		angle := 2 * math.Pi * float64(i) / float64(boundaryPoints)
		point, ok := solver.partialFootprintPointAt(jd, moon, axis, angle)
		samples[i] = solarEclipsePartialBoundarySample{
			point: point,
			ok:    ok,
		}
	}

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
}

func (solver solarEclipseSolver) partialFootprintPointAt(
	jd float64,
	moon [3]float64,
	axis solarEclipseAxis,
	angle float64,
) (SolarEclipsePathPoint, bool) {
	cosAngle := math.Cos(angle)
	sinAngle := math.Sin(angle)
	radius := solver.shadowRadiiAt(moon[2]).penumbraRadius
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

		nextRadius := solver.shadowRadiiAt(moon[2] - intersection.r2).penumbraRadius
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
	for _, sample := range samples {
		if !sample.ok {
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

	if len(segments) == 1 && len(segments[0]) > 2 && !solarEclipsePathCrossesAntimeridian(segments[0][len(segments[0])-1], segments[0][0]) {
		segments[0] = append(segments[0], segments[0][0])
		return segments, true
	}
	return segments, false
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

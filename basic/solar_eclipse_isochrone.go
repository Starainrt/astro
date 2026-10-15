package basic

import (
	"math"

	"github.com/starainrt/astro/internal/geodata"
)

const (
	solarEclipseGreatestTimeContourSeedLatitudeStepDegrees  = 5.0
	solarEclipseGreatestTimeContourSeedLongitudeStepDegrees = 5.0
	solarEclipseGreatestTimeContourSeedLatitudeLimitDegrees = 85.0
	solarEclipseGreatestTimeContourArcStepDegrees           = 1.5
	solarEclipseGreatestTimeContourMinArcStepDegrees        = 0.01
	solarEclipseGreatestTimeContourMaxArcSteps              = 4000
	solarEclipseGreatestTimeContourCorrectionIterations     = 12
	solarEclipseGreatestTimeContourGradientStepDegrees      = 1e-4
	solarEclipseGreatestTimeContourLatitudeLimitDegrees     = 88.0
)

// solarEclipseGreatestTimeArc 固定一个食甚时刻后的等时线求根器。
// 时刻固定后 g = ∂(separation²)/∂t 只随经纬度变化，其零集就是该时刻的食甚等值线：
// 一个约束、两个未知量，所以结果是曲线而不是区域，延拓成本正比于曲线长度。
type solarEclipseGreatestTimeArc struct {
	evaluation solarEclipseRiseSetEvaluation
}

// sample 返回残差与中心状态；离开可见偏食域、非极小点或数值无效时 ok 为 false。
func (arc solarEclipseGreatestTimeArc) sample(longitude, latitude float64) (float64, localSolarEclipseState, bool) {
	var state localSolarEclipseState
	if latitude <= -90 || latitude >= 90 {
		return 0, state, false
	}
	lonRad, latRad := longitude*rad, latitude*rad
	before := arc.evaluation.before.stateAt(lonRad, latRad, 0)
	state = arc.evaluation.center.stateAt(lonRad, latRad, 0)
	after := arc.evaluation.after.stateAt(lonRad, latRad, 0)
	if !finite(state.separationSquared) || state.sunAltitudeRad <= 0 {
		return 0, state, false
	}
	// 角距极小值处处存在，等时线必须再要求日月盘面真的相交，否则会在无食可见的海面上画出曲线。
	if solarEclipsePartialContactGap(state) > 1e-7 {
		return 0, state, false
	}
	stepSquared := solarEclipseRiseSetDerivativeStepDays * solarEclipseRiseSetDerivativeStepDays
	// separation² 在此取极小值才算食甚；二阶导非正说明该时刻不是本地的极大食。
	if (after.separationSquared-2*state.separationSquared+before.separationSquared)/stepSquared <= 0 {
		return 0, state, false
	}
	value := (after.separationSquared - before.separationSquared) / (2 * solarEclipseRiseSetDerivativeStepDays)
	if !finite(value) {
		return 0, state, false
	}
	return value, state, true
}

func (arc solarEclipseGreatestTimeArc) residual(longitude, latitude float64) float64 {
	value, _, ok := arc.sample(longitude, latitude)
	if !ok {
		return math.NaN()
	}
	return value
}

func (arc solarEclipseGreatestTimeArc) point(longitude, latitude float64, state localSolarEclipseState) SolarEclipsePathPoint {
	return SolarEclipsePathPoint{
		JDE:         arc.evaluation.jd,
		Longitude:   geodata.NormalizeLongitude(longitude),
		Latitude:    latitude,
		SunAltitude: state.sunAltitudeRad / rad,
	}
}

// metricGradient 返回 g 对地面东向、北向角度的偏导；东向角度 = 经度差 × cos(纬度)。
func (arc solarEclipseGreatestTimeArc) metricGradient(longitude, latitude float64) (float64, float64, bool) {
	step := solarEclipseGreatestTimeContourGradientStepDegrees
	value, _, ok := arc.sample(longitude, latitude)
	if !ok {
		return 0, 0, false
	}
	cosine := math.Cos(latitude * rad)
	if cosine < 1e-6 {
		return 0, 0, false
	}
	eastValue, _, eastOK := arc.sample(longitude+step, latitude)
	westValue, _, westOK := arc.sample(longitude-step, latitude)
	northValue, _, northOK := arc.sample(longitude, latitude+step)
	southValue, _, southOK := arc.sample(longitude, latitude-step)
	longitudeDerivative, ok := greatestTimeContourDifference(value, eastValue, westValue, step, eastOK, westOK)
	if !ok {
		return 0, 0, false
	}
	latitudeDerivative, ok := greatestTimeContourDifference(value, northValue, southValue, step, northOK, southOK)
	if !ok {
		return 0, 0, false
	}
	return longitudeDerivative / cosine, latitudeDerivative, true
}

// correct 把预测点沿残差梯度投影回零集；失败说明该方向已离开等时线定义域。
func (arc solarEclipseGreatestTimeArc) correct(longitude, latitude float64) (float64, float64, localSolarEclipseState, bool) {
	var state localSolarEclipseState
	for iteration := 0; iteration < solarEclipseGreatestTimeContourCorrectionIterations; iteration++ {
		value, current, ok := arc.sample(longitude, latitude)
		if !ok {
			return 0, 0, state, false
		}
		state = current
		if math.Abs(value) <= greatestTimeContourResidualTolerance {
			return longitude, latitude, state, true
		}
		east, north, ok := arc.metricGradient(longitude, latitude)
		if !ok {
			return 0, 0, state, false
		}
		denominator := east*east + north*north
		cosine := math.Cos(latitude * rad)
		if denominator < 1e-18 || cosine < 1e-6 {
			return 0, 0, state, false
		}
		// 完整牛顿步可能一步跨出可见域；逐步二分回退，只要还有一步落在域内就继续投影。
		scale, advanced := 1.0, false
		for attempt := 0; attempt < greatestTimeContourCorrectionBacktracking; attempt++ {
			nextLongitude := longitude - scale*value*east/denominator/cosine
			nextLatitude := latitude - scale*value*north/denominator
			scale /= 2
			if nextLatitude <= -90 || nextLatitude >= 90 {
				continue
			}
			if _, _, ok := arc.sample(nextLongitude, nextLatitude); !ok {
				continue
			}
			longitude, latitude = nextLongitude, nextLatitude
			advanced = true
			break
		}
		if !advanced {
			return 0, 0, state, false
		}
	}
	value, current, ok := arc.sample(longitude, latitude)
	if !ok || math.Abs(value) > 1e-6 {
		return 0, 0, state, false
	}
	return longitude, latitude, current, true
}

// traceGreatestTimeArc 从种子沿一个方向按弧长延拓，预测点落到定义域外时步长减半。
func (solver solarEclipseSolver) traceGreatestTimeArc(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude, direction float64,
) []SolarEclipsePathPoint {
	arc := solarEclipseGreatestTimeArc{evaluation: evaluation}
	_, state, ok := arc.sample(longitude, latitude)
	if !ok {
		return nil
	}
	points := []SolarEclipsePathPoint{arc.point(longitude, latitude, state)}
	step := solarEclipseGreatestTimeContourArcStepDegrees
	previousEast, previousNorth := 0.0, 0.0
	for count := 0; count < solarEclipseGreatestTimeContourMaxArcSteps; count++ {
		east, north, ok := arc.metricGradient(longitude, latitude)
		if !ok {
			break
		}
		norm := math.Hypot(east, north)
		cosine := math.Cos(latitude * rad)
		if norm < 1e-12 || cosine < 1e-6 {
			break
		}
		tangentEast, tangentNorth := -north/norm, east/norm
		if previousEast != 0 || previousNorth != 0 {
			if tangentEast*previousEast+tangentNorth*previousNorth < 0 {
				tangentEast, tangentNorth = -tangentEast, -tangentNorth
			}
		}
		nextLongitude, nextLatitude, nextState, ok := arc.correct(
			longitude+direction*step*tangentEast/cosine,
			latitude+direction*step*tangentNorth,
		)
		if !ok {
			step /= 2
			if step < solarEclipseGreatestTimeContourMinArcStepDegrees {
				break
			}
			continue
		}
		if math.Abs(nextLatitude) > solarEclipseGreatestTimeContourLatitudeLimitDegrees {
			break
		}
		next := arc.point(nextLongitude, nextLatitude, nextState)
		distance := solarEclipsePathDistanceKM(points[len(points)-1], next)
		// 校正回到原点说明该方向已经走到支路端点，继续只会原地打转。
		if distance < 1e-9 || distance > 4*step*geodata.EarthKMPerDegree {
			break
		}
		points = append(points, next)
		longitude, latitude = nextLongitude, nextLatitude
		previousEast, previousNorth = tangentEast, tangentNorth
		step = math.Min(solarEclipseGreatestTimeContourArcStepDegrees, step*1.5)
	}
	return points
}

// greatestTimeContourSeeds 用粗扫找延拓种子；扫描只用于定位零集，不参与曲线成型。
func (solver solarEclipseSolver) greatestTimeContourSeeds(evaluation solarEclipseRiseSetEvaluation) []SolarEclipsePathPoint {
	arc := solarEclipseGreatestTimeArc{evaluation: evaluation}
	seeds := make([]SolarEclipsePathPoint, 0, 16)
	latitudeStep := solarEclipseGreatestTimeContourSeedLatitudeStepDegrees
	longitudeStep := solarEclipseGreatestTimeContourSeedLongitudeStepDegrees
	limit := solarEclipseGreatestTimeContourSeedLatitudeLimitDegrees
	for latitude := -limit; latitude <= limit; latitude += latitudeStep {
		previousLongitude := -180.0
		previousValue := arc.residual(previousLongitude, latitude)
		for longitude := previousLongitude + longitudeStep; longitude <= 180; longitude += longitudeStep {
			value := arc.residual(longitude, latitude)
			if finite(previousValue) && finite(value) && previousValue*value <= 0 {
				root, ok := greatestTimeContourBisect(arc.residual, previousLongitude, longitude, latitude, previousValue)
				if !ok {
					continue
				}
				if _, state, sampled := arc.sample(root, latitude); sampled {
					seeds = append(seeds, arc.point(root, latitude, state))
				}
			}
			previousLongitude, previousValue = longitude, value
		}
	}
	return seeds
}

func solarEclipseGreatestTimeContourCovered(segments [][]SolarEclipsePathPoint, point SolarEclipsePathPoint) bool {
	for _, segment := range segments {
		for index := 1; index < len(segment); index++ {
			if greatestTimeContourPointSegmentKM(
				point.Longitude, point.Latitude,
				segment[index-1].Longitude, segment[index-1].Latitude,
				segment[index].Longitude, segment[index].Latitude,
			) <= greatestTimeContourCoverToleranceKM {
				return true
			}
		}
	}
	return false
}

// solarEclipseGreatestTimeContourSegmentCovered 判断整条支路是否已落在已绘曲线上（同一曲线被先后延拓两次时后一条可能更长）。
func solarEclipseGreatestTimeContourSegmentCovered(segments [][]SolarEclipsePathPoint, segment []SolarEclipsePathPoint) bool {
	for _, point := range segment {
		if !solarEclipseGreatestTimeContourCovered(segments, point) {
			return false
		}
	}
	return true
}

// solarEclipseGreatestTimeContourPruneCovered 丢弃已被新支路整条覆盖的旧支路。
func solarEclipseGreatestTimeContourPruneCovered(segments [][]SolarEclipsePathPoint, added []SolarEclipsePathPoint) [][]SolarEclipsePathPoint {
	kept := segments[:0]
	for _, segment := range segments {
		if solarEclipseGreatestTimeContourSegmentCovered([][]SolarEclipsePathPoint{added}, segment) {
			continue
		}
		kept = append(kept, segment)
	}
	return kept
}

// greatestTimeContourSegments 汇总一个时刻取值上的全部等时线支路。
func (solver solarEclipseSolver) greatestTimeContourSegments(level float64) [][]SolarEclipsePathPoint {
	evaluation := solver.magnitudeEvaluationAt(level)
	seeds := solver.greatestTimeContourSeeds(evaluation)
	segments := make([][]SolarEclipsePathPoint, 0, 2)
	for _, seed := range seeds {
		if solarEclipseGreatestTimeContourCovered(segments, seed) {
			continue
		}
		forward := solver.traceGreatestTimeArc(evaluation, seed.Longitude, seed.Latitude, 1)
		backward := solver.traceGreatestTimeArc(evaluation, seed.Longitude, seed.Latitude, -1)
		segment := make([]SolarEclipsePathPoint, 0, len(forward)+len(backward))
		for index := len(backward) - 1; index >= 1; index-- {
			segment = append(segment, backward[index])
		}
		segment = append(segment, forward...)
		if len(segment) < 2 {
			continue
		}
		// 先按整条支路去重：种子检查只能拦住"较短者先画"的情况，反序时需要在这里收口。
		if solarEclipseGreatestTimeContourSegmentCovered(segments, segment) {
			continue
		}
		segments = solarEclipseGreatestTimeContourPruneCovered(segments, segment)
		segments = append(segments, segment)
	}
	return segments
}

// greatestTimeContours 计算请求时刻取值的地方食甚时刻等值线。
func (solver solarEclipseSolver) greatestTimeContours(
	startJDE, endJDE float64,
	options SolarEclipsePartialFootprintOptions,
) []SolarEclipseGreatestTimeContour {
	if len(options.GreatestTimeValues) == 0 || startJDE == 0 || endJDE == 0 || endJDE <= startJDE {
		return nil
	}
	contours := make([]SolarEclipseGreatestTimeContour, 0, len(options.GreatestTimeValues))
	for _, level := range options.GreatestTimeValues {
		if !finite(level) || level < startJDE || level > endJDE {
			continue
		}
		segments := solver.greatestTimeContourSegments(level)
		if len(segments) == 0 {
			continue
		}
		contours = append(contours, SolarEclipseGreatestTimeContour{JDE: level, Segments: segments})
	}
	return contours
}

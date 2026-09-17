package basic

import (
	"math"
	"time"
)

const (
	occultationGreatestTimeContourSeedLatitudeStepDegrees  = 5.0
	occultationGreatestTimeContourSeedLongitudeStepDegrees = 5.0
	occultationGreatestTimeContourSeedLatitudeLimitDegrees = 85.0
	occultationGreatestTimeContourArcStepDegrees           = 1.5
	occultationGreatestTimeContourMinArcStepDegrees        = 0.01
	occultationGreatestTimeContourMaxArcSteps              = 4000
	occultationGreatestTimeContourCorrectionIterations     = 12
	occultationGreatestTimeContourGradientStepDegrees      = 1e-4
	occultationGreatestTimeContourLatitudeLimitDegrees     = 88.0
)

// occultationGreatestTimeArc 固定一个掩甚时刻后的等时线求根器。
// 时刻固定后判据的时间导数只随经纬度变化，其零集就是该时刻的掩甚等值线：
// 一个约束、两个未知量，所以结果是曲线而不是区域，延拓成本正比于曲线长度。
// useContactMetric 选择判据口径：点源恒星用月面中心角距（与库内 StarOccultationInfo.Greatest 同口径），
// 有限盘面行星用外接触度量（与库内 PlanetOccultationInfo.Greatest 同口径），两者相差数秒。
type occultationGreatestTimeArc struct {
	evaluation       occultationRiseSetEvaluation
	location         *time.Location
	useContactMetric bool
}

func (arc occultationGreatestTimeArc) metric(state occultationRiseSetState) float64 {
	if arc.useContactMetric {
		return state.contactMetric
	}
	return state.separationSquared
}

// sample 返回残差与中心状态；月面在地平下、未发生接触、非极小点或数值无效时 ok 为 false。
func (arc occultationGreatestTimeArc) sample(longitude, latitude float64) (float64, occultationRiseSetState, bool) {
	var state occultationRiseSetState
	if latitude <= -90 || latitude >= 90 {
		return 0, state, false
	}
	before := arc.evaluation.before.stateAt(longitude, latitude)
	state = arc.evaluation.center.stateAt(longitude, latitude)
	after := arc.evaluation.after.stateAt(longitude, latitude)
	if !state.valid || state.moonAltitude <= 0 {
		return 0, state, false
	}
	// 角距极小值处处存在，等时线必须再要求目标盘面真的与月面接触，否则会在无掩可见的区域画出曲线。
	if state.contactMetric > 1e-7 {
		return 0, state, false
	}
	stepSquared := occultationRiseSetDerivativeStepDays * occultationRiseSetDerivativeStepDays
	// 判据在此取极小值才是掩甚；二阶导非正说明该时刻不是本地的极大掩。
	center := arc.metric(state)
	if (arc.metric(after)-2*center+arc.metric(before))/stepSquared <= 0 {
		return 0, state, false
	}
	value := (arc.metric(after) - arc.metric(before)) / (2 * occultationRiseSetDerivativeStepDays)
	if !finite(value) {
		return 0, state, false
	}
	return value, state, true
}

func (arc occultationGreatestTimeArc) residual(longitude, latitude float64) float64 {
	value, _, ok := arc.sample(longitude, latitude)
	if !ok {
		return math.NaN()
	}
	return value
}

func (arc occultationGreatestTimeArc) point(longitude, latitude float64, state occultationRiseSetState) OccultationPathPoint {
	return OccultationPathPoint{
		Time:         occultationTTToLocation(arc.evaluation.tt, arc.location),
		Longitude:    normalizeLongitude(longitude),
		Latitude:     latitude,
		MoonAltitude: state.moonAltitude,
	}
}

// metricGradient 返回 g 对地面东向、北向角度的偏导；东向角度 = 经度差 × cos(纬度)。
func (arc occultationGreatestTimeArc) metricGradient(longitude, latitude float64) (float64, float64, bool) {
	step := occultationGreatestTimeContourGradientStepDegrees
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
func (arc occultationGreatestTimeArc) correct(longitude, latitude float64) (float64, float64, occultationRiseSetState, bool) {
	var state occultationRiseSetState
	for iteration := 0; iteration < occultationGreatestTimeContourCorrectionIterations; iteration++ {
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
		// 完整牛顿步可能一步跨出可见域（掩带很窄，限界附近的种子尤其容易）；逐步二分回退，
		// 只要还有一步落在域内就继续投影。
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

// traceOccultationGreatestTimeArc 从种子沿一个方向按弧长延拓，预测点落到定义域外时步长减半。
func traceOccultationGreatestTimeArc(
	evaluation occultationRiseSetEvaluation,
	location *time.Location,
	useContactMetric bool,
	longitude, latitude, direction float64,
) []OccultationPathPoint {
	arc := occultationGreatestTimeArc{evaluation: evaluation, location: location, useContactMetric: useContactMetric}
	_, state, ok := arc.sample(longitude, latitude)
	if !ok {
		return nil
	}
	points := []OccultationPathPoint{arc.point(longitude, latitude, state)}
	step := occultationGreatestTimeContourArcStepDegrees
	previousEast, previousNorth := 0.0, 0.0
	for count := 0; count < occultationGreatestTimeContourMaxArcSteps; count++ {
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
			if step < occultationGreatestTimeContourMinArcStepDegrees {
				break
			}
			continue
		}
		if math.Abs(nextLatitude) > occultationGreatestTimeContourLatitudeLimitDegrees {
			break
		}
		next := arc.point(nextLongitude, nextLatitude, nextState)
		distance := occultationPathDistanceKM(points[len(points)-1], next)
		// 校正回到原点说明该方向已经走到支路端点，继续只会原地打转。
		if distance < 1e-9 || distance > 4*step*greatestTimeContourKMPerDegree {
			break
		}
		points = append(points, next)
		longitude, latitude = nextLongitude, nextLatitude
		previousEast, previousNorth = tangentEast, tangentNorth
		step = math.Min(occultationGreatestTimeContourArcStepDegrees, step*1.5)
	}
	return points
}

// occultationGreatestTimeContourSeriesSeed 在一条序列上按时间插值取种子。
// 掩带是单条窄带，宽度常小于粗扫步长，二维经纬度扫描会整条漏掉，所以优先用中心线：
// 中心线各点的时刻就是该点的本地掩甚，按时间插值得到的点已经贴着等时线。
// 限界只在中心线的时间范围够不到该时刻时提供候选，且必须再经定义域校验（限界点的时刻是
// 擦边时刻，不等于本地掩甚，直接当种子可能落在可见域外）。
func occultationGreatestTimeContourSeriesSeed(points []OccultationPathPoint, levelTT float64) (float64, float64, bool) {
	for index := 1; index < len(points); index++ {
		before := occultationTimeToTT(points[index-1].Time)
		after := occultationTimeToTT(points[index].Time)
		if levelTT < before || levelTT > after || after <= before {
			continue
		}
		fraction := (levelTT - before) / (after - before)
		delta := math.Remainder(points[index].Longitude-points[index-1].Longitude, 360)
		return points[index-1].Longitude + delta*fraction,
			points[index-1].Latitude + (points[index].Latitude-points[index-1].Latitude)*fraction, true
	}
	return 0, 0, false
}

// occultationGreatestTimeContourSeeds 按序列顺序（中心线优先）给出该时刻的全部候选种子。
func occultationGreatestTimeContourSeeds(series [][]OccultationPathPoint, levelTT float64) [][2]float64 {
	seeds := make([][2]float64, 0, len(series))
	for _, points := range series {
		if longitude, latitude, ok := occultationGreatestTimeContourSeriesSeed(points, levelTT); ok {
			seeds = append(seeds, [2]float64{longitude, latitude})
		}
	}
	return seeds
}

// occultationGreatestTimeContourScanSeeds 是没有可用序列种子时的兜底粗扫。
func occultationGreatestTimeContourScanSeeds(arc occultationGreatestTimeArc) []OccultationPathPoint {
	seeds := make([]OccultationPathPoint, 0, 16)
	latitudeStep := occultationGreatestTimeContourSeedLatitudeStepDegrees
	longitudeStep := occultationGreatestTimeContourSeedLongitudeStepDegrees
	limit := occultationGreatestTimeContourSeedLatitudeLimitDegrees
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

func occultationGreatestTimeContourCovered(segments [][]OccultationPathPoint, point OccultationPathPoint) bool {
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

// occultationGreatestTimeContourSegmentCovered 判断整条支路是否已落在已绘曲线上（同一曲线被先后延拓两次时后一条可能更长）。
func occultationGreatestTimeContourSegmentCovered(segments [][]OccultationPathPoint, segment []OccultationPathPoint) bool {
	for _, point := range segment {
		if !occultationGreatestTimeContourCovered(segments, point) {
			return false
		}
	}
	return true
}

// occultationGreatestTimeContourPruneCovered 丢弃已被新支路整条覆盖的旧支路。
func occultationGreatestTimeContourPruneCovered(segments [][]OccultationPathPoint, added []OccultationPathPoint) [][]OccultationPathPoint {
	kept := segments[:0]
	for _, segment := range segments {
		if occultationGreatestTimeContourSegmentCovered([][]OccultationPathPoint{added}, segment) {
			continue
		}
		kept = append(kept, segment)
	}
	return kept
}

// occultationGreatestTimeContourSegments 汇总一个时刻取值上的全部等时线支路。
func occultationGreatestTimeContourSegments(
	evaluation occultationRiseSetEvaluation,
	location *time.Location,
	useContactMetric bool,
	series [][]OccultationPathPoint,
) [][]OccultationPathPoint {
	arc := occultationGreatestTimeArc{evaluation: evaluation, location: location, useContactMetric: useContactMetric}
	var seeds []OccultationPathPoint
	for _, candidate := range occultationGreatestTimeContourSeeds(series, evaluation.tt) {
		// 候选点可能落在可见域外（限界种子或该时刻已无接触），先校验再投影。
		if _, _, ok := arc.sample(candidate[0], candidate[1]); !ok {
			continue
		}
		correctedLongitude, correctedLatitude, state, corrected := arc.correct(candidate[0], candidate[1])
		if corrected {
			seeds = append(seeds, arc.point(correctedLongitude, correctedLatitude, state))
		}
	}
	// 每个时刻取值只保留能延拓出支路的种子；同一条曲线上的重复种子会被覆盖判据丢弃。
	if len(seeds) == 0 {
		seeds = occultationGreatestTimeContourScanSeeds(arc)
	}
	segments := make([][]OccultationPathPoint, 0, 2)
	for _, seed := range seeds {
		if occultationGreatestTimeContourCovered(segments, seed) {
			continue
		}
		forward := traceOccultationGreatestTimeArc(evaluation, location, useContactMetric, seed.Longitude, seed.Latitude, 1)
		backward := traceOccultationGreatestTimeArc(evaluation, location, useContactMetric, seed.Longitude, seed.Latitude, -1)
		segment := make([]OccultationPathPoint, 0, len(forward)+len(backward))
		for index := len(backward) - 1; index >= 1; index-- {
			segment = append(segment, backward[index])
		}
		segment = append(segment, forward...)
		if len(segment) < 2 {
			continue
		}
		// 先按整条支路去重：种子检查只能拦住"较短者先画"的情况，反序时需要在这里收口。
		if occultationGreatestTimeContourSegmentCovered(segments, segment) {
			continue
		}
		segments = occultationGreatestTimeContourPruneCovered(segments, segment)
		segments = append(segments, segment)
	}
	return segments
}

// occultationGreatestTimeLevels 取请求的掩甚时刻取值：显式取值优先（按时间截断到上限），
// 否则按步长对齐到 UTC 整刻度并覆盖可见窗口。掩星全球可见窗口通常只有数小时，间隔应比日食更密
// （15–30 分钟量级），否则整条掩带上只有寥寥几条线。
func occultationGreatestTimeLevels(options OccultationPathOptions, startTT, endTT float64) []greatestTimeContourLevel {
	if len(options.GreatestTimeValues) > 0 {
		values := options.GreatestTimeValues
		if len(values) > greatestTimeContourMaxLevels {
			values = values[:greatestTimeContourMaxLevels]
		}
		levels := make([]greatestTimeContourLevel, 0, len(values))
		for _, value := range values {
			levels = append(levels, greatestTimeContourLevel{
				tt: value,
				// 显式取值只有 TT 儒略日；抹掉亚毫秒噪声，避免整分被格式化成前一分钟。
				at: greatestTimeContourTTToUTC(value).Round(time.Millisecond),
			})
		}
		return levels
	}
	return greatestTimeContourAlignedLevels(startTT, endTT, options.GreatestTimeStep, greatestTimeContourMaxLevels)
}

// occultationGreatestTimeContours 计算请求时刻取值的地方掩甚时刻等值线。
func occultationGreatestTimeContours(
	levels []greatestTimeContourLevel,
	startTT, endTT float64,
	cache *occultationRiseSetEvaluationCache,
	series [][]OccultationPathPoint,
	useContactMetric bool,
	location *time.Location,
) []OccultationGreatestTimeContour {
	if len(levels) == 0 || cache == nil || startTT == 0 || endTT == 0 || endTT <= startTT {
		return nil
	}
	contours := make([]OccultationGreatestTimeContour, 0, len(levels))
	for _, level := range levels {
		if !finite(level.tt) || level.tt < startTT || level.tt > endTT {
			continue
		}
		segments := occultationGreatestTimeContourSegments(cache.evaluation(level.tt), location, useContactMetric, series)
		if len(segments) == 0 {
			continue
		}
		contours = append(contours, OccultationGreatestTimeContour{JDE: level.tt, Time: level.at, Segments: segments})
	}
	return contours
}

package basic

import (
	"math"
	"sort"
	"time"
)

const planetOccultationPathMaxTemporalSamples = 5000

type occultationPathFrameFunc func(float64) (occultationPathFrame, bool)

// FindPlanetOccultationPaths 搜索有限盘面行星月掩的全球外接触和内接触掩带。
// 查询窗口按全球几何掩甚点选择事件；端点容差 10 ms 与数值根精度一致。求解成功时，每条路径扩展到完整全球起止点。
// FindPlanetOccultationPaths searches the global outer- and inner-contact footprints of one finite-disk planet.
// The query window selects events by global geometric greatest, with a 10 ms endpoint tolerance matching the numerical root precision. Each returned path expands to its complete global start and end when solved.
func FindPlanetOccultationPaths(start, end time.Time, planet OccultationPlanet,
	options OccultationPathOptions) ([]PlanetOccultationPath, error) {
	if err := validateOccultationTimeRange(start, end); err != nil {
		return nil, err
	}
	if err := planet.Validate(); err != nil {
		return nil, err
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}

	config, _ := planetOccultationConfigFor(planet)
	options = normalizeOccultationPathOptions(options)
	startTT := occultationTimeToTT(start)
	endTT := occultationTimeToTT(end)
	candidateStartTT := startTT - occultationPathSearchSpanDays
	candidateEndTT := endTT + occultationPathSearchSpanDays
	candidates := planetOccultationCandidateGreatestTimes(
		candidateStartTT, candidateEndTT, planetOccultationDefaultStepDays, config, nil, 0,
	)
	paths := make([]PlanetOccultationPath, 0, len(candidates))
	for _, seedTT := range candidates {
		path, ok, err := planetOccultationPathAtSeed(seedTT, config, options, start.Location())
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if !occultationTimeInSelectionWindow(path.Greatest.Time, start, end) {
			continue
		}
		if len(paths) > 0 && math.Abs(paths[len(paths)-1].Greatest.Time.Sub(path.Greatest.Time).Seconds()) <= 60 {
			continue
		}
		paths = append(paths, path)
	}
	sort.SliceStable(paths, func(i, j int) bool {
		return paths[i].Greatest.Time.Before(paths[j].Greatest.Time)
	})
	return paths, nil
}

func planetOccultationPathAtSeed(
	seedTT float64,
	config planetOccultationConfig,
	options OccultationPathOptions,
	location *time.Location,
) (PlanetOccultationPath, bool, error) {
	frameAt := func(tt float64) (occultationPathFrame, bool) {
		return planetOccultationPathFrameAt(tt, config)
	}
	totalFrameAt := func(tt float64) (occultationPathFrame, bool) {
		return planetOccultationTotalPathFrameAt(tt, config)
	}
	searchStart := seedTT - occultationPathSearchSpanDays
	searchEnd := seedTT + occultationPathSearchSpanDays
	outerStart, outerEnd, ok := occultationPathWindowForFrame(seedTT, searchStart, searchEnd, frameAt, false)
	if !ok {
		return PlanetOccultationPath{}, false, nil
	}
	centerStart, centerEnd, hasCenter := occultationPathWindowForFrame(seedTT, searchStart, searchEnd, frameAt, true)
	greatestTT := occultationPathGreatestForFrame(seedTT, outerStart, outerEnd, frameAt)
	greatest, greatestOK := occultationPathCenterPointForFrame(greatestTT, frameAt, location)
	if !greatestOK && hasCenter {
		greatestTT = math.Max(centerStart, math.Min(centerEnd, greatestTT))
		greatest, greatestOK = occultationPathCenterPointForFrame(greatestTT, frameAt, location)
	}
	if !greatestOK {
		greatest, greatestOK = occultationPathBoundaryPointForFrame(greatestTT, frameAt, location)
	}
	if !greatestOK {
		return PlanetOccultationPath{}, false, nil
	}
	_, _, greatestWidth, greatestWidthOK := occultationPathLimitsAndWidthForFrame(greatestTT, frameAt)
	if !greatestWidthOK || greatestWidth <= 0 {
		return PlanetOccultationPath{}, false, nil
	}
	// 仅有边界的事件没有影轴与椭球交点；原回退点使用纬度极值弦宽，全掩带使用下方的地面横向宽度。统一两种接触带宽度定义，使有限盘面内外接触宽度可比较。
	// Boundary-only events do not have an axis/ellipsoid intersection. Their fallback point used to carry a latitude-extrema chord width, while total bands used the ground cross-track width below. Keep both contact bands on the same width definition so finite-disk inner/outer widths are comparable.
	greatest.WidthKM = greatestWidth
	totalStartTT, totalEndTT, hasTotal := occultationPathWindowForFrame(
		seedTT, searchStart, searchEnd, totalFrameAt, false,
	)
	hasTotal = hasTotal && greatestTT >= totalStartTT && greatestTT <= totalEndTT
	if planetOccultationPathTemporalSampleCount(
		outerStart, outerEnd, centerStart, centerEnd, hasCenter,
		totalStartTT, totalEndTT, hasTotal, greatestTT, options,
	) > planetOccultationPathMaxTemporalSamples {
		return PlanetOccultationPath{}, false, ErrOccultationPathSamplingLimit
	}

	start := occultationPathBoundaryEndpointForFrame(outerStart, frameAt, location, 1)
	end := occultationPathBoundaryEndpointForFrame(outerEnd, frameAt, location, -1)
	if !start.valid || !end.valid {
		return PlanetOccultationPath{}, false, nil
	}
	centerLine, northern, southern, err := planetOccultationPathSamples(
		outerStart, outerEnd, centerStart, centerEnd, hasCenter, greatestTT, frameAt, options, location,
	)
	if err != nil {
		return PlanetOccultationPath{}, false, err
	}
	path := PlanetOccultationPath{
		Planet:          config.planet,
		TargetID:        config.planet.String(),
		Start:           start.point,
		Greatest:        greatest,
		End:             end.point,
		Complete:        outerStart > searchStart && outerEnd < searchEnd,
		CenterLine:      centerLine,
		NorthernLimit:   occultationPathWithEndpoints(start.point, end.point, northern),
		SouthernLimit:   occultationPathWithEndpoints(start.point, end.point, southern),
		Step:            options.Step,
		TargetSpacingKM: options.TargetSpacingKM,
	}
	path.PartialFootprints = planetOccultationFootprints(
		outerStart, outerEnd, greatestTT, frameAt, options, location,
	)

	if !hasTotal {
		return path, true, nil
	}
	totalStart := occultationPathBoundaryEndpointForFrame(totalStartTT, totalFrameAt, location, 1)
	totalEnd := occultationPathBoundaryEndpointForFrame(totalEndTT, totalFrameAt, location, -1)
	_, _, totalWidth, totalWidthOK := occultationPathLimitsAndWidthForFrame(greatestTT, totalFrameAt)
	if !totalStart.valid || !totalEnd.valid || !totalWidthOK || totalWidth <= 0 {
		return path, true, nil
	}
	totalNorthern, totalSouthern := occultationPathBoundarySamplesForFrame(
		totalStartTT, totalEndTT, greatestTT, totalFrameAt, options, location,
	)
	if len(totalNorthern) == 0 || len(totalSouthern) == 0 {
		return path, true, nil
	}
	path.HasTotalBand = true
	path.TotalStart = totalStart.point
	path.TotalEnd = totalEnd.point
	path.TotalComplete = totalStartTT > searchStart && totalEndTT < searchEnd
	path.NorthernTotalLimit = occultationPathWithEndpoints(totalStart.point, totalEnd.point, totalNorthern)
	path.SouthernTotalLimit = occultationPathWithEndpoints(totalStart.point, totalEnd.point, totalSouthern)
	path.TotalFootprints = planetOccultationFootprints(
		totalStartTT, totalEndTT, greatestTT, totalFrameAt, options, location,
	)
	path.GreatestTotalWidthKM = totalWidth
	return path, true, nil
}

func planetOccultationPathTemporalSampleCount(
	outerStartTT, outerEndTT float64,
	centerStartTT, centerEndTT float64,
	hasCenter bool,
	totalStartTT, totalEndTT float64,
	hasTotal bool,
	greatestTT float64,
	options OccultationPathOptions,
) int {
	stepDays := float64(options.Step) / float64(24*time.Hour)
	count := len(occultationPathSampleTimes(outerStartTT, outerEndTT, greatestTT, stepDays))
	count += len(occultationPathSampleTimesWithLimit(
		outerStartTT, outerEndTT, greatestTT, stepDays, planetOccultationFootprintMaxSamples,
	))
	if hasCenter {
		count += len(occultationPathSampleTimes(centerStartTT, centerEndTT, greatestTT, stepDays))
	}
	if hasTotal {
		count += len(occultationPathSampleTimes(totalStartTT, totalEndTT, greatestTT, stepDays))
		count += len(occultationPathSampleTimesWithLimit(
			totalStartTT, totalEndTT, greatestTT, stepDays, planetOccultationFootprintMaxSamples,
		))
	}
	return count
}

func occultationPathWindowForFrame(
	seedTT, startTT, endTT float64,
	frameAt occultationPathFrameFunc,
	center bool,
) (float64, float64, bool) {
	left := math.Max(startTT, seedTT-occultationPathSearchSpanDays)
	right := math.Min(endTT, seedTT+occultationPathSearchSpanDays)
	if right <= left {
		return 0, 0, false
	}
	predicate := func(tt float64) bool {
		frame, ok := frameAt(tt)
		if !ok {
			return false
		}
		if center {
			_, _, ok = occultationEarthLineIntersection(frame.moon, frame.axis)
			return ok
		}
		return occultationPathFrameHasBoundary(frame)
	}

	step := occultationPathRangeStepDays
	first := math.NaN()
	previous := left
	previousOK := predicate(previous)
	if previousOK {
		first = previous
	} else {
		for tt := left + step; tt <= right; tt += step {
			current := math.Min(tt, right)
			currentOK := predicate(current)
			if currentOK {
				first = occultationPathRefineTransition(previous, current, predicate, false)
				break
			}
			previous = current
			previousOK = currentOK
		}
	}
	if math.IsNaN(first) {
		return 0, 0, false
	}

	last := first
	previous = first
	previousOK = true
	for tt := first + step; tt <= right; tt += step {
		current := math.Min(tt, right)
		currentOK := predicate(current)
		if !currentOK {
			last = occultationPathRefineTransition(previous, current, predicate, true)
			return first, last, true
		}
		last = current
		previous = current
		previousOK = currentOK
	}
	if previousOK {
		last = right
	}
	return first, last, true
}

func occultationPathGreatestForFrame(seedTT, startTT, endTT float64, frameAt occultationPathFrameFunc) float64 {
	left := math.Max(startTT, seedTT-0.75)
	right := math.Min(endTT, seedTT+0.75)
	if right <= left {
		return seedTT
	}
	impact := func(tt float64) float64 {
		frame, ok := frameAt(tt)
		if !ok {
			return math.Inf(1)
		}
		return math.Hypot(frame.moonProjectionX(), frame.moonProjectionY())
	}
	const goldenRatio = 0.6180339887498949
	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1 := impact(x1)
	f2 := impact(x2)
	for i := 0; i < 56; i++ {
		if f1 > f2 {
			left = x1
			x1, f1 = x2, f2
			x2 = left + goldenRatio*(right-left)
			f2 = impact(x2)
		} else {
			right = x2
			x2, f2 = x1, f1
			x1 = right - goldenRatio*(right-left)
			f1 = impact(x1)
		}
	}
	return (left + right) / 2
}

func planetOccultationPathSamples(
	outerStartTT, outerEndTT float64,
	centerStartTT, centerEndTT float64,
	hasCenter bool,
	greatestTT float64,
	frameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
) ([]OccultationPathPoint, []OccultationPathPoint, []OccultationPathPoint, error) {
	var centerLine []OccultationPathPoint
	if hasCenter {
		var err error
		centerLine, err = occultationPathCenterSamplesForFrame(centerStartTT, centerEndTT, greatestTT, frameAt, options, location)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	northern, southern := occultationPathBoundarySamplesForFrame(
		outerStartTT, outerEndTT, greatestTT, frameAt, options, location,
	)
	return centerLine, northern, southern, nil
}

func occultationPathBoundarySamplesForFrame(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
) ([]OccultationPathPoint, []OccultationPathPoint) {
	stepDays := float64(options.Step) / float64(24*time.Hour)
	times := occultationPathSampleTimes(startTT, endTT, greatestTT, stepDays)
	samples := make([]occultationPathBoundaryPairSample, 0, len(times))
	for _, tt := range times {
		firstVector, secondVector, ok := occultationPathCrossTrackLimitsForFrame(tt, frameAt)
		if !ok {
			continue
		}
		sample := occultationPathBoundaryPairSample{tt: tt, first: firstVector, second: secondVector}
		if len(samples) == 0 {
			samples = append(samples, sample)
			continue
		}
		samples = appendOccultationPathBoundaryPairSegment(samples, samples[len(samples)-1], sample, frameAt, 0)
	}
	first := make([]OccultationPathPoint, len(samples))
	second := make([]OccultationPathPoint, len(samples))
	for index, sample := range samples {
		first[index] = occultationPathPointFromVector(sample.tt, sample.first, 0, location)
		second[index] = occultationPathPointFromVector(sample.tt, sample.second, 0, location)
	}
	return occultationPathOrientBoundarySamples(first, second, greatestTT)
}

type occultationPathBoundaryPairSample struct {
	tt            float64
	first, second occultationPathVector
}

func appendOccultationPathBoundaryPairSegment(
	samples []occultationPathBoundaryPairSample,
	start, end occultationPathBoundaryPairSample,
	frameAt occultationPathFrameFunc,
	depth int,
) []occultationPathBoundaryPairSample {
	if depth >= occultationPathMaxAdaptiveDepth ||
		occultationPathBoundaryPairSpacing(start, end) <= occultationPathBoundarySpacingKM {
		return append(samples, end)
	}
	midTT := (start.tt + end.tt) / 2
	first, second, ok := occultationPathCrossTrackLimitsForFrame(midTT, frameAt)
	if !ok {
		return append(samples, end)
	}
	mid := occultationPathBoundaryPairSample{tt: midTT, first: first, second: second}
	samples = appendOccultationPathBoundaryPairSegment(samples, start, mid, frameAt, depth+1)
	return appendOccultationPathBoundaryPairSegment(samples, mid, end, frameAt, depth+1)
}

func occultationPathBoundaryPairSpacing(first, second occultationPathBoundaryPairSample) float64 {
	firstA := occultationPathEarthFixedVector(first.tt, first.first)
	firstB := occultationPathEarthFixedVector(first.tt, first.second)
	secondA := occultationPathEarthFixedVector(second.tt, second.first)
	secondB := occultationPathEarthFixedVector(second.tt, second.second)
	direct := math.Max(
		occultationPathNorm(occultationPathSub(secondA, firstA)),
		occultationPathNorm(occultationPathSub(secondB, firstB)),
	)
	swapped := math.Max(
		occultationPathNorm(occultationPathSub(secondB, firstA)),
		occultationPathNorm(occultationPathSub(secondA, firstB)),
	)
	return math.Min(direct, swapped)
}

func occultationPathCrossTrackLimitsForFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
) (occultationPathVector, occultationPathVector, bool) {
	frame, ok := frameAt(tt)
	before, beforeOK := frameAt(tt - occultationPathVelocityStepDays)
	after, afterOK := frameAt(tt + occultationPathVelocityStepDays)
	if !ok || !beforeOK || !afterOK {
		return occultationPathVector{}, occultationPathVector{}, false
	}
	vx := after.moonProjectionX() - before.moonProjectionX()
	vy := after.moonProjectionY() - before.moonProjectionY()
	if math.Hypot(vx, vy) <= 1e-12 {
		return occultationPathVector{}, occultationPathVector{}, false
	}

	// 复用日食中心线构造：取基准面中垂直于运动方向的两条影锥母线。点源掠过阶段将缺失母线限制到可见角度区间；有限目标使用最近可见区间两端，直到两条横向母线分别与地球相交。
	// Match the solar-eclipse central-path construction: take the two shadow generators perpendicular to motion in the fundamental plane. During a grazing point-source phase, clamp a missing generator to the visible-angle interval. For a finite target, use both ends of the nearest visible interval until the two cross-track generators intersect Earth independently.
	theta := math.Atan2(vx, -vy)
	first, _, firstOK := occultationPathBoundaryVector(frame, theta)
	second, _, secondOK := occultationPathBoundaryVector(frame, theta+math.Pi)
	if frame.targetRadius == 0 {
		if !firstOK {
			first, firstOK = occultationPathBoundaryAtNearestPointSourceTheta(frame, theta)
		}
		if !secondOK {
			second, secondOK = occultationPathBoundaryAtNearestPointSourceTheta(frame, theta+math.Pi)
		}
		if !firstOK || !secondOK {
			return occultationPathVector{}, occultationPathVector{}, false
		}
		return first, second, true
	}
	intervals := occultationPathBoundaryThetaIntervals(frame)
	interval, intervalOK := occultationPathNearestThetaInterval(intervals, theta)
	if firstOK && secondOK {
		if intervalOK && occultationPathAngleDistance(interval.left, theta) > occultationPathAngleDistance(interval.left, theta+math.Pi) {
			first, second = second, first
		}
		return first, second, true
	}
	if !intervalOK {
		return occultationPathVector{}, occultationPathVector{}, false
	}
	first, _, firstOK = occultationPathBoundaryVector(frame, interval.left)
	second, _, secondOK = occultationPathBoundaryVector(frame, interval.right)
	if !firstOK || !secondOK {
		return occultationPathVector{}, occultationPathVector{}, false
	}
	return first, second, true
}

func occultationPathAngleDistance(first, second float64) float64 {
	return math.Abs(math.Remainder(first-second, 2*math.Pi))
}

func occultationPathBoundaryAtNearestPointSourceTheta(
	frame occultationPathFrame,
	theta float64,
) (occultationPathVector, bool) {
	_, tangentTheta, tangentOK := occultationPathBoundaryTangent(frame)
	if !tangentOK {
		return occultationPathVector{}, false
	}
	left, right, intervalOK := occultationPathBoundaryThetaInterval(frame, tangentTheta)
	if !intervalOK {
		return occultationPathVector{}, false
	}
	middle := (left + right) / 2
	theta += 2 * math.Pi * math.Round((middle-theta)/(2*math.Pi))
	if theta < left {
		theta = left
	} else if theta > right {
		theta = right
	}
	point, _, ok := occultationPathBoundaryVector(frame, theta)
	return point, ok
}

type occultationPathThetaInterval struct {
	left, right float64
}

func occultationPathNearestThetaInterval(
	intervals []occultationPathThetaInterval,
	theta float64,
) (occultationPathThetaInterval, bool) {
	var closest occultationPathThetaInterval
	closestDistance := math.Inf(1)
	for _, interval := range intervals {
		middle := (interval.left + interval.right) / 2
		firstDelta := math.Abs(math.Remainder(theta-middle, 2*math.Pi))
		secondDelta := math.Abs(math.Remainder(theta+math.Pi-middle, 2*math.Pi))
		distance := math.Min(firstDelta, secondDelta)
		if distance < closestDistance {
			closest = interval
			closestDistance = distance
		}
	}
	if !finite(closestDistance) {
		return occultationPathThetaInterval{}, false
	}
	return closest, true
}

func occultationPathBoundaryThetaIntervals(frame occultationPathFrame) []occultationPathThetaInterval {
	step := 2 * math.Pi / float64(occultationPathBoundaryScanPoints)
	discriminants := make([]float64, occultationPathBoundaryScanPoints)
	for index := range discriminants {
		discriminant, _, _, ok := occultationPathBoundaryLine(frame, step*float64(index))
		if !ok {
			discriminant = math.Inf(-1)
		}
		discriminants[index] = discriminant
	}

	intervals := make([]occultationPathThetaInterval, 0, 2)
	for index, value := range discriminants {
		previous := discriminants[(index+len(discriminants)-1)%len(discriminants)]
		next := discriminants[(index+1)%len(discriminants)]
		if value < previous || value < next {
			continue
		}
		theta := occultationPathRefineBoundaryMaximum(frame, step*float64(index), step)
		discriminant, b, scale, ok := occultationPathBoundaryLine(frame, theta)
		tolerance := 1e-12 * math.Max(scale, 1)
		if !ok || b >= 0 || discriminant < -tolerance {
			continue
		}
		left, right, intervalOK := occultationPathBoundaryThetaInterval(frame, theta)
		if intervalOK {
			intervals = append(intervals, occultationPathThetaInterval{left: left, right: right})
		}
	}
	return intervals
}

func occultationPathRefineBoundaryMaximum(frame occultationPathFrame, center, step float64) float64 {
	left := center - step
	right := center + step
	const goldenRatio = 0.6180339887498949
	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1, _, _, _ := occultationPathBoundaryLine(frame, x1)
	f2, _, _, _ := occultationPathBoundaryLine(frame, x2)
	for iteration := 0; iteration < 40; iteration++ {
		if f1 < f2 {
			left = x1
			x1, f1 = x2, f2
			x2 = left + goldenRatio*(right-left)
			f2, _, _, _ = occultationPathBoundaryLine(frame, x2)
		} else {
			right = x2
			x2, f2 = x1, f1
			x1 = right - goldenRatio*(right-left)
			f1, _, _, _ = occultationPathBoundaryLine(frame, x1)
		}
	}
	return (left + right) / 2
}

func occultationPathOrientBoundarySamples(
	first, second []OccultationPathPoint,
	greatestTT float64,
) ([]OccultationPathPoint, []OccultationPathPoint) {
	if len(first) == 0 || len(first) != len(second) {
		return first, second
	}
	nearest := 0
	nearestDelta := math.Inf(1)
	for index := range first {
		delta := math.Abs(centerTimeTT(first[index].Time) - greatestTT)
		if delta < nearestDelta {
			nearest = index
			nearestDelta = delta
		}
	}
	if first[nearest].Latitude >= second[nearest].Latitude {
		return first, second
	}
	return second, first
}

func occultationPathCenterSamplesForFrame(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
) ([]OccultationPathPoint, error) {
	stepDays := float64(options.Step) / float64(24*time.Hour)
	times := occultationPathSampleTimes(startTT, endTT, greatestTT, stepDays)
	points := make([]OccultationPathPoint, 0, len(times))
	for _, tt := range times {
		point, ok := occultationPathCenterPointForFrame(tt, frameAt, location)
		if ok {
			points = append(points, point)
		}
	}
	if options.TargetSpacingKM > 0 {
		return refineOccultationPathSpacingForFrame(points, frameAt, options.TargetSpacingKM, location)
	}
	return points, nil
}

func refineOccultationPathSpacingForFrame(
	points []OccultationPathPoint,
	frameAt occultationPathFrameFunc,
	targetSpacingKM float64,
	location *time.Location,
) ([]OccultationPathPoint, error) {
	if len(points) < 2 || targetSpacingKM <= 0 {
		return points, nil
	}
	refined := make([]OccultationPathPoint, 0, len(points))
	refined = append(refined, points[0])
	widthAt := func(tt float64) (float64, bool) {
		_, _, width, ok := occultationPathLimitsAndWidthForFrame(tt, frameAt)
		return width, ok
	}
	for i := 1; i < len(points); i++ {
		segmentStart := len(refined) - 1
		var err error
		refined, err = appendOccultationPathSegmentForFrame(refined, points[i-1], points[i], frameAt, targetSpacingKM, location, 0)
		if err != nil {
			return nil, err
		}
		refineOccultationPathWidths(refined[segmentStart:], widthAt)
	}
	return refined, nil
}

func appendOccultationPathSegmentForFrame(
	points []OccultationPathPoint,
	start, end OccultationPathPoint,
	frameAt occultationPathFrameFunc,
	targetSpacingKM float64,
	location *time.Location,
	depth int,
) ([]OccultationPathPoint, error) {
	distance := occultationPathDistanceKM(start, end)
	if distance <= targetSpacingKM {
		if len(points) >= occultationPathMaxSampleCount {
			return nil, ErrOccultationPathSamplingLimit
		}
		return append(points, end), nil
	}
	if depth >= occultationPathMaxAdaptiveDepth || len(points) >= occultationPathMaxSampleCount {
		return nil, ErrOccultationPathSamplingLimit
	}
	midTT := (centerTimeTT(start.Time) + centerTimeTT(end.Time)) / 2
	mid, ok := occultationPathCenterPointForFrameWithoutWidth(midTT, frameAt, location)
	if !ok {
		return append(points, end), nil
	}
	mid.WidthKM = (start.WidthKM + end.WidthKM) / 2
	var err error
	points, err = appendOccultationPathSegmentForFrame(points, start, mid, frameAt, targetSpacingKM, location, depth+1)
	if err != nil {
		return nil, err
	}
	return appendOccultationPathSegmentForFrame(points, mid, end, frameAt, targetSpacingKM, location, depth+1)
}

func planetOccultationPathFrameAt(tt float64, config planetOccultationConfig) (occultationPathFrame, bool) {
	return planetOccultationContactPathFrameAt(tt, config, false)
}

func planetOccultationTotalPathFrameAt(tt float64, config planetOccultationConfig) (occultationPathFrame, bool) {
	return planetOccultationContactPathFrameAt(tt, config, true)
}

func planetOccultationContactPathFrameAt(tt float64, config planetOccultationConfig, total bool) (occultationPathFrame, bool) {
	moonRA, moonDec := HMoonGeocentricApparentRaDecN(tt, -1)
	moonDistance := HMoonAwayN(tt, -1)
	planetRA, planetDec := config.apparentRaDecN(tt, -1)
	planetDistance := config.earthDistanceN(tt, -1) * occultationPathAstronomicalUnitKM
	if !finite(moonRA) || !finite(moonDec) || !finite(moonDistance) || moonDistance <= 0 ||
		!finite(planetRA) || !finite(planetDec) || !finite(planetDistance) || planetDistance <= moonDistance {
		return occultationPathFrame{}, false
	}
	moon := occultationPathRaDecVector(moonRA, moonDec, moonDistance)
	target := occultationPathRaDecVector(planetRA, planetDec, planetDistance)
	moonToTarget := occultationPathSub(target, moon)
	moonToTargetDistance := occultationPathNorm(moonToTarget)
	moonRadius := MoonSemidiameter(tt) * math.Pi / (180 * 3600)
	moonRadiusKM := occultationPathNorm(moon) * math.Sin(moonRadius)
	contactRadiusKM := moonRadiusKM + config.equatorialRadiusKM
	if total {
		contactRadiusKM = moonRadiusKM - config.equatorialRadiusKM
	}
	if moonToTargetDistance <= math.Abs(contactRadiusKM) {
		return occultationPathFrame{}, false
	}
	axis := occultationPathUnit(occultationPathScale(moonToTarget, -1))
	north := occultationPathVector{z: 1}
	first := occultationPathCross(north, axis)
	if occultationPathNorm(first) < 1e-12 {
		first = occultationPathCross(occultationPathVector{x: 1}, axis)
	}
	first = occultationPathUnit(first)
	second := occultationPathUnit(occultationPathCross(axis, first))
	return occultationPathFrame{
		moon:         moon,
		axis:         axis,
		first:        first,
		second:       second,
		moonRadius:   moonRadius,
		targetRadius: math.Asin(contactRadiusKM / moonToTargetDistance),
	}, true
}

func occultationPathFrameHasBoundary(frame occultationPathFrame) bool {
	_, _, ok := occultationPathBoundaryTangent(frame)
	return ok
}

func occultationPathBoundaryEndpointForFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	direction int,
) occultationPathEndpoint {
	if _, ok := frameAt(tt); !ok {
		return occultationPathEndpoint{}
	}
	for offset := 0; offset <= 3; offset++ {
		candidateTT := tt + float64(direction*offset)*0.5/86400.0
		frame, ok := frameAt(candidateTT)
		if !ok {
			continue
		}
		vector, _, valid := occultationPathBoundaryTangent(frame)
		if valid {
			return occultationPathEndpoint{point: occultationPathPointFromVector(candidateTT, vector, 0, location), valid: true}
		}
	}
	return occultationPathEndpoint{}
}

func occultationPathCenterPointForFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
) (OccultationPathPoint, bool) {
	frame, ok := frameAt(tt)
	if !ok {
		return OccultationPathPoint{}, false
	}
	point, _, ok := occultationEarthLineIntersection(frame.moon, frame.axis)
	if !ok {
		return OccultationPathPoint{}, false
	}
	width := 0.0
	if _, _, tangentWidth, limitsOK := occultationPathLimitsAndWidthForFrame(tt, frameAt); limitsOK {
		width = tangentWidth
	}
	return occultationPathPointFromVectorWithMoon(tt, point, width, frame.moon, location), true
}

func occultationPathCenterPointForFrameWithoutWidth(
	tt float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
) (OccultationPathPoint, bool) {
	frame, ok := frameAt(tt)
	if !ok {
		return OccultationPathPoint{}, false
	}
	point, _, ok := occultationEarthLineIntersection(frame.moon, frame.axis)
	if !ok {
		return OccultationPathPoint{}, false
	}
	return occultationPathPointFromVectorWithMoon(tt, point, 0, frame.moon, location), true
}

func occultationPathBoundaryPointForFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
) (OccultationPathPoint, bool) {
	frame, ok := frameAt(tt)
	if !ok {
		return OccultationPathPoint{}, false
	}
	point, _, ok := occultationPathBoundaryTangent(frame)
	if !ok {
		return OccultationPathPoint{}, false
	}
	north, south, limitsOK := occultationPathScannedLimitsAtFrame(tt, frame)
	return occultationPathPointFromVector(tt, point, occultationPathBoundaryWidth(north, south, limitsOK), location), true
}

func occultationPathLimitsAndWidthForFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
) (occultationPathVector, occultationPathVector, float64, bool) {
	frame, ok := frameAt(tt)
	if !ok {
		return occultationPathVector{}, occultationPathVector{}, 0, false
	}
	beforeFrame, beforeOK := frameAt(tt - occultationPathVelocityStepDays)
	afterFrame, afterOK := frameAt(tt + occultationPathVelocityStepDays)
	if !beforeOK || !afterOK {
		north, south, scannedOK := occultationPathScannedLimitsForFrame(tt, frame)
		return north, south, occultationPathBoundaryWidth(north, south, scannedOK), scannedOK
	}

	vx := afterFrame.moonProjectionX() - beforeFrame.moonProjectionX()
	vy := afterFrame.moonProjectionY() - beforeFrame.moonProjectionY()
	speed := math.Hypot(vx, vy)
	if speed <= 1e-12 {
		north, south, scannedOK := occultationPathScannedLimitsForFrame(tt, frame)
		return north, south, occultationPathBoundaryWidth(north, south, scannedOK), scannedOK
	}
	planeCrossTrack := occultationPathUnit(occultationPathAdd(
		occultationPathScale(frame.first, -vy/speed),
		occultationPathScale(frame.second, vx/speed),
	))

	var centerFixed, groundCrossTrack occultationPathVector
	groundWidthOK := false
	center, centerOK := occultationPathTrackReference(frame)
	before, beforeCenterOK := occultationPathTrackReference(beforeFrame)
	after, afterCenterOK := occultationPathTrackReference(afterFrame)
	if centerOK && beforeCenterOK && afterCenterOK {
		centerFixed = occultationPathEarthFixedVector(tt, center)
		beforeFixed := occultationPathEarthFixedVector(tt-occultationPathVelocityStepDays, before)
		afterFixed := occultationPathEarthFixedVector(tt+occultationPathVelocityStepDays, after)
		polarRatioSquared := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
		normal := occultationPathUnit(occultationPathVector{x: centerFixed.x, y: centerFixed.y, z: centerFixed.z / polarRatioSquared})
		track := occultationPathSub(afterFixed, beforeFixed)
		track = occultationPathSub(track, occultationPathScale(normal, occultationPathDot(track, normal)))
		if occultationPathNorm(track) > 1e-12 {
			groundCrossTrack = occultationPathUnit(occultationPathCross(normal, occultationPathUnit(track)))
			groundWidthOK = true
		}
	}

	minimumOffset := math.Inf(1)
	maximumOffset := math.Inf(-1)
	minimumGroundOffset := math.Inf(1)
	maximumGroundOffset := math.Inf(-1)
	var minimumPoint, maximumPoint occultationPathVector
	var minimumGroundPoint, maximumGroundPoint occultationPathVector
	consider := func(point occultationPathVector) {
		offset := occultationPathDot(point, planeCrossTrack)
		if offset < minimumOffset {
			minimumOffset = offset
			minimumPoint = point
		}
		if offset > maximumOffset {
			maximumOffset = offset
			maximumPoint = point
		}
		if groundWidthOK {
			fixed := occultationPathEarthFixedVector(tt, point)
			groundOffset := occultationPathDot(occultationPathSub(fixed, centerFixed), groundCrossTrack)
			if groundOffset < minimumGroundOffset {
				minimumGroundOffset = groundOffset
				minimumGroundPoint = point
			}
			if groundOffset > maximumGroundOffset {
				maximumGroundOffset = groundOffset
				maximumGroundPoint = point
			}
		}
	}
	if tangentPoint, tangentTheta, tangentOK := occultationPathBoundaryTangent(frame); tangentOK {
		consider(tangentPoint)
		if leftTheta, rightTheta, intervalOK := occultationPathBoundaryThetaInterval(frame, tangentTheta); intervalOK {
			const intervalSamples = 128
			for i := 0; i <= intervalSamples; i++ {
				theta := leftTheta + (rightTheta-leftTheta)*float64(i)/intervalSamples
				if point, _, pointOK := occultationPathBoundaryVector(frame, theta); pointOK {
					consider(point)
				}
			}
		}
	}
	for i := 0; i < occultationPathBoundaryScanPoints; i++ {
		point, _, pointOK := occultationPathBoundaryVector(frame, 2*math.Pi*float64(i)/float64(occultationPathBoundaryScanPoints))
		if !pointOK {
			continue
		}
		consider(point)
	}
	if !finite(minimumOffset) || !finite(maximumOffset) {
		return occultationPathVector{}, occultationPathVector{}, 0, false
	}
	width := maximumOffset - minimumOffset
	if groundWidthOK && finite(minimumGroundOffset) && finite(maximumGroundOffset) {
		width = maximumGroundOffset - minimumGroundOffset
		// 有限目标会把月影打开或收束成圆锥；接近地平线时其影面投影可能折叠，使全球投影极值在远处地平线交点间跳变。地面轨迹极值仍位于掩带的同一物理侧。
		// A finite target opens or closes the lunar shadow into a cone. Near the horizon its shadow-plane projection can fold, causing the global projected extremum to jump between distant horizon intersections. Ground-track extrema remain on the same physical sides of the band.
		if frame.targetRadius != 0 {
			return maximumGroundPoint, minimumGroundPoint, width, true
		}
	}
	_, minimumLatitude := occultationPathGeodetic(tt, minimumPoint)
	_, maximumLatitude := occultationPathGeodetic(tt, maximumPoint)
	if maximumLatitude >= minimumLatitude {
		return maximumPoint, minimumPoint, width, true
	}
	return minimumPoint, maximumPoint, width, true
}

func occultationPathScannedLimitsForFrame(
	tt float64,
	frame occultationPathFrame,
) (occultationPathVector, occultationPathVector, bool) {
	return occultationPathScannedLimitsAtFrame(tt, frame)
}

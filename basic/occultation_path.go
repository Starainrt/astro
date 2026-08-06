package basic

import (
	"math"
	"sort"
	"time"
)

const (
	occultationPathDefaultStepDays    = 1.0 / 1440.0
	occultationPathMinStepDays        = 1.0 / 86400.0
	occultationPathMaxSampleCount     = 30000
	occultationPathMaxAdaptiveDepth   = 20
	occultationPathBoundarySpacingKM  = 500.0
	occultationPathVelocityStepDays   = 1.0 / 1440.0
	occultationPathBoundaryScanPoints = 720
	occultationPathRootToleranceDays  = occultationEventSelectionToleranceDays
	occultationPathRangeStepDays      = 5.0 / 1440.0
	occultationPathSearchSpanDays     = 2.0
	occultationPathWidthToleranceKM   = 0.005

	occultationPathEarthEquatorialRadiusKM = 6378.1366
	occultationPathEarthPolarRatio         = 0.99664719
	occultationPathAstronomicalUnitKM      = 149597870.7
)

// FindStarOccultationPaths 搜索单颗点源恒星月掩的全球掩带。
// 查询窗口按全球几何掩甚点选择事件，端点容差为 10 ms，与数值根精度一致。返回路径扩展到完整全球起止点；函数不会加载内嵌星表，调用者需显式提供坐标。
// FindStarOccultationPaths searches the global lunar-occultation footprint of one point-source star.
// The query window selects events by global geometric greatest, with a 10 ms endpoint tolerance matching the numerical root precision. Each returned path expands to its complete global start and end; the function does not load the embedded catalog.
func FindStarOccultationPaths(start, end time.Time, star StarCoordinate, options OccultationPathOptions) ([]StarOccultationPath, error) {
	if err := validateOccultationTimeRange(start, end); err != nil {
		return nil, err
	}
	if err := star.Validate(); err != nil {
		return nil, err
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}

	options = normalizeOccultationPathOptions(options)
	startTT := occultationTimeToTT(start)
	endTT := occultationTimeToTT(end)
	candidateStartTT := startTT - occultationPathSearchSpanDays
	candidateEndTT := endTT + occultationPathSearchSpanDays
	coarseOptions := OccultationSearchOptions{}
	candidates := starOccultationGeocentricCandidateGreatestTimes(
		candidateStartTT,
		candidateEndTT,
		starOccultationCoarseStepDays(coarseOptions),
		star,
		0,
	)
	paths := make([]StarOccultationPath, 0, len(candidates))
	for _, seedTT := range candidates {
		path, ok, err := starOccultationPathAtSeed(seedTT, star, options, start.Location())
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

func normalizeOccultationPathOptions(options OccultationPathOptions) OccultationPathOptions {
	if options.Step <= 0 {
		options.Step = time.Duration(occultationPathDefaultStepDays * float64(24*time.Hour))
	}
	if float64(options.Step)/float64(24*time.Hour) < occultationPathMinStepDays {
		options.Step = time.Second
	}
	if options.TargetSpacingKM <= 0 || math.IsNaN(options.TargetSpacingKM) || math.IsInf(options.TargetSpacingKM, 0) {
		options.TargetSpacingKM = 0
	}
	return options
}

func starOccultationPathAtSeed(
	seedTT float64,
	star StarCoordinate,
	options OccultationPathOptions,
	location *time.Location,
) (StarOccultationPath, bool, error) {
	searchStart := seedTT - occultationPathSearchSpanDays
	searchEnd := seedTT + occultationPathSearchSpanDays
	outerStart, outerEnd, ok := starOccultationPathWindow(seedTT, searchStart, searchEnd, star, false)
	if !ok {
		return StarOccultationPath{}, false, nil
	}

	centerStart, centerEnd, hasCenter := starOccultationPathWindow(seedTT, searchStart, searchEnd, star, true)
	greatestTT := starOccultationPathGreatest(seedTT, outerStart, outerEnd, star)
	greatest, greatestOK := starOccultationPathCenterPoint(greatestTT, star, location)
	if !greatestOK {
		if hasCenter {
			greatestTT = math.Max(centerStart, math.Min(centerEnd, greatestTT))
			greatest, greatestOK = starOccultationPathCenterPoint(greatestTT, star, location)
		}
	}
	if !greatestOK {
		frameAt := func(tt float64) (occultationPathFrame, bool) {
			return starOccultationPathFrameAt(tt, star)
		}
		greatest, greatestOK = occultationPathBoundaryPointForFrame(greatestTT, frameAt, location)
	}
	if !greatestOK {
		return StarOccultationPath{}, false, nil
	}

	start := starOccultationPathBoundaryEndpoint(outerStart, star, location, 1)
	end := starOccultationPathBoundaryEndpoint(outerEnd, star, location, -1)
	if !start.valid || !end.valid {
		return StarOccultationPath{}, false, nil
	}

	path := StarOccultationPath{
		TargetID:        star.ID,
		Start:           start.point,
		Greatest:        greatest,
		End:             end.point,
		Complete:        outerStart > searchStart && outerEnd < searchEnd,
		Step:            options.Step,
		TargetSpacingKM: options.TargetSpacingKM,
	}
	centerLine, northern, southern, err := starOccultationPathSamples(
		outerStart,
		outerEnd,
		centerStart,
		centerEnd,
		hasCenter,
		greatestTT,
		star,
		options,
		location,
	)
	if err != nil {
		return StarOccultationPath{}, false, err
	}
	path.CenterLine = centerLine
	path.NorthernLimit = occultationPathWithEndpoints(start.point, end.point, northern)
	path.SouthernLimit = occultationPathWithEndpoints(start.point, end.point, southern)
	return path, true, nil
}

func occultationPathWithEndpoints(start, end OccultationPathPoint, points []OccultationPathPoint) []OccultationPathPoint {
	result := make([]OccultationPathPoint, 0, len(points)+2)
	result = append(result, start)
	for _, point := range points {
		if point.Time.After(start.Time) && point.Time.Before(end.Time) {
			result = append(result, point)
		}
	}
	return append(result, end)
}

func starOccultationPathWindow(seedTT, startTT, endTT float64, star StarCoordinate, center bool) (float64, float64, bool) {
	left := math.Max(startTT, seedTT-occultationPathSearchSpanDays)
	right := math.Min(endTT, seedTT+occultationPathSearchSpanDays)
	if right <= left {
		return 0, 0, false
	}
	predicate := func(tt float64) bool {
		frame, ok := starOccultationPathFrameAt(tt, star)
		if !ok {
			return false
		}
		if center {
			_, _, ok = occultationEarthLineIntersection(frame.moon, frame.axis)
			return ok
		}
		return starOccultationPathHasBoundary(frame)
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

func occultationPathRefineTransition(left, right float64, predicate func(float64) bool, trueToFalse bool) float64 {
	leftOK := predicate(left)
	for i := 0; i < 48 && math.Abs(right-left) > occultationPathRootToleranceDays; i++ {
		mid := (left + right) / 2
		midOK := predicate(mid)
		if trueToFalse {
			if midOK {
				left = mid
			} else {
				right = mid
			}
			continue
		}
		if midOK {
			right = mid
		} else {
			left = mid
		}
	}
	if trueToFalse {
		return left
	}
	if leftOK {
		return left
	}
	return right
}

func starOccultationPathGreatest(seedTT, startTT, endTT float64, star StarCoordinate) float64 {
	left := math.Max(startTT, seedTT-0.75)
	right := math.Min(endTT, seedTT+0.75)
	if right <= left {
		return seedTT
	}
	const goldenRatio = 0.6180339887498949
	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1 := starOccultationPathImpact(x1, star)
	f2 := starOccultationPathImpact(x2, star)
	for i := 0; i < 56; i++ {
		if f1 > f2 {
			left = x1
			x1 = x2
			f1 = f2
			x2 = left + goldenRatio*(right-left)
			f2 = starOccultationPathImpact(x2, star)
		} else {
			right = x2
			x2 = x1
			f2 = f1
			x1 = right - goldenRatio*(right-left)
			f1 = starOccultationPathImpact(x1, star)
		}
	}
	return (left + right) / 2
}

func starOccultationPathImpact(tt float64, star StarCoordinate) float64 {
	frame, ok := starOccultationPathFrameAt(tt, star)
	if !ok {
		return math.Inf(1)
	}
	return math.Hypot(frame.moonProjectionX(), frame.moonProjectionY())
}

func starOccultationPathSamples(
	outerStartTT, outerEndTT float64,
	centerStartTT, centerEndTT float64,
	hasCenter bool,
	greatestTT float64,
	star StarCoordinate,
	options OccultationPathOptions,
	location *time.Location,
) ([]OccultationPathPoint, []OccultationPathPoint, []OccultationPathPoint, error) {
	var points []OccultationPathPoint
	if hasCenter {
		var err error
		points, err = starOccultationPathCenterSamples(centerStartTT, centerEndTT, greatestTT, star, options, location)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	frameAt := func(tt float64) (occultationPathFrame, bool) {
		return starOccultationPathFrameAt(tt, star)
	}
	northern, southern := occultationPathBoundarySamplesForFrame(
		outerStartTT, outerEndTT, greatestTT, frameAt, options, location,
	)
	return points, northern, southern, nil
}

func starOccultationPathCenterSamples(
	startTT, endTT, greatestTT float64,
	star StarCoordinate,
	options OccultationPathOptions,
	location *time.Location,
) ([]OccultationPathPoint, error) {
	stepDays := float64(options.Step) / float64(24*time.Hour)
	times := occultationPathSampleTimes(startTT, endTT, greatestTT, stepDays)
	points := make([]OccultationPathPoint, 0, len(times))
	for _, tt := range times {
		point, ok := starOccultationPathCenterPoint(tt, star, location)
		if ok {
			points = append(points, point)
		}
	}
	if options.TargetSpacingKM > 0 {
		return refineOccultationPathSpacing(points, star, options.TargetSpacingKM, location)
	}
	return points, nil
}

func occultationPathSampleTimes(startTT, endTT, greatestTT, stepDays float64) []float64 {
	return occultationPathSampleTimesWithLimit(
		startTT, endTT, greatestTT, stepDays, occultationPathMaxSampleCount,
	)
}

func occultationPathSampleTimesWithLimit(
	startTT, endTT, greatestTT, stepDays float64,
	maximumCount int,
) []float64 {
	if endTT < startTT {
		startTT, endTT = endTT, startTT
	}
	if maximumCount < 3 {
		maximumCount = 3
	}
	duration := endTT - startTT
	if duration <= 0 {
		return []float64{startTT}
	}
	if stepDays <= 0 || !finite(stepDays) {
		stepDays = duration
	}
	baseSampleCount := int(math.Ceil(duration/stepDays)) + 1
	if baseSampleCount+1 > maximumCount {
		times := []float64{startTT, greatestTT, endTT}
		interiorCount := maximumCount - len(times)
		for index := 1; index <= interiorCount; index++ {
			times = append(times, startTT+duration*float64(index)/float64(interiorCount+1))
		}
		sort.Float64s(times)
		return uniqueOccultationPathTimes(times)
	}
	times := []float64{startTT, greatestTT, endTT}
	for index := 1; ; index++ {
		tt := startTT + float64(index)*stepDays
		if tt >= endTT {
			break
		}
		times = append(times, tt)
	}
	sort.Float64s(times)
	return uniqueOccultationPathTimes(times)
}

func refineOccultationPathSpacing(
	points []OccultationPathPoint,
	star StarCoordinate,
	targetSpacingKM float64,
	location *time.Location,
) ([]OccultationPathPoint, error) {
	if len(points) < 2 || targetSpacingKM <= 0 {
		return points, nil
	}
	refined := make([]OccultationPathPoint, 0, len(points))
	refined = append(refined, points[0])
	widthAt := func(tt float64) (float64, bool) {
		_, _, width, ok := starOccultationPathLimitsAndWidthAt(tt, star)
		return width, ok
	}
	for i := 1; i < len(points); i++ {
		segmentStart := len(refined) - 1
		var err error
		refined, err = appendOccultationPathSegment(refined, points[i-1], points[i], star, targetSpacingKM, location, 0)
		if err != nil {
			return nil, err
		}
		refineOccultationPathWidths(refined[segmentStart:], widthAt)
	}
	return refined, nil
}

func appendOccultationPathSegment(
	points []OccultationPathPoint,
	start, end OccultationPathPoint,
	star StarCoordinate,
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
	startTT := centerTimeTT(start.Time)
	endTT := centerTimeTT(end.Time)
	midTT := (startTT + endTT) / 2
	mid, ok := starOccultationPathCenterPointWithoutWidth(midTT, star, location)
	if !ok {
		return append(points, end), nil
	}
	mid.WidthKM = (start.WidthKM + end.WidthKM) / 2
	var err error
	points, err = appendOccultationPathSegment(points, start, mid, star, targetSpacingKM, location, depth+1)
	if err != nil {
		return nil, err
	}
	return appendOccultationPathSegment(points, mid, end, star, targetSpacingKM, location, depth+1)
}

func uniqueOccultationPathTimes(times []float64) []float64 {
	if len(times) < 2 {
		return times
	}
	unique := times[:1]
	for _, tt := range times[1:] {
		if math.Abs(tt-unique[len(unique)-1]) <= 1e-10 {
			continue
		}
		unique = append(unique, tt)
	}
	return unique
}

type occultationPathFrame struct {
	moon         occultationPathVector
	axis         occultationPathVector
	first        occultationPathVector
	second       occultationPathVector
	moonRadius   float64
	targetRadius float64
}

type occultationPathEndpoint struct {
	point OccultationPathPoint
	valid bool
}

func (f occultationPathFrame) moonProjectionX() float64 { return occultationPathDot(f.moon, f.first) }
func (f occultationPathFrame) moonProjectionY() float64 { return occultationPathDot(f.moon, f.second) }

func starOccultationPathFrameAt(tt float64, star StarCoordinate) (occultationPathFrame, bool) {
	moonRA, moonDec := HMoonGeocentricApparentRaDecN(tt, -1)
	moonDistance := HMoonAwayN(tt, -1)
	if !finite(moonRA) || !finite(moonDec) || !finite(moonDistance) || moonDistance <= 0 {
		return occultationPathFrame{}, false
	}
	moon := occultationPathRaDecVector(moonRA, moonDec, moonDistance)
	starRA, starDec := starApparentRaDecGeocentric(tt, star)
	starDirection := occultationPathRaDecVector(starRA, starDec, 1)
	axis := occultationPathScale(starDirection, -1)
	if star.ParallaxMas > 0 {
		distanceAU := 206264806.247 / star.ParallaxMas
		target := occultationPathRaDecVector(starRA, starDec, distanceAU*occultationPathAstronomicalUnitKM)
		axis = occultationPathScale(occultationPathSub(target, moon), -1)
	}
	axis = occultationPathUnit(axis)
	north := occultationPathVector{z: 1}
	first := occultationPathCross(north, axis)
	if occultationPathNorm(first) < 1e-12 {
		first = occultationPathCross(occultationPathVector{x: 1}, axis)
	}
	first = occultationPathUnit(first)
	second := occultationPathUnit(occultationPathCross(axis, first))
	return occultationPathFrame{
		moon:       moon,
		axis:       axis,
		first:      first,
		second:     second,
		moonRadius: MoonSemidiameter(tt) * math.Pi / (180 * 3600),
	}, true
}

func starOccultationPathHasBoundary(frame occultationPathFrame) bool {
	_, _, ok := occultationPathBoundaryTangent(frame)
	return ok
}

func starOccultationPathBoundaryEndpoint(tt float64, star StarCoordinate, location *time.Location, direction int) occultationPathEndpoint {
	if _, ok := starOccultationPathFrameAt(tt, star); !ok {
		return occultationPathEndpoint{}
	}
	for offset := 0; offset <= 3; offset++ {
		candidateTT := tt + float64(direction)*float64(offset)*0.5/86400.0
		candidateFrame, candidateOK := starOccultationPathFrameAt(candidateTT, star)
		if !candidateOK {
			continue
		}
		vector, _, valid := occultationPathBoundaryTangent(candidateFrame)
		if valid {
			return occultationPathEndpoint{point: occultationPathPointFromVector(candidateTT, vector, 0, location), valid: true}
		}
	}
	return occultationPathEndpoint{}
}

func starOccultationPathCenterPoint(tt float64, star StarCoordinate, location *time.Location) (OccultationPathPoint, bool) {
	frame, ok := starOccultationPathFrameAt(tt, star)
	if !ok {
		return OccultationPathPoint{}, false
	}
	point, _, ok := occultationEarthLineIntersection(frame.moon, frame.axis)
	if !ok {
		return OccultationPathPoint{}, false
	}
	width := 0.0
	if _, _, tangentWidth, limitsOK := starOccultationPathLimitsAndWidthAt(tt, star); limitsOK {
		width = tangentWidth
	}
	return occultationPathPointFromVectorWithMoon(tt, point, width, frame.moon, location), true
}

func starOccultationPathCenterPointWithoutWidth(tt float64, star StarCoordinate, location *time.Location) (OccultationPathPoint, bool) {
	frame, ok := starOccultationPathFrameAt(tt, star)
	if !ok {
		return OccultationPathPoint{}, false
	}
	point, _, ok := occultationEarthLineIntersection(frame.moon, frame.axis)
	if !ok {
		return OccultationPathPoint{}, false
	}
	return occultationPathPointFromVectorWithMoon(tt, point, 0, frame.moon, location), true
}

func starOccultationPathLimitsAndWidthAt(tt float64, star StarCoordinate) (occultationPathVector, occultationPathVector, float64, bool) {
	frameAt := func(candidateTT float64) (occultationPathFrame, bool) {
		return starOccultationPathFrameAt(candidateTT, star)
	}
	return occultationPathLimitsAndWidthForFrame(tt, frameAt)
}

func occultationPathScannedLimitsAtFrame(tt float64, frame occultationPathFrame) (occultationPathVector, occultationPathVector, bool) {
	var northern, southern occultationPathVector
	northLatitude := math.Inf(-1)
	southLatitude := math.Inf(1)
	consider := func(vector occultationPathVector) {
		_, latitude := occultationPathGeodetic(tt, vector)
		if latitude > northLatitude {
			northLatitude = latitude
			northern = vector
		}
		if latitude < southLatitude {
			southLatitude = latitude
			southern = vector
		}
	}
	if tangentPoint, tangentTheta, tangentOK := occultationPathBoundaryTangent(frame); tangentOK {
		consider(tangentPoint)
		if leftTheta, rightTheta, intervalOK := occultationPathBoundaryThetaInterval(frame, tangentTheta); intervalOK {
			const intervalSamples = 128
			for i := 0; i <= intervalSamples; i++ {
				theta := leftTheta + (rightTheta-leftTheta)*float64(i)/intervalSamples
				if vector, _, ok := occultationPathBoundaryVector(frame, theta); ok {
					consider(vector)
				}
			}
		}
	}
	for i := 0; i < occultationPathBoundaryScanPoints; i++ {
		vector, _, ok := occultationPathBoundaryVector(frame, 2*math.Pi*float64(i)/float64(occultationPathBoundaryScanPoints))
		if !ok {
			continue
		}
		consider(vector)
	}
	return northern, southern, finite(northLatitude) && finite(southLatitude)
}

func occultationPathBoundaryThetaInterval(frame occultationPathFrame, centerTheta float64) (float64, float64, bool) {
	if !occultationPathBoundaryLineIntersects(frame, centerTheta) {
		return 0, 0, false
	}
	step := 2 * math.Pi / float64(occultationPathBoundaryScanPoints)
	findEdge := func(direction float64) (float64, bool) {
		inside := centerTheta
		for i := 1; i <= occultationPathBoundaryScanPoints; i++ {
			outside := centerTheta + direction*step*float64(i)
			if occultationPathBoundaryLineIntersects(frame, outside) {
				inside = outside
				continue
			}
			for iteration := 0; iteration < 56; iteration++ {
				mid := (inside + outside) / 2
				if occultationPathBoundaryLineIntersects(frame, mid) {
					inside = mid
				} else {
					outside = mid
				}
			}
			return inside, true
		}
		return 0, false
	}
	left, leftOK := findEdge(-1)
	right, rightOK := findEdge(1)
	return left, right, leftOK && rightOK && right > left
}

func occultationPathBoundaryLineIntersects(frame occultationPathFrame, theta float64) bool {
	discriminant, b, _, ok := occultationPathBoundaryLine(frame, theta)
	return ok && b < 0 && discriminant >= 0
}

func occultationPathBoundaryWidth(north, south occultationPathVector, ok bool) float64 {
	if !ok {
		return 0
	}
	return occultationPathNorm(occultationPathSub(north, south))
}

func occultationPathBoundaryVector(frame occultationPathFrame, theta float64) (occultationPathVector, float64, bool) {
	origin, direction, ok := occultationPathBoundaryRay(frame, theta)
	if !ok {
		return occultationPathVector{}, 0, false
	}
	return occultationEarthLineIntersection(origin, direction)
}

// occultationPathBoundaryRay 返回月缘圆柱或两球公切锥的一个母线 /
// occultationPathBoundaryRay returns one generator of the lunar-limb cylinder or a two-sphere common-tangent cone.
// targetRadius 带符号：正值表示异侧外切，负值表示同侧内切 /
// targetRadius is signed: positive for opposite-side outer tangency and negative for same-side inner tangency.
func occultationPathBoundaryRay(frame occultationPathFrame, theta float64) (occultationPathVector, occultationPathVector, bool) {
	moonDistance := occultationPathNorm(frame.moon)
	if moonDistance <= 0 || !finite(moonDistance) || !finite(frame.targetRadius) {
		return occultationPathVector{}, occultationPathVector{}, false
	}
	radial := occultationPathAdd(
		occultationPathScale(frame.first, math.Cos(theta)),
		occultationPathScale(frame.second, math.Sin(theta)),
	)
	sine, cosine := math.Sincos(frame.targetRadius)
	normal := occultationPathAdd(
		occultationPathScale(radial, cosine),
		occultationPathScale(frame.axis, -sine),
	)
	origin := occultationPathAdd(
		frame.moon,
		occultationPathScale(normal, moonDistance*math.Sin(frame.moonRadius)),
	)
	direction := occultationPathAdd(
		occultationPathScale(frame.axis, cosine),
		occultationPathScale(radial, sine),
	)
	return origin, occultationPathUnit(direction), true
}

// occultationPathBoundaryTangent 求边界锥与地球椭球的连续切点 /
// occultationPathBoundaryTangent finds the continuous tangency of a boundary cone with the Earth ellipsoid.
// 采样网格提供搜索盆地，再对线判别式做局部极大化以恢复网格点之间的切点 /
// A sample grid supplies a basin, while local maximization of the line discriminant recovers tangencies between grid points.
func occultationPathBoundaryTangent(frame occultationPathFrame) (occultationPathVector, float64, bool) {
	step := 2 * math.Pi / float64(occultationPathBoundaryScanPoints)
	bestTheta := 0.0
	bestDiscriminant := math.Inf(-1)
	for i := 0; i < occultationPathBoundaryScanPoints; i++ {
		theta := step * float64(i)
		discriminant, _, _, ok := occultationPathBoundaryLine(frame, theta)
		if ok && discriminant > bestDiscriminant {
			bestDiscriminant = discriminant
			bestTheta = theta
		}
	}
	if !finite(bestDiscriminant) {
		return occultationPathVector{}, 0, false
	}

	left := bestTheta - step
	right := bestTheta + step
	const goldenRatio = 0.6180339887498949
	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1, _, _, _ := occultationPathBoundaryLine(frame, x1)
	f2, _, _, _ := occultationPathBoundaryLine(frame, x2)
	for i := 0; i < 40; i++ {
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
	theta := (left + right) / 2
	discriminant, b, scale, ok := occultationPathBoundaryLine(frame, theta)
	if !ok {
		return occultationPathVector{}, 0, false
	}
	tolerance := 1e-12 * math.Max(scale, 1)
	if discriminant < -tolerance || b >= 0 {
		return occultationPathVector{}, 0, false
	}
	vector, _, valid := occultationPathBoundaryIntersection(frame, theta, tolerance)
	return vector, theta, valid
}

func occultationPathBoundaryLine(frame occultationPathFrame, theta float64) (discriminant, b, scale float64, ok bool) {
	origin, direction, rayOK := occultationPathBoundaryRay(frame, theta)
	if !rayOK {
		return 0, 0, 0, false
	}
	polarRatioSquared := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
	a := direction.x*direction.x + direction.y*direction.y + direction.z*direction.z/polarRatioSquared
	b = origin.x*direction.x + origin.y*direction.y + origin.z*direction.z/polarRatioSquared
	c := origin.x*origin.x + origin.y*origin.y + origin.z*origin.z/polarRatioSquared - occultationPathEarthEquatorialRadiusKM*occultationPathEarthEquatorialRadiusKM
	discriminant = b*b - a*c
	scale = math.Max(math.Abs(b*b), math.Abs(a*c))
	return discriminant, b, scale, a > 0 && finite(discriminant)
}

func occultationPathBoundaryIntersection(frame occultationPathFrame, theta, tolerance float64) (occultationPathVector, float64, bool) {
	origin, direction, ok := occultationPathBoundaryRay(frame, theta)
	if !ok {
		return occultationPathVector{}, 0, false
	}
	return occultationEarthLineIntersectionWithTolerance(origin, direction, tolerance)
}

func occultationPathPointFromVector(tt float64, vector occultationPathVector, width float64, location *time.Location) OccultationPathPoint {
	lon, lat := occultationPathGeodetic(tt, vector)
	moonRA, moonDec := moonTopocentricApparentRaDec(tt, Observer{Longitude: lon, Latitude: lat}, -1)
	return OccultationPathPoint{
		Time:         occultationTTToLocation(tt, location),
		Longitude:    lon,
		Latitude:     lat,
		MoonAltitude: occultationAltitude(tt, Observer{Longitude: lon, Latitude: lat}, moonRA, moonDec),
		WidthKM:      width,
	}
}

func occultationPathPointFromVectorWithMoon(
	tt float64,
	vector occultationPathVector,
	width float64,
	moon occultationPathVector,
	location *time.Location,
) OccultationPathPoint {
	lon, lat := occultationPathGeodetic(tt, vector)
	moonDistance := occultationPathNorm(moon)
	moonRA := normalizeRA(math.Atan2(moon.y, moon.x) * 180 / math.Pi)
	moonDec := math.Asin(math.Max(-1, math.Min(1, moon.z/moonDistance))) * 180 / math.Pi
	moonRA, moonDec = TopocentricRaDec(
		moonRA, moonDec, lat, lon, TD2UT(tt, false),
		moonDistance/occultationPathAstronomicalUnitKM, 0,
	)
	return OccultationPathPoint{
		Time:         occultationTTToLocation(tt, location),
		Longitude:    lon,
		Latitude:     lat,
		MoonAltitude: occultationAltitude(tt, Observer{Longitude: lon, Latitude: lat}, normalizeRA(moonRA), moonDec),
		WidthKM:      width,
	}
}

func centerTimeTT(value time.Time) float64 { return occultationTimeToTT(value) }

type occultationPathWidthFunc func(float64) (float64, bool)

func refineOccultationPathWidths(points []OccultationPathPoint, widthAt occultationPathWidthFunc) {
	if len(points) < 3 {
		return
	}
	cache := make(map[int]bool)
	var refine func(int, int, int)
	refine = func(left, right, depth int) {
		if right-left <= 1 || depth >= occultationPathMaxAdaptiveDepth {
			return
		}
		if right-left <= 4 {
			for index := left + 1; index < right; index++ {
				setOccultationPathExactWidth(points, index, widthAt, cache)
			}
			return
		}

		indices := uniqueOccultationPathWidthIndices(left, right)
		withinTolerance := true
		leftTime := points[left].Time
		duration := points[right].Time.Sub(leftTime).Seconds()
		for _, index := range indices {
			linear := (points[left].WidthKM + points[right].WidthKM) / 2
			if duration != 0 {
				fraction := points[index].Time.Sub(leftTime).Seconds() / duration
				linear = points[left].WidthKM + fraction*(points[right].WidthKM-points[left].WidthKM)
			}
			if !setOccultationPathExactWidth(points, index, widthAt, cache) ||
				math.Abs(points[index].WidthKM-linear) > occultationPathWidthToleranceKM {
				withinTolerance = false
			}
		}
		anchors := append([]int{left}, indices...)
		anchors = append(anchors, right)
		if withinTolerance {
			for index := 1; index < len(anchors); index++ {
				interpolateOccultationPathWidths(points, anchors[index-1], anchors[index])
			}
			return
		}
		for index := 1; index < len(anchors); index++ {
			refine(anchors[index-1], anchors[index], depth+1)
		}
	}
	refine(0, len(points)-1, 0)
}

func uniqueOccultationPathWidthIndices(left, right int) []int {
	indices := make([]int, 0, 3)
	for _, numerator := range []int{1, 2, 3} {
		index := left + (right-left)*numerator/4
		if index <= left || index >= right || len(indices) > 0 && index == indices[len(indices)-1] {
			continue
		}
		indices = append(indices, index)
	}
	return indices
}

func setOccultationPathExactWidth(
	points []OccultationPathPoint,
	index int,
	widthAt occultationPathWidthFunc,
	cache map[int]bool,
) bool {
	if ok, found := cache[index]; found {
		return ok
	}
	width, ok := widthAt(centerTimeTT(points[index].Time))
	if ok && finite(width) && width >= 0 {
		points[index].WidthKM = width
	} else {
		ok = false
	}
	cache[index] = ok
	return ok
}

func interpolateOccultationPathWidths(points []OccultationPathPoint, left, right int) {
	if right-left <= 1 {
		return
	}
	leftTime := points[left].Time
	duration := points[right].Time.Sub(leftTime).Seconds()
	for index := left + 1; index < right; index++ {
		fraction := float64(index-left) / float64(right-left)
		if duration != 0 {
			fraction = points[index].Time.Sub(leftTime).Seconds() / duration
		}
		points[index].WidthKM = points[left].WidthKM + fraction*(points[right].WidthKM-points[left].WidthKM)
	}
}

func occultationPathDistanceKM(a, b OccultationPathPoint) float64 {
	return occultationPathDistanceKMValues(a.Longitude, a.Latitude, b.Longitude, b.Latitude)
}

func occultationPathDistanceKMValues(lon1, lat1, lon2, lat2 float64) float64 {
	lat1 *= math.Pi / 180
	lat2 *= math.Pi / 180
	dLat := lat2 - lat1
	dLon := (lon2 - lon1) * math.Pi / 180
	dLon = math.Mod(dLon+math.Pi, 2*math.Pi)
	if dLon < 0 {
		dLon += 2 * math.Pi
	}
	dLon -= math.Pi
	h := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * occultationPathEarthEquatorialRadiusKM * math.Asin(math.Sqrt(math.Min(1, h)))
}

type occultationPathVector struct{ x, y, z float64 }

func occultationPathRaDecVector(ra, dec, distance float64) occultationPathVector {
	ra *= math.Pi / 180
	dec *= math.Pi / 180
	return occultationPathVector{
		x: distance * math.Cos(dec) * math.Cos(ra),
		y: distance * math.Cos(dec) * math.Sin(ra),
		z: distance * math.Sin(dec),
	}
}

func occultationPathAdd(a, b occultationPathVector) occultationPathVector {
	return occultationPathVector{x: a.x + b.x, y: a.y + b.y, z: a.z + b.z}
}

func occultationPathSub(a, b occultationPathVector) occultationPathVector {
	return occultationPathVector{x: a.x - b.x, y: a.y - b.y, z: a.z - b.z}
}

func occultationPathScale(a occultationPathVector, scalar float64) occultationPathVector {
	return occultationPathVector{x: a.x * scalar, y: a.y * scalar, z: a.z * scalar}
}

func occultationPathDot(a, b occultationPathVector) float64 {
	return a.x*b.x + a.y*b.y + a.z*b.z
}

func occultationPathCross(a, b occultationPathVector) occultationPathVector {
	return occultationPathVector{
		x: a.y*b.z - a.z*b.y,
		y: a.z*b.x - a.x*b.z,
		z: a.x*b.y - a.y*b.x,
	}
}

func occultationPathNorm(value occultationPathVector) float64 {
	return math.Sqrt(occultationPathDot(value, value))
}

func occultationPathUnit(value occultationPathVector) occultationPathVector {
	norm := occultationPathNorm(value)
	if norm <= 0 {
		return occultationPathVector{}
	}
	return occultationPathScale(value, 1/norm)
}

func occultationEarthLineIntersection(origin, direction occultationPathVector) (occultationPathVector, float64, bool) {
	return occultationEarthLineIntersectionWithTolerance(origin, direction, 0)
}

// occultationPathTrackReference 将影轴地面轨迹连续延伸到掠过阶段 /
// occultationPathTrackReference extends the shadow-axis ground track through grazing phases.
// 在地球外时，将椭球度量下最近点径向投影到表面，并在相切处与真实近侧交点连续连接 /
// Outside the Earth, the closest point under the ellipsoid metric is projected radially onto the surface and joined continuously to the real near-side intersection at tangency.
func occultationPathTrackReference(frame occultationPathFrame) (occultationPathVector, bool) {
	if point, _, ok := occultationEarthLineIntersection(frame.moon, frame.axis); ok {
		return point, true
	}

	polarRatioSquared := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
	a := frame.axis.x*frame.axis.x + frame.axis.y*frame.axis.y + frame.axis.z*frame.axis.z/polarRatioSquared
	b := frame.moon.x*frame.axis.x + frame.moon.y*frame.axis.y + frame.moon.z*frame.axis.z/polarRatioSquared
	if a <= 0 || !finite(a) || !finite(b) {
		return occultationPathVector{}, false
	}
	closest := occultationPathAdd(frame.moon, occultationPathScale(frame.axis, -b/a))
	metricRadius := math.Sqrt(closest.x*closest.x + closest.y*closest.y + closest.z*closest.z/polarRatioSquared)
	if metricRadius <= 1e-9 || !finite(metricRadius) {
		return occultationPathVector{}, false
	}
	return occultationPathScale(closest, occultationPathEarthEquatorialRadiusKM/metricRadius), true
}

func occultationEarthLineIntersectionWithTolerance(origin, direction occultationPathVector, tolerance float64) (occultationPathVector, float64, bool) {
	polarRatioSquared := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
	a := direction.x*direction.x + direction.y*direction.y + direction.z*direction.z/polarRatioSquared
	b := origin.x*direction.x + origin.y*direction.y + origin.z*direction.z/polarRatioSquared
	c := origin.x*origin.x + origin.y*origin.y + origin.z*origin.z/polarRatioSquared - occultationPathEarthEquatorialRadiusKM*occultationPathEarthEquatorialRadiusKM
	discriminant := b*b - a*c
	if discriminant < -tolerance || a <= 0 {
		return occultationPathVector{}, 0, false
	}
	if discriminant < 0 {
		discriminant = 0
	}
	root := math.Sqrt(discriminant)
	roots := [2]float64{(-b - root) / a, (-b + root) / a}
	chosen := math.Inf(1)
	for _, root := range roots {
		if root >= 0 && root < chosen {
			chosen = root
		}
	}
	if math.IsInf(chosen, 1) {
		return occultationPathVector{}, 0, false
	}
	return occultationPathAdd(origin, occultationPathScale(direction, chosen)), chosen, true
}

func occultationPathGeodetic(tt float64, vector occultationPathVector) (float64, float64) {
	ut := TD2UT(tt, false)
	gst := ApparentSiderealTime(ut) * 15
	longitude := normalizeLongitude(math.Atan2(vector.y, vector.x)*180/math.Pi - gst)
	latitude := math.Atan2(
		vector.z,
		occultationPathEarthPolarRatio*occultationPathEarthPolarRatio*math.Hypot(vector.x, vector.y),
	) * 180 / math.Pi
	return longitude, latitude
}

func occultationPathEarthFixedVector(tt float64, vector occultationPathVector) occultationPathVector {
	angle := ApparentSiderealTime(TD2UT(tt, false)) * 15 * math.Pi / 180
	cosAngle := math.Cos(angle)
	sinAngle := math.Sin(angle)
	return occultationPathVector{
		x: cosAngle*vector.x + sinAngle*vector.y,
		y: -sinAngle*vector.x + cosAngle*vector.y,
		z: vector.z,
	}
}

func normalizeLongitude(longitude float64) float64 {
	longitude = math.Mod(longitude+180, 360)
	if longitude < 0 {
		longitude += 360
	}
	return longitude - 180
}

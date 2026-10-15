package basic

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/starainrt/astro/internal/geodata"
)

const (
	// occultationPathThetaIntervalCacheMaximumEntries 限制每个 frame 记忆的中心角数量；一次求解只会用到个位数个。
	// occultationPathThetaIntervalCacheMaximumEntries bounds the memory of center angles per frame; a solve uses only a handful.
	occultationPathThetaIntervalCacheMaximumEntries = 8
	occultationPathDefaultStepDays                  = 1.0 / 1440.0
	occultationPathMinStepDays                      = 1.0 / 86400.0
	occultationPathMaxSampleCount                   = 30000
	occultationPathMaxAdaptiveDepth                 = 20
	occultationPathBoundarySpacingKM                = 500.0
	occultationPathContourSpacingKM                 = 40.0
	occultationPathBoundaryMergeKM                  = 1.0
	occultationPathVelocityStepDays                 = 1.0 / 1440.0
	occultationPathBoundaryScanPoints               = 720
	occultationPathRootToleranceDays                = occultationEventSelectionToleranceDays
	occultationPathRangeStepDays                    = 5.0 / 1440.0
	occultationPathSearchSpanDays                   = 2.0
	occultationPathWidthToleranceKM                 = 0.005
	occultationPathMaxOutputPointCount              = 2000000
	occultationPathFootprintPointBudget             = 2048
	occultationPathBoundaryBranchJumpKM             = 120.0
	occultationPathBoundaryBranchSpeedKMPerSecond   = 10.0

	occultationPathEarthPolarRatio    = 0.99664719
	occultationPathAstronomicalUnitKM = 149597870.7
)

// FindStarOccultationPathsWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func FindStarOccultationPathsWithInstance(start, end time.Time, star StarCoordinate, options OccultationPathOptions, inst *Instance) ([]StarOccultationPath, error) {
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
		0, inst,
	)
	paths := make([]StarOccultationPath, 0, len(candidates))
	for _, seedTT := range candidates {
		path, ok, err := starOccultationPathAtSeed(seedTT, star, options, start, end, inst)
		if err != nil {
			return nil, err
		}
		if !ok {
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
	if options.Algorithm == "" {
		options.Algorithm = OccultationPathAlgorithmOptimized
	}
	if options.Step <= 0 {
		options.Step = time.Duration(occultationPathDefaultStepDays * float64(24*time.Hour))
	}
	if float64(options.Step)/float64(24*time.Hour) < occultationPathMinStepDays {
		options.Step = time.Second
	}
	if options.TargetSpacingKM <= 0 || math.IsNaN(options.TargetSpacingKM) || math.IsInf(options.TargetSpacingKM, 0) {
		options.TargetSpacingKM = 0
	}
	if len(options.GreatestTimeValues) > 0 {
		values := make([]float64, 0, len(options.GreatestTimeValues))
		for _, value := range options.GreatestTimeValues {
			if !finite(value) {
				continue
			}
			duplicate := false
			for _, existing := range values {
				if math.Abs(existing-value) <= 1e-9 {
					duplicate = true
					break
				}
			}
			if !duplicate {
				values = append(values, value)
			}
		}
		sort.Float64s(values)
		options.GreatestTimeValues = values
	}
	return options
}

func starOccultationPathAtSeed(
	seedTT float64,
	star StarCoordinate,
	options OccultationPathOptions,
	selectionStart, selectionEnd time.Time, instances ...*Instance) (StarOccultationPath, bool, error) {
	inst := instanceValue(instances)

	location := selectionStart.Location()
	cache := newStarOccultationEventCache(star, inst)
	cache.prepareLocalEphemeris(seedTT)
	frameAt := cache.frameAt
	candidateFrameAt := cache.candidateFrameAt
	searchStart := seedTT - occultationPathSearchSpanDays
	searchEnd := seedTT + occultationPathSearchSpanDays
	outerStart, outerEnd, ok := starOccultationPathWindowWithCandidateFrames(
		seedTT, searchStart, searchEnd, candidateFrameAt, frameAt, false,
	)
	if !ok {
		return StarOccultationPath{}, false, nil
	}

	centerStart, centerEnd, hasCenter := starOccultationPathWindowWithCandidateFrames(
		seedTT, searchStart, searchEnd, candidateFrameAt, frameAt, true,
	)
	// Global markers retain the same full-term ephemerides as event-only
	// queries. The optimized path cache is selected after these markers.
	greatestTT := starOccultationPathGreatestWithFrame(seedTT, outerStart, outerEnd, frameAt)
	greatest, greatestOK := starOccultationPathCenterPointWithFrame(greatestTT, frameAt, location)
	if !greatestOK {
		if hasCenter {
			greatestTT = math.Max(centerStart, math.Min(centerEnd, greatestTT))
			greatest, greatestOK = starOccultationPathCenterPointWithFrame(greatestTT, frameAt, location)
		}
	}
	if !greatestOK {
		// For a non-central path the shadow axis misses the ellipsoid. Greatest
		// is the nearest point on the ellipsoid to that axis, not an outer
		// contact tangent; the latter is a band edge and shifts the marker.
		greatest, greatestOK = occultationPathTrackPointForFrame(greatestTT, frameAt, location)
	}
	if !greatestOK {
		return StarOccultationPath{}, false, nil
	}
	if !occultationTimeInSelectionWindow(greatest.Time, selectionStart, selectionEnd) {
		return StarOccultationPath{}, false, nil
	}
	if occultationPathEstimatedPointCount(
		outerStart, outerEnd, centerStart, centerEnd, hasCenter,
		0, 0, false, greatestTT, options,
	) > occultationPathMaxOutputPointCount {
		return StarOccultationPath{}, false, ErrOccultationPathSamplingLimit
	}

	start := starOccultationPathBoundaryEndpointWithFrame(outerStart, frameAt, location, 1, inst)
	end := starOccultationPathBoundaryEndpointWithFrame(outerEnd, frameAt, location, -1, inst)
	if !start.valid || !end.valid {
		return StarOccultationPath{}, false, nil
	}
	exactFrameAt := frameAt
	if options.Algorithm != OccultationPathAlgorithmExact {
		optimized := newStarOccultationEventCache(star, inst)
		optimized.preparePathEphemeris(seedTT, options.Algorithm)
		if optimized.local.dense {
			cache = optimized
			frameAt = cache.frameAt
		}
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
	centerLine, northern, southern, err := starOccultationPathSamplesWithFrame(
		outerStart,
		outerEnd,
		centerStart,
		centerEnd,
		hasCenter,
		greatestTT,
		frameAt,
		options,
		location,
	)
	if err != nil {
		return StarOccultationPath{}, false, err
	}
	if cache.local.dense {
		correctOccultationCenterWidths(centerLine, exactFrameAt)
	}
	path.CenterLine = centerLine
	path.NorthernLimit = occultationPathWithEndpoints(start.point, end.point, northern)
	path.SouthernLimit = occultationPathWithEndpoints(start.point, end.point, southern)
	path.NorthernLimit = occultationStationCorrectLimitSeries(
		path.NorthernLimit, frameAt, cache.riseSetContextAt, false, location,
	)
	path.SouthernLimit = occultationStationCorrectLimitSeries(
		path.SouthernLimit, frameAt, cache.riseSetContextAt, false, location,
	)
	var northApplied, southApplied bool
	path.NorthernLimit, northApplied = occultationLimbCorrectLimitSeries(path.NorthernLimit, cache.riseSetContextAt, options.LunarLimb, location, inst)
	path.SouthernLimit, southApplied = occultationLimbCorrectLimitSeries(path.SouthernLimit, cache.riseSetContextAt, options.LunarLimb, location, inst)
	path.LunarLimbApplied = northApplied || southApplied
	path.GreatestLimitSeparationKM, _ = occultationPathLimitSeparations(
		path.NorthernLimit, path.SouthernLimit, greatestTT,
	)
	if !options.DisableFootprints {
		path.Footprints = planetOccultationFootprints(outerStart, outerEnd, greatestTT, frameAt, options, location)
	}
	path.RiseSetCurves = occultationRiseSetCurvesWithCache(
		outerStart, outerEnd, greatestTT, options, location, cache.riseSetCache,
	)
	path.GreatestTimeContours = occultationGreatestTimeContours(
		occultationGreatestTimeLevels(options, outerStart, outerEnd), outerStart, outerEnd, cache.riseSetCache,
		[][]OccultationPathPoint{path.CenterLine, path.NorthernLimit, path.SouthernLimit},
		false, location,
	)
	bandContourTimes := occultationRiseSetEndpointTimes(path.RiseSetCurves)
	path.BandContours = occultationContactBandContoursWithAdditionalTimes(
		start.point, end.point, outerStart, outerEnd, greatestTT, frameAt, options, location, bandContourTimes,
	)
	if options.DisableFootprints {
		path.BandFootprints = planetOccultationBandFootprints(
			outerStart, outerEnd, greatestTT, frameAt, location, bandContourTimes,
		)
		if options.IncludeFootprintTimeline {
			path.Footprints = planetOccultationTimelineFootprints(
				outerStart, outerEnd, greatestTT, frameAt, options, location,
			)
		}
	}
	if len(path.Footprints) > 0 {
		path.Footprints = occultationStationCorrectFootprintEdges(
			path.Footprints, frameAt, cache.riseSetContextAt, false, location,
		)
	}
	if len(path.BandFootprints) > 0 {
		path.BandFootprints = occultationStationCorrectFootprintEdges(
			path.BandFootprints, frameAt, cache.riseSetContextAt, false, location,
		)
	}
	path.BandContours = occultationStationCorrectContours(
		path.BandContours, path.RiseSetCurves, cache.riseSetCache, location,
	)
	path.VisibilityContours = occultationStationVisibilityEnvelopeContours(
		path.RiseSetCurves, cache.riseSetCache, location,
	)
	return path, true, nil
}

func occultationPathEstimatedPointCount(
	outerStart, outerEnd, centerStart, centerEnd float64,
	hasCenter bool,
	totalStart, totalEnd float64,
	hasTotal bool,
	greatestTT float64,
	options OccultationPathOptions,
) int {
	stepDays := float64(options.Step) / float64(24*time.Hour)
	if stepDays <= 0 {
		stepDays = occultationPathDefaultStepDays
	}
	estimate := 0
	add := func(value int) {
		estimate = occultationPathAccumulatePointEstimate(estimate, value)
	}
	add(3 * len(occultationPathSampleTimes(outerStart, outerEnd, greatestTT, stepDays)))
	contourStepDays := occultationPathContourStepDays(options)
	add(2 * len(occultationPathSampleTimes(outerStart, outerEnd, greatestTT, contourStepDays)))
	if hasCenter {
		add(3 * len(occultationPathSampleTimes(centerStart, centerEnd, greatestTT, stepDays)))
	}
	if hasTotal {
		add(3 * len(occultationPathSampleTimes(totalStart, totalEnd, greatestTT, stepDays)))
		add(2 * len(occultationPathSampleTimes(totalStart, totalEnd, greatestTT, contourStepDays)))
	}
	footprintStep := planetOccultationFootprintSampleStepDays(options)
	footprintLimit := planetOccultationFootprintMaxSamples
	if options.DisableFootprints && !options.IncludeFootprintTimeline {
		footprintStep = planetOccultationBandSampleStepDays()
		footprintLimit = planetOccultationBandMaxSamples
	}
	add(len(occultationPathSampleTimesWithLimit(outerStart, outerEnd, greatestTT, footprintStep, footprintLimit)) * occultationPathFootprintPointBudget)
	if options.DisableFootprints && options.IncludeFootprintTimeline {
		add(len(planetOccultationBandSampleTimes(
			outerStart, outerEnd, greatestTT, planetOccultationBandSampleStepDays(), planetOccultationBandMaxSamples,
		)) * occultationPathFootprintPointBudget)
	}
	if hasTotal {
		add(len(occultationPathSampleTimesWithLimit(totalStart, totalEnd, greatestTT, footprintStep, footprintLimit)) * occultationPathFootprintPointBudget)
		if options.DisableFootprints && options.IncludeFootprintTimeline {
			add(len(planetOccultationBandSampleTimes(
				totalStart, totalEnd, greatestTT, planetOccultationBandSampleStepDays(), planetOccultationBandMaxSamples,
			)) * occultationPathFootprintPointBudget)
		}
	}
	if !options.DisableRiseSet {
		riseSetStep := 5 * time.Minute
		if options.RiseSetStep > 0 {
			riseSetStep = options.RiseSetStep
		}
		riseSetCount := len(occultationPathSampleTimes(
			outerStart, outerEnd, greatestTT,
			float64(riseSetStep)/float64(24*time.Hour),
		))
		add(6 * riseSetCount)
	}
	return estimate
}

func occultationPathAccumulatePointEstimate(estimate, value int) int {
	if value <= 0 || estimate > occultationPathMaxOutputPointCount {
		return estimate
	}
	if value > occultationPathMaxOutputPointCount-estimate {
		return occultationPathMaxOutputPointCount + 1
	}
	return estimate + value
}

func occultationPathWithEndpoints(start, end OccultationPathPoint, points []OccultationPathPoint) []OccultationPathPoint {
	result := make([]OccultationPathPoint, 0, len(points)+2)
	result = append(result, start)
	for _, point := range points {
		if point.Time.After(result[len(result)-1].Time) && point.Time.Before(end.Time) {
			result = append(result, point)
		}
	}
	return append(result, end)
}

// occultationPathLimitSeparations 在同一时刻的南北限采样对上填写地面间距，并返回最接近掩甚的那一对的间距。
func occultationPathLimitSeparations(
	north, south []OccultationPathPoint,
	greatestTT float64,
) (float64, bool) {
	if len(north) == 0 || len(north) != len(south) {
		return 0, false
	}
	separation, found := 0.0, false
	nearestDelta := math.Inf(1)
	for index := range north {
		if !north[index].Time.Equal(south[index].Time) {
			continue
		}
		value := occultationPathDistanceKM(north[index], south[index])
		if !finite(value) || value < 0 {
			continue
		}
		north[index].LimitSeparationKM = value
		south[index].LimitSeparationKM = value
		if delta := math.Abs(centerTimeTT(north[index].Time) - greatestTT); delta < nearestDelta {
			nearestDelta, separation, found = delta, value, true
		}
	}
	return separation, found
}

func occultationContactBandContours(
	northern, southern []OccultationPathPoint,
) [][]OccultationPathPoint {
	contours := make([][]OccultationPathPoint, 0, 4)
	for _, source := range [][]OccultationPathPoint{northern, southern} {
		for _, sampleRange := range occultationContinuousBoundaryRanges(source) {
			if sampleRange.end-sampleRange.start < 2 {
				continue
			}
			contour := append([]OccultationPathPoint(nil), source[sampleRange.start:sampleRange.end]...)
			contours = append(contours, contour)
		}
	}
	return contours
}

type occultationPathSampleRange struct {
	start int
	end   int
}

func occultationContinuousBoundaryRanges(points []OccultationPathPoint) []occultationPathSampleRange {
	if len(points) == 0 {
		return nil
	}
	ranges := make([]occultationPathSampleRange, 0, 4)
	start := 0
	for index := 1; index < len(points); index++ {
		if !occultationPathBoundaryBranchChanged(points[index-1], points[index]) {
			continue
		}
		ranges = append(ranges, occultationPathSampleRange{start: start, end: index})
		start = index
	}
	return append(ranges, occultationPathSampleRange{start: start, end: len(points)})
}

func occultationPathBoundaryBranchChanged(first, second OccultationPathPoint) bool {
	distance := occultationPathDistanceKM(first, second)
	if distance <= occultationPathBoundaryBranchJumpKM {
		return false
	}
	duration := math.Abs(second.Time.Sub(first.Time).Seconds())
	return duration == 0 || distance/duration > occultationPathBoundaryBranchSpeedKMPerSecond
}

func occultationContactBandContoursWithAdditionalTimes(
	start, end OccultationPathPoint,
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
	additionalTimes []float64,
) [][]OccultationPathPoint {
	stepDays := occultationPathContourStepDays(options)
	northern, southern := occultationPathBoundaryContourSamplesForFrame(
		startTT, endTT, greatestTT, frameAt, stepDays, location, additionalTimes,
	)
	return occultationContactBandContours(
		occultationPathWithEndpoints(start, end, northern),
		occultationPathWithEndpoints(start, end, southern),
	)
}

func occultationPathContourStepDays(options OccultationPathOptions) float64 {
	stepDays := float64(options.Step) / float64(24*time.Hour)
	if stepDays <= 0 {
		stepDays = occultationPathDefaultStepDays
	}
	if options.DisableRiseSet {
		return stepDays
	}
	riseSetStep := options.RiseSetStep
	if riseSetStep <= 0 {
		riseSetStep = 5 * time.Minute
	}
	riseSetStepDays := float64(riseSetStep) / float64(24*time.Hour)
	if riseSetStepDays > 0 && riseSetStepDays < stepDays {
		stepDays = riseSetStepDays
	}
	return stepDays
}

func starOccultationPathWindow(seedTT, startTT, endTT float64, star StarCoordinate, center bool, instances ...*Instance) (float64, float64, bool) {
	inst := instanceValue(instances)

	return starOccultationPathWindowWithFrame(seedTT, startTT, endTT, func(tt float64) (occultationPathFrame, bool) {
		return starOccultationPathFrameAt(tt, star, inst)
	}, center)
}

func starOccultationPathWindowWithFrame(
	seedTT, startTT, endTT float64,
	frameAt occultationPathFrameFunc,
	center bool,
) (float64, float64, bool) {
	return occultationPathWindowWithCandidateFrames(
		seedTT, startTT, endTT, frameAt, frameAt, center,
		starOccultationPathHasBoundary,
	)
}

func starOccultationPathWindowWithCandidateFrames(
	seedTT, startTT, endTT float64,
	candidateFrameAt, exactFrameAt occultationPathFrameFunc,
	center bool,
) (float64, float64, bool) {
	return occultationPathWindowWithCandidateFrames(
		seedTT, startTT, endTT, candidateFrameAt, exactFrameAt, center,
		starOccultationPathHasBoundary,
	)
}

func occultationPathWindowWithCandidateFrames(
	seedTT, startTT, endTT float64,
	candidateFrameAt, exactFrameAt occultationPathFrameFunc,
	center bool,
	hasBoundary func(occultationPathFrame) bool,
) (float64, float64, bool) {
	predicate := func(frameAt occultationPathFrameFunc, tt float64) bool {
		frame, ok := frameAt(tt)
		if !ok {
			return false
		}
		if center {
			_, _, ok = occultationEarthLineIntersection(frame.moon, frame.axis)
			return ok
		}
		return hasBoundary(frame)
	}
	engine := occultationMovingDiskEngine()
	return engine.window(
		seedTT, startTT, endTT,
		func(tt float64) bool { return predicate(candidateFrameAt, tt) },
		func(tt float64) bool { return predicate(exactFrameAt, tt) },
	)
}

func starOccultationPathGreatest(seedTT, startTT, endTT float64, star StarCoordinate, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return starOccultationPathGreatestWithFrame(seedTT, startTT, endTT, func(tt float64) (occultationPathFrame, bool) {
		return starOccultationPathFrameAt(tt, star, inst)
	})
}

func starOccultationPathGreatestWithFrame(
	seedTT, startTT, endTT float64,
	frameAt occultationPathFrameFunc,
) float64 {
	return starOccultationPathGreatestWithIterations(seedTT, startTT, endTT, frameAt, 56)
}

func starOccultationPathGreatestWithIterations(
	seedTT, startTT, endTT float64,
	frameAt occultationPathFrameFunc,
	iterations int,
) float64 {
	return occultationMovingDiskEngine().greatest(
		seedTT, startTT, endTT,
		func(tt float64) (float64, bool) {
			frame, ok := frameAt(tt)
			if !ok {
				return 0, false
			}
			return math.Hypot(frame.moonProjectionX(), frame.moonProjectionY()), true
		}, iterations,
	)
}

func starOccultationPathImpactWithFrame(tt float64, frameAt occultationPathFrameFunc) float64 {
	frame, ok := frameAt(tt)
	if !ok {
		return math.Inf(1)
	}
	return math.Hypot(frame.moonProjectionX(), frame.moonProjectionY())
}

func starOccultationPathSamplesWithFrame(
	outerStartTT, outerEndTT float64,
	centerStartTT, centerEndTT float64,
	hasCenter bool,
	greatestTT float64,
	frameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
) ([]OccultationPathPoint, []OccultationPathPoint, []OccultationPathPoint, error) {
	var points []OccultationPathPoint
	if hasCenter {
		var err error
		points, err = starOccultationPathCenterSamplesWithFrame(centerStartTT, centerEndTT, greatestTT, frameAt, options, location)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	northern, southern := occultationPathBoundarySamplesForFrame(
		outerStartTT, outerEndTT, greatestTT, frameAt, occultationPathContourStepDays(options), location,
	)
	return points, northern, southern, nil
}

func starOccultationPathCenterSamplesWithFrame(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
) ([]OccultationPathPoint, error) {
	stepDays := float64(options.Step) / float64(24*time.Hour)
	times := occultationPathSampleTimes(startTT, endTT, greatestTT, stepDays)
	points := make([]OccultationPathPoint, 0, len(times))
	for _, tt := range times {
		point, ok := starOccultationPathCenterPointWithFrame(tt, frameAt, location)
		if ok && (len(points) == 0 || point.Time.After(points[len(points)-1].Time)) {
			points = append(points, point)
		}
	}
	if options.TargetSpacingKM > 0 {
		return refineOccultationPathSpacingWithFrame(points, frameAt, options.TargetSpacingKM, location)
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
	if maximumCount < 3 {
		maximumCount = 3
	}
	engine := occultationMovingDiskEngine()
	engine.maxSampleCount = maximumCount
	times, _ := engine.sampleTimes(startTT, endTT, greatestTT, stepDays)
	return times
}

func refineOccultationPathSpacingWithFrame(
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
		refined, err = appendOccultationPathSegmentWithFrame(refined, points[i-1], points[i], frameAt, targetSpacingKM, location, 0)
		if err != nil {
			return nil, err
		}
		refineOccultationPathWidths(refined[segmentStart:], widthAt)
	}
	return refined, nil
}

func appendOccultationPathSegmentWithFrame(
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
	startTT := centerTimeTT(start.Time)
	endTT := centerTimeTT(end.Time)
	midTT := (startTT + endTT) / 2
	midTime := occultationTTToLocation(midTT, location)
	if !midTime.After(start.Time) || !midTime.Before(end.Time) {
		return append(points, end), nil
	}
	mid, ok := starOccultationPathCenterPointWithoutWidthWithFrame(midTT, frameAt, location)
	if !ok {
		return append(points, end), nil
	}
	mid.WidthKM = (start.WidthKM + end.WidthKM) / 2
	var err error
	points, err = appendOccultationPathSegmentWithFrame(points, start, mid, frameAt, targetSpacingKM, location, depth+1)
	if err != nil {
		return nil, err
	}
	return appendOccultationPathSegmentWithFrame(points, mid, end, frameAt, targetSpacingKM, location, depth+1)
}

func uniqueOccultationPathTimes(times []float64) []float64 {
	return movingDiskUniqueTimes(times)
}

type occultationPathFrame struct {
	moon         occultationPathVector
	axis         occultationPathVector
	first        occultationPathVector
	second       occultationPathVector
	moonRadius   float64
	targetRadius float64
	// boundary 挂在被事件缓存复用的 frame 上，记忆化只依赖 frame 的边界搜索；零值 frame 为 nil 时退化为直接计算。
	// boundary is attached to frames reused through the event cache and memoizes the boundary
	// searches that depend only on the frame; a zero frame keeps it nil and computes directly.
	boundary *occultationPathBoundaryCache
}

// occultationPathBoundaryCache 记忆化只依赖 frame 的两个边界搜索：切点与可见 θ 区间。
// 两者各自要扫描 720 个网格点，而同一 frame 会在多个调用点被反复查询。
// occultationPathBoundaryCache memoizes the two frame-only boundary searches (the tangency
// and the visible theta intervals). Each scans 720 grid points and the same frame is queried
// repeatedly from several call sites.
type occultationPathBoundaryCache struct {
	// boundaryOnce 用同一份 720 点判别式网格同时恢复切点与可见 θ 区间，避免两条路径各扫一遍。
	// boundaryOnce recovers both the tangency and the visible theta intervals from one 720-point
	// discriminant grid instead of scanning it once per path.
	boundaryOnce sync.Once
	tangentPoint occultationPathVector
	tangentTheta float64
	tangentOK    bool
	intervals    []occultationPathThetaInterval

	// intervalMu 保护按 centerTheta 复用的 θ 区间搜索；每个 frame 只会用到个位数个中心角。
	// intervalMu guards the theta-interval searches reused per center angle; a frame only ever
	// uses a handful of center angles.
	intervalMu   sync.Mutex
	intervalKeys []uint64
	intervalVals []occultationPathThetaIntervalResult
}

// occultationPathFrameGeometryEqual 比较两个 frame 的几何内容，忽略只用于记忆化的 boundary 指针。
// occultationPathFrameGeometryEqual compares the geometric content of two frames and ignores
// the memo-only boundary pointer.
func occultationPathFrameGeometryEqual(first, second occultationPathFrame) bool {
	return first.moon == second.moon &&
		first.axis == second.axis &&
		first.first == second.first &&
		first.second == second.second &&
		first.moonRadius == second.moonRadius &&
		first.targetRadius == second.targetRadius
}

// occultationPathThetaIntervalResult 是 occultationPathBoundaryThetaInterval 的缓存结果。
// occultationPathThetaIntervalResult is the cached result of occultationPathBoundaryThetaInterval.
type occultationPathThetaIntervalResult struct {
	left, right float64
	ok          bool
}

type starOccultationEphemerisState struct {
	moonRA, moonDec float64
	moonDistanceKM  float64
	starRA, starDec float64
	starDistanceKM  float64
	valid           bool
}

type starOccultationEventCache struct {
	inst *Instance

	star         StarCoordinate
	states       map[uint64]starOccultationEphemerisState
	frames       map[uint64]planetOccultationFrameCacheEntry
	riseSetCache *occultationRiseSetEvaluationCache
	local        *starOccultationLocalEphemeris
}

func newStarOccultationEventCache(star StarCoordinate, instances ...*Instance) *starOccultationEventCache {
	inst := instanceValue(instances)

	cache := &starOccultationEventCache{inst: inst,
		star:   star,
		states: make(map[uint64]starOccultationEphemerisState),
		frames: make(map[uint64]planetOccultationFrameCacheEntry),
	}
	cache.riseSetCache = newOccultationRiseSetEvaluationCacheWithCandidate(
		cache.riseSetContextAt,
		cache.candidateRiseSetContextAt,
	)
	return cache
}

func (cache *starOccultationEventCache) stateAt(tt float64) starOccultationEphemerisState {
	key := math.Float64bits(tt)
	if state, ok := cache.states[key]; ok {
		return state
	}
	if len(cache.states) >= planetOccultationEventCacheMaximumEntries {
		for cachedKey := range cache.states {
			delete(cache.states, cachedKey)
		}
		for cachedKey := range cache.frames {
			delete(cache.frames, cachedKey)
		}
	}
	var state starOccultationEphemerisState
	interpolated := false
	if cache.local != nil && cache.local.dense {
		state, interpolated = cache.local.stateAt(tt)
	}
	if !interpolated {
		state = starOccultationEphemerisStateAt(tt, cache.star, cache.inst)
	}
	cache.states[key] = state
	return state
}

func (cache *starOccultationEventCache) prepareLocalEphemeris(center float64) {
	if cache.local == nil {
		cache.local = newStarOccultationLocalEphemeris(center, cache.star, cache.inst)
	}
}

func (cache *starOccultationEventCache) candidateFrameAt(tt float64) (occultationPathFrame, bool) {
	if cache.local != nil {
		if state, ok := cache.local.stateAt(tt); ok {
			return starOccultationPathFrameFromState(state)
		}
	}
	return cache.frameAt(tt)
}

func (cache *starOccultationEventCache) frameAt(tt float64) (occultationPathFrame, bool) {
	key := math.Float64bits(tt)
	if entry, ok := cache.frames[key]; ok {
		return entry.frame, entry.ok
	}
	frame, ok := starOccultationPathFrameFromState(cache.stateAt(tt))
	if ok {
		frame.boundary = &occultationPathBoundaryCache{}
	}
	cache.frames[key] = planetOccultationFrameCacheEntry{frame: frame, ok: ok}
	return frame, ok
}

func (cache *starOccultationEventCache) riseSetContextAt(tt float64) occultationRiseSetContext {
	state := cache.stateAt(tt)
	return newOccultationRiseSetContext(
		tt, state.moonRA, state.moonDec, state.moonDistanceKM,
		state.starRA, state.starDec, state.starDistanceKM, 0,
	)
}

func (cache *starOccultationEventCache) candidateRiseSetContextAt(tt float64) occultationRiseSetContext {
	if cache.local != nil {
		moonXYZ, targetXYZ, ok := cache.local.vectorsAt(tt)
		if ok {
			return newOccultationRiseSetContextFromVectors(
				tt,
				moonXYZ,
				targetXYZ,
				cache.local.starDistanceKM() > 0,
				0,
			)
		}
	}
	return cache.riseSetContextAt(tt)
}

type occultationPathEndpoint struct {
	point OccultationPathPoint
	valid bool
}

func (f occultationPathFrame) moonProjectionX() float64 { return occultationPathDot(f.moon, f.first) }
func (f occultationPathFrame) moonProjectionY() float64 { return occultationPathDot(f.moon, f.second) }

func starOccultationPathFrameAt(tt float64, star StarCoordinate, instances ...*Instance) (occultationPathFrame, bool) {
	inst := instanceValue(instances)

	return starOccultationPathFrameFromState(starOccultationEphemerisStateAt(tt, star, inst))
}

func starOccultationEphemerisStateAt(tt float64, star StarCoordinate, instances ...*Instance) starOccultationEphemerisState {
	inst := instanceValue(instances)

	moonRA, moonDec := HMoonGeocentricApparentRaDecNWithInstance(tt, -1, inst)
	moonDistanceKM := HMoonAwayNWithInstance(tt, -1, inst)
	// 方向与距离必须来自同一次三维推进，否则视差修正会退回历元距离。
	starRA, starDec, starDistanceAU := starApparentRaDecDistanceGeocentric(tt, star, inst)
	starDistanceKM := starDistanceAU * occultationPathAstronomicalUnitKM
	return starOccultationEphemerisState{
		moonRA: moonRA, moonDec: moonDec, moonDistanceKM: moonDistanceKM,
		starRA: starRA, starDec: starDec, starDistanceKM: starDistanceKM,
		valid: finite(moonRA) && finite(moonDec) && finite(moonDistanceKM) && moonDistanceKM > 0 &&
			finite(starRA) && finite(starDec) && finite(starDistanceKM) && starDistanceKM >= 0,
	}
}

func starOccultationPathFrameFromState(state starOccultationEphemerisState) (occultationPathFrame, bool) {
	if !state.valid {
		return occultationPathFrame{}, false
	}
	moon := occultationPathRaDecVector(state.moonRA, state.moonDec, state.moonDistanceKM)
	starDirection := occultationPathRaDecVector(state.starRA, state.starDec, 1)
	axis := occultationPathScale(starDirection, -1)
	if state.starDistanceKM > 0 {
		target := occultationPathRaDecVector(state.starRA, state.starDec, state.starDistanceKM)
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
		moonRadius: math.Asin(moonEquatorialRadiusKM / state.moonDistanceKM),
	}, true
}

func starOccultationPathHasBoundary(frame occultationPathFrame) bool {
	_, _, ok := occultationPathBoundaryTangent(frame)
	return ok
}

func starOccultationPathBoundaryEndpointWithFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	direction int, instances ...*Instance) occultationPathEndpoint {
	inst := instanceValue(instances)

	if _, ok := frameAt(tt); !ok {
		return occultationPathEndpoint{}
	}
	for offset := 0; offset <= 3; offset++ {
		candidateTT := tt + float64(direction)*float64(offset)*0.5/86400.0
		candidateFrame, candidateOK := frameAt(candidateTT)
		if !candidateOK {
			continue
		}
		vector, _, valid := occultationPathBoundaryTangent(candidateFrame)
		if valid {
			return occultationPathEndpoint{point: occultationPathPointFromVector(candidateTT, vector, 0, location, inst), valid: true}
		}
	}
	return occultationPathEndpoint{}
}

func starOccultationPathCenterPoint(tt float64, star StarCoordinate, location *time.Location, instances ...*Instance) (OccultationPathPoint, bool) {
	inst := instanceValue(instances)

	return starOccultationPathCenterPointWithFrame(tt, func(candidateTT float64) (occultationPathFrame, bool) {
		return starOccultationPathFrameAt(candidateTT, star, inst)
	}, location)
}

func starOccultationPathCenterPointWithFrame(
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

func starOccultationPathCenterPointWithoutWidthWithFrame(
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

func occultationPathScannedLimitsAtFrame(tt float64, frame occultationPathFrame) (occultationPathVector, occultationPathVector, bool) {
	var northern, southern occultationPathVector
	northLatitude := math.Inf(-1)
	southLatitude := math.Inf(1)
	consider := func(vector occultationPathVector) {
		latitude := occultationPathGeodeticLatitude(vector)
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
			// 同 occultationPathFiniteCrossTrackExtrema：等差末样本与 rightTheta 差 1 ULP，
			// 掠射时端点会被判"无地面交点"而丢掉极值，端点原值必须补采一次。
			if vector, _, ok := occultationPathBoundaryVector(frame, rightTheta); ok {
				consider(vector)
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
	if cached, hit := occultationPathCachedThetaInterval(frame, centerTheta); hit {
		return cached.left, cached.right, cached.ok
	}
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
	result := occultationPathThetaIntervalResult{left: left, right: right, ok: leftOK && rightOK && right > left}
	occultationPathStoreThetaInterval(frame, centerTheta, result)
	return result.left, result.right, result.ok
}

// occultationPathCachedThetaInterval 查询 frame 上按 centerTheta 缓存的 θ 区间。
// occultationPathCachedThetaInterval looks up a theta interval cached on the frame.
func occultationPathCachedThetaInterval(
	frame occultationPathFrame,
	centerTheta float64,
) (occultationPathThetaIntervalResult, bool) {
	if frame.boundary == nil {
		return occultationPathThetaIntervalResult{}, false
	}
	key := math.Float64bits(centerTheta)
	frame.boundary.intervalMu.Lock()
	defer frame.boundary.intervalMu.Unlock()
	for index, cached := range frame.boundary.intervalKeys {
		if cached == key {
			return frame.boundary.intervalVals[index], true
		}
	}
	return occultationPathThetaIntervalResult{}, false
}

// occultationPathStoreThetaInterval 记住 frame 上某个 centerTheta 的 θ 区间结果。
// occultationPathStoreThetaInterval remembers one theta interval computed on a frame.
func occultationPathStoreThetaInterval(
	frame occultationPathFrame,
	centerTheta float64,
	result occultationPathThetaIntervalResult,
) {
	if frame.boundary == nil {
		return
	}
	key := math.Float64bits(centerTheta)
	frame.boundary.intervalMu.Lock()
	defer frame.boundary.intervalMu.Unlock()
	if len(frame.boundary.intervalKeys) >= occultationPathThetaIntervalCacheMaximumEntries {
		return
	}
	frame.boundary.intervalKeys = append(frame.boundary.intervalKeys, key)
	frame.boundary.intervalVals = append(frame.boundary.intervalVals, result)
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
	sineTheta, cosineTheta := math.Sincos(theta)
	radial := occultationPathAdd(
		occultationPathScale(frame.first, cosineTheta),
		occultationPathScale(frame.second, sineTheta),
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
	if frame.boundary != nil {
		frame.boundary.boundaryOnce.Do(func() { occultationPathBoundaryPrecompute(frame) })
		return frame.boundary.tangentPoint, frame.boundary.tangentTheta, frame.boundary.tangentOK
	}
	return occultationPathBoundaryTangentUncached(frame)
}

// occultationPathBoundaryPrecompute 用一份网格同时算好切点与可见 θ 区间。
// occultationPathBoundaryPrecompute derives both the tangency and the visible theta intervals
// from one discriminant grid.
func occultationPathBoundaryPrecompute(frame occultationPathFrame) {
	step := 2 * math.Pi / float64(occultationPathBoundaryScanPoints)
	discriminants := make([]float64, occultationPathBoundaryScanPoints)
	occultationPathBoundaryFillGrid(frame, discriminants)
	frame.boundary.tangentPoint, frame.boundary.tangentTheta, frame.boundary.tangentOK =
		occultationPathBoundaryTangentFromGrid(frame, step, discriminants)
	frame.boundary.intervals = occultationPathBoundaryThetaIntervalsFromGrid(frame, step, discriminants)
}

// occultationPathBoundaryTangentUncached 是切点搜索本体；调用方通过 frame 级缓存复用结果。
// occultationPathBoundaryTangentUncached is the tangency search itself; callers reuse it through the frame cache.
func occultationPathBoundaryTangentUncached(frame occultationPathFrame) (occultationPathVector, float64, bool) {
	step := 2 * math.Pi / float64(occultationPathBoundaryScanPoints)
	discriminants := make([]float64, occultationPathBoundaryScanPoints)
	occultationPathBoundaryFillGrid(frame, discriminants)
	return occultationPathBoundaryTangentFromGrid(frame, step, discriminants)
}

// occultationPathBoundaryTangentFromGrid 在已算好的网格上恢复切点。
// occultationPathBoundaryTangentFromGrid recovers the tangency from a prepared grid.
func occultationPathBoundaryTangentFromGrid(
	frame occultationPathFrame,
	step float64,
	discriminants []float64,
) (occultationPathVector, float64, bool) {
	bestTheta := 0.0
	bestDiscriminant := math.Inf(-1)
	for i, discriminant := range discriminants {
		if discriminant > bestDiscriminant {
			bestDiscriminant = discriminant
			bestTheta = step * float64(i)
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
	c := origin.x*origin.x + origin.y*origin.y + origin.z*origin.z/polarRatioSquared - geodata.EarthEquatorialRadiusKM*geodata.EarthEquatorialRadiusKM
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

func occultationPathPointFromVector(tt float64, vector occultationPathVector, width float64, location *time.Location, instances ...*Instance) OccultationPathPoint {
	inst := instanceValue(instances)

	lon, lat := occultationPathGeodetic(tt, vector)
	moonRA, moonDec := moonTopocentricApparentRaDec(tt, Observer{Longitude: lon, Latitude: lat}, -1, inst)
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
	return occultationPathPointFromVectorWithMoonSidereal(
		tt, vector, width, moon, ApparentSiderealTime(TT2UT1(tt))*15, location,
	)
}

func occultationPathPointFromVectorWithMoonSidereal(
	tt float64,
	vector occultationPathVector,
	width float64,
	moon occultationPathVector,
	siderealDegrees float64,
	location *time.Location,
) OccultationPathPoint {
	lon, lat := occultationPathGeodeticWithSidereal(vector, siderealDegrees)
	moonDistance := occultationPathNorm(moon)
	moonRA := normalizeRA(math.Atan2(moon.y, moon.x) * 180 / math.Pi)
	moonDec := math.Asin(math.Max(-1, math.Min(1, moon.z/moonDistance))) * 180 / math.Pi
	moonRA, moonDec = topocentricRaDecWithSidereal(
		moonRA, moonDec, lat, lon, siderealDegrees,
		moonDistance/occultationPathAstronomicalUnitKM, 0,
	)
	return OccultationPathPoint{
		Time:      occultationTTToLocation(tt, location),
		Longitude: lon,
		Latitude:  lat,
		MoonAltitude: occultationAltitudeWithSidereal(
			siderealDegrees, Observer{Longitude: lon, Latitude: lat}, normalizeRA(moonRA), moonDec,
		),
		WidthKM: width,
	}
}

func occultationPathDeduplicateBoundaryPoints(
	first, second []OccultationPathPoint,
) ([]OccultationPathPoint, []OccultationPathPoint) {
	if len(first) != len(second) || len(first) == 0 {
		return first, second
	}
	uniqueFirst := make([]OccultationPathPoint, 0, len(first))
	uniqueSecond := make([]OccultationPathPoint, 0, len(second))
	for index := range first {
		if len(uniqueFirst) > 0 && !first[index].Time.After(uniqueFirst[len(uniqueFirst)-1].Time) {
			continue
		}
		if len(uniqueFirst) > 0 && first[index].Time.Sub(uniqueFirst[len(uniqueFirst)-1].Time) < time.Second &&
			occultationPathDistanceKM(first[index], uniqueFirst[len(uniqueFirst)-1]) < occultationPathBoundaryMergeKM &&
			occultationPathDistanceKM(second[index], uniqueSecond[len(uniqueSecond)-1]) < occultationPathBoundaryMergeKM {
			uniqueFirst[len(uniqueFirst)-1] = first[index]
			uniqueSecond[len(uniqueSecond)-1] = second[index]
			continue
		}
		uniqueFirst = append(uniqueFirst, first[index])
		uniqueSecond = append(uniqueSecond, second[index])
	}
	return uniqueFirst, uniqueSecond
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
	return geodata.EarthEquatorialRadiusKM * geodata.CentralAngle(lon1, lat1, lon2, lat2)
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
	return occultationPathScale(closest, geodata.EarthEquatorialRadiusKM/metricRadius), true
}

func occultationEarthLineIntersectionWithTolerance(origin, direction occultationPathVector, tolerance float64) (occultationPathVector, float64, bool) {
	polarRatioSquared := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
	a := direction.x*direction.x + direction.y*direction.y + direction.z*direction.z/polarRatioSquared
	b := origin.x*direction.x + origin.y*direction.y + origin.z*direction.z/polarRatioSquared
	c := origin.x*origin.x + origin.y*origin.y + origin.z*origin.z/polarRatioSquared - geodata.EarthEquatorialRadiusKM*geodata.EarthEquatorialRadiusKM
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
	ut1 := TT2UT1(tt)
	return occultationPathGeodeticWithSidereal(vector, ApparentSiderealTime(ut1)*15)
}

func occultationPathGeodeticWithSidereal(
	vector occultationPathVector,
	siderealDegrees float64,
) (float64, float64) {
	longitude := geodata.NormalizeLongitude(math.Atan2(vector.y, vector.x)*180/math.Pi - siderealDegrees)
	return longitude, occultationPathGeodeticLatitude(vector)
}

func occultationPathGeodeticLatitude(vector occultationPathVector) float64 {
	return math.Atan2(
		vector.z,
		occultationPathEarthPolarRatio*occultationPathEarthPolarRatio*math.Hypot(vector.x, vector.y),
	) * 180 / math.Pi
}

func occultationPathEarthFixedVector(tt float64, vector occultationPathVector) occultationPathVector {
	return occultationPathEarthFixedVectorWithRotation(vector, occultationPathEarthRotationAt(tt))
}

type occultationPathEarthRotation struct {
	cosine float64
	sine   float64
}

func occultationPathEarthRotationAt(tt float64) occultationPathEarthRotation {
	angle := ApparentSiderealTime(TT2UT1(tt)) * 15 * math.Pi / 180
	return occultationPathEarthRotation{cosine: math.Cos(angle), sine: math.Sin(angle)}
}

func occultationPathEarthFixedVectorWithRotation(
	vector occultationPathVector,
	rotation occultationPathEarthRotation,
) occultationPathVector {
	return occultationPathVector{
		x: rotation.cosine*vector.x + rotation.sine*vector.y,
		y: -rotation.sine*vector.x + rotation.cosine*vector.y,
		z: vector.z,
	}
}

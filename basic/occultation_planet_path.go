package basic

import (
	"math"
	"sort"
	"time"

	"github.com/starainrt/astro/internal/geodata"
)

const (
	planetOccultationPathMaxTemporalSamples = 5000
	// A single dense path evaluates several independent temporal grids and
	// contour refinements. Keep all exact frames for that event so the shared
	// ephemeris is not recomputed every time the bounded cache rolls over.
	planetOccultationEventCacheMaximumEntries = 16384
)

type occultationPathFrameFunc func(float64) (occultationPathFrame, bool)

type planetOccultationEphemerisState struct {
	moonRA, moonDec     float64
	moonDistanceKM      float64
	planetRA, planetDec float64
	planetDistanceKM    float64
	valid               bool
}

type planetOccultationFrameCacheEntry struct {
	frame occultationPathFrame
	ok    bool
}

type planetOccultationEventCache struct {
	inst *Instance

	config            planetOccultationConfig
	states            map[uint64]planetOccultationEphemerisState
	outerFrames       map[uint64]planetOccultationFrameCacheEntry
	totalFrames       map[uint64]planetOccultationFrameCacheEntry
	riseSetCache      *occultationRiseSetEvaluationCache
	totalRiseSetCache *occultationRiseSetEvaluationCache
	local             *planetOccultationLocalEphemeris
}

func newPlanetOccultationEventCache(config planetOccultationConfig, instances ...*Instance) *planetOccultationEventCache {
	inst := instanceValue(instances)

	cache := &planetOccultationEventCache{inst: inst,
		config:      config,
		states:      make(map[uint64]planetOccultationEphemerisState),
		outerFrames: make(map[uint64]planetOccultationFrameCacheEntry),
		totalFrames: make(map[uint64]planetOccultationFrameCacheEntry),
	}
	cache.riseSetCache = newOccultationRiseSetEvaluationCacheWithCandidate(
		cache.riseSetContextAt,
		cache.candidateRiseSetContextAt,
	)
	cache.totalRiseSetCache = newOccultationRiseSetEvaluationCacheWithCandidate(
		cache.totalRiseSetContextAt,
		cache.candidateTotalRiseSetContextAt,
	)
	return cache
}

// FindPlanetOccultationPathsWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func FindPlanetOccultationPathsWithInstance(start, end time.Time, planet OccultationPlanet,
	options OccultationPathOptions, inst *Instance) ([]PlanetOccultationPath, error) {
	if err := validateOccultationTimeRange(start, end); err != nil {
		return nil, err
	}
	if err := planet.Validate(); err != nil {
		return nil, err
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}

	config, _ := planetOccultationConfigFor(planet, inst)
	options = normalizeOccultationPathOptions(options)
	startTT := occultationTimeToTT(start)
	endTT := occultationTimeToTT(end)
	candidateStartTT := startTT - occultationPathSearchSpanDays
	candidateEndTT := endTT + occultationPathSearchSpanDays
	candidates := planetOccultationCandidateGreatestTimes(
		candidateStartTT, candidateEndTT, planetOccultationDefaultStepDays, config, nil, 0, inst,
	)
	paths := make([]PlanetOccultationPath, 0, len(candidates))
	for _, seedTT := range candidates {
		path, ok, err := planetOccultationPathAtSeed(seedTT, config, options, start, end, inst)
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

func planetOccultationPathAtSeed(
	seedTT float64,
	config planetOccultationConfig,
	options OccultationPathOptions,
	selectionStart, selectionEnd time.Time, instances ...*Instance) (PlanetOccultationPath, bool, error) {
	inst := instanceValue(instances)

	location := selectionStart.Location()
	cache := newPlanetOccultationEventCache(config, inst)
	cache.prepareLocalEphemeris(seedTT)
	frameAt := cache.outerFrameAt
	totalFrameAt := cache.totalFrameAt
	candidateFrameAt := cache.candidateFrameAt
	candidateTotalFrameAt := cache.candidateTotalFrameAt
	searchStart := seedTT - occultationPathSearchSpanDays
	searchEnd := seedTT + occultationPathSearchSpanDays
	outerStart, outerEnd, ok := occultationPathWindowWithCandidateFrames(
		seedTT, searchStart, searchEnd, candidateFrameAt, frameAt, false,
		occultationPathFrameHasBoundary,
	)
	if !ok {
		return PlanetOccultationPath{}, false, nil
	}
	centerStart, centerEnd, hasCenter := occultationPathWindowWithCandidateFrames(
		seedTT, searchStart, searchEnd, candidateFrameAt, frameAt, true,
		occultationPathFrameHasBoundary,
	)
	// Use the local interpolated ephemeris to predict the maximum first, then
	// run the exact golden-section search in a bounded neighbourhood. Global
	// markers stay identical to event-only queries in both path branches.
	candidateGreatestTT := occultationPathGreatestForFrame(seedTT, outerStart, outerEnd, candidateFrameAt)
	exactSearchMarginDays := 0.10
	exactStart := math.Max(outerStart, candidateGreatestTT-exactSearchMarginDays)
	exactEnd := math.Min(outerEnd, candidateGreatestTT+exactSearchMarginDays)
	if exactEnd <= exactStart {
		exactStart, exactEnd = outerStart, outerEnd
	}
	greatestTT := occultationPathGreatestForFrame(candidateGreatestTT, exactStart, exactEnd, frameAt)
	greatest, greatestOK := occultationPathCenterPointForFrame(greatestTT, frameAt, location)
	if !greatestOK && hasCenter {
		greatestTT = math.Max(centerStart, math.Min(centerEnd, greatestTT))
		greatest, greatestOK = occultationPathCenterPointForFrame(greatestTT, frameAt, location)
	}
	if !greatestOK {
		// Non-central events have no Earth intersection with the shadow axis.
		// Greatest is the ellipsoid point nearest to that axis, not an outer
		// contact tangent. The tangent fallback can place Greatest outside the
		// total band for grazing finite-disk occultations.
		greatest, greatestOK = occultationPathTrackPointForFrame(greatestTT, frameAt, location)
	}
	if !greatestOK {
		return PlanetOccultationPath{}, false, nil
	}
	if !occultationTimeInSelectionWindow(greatest.Time, selectionStart, selectionEnd) {
		return PlanetOccultationPath{}, false, nil
	}
	_, _, greatestWidth, greatestWidthOK := occultationPathLimitsAndWidthForFrame(greatestTT, frameAt)
	if !greatestWidthOK || greatestWidth <= 0 {
		return PlanetOccultationPath{}, false, nil
	}
	// 仅有边界的事件没有影轴与椭球交点；原回退点使用纬度极值弦宽，全掩带使用下方的地面横向宽度。统一两种接触带宽度定义，使有限盘面内外接触宽度可比较。
	// Boundary-only events do not have an axis/ellipsoid intersection. Their fallback point used to carry a latitude-extrema chord width, while total bands used the ground cross-track width below. Keep both contact bands on the same width definition so finite-disk inner/outer widths are comparable.
	greatest.WidthKM = greatestWidth
	totalStartTT, totalEndTT, hasTotal := occultationPathWindowWithCandidateFrames(
		seedTT, searchStart, searchEnd, candidateTotalFrameAt, totalFrameAt, false,
		occultationPathFrameHasBoundary,
	)
	hasTotal = hasTotal && greatestTT >= totalStartTT && greatestTT <= totalEndTT
	if planetOccultationPathTemporalSampleCount(
		outerStart, outerEnd, centerStart, centerEnd, hasCenter,
		totalStartTT, totalEndTT, hasTotal, greatestTT, options,
	) > planetOccultationPathMaxTemporalSamples {
		return PlanetOccultationPath{}, false, ErrOccultationPathSamplingLimit
	}
	if occultationPathEstimatedPointCount(
		outerStart, outerEnd, centerStart, centerEnd, hasCenter,
		totalStartTT, totalEndTT, hasTotal, greatestTT, options,
	) > occultationPathMaxOutputPointCount {
		return PlanetOccultationPath{}, false, ErrOccultationPathSamplingLimit
	}

	start := occultationPathBoundaryEndpointForFrame(outerStart, frameAt, location, 1, inst)
	end := occultationPathBoundaryEndpointForFrame(outerEnd, frameAt, location, -1, inst)
	if !start.valid || !end.valid {
		return PlanetOccultationPath{}, false, nil
	}
	exactFrameAt, exactTotalFrameAt := frameAt, totalFrameAt
	if options.Algorithm != OccultationPathAlgorithmExact {
		optimized := newPlanetOccultationEventCache(config, inst)
		optimized.preparePathEphemeris(seedTT, options.Algorithm)
		if optimized.local.dense {
			cache = optimized
			frameAt, totalFrameAt = cache.outerFrameAt, cache.totalFrameAt
		}
	}
	centerLine, northern, southern, err := planetOccultationPathSamples(
		outerStart, outerEnd, centerStart, centerEnd, hasCenter, greatestTT, frameAt, exactFrameAt, options, location,
	)
	if err != nil {
		return PlanetOccultationPath{}, false, err
	}
	if cache.local.dense {
		correctOccultationCenterWidths(centerLine, exactFrameAt)
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
	// The static visible fill is evaluated from these limit tracks. Keep it on
	// the same station-centred contact equation as the contact contours and
	// rise/set curves; otherwise one event mixes geocentric limits with
	// topocentric contours and the selected envelope can miss a branch.
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
		path.PartialFootprints = planetOccultationFootprints(
			outerStart, outerEnd, greatestTT, frameAt, options, location,
		)
	}
	path.RiseSetCurves = occultationRiseSetCurvesWithCache(
		outerStart, outerEnd, greatestTT, options, location, cache.riseSetCache,
	)
	path.GreatestTimeContours = occultationGreatestTimeContours(
		occultationGreatestTimeLevels(options, outerStart, outerEnd), outerStart, outerEnd, cache.riseSetCache,
		[][]OccultationPathPoint{path.CenterLine, path.NorthernLimit, path.SouthernLimit},
		true, location,
	)
	partialContourTimes := occultationRiseSetEndpointTimes(path.RiseSetCurves)
	if options.DisableFootprints {
		path.PartialBandFootprints, partialContourTimes = planetOccultationBandFootprintsWithContourTimes(
			outerStart, outerEnd, greatestTT, frameAt, location, partialContourTimes,
		)
	}
	if len(path.PartialFootprints) > 0 {
		path.PartialFootprints = occultationStationCorrectFootprintEdges(
			path.PartialFootprints, frameAt, cache.riseSetContextAt, false, location,
		)
	}
	if len(path.PartialBandFootprints) > 0 {
		path.PartialBandFootprints = occultationStationCorrectFootprintEdges(
			path.PartialBandFootprints, frameAt, cache.riseSetContextAt, false, location,
		)
	}
	path.PartialBandContours = occultationContactBandContoursWithAdditionalTimes(
		start.point, end.point, outerStart, outerEnd, greatestTT, frameAt, options, location, partialContourTimes,
	)
	// The geocentric cone supplies the contour topology and time samples; the
	// station oracle corrects each fixed-time contact onto the same topocentric
	// equation used by rise/set curves.  Keep the geocentric seed when a fixed
	// time has no station root yet (normally the short endpoint sliver).
	path.PartialBandContours = occultationStationCorrectContours(
		path.PartialBandContours, path.RiseSetCurves, cache.riseSetCache, location,
	)
	path.PartialVisibilityContours = occultationStationVisibilityEnvelopeContours(
		path.RiseSetCurves, cache.riseSetCache, location,
	)
	if options.DisableFootprints && options.IncludeFootprintTimeline {
		path.PartialFootprints = planetOccultationTimelineFootprints(
			outerStart, outerEnd, greatestTT, frameAt, options, location,
		)
	}

	if !hasTotal {
		return path, true, nil
	}
	totalStart := occultationPathBoundaryEndpointForFrame(totalStartTT, exactTotalFrameAt, location, 1, inst)
	totalEnd := occultationPathBoundaryEndpointForFrame(totalEndTT, exactTotalFrameAt, location, -1, inst)
	_, _, totalWidth, totalWidthOK := occultationPathLimitsAndWidthForFrame(greatestTT, exactTotalFrameAt)
	if !totalStart.valid || !totalEnd.valid || !totalWidthOK || totalWidth <= 0 {
		return path, true, nil
	}
	// The global extrema used for a boundary-only footprint can select
	// different tangent branches for outer and inner cones. At greatest, the
	// physical width is the paired cross-track separation on the same moving
	// track; use it for the finite target before applying the public invariant.
	if crossTrackWidth, ok := occultationPathCrossTrackWidthForFrame(greatestTT, exactTotalFrameAt); ok {
		totalWidth = crossTrackWidth
	}
	if totalWidth >= greatestWidth {
		if crossTrackWidth, ok := occultationPathCrossTrackWidthForFrame(greatestTT, exactFrameAt); ok && crossTrackWidth > 0 {
			greatestWidth = crossTrackWidth
			path.Greatest.WidthKM = crossTrackWidth
		}
	}
	if totalWidth >= greatestWidth {
		totalWidth = math.Nextafter(greatestWidth, 0)
	}
	contourStepDays := occultationPathContourStepDays(options)
	totalNorthern, totalSouthern := occultationPathBoundarySamplesForFrame(
		totalStartTT, totalEndTT, greatestTT, totalFrameAt, contourStepDays, location,
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
	path.NorthernTotalLimit = occultationStationCorrectLimitSeries(
		path.NorthernTotalLimit, totalFrameAt, cache.riseSetContextAt, true, location,
	)
	path.SouthernTotalLimit = occultationStationCorrectLimitSeries(
		path.SouthernTotalLimit, totalFrameAt, cache.riseSetContextAt, true, location,
	)
	_, _ = occultationPathLimitSeparations(path.NorthernTotalLimit, path.SouthernTotalLimit, greatestTT)
	path.TotalRiseSetCurves = occultationRiseSetCurvesWithCache(
		totalStartTT, totalEndTT, greatestTT, options, location, cache.totalRiseSetCache,
	)
	totalContourTimes := occultationRiseSetEndpointTimes(path.TotalRiseSetCurves)
	if !options.DisableFootprints {
		path.TotalFootprints = planetOccultationFootprints(
			totalStartTT, totalEndTT, greatestTT, totalFrameAt, options, location,
		)
	}
	if options.DisableFootprints {
		path.TotalBandFootprints, totalContourTimes = planetOccultationBandFootprintsWithContourTimes(
			totalStartTT, totalEndTT, greatestTT, totalFrameAt, location, totalContourTimes,
		)
	}
	if len(path.TotalFootprints) > 0 {
		path.TotalFootprints = occultationStationCorrectFootprintEdges(
			path.TotalFootprints, totalFrameAt, cache.riseSetContextAt, true, location,
		)
	}
	if len(path.TotalBandFootprints) > 0 {
		path.TotalBandFootprints = occultationStationCorrectFootprintEdges(
			path.TotalBandFootprints, totalFrameAt, cache.riseSetContextAt, true, location,
		)
	}
	path.TotalBandContours = occultationContactBandContoursWithAdditionalTimes(
		totalStart.point, totalEnd.point, totalStartTT, totalEndTT, greatestTT,
		totalFrameAt, options, location, totalContourTimes,
	)
	path.TotalBandContours = occultationStationCorrectContours(
		path.TotalBandContours, path.TotalRiseSetCurves, cache.totalRiseSetCache, location,
	)
	path.TotalVisibilityContours = occultationStationVisibilityEnvelopeContours(
		path.TotalRiseSetCurves, cache.totalRiseSetCache, location,
	)
	if options.DisableFootprints && options.IncludeFootprintTimeline {
		path.TotalFootprints = planetOccultationTimelineFootprints(
			totalStartTT, totalEndTT, greatestTT, totalFrameAt, options, location,
		)
	}
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
	if !options.DisableFootprints || options.IncludeFootprintTimeline {
		count += len(occultationPathSampleTimesWithLimit(
			outerStartTT, outerEndTT, greatestTT, planetOccultationFootprintSampleStepDays(options), planetOccultationFootprintMaxSamples,
		))
	} else {
		count += len(planetOccultationBandSampleTimes(
			outerStartTT, outerEndTT, greatestTT, planetOccultationBandSampleStepDays(), planetOccultationBandMaxSamples,
		))
	}
	if options.DisableFootprints && options.IncludeFootprintTimeline {
		count += len(planetOccultationBandSampleTimes(
			outerStartTT, outerEndTT, greatestTT, planetOccultationBandSampleStepDays(), planetOccultationBandMaxSamples,
		))
	}
	if hasCenter {
		count += len(occultationPathSampleTimes(centerStartTT, centerEndTT, greatestTT, stepDays))
	}
	contourStepDays := occultationPathContourStepDays(options)
	count += 2 * len(occultationPathSampleTimes(outerStartTT, outerEndTT, greatestTT, contourStepDays))
	if hasTotal {
		count += len(occultationPathSampleTimes(totalStartTT, totalEndTT, greatestTT, stepDays))
		count += 2 * len(occultationPathSampleTimes(totalStartTT, totalEndTT, greatestTT, contourStepDays))
		if !options.DisableFootprints || options.IncludeFootprintTimeline {
			count += len(occultationPathSampleTimesWithLimit(
				totalStartTT, totalEndTT, greatestTT, planetOccultationFootprintSampleStepDays(options), planetOccultationFootprintMaxSamples,
			))
		} else {
			count += len(planetOccultationBandSampleTimes(
				totalStartTT, totalEndTT, greatestTT, planetOccultationBandSampleStepDays(), planetOccultationBandMaxSamples,
			))
		}
		if options.DisableFootprints && options.IncludeFootprintTimeline {
			count += len(planetOccultationBandSampleTimes(
				totalStartTT, totalEndTT, greatestTT, planetOccultationBandSampleStepDays(), planetOccultationBandMaxSamples,
			))
		}
	}
	return count
}

func occultationPathWindowForFrame(
	seedTT, startTT, endTT float64,
	frameAt occultationPathFrameFunc,
	center bool,
) (float64, float64, bool) {
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
	engine := occultationMovingDiskEngine()
	return engine.window(seedTT, startTT, endTT, predicate, predicate)
}

func occultationPathGreatestForFrame(seedTT, startTT, endTT float64, frameAt occultationPathFrameFunc) float64 {
	return occultationPathGreatestForFrameIterations(seedTT, startTT, endTT, frameAt, 56)
}

func occultationPathGreatestForFrameIterations(
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

func planetOccultationPathSamples(
	outerStartTT, outerEndTT float64,
	centerStartTT, centerEndTT float64,
	hasCenter bool,
	greatestTT float64,
	frameAt, exactFrameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
) ([]OccultationPathPoint, []OccultationPathPoint, []OccultationPathPoint, error) {
	var centerLine []OccultationPathPoint
	if hasCenter {
		var err error
		centerLine, err = occultationPathCenterSamplesForFrame(centerStartTT, centerEndTT, greatestTT, frameAt, exactFrameAt, options, location)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	contourStepDays := occultationPathContourStepDays(options)
	northern, southern := occultationPathBoundarySamplesForFrame(
		outerStartTT, outerEndTT, greatestTT, frameAt, contourStepDays, location,
	)
	return centerLine, northern, southern, nil
}

func occultationPathBoundarySamplesForFrame(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	stepDays float64,
	location *time.Location,
) ([]OccultationPathPoint, []OccultationPathPoint) {
	times := occultationPathSampleTimes(startTT, endTT, greatestTT, stepDays)
	return occultationPathBoundarySamplesAtTimesForFrame(
		times, greatestTT, frameAt, location, occultationPathContourSpacingKM, false,
	)
}

func occultationPathBoundaryContourSamplesForFrame(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	stepDays float64,
	location *time.Location,
	additionalTimes []float64,
) ([]OccultationPathPoint, []OccultationPathPoint) {
	times := occultationPathSampleTimes(startTT, endTT, greatestTT, stepDays)
	for _, tt := range additionalTimes {
		if !finite(tt) || tt < startTT || tt > endTT {
			continue
		}
		times = append(times, tt)
	}
	sort.Float64s(times)
	times = uniqueOccultationPathTimes(times)
	return occultationPathBoundarySamplesAtTimesForFrame(
		times, greatestTT, frameAt, location, occultationPathContourSpacingKM, true,
	)
}

func occultationPathBoundarySamplesAtTimesForFrame(
	times []float64,
	greatestTT float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	targetSpacingKM float64,
	useFiniteArcExtrema bool,
) ([]OccultationPathPoint, []OccultationPathPoint) {
	limitsAt := occultationPathCrossTrackLimitsForFrame
	if useFiniteArcExtrema {
		limitsAt = occultationPathContourCrossTrackLimitsForFrame
	}
	samples := make([]occultationPathBoundaryPairSample, 0, len(times))
	for _, tt := range times {
		firstVector, secondVector, moon, ok := occultationPathContourBoundaryPairAt(
			tt, frameAt, limitsAt, useFiniteArcExtrema, nil,
		)
		if !ok {
			continue
		}
		sample := occultationPathBoundaryPairSample{tt: tt, first: firstVector, second: secondVector, moon: moon}
		if len(samples) == 0 {
			samples = append(samples, sample)
			continue
		}
		samples = appendOccultationPathBoundaryPairSegment(
			samples, samples[len(samples)-1], sample, frameAt, targetSpacingKM, limitsAt, useFiniteArcExtrema, 0,
		)
	}
	first := make([]OccultationPathPoint, len(samples))
	second := make([]OccultationPathPoint, len(samples))
	for index, sample := range samples {
		siderealDegrees := ApparentSiderealTime(TT2UT1(sample.tt)) * 15
		first[index] = occultationPathPointFromVectorWithMoonSidereal(
			sample.tt, sample.first, 0, sample.moon, siderealDegrees, location,
		)
		second[index] = occultationPathPointFromVectorWithMoonSidereal(
			sample.tt, sample.second, 0, sample.moon, siderealDegrees, location,
		)
	}
	first, second = occultationPathDeduplicateBoundaryPoints(first, second)
	return occultationPathOrientBoundarySamples(first, second, greatestTT)
}

type occultationPathBoundaryPairSample struct {
	tt            float64
	first, second occultationPathVector
	moon          occultationPathVector
}

func appendOccultationPathBoundaryPairSegment(
	samples []occultationPathBoundaryPairSample,
	start, end occultationPathBoundaryPairSample,
	frameAt occultationPathFrameFunc,
	targetSpacingKM float64,
	limitsAt func(float64, occultationPathFrameFunc) (occultationPathVector, occultationPathVector, occultationPathVector, bool),
	useFiniteArcExtrema bool,
	depth int,
) []occultationPathBoundaryPairSample {
	if useFiniteArcExtrema {
		end = occultationPathRepairContourBoundaryPair(start, end, frameAt, limitsAt)
	}
	end = occultationPathOrientBoundaryPair(start, end)
	spacing := occultationPathBoundaryPairDirectSpacing(start, end)
	if useFiniteArcExtrema {
		// The public maps use Web Mercator. Near a pole a physically short
		// ground-track step can span a much larger projected x/y distance, so
		// a ground-distance-only sampler renders visible corners and notches.
		// Refine only finite-disk contact contours using the same projection
		// metric that the map consumes; ordinary event/path sampling remains
		// on the cheaper physical-distance criterion.
		projectedSpacing := occultationPathBoundaryPairWebMercatorSpacing(start, end)
		projectedTarget := occultationPathBoundaryPairWebMercatorTarget(start, end, targetSpacingKM)
		if projectedSpacing > projectedTarget {
			spacing = math.Max(spacing, projectedSpacing*targetSpacingKM/projectedTarget)
		}
	}
	if depth >= occultationPathMaxAdaptiveDepth || spacing <= targetSpacingKM {
		return append(samples, end)
	}
	midTT := (start.tt + end.tt) / 2
	if midTT <= start.tt || midTT >= end.tt {
		return append(samples, end)
	}
	first, second, moon, ok := occultationPathContourBoundaryPairAt(
		midTT, frameAt, limitsAt, useFiniteArcExtrema, &start,
	)
	if !ok {
		return append(samples, end)
	}
	mid := occultationPathBoundaryPairSample{tt: midTT, first: first, second: second, moon: moon}
	samples = appendOccultationPathBoundaryPairSegment(samples, start, mid, frameAt, targetSpacingKM, limitsAt, useFiniteArcExtrema, depth+1)
	return appendOccultationPathBoundaryPairSegment(
		samples, samples[len(samples)-1], end, frameAt, targetSpacingKM, limitsAt, useFiniteArcExtrema, depth+1,
	)
}

func occultationPathBoundaryPairWebMercatorSpacing(
	first, second occultationPathBoundaryPairSample,
) float64 {
	return math.Max(
		occultationPathWebMercatorPointSpacing(first.tt, first.first, second.tt, second.first),
		occultationPathWebMercatorPointSpacing(first.tt, first.second, second.tt, second.second),
	)
}

func occultationPathBoundaryPairWebMercatorTarget(
	first, second occultationPathBoundaryPairSample,
	targetSpacingKM float64,
) float64 {
	if targetSpacingKM <= 0 {
		return targetSpacingKM
	}
	_, firstStartLatitude := occultationPathGeodetic(first.tt, first.first)
	_, firstEndLatitude := occultationPathGeodetic(second.tt, second.first)
	_, secondStartLatitude := occultationPathGeodetic(first.tt, first.second)
	_, secondEndLatitude := occultationPathGeodetic(second.tt, second.second)
	latitude := math.Max(
		math.Max(math.Abs(firstStartLatitude), math.Abs(firstEndLatitude)),
		math.Max(math.Abs(secondStartLatitude), math.Abs(secondEndLatitude)),
	) * math.Pi / 180
	// Longitude is stretched by sec(latitude) in Web Mercator. Keep the
	// projected contour step near the physical 40 km target at low latitudes,
	// but cap it at roughly 10 km around the polar part of this event.
	return targetSpacingKM * math.Max(0.75, math.Cos(latitude))
}

func occultationPathWebMercatorPointSpacing(
	firstTT float64,
	firstVector occultationPathVector,
	secondTT float64,
	secondVector occultationPathVector,
) float64 {
	firstLongitude, firstLatitude := occultationPathGeodetic(firstTT, firstVector)
	secondLongitude, secondLatitude := occultationPathGeodetic(secondTT, secondVector)
	return geodata.WebMercatorDistanceKM(
		geodata.GeoPoint{Longitude: firstLongitude, Latitude: firstLatitude},
		geodata.GeoPoint{Longitude: secondLongitude, Latitude: secondLatitude},
	)
}

func occultationPathContourBoundaryPairAt(
	tt float64,
	frameAt occultationPathFrameFunc,
	limitsAt func(float64, occultationPathFrameFunc) (occultationPathVector, occultationPathVector, occultationPathVector, bool),
	useFiniteArcExtrema bool,
	previous *occultationPathBoundaryPairSample,
) (occultationPathVector, occultationPathVector, occultationPathVector, bool) {
	first, second, moon, ok := limitsAt(tt, frameAt)
	if !ok {
		if useFiniteArcExtrema && previous != nil {
			return previous.first, previous.second, previous.moon, true
		}
		return occultationPathVector{}, occultationPathVector{}, occultationPathVector{}, false
	}
	candidate := occultationPathBoundaryPairSample{tt: tt, first: first, second: second, moon: moon}
	if !useFiniteArcExtrema || previous == nil {
		return candidate.first, candidate.second, candidate.moon, true
	}
	candidate = occultationPathOrientBoundaryPair(*previous, candidate)
	if occultationPathBoundaryPairDirectSpacing(*previous, candidate) <= occultationPathBoundarySpacingKM {
		return candidate.first, candidate.second, candidate.moon, true
	}
	if frame, frameOK := frameAt(tt); frameOK && frame.targetRadius != 0 {
		if first, second, ok := occultationPathContinueFiniteBoundaryPair(
			frame, tt, *previous,
		); ok {
			return first, second, frame.moon, true
		}
	}
	return candidate.first, candidate.second, candidate.moon, true
}

func occultationPathContinueFiniteBoundaryPair(
	frame occultationPathFrame,
	tt float64,
	previous occultationPathBoundaryPairSample,
) (occultationPathVector, occultationPathVector, bool) {
	intervals := occultationPathBoundaryThetaIntervals(frame)
	if len(intervals) == 0 {
		return occultationPathVector{}, occultationPathVector{}, false
	}
	currentRotation := occultationPathEarthRotationAt(tt)
	previousRotation := occultationPathEarthRotationAt(previous.tt)
	nearest := func(target occultationPathVector) (occultationPathVector, bool) {
		targetFixed := occultationPathEarthFixedVectorWithRotation(target, previousRotation)
		bestDistance := math.Inf(1)
		var best occultationPathVector
		for _, interval := range intervals {
			const samples = 48
			bestIndex := -1
			for index := 0; index <= samples; index++ {
				theta := interval.left + (interval.right-interval.left)*float64(index)/samples
				point, _, ok := occultationPathBoundaryVector(frame, theta)
				if !ok {
					continue
				}
				fixed := occultationPathEarthFixedVectorWithRotation(point, currentRotation)
				distance := occultationPathNorm(occultationPathSub(fixed, targetFixed))
				if distance < bestDistance {
					bestDistance, best, bestIndex = distance, point, index
				}
			}
			if bestIndex < 0 {
				continue
			}
			leftIndex := math.Max(float64(bestIndex-1), 0)
			rightIndex := math.Min(float64(bestIndex+1), samples)
			left := interval.left + (interval.right-interval.left)*leftIndex/samples
			right := interval.left + (interval.right-interval.left)*rightIndex/samples
			const goldenRatio = 0.6180339887498949
			x1 := right - goldenRatio*(right-left)
			x2 := left + goldenRatio*(right-left)
			distanceAt := func(theta float64) (float64, occultationPathVector, bool) {
				point, _, ok := occultationPathBoundaryVector(frame, theta)
				if !ok {
					return math.Inf(1), occultationPathVector{}, false
				}
				fixed := occultationPathEarthFixedVectorWithRotation(point, currentRotation)
				return occultationPathNorm(occultationPathSub(fixed, targetFixed)), point, true
			}
			f1, _, _ := distanceAt(x1)
			f2, _, _ := distanceAt(x2)
			for iteration := 0; iteration < 24; iteration++ {
				if f1 > f2 {
					left = x1
					x1, f1 = x2, f2
					x2 = left + goldenRatio*(right-left)
					f2, _, _ = distanceAt(x2)
				} else {
					right = x2
					x2, f2 = x1, f1
					x1 = right - goldenRatio*(right-left)
					f1, _, _ = distanceAt(x1)
				}
			}
			distance, point, ok := distanceAt((left + right) / 2)
			if ok && distance < bestDistance {
				bestDistance, best = distance, point
			}
		}
		return best, finite(bestDistance)
	}
	first, firstOK := nearest(previous.first)
	second, secondOK := nearest(previous.second)
	if !firstOK || !secondOK || occultationPathNorm(occultationPathSub(first, second)) < 1 {
		return occultationPathVector{}, occultationPathVector{}, false
	}
	return first, second, true
}

func occultationPathRepairContourBoundaryPair(
	previous, current occultationPathBoundaryPairSample,
	frameAt occultationPathFrameFunc,
	limitsAt func(float64, occultationPathFrameFunc) (occultationPathVector, occultationPathVector, occultationPathVector, bool),
) occultationPathBoundaryPairSample {
	current = occultationPathOrientBoundaryPair(previous, current)
	if occultationPathBoundaryPairDirectSpacing(previous, current) <= occultationPathBoundarySpacingKM {
		return current
	}
	if frame, ok := frameAt(current.tt); ok && frame.targetRadius != 0 {
		if first, second, ok := occultationPathContinueFiniteBoundaryPair(
			frame, current.tt, previous,
		); ok {
			continued := occultationPathOrientBoundaryPair(previous, occultationPathBoundaryPairSample{
				tt: current.tt, first: first, second: second, moon: frame.moon,
			})
			if occultationPathBoundaryPairDirectSpacing(previous, continued) <=
				occultationPathBoundaryPairDirectSpacing(previous, current) {
				return continued
			}
		}
		if _, _, _, finiteOK := limitsAt(current.tt, frameAt); !finiteOK {
			return previous
		}
	}
	_ = limitsAt
	return current
}

func occultationPathOrientBoundaryPair(
	previous, current occultationPathBoundaryPairSample,
) occultationPathBoundaryPairSample {
	direct, swapped := occultationPathBoundaryPairSpacings(previous, current)
	if swapped < direct {
		current.first, current.second = current.second, current.first
	}
	return current
}

func occultationPathBoundaryPairDirectSpacing(first, second occultationPathBoundaryPairSample) float64 {
	direct, _ := occultationPathBoundaryPairSpacings(first, second)
	return direct
}

func occultationPathBoundaryPairSpacings(first, second occultationPathBoundaryPairSample) (float64, float64) {
	firstRotation := occultationPathEarthRotationAt(first.tt)
	secondRotation := occultationPathEarthRotationAt(second.tt)
	firstA := occultationPathEarthFixedVectorWithRotation(first.first, firstRotation)
	firstB := occultationPathEarthFixedVectorWithRotation(first.second, firstRotation)
	secondA := occultationPathEarthFixedVectorWithRotation(second.first, secondRotation)
	secondB := occultationPathEarthFixedVectorWithRotation(second.second, secondRotation)
	direct := math.Max(
		occultationPathNorm(occultationPathSub(secondA, firstA)),
		occultationPathNorm(occultationPathSub(secondB, firstB)),
	)
	swapped := math.Max(
		occultationPathNorm(occultationPathSub(secondB, firstA)),
		occultationPathNorm(occultationPathSub(secondA, firstB)),
	)
	return direct, swapped
}

func occultationPathCrossTrackLimitsForFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
) (occultationPathVector, occultationPathVector, occultationPathVector, bool) {
	frame, ok := frameAt(tt)
	before, beforeOK := frameAt(tt - occultationPathVelocityStepDays)
	after, afterOK := frameAt(tt + occultationPathVelocityStepDays)
	if !ok || !beforeOK || !afterOK {
		return occultationPathVector{}, occultationPathVector{}, occultationPathVector{}, false
	}
	vx := after.moonProjectionX() - before.moonProjectionX()
	vy := after.moonProjectionY() - before.moonProjectionY()
	if math.Hypot(vx, vy) <= 1e-12 {
		return occultationPathVector{}, occultationPathVector{}, occultationPathVector{}, false
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
			return occultationPathVector{}, occultationPathVector{}, occultationPathVector{}, false
		}
		return first, second, frame.moon, true
	}
	intervals := occultationPathBoundaryThetaIntervals(frame)
	interval, intervalOK := occultationPathNearestThetaInterval(intervals, theta)
	if firstOK && secondOK {
		if intervalOK && occultationPathAngleDistance(interval.left, theta) > occultationPathAngleDistance(interval.left, theta+math.Pi) {
			first, second = second, first
		}
		return first, second, frame.moon, true
	}
	if !intervalOK {
		return occultationPathVector{}, occultationPathVector{}, occultationPathVector{}, false
	}
	first, _, firstOK = occultationPathBoundaryVector(frame, interval.left)
	second, _, secondOK = occultationPathBoundaryVector(frame, interval.right)
	if !firstOK || !secondOK {
		return occultationPathVector{}, occultationPathVector{}, occultationPathVector{}, false
	}
	return first, second, frame.moon, true
}

// occultationPathContourCrossTrackLimitsForFrame is used only for the static
// contact-contour band. A finite target may expose one long visible
// contact-cone arc; its horizon endpoints are closure points, not necessarily
// the two cross-track sides of the band. Scan that arc for the physical sides,
// while retaining the normal paired-limit fallback for grazing/degenerate
// samples.
func occultationPathContourCrossTrackLimitsForFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
) (occultationPathVector, occultationPathVector, occultationPathVector, bool) {
	frame, ok := frameAt(tt)
	if !ok || frame.targetRadius == 0 {
		return occultationPathCrossTrackLimitsForFrame(tt, frameAt)
	}
	before, beforeOK := frameAt(tt - occultationPathVelocityStepDays)
	after, afterOK := frameAt(tt + occultationPathVelocityStepDays)
	if !beforeOK || !afterOK {
		return occultationPathCrossTrackLimitsForFrame(tt, frameAt)
	}
	first, second, _, extremaOK := occultationPathFiniteCrossTrackExtrema(tt, frame, before, after)
	if extremaOK {
		return first, second, frame.moon, true
	}
	return occultationPathCrossTrackLimitsForFrame(tt, frameAt)
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
	if frame.boundary != nil {
		frame.boundary.boundaryOnce.Do(func() { occultationPathBoundaryPrecompute(frame) })
		return frame.boundary.intervals
	}
	return occultationPathBoundaryThetaIntervalsUncached(frame)
}

// occultationPathBoundaryThetaIntervalsUncached 是可见 θ 区间扫描本体；调用方通过 frame 级缓存复用结果。
// occultationPathBoundaryThetaIntervalsUncached is the visible-theta scan itself; callers reuse it through the frame cache.
func occultationPathBoundaryThetaIntervalsUncached(frame occultationPathFrame) []occultationPathThetaInterval {
	step := 2 * math.Pi / float64(occultationPathBoundaryScanPoints)
	discriminants := make([]float64, occultationPathBoundaryScanPoints)
	occultationPathBoundaryFillGrid(frame, discriminants)
	return occultationPathBoundaryThetaIntervalsFromGrid(frame, step, discriminants)
}

// occultationPathBoundaryFillGrid 计算整圈 720 点判别式网格，失败点记为 -Inf。
// occultationPathBoundaryFillGrid evaluates the 720-point discriminant grid, marking failures as -Inf.
func occultationPathBoundaryFillGrid(frame occultationPathFrame, discriminants []float64) {
	step := 2 * math.Pi / float64(occultationPathBoundaryScanPoints)
	for index := range discriminants {
		discriminant, _, _, ok := occultationPathBoundaryLine(frame, step*float64(index))
		if !ok {
			discriminant = math.Inf(-1)
		}
		discriminants[index] = discriminant
	}
}

// occultationPathBoundaryThetaIntervalsFromGrid 在已算好的网格上恢复可见 θ 区间。
// occultationPathBoundaryThetaIntervalsFromGrid recovers the visible theta intervals from a prepared grid.
func occultationPathBoundaryThetaIntervalsFromGrid(
	frame occultationPathFrame,
	step float64,
	discriminants []float64,
) []occultationPathThetaInterval {
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
	frameAt, exactFrameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
) ([]OccultationPathPoint, error) {
	stepDays := float64(options.Step) / float64(24*time.Hour)
	times := occultationPathSampleTimes(startTT, endTT, greatestTT, stepDays)
	points := make([]OccultationPathPoint, 0, len(times))
	for _, tt := range times {
		point, ok := occultationPathCenterPointForFrame(tt, frameAt, location)
		// 中心窗口端点由精确帧求得，插值帧在端点附近可能判无交点，此时回退精确帧。
		if !ok && exactFrameAt != nil {
			point, ok = occultationPathCenterPointForFrame(tt, exactFrameAt, location)
		}
		if ok && (len(points) == 0 || point.Time.After(points[len(points)-1].Time)) {
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
	midTime := occultationTTToLocation(midTT, location)
	if !midTime.After(start.Time) || !midTime.Before(end.Time) {
		return append(points, end), nil
	}
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

func planetOccultationPathFrameAt(tt float64, config planetOccultationConfig, instances ...*Instance) (occultationPathFrame, bool) {
	inst := instanceValue(instances)

	return planetOccultationContactPathFrameAt(tt, config, false, inst)
}

func planetOccultationTotalPathFrameAt(tt float64, config planetOccultationConfig) (occultationPathFrame, bool) {
	return planetOccultationContactPathFrameAt(tt, config, true)
}

func planetOccultationContactPathFrameAt(tt float64, config planetOccultationConfig, total bool, instances ...*Instance) (occultationPathFrame, bool) {
	inst := instanceValue(instances)

	state := planetOccultationEphemerisStateAt(tt, config, inst)
	return planetOccultationContactPathFrameFromState(state, config, total)
}

func planetOccultationEphemerisStateAt(tt float64, config planetOccultationConfig, instances ...*Instance) planetOccultationEphemerisState {
	inst := instanceValue(instances)

	moonRA, moonDec := HMoonGeocentricApparentRaDecNWithInstance(tt, -1, inst)
	moonDistance := HMoonAwayNWithInstance(tt, -1, inst)
	planetRA, planetDec, planetDistanceAU := planetOccultationApparentPositionAndDistanceN(tt, config, -1, inst)
	planetDistance := planetDistanceAU * occultationPathAstronomicalUnitKM
	return planetOccultationEphemerisState{
		moonRA: moonRA, moonDec: moonDec, moonDistanceKM: moonDistance,
		planetRA: planetRA, planetDec: planetDec, planetDistanceKM: planetDistance,
		valid: finite(moonRA) && finite(moonDec) && finite(moonDistance) && moonDistance > 0 &&
			finite(planetRA) && finite(planetDec) && finite(planetDistance) && planetDistance > 0,
	}
}

func planetOccultationApparentPositionAndDistanceN(
	tt float64,
	config planetOccultationConfig,
	n int, instances ...*Instance) (float64, float64, float64) {
	inst := instanceValue(instances)

	position, _, distanceAU := planetApparentGeocentricPositionAndDistanceN(config.planetIndex, tt, n, inst)
	ra, dec := planetApparentRaDecFromLoBo(tt, position.lo, position.bo)
	return ra, dec, distanceAU
}

func planetOccultationContactPathFrameFromState(
	state planetOccultationEphemerisState,
	config planetOccultationConfig,
	total bool,
) (occultationPathFrame, bool) {
	if !state.valid || state.planetDistanceKM <= state.moonDistanceKM {
		return occultationPathFrame{}, false
	}
	moon := occultationPathRaDecVector(state.moonRA, state.moonDec, state.moonDistanceKM)
	target := occultationPathRaDecVector(state.planetRA, state.planetDec, state.planetDistanceKM)
	moonToTarget := occultationPathSub(target, moon)
	moonToTargetDistance := occultationPathNorm(moonToTarget)
	moonRadius := math.Asin(moonEquatorialRadiusKM / state.moonDistanceKM)
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

func (cache *planetOccultationEventCache) stateAt(tt float64) planetOccultationEphemerisState {
	key := math.Float64bits(tt)
	if state, ok := cache.states[key]; ok {
		return state
	}
	if len(cache.states) >= planetOccultationEventCacheMaximumEntries {
		for cachedKey := range cache.states {
			delete(cache.states, cachedKey)
		}
		for cachedKey := range cache.outerFrames {
			delete(cache.outerFrames, cachedKey)
		}
		for cachedKey := range cache.totalFrames {
			delete(cache.totalFrames, cachedKey)
		}
	}
	var state planetOccultationEphemerisState
	interpolated := false
	if cache.local != nil && cache.local.dense {
		state, interpolated = cache.local.stateAt(tt)
	}
	if !interpolated {
		state = planetOccultationEphemerisStateAt(tt, cache.config, cache.inst)
	}
	cache.states[key] = state
	return state
}

func (cache *planetOccultationEventCache) prepareLocalEphemeris(center float64) {
	if cache.local == nil {
		cache.local = newPlanetOccultationLocalEphemeris(center, cache.config, cache.inst)
	}
}

func (cache *planetOccultationEventCache) candidateFrameAt(tt float64) (occultationPathFrame, bool) {
	return cache.candidateFrameAtKind(tt, false)
}

func (cache *planetOccultationEventCache) candidateTotalFrameAt(tt float64) (occultationPathFrame, bool) {
	return cache.candidateFrameAtKind(tt, true)
}

func (cache *planetOccultationEventCache) candidateFrameAtKind(tt float64, total bool) (occultationPathFrame, bool) {
	if cache.local != nil {
		if state, ok := cache.local.stateAt(tt); ok {
			return planetOccultationContactPathFrameFromState(state, cache.config, total)
		}
	}
	return cache.frameAt(tt, total)
}

func (cache *planetOccultationEventCache) outerFrameAt(tt float64) (occultationPathFrame, bool) {
	return cache.frameAt(tt, false)
}

func (cache *planetOccultationEventCache) totalFrameAt(tt float64) (occultationPathFrame, bool) {
	return cache.frameAt(tt, true)
}

func (cache *planetOccultationEventCache) frameAt(tt float64, total bool) (occultationPathFrame, bool) {
	key := math.Float64bits(tt)
	frames := cache.outerFrames
	if total {
		frames = cache.totalFrames
	}
	if entry, ok := frames[key]; ok {
		return entry.frame, entry.ok
	}
	frame, ok := planetOccultationContactPathFrameFromState(cache.stateAt(tt), cache.config, total)
	if ok {
		frame.boundary = &occultationPathBoundaryCache{}
	}
	frames[key] = planetOccultationFrameCacheEntry{frame: frame, ok: ok}
	return frame, ok
}

func (cache *planetOccultationEventCache) riseSetContextAt(tt float64) occultationRiseSetContext {
	state := cache.stateAt(tt)
	return newOccultationRiseSetContext(
		tt, state.moonRA, state.moonDec, state.moonDistanceKM,
		state.planetRA, state.planetDec, state.planetDistanceKM, cache.config.equatorialRadiusKM,
	)
}

func (cache *planetOccultationEventCache) candidateRiseSetContextAt(tt float64) occultationRiseSetContext {
	var context occultationRiseSetContext
	if cache.local != nil {
		moonXYZ, targetXYZ, ok := cache.local.vectorsAt(tt)
		if ok {
			context = newOccultationRiseSetContextFromVectors(
				tt,
				moonXYZ,
				targetXYZ,
				true,
				cache.config.equatorialRadiusKM,
			)
		} else {
			context = cache.riseSetContextAt(tt)
		}
	} else {
		context = cache.riseSetContextAt(tt)
	}
	return context
}

func (cache *planetOccultationEventCache) totalRiseSetContextAt(tt float64) occultationRiseSetContext {
	return cache.riseSetContextAt(tt).withInternalContact()
}

func (cache *planetOccultationEventCache) candidateTotalRiseSetContextAt(tt float64) occultationRiseSetContext {
	return cache.candidateRiseSetContextAt(tt).withInternalContact()
}

func occultationPathFrameHasBoundary(frame occultationPathFrame) bool {
	_, _, ok := occultationPathBoundaryTangent(frame)
	return ok
}

func occultationPathBoundaryEndpointForFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
	direction int, instances ...*Instance) occultationPathEndpoint {
	inst := instanceValue(instances)

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
			return occultationPathEndpoint{point: occultationPathPointFromVector(candidateTT, vector, 0, location, inst), valid: true}
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
	location *time.Location, instances ...*Instance) (OccultationPathPoint, bool) {
	inst := instanceValue(instances)

	frame, ok := frameAt(tt)
	if !ok {
		return OccultationPathPoint{}, false
	}
	point, _, ok := occultationPathBoundaryTangent(frame)
	if !ok {
		return OccultationPathPoint{}, false
	}
	north, south, limitsOK := occultationPathScannedLimitsAtFrame(tt, frame)
	return occultationPathPointFromVector(tt, point, occultationPathBoundaryWidth(north, south, limitsOK), location, inst), true
}

func occultationPathTrackPointForFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
) (OccultationPathPoint, bool) {
	frame, ok := frameAt(tt)
	if !ok {
		return OccultationPathPoint{}, false
	}
	point, ok := occultationPathTrackReference(frame)
	if !ok {
		return OccultationPathPoint{}, false
	}
	width := 0.0
	if _, _, candidate, limitsOK := occultationPathLimitsAndWidthForFrame(tt, frameAt); limitsOK {
		width = candidate
	}
	return occultationPathPointFromVectorWithMoon(tt, point, width, frame.moon, location), true
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
	centerRotation := occultationPathEarthRotation{}
	if centerOK && beforeCenterOK && afterCenterOK {
		centerRotation = occultationPathEarthRotationAt(tt)
		centerFixed = occultationPathEarthFixedVectorWithRotation(center, centerRotation)
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
			fixed := occultationPathEarthFixedVectorWithRotation(point, centerRotation)
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
			// 末样本的等差写法与 rightTheta 相差 1 ULP：掠射时该 ULP 会把端点的判别式推成负值，
			// 端点处的横向极值就被当成"无地面交点"丢掉，故端点原值必须补采一次。
			if point, _, pointOK := occultationPathBoundaryVector(frame, rightTheta); pointOK {
				consider(point)
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
	minimumLatitude := occultationPathGeodeticLatitude(minimumPoint)
	maximumLatitude := occultationPathGeodeticLatitude(maximumPoint)
	if maximumLatitude >= minimumLatitude {
		return maximumPoint, minimumPoint, width, true
	}
	return minimumPoint, maximumPoint, width, true
}

// occultationPathFiniteCrossTrackExtrema scans the complete Earth-intersecting
// contact-cone arc and returns its physical cross-track extrema. For a finite
// target the visible arc can be much longer than the two horizon endpoints;
// those endpoints are only the moonrise/moonset closure, not the band sides.
func occultationPathFiniteCrossTrackExtrema(
	tt float64,
	frame, beforeFrame, afterFrame occultationPathFrame,
) (occultationPathVector, occultationPathVector, float64, bool) {
	vx := afterFrame.moonProjectionX() - beforeFrame.moonProjectionX()
	vy := afterFrame.moonProjectionY() - beforeFrame.moonProjectionY()
	speed := math.Hypot(vx, vy)
	if speed <= 1e-12 {
		return occultationPathVector{}, occultationPathVector{}, 0, false
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
	centerRotation := occultationPathEarthRotation{}
	if centerOK && beforeCenterOK && afterCenterOK {
		centerRotation = occultationPathEarthRotationAt(tt)
		centerFixed = occultationPathEarthFixedVectorWithRotation(center, centerRotation)
		beforeFixed := occultationPathEarthFixedVector(
			tt-occultationPathVelocityStepDays, before,
		)
		afterFixed := occultationPathEarthFixedVector(
			tt+occultationPathVelocityStepDays, after,
		)
		polarRatioSquared := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
		normal := occultationPathUnit(occultationPathVector{
			x: centerFixed.x, y: centerFixed.y, z: centerFixed.z / polarRatioSquared,
		})
		track := occultationPathSub(afterFixed, beforeFixed)
		track = occultationPathSub(track, occultationPathScale(normal, occultationPathDot(track, normal)))
		if occultationPathNorm(track) > 1e-12 {
			groundCrossTrack = occultationPathUnit(occultationPathCross(normal, occultationPathUnit(track)))
			groundWidthOK = true
		}
	}

	minimumOffset, maximumOffset := math.Inf(1), math.Inf(-1)
	minimumGroundOffset, maximumGroundOffset := math.Inf(1), math.Inf(-1)
	var minimumPoint, maximumPoint occultationPathVector
	var minimumGroundPoint, maximumGroundPoint occultationPathVector
	consider := func(point occultationPathVector) {
		offset := occultationPathDot(point, planeCrossTrack)
		if offset < minimumOffset {
			minimumOffset, minimumPoint = offset, point
		}
		if offset > maximumOffset {
			maximumOffset, maximumPoint = offset, point
		}
		if groundWidthOK {
			fixed := occultationPathEarthFixedVectorWithRotation(point, centerRotation)
			groundOffset := occultationPathDot(
				occultationPathSub(fixed, centerFixed), groundCrossTrack,
			)
			if groundOffset < minimumGroundOffset {
				minimumGroundOffset, minimumGroundPoint = groundOffset, point
			}
			if groundOffset > maximumGroundOffset {
				maximumGroundOffset, maximumGroundPoint = groundOffset, point
			}
		}
	}
	if tangentPoint, tangentTheta, tangentOK := occultationPathBoundaryTangent(frame); tangentOK {
		consider(tangentPoint)
		if leftTheta, rightTheta, intervalOK := occultationPathBoundaryThetaInterval(frame, tangentTheta); intervalOK {
			for index := 0; index <= 128; index++ {
				theta := leftTheta + (rightTheta-leftTheta)*float64(index)/128
				if point, _, pointOK := occultationPathBoundaryVector(frame, theta); pointOK {
					consider(point)
				}
			}
			// 末样本的等差写法与 rightTheta 相差 1 ULP：掠射时该 ULP 会把端点的判别式推成负值，
			// 端点处的极值就被当成"无地面交点"丢掉，故端点原值必须补采一次。
			if point, _, pointOK := occultationPathBoundaryVector(frame, rightTheta); pointOK {
				consider(point)
			}
		}
	}
	for _, interval := range occultationPathBoundaryThetaIntervals(frame) {
		for index := 0; index <= 128; index++ {
			theta := interval.left + (interval.right-interval.left)*float64(index)/128
			if point, _, pointOK := occultationPathBoundaryVector(frame, theta); pointOK {
				consider(point)
			}
		}
		// 与 occultationPathLimitsAndWidthForFrame 同因：等差末样本与 interval.right 差 1 ULP，
		// 掠射时会把端点判别式推成负值而丢掉该处极值，端点原值必须补采一次。
		if point, _, pointOK := occultationPathBoundaryVector(frame, interval.right); pointOK {
			consider(point)
		}
	}
	if !finite(minimumOffset) || !finite(maximumOffset) {
		return occultationPathVector{}, occultationPathVector{}, 0, false
	}
	if groundWidthOK && finite(minimumGroundOffset) && finite(maximumGroundOffset) {
		width := maximumGroundOffset - minimumGroundOffset
		if width > 0 {
			return maximumGroundPoint, minimumGroundPoint, width, true
		}
	}
	minimumLatitude := occultationPathGeodeticLatitude(minimumPoint)
	maximumLatitude := occultationPathGeodeticLatitude(maximumPoint)
	width := maximumOffset - minimumOffset
	if maximumLatitude >= minimumLatitude {
		return maximumPoint, minimumPoint, width, width > 0
	}
	return minimumPoint, maximumPoint, width, width > 0
}

func occultationPathScannedLimitsForFrame(
	tt float64,
	frame occultationPathFrame,
) (occultationPathVector, occultationPathVector, bool) {
	return occultationPathScannedLimitsAtFrame(tt, frame)
}

func occultationPathCrossTrackWidthForFrame(
	tt float64,
	frameAt occultationPathFrameFunc,
) (float64, bool) {
	first, second, _, ok := occultationPathCrossTrackLimitsForFrame(tt, frameAt)
	if !ok {
		return 0, false
	}
	width := occultationPathNorm(occultationPathSub(first, second))
	return width, finite(width) && width > 0
}

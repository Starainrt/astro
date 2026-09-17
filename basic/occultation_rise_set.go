package basic

import (
	"math"
	"time"
)

const (
	occultationRiseSetBoundaryPoints     = 180
	occultationRiseSetDerivativeStepDays = 5.0 / 86400.0
	// A polar phase fold can create or remove a root branch inside one coarse
	// rise/set interval. Refine only those topology-changing intervals rather
	// than lowering the global sampling step for every event.
	occultationRiseSetAdaptiveMaximumPasses       = 1
	occultationRiseSetAdaptiveMinimumStepDays     = 5.0 / 86400.0
	occultationRiseSetAdaptiveMaximumIntervals    = 64
	occultationRiseSetFoldRootToleranceDays       = 1e-10
	occultationRiseSetJunctionDerivativeTolerance = 1e-6
	// Beyond this point, recomputing the small vector state costs less than growing every per-TT map.
	occultationRiseSetStateCacheMaximumEntries = 8
	// 48 h 窗口按 1 分钟采样需要 2880 个时刻、每个时刻三个上下文（中心/前/后），
	// 4096 个槽位会在事件中途反复整表清空并重算精确星历。
	// A 48 h window sampled every minute needs 2880 instants with three contexts each
	// (center/before/after); 4096 slots cleared the whole table repeatedly mid-event and
	// recomputed the exact ephemerides it had just dropped.
	occultationRiseSetEvaluationCacheMaximumEntries = 16384
	occultationRiseSetTimeEpsilonDays               = 1e-8
	occultationRiseSetTargetSpacingKM               = 100.0
	occultationTopocentricEarthRadiusKM             = 6378.14
	occultationTopocentricEarthPolarRadiusKM        = 6356.755
)

// 站心赤经赤纬的历史公式使用固定地平视差常数；保留该尺度可使向量算法与既有结果数值等价。
// The legacy topocentric RA/Dec formula uses a fixed horizontal-parallax constant; retaining that scale keeps the vector implementation numerically equivalent.
var occultationLegacyParallaxRadiusKM = math.Sin(0.0024427777777*rad) * occultationPathAstronomicalUnitKM

type occultationRiseSetBody struct {
	direction  occultationPathVector
	positionKM occultationPathVector
	distanceKM float64
}

type occultationRiseSetContext struct {
	tt              float64
	siderealDegrees float64
	moonRA          float64
	moonDec         float64
	moon            occultationRiseSetBody
	target          occultationRiseSetBody
	targetRadiusKM  float64
	internalContact bool
	valid           bool
	states          map[occultationRiseSetStateKey]occultationRiseSetState
}

type occultationRiseSetStateKey struct {
	longitude uint64
	latitude  uint64
}

type occultationRiseSetState struct {
	// contactMetric is selected for the context that reads the cached state.
	// The two underlying metrics are retained so external and internal contact
	// contexts can share the expensive topocentric geometry without sharing the
	// wrong contact equation.
	contactMetric         float64
	externalContactMetric float64
	internalContactMetric float64
	separationSquared     float64
	moonAltitude          float64
	valid                 bool
}

type occultationRiseSetEvaluation struct {
	tt     float64
	center occultationRiseSetContext
	before occultationRiseSetContext
	after  occultationRiseSetContext
}

type occultationRiseSetCurveKey struct {
	phase     RiseSetPhase
	direction RiseSetDirection
}

type occultationRiseSetTrack struct {
	segments [][]OccultationPathPoint
}

type occultationRiseSetContextFunc func(float64) occultationRiseSetContext

// occultationRiseSetEvaluationCache keeps ephemeris contexts and their
// center/before/after derivative bundle at the event level. Root refinements
// revisit the same TT values many times, especially while joining branches.
type occultationRiseSetEvaluationCache struct {
	contextAt            occultationRiseSetContextFunc
	candidateContextAt   occultationRiseSetContextFunc
	contexts             map[uint64]occultationRiseSetContext
	candidateContexts    map[uint64]occultationRiseSetContext
	evaluations          map[uint64]occultationRiseSetEvaluation
	candidateEvaluations map[uint64]occultationRiseSetEvaluation
}

func newOccultationRiseSetEvaluationCache(contextAt occultationRiseSetContextFunc) *occultationRiseSetEvaluationCache {
	return newOccultationRiseSetEvaluationCacheWithCandidate(contextAt, nil)
}

func newOccultationRiseSetEvaluationCacheWithCandidate(
	contextAt, candidateContextAt occultationRiseSetContextFunc,
) *occultationRiseSetEvaluationCache {
	return &occultationRiseSetEvaluationCache{
		contextAt:            contextAt,
		candidateContextAt:   candidateContextAt,
		contexts:             make(map[uint64]occultationRiseSetContext),
		candidateContexts:    make(map[uint64]occultationRiseSetContext),
		evaluations:          make(map[uint64]occultationRiseSetEvaluation),
		candidateEvaluations: make(map[uint64]occultationRiseSetEvaluation),
	}
}

func (cache *occultationRiseSetEvaluationCache) context(tt float64) occultationRiseSetContext {
	key := math.Float64bits(tt)
	if context, ok := cache.contexts[key]; ok {
		return context
	}
	context := cache.contextAt(tt)
	if len(cache.contexts) >= occultationRiseSetEvaluationCacheMaximumEntries {
		clearOccultationRiseSetContexts(cache.contexts)
		clearOccultationRiseSetEvaluations(cache.evaluations)
	}
	cache.contexts[key] = context
	return context
}

func (cache *occultationRiseSetEvaluationCache) candidateContext(tt float64) occultationRiseSetContext {
	if cache.candidateContextAt == nil {
		return cache.context(tt)
	}
	key := math.Float64bits(tt)
	if context, ok := cache.candidateContexts[key]; ok {
		return context
	}
	context := cache.candidateContextAt(tt)
	if len(cache.candidateContexts) >= occultationRiseSetEvaluationCacheMaximumEntries {
		clearOccultationRiseSetContexts(cache.candidateContexts)
		clearOccultationRiseSetEvaluations(cache.candidateEvaluations)
	}
	cache.candidateContexts[key] = context
	return context
}

func (cache *occultationRiseSetEvaluationCache) evaluation(tt float64) occultationRiseSetEvaluation {
	return cache.evaluationAt(tt, false)
}

func (cache *occultationRiseSetEvaluationCache) candidateEvaluation(tt float64) occultationRiseSetEvaluation {
	return cache.evaluationAt(tt, true)
}

func (cache *occultationRiseSetEvaluationCache) candidateOnly() *occultationRiseSetEvaluationCache {
	if cache == nil || cache.candidateContextAt == nil {
		return cache
	}
	return &occultationRiseSetEvaluationCache{
		contextAt:   cache.candidateContextAt,
		contexts:    cache.candidateContexts,
		evaluations: cache.candidateEvaluations,
	}
}

func (cache *occultationRiseSetEvaluationCache) evaluationAt(tt float64, candidate bool) occultationRiseSetEvaluation {
	key := math.Float64bits(tt)
	evaluations := cache.evaluations
	context := cache.context
	if candidate && cache.candidateContextAt != nil {
		evaluations = cache.candidateEvaluations
		context = cache.candidateContext
	}
	if evaluation, ok := evaluations[key]; ok {
		return evaluation
	}
	evaluation := occultationRiseSetEvaluation{
		tt:     tt,
		center: context(tt),
		before: context(tt - occultationRiseSetDerivativeStepDays),
		after:  context(tt + occultationRiseSetDerivativeStepDays),
	}
	if len(evaluations) >= occultationRiseSetEvaluationCacheMaximumEntries {
		clearOccultationRiseSetEvaluations(evaluations)
		if candidate {
			clearOccultationRiseSetContexts(cache.candidateContexts)
		} else {
			clearOccultationRiseSetContexts(cache.contexts)
		}
	}
	evaluations[key] = evaluation
	return evaluation
}

func clearOccultationRiseSetContexts(values map[uint64]occultationRiseSetContext) {
	for key := range values {
		delete(values, key)
	}
}

func clearOccultationRiseSetEvaluations(values map[uint64]occultationRiseSetEvaluation) {
	for key := range values {
		delete(values, key)
	}
}

func newOccultationRiseSetContext(
	tt, moonRA, moonDec, moonDistanceKM,
	targetRA, targetDec, targetDistanceKM, targetRadiusKM float64,
) occultationRiseSetContext {
	moon := newOccultationRiseSetBody(moonRA, moonDec, moonDistanceKM)
	target := newOccultationRiseSetBody(targetRA, targetDec, targetDistanceKM)
	return occultationRiseSetContext{
		tt:              tt,
		siderealDegrees: ApparentSiderealTime(TD2UT(tt, false)) * 15,
		moonRA:          moonRA,
		moonDec:         moonDec,
		moon:            moon,
		target:          target,
		targetRadiusKM:  targetRadiusKM,
		valid: finite(moonRA) && finite(moonDec) && finite(moonDistanceKM) && moonDistanceKM > 0 &&
			finite(targetRA) && finite(targetDec) && finite(targetDistanceKM) && targetDistanceKM >= 0 &&
			finite(targetRadiusKM) && targetRadiusKM >= 0,
		states: make(map[occultationRiseSetStateKey]occultationRiseSetState),
	}
}

func (context occultationRiseSetContext) withInternalContact() occultationRiseSetContext {
	context.internalContact = true
	return context
}

func newOccultationRiseSetContextFromVectors(
	tt float64,
	moonXYZ, targetXYZ [3]float64,
	targetAtFiniteDistance bool,
	targetRadiusKM float64,
) occultationRiseSetContext {
	moon, moonRA, moonDec, moonOK := occultationRiseSetBodyFromVector(moonXYZ, true)
	target, _, _, targetOK := occultationRiseSetBodyFromVector(targetXYZ, targetAtFiniteDistance)
	return occultationRiseSetContext{
		tt:              tt,
		siderealDegrees: ApparentSiderealTime(TD2UT(tt, false)) * 15,
		moonRA:          moonRA,
		moonDec:         moonDec,
		moon:            moon,
		target:          target,
		targetRadiusKM:  targetRadiusKM,
		valid:           moonOK && targetOK && finite(targetRadiusKM) && targetRadiusKM >= 0,
		states:          make(map[occultationRiseSetStateKey]occultationRiseSetState),
	}
}

func occultationRiseSetBodyFromVector(
	xyz [3]float64,
	atFiniteDistance bool,
) (occultationRiseSetBody, float64, float64, bool) {
	vector := occultationPathVector{x: xyz[0], y: xyz[1], z: xyz[2]}
	distanceKM := occultationPathNorm(vector)
	if !finite(distanceKM) || distanceKM <= 0 {
		return occultationRiseSetBody{}, 0, 0, false
	}
	direction := occultationPathScale(vector, 1/distanceKM)
	ra := normalizeRA(math.Atan2(direction.y, direction.x) / rad)
	dec := math.Asin(math.Max(-1, math.Min(1, direction.z))) / rad
	body := occultationRiseSetBody{direction: direction}
	if atFiniteDistance {
		body.positionKM = vector
		body.distanceKM = distanceKM
	}
	return body, ra, dec, finite(ra) && finite(dec)
}

func newOccultationRiseSetBody(ra, dec, distanceKM float64) occultationRiseSetBody {
	direction := occultationPathRaDecVector(ra, dec, 1)
	body := occultationRiseSetBody{direction: direction, distanceKM: distanceKM}
	if distanceKM > 0 {
		body.positionKM = occultationPathScale(direction, distanceKM)
	}
	return body
}

func (context occultationRiseSetContext) stateAt(longitude, latitude float64) occultationRiseSetState {
	key := occultationRiseSetStateKey{math.Float64bits(longitude), math.Float64bits(latitude)}
	if state, ok := context.states[key]; ok {
		return state.withContactMetric(context.internalContact)
	}
	if !context.valid {
		return context.storeState(key, occultationRiseSetState{})
	}
	observerParallaxKM, observerDistanceKM, zenith := occultationRiseSetObserverVectors(
		context.siderealDegrees, longitude, latitude,
	)
	moonDirection := occultationRiseSetTopocentricDirection(context.moon, observerParallaxKM)
	targetDirection := occultationRiseSetTopocentricDirection(context.target, observerParallaxKM)
	moonDistanceKM := occultationRiseSetTopocentricDistance(context.moon, observerDistanceKM)
	if !finite(moonDistanceKM) || moonDistanceKM <= moonEquatorialRadiusKM {
		return context.storeState(key, occultationRiseSetState{})
	}
	moonRadius := angularSemidiameterArcsec(moonEquatorialRadiusKM, moonDistanceKM) / 3600
	targetRadius := 0.0
	if context.targetRadiusKM > 0 {
		targetDistanceKM := occultationRiseSetTopocentricDistance(context.target, observerDistanceKM)
		if !finite(targetDistanceKM) || targetDistanceKM <= context.targetRadiusKM {
			return context.storeState(key, occultationRiseSetState{})
		}
		targetRadius = angularSemidiameterArcsec(context.targetRadiusKM, targetDistanceKM) / 3600
	}
	cosSeparation := math.Max(-1, math.Min(1, occultationPathDot(moonDirection, targetDirection)))
	separationRad := math.Acos(cosSeparation)
	separation := separationRad / rad
	moonAltitude := math.Asin(math.Max(-1, math.Min(1, occultationPathDot(moonDirection, zenith)))) / rad
	contactState := movingDiskContactState{
		separation:           separation,
		occultingOuterRadius: moonRadius,
		occultingInnerRadius: moonRadius,
		targetRadius:         targetRadius,
		valid: movingDiskContactStateValid(
			separation, moonRadius, moonRadius, targetRadius,
		),
	}
	externalContactMetric := contactState.externalContactGap()
	internalContactMetric := contactState.internalContactGap()
	return context.storeState(key, occultationRiseSetState{
		externalContactMetric: externalContactMetric,
		internalContactMetric: internalContactMetric,
		separationSquared:     2 - 2*cosSeparation,
		moonAltitude:          moonAltitude,
		valid:                 contactState.valid && finite(moonAltitude),
	})
}

// storeState 记住一个站点状态；表满时整表清空后复用，避免为每个解分配新 map。
func (context occultationRiseSetContext) storeState(
	key occultationRiseSetStateKey,
	state occultationRiseSetState,
) occultationRiseSetState {
	if context.states != nil {
		// Newton steps reuse a few nearby stations, not the full history.
		// Keep the small map allocation and recycle it as the solve moves.
		if len(context.states) >= occultationRiseSetStateCacheMaximumEntries {
			for cachedKey := range context.states {
				delete(context.states, cachedKey)
			}
		}
		context.states[key] = state
	}
	return state.withContactMetric(context.internalContact)
}

func (state occultationRiseSetState) withContactMetric(internal bool) occultationRiseSetState {
	if internal {
		state.contactMetric = state.internalContactMetric
	} else {
		state.contactMetric = state.externalContactMetric
	}
	return state
}

func (context occultationRiseSetContext) moonHorizonResidual(longitude, latitude float64) (float64, bool) {
	if !context.valid || context.moon.distanceKM <= 0 {
		return 0, false
	}
	latitudeRad := latitude * rad
	theta := (context.siderealDegrees + longitude) * rad
	sinLatitude, cosLatitude := math.Sincos(latitudeRad)
	sinTheta, cosTheta := math.Sincos(theta)
	moonDotZenith := cosLatitude*(context.moon.positionKM.x*cosTheta+context.moon.positionKM.y*sinTheta) +
		context.moon.positionKM.z*sinLatitude
	polarRatio := occultationTopocentricEarthPolarRadiusKM / occultationTopocentricEarthRadiusKM
	shapeDotZenith := math.Sqrt(cosLatitude*cosLatitude + polarRatio*polarRatio*sinLatitude*sinLatitude)
	residual := moonDotZenith - occultationLegacyParallaxRadiusKM*shapeDotZenith
	return residual, finite(residual)
}

func occultationRiseSetObserverVectors(
	siderealDegrees, longitude, latitude float64,
) (occultationPathVector, occultationPathVector, occultationPathVector) {
	latitudeRad := latitude * rad
	theta := (siderealDegrees + longitude) * rad
	sinLatitude, cosLatitude := math.Sincos(latitudeRad)
	sinTheta, cosTheta := math.Sincos(theta)
	polarRatio := occultationTopocentricEarthPolarRadiusKM / occultationTopocentricEarthRadiusKM
	u := math.Atan(polarRatio * math.Tan(latitudeRad))
	sinU, cosU := math.Sincos(u)
	shape := occultationPathVector{
		x: cosU * cosTheta,
		y: cosU * sinTheta,
		z: polarRatio * sinU,
	}
	zenith := occultationPathVector{x: cosLatitude * cosTheta, y: cosLatitude * sinTheta, z: sinLatitude}
	return occultationPathScale(shape, occultationLegacyParallaxRadiusKM),
		occultationPathScale(shape, occultationTopocentricEarthRadiusKM), zenith
}

func occultationRiseSetTopocentricDirection(
	body occultationRiseSetBody,
	observerKM occultationPathVector,
) occultationPathVector {
	if body.distanceKM <= 0 {
		return body.direction
	}
	return occultationPathUnit(occultationPathSub(body.positionKM, observerKM))
}

func occultationRiseSetTopocentricDistance(
	body occultationRiseSetBody,
	observerKM occultationPathVector,
) float64 {
	if body.distanceKM <= 0 {
		return math.Inf(1)
	}
	return occultationPathNorm(occultationPathSub(body.positionKM, observerKM))
}

func occultationRiseSetCurves(
	startTT, endTT, greatestTT float64,
	options OccultationPathOptions,
	location *time.Location,
	contextAt occultationRiseSetContextFunc,
) []OccultationRiseSetCurve {
	cache := newOccultationRiseSetEvaluationCache(contextAt)
	return occultationRiseSetCurvesWithCache(startTT, endTT, greatestTT, options, location, cache)
}

func occultationRiseSetCurvesWithCache(
	startTT, endTT, greatestTT float64,
	options OccultationPathOptions,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) []OccultationRiseSetCurve {
	curves, _, _ := occultationRiseSetCurvesWithRecoveryReport(
		startTT, endTT, greatestTT, options, location, cache,
	)
	return curves
}

// occultationRiseSetCurvesWithRecoveryReport 额外返回折点补根重建前的相位图和重建候选，
// 便于调用方核验“窗口内部未成对端点”判据；正常路径只使用第一个返回值。
// occultationRiseSetCurvesWithRecoveryReport also returns the phase graph before fold
// recovery and the rebuilt candidate, so callers can verify the interior-endpoint test;
// the normal path uses the first result only.
func occultationRiseSetCurvesWithRecoveryReport(
	startTT, endTT, greatestTT float64,
	options OccultationPathOptions,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
) ([]OccultationRiseSetCurve, []OccultationRiseSetCurve, []OccultationRiseSetCurve) {
	if options.DisableRiseSet || startTT == 0 || endTT == 0 || endTT <= startTT {
		return nil, nil, nil
	}
	if cache == nil {
		return nil, nil, nil
	}
	step := options.RiseSetStep
	if step <= 0 {
		step = 5 * time.Minute
	}
	stepDays := float64(step) / float64(24*time.Hour)
	times := occultationPathSampleTimes(startTT, endTT, greatestTT, stepDays)
	keys := []occultationRiseSetCurveKey{
		{RiseSetPhaseStart, RiseSetDirectionRise},
		{RiseSetPhaseStart, RiseSetDirectionSet},
		{RiseSetPhaseGreatest, RiseSetDirectionRise},
		{RiseSetPhaseGreatest, RiseSetDirectionSet},
		{RiseSetPhaseEnd, RiseSetDirectionRise},
		{RiseSetPhaseEnd, RiseSetDirectionSet},
	}
	// The ordinary grid is intentionally retained for performance. Near a
	// grazing phase junction a root branch may exist for less than one grid
	// interval, so evaluate only intervals whose root topology changes and
	// bisect those locally. This recovers the missing branch without doubling
	// the cost of every rise/set curve.
	samples := make([]occultationRiseSetSampledPoints, 0, len(times))
	evaluate := func(tt float64, foldRecovery bool) occultationRiseSetSampledPoints {
		return occultationRiseSetSampledPoints{
			tt: tt,
			points: cache.candidateEvaluation(tt).pointsAt(
				occultationRiseSetBoundaryPoints, location, foldRecovery,
			),
		}
	}
	pointSource := cache.candidateEvaluation(startTT).center.targetRadiusKM <= 0
	for _, tt := range times {
		samples = append(samples, evaluate(tt, false))
	}
	buildCurves := func(samples []occultationRiseSetSampledPoints) []OccultationRiseSetCurve {
		tracks := make(map[occultationRiseSetCurveKey][]*occultationRiseSetTrack, len(keys))
		curves := make([]OccultationRiseSetCurve, 0, len(keys))
		for _, sample := range samples {
			for _, key := range keys {
				if sample.adaptive && key.phase != RiseSetPhaseGreatest {
					continue
				}
				tracks[key] = appendOccultationRiseSetSamples(tracks[key], sample.points[key], stepDays)
			}
		}
		for _, key := range keys {
			segments := make([][]OccultationPathPoint, 0, len(tracks[key]))
			for _, track := range tracks[key] {
				for _, segment := range track.segments {
					if len(segment) >= 1 {
						segments = append(segments, segment)
					}
				}
			}
			if len(segments) > 0 {
				curves = append(curves, OccultationRiseSetCurve{
					Phase: key.phase, Direction: key.direction, Segments: segments,
				})
			}
		}
		completeOccultationRiseSetCurveEndpoints(curves, stepDays, location, cache)
		return curves
	}
	for pass := 0; pointSource && pass < occultationRiseSetAdaptiveMaximumPasses; pass++ {
		if len(samples) < 2 {
			break
		}
		intervals := make([]int, 0, occultationRiseSetAdaptiveMaximumIntervals)
		for index := 0; index+1 < len(samples) && len(intervals) < occultationRiseSetAdaptiveMaximumIntervals; index++ {
			left, right := samples[index], samples[index+1]
			if right.tt <= left.tt || right.tt-left.tt <= occultationRiseSetAdaptiveMinimumStepDays {
				continue
			}
			if !occultationRiseSetSamplesChangeTopology(left.points, right.points, keys) {
				continue
			}
			// Refine only when the interval hides an additional branch at its
			// interior. A simple 0->1 appearance is already represented by the
			// endpoint sample; refining every disappearance interval can alter
			// branch assignment for polar events whose midpoint has no root.
			midpoint := evaluate((left.tt+right.tt)/2, false)
			hiddenBranch := false
			for _, key := range keys {
				if len(midpoint.points[key]) > len(left.points[key]) &&
					len(midpoint.points[key]) > len(right.points[key]) {
					hiddenBranch = true
					break
				}
			}
			if hiddenBranch {
				intervals = append(intervals, index)
			}
		}
		if len(intervals) == 0 {
			break
		}
		refined := make([]occultationRiseSetSampledPoints, 0, len(samples)+len(intervals))
		intervalSet := make(map[int]bool, len(intervals))
		for _, index := range intervals {
			intervalSet[index] = true
		}
		for index, sample := range samples {
			refined = append(refined, sample)
			if !intervalSet[index] || index+1 >= len(samples) {
				continue
			}
			midpoint := evaluate((sample.tt+samples[index+1].tt)/2, false)
			midpoint.adaptive = true
			refined = append(refined, midpoint)
		}
		samples = refined
	}
	firstTT, lastTT := times[0], times[len(times)-1]
	curves := buildCurves(samples)
	baseCurves := curves
	recoveredCurves := occultationRiseSetFoldRecoveryCandidate(samples, buildCurves, evaluate, curves, firstTT, lastTT)
	if recoveredCurves != nil &&
		occultationRiseSetUnclosedEndpointCount(recoveredCurves, true, firstTT, lastTT) <
			occultationRiseSetUnclosedEndpointCount(baseCurves, true, firstTT, lastTT) {
		return recoveredCurves, baseCurves, recoveredCurves
	}
	return baseCurves, baseCurves, recoveredCurves
}

// occultationRiseSetFoldRecoveryCandidate 在窗口内部仍有未成对端点时用折点补根重建相位图。
// 没有窗口内部未成对端点时直接返回 nil：查询窗口边界上的裁剪端点没有任何折点可以配对，
// 为它们重跑整条流水线只会被丢弃。
// occultationRiseSetFoldRecoveryCandidate rebuilds the phase graph with fold roots while
// interior endpoints are still unpaired, and returns nil when none is: an endpoint clipped
// by the query window has no fold to pair with, so rebuilding for it is discarded work.
func occultationRiseSetFoldRecoveryCandidate(
	samples []occultationRiseSetSampledPoints,
	buildCurves func([]occultationRiseSetSampledPoints) []OccultationRiseSetCurve,
	evaluate func(float64, bool) occultationRiseSetSampledPoints,
	base []OccultationRiseSetCurve,
	firstTT, lastTT float64,
) []OccultationRiseSetCurve {
	if occultationRiseSetUnclosedEndpointCount(base, true, firstTT, lastTT) == 0 {
		return nil
	}
	// A polar tangent fold dropped by the sign scan leaves the two branches of a
	// phase unjoined. Events that already close keep their existing sampling and
	// endpoints byte-identical.
	recovered := make([]occultationRiseSetSampledPoints, len(samples))
	for index, sample := range samples {
		recovered[index] = evaluate(sample.tt, true)
		recovered[index].adaptive = sample.adaptive
	}
	return buildCurves(recovered)
}

type occultationRiseSetSampledPoints struct {
	tt       float64
	adaptive bool
	points   map[occultationRiseSetCurveKey][]OccultationPathPoint
}

func occultationRiseSetSamplesChangeTopology(
	left, right map[occultationRiseSetCurveKey][]OccultationPathPoint,
	keys []occultationRiseSetCurveKey,
) bool {
	for _, key := range keys {
		leftCount := len(left[key])
		rightCount := len(right[key])
		if leftCount != rightCount {
			return leftCount > 0 || rightCount > 0
		}
	}
	return false
}

func (evaluation occultationRiseSetEvaluation) pointsAt(
	boundaryPoints int,
	location *time.Location,
	foldRecovery bool,
) map[occultationRiseSetCurveKey][]OccultationPathPoint {
	result := make(map[occultationRiseSetCurveKey][]OccultationPathPoint, 6)
	if !evaluation.center.valid || !evaluation.before.valid || !evaluation.after.valid {
		return result
	}
	gst := ApparentSiderealTime(TD2UT(evaluation.tt, false)) * 15
	centerLongitude := normalizeLongitude(evaluation.center.moonRA - gst)
	centerLatitude := evaluation.center.moonDec
	appendRoots := func(greatest bool) {
		// Keep the historical geographic refinement for ordinary latitudes; it
		// is both cheaper and preserves the established endpoint network.  A
		// polar root needs the station-corrected horizon parameterization because
		// its geographic Jacobian becomes singular and can switch siblings.
		valueAt := func(angle float64) (float64, bool) {
			longitude, latitude := riseSetHorizonPoint(centerLongitude, centerLatitude, angle)
			state := evaluation.center.stateAt(longitude, latitude)
			if !state.valid {
				return 0, false
			}
			if greatest {
				return evaluation.separationDerivative(longitude, latitude), true
			}
			return state.contactMetric, true
		}
		foldTolerance := 0.0
		if foldRecovery && !greatest {
			foldTolerance = riseSetFoldRootResidualToleranceDeg
		}
		roots := riseSetCyclicRootsWithFoldTolerance(boundaryPoints, foldTolerance, valueAt)
		polar := false
		for _, angle := range roots {
			_, latitude := riseSetHorizonPoint(centerLongitude, centerLatitude, angle)
			if math.Abs(latitude) >= 70 {
				polar = true
				break
			}
		}
		var stationHorizonRoots []float64
		if polar {
			stationValueAt := func(angle float64) (float64, bool) {
				longitude, latitude, ok := occultationRiseSetHorizonPointFromContext(evaluation.center, angle)
				if !ok {
					return 0, false
				}
				state := evaluation.center.stateAt(longitude, latitude)
				if !state.valid {
					return 0, false
				}
				if greatest {
					return evaluation.separationDerivative(longitude, latitude), true
				}
				return state.contactMetric, true
			}
			stationHorizonRoots = riseSetCyclicRootsWithFoldTolerance(boundaryPoints, foldTolerance, stationValueAt)
		}
		for _, angle := range roots {
			var longitude, latitude float64
			var ok bool
			longitude, latitude = riseSetHorizonPoint(centerLongitude, centerLatitude, angle)
			if math.Abs(latitude) >= 70 {
				stationAngle, found := occultationRiseSetClosestHorizonAngle(angle, stationHorizonRoots)
				if !found {
					continue
				}
				// `angle` is the continuation parameter on the station-corrected
				// horizon.  Do not refine this root in longitude/latitude: near a
				// polar fold that 2-D Newton system can converge to its sibling.
				longitude, latitude, ok = occultationRiseSetHorizonPointFromContext(evaluation.center, stationAngle)
			} else {
				longitude, latitude, ok = riseSetRefineGeographicRoot(
					longitude, latitude,
					func(lon, lat float64) (float64, float64, bool) {
						state := evaluation.center.stateAt(lon, lat)
						if !state.valid {
							return 0, 0, false
						}
						first := state.contactMetric
						if greatest {
							first = evaluation.separationDerivative(lon, lat)
						}
						return first, state.moonAltitude, finite(first)
					},
				)
			}
			if !ok {
				continue
			}
			point, key, valid := evaluation.classify(longitude, latitude, greatest, location)
			if !valid || occultationRiseSetPointExists(result[key], point) {
				continue
			}
			result[key] = append(result[key], point)
		}
	}
	appendRoots(false)
	appendRoots(true)
	return result
}

func occultationRiseSetClosestHorizonAngle(target float64, candidates []float64) (float64, bool) {
	closest := 0.0
	distance := math.Inf(1)
	for _, candidate := range candidates {
		current := riseSetAngularDistance(target, candidate)
		if current < distance {
			closest, distance = candidate, current
		}
	}
	return closest, finite(distance) && distance <= math.Pi/6
}

func (evaluation occultationRiseSetEvaluation) classify(
	longitude, latitude float64,
	greatest bool,
	location *time.Location,
) (OccultationPathPoint, occultationRiseSetCurveKey, bool) {
	state := evaluation.center.stateAt(longitude, latitude)
	altitudeDerivative := evaluation.moonAltitudeDerivative(longitude, latitude)
	if !state.valid || !finite(altitudeDerivative) || math.Abs(altitudeDerivative) < 1e-8 {
		return OccultationPathPoint{}, occultationRiseSetCurveKey{}, false
	}
	direction := RiseSetDirectionSet
	if altitudeDerivative > 0 {
		direction = RiseSetDirectionRise
	}
	phase := RiseSetPhaseGreatest
	if greatest {
		if state.contactMetric > 1e-7 || evaluation.separationSecondDerivative(longitude, latitude) <= 0 {
			return OccultationPathPoint{}, occultationRiseSetCurveKey{}, false
		}
	} else {
		contactDerivative := evaluation.contactDerivative(longitude, latitude)
		if !finite(contactDerivative) || math.Abs(contactDerivative) < 1e-8 {
			return OccultationPathPoint{}, occultationRiseSetCurveKey{}, false
		}
		phase = RiseSetPhaseEnd
		if contactDerivative < 0 {
			phase = RiseSetPhaseStart
		}
	}
	return OccultationPathPoint{
		Time:      occultationTTToLocation(evaluation.tt, location),
		Longitude: longitude, Latitude: latitude, MoonAltitude: state.moonAltitude,
	}, occultationRiseSetCurveKey{phase: phase, direction: direction}, true
}

func (evaluation occultationRiseSetEvaluation) contactDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude, latitude)
	after := evaluation.after.stateAt(longitude, latitude)
	if !before.valid || !after.valid {
		return math.NaN()
	}
	return (after.contactMetric - before.contactMetric) / (2 * occultationRiseSetDerivativeStepDays)
}

func (evaluation occultationRiseSetEvaluation) contactSecondDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude, latitude)
	center := evaluation.center.stateAt(longitude, latitude)
	after := evaluation.after.stateAt(longitude, latitude)
	if !before.valid || !center.valid || !after.valid {
		return math.NaN()
	}
	stepSquared := occultationRiseSetDerivativeStepDays * occultationRiseSetDerivativeStepDays
	return (after.contactMetric - 2*center.contactMetric + before.contactMetric) / stepSquared
}

func (evaluation occultationRiseSetEvaluation) separationDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude, latitude)
	after := evaluation.after.stateAt(longitude, latitude)
	if !before.valid || !after.valid {
		return math.NaN()
	}
	return (after.separationSquared - before.separationSquared) / (2 * occultationRiseSetDerivativeStepDays)
}

func (evaluation occultationRiseSetEvaluation) separationSecondDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude, latitude)
	center := evaluation.center.stateAt(longitude, latitude)
	after := evaluation.after.stateAt(longitude, latitude)
	if !before.valid || !center.valid || !after.valid {
		return math.NaN()
	}
	stepSquared := occultationRiseSetDerivativeStepDays * occultationRiseSetDerivativeStepDays
	return (after.separationSquared - 2*center.separationSquared + before.separationSquared) / stepSquared
}

func (evaluation occultationRiseSetEvaluation) moonAltitudeDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude, latitude)
	after := evaluation.after.stateAt(longitude, latitude)
	if !before.valid || !after.valid {
		return math.NaN()
	}
	return (after.moonAltitude - before.moonAltitude) / (2 * occultationRiseSetDerivativeStepDays)
}

func (evaluation occultationRiseSetEvaluation) moonAltitudeSecondDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude, latitude)
	center := evaluation.center.stateAt(longitude, latitude)
	after := evaluation.after.stateAt(longitude, latitude)
	if !before.valid || !center.valid || !after.valid {
		return math.NaN()
	}
	stepSquared := occultationRiseSetDerivativeStepDays * occultationRiseSetDerivativeStepDays
	return (after.moonAltitude - 2*center.moonAltitude + before.moonAltitude) / stepSquared
}

func occultationRiseSetPointExists(points []OccultationPathPoint, candidate OccultationPathPoint) bool {
	for _, point := range points {
		if occultationPathDistanceKM(point, candidate) < 0.01 {
			return true
		}
	}
	return false
}

// occultationRiseSetBranchChanged is stricter than the coarse solar-path
// splitter for short occultation segments. A polar rise/set root can move
// hundreds of kilometres in nearly the same timestamp when two unrelated
// horizon branches are paired. Distances up to 500 km remain valid for the
// sampled fold attachments; a larger sub-second jump is rejected as a branch
// change before adaptive refinement can reconnect it.
func occultationRiseSetBranchChanged(distanceKM, deltaDays float64) bool {
	if !finite(distanceKM) || !finite(deltaDays) {
		return true
	}
	if distanceKM <= 500 {
		return false
	}
	seconds := deltaDays * 86400
	// Equal-time endpoints can only be joined when they are already spatially
	// close. A larger jump in the sub-second interval is a different horizon
	// root, not a fast-moving physical branch.
	if seconds <= 0 {
		return true
	}
	return seconds < 1 && distanceKM/seconds > 10
}

// occultationRiseSetWindowEdgeToleranceDays 判定端点是否落在查询窗口边界上。
// 窗口边界样本经过民用时往返的误差在微秒量级，远小于任何采样步长。
const occultationRiseSetWindowEdgeToleranceDays = 1e-6

// occultationRiseSetUnclosedEndpointCount 统计未被其他相位曲线端点共享的端点数量，
// 这是“相位图是否闭合”的判据。interiorOnly 时跳过窗口裁剪端点：它们没有任何折点
// 可以配对，只反映查询窗口而不是几何缺陷。
// occultationRiseSetUnclosedEndpointCount counts the endpoints that no other phase
// curve endpoint shares, the closure test for the phase graph. With interiorOnly,
// endpoints clipped by the query window are skipped: no fold can pair them, so they
// describe the window rather than the geometry.
func occultationRiseSetUnclosedEndpointCount(
	curves []OccultationRiseSetCurve,
	interiorOnly bool,
	firstTT, lastTT float64,
) int {
	unclosed := 0
	for ci, curve := range curves {
		if curve.Phase == RiseSetPhaseGreatest {
			continue
		}
		for si, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			for _, endpoint := range []OccultationPathPoint{segment[0], segment[len(segment)-1]} {
				if interiorOnly && occultationRiseSetEndpointClipped(endpoint, firstTT, lastTT) {
					continue
				}
				shared := false
				for oi, other := range curves {
					if other.Phase == RiseSetPhaseGreatest {
						continue
					}
					for os, points := range other.Segments {
						if (ci == oi && si == os) || len(points) < 2 {
							continue
						}
						for _, point := range []OccultationPathPoint{points[0], points[len(points)-1]} {
							dt := point.Time.Sub(endpoint.Time)
							if dt >= -time.Second && dt <= time.Second &&
								occultationPathDistanceKM(endpoint, point) < 0.01 {
								shared = true
							}
						}
					}
				}
				if !shared {
					unclosed++
				}
			}
		}
	}
	return unclosed
}

func occultationRiseSetEndpointClipped(endpoint OccultationPathPoint, firstTT, lastTT float64) bool {
	tt := occultationTimeToTT(endpoint.Time)
	return tt <= firstTT+occultationRiseSetWindowEdgeToleranceDays ||
		tt >= lastTT-occultationRiseSetWindowEdgeToleranceDays
}

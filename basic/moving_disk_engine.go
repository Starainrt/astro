package basic

import (
	"math"
	"sort"
)

// movingDiskEventEngine contains the time-domain part shared by solar
// eclipses and lunar occultations.  Geometry remains in the caller: a solar
// adapter evaluates the Bessel projection, while an occultation adapter
// evaluates the Earth-vector frame.  Keeping that boundary explicit is what
// lets the solar implementation remain the numerical baseline.
type movingDiskEventEngine struct {
	maxSampleCount        int
	searchSpanDays        float64
	rangeStepDays         float64
	rootToleranceDays     float64
	greatestSpanDays      float64
	greatestToleranceDays float64
	reserveAnchorSlot     bool
	uniformOverflow       bool
	indexedSampleTimes    bool
}

// movingDiskContactState is the dimensionless circular-disk model shared by
// solar eclipse and lunar occultation adapters.  The radii and separation may
// be radians, degrees, or arcseconds, but all four values must use the same
// unit.  Solar eclipses provide distinct outer and inner occulting radii;
// finite-planet occultations normally use the same lunar radius for both.
type movingDiskContactState struct {
	separation           float64
	occultingOuterRadius float64
	occultingInnerRadius float64
	targetRadius         float64
	valid                bool
}

// movingDiskEventCacheMaximumEntries is shared by event-local caches so a
// dense path cannot grow without bound in native or TinyGo/WASM execution.
const movingDiskEventCacheMaximumEntries = 2048
const movingDiskContactCacheMaximumEntries = movingDiskEventCacheMaximumEntries

// movingDiskContactCache stores the disk state for one event evaluation.  The
// valid bit is part of the entry so failed ephemeris evaluations are cached as
// well; otherwise a pair of external/internal roots can repeat the same
// invalid star or planet calculation indefinitely.  Entries are keyed by the
// exact TT bits because root refinement intentionally revisits exact endpoints.
type movingDiskContactCache struct {
	entries    map[uint64]movingDiskContactCacheEntry
	maxEntries int
}

type movingDiskContactCacheEntry struct {
	state movingDiskContactState
	ok    bool
}

func newMovingDiskContactCache() *movingDiskContactCache {
	return &movingDiskContactCache{
		entries:    make(map[uint64]movingDiskContactCacheEntry),
		maxEntries: movingDiskContactCacheMaximumEntries,
	}
}

func (cache *movingDiskContactCache) lookup(tt float64) (movingDiskContactState, bool, bool) {
	if cache == nil || cache.entries == nil {
		return movingDiskContactState{}, false, false
	}
	entry, ok := cache.entries[math.Float64bits(tt)]
	if !ok {
		return movingDiskContactState{}, false, false
	}
	return entry.state, entry.ok, true
}

func (cache *movingDiskContactCache) store(tt float64, state movingDiskContactState, ok bool) {
	if cache == nil {
		return
	}
	if cache.entries == nil {
		cache.entries = make(map[uint64]movingDiskContactCacheEntry)
	}
	limit := cache.maxEntries
	if limit <= 0 {
		limit = movingDiskContactCacheMaximumEntries
	}
	key := math.Float64bits(tt)
	if _, exists := cache.entries[key]; !exists && len(cache.entries) >= limit {
		// Contact roots are local to one event and are naturally clustered in
		// time.  Clearing the bounded table is cheaper and more predictable
		// than maintaining an eviction list in TinyGo/WASM.
		for key := range cache.entries {
			delete(cache.entries, key)
		}
	}
	cache.entries[key] = movingDiskContactCacheEntry{state: state, ok: ok}
}

// movingDiskContactEvaluator adapts an event-specific geometry calculation to
// the shared contact root engine.  Its callback must be pure for a given TT;
// all observer/configuration values belong to the evaluator closure.
type movingDiskContactEvaluator struct {
	evaluate func(float64) (movingDiskContactState, bool)
	cache    *movingDiskContactCache
}

func newMovingDiskContactEvaluator(
	evaluate func(float64) (movingDiskContactState, bool),
) *movingDiskContactEvaluator {
	return &movingDiskContactEvaluator{
		evaluate: evaluate,
		cache:    newMovingDiskContactCache(),
	}
}

func (evaluator *movingDiskContactEvaluator) stateAt(tt float64) (movingDiskContactState, bool) {
	if evaluator == nil || evaluator.evaluate == nil || !finiteMovingDiskValue(tt) {
		return movingDiskContactState{}, false
	}
	if state, ok, hit := evaluator.cache.lookup(tt); hit {
		return state, ok
	}
	state, ok := evaluator.evaluate(tt)
	if !ok {
		state = movingDiskContactState{}
	}
	evaluator.cache.store(tt, state, ok)
	return state, ok
}

// prime inserts a state already computed by the caller.  Greatest-point
// evaluation often precedes both external and internal contact roots; priming
// avoids evaluating that same TT a second time while preserving the callback
// as the source of truth for all other samples.
func (evaluator *movingDiskContactEvaluator) prime(
	tt float64,
	state movingDiskContactState,
	ok bool,
) {
	if evaluator == nil || !finiteMovingDiskValue(tt) {
		return
	}
	evaluator.cache.store(tt, state, ok)
}

func (evaluator *movingDiskContactEvaluator) gap(tt float64, internal bool) (float64, bool) {
	state, ok := evaluator.stateAt(tt)
	if !ok || !state.valid {
		return 0, false
	}
	if internal {
		return state.internalContactGap(), true
	}
	return state.externalContactGap(), true
}

func (state movingDiskContactState) externalContactGap() float64 {
	if !state.valid {
		return math.NaN()
	}
	return state.separation - state.occultingOuterRadius - state.targetRadius
}

func (state movingDiskContactState) internalContactGap() float64 {
	if !state.valid {
		return math.NaN()
	}
	return state.separation - math.Abs(state.occultingInnerRadius-state.targetRadius)
}

func movingDiskContactStateValid(separation, outerRadius, innerRadius, targetRadius float64) bool {
	return finiteMovingDiskValue(separation) && separation >= 0 &&
		finiteMovingDiskValue(outerRadius) && outerRadius > 0 &&
		finiteMovingDiskValue(innerRadius) && innerRadius > 0 &&
		finiteMovingDiskValue(targetRadius) && targetRadius >= 0
}

func finiteMovingDiskValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func (engine movingDiskEventEngine) sampleTimes(
	start, end, greatest, requestedStep float64,
) ([]float64, float64) {
	if end < start {
		start, end = end, start
	}
	maximum := engine.maxSampleCount
	if maximum < 3 {
		maximum = 3
	}
	duration := end - start
	if duration <= 0 {
		return []float64{start}, requestedStep
	}
	step := requestedStep
	if step <= 0 || math.IsNaN(step) || math.IsInf(step, 0) {
		step = duration
	}
	// 先按浮点比较，避免极小步长下 int 转换溢出成负数而绕过上限。
	baseSampleCount := math.Ceil(duration/step) + 1
	if engine.reserveAnchorSlot {
		baseSampleCount++
	}
	if baseSampleCount > float64(maximum) && engine.uniformOverflow {
		interiorCount := maximum - 3
		if interiorCount < 0 {
			interiorCount = 0
		}
		bounded := make([]float64, 0, maximum)
		bounded = append(bounded, start, greatest, end)
		for index := 1; index <= interiorCount; index++ {
			bounded = append(bounded, start+duration*float64(index)/float64(interiorCount+1))
		}
		sort.Float64s(bounded)
		return movingDiskUniqueTimes(bounded), step
	}
	if baseSampleCount > float64(maximum) {
		step = duration / float64(maximum-1)
	}

	times := []float64{start, greatest, end}
	if engine.indexedSampleTimes {
		for index := 1; ; index++ {
			current := start + float64(index)*step
			if current >= end {
				break
			}
			times = append(times, current)
		}
	} else {
		for current := start + step; current < end; current += step {
			times = append(times, current)
		}
	}
	sort.Float64s(times)
	return movingDiskUniqueTimes(times), step
}

func (engine movingDiskEventEngine) window(
	seed, start, end float64,
	candidateAt, exactAt func(float64) bool,
) (float64, float64, bool) {
	if exactAt == nil {
		return 0, 0, false
	}
	left := start
	right := end
	if engine.searchSpanDays > 0 {
		left = math.Max(left, seed-engine.searchSpanDays)
		right = math.Min(right, seed+engine.searchSpanDays)
	}
	if right <= left {
		return 0, 0, false
	}
	if candidateAt == nil {
		candidateAt = exactAt
	}
	step := engine.rangeStepDays
	if step <= 0 || math.IsNaN(step) || math.IsInf(step, 0) {
		step = right - left
	}

	first := math.NaN()
	previous := left
	if candidateAt(previous) && exactAt(previous) {
		first = previous
	}
	if math.IsNaN(first) {
		for current := left + step; current <= right; current += step {
			current = math.Min(current, right)
			if candidateAt(current) {
				first = engine.refineTransition(previous, current, exactAt, false)
				break
			}
			previous = current
		}
	}
	if math.IsNaN(first) {
		return 0, 0, false
	}

	last := first
	previous = first
	for current := first + step; current <= right; current += step {
		current = math.Min(current, right)
		if !candidateAt(current) {
			last = engine.refineTransition(previous, current, exactAt, true)
			return first, last, true
		}
		last = current
		previous = current
	}
	return first, right, true
}

func (engine movingDiskEventEngine) refineTransition(
	left, right float64,
	predicate func(float64) bool,
	trueToFalse bool,
) float64 {
	leftOK := predicate(left)
	tolerance := engine.rootToleranceDays
	if tolerance <= 0 || math.IsNaN(tolerance) || math.IsInf(tolerance, 0) {
		tolerance = 1e-10
	}
	for iteration := 0; iteration < 48 && math.Abs(right-left) > tolerance; iteration++ {
		middle := (left + right) / 2
		middleOK := predicate(middle)
		if trueToFalse {
			if middleOK {
				left = middle
			} else {
				right = middle
			}
			continue
		}
		if middleOK {
			right = middle
		} else {
			left = middle
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

// greatest returns the minimum impact value in the event-local interval.
// The callback returns (impact, valid); invalid states are treated as +Inf.
func (engine movingDiskEventEngine) greatest(
	seed, start, end float64,
	impactAt func(float64) (float64, bool),
	iterations int,
) float64 {
	if impactAt == nil {
		return seed
	}
	left := start
	right := end
	span := engine.greatestSpanDays
	if span > 0 {
		left = math.Max(left, seed-span)
		right = math.Min(right, seed+span)
	}
	if right <= left {
		return seed
	}
	if iterations <= 0 {
		iterations = 56
	}
	const goldenRatio = 0.6180339887498949
	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1 := movingDiskImpact(impactAt, x1)
	f2 := movingDiskImpact(impactAt, x2)
	for iteration := 0; iteration < iterations &&
		(engine.greatestToleranceDays <= 0 || right-left > engine.greatestToleranceDays); iteration++ {
		if f1 > f2 {
			left = x1
			x1, f1 = x2, f2
			x2 = left + goldenRatio*(right-left)
			f2 = movingDiskImpact(impactAt, x2)
			continue
		}
		right = x2
		x2, f2 = x1, f1
		x1 = right - goldenRatio*(right-left)
		f1 = movingDiskImpact(impactAt, x1)
	}
	return (left + right) / 2
}

// contactRoot finds one external or internal disk-contact root by walking
// away from greatest and then bisecting the first valid sign change.  The
// callback may reject an ephemeris sample; rejected samples are skipped while
// searching, but an invalid value inside a confirmed bisection bracket aborts
// that root rather than inventing a crossing.
func (engine movingDiskEventEngine) contactRoot(
	greatest, direction, stepDays, spanDays, tolerance float64,
	metric func(float64) (float64, bool),
	iterations int,
) (float64, bool) {
	if metric == nil || (direction != -1 && direction != 1) ||
		stepDays <= 0 || spanDays <= 0 || tolerance <= 0 {
		return 0, false
	}
	nearTT := greatest
	nearValue, nearOK := metric(nearTT)
	if !nearOK || !finiteMovingDiskValue(nearValue) || nearValue > 0 {
		return 0, false
	}
	maxSteps := int(math.Ceil(spanDays / stepDays))
	if maxSteps < 1 {
		maxSteps = 1
	}
	for index := 1; index <= maxSteps; index++ {
		farTT := greatest + direction*float64(index)*stepDays
		farValue, farOK := metric(farTT)
		if !farOK || !finiteMovingDiskValue(farValue) {
			continue
		}
		if farValue < 0 {
			nearTT, nearValue = farTT, farValue
			continue
		}
		return movingDiskContactBracketRoot(
			nearTT, farTT, nearValue, farValue, metric, tolerance, iterations,
		)
	}
	return 0, false
}

func movingDiskContactBracketRoot(
	left, right, leftValue, rightValue float64,
	metric func(float64) (float64, bool),
	tolerance float64,
	iterations int,
) (float64, bool) {
	if left > right {
		left, right = right, left
		leftValue, rightValue = rightValue, leftValue
	}
	if !finiteMovingDiskValue(leftValue) || !finiteMovingDiskValue(rightValue) ||
		leftValue*rightValue > 0 {
		return 0, false
	}
	if leftValue == 0 {
		return left, true
	}
	if rightValue == 0 {
		return right, true
	}
	if iterations <= 0 {
		iterations = 64
	}
	for index := 0; index < iterations && right-left > tolerance; index++ {
		middle := (left + right) / 2
		middleValue, ok := metric(middle)
		if !ok || !finiteMovingDiskValue(middleValue) {
			return 0, false
		}
		if leftValue*middleValue <= 0 {
			right, rightValue = middle, middleValue
		} else {
			left, leftValue = middle, middleValue
		}
	}
	return (left + right) / 2, true
}

func movingDiskImpact(impactAt func(float64) (float64, bool), tt float64) float64 {
	value, ok := impactAt(tt)
	if !ok || math.IsNaN(value) || math.IsInf(value, 0) {
		return math.Inf(1)
	}
	return value
}

func movingDiskUniqueTimes(times []float64) []float64 {
	if len(times) < 2 {
		return times
	}
	unique := times[:1]
	for _, current := range times[1:] {
		if math.Abs(current-unique[len(unique)-1]) <= 1e-10 {
			continue
		}
		unique = append(unique, current)
	}
	return unique
}

func solarEclipseMovingDiskEngine() movingDiskEventEngine {
	return movingDiskEventEngine{
		maxSampleCount:        solarEclipsePathMaxSampleCount,
		rootToleranceDays:     solarEclipseShadowContactToleranceDays,
		greatestToleranceDays: localSolarEclipseGreatestTolerance,
	}
}

func occultationMovingDiskEngine() movingDiskEventEngine {
	return movingDiskEventEngine{
		maxSampleCount:     occultationPathMaxSampleCount,
		searchSpanDays:     occultationPathSearchSpanDays,
		rangeStepDays:      occultationPathRangeStepDays,
		rootToleranceDays:  occultationPathRootToleranceDays,
		greatestSpanDays:   0.75,
		reserveAnchorSlot:  true,
		uniformOverflow:    true,
		indexedSampleTimes: true,
	}
}

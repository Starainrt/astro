package basic

import "math"

const (
	moonPlanetConjunctionEstimateN         = 8
	moonPlanetConjunctionNearQueryDeltaDeg = 3.0
	moonPlanetConjunctionBracketStepDays   = 0.5
	moonPlanetConjunctionNearQueryStepDays = 0.25
	moonPlanetConjunctionNearQueryHalfSpan = 1.5
	moonPlanetConjunctionBracketHalfSpan   = 2.0
	moonPlanetConjunctionBracketGrowth     = 2.0
	moonPlanetConjunctionBracketAttempts   = 3
	moonPlanetConjunctionRefineStepDays    = 0.5 / 86400.0
	moonPlanetConjunctionEventTolerance    = 0.01
	moonPlanetConjunctionFallbackSpanScale = 1.5
)

type moonPlanetConjunctionLocalResult struct {
	lastUT float64
	nextUT float64
}

func emptyMoonPlanetConjunctionLocalResult() moonPlanetConjunctionLocalResult {
	return moonPlanetConjunctionLocalResult{
		lastUT: math.NaN(),
		nextUT: math.NaN(),
	}
}

// MoonPlanetConjunctionPlanet 月球合月目标行星 / target planet for Moon-planet conjunction events.
type MoonPlanetConjunctionPlanet int

const (
	MoonPlanetConjunctionMercury MoonPlanetConjunctionPlanet = iota + 1
	MoonPlanetConjunctionVenus
	MoonPlanetConjunctionMars
	MoonPlanetConjunctionJupiter
	MoonPlanetConjunctionSaturn
	MoonPlanetConjunctionUranus
	MoonPlanetConjunctionNeptune
)

func moonPlanetConjunctionWrappedDelta(diff float64) float64 {
	diff = math.Mod(diff+180, 360)
	if diff < 0 {
		diff += 360
	}
	return diff - 180
}

func moonPlanetConjunctionDeltaAt(jdTT float64, planet MoonPlanetConjunctionPlanet, n int) float64 {
	moonRA := HMoonGeocentricApparentRaN(jdTT, n)
	var planetRA float64
	switch planet {
	case MoonPlanetConjunctionMercury:
		planetRA = MercuryApparentRaN(jdTT, n)
	case MoonPlanetConjunctionVenus:
		planetRA = VenusApparentRaN(jdTT, n)
	case MoonPlanetConjunctionMars:
		planetRA = MarsApparentRaN(jdTT, n)
	case MoonPlanetConjunctionJupiter:
		planetRA = JupiterApparentRaN(jdTT, n)
	case MoonPlanetConjunctionSaturn:
		planetRA = SaturnApparentRaN(jdTT, n)
	case MoonPlanetConjunctionUranus:
		planetRA = UranusApparentRaN(jdTT, n)
	case MoonPlanetConjunctionNeptune:
		planetRA = NeptuneApparentRaN(jdTT, n)
	default:
		return math.NaN()
	}
	return moonPlanetConjunctionWrappedDelta(moonRA - planetRA)
}

func moonPlanetConjunctionPeriodDays(planet MoonPlanetConjunctionPlanet) float64 {
	switch planet {
	case MoonPlanetConjunctionMercury:
		return 28.1
	case MoonPlanetConjunctionVenus:
		return 28.4
	case MoonPlanetConjunctionMars:
		return 29.2
	case MoonPlanetConjunctionJupiter:
		return 28.0
	case MoonPlanetConjunctionSaturn:
		return 27.4
	case MoonPlanetConjunctionUranus:
		return 27.3
	case MoonPlanetConjunctionNeptune:
		return 27.3
	default:
		return math.NaN()
	}
}

func moonPlanetConjunctionInDirection(eventUT, queryTT float64, direction int) bool {
	switch direction {
	case -1:
		return eventUTQueryBeforeOrEqual(eventUT, queryTT)
	case 1:
		return eventUTQueryAfterOrEqual(eventUT, queryTT)
	default:
		return true
	}
}

func moonPlanetConjunctionFindBracket(centerTT, halfSpan, step float64, planet MoonPlanetConjunctionPlanet) (float64, float64, bool) {
	if math.IsNaN(centerTT) || math.IsNaN(halfSpan) || math.IsNaN(step) || halfSpan <= 0 || step <= 0 {
		return 0, 0, false
	}
	start := centerTT - halfSpan
	end := centerTT + halfSpan
	samples := int(math.Ceil((end-start)/step)) + 1
	prevTT := start
	prevVal := moonPlanetConjunctionDeltaAt(prevTT, planet, -1)
	if math.IsNaN(prevVal) {
		return 0, 0, false
	}
	if prevVal == 0 {
		return prevTT, prevTT, true
	}
	bestLeft := 0.0
	bestRight := 0.0
	bestDistance := math.Inf(1)
	for i := 1; i <= samples; i++ {
		tt := start + float64(i)*step
		if tt > end {
			tt = end
		}
		val := moonPlanetConjunctionDeltaAt(tt, planet, -1)
		if math.IsNaN(val) {
			return 0, 0, false
		}
		if val == 0 {
			return tt, tt, true
		}
		if prevVal*val < 0 {
			mid := (prevTT + tt) / 2.0
			distance := math.Abs(mid - centerTT)
			if distance < bestDistance {
				bestLeft = prevTT
				bestRight = tt
				bestDistance = distance
			}
		}
		if tt == end {
			break
		}
		prevTT = tt
		prevVal = val
	}
	if math.IsInf(bestDistance, 1) {
		return 0, 0, false
	}
	return bestLeft, bestRight, true
}

func moonPlanetConjunctionRefineBracket(leftTT, rightTT float64, planet MoonPlanetConjunctionPlanet) float64 {
	if leftTT > rightTT {
		leftTT, rightTT = rightTT, leftTT
	}
	if leftTT == rightTT {
		return leftTT
	}
	center := (leftTT + rightTT) / 2.0
	halfWindow := (rightTT - leftTT) / 2.0
	return eventZeroRefine(center, halfWindow, moonPlanetConjunctionRefineStepDays, func(sampleTT float64) float64 {
		return moonPlanetConjunctionDeltaAt(sampleTT, planet, -1)
	})
}

func moonPlanetConjunctionEventUT(leftTT, rightTT float64, planet MoonPlanetConjunctionPlanet) float64 {
	eventTT := moonPlanetConjunctionRefineBracket(leftTT, rightTT, planet)
	if math.Abs(moonPlanetConjunctionDeltaAt(eventTT, planet, -1)) > moonPlanetConjunctionEventTolerance {
		return math.NaN()
	}
	return TD2UT(eventTT, false)
}

func moonPlanetConjunctionCollectLocalEvent(result *moonPlanetConjunctionLocalResult, queryTT, eventUT float64) {
	if math.IsNaN(eventUT) {
		return
	}
	if eventUTQueryBeforeOrEqual(eventUT, queryTT) {
		if math.IsNaN(result.lastUT) || math.Abs(eventUTQueryTTDelta(eventUT, queryTT)) < math.Abs(eventUTQueryTTDelta(result.lastUT, queryTT)) {
			result.lastUT = eventUT
		}
	}
	if eventUTQueryAfterOrEqual(eventUT, queryTT) {
		if math.IsNaN(result.nextUT) || math.Abs(eventUTQueryTTDelta(eventUT, queryTT)) < math.Abs(eventUTQueryTTDelta(result.nextUT, queryTT)) {
			result.nextUT = eventUT
		}
	}
}

func moonPlanetConjunctionShouldCheckLocal(queryTT float64, planet MoonPlanetConjunctionPlanet) bool {
	delta := moonPlanetConjunctionDeltaAt(queryTT, planet, moonPlanetConjunctionEstimateN)
	if math.IsNaN(delta) {
		return false
	}
	return math.Abs(delta) <= moonPlanetConjunctionNearQueryDeltaDeg
}

func moonPlanetConjunctionLocalEvents(queryTT float64, planet MoonPlanetConjunctionPlanet) moonPlanetConjunctionLocalResult {
	result := emptyMoonPlanetConjunctionLocalResult()
	start := queryTT - moonPlanetConjunctionNearQueryHalfSpan
	end := queryTT + moonPlanetConjunctionNearQueryHalfSpan
	step := moonPlanetConjunctionNearQueryStepDays
	prevTT := start
	prevVal := moonPlanetConjunctionDeltaAt(prevTT, planet, -1)
	if math.IsNaN(prevVal) {
		return result
	}
	samples := int(math.Ceil((end-start)/step)) + 1
	for i := 1; i <= samples; i++ {
		tt := start + float64(i)*step
		if tt > end {
			tt = end
		}
		val := moonPlanetConjunctionDeltaAt(tt, planet, -1)
		if math.IsNaN(val) {
			return emptyMoonPlanetConjunctionLocalResult()
		}
		if prevVal == 0 || val == 0 || prevVal*val < 0 {
			moonPlanetConjunctionCollectLocalEvent(&result, queryTT, moonPlanetConjunctionEventUT(prevTT, tt, planet))
		}
		if tt == end {
			break
		}
		prevTT = tt
		prevVal = val
	}
	return result
}

func moonPlanetConjunctionMaybeLocalEvents(queryTT float64, planet MoonPlanetConjunctionPlanet) moonPlanetConjunctionLocalResult {
	if !moonPlanetConjunctionShouldCheckLocal(queryTT, planet) {
		return emptyMoonPlanetConjunctionLocalResult()
	}
	return moonPlanetConjunctionLocalEvents(queryTT, planet)
}

func moonPlanetConjunctionGuessTT(queryTT float64, planet MoonPlanetConjunctionPlanet, direction int) float64 {
	delta := moonPlanetConjunctionDeltaAt(queryTT, planet, moonPlanetConjunctionEstimateN)
	if math.IsNaN(delta) {
		return math.NaN()
	}
	period := moonPlanetConjunctionPeriodDays(planet)
	if math.IsNaN(period) {
		return math.NaN()
	}
	switch direction {
	case -1:
		return queryTT - innerLastCycleOffset(delta, period)
	case 1:
		return queryTT + innerNextCycleOffset(delta, period)
	default:
		return math.NaN()
	}
}

func moonPlanetConjunctionDirectionalFallback(queryTT float64, planet MoonPlanetConjunctionPlanet, direction int) float64 {
	period := moonPlanetConjunctionPeriodDays(planet)
	if math.IsNaN(period) {
		return math.NaN()
	}
	span := period * moonPlanetConjunctionFallbackSpanScale
	if span <= 0 {
		return math.NaN()
	}
	step := moonPlanetConjunctionNearQueryStepDays
	start := queryTT
	end := queryTT
	switch direction {
	case -1:
		start -= span
	case 1:
		end += span
	default:
		return math.NaN()
	}

	prevTT := start
	prevVal := moonPlanetConjunctionDeltaAt(prevTT, planet, -1)
	if math.IsNaN(prevVal) {
		return math.NaN()
	}

	bestEventUT := math.NaN()
	for tt := start + step; ; tt += step {
		if tt > end {
			tt = end
		}
		val := moonPlanetConjunctionDeltaAt(tt, planet, -1)
		if math.IsNaN(val) {
			return math.NaN()
		}
		if prevVal == 0 || val == 0 || prevVal*val < 0 {
			eventUT := moonPlanetConjunctionEventUT(prevTT, tt, planet)
			if !math.IsNaN(eventUT) && moonPlanetConjunctionInDirection(eventUT, queryTT, direction) {
				if direction == 1 {
					return eventUT
				}
				bestEventUT = eventUT
			}
		}
		if tt == end {
			break
		}
		prevTT = tt
		prevVal = val
	}
	return bestEventUT
}

func moonPlanetConjunctionDirectionalEventWithLocal(queryTT float64, planet MoonPlanetConjunctionPlanet, direction int, local moonPlanetConjunctionLocalResult) float64 {
	switch direction {
	case -1:
		if !math.IsNaN(local.lastUT) {
			return local.lastUT
		}
	case 1:
		if !math.IsNaN(local.nextUT) {
			return local.nextUT
		}
	}
	guessTT := moonPlanetConjunctionGuessTT(queryTT, planet, direction)
	if math.IsNaN(guessTT) {
		return math.NaN()
	}
	halfSpan := moonPlanetConjunctionBracketHalfSpan
	for attempt := 0; attempt < moonPlanetConjunctionBracketAttempts; attempt++ {
		left, right, ok := moonPlanetConjunctionFindBracket(guessTT, halfSpan, moonPlanetConjunctionBracketStepDays, planet)
		if ok {
			eventUT := moonPlanetConjunctionEventUT(left, right, planet)
			if math.IsNaN(eventUT) {
				halfSpan *= moonPlanetConjunctionBracketGrowth
				continue
			}
			if moonPlanetConjunctionInDirection(eventUT, queryTT, direction) {
				return eventUT
			}
		}
		halfSpan *= moonPlanetConjunctionBracketGrowth
	}
	return moonPlanetConjunctionDirectionalFallback(queryTT, planet, direction)
}

func moonPlanetConjunctionDirectionalEvent(queryTT float64, planet MoonPlanetConjunctionPlanet, direction int) float64 {
	return moonPlanetConjunctionDirectionalEventWithLocal(queryTT, planet, direction, moonPlanetConjunctionMaybeLocalEvents(queryTT, planet))
}

// LastMoonPlanetConjunction 指定时刻之前最近一次行星合月（赤经合） / previous Moon-planet conjunction at or before jde.
func LastMoonPlanetConjunction(jde float64, planet MoonPlanetConjunctionPlanet) float64 {
	return moonPlanetConjunctionDirectionalEvent(jde, planet, -1)
}

// NextMoonPlanetConjunction 指定时刻之后最近一次行星合月（赤经合） / next Moon-planet conjunction at or after jde.
func NextMoonPlanetConjunction(jde float64, planet MoonPlanetConjunctionPlanet) float64 {
	return moonPlanetConjunctionDirectionalEvent(jde, planet, 1)
}

// ClosestMoonPlanetConjunction 离指定时刻最近一次行星合月（赤经合） / closest Moon-planet conjunction to jde.
func ClosestMoonPlanetConjunction(jde float64, planet MoonPlanetConjunctionPlanet) float64 {
	local := moonPlanetConjunctionMaybeLocalEvents(jde, planet)
	if !math.IsNaN(local.lastUT) && !math.IsNaN(local.nextUT) {
		if sameEventJD(local.lastUT, local.nextUT) {
			return local.lastUT
		}
		return closestEventUTToQueryTT(jde, local.lastUT, local.nextUT)
	}
	if !math.IsNaN(local.lastUT) {
		return local.lastUT
	}
	if !math.IsNaN(local.nextUT) {
		return local.nextUT
	}
	last := moonPlanetConjunctionDirectionalEventWithLocal(jde, planet, -1, local)
	next := moonPlanetConjunctionDirectionalEventWithLocal(jde, planet, 1, local)
	if math.IsNaN(last) {
		return next
	}
	if math.IsNaN(next) {
		return last
	}
	return closestEventUTToQueryTT(jde, last, next)
}

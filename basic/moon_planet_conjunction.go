package basic

import (
	"math"

	"github.com/starainrt/astro/internal/geodata"
)

const (
	moonPlanetConjunctionEstimateN         = 8
	moonPlanetConjunctionNearQueryDeltaDeg = 3.0
	moonPlanetConjunctionDirectionEpsilon  = 0.1 / 86400.0
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

func moonPlanetConjunctionDeltaAt(jdTT float64, planet MoonPlanetConjunctionPlanet, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	moonRA := HMoonGeocentricApparentRaNWithInstance(jdTT, n, inst)
	var planetRA float64
	switch planet {
	case MoonPlanetConjunctionMercury:
		planetRA = MercuryApparentRaNWithInstance(jdTT, n, inst)
	case MoonPlanetConjunctionVenus:
		planetRA = VenusApparentRaNWithInstance(jdTT, n, inst)
	case MoonPlanetConjunctionMars:
		planetRA = MarsApparentRaNWithInstance(jdTT, n, inst)
	case MoonPlanetConjunctionJupiter:
		planetRA = JupiterApparentRaNWithInstance(jdTT, n, inst)
	case MoonPlanetConjunctionSaturn:
		planetRA = SaturnApparentRaNWithInstance(jdTT, n, inst)
	case MoonPlanetConjunctionUranus:
		planetRA = UranusApparentRaNWithInstance(jdTT, n, inst)
	case MoonPlanetConjunctionNeptune:
		planetRA = NeptuneApparentRaNWithInstance(jdTT, n, inst)
	default:
		return math.NaN()
	}
	return geodata.NormalizeLongitude(moonRA - planetRA)
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

func moonPlanetConjunctionBeforeOrEqual(eventUT, queryTT float64) bool {
	return eventUTQueryTTDelta(eventUT, queryTT) <= moonPlanetConjunctionDirectionEpsilon
}

func moonPlanetConjunctionAfterOrEqual(eventUT, queryTT float64) bool {
	return eventUTQueryTTDelta(eventUT, queryTT) >= -moonPlanetConjunctionDirectionEpsilon
}

func moonPlanetConjunctionInDirection(eventUT, queryTT float64, direction int) bool {
	switch direction {
	case -1:
		return moonPlanetConjunctionBeforeOrEqual(eventUT, queryTT)
	case 1:
		return moonPlanetConjunctionAfterOrEqual(eventUT, queryTT)
	default:
		return true
	}
}

func moonPlanetConjunctionFindBracket(centerTT, halfSpan, step float64, planet MoonPlanetConjunctionPlanet, instances ...*Instance) (float64, float64, bool) {
	inst := instanceValue(instances)

	if math.IsNaN(centerTT) || math.IsNaN(halfSpan) || math.IsNaN(step) || halfSpan <= 0 || step <= 0 {
		return 0, 0, false
	}
	start := centerTT - halfSpan
	end := centerTT + halfSpan
	samples := int(math.Ceil((end-start)/step)) + 1
	prevTT := start
	prevVal := moonPlanetConjunctionDeltaAt(prevTT, planet, -1, inst)
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
		val := moonPlanetConjunctionDeltaAt(tt, planet, -1, inst)
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

func moonPlanetConjunctionRefineBracket(leftTT, rightTT float64, planet MoonPlanetConjunctionPlanet, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if leftTT > rightTT {
		leftTT, rightTT = rightTT, leftTT
	}
	if leftTT == rightTT {
		return leftTT
	}
	center := (leftTT + rightTT) / 2.0
	halfWindow := (rightTT - leftTT) / 2.0
	return eventZeroRefine(center, halfWindow, moonPlanetConjunctionRefineStepDays, func(sampleTT float64) float64 {
		return moonPlanetConjunctionDeltaAt(sampleTT, planet, -1, inst)
	})
}

func moonPlanetConjunctionEventUT(leftTT, rightTT float64, planet MoonPlanetConjunctionPlanet, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	eventTT := moonPlanetConjunctionRefineBracket(leftTT, rightTT, planet, inst)
	if math.Abs(moonPlanetConjunctionDeltaAt(eventTT, planet, -1, inst)) > moonPlanetConjunctionEventTolerance {
		return math.NaN()
	}
	return TT2UTC(eventTT)
}

func moonPlanetConjunctionCollectLocalEvent(result *moonPlanetConjunctionLocalResult, queryTT, eventUT float64) {
	if math.IsNaN(eventUT) {
		return
	}
	if moonPlanetConjunctionBeforeOrEqual(eventUT, queryTT) {
		if math.IsNaN(result.lastUT) || math.Abs(eventUTQueryTTDelta(eventUT, queryTT)) < math.Abs(eventUTQueryTTDelta(result.lastUT, queryTT)) {
			result.lastUT = eventUT
		}
	}
	if moonPlanetConjunctionAfterOrEqual(eventUT, queryTT) {
		if math.IsNaN(result.nextUT) || math.Abs(eventUTQueryTTDelta(eventUT, queryTT)) < math.Abs(eventUTQueryTTDelta(result.nextUT, queryTT)) {
			result.nextUT = eventUT
		}
	}
}

func moonPlanetConjunctionShouldCheckLocal(queryTT float64, planet MoonPlanetConjunctionPlanet, instances ...*Instance) bool {
	inst := instanceValue(instances)

	delta := moonPlanetConjunctionDeltaAt(queryTT, planet, moonPlanetConjunctionEstimateN, inst)
	if math.IsNaN(delta) {
		return false
	}
	return math.Abs(delta) <= moonPlanetConjunctionNearQueryDeltaDeg
}

func moonPlanetConjunctionLocalEvents(queryTT float64, planet MoonPlanetConjunctionPlanet, instances ...*Instance) moonPlanetConjunctionLocalResult {
	inst := instanceValue(instances)

	result := emptyMoonPlanetConjunctionLocalResult()
	start := queryTT - moonPlanetConjunctionNearQueryHalfSpan
	end := queryTT + moonPlanetConjunctionNearQueryHalfSpan
	step := moonPlanetConjunctionNearQueryStepDays
	prevTT := start
	prevVal := moonPlanetConjunctionDeltaAt(prevTT, planet, -1, inst)
	if math.IsNaN(prevVal) {
		return result
	}
	samples := int(math.Ceil((end-start)/step)) + 1
	for i := 1; i <= samples; i++ {
		tt := start + float64(i)*step
		if tt > end {
			tt = end
		}
		val := moonPlanetConjunctionDeltaAt(tt, planet, -1, inst)
		if math.IsNaN(val) {
			return emptyMoonPlanetConjunctionLocalResult()
		}
		if prevVal == 0 || val == 0 || prevVal*val < 0 {
			moonPlanetConjunctionCollectLocalEvent(&result, queryTT, moonPlanetConjunctionEventUT(prevTT, tt, planet, inst))
		}
		if tt == end {
			break
		}
		prevTT = tt
		prevVal = val
	}
	return result
}

func moonPlanetConjunctionMaybeLocalEvents(queryTT float64, planet MoonPlanetConjunctionPlanet, instances ...*Instance) moonPlanetConjunctionLocalResult {
	inst := instanceValue(instances)

	if !moonPlanetConjunctionShouldCheckLocal(queryTT, planet, inst) {
		return emptyMoonPlanetConjunctionLocalResult()
	}
	return moonPlanetConjunctionLocalEvents(queryTT, planet, inst)
}

func moonPlanetConjunctionGuessTT(queryTT float64, planet MoonPlanetConjunctionPlanet, direction int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	delta := moonPlanetConjunctionDeltaAt(queryTT, planet, moonPlanetConjunctionEstimateN, inst)
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

func moonPlanetConjunctionDirectionalFallback(queryTT float64, planet MoonPlanetConjunctionPlanet, direction int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

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
	prevVal := moonPlanetConjunctionDeltaAt(prevTT, planet, -1, inst)
	if math.IsNaN(prevVal) {
		return math.NaN()
	}

	bestEventUT := math.NaN()
	for tt := start + step; ; tt += step {
		if tt > end {
			tt = end
		}
		val := moonPlanetConjunctionDeltaAt(tt, planet, -1, inst)
		if math.IsNaN(val) {
			return math.NaN()
		}
		if prevVal == 0 || val == 0 || prevVal*val < 0 {
			eventUT := moonPlanetConjunctionEventUT(prevTT, tt, planet, inst)
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

func moonPlanetConjunctionDirectionalEventWithLocal(queryTT float64, planet MoonPlanetConjunctionPlanet, direction int, local moonPlanetConjunctionLocalResult, instances ...*Instance) float64 {
	inst := instanceValue(instances)

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
	guessTT := moonPlanetConjunctionGuessTT(queryTT, planet, direction, inst)
	if math.IsNaN(guessTT) {
		return math.NaN()
	}
	halfSpan := moonPlanetConjunctionBracketHalfSpan
	for attempt := 0; attempt < moonPlanetConjunctionBracketAttempts; attempt++ {
		left, right, ok := moonPlanetConjunctionFindBracket(guessTT, halfSpan, moonPlanetConjunctionBracketStepDays, planet, inst)
		if ok {
			eventUT := moonPlanetConjunctionEventUT(left, right, planet, inst)
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
	return moonPlanetConjunctionDirectionalFallback(queryTT, planet, direction, inst)
}

func moonPlanetConjunctionDirectionalEvent(queryTT float64, planet MoonPlanetConjunctionPlanet, direction int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return moonPlanetConjunctionDirectionalEventWithLocal(queryTT, planet, direction, moonPlanetConjunctionMaybeLocalEvents(queryTT, planet, inst), inst)
}

// LastMoonPlanetConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMoonPlanetConjunctionWithInstance(jde float64, planet MoonPlanetConjunctionPlanet, inst *Instance) float64 {
	return moonPlanetConjunctionDirectionalEvent(jde, planet, -1, inst)
}

// NextMoonPlanetConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMoonPlanetConjunctionWithInstance(jde float64, planet MoonPlanetConjunctionPlanet, inst *Instance) float64 {
	return moonPlanetConjunctionDirectionalEvent(jde, planet, 1, inst)
}

// ClosestMoonPlanetConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func ClosestMoonPlanetConjunctionWithInstance(jde float64, planet MoonPlanetConjunctionPlanet, inst *Instance) float64 {
	local := moonPlanetConjunctionMaybeLocalEvents(jde, planet, inst)
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
	last := moonPlanetConjunctionDirectionalEventWithLocal(jde, planet, -1, local, inst)
	next := moonPlanetConjunctionDirectionalEventWithLocal(jde, planet, 1, local, inst)
	if math.IsNaN(last) {
		return next
	}
	if math.IsNaN(next) {
		return last
	}
	return closestEventUTToQueryTT(jde, last, next)
}

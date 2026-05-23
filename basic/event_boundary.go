package basic

import "math"

const (
	exactEventTolerance     = 2.0 / 86400.0
	exactQueryTTToleranceUT = 0.1 / 86400.0
)

func sameEventJD(a, b float64) bool {
	return math.Abs(a-b) <= exactEventTolerance
}

func sameEventUTQueryTT(eventUT, queryTT float64) bool {
	return math.Abs(eventUTQueryTTDelta(eventUT, queryTT)) <= exactQueryTTToleranceUT
}

func closestEventUTToQueryTT(queryTT, best float64, candidates ...float64) float64 {
	bestAbs := math.Abs(eventUTQueryTTDelta(best, queryTT))
	for _, candidate := range candidates {
		candidateAbs := math.Abs(eventUTQueryTTDelta(candidate, queryTT))
		if candidateAbs < bestAbs {
			best = candidate
			bestAbs = candidateAbs
		}
	}
	return best
}

type phaseEventSearchFunc func(jde, degree float64, next uint8) float64
type simpleEventSearchFunc func(jde float64) float64

func inclusiveLastPhaseEvent(jde, degree float64, fn phaseEventSearchFunc) float64 {
	last := fn(jde, degree, 0)
	next := fn(jde, degree, 1)
	if eventUTQueryBeforeOrEqual(next, jde) && eventUTQueryAfterOrEqual(next, jde) {
		return next
	}
	if eventUTQueryBeforeOrEqual(last, jde) {
		return last
	}
	return last
}

func inclusiveNextPhaseEvent(jde, degree float64, fn phaseEventSearchFunc) float64 {
	last := fn(jde, degree, 0)
	if eventUTQueryBeforeOrEqual(last, jde) && eventUTQueryAfterOrEqual(last, jde) {
		return last
	}
	next := fn(jde, degree, 1)
	if eventUTQueryAfterOrEqual(next, jde) {
		return next
	}
	return next
}

func inclusiveLastSimpleEvent(jde float64, lastFn, nextFn simpleEventSearchFunc) float64 {
	last := lastFn(jde)
	next := nextFn(jde)
	if eventUTQueryBeforeOrEqual(next, jde) && eventUTQueryAfterOrEqual(next, jde) {
		return next
	}
	if eventUTQueryBeforeOrEqual(last, jde) {
		return last
	}
	return last
}

func inclusiveNextSimpleEvent(jde float64, lastFn, nextFn simpleEventSearchFunc) float64 {
	last := lastFn(jde)
	if eventUTQueryBeforeOrEqual(last, jde) && eventUTQueryAfterOrEqual(last, jde) {
		return last
	}
	next := nextFn(jde)
	if eventUTQueryAfterOrEqual(next, jde) {
		return next
	}
	return next
}

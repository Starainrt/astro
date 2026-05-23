package basic

import "math"

const (
	innerEventEpsilon         = 0.1 / 86400.0
	innerEventWindowPadding   = 4.0 / 86400.0
	innerEventMaximizeEpsilon = 4.0 / 86400.0
)

func eventQueryTTAsUT(queryTT float64) float64 {
	return TD2UT(queryTT, false)
}

func eventUTQueryTTDelta(eventUT, queryTT float64) float64 {
	return eventUT - eventQueryTTAsUT(queryTT)
}

func eventUTQueryBeforeOrEqual(eventUT, queryTT float64) bool {
	return eventUTQueryTTDelta(eventUT, queryTT) <= innerEventEpsilon
}

func eventUTQueryAfterOrEqual(eventUT, queryTT float64) bool {
	return eventUTQueryTTDelta(eventUT, queryTT) >= -innerEventEpsilon
}

func eventUTNextQueryTT(eventUT float64) float64 {
	return TD2UT(eventUT, true) + 1.0
}

func eventUTLastQueryTT(eventUT float64) float64 {
	return TD2UT(eventUT, true) - 1.0
}

func innerNextCycleOffset(delta, period float64) float64 {
	if delta <= 0 {
		return -delta * period / 360.0
	}
	return (360.0 - delta) * period / 360.0
}

func innerLastCycleOffset(delta, period float64) float64 {
	if delta >= 0 {
		return delta * period / 360.0
	}
	return (360.0 + delta) * period / 360.0
}

func clampFloat64(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func scanWindowForMinAbs(start, end, step float64, fn func(float64) float64) float64 {
	if end < start {
		start, end = end, start
	}
	if step <= 0 || end == start {
		return start
	}
	bestJD := start
	bestAbs := math.Abs(fn(start))
	for jd := start + step; jd < end; jd += step {
		candidateAbs := math.Abs(fn(jd))
		if candidateAbs < bestAbs {
			bestAbs = candidateAbs
			bestJD = jd
		}
	}
	endAbs := math.Abs(fn(end))
	if endAbs < bestAbs {
		return end
	}
	return bestJD
}

func scanWindowForMax(start, end, step float64, fn func(float64) float64) float64 {
	if end < start {
		start, end = end, start
	}
	if step <= 0 || end == start {
		return start
	}
	bestJD := start
	bestVal := fn(start)
	for jd := start + step; jd < end; jd += step {
		candidateVal := fn(jd)
		if candidateVal > bestVal {
			bestVal = candidateVal
			bestJD = jd
		}
	}
	endVal := fn(end)
	if endVal > bestVal {
		return end
	}
	return bestJD
}

func boundedEventZeroRefine(seed, start, end, halfWindow, step float64, fn func(float64) float64) float64 {
	if end < start {
		start, end = end, start
	}
	if end <= start {
		return start
	}
	maxHalfWindow := (end - start) / 2
	if halfWindow > maxHalfWindow {
		halfWindow = maxHalfWindow
	}
	if halfWindow <= 0 {
		return clampFloat64(seed, start, end)
	}
	seed = clampFloat64(seed, start+halfWindow, end-halfWindow)
	return eventZeroRefine(seed, halfWindow, step, fn)
}

func zeroEventInWindow(start, end, coarseStep, halfWindow, refineStep float64, coarseFn, exactFn func(float64) float64) float64 {
	if end < start {
		start, end = end, start
	}
	if end <= start {
		return start
	}
	rangeDays := end - start
	if coarseStep <= 0 || coarseStep > rangeDays {
		coarseStep = rangeDays / 6.0
	}
	if coarseStep < 0.5 {
		coarseStep = 0.5
	}
	if refineStep <= 0 {
		refineStep = 0.5 / 86400.0
	}
	if halfWindow <= 0 {
		halfWindow = coarseStep
	}
	guess := scanWindowForMinAbs(start, end, coarseStep, coarseFn)
	return boundedEventZeroRefine(guess, start, end, halfWindow, refineStep, exactFn)
}

func maximizeInWindow(start, end, coarseStep float64, coarseFn, exactFn func(float64) float64) float64 {
	if end < start {
		start, end = end, start
	}
	if end <= start {
		return start
	}
	rangeDays := end - start
	if coarseStep <= 0 || coarseStep > rangeDays {
		coarseStep = rangeDays / 6.0
	}
	if coarseStep < 0.5 {
		coarseStep = 0.5
	}
	guess := scanWindowForMax(start, end, coarseStep, coarseFn)
	left := clampFloat64(guess-coarseStep, start, end)
	right := clampFloat64(guess+coarseStep, start, end)
	if right-left <= innerEventMaximizeEpsilon {
		return guess
	}
	for i := 0; i < 20; i++ {
		third := (right - left) / 3.0
		leftThird := left + third
		rightThird := right - third
		if exactFn(leftThird) <= exactFn(rightThird) {
			left = leftThird
			continue
		}
		right = rightThird
	}
	bestJD := guess
	bestVal := exactFn(bestJD)
	for _, jd := range []float64{left, (left + right) / 2.0, right} {
		candidateVal := exactFn(jd)
		if candidateVal > bestVal {
			bestVal = candidateVal
			bestJD = jd
		}
	}
	return bestJD
}

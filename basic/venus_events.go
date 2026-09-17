package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

const (
	VENUS_S_PERIOD    = 1 / ((1 / 224.701) - (1 / 365.256363004))
	venusEventSearchN = 16
	// venusConjunctionSameInstantDegrees 「同刻」快速路径的视黄经差触发阈值（度）。
	// 实测截断级数（n=16）与全项级数的视黄经差最大 0.0031 度：若阈值小于它，
	// 「查询落在合前几秒~几分钟」时截断级数在起点处符号相反，方向扫描会直接跨过这次合，
	// 返回下一个会合周期的合（实测 1950–2055 年有 17~29 个查询点命中）。
	// 0.05 度约合 38 分钟，留有 16 倍余量；侧向判定仍用 0.1 秒口径，语义不变。
	venusConjunctionSameInstantDegrees = 5.0e-2
)

func venusSunLongitudeDelta(jde float64) float64 {
	sub := Limit360(VenusApparentLo(jde) - HSunApparentLo(jde))
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub
}

func venusSunLongitudeDeltaN(jde float64, n int) float64 {
	sub := Limit360(VenusApparentLoN(jde, n) - HSunApparentLoN(jde, n))
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub
}

func venusRADerivative(jde, val float64) float64 {
	sub := VenusApparentRa(jde+val) - VenusApparentRa(jde-val)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * val)
}

func venusRAContinuousForMax(jde float64) float64 {
	ra := VenusApparentRa(jde)
	if ra < 180 {
		return ra + 360
	}
	return ra
}

func venusRAContinuousForMaxN(jde float64, n int) float64 {
	ra := VenusApparentRaN(jde, n)
	if ra < 180 {
		return ra + 360
	}
	return ra
}

func venusRAContinuousForMin(jde float64) float64 {
	ra := VenusApparentRa(jde)
	if ra > 180 {
		return ra - 360
	}
	return ra
}

func venusRAContinuousForMinN(jde float64, n int) float64 {
	ra := VenusApparentRaN(jde, n)
	if ra > 180 {
		return ra - 360
	}
	return ra
}

func venusRAExtremumRefine(seed, start, end, step float64, fn func(float64) float64) float64 {
	centerJD := clampFloat64(seed, start, end)
	halfStep := step
	bestJD := centerJD
	bestVal := fn(centerJD)
	for i := 0; i < 8; i++ {
		leftJD := clampFloat64(centerJD-halfStep, start, end)
		rightJD := clampFloat64(centerJD+halfStep, start, end)
		leftVal := fn(leftJD)
		centerVal := fn(centerJD)
		rightVal := fn(rightJD)
		if leftVal > bestVal {
			bestVal = leftVal
			bestJD = leftJD
		}
		if centerVal > bestVal {
			bestVal = centerVal
			bestJD = centerJD
		}
		if rightVal > bestVal {
			bestVal = rightVal
			bestJD = rightJD
		}
		denominator := leftVal - 2*centerVal + rightVal
		if denominator == 0 {
			centerJD = bestJD
			halfStep /= 2
			continue
		}
		vertexJD := centerJD + 0.5*halfStep*(leftVal-rightVal)/denominator
		vertexJD = clampFloat64(vertexJD, leftJD, rightJD)
		vertexVal := fn(vertexJD)
		if vertexVal > bestVal {
			bestVal = vertexVal
			bestJD = vertexJD
		}
		centerJD = bestJD
		halfStep /= 2
	}
	return bestJD
}

func venusSunElongationN(jde float64, n int) float64 {
	lo1, bo1 := VenusApparentLoBoN(jde, n)
	lo2 := HSunApparentLoN(jde, n)
	bo2 := HSunTrueBoN(jde, n)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

func venusConjunction(jde float64, next uint8) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	queryTT := jde
	direction := -1.0
	if next == 1 {
		direction = 1
	}
	left := queryTT
	leftVal := venusSunLongitudeDeltaN(left, venusEventSearchN)
	if math.Abs(venusSunLongitudeDelta(queryTT)) <= venusConjunctionSameInstantDegrees {
		if exact, ok := venusConjunctionRefine(left, 1.0); ok {
			eventUT := TD2UT(exact, false)
			if next == 0 && eventUTQueryBeforeOrEqual(eventUT, queryTT) {
				return eventUT
			}
			if next == 1 && eventUTQueryAfterOrEqual(eventUT, queryTT) {
				return eventUT
			}
		}
	}
	const step = 8.0
	for i := 0; i < 80; i++ {
		right := queryTT + direction*step*float64(i+1)
		rightVal := venusSunLongitudeDeltaN(right, venusEventSearchN)
		if leftVal == 0 || rightVal == 0 || leftVal*rightVal <= 0 {
			center := (left + right) / 2.0
			halfWindow := math.Abs(right-left) / 2.0
			if exact, ok := venusConjunctionRefine(center, halfWindow); ok {
				return TD2UT(exact, false)
			}
			// 截断级数在根附近的符号可能与全项不一致（查询几乎正好落在合上），
			// 此时改用全项级数做方向扫描兜底，而不是直接有界失败。
			return venusConjunctionFullDirectionalScan(queryTT, direction)
		}
		left = right
		leftVal = rightVal
	}
	// 640 天已经覆盖一个金星会合周期；仍无括号通常表示输入超出解析项的可靠范围。
	// 继续按 5 微日扫描整个周期会产生数亿次星历计算，因此在这里有界失败。
	// The 640-day directional scan already exceeds one Venus synodic period. If it
	// still finds no bracket, fail in a bounded way instead of scanning hundreds
	// of millions of five-microday samples across the full fallback window.
	return math.NaN()
}

// venusConjunctionFullDirectionalScan 全项级数的方向扫描（截断括号未被确认时的兜底）。
func venusConjunctionFullDirectionalScan(queryTT, direction float64) float64 {
	const (
		step     = 8.0
		maxSteps = 80
	)
	left := queryTT
	leftVal := venusSunLongitudeDelta(left)
	if !isFiniteFloat(leftVal) {
		return math.NaN()
	}
	for i := 0; i < maxSteps; i++ {
		right := queryTT + direction*step*float64(i+1)
		rightVal := venusSunLongitudeDelta(right)
		if !isFiniteFloat(rightVal) {
			return math.NaN()
		}
		if leftVal == 0 || rightVal == 0 || leftVal*rightVal < 0 {
			center := (left + right) / 2.0
			if exact, ok := venusConjunctionRefine(center, math.Abs(right-left)/2.0); ok {
				return TD2UT(exact, false)
			}
			return math.NaN()
		}
		left = right
		leftVal = rightVal
	}
	return math.NaN()
}

func venusConjunctionRefine(seed, halfWindow float64) (float64, bool) {
	leftJD := seed - halfWindow
	centerJD := seed
	rightJD := seed + halfWindow
	leftVal := venusSunLongitudeDelta(leftJD)
	centerVal := venusSunLongitudeDelta(centerJD)
	rightVal := venusSunLongitudeDelta(rightJD)
	if !isFiniteFloat(leftVal) || !isFiniteFloat(centerVal) || !isFiniteFloat(rightVal) {
		return math.NaN(), false
	}
	if _, _, _, _, ok := eventZeroBracket(leftJD, leftVal, centerJD, centerVal, rightJD, rightVal); !ok {
		return math.NaN(), false
	}
	return eventZeroRefine(seed, halfWindow, 0.000005, venusSunLongitudeDelta), true
}

func venusConjunctionTypeAt(eventUT float64) bool {
	return EarthVenusAway(eventUT) <= EarthAway(eventUT)
}

func nextVenusTypedConjunctionFromEvent(jde float64, inferior bool) float64 {
	date := NextVenusConjunctionStrict(jde)
	if venusConjunctionTypeAt(date) == inferior {
		return date
	}
	return NextVenusConjunctionStrict(eventUTNextQueryTT(date))
}

func lastVenusTypedConjunctionFromEvent(jde float64, inferior bool) float64 {
	date := LastVenusConjunctionStrict(jde)
	if venusConjunctionTypeAt(date) == inferior {
		return date
	}
	return LastVenusConjunctionStrict(eventUTLastQueryTT(date))
}

func LastVenusConjunction(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusConjunctionStrict, NextVenusConjunctionStrict)
}

func NextVenusConjunction(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusConjunctionStrict, NextVenusConjunctionStrict)
}

func LastVenusConjunctionStrict(jde float64) float64 {
	return venusConjunction(jde, 0)
}

func NextVenusConjunctionStrict(jde float64) float64 {
	return venusConjunction(jde, 1)
}

func nextVenusTypedConjunction(jde float64, inferior bool) float64 {
	return nextVenusTypedConjunctionFromEvent(jde, inferior)
}

func lastVenusTypedConjunction(jde float64, inferior bool) float64 {
	return lastVenusTypedConjunctionFromEvent(jde, inferior)
}

func NextVenusInferiorConjunction(jde float64) float64 {
	return nextVenusTypedConjunction(jde, true)
}

func NextVenusSuperiorConjunction(jde float64) float64 {
	return nextVenusTypedConjunction(jde, false)
}

func LastVenusInferiorConjunction(jde float64) float64 {
	return lastVenusTypedConjunction(jde, true)
}

func LastVenusSuperiorConjunction(jde float64) float64 {
	return lastVenusTypedConjunction(jde, false)
}

func NextVenusRetrograde(jde float64) float64 {
	p2r := NextVenusProgradeToRetrograde(jde)
	r2p := NextVenusRetrogradeToPrograde(jde)
	if sameEventJD(p2r, r2p) {
		return p2r
	}
	return earliestFiniteEventUT(p2r, r2p)
}

func LastVenusRetrograde(jde float64) float64 {
	p2r := LastVenusProgradeToRetrograde(jde)
	r2p := LastVenusRetrogradeToPrograde(jde)
	if sameEventJD(p2r, r2p) {
		return p2r
	}
	return latestFiniteEventUT(p2r, r2p)
}

func venusStationInWindow(start, end float64, progradeToRetrograde bool) float64 {
	var best float64
	if progradeToRetrograde {
		guess := scanWindowForMax(start, end, 2.0, func(jd float64) float64 {
			return venusRAContinuousForMaxN(jd, venusEventSearchN)
		})
		best = venusRAExtremumRefine(guess, start, end, 1.0, func(jd float64) float64 {
			return venusRAContinuousForMax(jd)
		})
	} else {
		guess := scanWindowForMax(start, end, 2.0, func(jd float64) float64 {
			return -venusRAContinuousForMinN(jd, venusEventSearchN)
		})
		best = venusRAExtremumRefine(guess, start, end, 1.0, func(jd float64) float64 {
			return -venusRAContinuousForMin(jd)
		})
	}
	// 抛物顶点法 8 次迭代后仍有秒级残差（实测最差 2.4 s），会让「查询落在站前 1 秒」
	// 这类边界查询跳到下一个会合周期。这里再用 RA 变化率的零点做一次割线抛光，
	// 把站时刻精度压到亚秒级；抛光失败（括号内无异号）时保留顶点法结果。
	if polished, ok := venusStationPolish(best, progradeToRetrograde); ok {
		best = polished
	}
	return TD2UT(best, false)
}

// venusStationPolish 在抛物顶点结果附近求 RA 变化率的零点。
// progradeToRetrograde=true 时 RA 变化率由正转负（极大值），否则由负转正（极小值）。
func venusStationPolish(seed float64, progradeToRetrograde bool) (float64, bool) {
	rate := func(jd float64) float64 { return venusRADerivative(jd, stationDerivativeStepDay) }
	for _, halfWindow := range []float64{120.0 / 86400.0, 600.0 / 86400.0} {
		leftJD := seed - halfWindow
		rightJD := seed + halfWindow
		leftValue := rate(leftJD)
		rightValue := rate(rightJD)
		if !isFiniteFloat(leftValue) || !isFiniteFloat(rightValue) {
			return math.NaN(), false
		}
		if leftValue == 0 {
			return leftJD, true
		}
		if rightValue == 0 {
			return rightJD, true
		}
		if leftValue*rightValue > 0 {
			continue
		}
		if progradeToRetrograde != (leftValue > 0) {
			return math.NaN(), false
		}
		root, ok := eventBracketSecantRoot(leftJD, rightJD, leftValue, rightValue, 0.01/86400.0, rate)
		if !ok {
			return math.NaN(), false
		}
		return root, true
	}
	return math.NaN(), false
}

func venusProgradeToRetrogradeAroundInferior(inferior float64) float64 {
	return venusStationInWindow(inferior-30.0, inferior-14.0, true)
}

func venusRetrogradeToProgradeAroundInferior(inferior float64) float64 {
	return venusStationInWindow(inferior+14.0, inferior+24.0, false)
}

func NextVenusProgradeToRetrograde(jde float64) float64 {
	inferior := NextVenusInferiorConjunction(jde)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusProgradeToRetrogradeAroundInferior(inferior)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if stationUTQueryAfterOrEqual(date, jde) {
			return date
		}
		inferior = NextVenusInferiorConjunction(eventUTNextQueryTT(inferior))
	}
	return math.NaN()
}

func NextVenusRetrogradeToPrograde(jde float64) float64 {
	inferior := LastVenusInferiorConjunction(jde)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusRetrogradeToProgradeAroundInferior(inferior)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if stationUTQueryAfterOrEqual(date, jde) {
			return date
		}
		inferior = NextVenusInferiorConjunction(eventUTNextQueryTT(inferior))
	}
	return math.NaN()
}

func LastVenusProgradeToRetrograde(jde float64) float64 {
	inferior := NextVenusInferiorConjunction(jde)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusProgradeToRetrogradeAroundInferior(inferior)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if stationUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		inferior = LastVenusInferiorConjunction(eventUTLastQueryTT(inferior))
	}
	return math.NaN()
}

func LastVenusRetrogradeToPrograde(jde float64) float64 {
	inferior := LastVenusInferiorConjunction(jde)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusRetrogradeToProgradeAroundInferior(inferior)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if stationUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		inferior = LastVenusInferiorConjunction(eventUTLastQueryTT(inferior))
	}
	return math.NaN()
}

func VenusSunElongation(jde float64) float64 {
	lo1, bo1 := VenusApparentLoBo(jde)
	lo2 := HSunApparentLo(jde)
	bo2 := HSunTrueBo(jde)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

// venusGreatestElongationInWindow 求窗口内大距：目标是公开 VenusSunElongation 的极大（视距角），
// 而不是忽略光行差/视位置修正的真距角，否则返回的时刻不是调用方能量到的那个极值。
// 窗口两端是世界时，目标函数收力学时，因此逐次换算。
func venusGreatestElongationInWindow(start, end float64) float64 {
	return maximizeInWindow(start, end, 5.0, func(utJD float64) float64 {
		return venusSunElongationN(TD2UT(utJD, true), venusEventSearchN)
	}, func(utJD float64) float64 {
		return VenusSunElongation(TD2UT(utJD, true))
	})
}

func venusEastElongationWindowEndingAt(inferior float64) (float64, float64) {
	lastSuperior := LastVenusSuperiorConjunction(eventUTLastQueryTT(inferior))
	return lastSuperior + innerEventWindowPadding, inferior - innerEventWindowPadding
}

func venusWestElongationWindowEndingAt(superior float64) (float64, float64) {
	lastInferior := LastVenusInferiorConjunction(eventUTLastQueryTT(superior))
	return lastInferior + innerEventWindowPadding, superior - innerEventWindowPadding
}

func venusEastElongationWindowContaining(jde float64) (float64, float64) {
	nextInferior := NextVenusInferiorConjunction(jde)
	start, end := venusEastElongationWindowEndingAt(nextInferior)
	if eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde) {
		return start, end
	}
	currentInferior := LastVenusInferiorConjunction(jde)
	return venusEastElongationWindowEndingAt(currentInferior)
}

func venusWestElongationWindowContaining(jde float64) (float64, float64) {
	nextSuperior := NextVenusSuperiorConjunction(jde)
	start, end := venusWestElongationWindowEndingAt(nextSuperior)
	if eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde) {
		return start, end
	}
	currentSuperior := LastVenusSuperiorConjunction(jde)
	return venusWestElongationWindowEndingAt(currentSuperior)
}

func nextVenusGreatestElongationTyped(jde float64, east bool) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if east {
		start, windowEnd := venusEastElongationWindowContaining(jde)
		for i := 0; i < eventDirectionalSearchIterations; i++ {
			date := venusGreatestElongationInWindow(start, windowEnd)
			if !isFiniteFloat(date) {
				return math.NaN()
			}
			if eventUTQueryAfterOrEqual(date, jde) {
				return date
			}
			nextInferior := NextVenusInferiorConjunction(eventUTNextQueryTT(windowEnd))
			start, windowEnd = venusEastElongationWindowEndingAt(nextInferior)
		}
		return math.NaN()
	}
	start, windowEnd := venusWestElongationWindowContaining(jde)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusGreatestElongationInWindow(start, windowEnd)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		nextSuperior := NextVenusSuperiorConjunction(eventUTNextQueryTT(windowEnd))
		start, windowEnd = venusWestElongationWindowEndingAt(nextSuperior)
	}
	return math.NaN()
}

func lastVenusGreatestElongationTyped(jde float64, east bool) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if east {
		start, windowEnd := venusEastElongationWindowContaining(jde)
		for i := 0; i < eventDirectionalSearchIterations; i++ {
			date := venusGreatestElongationInWindow(start, windowEnd)
			if !isFiniteFloat(date) {
				return math.NaN()
			}
			if eventUTQueryBeforeOrEqual(date, jde) {
				return date
			}
			prevInferior := LastVenusInferiorConjunction(eventUTLastQueryTT(start))
			start, windowEnd = venusEastElongationWindowEndingAt(prevInferior)
		}
		return math.NaN()
	}
	start, windowEnd := venusWestElongationWindowContaining(jde)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusGreatestElongationInWindow(start, windowEnd)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		prevSuperior := LastVenusSuperiorConjunction(eventUTLastQueryTT(start))
		start, windowEnd = venusWestElongationWindowEndingAt(prevSuperior)
	}
	return math.NaN()
}

// venusElongationWindowAt 返回包含 jde 的该侧大距窗口；查询落在合的 4 秒内边距里时两侧都不包含。
func venusElongationWindowAt(jde float64, east bool) (float64, float64, bool) {
	if east {
		start, end := venusEastElongationWindowEndingAt(NextVenusInferiorConjunction(jde))
		return start, end, eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde)
	}
	start, end := venusWestElongationWindowEndingAt(NextVenusSuperiorConjunction(jde))
	return start, end, eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde)
}

// 无东西侧参数的 Next/Last 先只看查询所在窗口那一侧：该侧大距未过就是答案，已过则另一侧紧接着的
// 下一个才是答案，因此只有在查询贴住合（两侧窗口都不含）时才退化为两侧都算。赤经差在合附近会提前
// 变号，不能用它定窗口。
func NextVenusGreatestElongation(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	for _, east := range [2]bool{true, false} {
		start, end, ok := venusElongationWindowAt(jde, east)
		if !ok {
			continue
		}
		if date := venusGreatestElongationInWindow(start, end); isFiniteFloat(date) && eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		if date := nextVenusGreatestElongationTyped(jde, !east); isFiniteFloat(date) && eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		break
	}
	return earliestFiniteEventUT(nextVenusGreatestElongationTyped(jde, true), nextVenusGreatestElongationTyped(jde, false))
}

func LastVenusGreatestElongation(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	for _, east := range [2]bool{true, false} {
		start, end, ok := venusElongationWindowAt(jde, east)
		if !ok {
			continue
		}
		if date := venusGreatestElongationInWindow(start, end); isFiniteFloat(date) && eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		if date := lastVenusGreatestElongationTyped(jde, !east); isFiniteFloat(date) && eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		break
	}
	return latestFiniteEventUT(lastVenusGreatestElongationTyped(jde, true), lastVenusGreatestElongationTyped(jde, false))
}

func LastVenusInferiorConjunctionInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusInferiorConjunction, NextVenusInferiorConjunction)
}

func NextVenusInferiorConjunctionInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusInferiorConjunction, NextVenusInferiorConjunction)
}

func LastVenusSuperiorConjunctionInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusSuperiorConjunction, NextVenusSuperiorConjunction)
}

func NextVenusSuperiorConjunctionInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusSuperiorConjunction, NextVenusSuperiorConjunction)
}

func LastVenusRetrogradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusRetrograde, NextVenusRetrograde)
}

func NextVenusRetrogradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusRetrograde, NextVenusRetrograde)
}

func LastVenusProgradeToRetrogradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusProgradeToRetrograde, NextVenusProgradeToRetrograde)
}

func NextVenusProgradeToRetrogradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusProgradeToRetrograde, NextVenusProgradeToRetrograde)
}

func LastVenusRetrogradeToProgradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusRetrogradeToPrograde, NextVenusRetrogradeToPrograde)
}

func NextVenusRetrogradeToProgradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusRetrogradeToPrograde, NextVenusRetrogradeToPrograde)
}

func LastVenusGreatestElongationInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusGreatestElongation, NextVenusGreatestElongation)
}

func NextVenusGreatestElongationInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusGreatestElongation, NextVenusGreatestElongation)
}

func LastVenusGreatestElongationEastInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusGreatestElongationEast, NextVenusGreatestElongationEast)
}

func NextVenusGreatestElongationEastInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusGreatestElongationEast, NextVenusGreatestElongationEast)
}

func LastVenusGreatestElongationWestInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastVenusGreatestElongationWest, NextVenusGreatestElongationWest)
}

func NextVenusGreatestElongationWestInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastVenusGreatestElongationWest, NextVenusGreatestElongationWest)
}

func NextVenusGreatestElongationEast(jde float64) float64 {
	return nextVenusGreatestElongationTyped(jde, true)
}

func NextVenusGreatestElongationWest(jde float64) float64 {
	return nextVenusGreatestElongationTyped(jde, false)
}

func LastVenusGreatestElongationEast(jde float64) float64 {
	return lastVenusGreatestElongationTyped(jde, true)
}

func LastVenusGreatestElongationWest(jde float64) float64 {
	return lastVenusGreatestElongationTyped(jde, false)
}

package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

const (
	VENUS_S_PERIOD    = 1 / ((1 / 224.701) - (1 / 365.256363004))
	venusEventSearchN = 16
	// venusConjunctionSameInstantDegrees 「同刻」快速路径的视黄经差触发阈值（度）：必须明显大于
	// 截断级数与全项级数的最大视黄经差，否则「查询落在合前几秒」时起点符号相反，方向扫描会跨过这次合。
	// 0.05 度约合 38 分钟；侧向判定仍用 0.1 秒口径，语义不变。
	venusConjunctionSameInstantDegrees = 5.0e-2
)

func venusSunLongitudeDelta(jde float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(VenusApparentLoWithInstance(jde, inst) - HSunApparentLoWithInstance(jde, inst))
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub
}

func venusSunLongitudeDeltaN(jde float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(VenusApparentLoNWithInstance(jde, n, inst) - HSunApparentLoNWithInstance(jde, n, inst))
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub
}

func venusRADerivative(jde, val float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := VenusApparentRaWithInstance(jde+val, inst) - VenusApparentRaWithInstance(jde-val, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * val)
}

func venusRAContinuousForMax(jde float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	ra := VenusApparentRaWithInstance(jde, inst)
	if ra < 180 {
		return ra + 360
	}
	return ra
}

func venusRAContinuousForMaxN(jde float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	ra := VenusApparentRaNWithInstance(jde, n, inst)
	if ra < 180 {
		return ra + 360
	}
	return ra
}

func venusRAContinuousForMin(jde float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	ra := VenusApparentRaWithInstance(jde, inst)
	if ra > 180 {
		return ra - 360
	}
	return ra
}

func venusRAContinuousForMinN(jde float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	ra := VenusApparentRaNWithInstance(jde, n, inst)
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

func venusSunElongationN(jde float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	lo1, bo1 := VenusApparentLoBoNWithInstance(jde, n, inst)
	lo2 := HSunApparentLoNWithInstance(jde, n, inst)
	bo2 := HSunTrueBoNWithInstance(jde, n, inst)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

func venusConjunction(jde float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	queryTT := jde
	direction := -1.0
	if next == 1 {
		direction = 1
	}
	left := queryTT
	leftVal := venusSunLongitudeDeltaN(left, venusEventSearchN, inst)
	if math.Abs(venusSunLongitudeDelta(queryTT, inst)) <= venusConjunctionSameInstantDegrees {
		if exact, ok := venusConjunctionRefine(left, 1.0, inst); ok {
			eventUT := TT2UTC(exact)
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
		rightVal := venusSunLongitudeDeltaN(right, venusEventSearchN, inst)
		if leftVal == 0 || rightVal == 0 || leftVal*rightVal <= 0 {
			center := (left + right) / 2.0
			halfWindow := math.Abs(right-left) / 2.0
			if exact, ok := venusConjunctionRefine(center, halfWindow, inst); ok {
				return TT2UTC(exact)
			}
			// 截断级数在根附近的符号可能与全项不一致（查询几乎正好落在合上），
			// 此时改用全项级数做方向扫描兜底，而不是直接有界失败。
			return venusConjunctionFullDirectionalScan(queryTT, direction, inst)
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
func venusConjunctionFullDirectionalScan(queryTT, direction float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	const (
		step     = 8.0
		maxSteps = 80
	)
	left := queryTT
	leftVal := venusSunLongitudeDelta(left, inst)
	if !isFiniteFloat(leftVal) {
		return math.NaN()
	}
	for i := 0; i < maxSteps; i++ {
		right := queryTT + direction*step*float64(i+1)
		rightVal := venusSunLongitudeDelta(right, inst)
		if !isFiniteFloat(rightVal) {
			return math.NaN()
		}
		if leftVal == 0 || rightVal == 0 || leftVal*rightVal < 0 {
			center := (left + right) / 2.0
			if exact, ok := venusConjunctionRefine(center, math.Abs(right-left)/2.0, inst); ok {
				return TT2UTC(exact)
			}
			return math.NaN()
		}
		left = right
		leftVal = rightVal
	}
	return math.NaN()
}

func venusConjunctionRefine(seed, halfWindow float64, instances ...*Instance) (float64, bool) {
	inst := instanceValue(instances)

	leftJDE := seed - halfWindow
	centerJDE := seed
	rightJDE := seed + halfWindow
	leftVal := venusSunLongitudeDelta(leftJDE, inst)
	centerVal := venusSunLongitudeDelta(centerJDE, inst)
	rightVal := venusSunLongitudeDelta(rightJDE, inst)
	if !isFiniteFloat(leftVal) || !isFiniteFloat(centerVal) || !isFiniteFloat(rightVal) {
		return math.NaN(), false
	}
	if _, _, _, _, ok := eventZeroBracket(leftJDE, leftVal, centerJDE, centerVal, rightJDE, rightVal); !ok {
		return math.NaN(), false
	}
	return eventZeroRefine(seed, halfWindow, 0.000005, func(jde float64) float64 { return venusSunLongitudeDelta(jde, inst) }), true
}

func venusConjunctionTypeAt(eventUT float64, instances ...*Instance) bool {
	inst := instanceValue(instances)

	return EarthVenusAwayWithInstance(eventUT, inst) <= EarthAwayWithInstance(eventUT, inst)
}

func nextVenusTypedConjunctionFromEvent(jde float64, inferior bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	date := NextVenusConjunctionStrictWithInstance(jde, inst)
	if venusConjunctionTypeAt(date, inst) == inferior {
		return date
	}
	return NextVenusConjunctionStrictWithInstance(eventUTNextQueryTT(date), inst)
}

func lastVenusTypedConjunctionFromEvent(jde float64, inferior bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	date := LastVenusConjunctionStrictWithInstance(jde, inst)
	if venusConjunctionTypeAt(date, inst) == inferior {
		return date
	}
	return LastVenusConjunctionStrictWithInstance(eventUTLastQueryTT(date), inst)
}

// LastVenusConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastVenusConjunctionStrictWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusConjunctionStrictWithInstance(jde, inst) })
}

// NextVenusConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastVenusConjunctionStrictWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusConjunctionStrictWithInstance(jde, inst) })
}

// LastVenusConjunctionStrictWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusConjunctionStrictWithInstance(jde float64, inst *Instance) float64 {
	return venusConjunction(jde, 0, inst)
}

// NextVenusConjunctionStrictWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusConjunctionStrictWithInstance(jde float64, inst *Instance) float64 {
	return venusConjunction(jde, 1, inst)
}

func nextVenusTypedConjunction(jde float64, inferior bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return nextVenusTypedConjunctionFromEvent(jde, inferior, inst)
}

func lastVenusTypedConjunction(jde float64, inferior bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return lastVenusTypedConjunctionFromEvent(jde, inferior, inst)
}

// NextVenusInferiorConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusInferiorConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return nextVenusTypedConjunction(jde, true, inst)
}

// NextVenusSuperiorConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusSuperiorConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return nextVenusTypedConjunction(jde, false, inst)
}

// LastVenusInferiorConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusInferiorConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return lastVenusTypedConjunction(jde, true, inst)
}

// LastVenusSuperiorConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusSuperiorConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return lastVenusTypedConjunction(jde, false, inst)
}

// NextVenusRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	p2r := NextVenusProgradeToRetrogradeWithInstance(jde, inst)
	r2p := NextVenusRetrogradeToProgradeWithInstance(jde, inst)
	if sameEventJD(p2r, r2p) {
		return p2r
	}
	return earliestFiniteEventUT(p2r, r2p)
}

// LastVenusRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	p2r := LastVenusProgradeToRetrogradeWithInstance(jde, inst)
	r2p := LastVenusRetrogradeToProgradeWithInstance(jde, inst)
	if sameEventJD(p2r, r2p) {
		return p2r
	}
	return latestFiniteEventUT(p2r, r2p)
}

func venusStationInWindow(start, end float64, progradeToRetrograde bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	var best float64
	if progradeToRetrograde {
		guess := scanWindowForMax(start, end, 2.0, func(jd float64) float64 {
			return venusRAContinuousForMaxN(jd, venusEventSearchN, inst)
		})
		best = venusRAExtremumRefine(guess, start, end, 1.0, func(jd float64) float64 {
			return venusRAContinuousForMax(jd, inst)
		})
	} else {
		guess := scanWindowForMax(start, end, 2.0, func(jd float64) float64 {
			return -venusRAContinuousForMinN(jd, venusEventSearchN, inst)
		})
		best = venusRAExtremumRefine(guess, start, end, 1.0, func(jd float64) float64 {
			return -venusRAContinuousForMin(jd, inst)
		})
	}
	// 抛物顶点法仍有秒级残差，会让「查询落在站前 1 秒」这类边界查询跳到下一个会合周期；
	// 这里再用 RA 变化率的零点做一次割线抛光，把站时刻精度压到亚秒级，
	// 抛光失败（括号内无异号）时保留顶点法结果。
	if polished, ok := venusStationPolish(best, progradeToRetrograde, inst); ok {
		best = polished
	}
	return TT2UTC(best)
}

// venusStationPolish 在抛物顶点结果附近求 RA 变化率的零点。
// progradeToRetrograde=true 时 RA 变化率由正转负（极大值），否则由负转正（极小值）。
func venusStationPolish(seed float64, progradeToRetrograde bool, instances ...*Instance) (float64, bool) {
	inst := instanceValue(instances)

	rate := func(jd float64) float64 { return venusRADerivative(jd, stationDerivativeStepDay, inst) }
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

func venusProgradeToRetrogradeAroundInferior(inferior float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return venusStationInWindow(inferior-30.0, inferior-14.0, true, inst)
}

func venusRetrogradeToProgradeAroundInferior(inferior float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return venusStationInWindow(inferior+14.0, inferior+24.0, false, inst)
}

// NextVenusProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	inferior := NextVenusInferiorConjunctionWithInstance(jde, inst)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusProgradeToRetrogradeAroundInferior(inferior, inst)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if stationUTQueryAfterOrEqual(date, jde) {
			return date
		}
		inferior = NextVenusInferiorConjunctionWithInstance(eventUTNextQueryTT(inferior), inst)
	}
	return math.NaN()
}

// NextVenusRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	inferior := LastVenusInferiorConjunctionWithInstance(jde, inst)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusRetrogradeToProgradeAroundInferior(inferior, inst)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if stationUTQueryAfterOrEqual(date, jde) {
			return date
		}
		inferior = NextVenusInferiorConjunctionWithInstance(eventUTNextQueryTT(inferior), inst)
	}
	return math.NaN()
}

// LastVenusProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	inferior := NextVenusInferiorConjunctionWithInstance(jde, inst)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusProgradeToRetrogradeAroundInferior(inferior, inst)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if stationUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		inferior = LastVenusInferiorConjunctionWithInstance(eventUTLastQueryTT(inferior), inst)
	}
	return math.NaN()
}

// LastVenusRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	inferior := LastVenusInferiorConjunctionWithInstance(jde, inst)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusRetrogradeToProgradeAroundInferior(inferior, inst)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if stationUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		inferior = LastVenusInferiorConjunctionWithInstance(eventUTLastQueryTT(inferior), inst)
	}
	return math.NaN()
}

// VenusSunElongationWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusSunElongationWithInstance(jde float64, inst *Instance) float64 {
	lo1, bo1 := VenusApparentLoBoWithInstance(jde, inst)
	lo2 := HSunApparentLoWithInstance(jde, inst)
	bo2 := HSunTrueBoWithInstance(jde, inst)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

// venusGreatestElongationInWindow 求窗口内大距：目标是公开 VenusSunElongation 的极大（视距角），
// 而不是忽略光行差/视位置修正的真距角，否则返回的时刻不是调用方能量到的那个极值。
// 窗口两端是世界时，目标函数收力学时，因此逐次换算。
func venusGreatestElongationInWindow(start, end float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return maximizeInWindow(start, end, 5.0, func(utJD float64) float64 {
		return venusSunElongationN(UTC2TT(utJD), venusEventSearchN, inst)
	}, func(utJD float64) float64 {
		return VenusSunElongationWithInstance(UTC2TT(utJD), inst)
	})
}

func venusEastElongationWindowEndingAt(inferior float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	lastSuperior := LastVenusSuperiorConjunctionWithInstance(eventUTLastQueryTT(inferior), inst)
	return lastSuperior + innerEventWindowPadding, inferior - innerEventWindowPadding
}

func venusWestElongationWindowEndingAt(superior float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	lastInferior := LastVenusInferiorConjunctionWithInstance(eventUTLastQueryTT(superior), inst)
	return lastInferior + innerEventWindowPadding, superior - innerEventWindowPadding
}

func venusEastElongationWindowContaining(jde float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	nextInferior := NextVenusInferiorConjunctionWithInstance(jde, inst)
	start, end := venusEastElongationWindowEndingAt(nextInferior, inst)
	if eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde) {
		return start, end
	}
	currentInferior := LastVenusInferiorConjunctionWithInstance(jde, inst)
	return venusEastElongationWindowEndingAt(currentInferior, inst)
}

func venusWestElongationWindowContaining(jde float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	nextSuperior := NextVenusSuperiorConjunctionWithInstance(jde, inst)
	start, end := venusWestElongationWindowEndingAt(nextSuperior, inst)
	if eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde) {
		return start, end
	}
	currentSuperior := LastVenusSuperiorConjunctionWithInstance(jde, inst)
	return venusWestElongationWindowEndingAt(currentSuperior, inst)
}

func nextVenusGreatestElongationTyped(jde float64, east bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if east {
		start, windowEnd := venusEastElongationWindowContaining(jde, inst)
		for i := 0; i < eventDirectionalSearchIterations; i++ {
			date := venusGreatestElongationInWindow(start, windowEnd, inst)
			if !isFiniteFloat(date) {
				return math.NaN()
			}
			if eventUTQueryAfterOrEqual(date, jde) {
				return date
			}
			nextInferior := NextVenusInferiorConjunctionWithInstance(eventUTNextQueryTT(windowEnd), inst)
			start, windowEnd = venusEastElongationWindowEndingAt(nextInferior, inst)
		}
		return math.NaN()
	}
	start, windowEnd := venusWestElongationWindowContaining(jde, inst)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusGreatestElongationInWindow(start, windowEnd, inst)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		nextSuperior := NextVenusSuperiorConjunctionWithInstance(eventUTNextQueryTT(windowEnd), inst)
		start, windowEnd = venusWestElongationWindowEndingAt(nextSuperior, inst)
	}
	return math.NaN()
}

func lastVenusGreatestElongationTyped(jde float64, east bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if east {
		start, windowEnd := venusEastElongationWindowContaining(jde, inst)
		for i := 0; i < eventDirectionalSearchIterations; i++ {
			date := venusGreatestElongationInWindow(start, windowEnd, inst)
			if !isFiniteFloat(date) {
				return math.NaN()
			}
			if eventUTQueryBeforeOrEqual(date, jde) {
				return date
			}
			prevInferior := LastVenusInferiorConjunctionWithInstance(eventUTLastQueryTT(start), inst)
			start, windowEnd = venusEastElongationWindowEndingAt(prevInferior, inst)
		}
		return math.NaN()
	}
	start, windowEnd := venusWestElongationWindowContaining(jde, inst)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := venusGreatestElongationInWindow(start, windowEnd, inst)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		prevSuperior := LastVenusSuperiorConjunctionWithInstance(eventUTLastQueryTT(start), inst)
		start, windowEnd = venusWestElongationWindowEndingAt(prevSuperior, inst)
	}
	return math.NaN()
}

// venusElongationWindowAt 返回包含 jde 的该侧大距窗口；查询落在合的 4 秒内边距里时两侧都不包含。
func venusElongationWindowAt(jde float64, east bool, instances ...*Instance) (float64, float64, bool) {
	inst := instanceValue(instances)

	if east {
		start, end := venusEastElongationWindowEndingAt(NextVenusInferiorConjunctionWithInstance(jde, inst), inst)
		return start, end, eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde)
	}
	start, end := venusWestElongationWindowEndingAt(NextVenusSuperiorConjunctionWithInstance(jde, inst), inst)
	return start, end, eventUTQueryBeforeOrEqual(start, jde) && eventUTQueryAfterOrEqual(end, jde)
}

// NextVenusGreatestElongationWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusGreatestElongationWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	// 先只看查询所在一侧的窗口，未过即答案，已过则取另一侧的下一个；赤经差在合附近会提前变号，不能用来定窗口。
	for _, east := range [2]bool{true, false} {
		start, end, ok := venusElongationWindowAt(jde, east, inst)
		if !ok {
			continue
		}
		if date := venusGreatestElongationInWindow(start, end, inst); isFiniteFloat(date) && eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		if date := nextVenusGreatestElongationTyped(jde, !east, inst); isFiniteFloat(date) && eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		break
	}
	return earliestFiniteEventUT(nextVenusGreatestElongationTyped(jde, true, inst), nextVenusGreatestElongationTyped(jde, false, inst))
}

// LastVenusGreatestElongationWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusGreatestElongationWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	// 先只看查询所在一侧的窗口，未过即答案，已过则取另一侧的下一个；赤经差在合附近会提前变号，不能用来定窗口。
	for _, east := range [2]bool{true, false} {
		start, end, ok := venusElongationWindowAt(jde, east, inst)
		if !ok {
			continue
		}
		if date := venusGreatestElongationInWindow(start, end, inst); isFiniteFloat(date) && eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		if date := lastVenusGreatestElongationTyped(jde, !east, inst); isFiniteFloat(date) && eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		break
	}
	return latestFiniteEventUT(lastVenusGreatestElongationTyped(jde, true, inst), lastVenusGreatestElongationTyped(jde, false, inst))
}

// LastVenusInferiorConjunctionInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusInferiorConjunctionInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastVenusInferiorConjunctionWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusInferiorConjunctionWithInstance(jde, inst) })
}

// NextVenusInferiorConjunctionInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusInferiorConjunctionInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastVenusInferiorConjunctionWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusInferiorConjunctionWithInstance(jde, inst) })
}

// LastVenusSuperiorConjunctionInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusSuperiorConjunctionInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastVenusSuperiorConjunctionWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusSuperiorConjunctionWithInstance(jde, inst) })
}

// NextVenusSuperiorConjunctionInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusSuperiorConjunctionInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastVenusSuperiorConjunctionWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusSuperiorConjunctionWithInstance(jde, inst) })
}

// LastVenusRetrogradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusRetrogradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastVenusRetrogradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusRetrogradeWithInstance(jde, inst) })
}

// NextVenusRetrogradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusRetrogradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastVenusRetrogradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusRetrogradeWithInstance(jde, inst) })
}

// LastVenusProgradeToRetrogradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusProgradeToRetrogradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastVenusProgradeToRetrogradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusProgradeToRetrogradeWithInstance(jde, inst) })
}

// NextVenusProgradeToRetrogradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusProgradeToRetrogradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastVenusProgradeToRetrogradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusProgradeToRetrogradeWithInstance(jde, inst) })
}

// LastVenusRetrogradeToProgradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusRetrogradeToProgradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastVenusRetrogradeToProgradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusRetrogradeToProgradeWithInstance(jde, inst) })
}

// NextVenusRetrogradeToProgradeInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusRetrogradeToProgradeInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastVenusRetrogradeToProgradeWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusRetrogradeToProgradeWithInstance(jde, inst) })
}

// LastVenusGreatestElongationInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusGreatestElongationInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastVenusGreatestElongationWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusGreatestElongationWithInstance(jde, inst) })
}

// NextVenusGreatestElongationInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusGreatestElongationInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastVenusGreatestElongationWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusGreatestElongationWithInstance(jde, inst) })
}

// LastVenusGreatestElongationEastInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusGreatestElongationEastInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastVenusGreatestElongationEastWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusGreatestElongationEastWithInstance(jde, inst) })
}

// NextVenusGreatestElongationEastInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusGreatestElongationEastInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastVenusGreatestElongationEastWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusGreatestElongationEastWithInstance(jde, inst) })
}

// LastVenusGreatestElongationWestInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusGreatestElongationWestInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastSimpleEvent(jde, func(jde float64) float64 { return LastVenusGreatestElongationWestWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusGreatestElongationWestWithInstance(jde, inst) })
}

// NextVenusGreatestElongationWestInclusiveWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusGreatestElongationWestInclusiveWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextSimpleEvent(jde, func(jde float64) float64 { return LastVenusGreatestElongationWestWithInstance(jde, inst) }, func(jde float64) float64 { return NextVenusGreatestElongationWestWithInstance(jde, inst) })
}

// NextVenusGreatestElongationEastWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusGreatestElongationEastWithInstance(jde float64, inst *Instance) float64 {
	return nextVenusGreatestElongationTyped(jde, true, inst)
}

// NextVenusGreatestElongationWestWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusGreatestElongationWestWithInstance(jde float64, inst *Instance) float64 {
	return nextVenusGreatestElongationTyped(jde, false, inst)
}

// LastVenusGreatestElongationEastWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusGreatestElongationEastWithInstance(jde float64, inst *Instance) float64 {
	return lastVenusGreatestElongationTyped(jde, true, inst)
}

// LastVenusGreatestElongationWestWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusGreatestElongationWestWithInstance(jde float64, inst *Instance) float64 {
	return lastVenusGreatestElongationTyped(jde, false, inst)
}

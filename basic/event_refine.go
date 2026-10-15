package basic

import "math"

const (
	eventNewtonMaxIterations         = 24
	eventDirectionalSearchIterations = 128
	eventRiseSetScanStep             = 1.0 / 1440
	// stationDerivativeStepDay 「留」精修用的中心差分步长（天）：更细的步长会被浮点相消噪声支配。
	stationDerivativeStepDay = 0.01
)

func isFiniteFloat(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// eventNewtonRefine 执行有界牛顿迭代；修正函数返回 f(x)/f'(x)，调用者保留现有导数计算 / eventNewtonRefine performs a bounded Newton iteration. The correction
// 对格式错误输入和不收敛迭代快速失败 / function returns f(x)/f'(x), so callers retain their existing derivative
// 计算 / calculation while malformed input and non-convergent iterations fail fast.
func eventNewtonRefine(seed, tolerance float64, correction func(float64) float64) (float64, bool) {
	if !isFiniteFloat(seed) || !isFiniteFloat(tolerance) || tolerance <= 0 {
		return math.NaN(), false
	}
	current := seed
	for i := 0; i < eventNewtonMaxIterations; i++ {
		step := correction(current)
		if !isFiniteFloat(step) {
			return math.NaN(), false
		}
		next := current - step
		if !isFiniteFloat(next) {
			return math.NaN(), false
		}
		if math.Abs(next-current) <= tolerance {
			return next, true
		}
		current = next
	}
	return math.NaN(), false
}

func eventRiseSetCandidateValid(candidate, civilDayStart, slope float64, isRise bool) bool {
	if !isFiniteFloat(candidate) || !isFiniteFloat(civilDayStart) || !isFiniteFloat(slope) ||
		candidate < civilDayStart || candidate >= civilDayStart+1 {
		return false
	}
	if isRise {
		return slope > 0
	}
	return slope < 0
}

func eventDirectionalRiseSetSearch(civilDayStart float64, isRise bool, fallbackErr error,
	residual func(float64) float64) (float64, error) {
	if !isFiniteFloat(civilDayStart) {
		return 0, ErrInvalidObservationInput
	}

	previousJD := civilDayStart
	previousValue := residual(previousJD)
	if !isFiniteFloat(previousValue) {
		return 0, ErrInvalidObservationInput
	}
	minimum, maximum := previousValue, previousValue
	steps := int(math.Round(1 / eventRiseSetScanStep))
	for i := 1; i <= steps; i++ {
		currentJD := civilDayStart + float64(i)*eventRiseSetScanStep
		currentValue := residual(currentJD)
		if !isFiniteFloat(currentValue) {
			return 0, ErrInvalidObservationInput
		}
		minimum = math.Min(minimum, currentValue)
		maximum = math.Max(maximum, currentValue)
		if eventCrossesDirection(previousValue, currentValue, isRise) {
			eventJD := eventDirectionalBracketRefine(previousJD, currentJD, previousValue, currentValue, residual)
			if eventJD < civilDayStart+1 {
				return eventJD, nil
			}
		}
		previousJD = currentJD
		previousValue = currentValue
	}

	switch {
	case fallbackErr != nil:
		return 0, fallbackErr
	case maximum < 0:
		return 0, ErrNeverRise
	case minimum > 0:
		return 0, ErrNeverSet
	default:
		return 0, ErrNotOnThisDate
	}
}

func eventCrossesDirection(leftValue, rightValue float64, isRise bool) bool {
	if isRise {
		return leftValue <= 0 && rightValue >= 0 && leftValue != rightValue
	}
	return leftValue >= 0 && rightValue <= 0 && leftValue != rightValue
}

func eventDirectionalBracketRefine(leftJD, rightJD, leftValue, rightValue float64, residual func(float64) float64) float64 {
	if leftValue == 0 {
		return leftJD
	}
	if rightValue == 0 {
		return rightJD
	}
	for i := 0; i < 48; i++ {
		middleJD := (leftJD + rightJD) / 2
		middleValue := residual(middleJD)
		if middleValue == 0 {
			return middleJD
		}
		if (leftValue < 0) == (middleValue < 0) {
			leftJD = middleJD
			leftValue = middleValue
		} else {
			rightJD = middleJD
		}
	}
	return (leftJD + rightJD) / 2
}

// eventBracketSecantRoot 在已知异号的括号内用割线法（带中点兜底）求根。
//
// 与 eventDirectionalBracketRefine（固定 48 次二分）相比，割线法通常 5~8 次求值即可达到
// 亚秒精度，适合「括号由廉价截断级数给出、抛光必须用全项级数」的两段式搜索。
// 括号每一步都收缩，因此不会跑到括号外；括号端点同号或出现非有限值时返回 false。
func eventBracketSecantRoot(leftJD, rightJD, leftValue, rightValue, tolerance float64,
	fn func(float64) float64) (float64, bool) {
	if leftValue == 0 {
		return leftJD, true
	}
	if rightValue == 0 {
		return rightJD, true
	}
	if !isFiniteFloat(leftValue) || !isFiniteFloat(rightValue) || leftValue*rightValue > 0 {
		return math.NaN(), false
	}
	if !isFiniteFloat(tolerance) || tolerance <= 0 {
		tolerance = 0.5 / 86400.0
	}
	bestJD := (leftJD + rightJD) / 2
	for i := 0; i < 48; i++ {
		candidateJD := (leftJD + rightJD) / 2
		if rightValue != leftValue {
			secantJD := rightJD - rightValue*(rightJD-leftJD)/(rightValue-leftValue)
			if secantJD > leftJD && secantJD < rightJD {
				candidateJD = secantJD
			}
		}
		candidateValue := fn(candidateJD)
		if !isFiniteFloat(candidateValue) {
			return math.NaN(), false
		}
		bestJD = candidateJD
		if candidateValue == 0 || math.Abs(rightJD-leftJD) <= tolerance {
			return candidateJD, true
		}
		if (leftValue < 0) == (candidateValue < 0) {
			leftJD, leftValue = candidateJD, candidateValue
			continue
		}
		rightJD, rightValue = candidateJD, candidateValue
	}
	if math.Abs(rightJD-leftJD) <= tolerance {
		return bestJD, true
	}
	// 48 次迭代后括号仍宽于容差（例如容差低于该儒略日的 ULP）：调用方把 ok 当作
	// “已抛光到容差”，这里必须报 false，不能返回一个精度未达标的时刻。
	return bestJD, false
}

func eventFixedScanRefine(seed, halfWindow, step float64, fn func(float64) float64) float64 {
	start := seed - halfWindow
	bestJD := start
	bestAbs := math.Abs(fn(start))
	samples := int(math.Round((2 * halfWindow) / step))
	for i := 1; i < samples; i++ {
		candidateJD := start + float64(i)*step
		candidateAbs := math.Abs(fn(candidateJD))
		if candidateAbs < bestAbs {
			bestAbs = candidateAbs
			bestJD = candidateJD
		}
	}
	return bestJD
}

func eventZeroBracket(leftJD, leftVal, centerJD, centerVal, rightJD, rightVal float64) (float64, float64, float64, float64, bool) {
	if leftVal == 0 {
		return leftJD, leftJD, leftVal, leftVal, true
	}
	if centerVal == 0 {
		return centerJD, centerJD, centerVal, centerVal, true
	}
	if rightVal == 0 {
		return rightJD, rightJD, rightVal, rightVal, true
	}
	if leftVal*centerVal < 0 {
		return leftJD, centerJD, leftVal, centerVal, true
	}
	if centerVal*rightVal < 0 {
		return centerJD, rightJD, centerVal, rightVal, true
	}
	if leftVal*rightVal < 0 {
		return leftJD, rightJD, leftVal, rightVal, true
	}
	return 0, 0, 0, 0, false
}

// eventZeroRefine 细化 seed 附近的零点；无可用括号区间时退回固定步长扫描。
func eventZeroRefine(seed, halfWindow, step float64, fn func(float64) float64) float64 {
	leftJD := seed - halfWindow
	centerJD := seed
	rightJD := seed + halfWindow
	leftVal := fn(leftJD)
	centerVal := fn(centerJD)
	rightVal := fn(rightJD)

	bestJD := centerJD
	bestAbs := math.Abs(centerVal)
	if candidateAbs := math.Abs(leftVal); candidateAbs < bestAbs {
		bestAbs = candidateAbs
		bestJD = leftJD
	}
	if candidateAbs := math.Abs(rightVal); candidateAbs < bestAbs {
		bestAbs = candidateAbs
		bestJD = rightJD
	}

	bracketLeftJD, bracketRightJD, bracketLeftVal, bracketRightVal, ok := eventZeroBracket(leftJD, leftVal, centerJD, centerVal, rightJD, rightVal)
	if !ok {
		return eventFixedScanRefine(seed, halfWindow, step, fn)
	}
	if bracketLeftJD == bracketRightJD {
		return bracketLeftJD
	}

	for i := 0; i < 8; i++ {
		candidateJD := (bracketLeftJD + bracketRightJD) / 2
		if bracketRightVal != bracketLeftVal {
			secantJD := bracketRightJD - bracketRightVal*(bracketRightJD-bracketLeftJD)/(bracketRightVal-bracketLeftVal)
			if secantJD > bracketLeftJD && secantJD < bracketRightJD {
				candidateJD = secantJD
			}
		}
		candidateVal := fn(candidateJD)
		candidateAbs := math.Abs(candidateVal)
		if candidateAbs < bestAbs {
			bestAbs = candidateAbs
			bestJD = candidateJD
		}
		if candidateVal == 0 || math.Abs(bracketRightJD-bracketLeftJD) <= step {
			break
		}
		if bracketLeftVal*candidateVal < 0 {
			bracketRightJD = candidateJD
			bracketRightVal = candidateVal
			continue
		}
		bracketLeftJD = candidateJD
		bracketLeftVal = candidateVal
	}
	return bestJD
}

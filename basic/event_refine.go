package basic

import "math"

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

// eventZeroRefine 细化 seed 附近的零点；若找不到可用括号区间，则退回旧的固定步长扫描。
// eventZeroRefine refines a nearby zero crossing and falls back to the legacy
// fixed-step scan when no usable bracket is found.
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

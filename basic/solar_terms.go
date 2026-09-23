package basic

import "math"

func GetMoonLoops(year float64, loop int) []float64 {
	var start float64
	var newMoon, lastNewMoon float64
	moonLoops := make([]float64, loop)
	if year < 6000 {
		start = year + 11.00/12.00 + 5.00/30.00/12.00
	} else {
		start = year + 9.00/12.00 + 5.00/30.00/12.00
	}
	i := 1
	for j := 0; j < loop; j++ {
		if year > 3000 {
			newMoon = TT2UTC(CalcMoonSH(start+float64(i-1)/12.5, 0) + 8.0/24.0)
		} else {
			newMoon = TT2UTC(CalcMoonS(start+float64(i-1)/12.5, 0) + 8.0/24.0)
		}
		if i != 1 {
			if newMoon == lastNewMoon {
				j--
				i++
				continue
			}
		}
		moonLoops[j] = newMoon
		lastNewMoon = moonLoops[j]
		i++
	}
	return moonLoops
}

// GetJieqiLoops 返回从该年冬至起连续 loop 个节气时刻（北京时间，按 15° 一步）：
// 每 24 个节气跨一年，loop<=0 返回 nil；黄经一律归化到 (0, 360]，
// 因此 loop 超过 31 时也不会把 >360° 的角度丢给 GetJQTime（那里会静默 NaN）。
// GetJieqiLoops returns loop consecutive solar-term instants starting at the winter solstice of
// year; 24 terms span one year. Non-positive loop returns nil, and the longitude is normalised into
// (0, 360] so loops beyond 31 never hand an out-of-range angle to GetJQTime.
func GetJieqiLoops(year, loop int) []float64 {
	if loop <= 0 {
		return nil
	}
	start := 270
	jq := make([]float64, loop)
	for i := 1; i <= loop; i++ {
		angle := start + 15*(i-1)
		for angle > 360 {
			angle -= 360
		}
		jq[i-1] = GetJQTime(year+int(math.Ceil(float64(i-1)/24.000)), angle) + 8.0/24.0
	}
	return jq
}

func GetJQTime(year, angle int) float64 {
	// Calculate initial day based on angle parity
	var initialDay float64
	if angle%2 == 0 {
		initialDay = 18
	} else {
		initialDay = 3
	}

	// Calculate temporary factor for month offset
	var tempFactor float64
	if angle%10 != 0 {
		tempFactor = float64(angle+15) / 30.0
	} else {
		tempFactor = float64(angle) / 30.0
	}

	// Calculate initial month, adjusting if超过 12
	initialMonth := 3.0 + tempFactor
	if initialMonth > 12.0 {
		initialMonth -= 12.0
	}

	// Calculate initial Julian date
	initialJD := JDCalc(year, int(initialMonth), initialDay)

	// Set target angle for iteration; if angle is 0, use 360
	targetAngle := float64(angle)
	if angle == 0 {
		targetAngle = 360.0
	}

	// Newton-Raphson iteration to find precise Julian date
	currentJDE := initialJD
	var ok bool
	currentJDE, ok = eventNewtonRefine(currentJDE, 0.00001, func(previousJD float64) float64 {
		errorValue := JQLospec(previousJD, targetAngle) - targetAngle
		derivative := (JQLospec(previousJD+0.000005, targetAngle) - JQLospec(previousJD-0.000005, targetAngle)) / 0.00001
		return errorValue / derivative
	})
	if !ok {
		return math.NaN()
	}

	// Convert to UT and return
	return TT2UTC(currentJDE)
}

func JQLospec(jde float64, target float64) float64 {
	sunLo := HSunApparentLo(jde)
	if target >= 345 {
		if sunLo <= 12 {
			sunLo += 360
		}
	} else if target <= 15 {
		if sunLo >= 350 {
			sunLo -= 360
		}
	}
	return sunLo
}

package basic

import "math"

func GetMoonLoops(year float64, loop int) []float64 {
	var start float64
	var newMoon, lastNewMoon float64
	if loop <= 0 {
		return nil
	}
	moonLoops := make([]float64, loop)
	if year < 6000 {
		start = year + 11.00/12.00 + 5.00/30.00/12.00
	} else {
		start = year + 9.00/12.00 + 5.00/30.00/12.00
	}
	i := 1
	// 重复朔只跳过、不计入，最多容忍与 loop 同量级的重试。
	maxAttempts := 2*loop + 4
	j := 0
	for ; j < loop && i <= maxAttempts; j++ {
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
	return moonLoops[:j]
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
	target := solarTermTarget(angle)
	tt, ok := solveSolarTermTT(solarTermSeed(year, angle), target, func(jde float64) float64 {
		return JQLospec(jde, target)
	})
	if !ok {
		return math.NaN()
	}
	return TT2UTC(tt)
}

func JQLospec(jde float64, target float64) float64 {
	return wrapSolarLongitude(HSunApparentLo(jde), target)
}

// 太阳视黄经平均日变化（度/天），只用于迭代步长。
const solarTermMeanRate = 360 / 365.2422

const solarTermToleranceDay = 0.00001

// 粗解收敛到该步长（天）即转精确求解。
const solarTermCoarseToleranceDay = 0.001

// solarTermSeed 按节气角度给出所在月份的迭代初值。
func solarTermSeed(year, angle int) float64 {
	initialDay := 18.0
	if angle%2 != 0 {
		initialDay = 3
	}
	tempFactor := float64(angle) / 30
	if angle%10 != 0 {
		tempFactor = float64(angle+15) / 30
	}
	initialMonth := 3 + tempFactor
	if initialMonth > 12 {
		initialMonth -= 12
	}
	return JDCalc(year, int(initialMonth), initialDay)
}

// solarTermTarget 0° 按 360° 求解，与春分前的视黄经展开方向一致。
func solarTermTarget(angle int) float64 {
	if angle == 0 {
		return 360
	}
	return float64(angle)
}

// wrapSolarLongitude 把视黄经展开到目标附近，避免 0/360° 跳变。
func wrapSolarLongitude(sunLo, target float64) float64 {
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

// 割线解离初值超过该天数时改走原牛顿迭代，以保持原选根结果。
const solarTermSameRootDays = 19

// solveSolarTermTT 先用低精度视黄经逼近，再以精确视黄经割线法收敛，返回 TT。
func solveSolarTermTT(seed, target float64, precise func(float64) float64) (float64, bool) {
	if !isFiniteFloat(seed) {
		return math.NaN(), false
	}
	if coarse, ok := solarTermCoarse(seed, target); ok {
		if tt, ok := solarTermSecant(coarse, target, precise); ok && math.Abs(tt-seed) <= solarTermSameRootDays {
			return tt, true
		}
	}
	// 古代儒略历初值偏离较大时，原牛顿法可能收敛到相邻年份的根，须保持一致。
	return solarTermNewton(seed, target, precise)
}

func solarTermNewton(seed, target float64, precise func(float64) float64) (float64, bool) {
	const h = solarTermToleranceDay / 2
	return eventNewtonRefine(seed, solarTermToleranceDay, func(x float64) float64 {
		residual := precise(x) - target
		derivative := (precise(x+h) - precise(x-h)) / solarTermToleranceDay
		return residual / derivative
	})
}

func solarTermCoarse(seed, target float64) (float64, bool) {
	x := seed
	for i := 0; i < eventNewtonMaxIterations; i++ {
		step := (wrapSolarLongitude(SunApparentLo(x), target) - target) / solarTermMeanRate
		if !isFiniteFloat(step) {
			return math.NaN(), false
		}
		x -= step
		if math.Abs(step) <= solarTermCoarseToleranceDay {
			return x, true
		}
	}
	return math.NaN(), false
}

// solarTermSecant 导数取相邻两次残差，每步只求一次精确视黄经。
func solarTermSecant(seed, target float64, precise func(float64) float64) (float64, bool) {
	x0 := seed
	f0 := precise(x0) - target
	if !isFiniteFloat(f0) {
		return math.NaN(), false
	}
	x1 := x0 - f0/solarTermMeanRate
	for i := 0; i < eventNewtonMaxIterations; i++ {
		f1 := precise(x1) - target
		if !isFiniteFloat(f1) {
			return math.NaN(), false
		}
		if f1 == f0 {
			// 残差不再变化说明已到浮点分辨率；残差仍大则是落在跳变两侧。
			return x1, math.Abs(f1) <= solarTermMeanRate*solarTermToleranceDay
		}
		x2 := x1 - f1*(x1-x0)/(f1-f0)
		if !isFiniteFloat(x2) {
			return math.NaN(), false
		}
		if math.Abs(x2-x1) <= solarTermToleranceDay {
			return x2, true
		}
		x0, f0, x1 = x1, f1, x2
	}
	return math.NaN(), false
}

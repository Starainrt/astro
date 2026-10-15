package basic

import (
	"math"
	"sync"
)

// TimeScaleFuturePolicy 民用时标换算政策，零值为闰时 / civil-time conversion policy, leap hour by default.
type TimeScaleFuturePolicy int

const (
	// TimeScaleLeapHour 按 ΔT 越限施加整小时校正（2027-05-20 起 CGPM 决议） / applies hour-sized corrections as adopted by CGPM from 2027.
	TimeScaleLeapHour TimeScaleFuturePolicy = iota
	// TimeScaleLeapSecond 按 ΔT 越限施加整数秒校正（1972–2026 确定值） / integer-second corrections (definitive 1972–2026).
	TimeScaleLeapSecond
	// TimeScaleAssumeUT1Tracking 窗口外固定末端 DUT1，平滑跟随 UT1 / holds the last observed DUT1 beyond the window.
	TimeScaleAssumeUT1Tracking
	// TimeScaleFreezeUTCOffset 窗口外固定末端 TT−UTC / holds the last TT−UTC offset beyond the window.
	TimeScaleFreezeUTCOffset
	// TimeScaleUT1Civil 全时轴民用时标等同 UT1，TT−UTC 覆盖仍优先 / uses UT1 as civil time everywhere unless TT−UTC is overridden.
	TimeScaleUT1Civil
)

// 闰时情景的校正步长与容限，单位秒。
const timeScaleLeapHourSeconds = 3600.0

// 闰秒情景的 DUT1 容限，单位秒。
const utcDUT1ToleranceSeconds = 0.9

// utcEraStartJDE 是 UTC 按 SI 秒运行的起点（1972-01-01），此前民用时标即 UT1。
const utcEraStartJDE = 2441317.5

// 月度 ΔT 表覆盖窗口末端，必须与表末项同步。
const timeScaleExactEndJDE = 2461314.5

// 闰秒生效的儒略日按降序排列，与 utcLeapOffsets 一一对应。
var utcLeapJDEs = [...]float64{
	2457754.5, 2457204.5, 2456109.5, 2454832.5,
	2453736.5, 2451179.5, 2450630.5, 2450083.5,
	2449534.5, 2449169.5, 2448804.5, 2448257.5,
	2447892.5, 2447161.5, 2446247.5, 2445516.5,
	2445151.5, 2444786.5, 2444239.5, 2443874.5,
	2443509.5, 2443144.5, 2442778.5, 2442413.5,
	2442048.5, 2441683.5, 2441499.5, 2441133.5,
}

// utcLeapOffsets[i] 是 utcLeapJDEs[i] 起生效的 TT−UTC（秒），与上表顺序一致。
var utcLeapOffsets = [...]float64{
	69.184, 68.184, 67.184, 66.184,
	65.184, 64.184, 63.184, 62.184,
	61.184, 60.184, 59.184, 58.184,
	57.184, 56.184, 55.184, 54.184,
	53.184, 52.184, 51.184, 50.184,
	49.184, 48.184, 47.184, 46.184,
	45.184, 44.184, 43.184, 42.184,
}

// utcLeapBaseSeconds 是 1972-01-01 起算的 TT−UTC 基值：32.184 + (TAI−UTC 的 10 秒)。
const utcLeapBaseSeconds = 42.184

var (
	timeScaleMu  sync.RWMutex
	ttMinusUTCFn func(float64) float64
	futurePolicy TimeScaleFuturePolicy
)

// SetTTMinusUTCFn 覆盖 TT−UTC，nil 恢复内置表与政策 / overrides TT−UTC; nil restores the built-in table and policy.
func SetTTMinusUTCFn(fn func(float64) float64) {
	timeScaleMu.Lock()
	ttMinusUTCFn = fn
	timeScaleMu.Unlock()
	deltaTFnMu.Lock()
	deltaTGeneration++
	deltaTFnMu.Unlock()
}

// GetTTMinusUTCFn 返回当前 TT−UTC 覆盖函数，未设置时为 nil / current TT−UTC override, or nil.
func GetTTMinusUTCFn() func(float64) float64 {
	timeScaleMu.RLock()
	defer timeScaleMu.RUnlock()
	return ttMinusUTCFn
}

// SetTimeScaleFuturePolicy 设置民用时标换算政策 / sets the civil-time conversion policy.
func SetTimeScaleFuturePolicy(policy TimeScaleFuturePolicy) {
	timeScaleMu.Lock()
	futurePolicy = policy
	timeScaleMu.Unlock()
	deltaTFnMu.Lock()
	deltaTGeneration++
	deltaTFnMu.Unlock()
}

// GetTimeScaleFuturePolicy 返回当前民用时标换算政策 / current civil-time conversion policy.
func GetTimeScaleFuturePolicy() TimeScaleFuturePolicy {
	timeScaleMu.RLock()
	defer timeScaleMu.RUnlock()
	return futurePolicy
}

// TTMinusUTCSeconds 返回 TT−UTC 覆盖值或内置表值，不应用未来政策 / TT−UTC override or leap-table seconds, without future policy.
func TTMinusUTCSeconds(jd float64) float64 {
	timeScaleMu.RLock()
	fn := ttMinusUTCFn
	timeScaleMu.RUnlock()
	if fn != nil {
		return fn(jd)
	}
	return ttMinusUTCSecondsDefault(jd)
}

// UTC2TT 民用时刻转 TT，1972 年前按 UT1、之后按覆盖或所选政策 / converts civil time to TT using UT1 before 1972, then the override or selected policy.
func UTC2TT(jd float64) float64 {
	return utcToTTJDE(jd)
}

// TT2UTC 反解 UTC2TT，阶跃区间内往返不唯一 / inverts UTC2TT, with ambiguous round trips within step intervals.
func TT2UTC(ttJDE float64) float64 {
	return ttToUTCJDE(ttJDE)
}

// UT12TT 按当前 ΔT 模型将 UT1 转为 TT / converts UT1 to TT using the active ΔT model.
func UT12TT(jd float64) float64 {
	return ut1ToTTJDE(jd)
}

// TT2UT1 是 UT12TT 的逆 / inverts UT12TT.
func TT2UT1(ttJDE float64) float64 {
	return ttToUT1JDE(ttJDE)
}

// UTC2UT1 民用时刻转 UT1，窗口之后按当前未来政策 / converts civil time to UT1 under the active future policy.
func UTC2UT1(jd float64) float64 {
	return ttToUT1JDE(utcToTTJDE(jd))
}

// UT12UTC 是 UTC2UT1 的逆 / inverts UTC2UT1.
func UT12UTC(ut1JDE float64) float64 {
	return ttToUTCJDE(ut1ToTTJDE(ut1JDE))
}

// DUT1Seconds 返回 UT1−UTC（秒），采用当前覆盖与政策 / UT1−UTC in seconds under the active overrides and policy.
func DUT1Seconds(jd float64) float64 {
	return utcToTTOffsetSeconds(jd) - ut1ToTTOffsetSeconds(jd)
}

// 表外沿用最近的已知偏移，不应用未来政策。
func ttMinusUTCSecondsDefault(jd float64) float64 {
	for i := range utcLeapJDEs {
		if jd >= utcLeapJDEs[i] {
			return utcLeapOffsets[i]
		}
	}
	return utcLeapBaseSeconds
}

// TTMinusUTCSecondsDefault 返回内置闰秒表的 TT−UTC（秒），忽略覆盖 / built-in TT−UTC seconds, ignoring overrides.
func TTMinusUTCSecondsDefault(jd float64) float64 {
	return ttMinusUTCSecondsDefault(jd)
}

// deltaTModelSecondsAtUT 是内置 ΔT 模型（秒）：覆盖期内用逐月实测表（毫秒级），表外用外推样条。
func deltaTModelSecondsAtUT(jd float64) float64 {
	if seconds, ok := deltaTMonthlyAt(jd); ok {
		return seconds
	}
	// 两端按常值锚定，避免 ΔT 跳变使反解落到窗口另一侧。
	if jd < deltaTMonthlyJDE[0] {
		return deltaTSplineAtJDE(jd) + deltaTMonthlyObserved[0] - deltaTSplineAtJDE(deltaTMonthlyJDE[0])
	}
	last := len(deltaTMonthlyJDE) - 1
	return deltaTSplineAtJDE(jd) + deltaTMonthlyObserved[last] - deltaTSplineAtJDE(deltaTMonthlyJDE[last])
}

// deltaTMonthlyAt 在逐月实测表覆盖范围内线性插值，范围外 ok 为 false。
func deltaTMonthlyAt(jd float64) (float64, bool) {
	count := len(deltaTMonthlyJDE)
	if count == 0 || jd < deltaTMonthlyJDE[0] || jd > deltaTMonthlyJDE[count-1] {
		return 0, false
	}
	low, high := 0, count-1
	for high-low > 1 {
		mid := (low + high) / 2
		if deltaTMonthlyJDE[mid] <= jd {
			low = mid
		} else {
			high = mid
		}
	}
	span := deltaTMonthlyJDE[high] - deltaTMonthlyJDE[low]
	if span <= 0 {
		return deltaTMonthlyObserved[low], true
	}
	ratio := (jd - deltaTMonthlyJDE[low]) / span
	return deltaTMonthlyObserved[low] + ratio*(deltaTMonthlyObserved[high]-deltaTMonthlyObserved[low]), true
}

// deltaTYearAtJDE 把 UT 儒略日换成十进制年：1582 改历之后按格里高利年平均长度，之前按儒略年。
func deltaTYearAtJDE(jd float64) float64 {
	if jd >= 2299160.5 {
		return (jd-2451544.5)/365.2425 + 2000
	}
	return (jd+0.5)/365.25 - 4712
}

// deltaTSplineAtJDE 用外推样条求 TT−UT1（秒），覆盖表外的古代与未来。
func deltaTSplineAtJDE(jd float64) float64 {
	year := deltaTYearAtJDE(jd)
	return DeltaTSplineY(year)
}

// 直接求秒差，避免两个大 JD 相减导致精度损失。
func utcToTTOffsetSeconds(jd float64) float64 {
	if jd < utcEraStartJDE {
		// 1972 前民用时标即 UT1。
		return ut1ToTTOffsetSeconds(jd)
	}
	// 1972 年后显式覆盖优先于全部政策。
	if fn := GetTTMinusUTCFn(); fn != nil {
		return fn(jd)
	}
	// UT1Civil 替换全时轴定义，须先于窗口内闰秒表判断。
	if GetTimeScaleFuturePolicy() == TimeScaleUT1Civil {
		return ut1ToTTOffsetSeconds(jd)
	}
	// 内置实测窗口（截至 2026-10-01）一律取确定闰秒表值，窗口外才按政策外推。
	if jd <= timeScaleExactEndJDE {
		return TTMinusUTCSeconds(jd)
	}
	switch GetTimeScaleFuturePolicy() {
	case TimeScaleLeapHour:
		return steppedUTCOffsetSeconds(jd, timeScaleLeapHourSeconds, timeScaleLeapHourSeconds)
	case TimeScaleLeapSecond:
		return steppedUTCOffsetSeconds(jd, 1, utcDUT1ToleranceSeconds)
	case TimeScaleAssumeUT1Tracking:
		return ut1ToTTOffsetSeconds(jd) + dut1AtTimeScaleExactEnd()
	case TimeScaleFreezeUTCOffset:
		return TTMinusUTCSeconds(timeScaleExactEndJDE)
	}
	return steppedUTCOffsetSeconds(jd, timeScaleLeapHourSeconds, timeScaleLeapHourSeconds)
}

// 相对窗口末端偏移施加最少整数步校正，使 DUT1 落回容限内。
func steppedUTCOffsetSeconds(jd, step, tolerance float64) float64 {
	base := TTMinusUTCSeconds(timeScaleExactEndJDE)
	drift := base - ut1ToTTOffsetSeconds(jd)
	if math.IsNaN(drift) || math.IsInf(drift, 0) {
		return math.NaN()
	}
	switch {
	case drift <= -tolerance:
		return base + step*(math.Floor((-drift-tolerance)/step)+1)
	case drift >= tolerance:
		return base - step*(math.Floor((drift-tolerance)/step)+1)
	}
	return base
}

// dut1AtTimeScaleExactEnd 是窗口末端的实测 UT1−UTC（秒），TimeScaleAssumeUT1Tracking 沿用此值。
func dut1AtTimeScaleExactEnd() float64 {
	return TTMinusUTCSeconds(timeScaleExactEndJDE) - ut1ToTTOffsetSeconds(timeScaleExactEndJDE)
}

func ut1ToTTOffsetSeconds(jd float64) float64 {
	return DeltaT(jd, true)
}

func utcToTTJDE(jd float64) float64 {
	return jd + utcToTTOffsetSeconds(jd)/86400
}

func ut1ToTTJDE(jd float64) float64 {
	return jd + ut1ToTTOffsetSeconds(jd)/86400
}

// ttToUT1JDE 解 TT = UT12TT(ut)，与正向使用同一模型，保证往返一致。
func ttToUT1JDE(ttJDE float64) float64 {
	ut := ttJDE - ut1ToTTOffsetSeconds(ttJDE)/86400
	for iteration := 0; iteration < 4; iteration++ {
		next := ttJDE - ut1ToTTOffsetSeconds(ut)/86400
		if next == ut {
			break
		}
		ut = next
	}
	return ut
}

// ttToUTCJDE 解 TT = UTC2TT(utc)：分支由候选时刻判定，否则边界处会来回振荡。
func ttToUTCJDE(ttJDE float64) float64 {
	if math.IsNaN(ttJDE) || math.IsInf(ttJDE, 0) {
		return math.NaN()
	}
	utc := ttJDE - utcToTTOffsetSeconds(ttJDE)/86400
	for iteration := 0; iteration < 4; iteration++ {
		next := ttJDE - utcToTTOffsetSeconds(utc)/86400
		if next == utc {
			break
		}
		utc = next
	}
	return utc
}

var defDeltaTFn = DefaultDeltaTv2
var deltaTFnMu sync.RWMutex
var activeDeltaTModel = DeltaTModelDefault
var activeDeltaTKeepObserved = true

// 配置变化使缓存世代递增，起始为 1 以排除零值缓存。
var deltaTGeneration uint64 = 1

// 两种输入口径共用的 ΔT 模型年限。
const deltaTValidYearSpan = 40000.0

// DeltaT 返回 TT−UT1（秒），julianDay 为真取 UT JD、否则取十进制年，超出 ±40000 年返回 NaN / TT−UT1 seconds from a UT JD if julianDay, otherwise a decimal year; NaN beyond ±40000 years.
func DeltaT(date float64, julianDay bool) float64 {
	if !math.IsNaN(date) && !math.IsInf(date, 0) && math.Abs(deltaTArgumentYear(date, julianDay)) > deltaTValidYearSpan {
		return math.NaN()
	}
	deltaTFnMu.RLock()
	fn := defDeltaTFn
	deltaTFnMu.RUnlock()
	return fn(date, julianDay)
}

// deltaTArgumentYear 把两种自变量口径统一成十进制年，只用于越界判定。
func deltaTArgumentYear(date float64, julianDay bool) float64 {
	if julianDay {
		return deltaTYearAtJDE(date)
	}
	return math.Floor(date)
}

// SetDeltaTFn 注入秒单位的 ΔT 模型并标记为 Manual，nil 恢复默认 / installs a manual ΔT model in seconds; nil restores the default.
func SetDeltaTFn(fn func(float64, bool) float64) {
	if fn == nil {
		installDeltaTFn(DefaultDeltaTv2, DeltaTModelDefault, true)
		return
	}
	installDeltaTFn(fn, DeltaTModelManual, false)
}

// installDeltaTFn 是 ΔT 钩子的唯一写入口：函数、模型标记与世代号一起更新。
func installDeltaTFn(fn func(float64, bool) float64, model DeltaTModel, keepObserved bool) {
	deltaTFnMu.Lock()
	defDeltaTFn = fn
	activeDeltaTModel = model
	activeDeltaTKeepObserved = keepObserved
	deltaTGeneration++
	deltaTFnMu.Unlock()
}

func deltaTGenerationValue() uint64 {
	deltaTFnMu.RLock()
	value := deltaTGeneration
	deltaTFnMu.RUnlock()
	return value
}

// GetDeltaTFn 返回当前生效的 ΔT 计算函数，其入参为年或儒略日 / current ΔT function, taking a year or a Julian day.
func GetDeltaTFn() func(float64, bool) float64 {
	deltaTFnMu.RLock()
	fn := defDeltaTFn
	deltaTFnMu.RUnlock()
	return fn
}

// DefaultDeltaTv2 库内默认 ΔT 计算，date 为年或儒略日（isJd），返回秒 / default ΔT computation in seconds.
func DefaultDeltaTv2(date float64, isJd bool) float64 {
	if math.IsNaN(date) || math.IsInf(date, 0) {
		return math.NaN()
	}
	if !isJd {
		year := math.Floor(date)
		start := JDCalc(int(year), 1, 1)
		end := JDCalc(int(year)+1, 1, 1)
		date = start + (date-year)*(end-start)
	}
	return DeltaTv2(date)
}

// DeltaTSplineY 按十进制年计算样条与长期外推 ΔT（秒）/ spline and long-term ΔT in seconds for a decimal year.
func DeltaTSplineY(y float64) float64 {
	if math.IsNaN(y) || math.IsInf(y, 0) {
		return math.NaN()
	}
	// 日长偏差积分后的秒差，积分常数在两侧分别锚定。
	integratedLod := func(x float64) float64 {
		u := x - 1825
		return 3.14115e-3*u*u + 284.8435805251424*math.Cos(0.4487989505128276*(0.01*u+0.75))
	}

	if y < -720 {
		const c = 1.007739546148514
		return integratedLod(y) + c
	}
	if y > 2025 {
		const c = -150.56787057979514
		return integratedLod(y) + c
	}

	n := len(deltaTSplineY0)
	var i int
	for i = n - 1; i >= 0; i-- {
		if y >= deltaTSplineY0[i] {
			break
		}
	}
	t := (y - deltaTSplineY0[i]) / (deltaTSplineY1[i] - deltaTSplineY0[i])
	dT := deltaTSplineA0[i] + t*(deltaTSplineA1[i]+t*(deltaTSplineA2[i]+t*deltaTSplineA3[i]))
	return dT
}

// 区间内 t = (year−Y0)/(Y1−Y0)，ΔT = A0 + t·(A1 + t·(A2 + t·A3))。
var deltaTSplineY0 = [...]float64{-720, -100, 400, 1000, 1150, 1300, 1500, 1600, 1650, 1720, 1800, 1810, 1820, 1830, 1840, 1850, 1855, 1860, 1865, 1870, 1875, 1880, 1885, 1890, 1895, 1900, 1905, 1910, 1915, 1920, 1925, 1930, 1935, 1940, 1945, 1950, 1953, 1956, 1959, 1962, 1965, 1968, 1971, 1974, 1977, 1980, 1983, 1986, 1989, 1992, 1995, 1998, 2001, 2004, 2007, 2010, 2013, 2016, 2019, 2022}
var deltaTSplineY1 = [...]float64{-100, 400, 1000, 1150, 1300, 1500, 1600, 1650, 1720, 1800, 1810, 1820, 1830, 1840, 1850, 1855, 1860, 1865, 1870, 1875, 1880, 1885, 1890, 1895, 1900, 1905, 1910, 1915, 1920, 1925, 1930, 1935, 1940, 1945, 1950, 1953, 1956, 1959, 1962, 1965, 1968, 1971, 1974, 1977, 1980, 1983, 1986, 1989, 1992, 1995, 1998, 2001, 2004, 2007, 2010, 2013, 2016, 2019, 2022, 2025}
var deltaTSplineA0 = [...]float64{20371.848, 11557.668, 6535.116, 1650.393, 1056.647, 681.149, 292.343, 109.127, 43.952, 12.068, 18.367, 15.678, 16.516, 10.804, 7.634, 9.338, 10.357, 9.04, 8.255, 2.371, -1.126, -3.21, -4.388, -3.884, -5.017, -1.977, 4.923, 11.142, 17.479, 21.617, 23.789, 24.418, 24.164, 24.426, 27.05, 28.932, 30.002, 30.76, 32.652, 33.621, 35.093, 37.956, 40.951, 44.244, 47.291, 50.361, 52.936, 54.984, 56.373, 58.453, 60.678, 62.898, 64.083, 64.553, 65.197, 66.061, 66.919, 68.130, 69.250, 69.296}
var deltaTSplineA1 = [...]float64{-9999.586, -5822.27, -5671.519, -753.21, -459.628, -421.345, -192.841, -78.697, -68.089, 2.507, -3.481, 0.021, -2.157, -6.018, -0.416, 1.642, -0.486, -0.591, -3.456, -5.593, -2.314, -1.893, 0.101, -0.531, 0.134, 5.715, 6.828, 6.33, 5.518, 3.02, 1.333, 0.052, -0.419, 1.645, 2.499, 1.127, 0.737, 1.409, 1.577, 0.868, 2.275, 3.035, 3.157, 3.199, 3.069, 2.878, 2.354, 1.577, 1.648, 2.235, 2.324, 1.804, 0.674, 0.466, 0.804, 0.839, 1.005, 1.348, 0.594, -0.227}
var deltaTSplineA2 = [...]float64{776.247, 1303.151, -298.291, 184.811, 108.771, 61.953, -6.572, 10.505, 38.333, 41.731, -1.126, 4.629, -6.806, 2.944, 2.658, 0.261, -2.389, 2.284, -5.148, 3.011, 0.269, 0.152, 1.842, -2.474, 3.138, 2.443, -1.329, 0.831, -1.643, -0.856, -0.831, -0.449, -0.022, 2.086, -1.232, 0.22, -0.61, 1.282, -1.115, 0.406, 1.002, -0.242, 0.364, -0.323, 0.193, -0.384, -0.14, -0.637, 0.708, -0.121, 0.21, -0.729, -0.402, 0.194, 0.144, -0.109, 0.275, 0.068, -0.822, 0.001}
var deltaTSplineA3 = [...]float64{409.16, -503.433, 1085.087, -25.346, -24.641, -29.414, 16.197, 3.018, -2.127, -37.939, 1.918, -3.812, 3.25, -0.096, -0.539, -0.883, 1.558, -2.477, 2.72, -0.914, -0.039, 0.563, -1.438, 1.871, -0.232, -1.257, 0.72, -0.825, 0.262, 0.008, 0.127, 0.142, 0.702, -1.106, 0.614, -0.277, 0.631, -0.799, 0.507, 0.199, -0.414, 0.202, -0.229, 0.172, -0.192, 0.081, -0.165, 0.448, -0.276, 0.11, -0.313, 0.109, 0.199, -0.017, -0.084, 0.128, -0.069, -0.297, 0.274, 0.086}

// DeltaTv2 返回内置模型的 TT−UT1（秒），实测表外采用外推样条 / built-in TT−UT1 seconds, with spline extrapolation outside the observed table.
func DeltaTv2(jd float64) float64 {
	if math.IsNaN(jd) || math.IsInf(jd, 0) {
		return math.NaN()
	}
	return deltaTModelSecondsAtUT(jd)
}

// DeltaTSecondsAt 返回 TT 时刻的 ΔT，有限覆盖值优先，NaN/Inf 选择当前模型 / ΔT at a TT instant, using a finite override or the active model.
func DeltaTSecondsAt(jdeTT, overrideSeconds float64) float64 {
	if !math.IsNaN(overrideSeconds) && !math.IsInf(overrideSeconds, 0) {
		return overrideSeconds
	}
	return deltaTModelSecondsAtTT(jdeTT)
}

func deltaTModelSecondsAtTT(jdeTT float64) float64 {
	// ΔT 模型以 UT 为自变量，不能直接传 TT。
	return ut1ToTTOffsetSeconds(ttToUT1JDE(jdeTT))
}

// DeltaTGroundShiftKM 将 ΔT 误差换算为站点相对影子的地面横移距离（千米）/ ground displacement in kilometres from a ΔT error.
func DeltaTGroundShiftKM(deltaTSeconds, latitudeDeg float64) float64 {
	if math.IsNaN(deltaTSeconds) || math.IsInf(deltaTSeconds, 0) ||
		math.IsNaN(latitudeDeg) || math.IsInf(latitudeDeg, 0) {
		return math.NaN()
	}
	return solarEclipseEarthEquatorialRotationKMPerSecond * math.Abs(deltaTSeconds) * math.Cos(latitudeDeg*rad)
}

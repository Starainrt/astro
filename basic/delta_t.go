package basic

import (
	"math"
	"sync"
)

var defDeltaTFn = DefaultDeltaTv2
var deltaTFnMu sync.RWMutex

// deltaTGeneration 随每次 ΔT 覆盖递增，供依赖 ΔT 的只读记忆表判断自身是否过期。
// 起始为 1，使零值缓存条目（世代 0）天然视为未命中。
// deltaTGeneration increments on every ΔT override so ΔT-dependent memo tables can detect
// staleness. It starts at 1 so a zero-valued cache entry (generation 0) is never a hit.
var deltaTGeneration uint64 = 1

func DeltaT(date float64, isJDE bool) float64 {
	deltaTFnMu.RLock()
	fn := defDeltaTFn
	deltaTFnMu.RUnlock()
	return fn(date, isJDE)
}

func SetDeltaTFn(fn func(float64, bool) float64) {
	if fn != nil {
		deltaTFnMu.Lock()
		defDeltaTFn = fn
		deltaTGeneration++
		deltaTFnMu.Unlock()
	}
}

// deltaTGenerationValue 返回当前 ΔT 世代，用于让只读记忆表在 ΔT 改变后整体失效。
// deltaTGenerationValue returns the current ΔT generation so memo tables can be invalidated.
func deltaTGenerationValue() uint64 {
	deltaTFnMu.RLock()
	value := deltaTGeneration
	deltaTFnMu.RUnlock()
	return value
}

func GetDeltaTFn() func(float64, bool) float64 {
	deltaTFnMu.RLock()
	fn := defDeltaTFn
	deltaTFnMu.RUnlock()
	return fn
}

func DefaultDeltaTv2(date float64, isJd bool) float64 { //传入年或儒略日，传出为秒
	if math.IsNaN(date) || math.IsInf(date, 0) {
		return math.NaN()
	}
	if !isJd {
		year := math.Floor(date)
		start := JDECalc(int(year), 1, 1)
		end := JDECalc(int(year)+1, 1, 1)
		date = start + (date-year)*(end-start)
	}
	return DeltaTv2(date)
}

// 使用Stephenson等人(2016)和Morrison等人(2021)的拟合和外推公式计算Delta T
// http://astro.ukho.gov.uk/nao/lvm/
// 2010年后的系数已修改以包含2019年后的数据
// 返回Delta T，单位为秒
func DeltaTSplineY(y float64) float64 {
	if math.IsNaN(y) || math.IsInf(y, 0) {
		return math.NaN()
	}
	// 积分lod（平均太阳日偏离86400秒的偏差）方程：
	// 来自 http://astro.ukho.gov.uk/nao/lvm/：
	// lod = 1.72 t − 3.5 sin(2*pi*(t+0.75)/14) 单位ms/day，其中 t = (y - 1825)/100
	// 是从1825年开始的世纪数
	// 使用 1ms = 1e-3s 和 1儒略年 = 365.25天，
	// lod = 6.2823e-3 * Delta y - 1.278375*sin(2*pi/14*(Delta y /100 + 0.75) 单位s/year
	// 其中 Delta y = y - 1825。积分该方程得到
	// Integrate[lod, y] = 3.14115e-3*(Delta y)^2 + 894.8625/pi*cos(2*pi/14*(Delta y /100 + 0.75)
	// 单位为秒。积分常数设为0。
	integratedLod := func(x float64) float64 {
		u := x - 1825
		return 3.14115e-3*u*u + 284.8435805251424*math.Cos(0.4487989505128276*(0.01*u+0.75))
	}

	if y < -720 {
		// 使用积分lod + 常数
		const c = 1.007739546148514
		return integratedLod(y) + c
	}
	if y > 2025 {
		// 使用积分lod + 常数
		const c = -150.56787057979514
		return integratedLod(y) + c
	}

	// 使用三次样条拟合
	y0 := []float64{-720, -100, 400, 1000, 1150, 1300, 1500, 1600, 1650, 1720, 1800, 1810, 1820, 1830, 1840, 1850, 1855, 1860, 1865, 1870, 1875, 1880, 1885, 1890, 1895, 1900, 1905, 1910, 1915, 1920, 1925, 1930, 1935, 1940, 1945, 1950, 1953, 1956, 1959, 1962, 1965, 1968, 1971, 1974, 1977, 1980, 1983, 1986, 1989, 1992, 1995, 1998, 2001, 2004, 2007, 2010, 2013, 2016, 2019, 2022}
	y1 := []float64{-100, 400, 1000, 1150, 1300, 1500, 1600, 1650, 1720, 1800, 1810, 1820, 1830, 1840, 1850, 1855, 1860, 1865, 1870, 1875, 1880, 1885, 1890, 1895, 1900, 1905, 1910, 1915, 1920, 1925, 1930, 1935, 1940, 1945, 1950, 1953, 1956, 1959, 1962, 1965, 1968, 1971, 1974, 1977, 1980, 1983, 1986, 1989, 1992, 1995, 1998, 2001, 2004, 2007, 2010, 2013, 2016, 2019, 2022, 2025}
	a0 := []float64{20371.848, 11557.668, 6535.116, 1650.393, 1056.647, 681.149, 292.343, 109.127, 43.952, 12.068, 18.367, 15.678, 16.516, 10.804, 7.634, 9.338, 10.357, 9.04, 8.255, 2.371, -1.126, -3.21, -4.388, -3.884, -5.017, -1.977, 4.923, 11.142, 17.479, 21.617, 23.789, 24.418, 24.164, 24.426, 27.05, 28.932, 30.002, 30.76, 32.652, 33.621, 35.093, 37.956, 40.951, 44.244, 47.291, 50.361, 52.936, 54.984, 56.373, 58.453, 60.678, 62.898, 64.083, 64.553, 65.197, 66.061, 66.919, 68.130, 69.250, 69.296}
	a1 := []float64{-9999.586, -5822.27, -5671.519, -753.21, -459.628, -421.345, -192.841, -78.697, -68.089, 2.507, -3.481, 0.021, -2.157, -6.018, -0.416, 1.642, -0.486, -0.591, -3.456, -5.593, -2.314, -1.893, 0.101, -0.531, 0.134, 5.715, 6.828, 6.33, 5.518, 3.02, 1.333, 0.052, -0.419, 1.645, 2.499, 1.127, 0.737, 1.409, 1.577, 0.868, 2.275, 3.035, 3.157, 3.199, 3.069, 2.878, 2.354, 1.577, 1.648, 2.235, 2.324, 1.804, 0.674, 0.466, 0.804, 0.839, 1.005, 1.348, 0.594, -0.227}
	a2 := []float64{776.247, 1303.151, -298.291, 184.811, 108.771, 61.953, -6.572, 10.505, 38.333, 41.731, -1.126, 4.629, -6.806, 2.944, 2.658, 0.261, -2.389, 2.284, -5.148, 3.011, 0.269, 0.152, 1.842, -2.474, 3.138, 2.443, -1.329, 0.831, -1.643, -0.856, -0.831, -0.449, -0.022, 2.086, -1.232, 0.22, -0.61, 1.282, -1.115, 0.406, 1.002, -0.242, 0.364, -0.323, 0.193, -0.384, -0.14, -0.637, 0.708, -0.121, 0.21, -0.729, -0.402, 0.194, 0.144, -0.109, 0.275, 0.068, -0.822, 0.001}
	a3 := []float64{409.16, -503.433, 1085.087, -25.346, -24.641, -29.414, 16.197, 3.018, -2.127, -37.939, 1.918, -3.812, 3.25, -0.096, -0.539, -0.883, 1.558, -2.477, 2.72, -0.914, -0.039, 0.563, -1.438, 1.871, -0.232, -1.257, 0.72, -0.825, 0.262, 0.008, 0.127, 0.142, 0.702, -1.106, 0.614, -0.277, 0.631, -0.799, 0.507, 0.199, -0.414, 0.202, -0.229, 0.172, -0.192, 0.081, -0.165, 0.448, -0.276, 0.11, -0.313, 0.109, 0.199, -0.017, -0.084, 0.128, -0.069, -0.297, 0.274, 0.086}

	n := len(y0)
	var i int
	for i = n - 1; i >= 0; i-- {
		if y >= y0[i] {
			break
		}
	}
	t := (y - y0[i]) / (y1[i] - y0[i])
	dT := a0[i] + t*(a1[i]+t*(a2[i]+t*a3[i]))
	return dT
}

func DeltaTv2(jd float64) float64 {
	if math.IsNaN(jd) || math.IsInf(jd, 0) {
		return math.NaN()
	}
	if jd > 2461041.5 || jd < 2441317.5 {
		var y float64
		if jd >= 2299160.5 {
			y = (jd-2451544.5)/365.2425 + 2000
		} else {
			y = (jd+0.5)/365.25 - 4712
		}
		return DeltaTSplineY(y)
	}

	// 闰秒JD值
	jdLeaps := []float64{2457754.5, 2457204.5, 2456109.5, 2454832.5,
		2453736.5, 2451179.5, 2450630.5, 2450083.5,
		2449534.5, 2449169.5, 2448804.5, 2448257.5,
		2447892.5, 2447161.5, 2446247.5, 2445516.5,
		2445151.5, 2444786.5, 2444239.5, 2443874.5,
		2443509.5, 2443144.5, 2442778.5, 2442413.5,
		2442048.5, 2441683.5, 2441499.5, 2441133.5}
	n := len(jdLeaps)
	deltaTSeconds := 42.184
	for i := 0; i < n; i++ {
		if jd >= jdLeaps[i] {
			deltaTSeconds += float64(n - i - 1)
			break
		}
	}
	return deltaTSeconds
}

// DeltaTSecondsAt 返回某个 TT 时刻实际使用的 ΔT（秒）：overrideSeconds 是有限值时直接采用
// （含 0，可显式要求 ΔT=0），为 NaN/±Inf 时改用进程级模型。模型按 UT 键控，因此这里先解
// TT−ΔT(UT) 再求值，不把 TT 直接当作 UT 送进模型（差约 2e-6 s）。
// DeltaTSecondsAt returns the ΔT in seconds used at one TT instant: a finite override wins
// (including 0, which requests ΔT = 0 explicitly), while NaN or ±Inf selects the process-wide
// model. The model is keyed by UT, so the equation TT - ΔT(UT) is solved instead of feeding TT.
func DeltaTSecondsAt(jdeTT, overrideSeconds float64) float64 {
	if !math.IsNaN(overrideSeconds) && !math.IsInf(overrideSeconds, 0) {
		return overrideSeconds
	}
	return deltaTModelSecondsAtTT(jdeTT)
}

func deltaTModelSecondsAtTT(jdeTT float64) float64 {
	ut := jdeTT - DeltaT(jdeTT, true)/86400.0
	for iteration := 0; iteration < 4; iteration++ {
		next := jdeTT - DeltaT(ut, true)/86400.0
		if next == ut {
			break
		}
		ut = next
	}
	return DeltaT(ut, true)
}

// DeltaTGroundShiftKM 把 ΔT 误差换算为站点相对影子的地面横移距离（千米）。
// 地球赤道自转线速度 465.1 m/s，因此 ΔT 相差 Δ 秒时，地面点相对影子横移
// 0.4651·|Δ|·cos(纬度) 千米；±400 年跨度上 ΔT 外推差几百到几千秒，足以挪动本影
// 边界数百千米，调用方可用本函数把外部给出的 ΔT 不确定度换算成几何不确定度。
// DeltaTGroundShiftKM converts a ΔT error into the ground displacement of a station
// relative to the shadow, in kilometres: 0.4651 * |ΔT| * cos(latitude).
func DeltaTGroundShiftKM(deltaTSeconds, latitudeDeg float64) float64 {
	if math.IsNaN(deltaTSeconds) || math.IsInf(deltaTSeconds, 0) ||
		math.IsNaN(latitudeDeg) || math.IsInf(latitudeDeg, 0) {
		return math.NaN()
	}
	return solarEclipseEarthEquatorialRotationKMPerSecond * math.Abs(deltaTSeconds) * math.Cos(latitudeDeg*rad)
}

func TD2UT(jde float64, utToTD bool) float64 { // true 世界时转力学时CC，false 力学时转世界时VV
	deltaTSeconds := DeltaT(jde, true)
	if utToTD {
		return jde + deltaTSeconds/3600/24
	}
	// Delta T is evaluated at UT in the forward conversion. Solve the same
	// equation in reverse so distant-epoch contact times survive a round trip.
	ut := jde - deltaTSeconds/3600/24
	for iteration := 0; iteration < 4; iteration++ {
		next := jde - DeltaT(ut, true)/3600/24
		if next == ut {
			break
		}
		ut = next
	}
	return ut
}

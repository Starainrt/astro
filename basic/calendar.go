package basic

import (
	"fmt"
	"math"
	"strings"
	"time"
)

var defDeltaTFn = DefaultDeltaTv2

/*
@name: 儒略日计算
@dec: 计算给定时间的儒略日，1582年改力后为格里高利历，之前为儒略历
@     请注意，传入的时间在天文计算中一般为力学时，应当注意和世界时的转化
*/
func JDECalc(Year, Month int, Day float64) float64 {
	if Month == 1 || Month == 2 {
		Year--
		Month += 12
	}
	var tmpvarB int
	tmpvar := fmt.Sprintf("%04d-%02d-%2d", Year, Month, int(math.Floor(Day)))
	if strings.Compare(tmpvar, `1582-10-04`) != 1 {
		tmpvarB = 0
	} else {
		tmpvarA := int(Year / 100)
		tmpvarB = 2 - tmpvarA + int(tmpvarA/4)
	}
	return (math.Floor(365.25*(float64(Year)+4716.0)) + math.Floor(30.6001*float64(Month+1)) + Day + float64(tmpvarB) - 1524.5)
}

/*
@name: 获得当前儒略日时间：当地世界时，非格林尼治时间
*/
func GetNowJDE() (NowJDE float64) {
	Time := float64(time.Now().Second())/3600.0/24.0 + float64(time.Now().Minute())/60.0/24.0 + float64(time.Now().Hour())/24.0
	NowJDE = JDECalc(time.Now().Year(), int(time.Now().Month()), float64(time.Now().Day())+Time)
	return
}

func dt_ext(y, jsd float64) float64 { // 二次曲线外推，用于数值外插
	dy := (y - 1820) / 100.00
	return -20 + jsd*dy*dy
}
func dt_cal(y float64) float64 { //传入年， 返回世界时UT与原子时（力学时 TD）之差, ΔT = TD - UT
	dt_at := []float64{
		-4000, 108371.7, -13036.80, 392.000, 0.0000,
		-500, 17201.0, -627.82, 16.170, -0.3413,
		-150, 12200.6, -346.41, 5.403, -0.1593,
		150, 9113.8, -328.13, -1.647, 0.0377,
		500, 5707.5, -391.41, 0.915, 0.3145,
		900, 2203.4, -283.45, 13.034, -0.1778,
		1300, 490.1, -57.35, 2.085, -0.0072,
		1600, 120.0, -9.81, -1.532, 0.1403,
		1700, 10.2, -0.91, 0.510, -0.0370,
		1800, 13.4, -0.72, 0.202, -0.0193,
		1830, 7.8, -1.81, 0.416, -0.0247,
		1860, 8.3, -0.13, -0.406, 0.0292,
		1880, -5.4, 0.32, -0.183, 0.0173,
		1900, -2.3, 2.06, 0.169, -0.0135,
		1920, 21.2, 1.69, -0.304, 0.0167,
		1940, 24.2, 1.22, -0.064, 0.0031,
		1960, 33.2, 0.51, 0.231, -0.0109,
		1980, 51.0, 1.29, -0.026, 0.0032,
		2000, 63.87, 0.1, 0, 0,
		2005, 64.7, 0.4, 0, 0, //一次项记为x,则 10x=0.4秒/年*(2015-2005),解得x=0.4
		2015, 69,
	}
	y0 := dt_at[len(dt_at)-2] //表中最后一年
	t0 := dt_at[len(dt_at)-1] //表中最后一年的 deltatT
	if y >= y0 {
		jsd := float64(31) // sjd是y1年之后的加速度估计
		// 瑞士星历表jsd=31, NASA网站jsd=32, skmap的jsd=29
		if y > y0+100.00 {
			return dt_ext(y, jsd)
		}
		v := dt_ext(y, jsd)        //二次曲线外推
		dv := dt_ext(y0, jsd) - t0 // ye年的二次外推与te的差
		return (v - dv*(y0+100.00-y)/100.00)
	}
	d := dt_at
	var i int
	for i = 0; i < len(d); i += 5 {
		if float64(y) < d[i+5] {
			break
			// 判断年所在的区间
		}
	}
	t1 := (y - d[i]) / (d[i+5] - d[i]) * 10.00 //////// 三次插值， 保证精确性
	t2 := t1 * t1
	t3 := t2 * t1
	res := d[i+1] + d[i+2]*t1 + d[i+3]*t2 + d[i+4]*t3
	return (res)
}

func DeltaT(date float64, isJDE bool) float64 {
	return defDeltaTFn(date, isJDE)
}

func SetDeltaTFn(fn func(float64, bool) float64) {
	if fn != nil {
		defDeltaTFn = fn
	}
}

func GetDeltaTFn() func(float64, bool) float64 {
	return defDeltaTFn
}

func OldDefaultDeltaT(Date float64, IsJDE bool) (Result float64) { //传入年或儒略日，传出为秒
	var Year float64
	if IsJDE {
		dates := JDE2Date(Date)
		Year = float64(dates.Year()) + float64(dates.YearDay())/365.0
	} else {
		Year = Date
	}
	if Year < 2100 && Year >= 2010 {
		return dt_cal(Year)
	}
	return DefaultDeltaT(Date, IsJDE)
}

func DefaultDeltaT(Date float64, IsJDE bool) (Result float64) { //传入年或儒略日，传出为秒
	var Year float64
	if IsJDE {
		dates := JDE2Date(Date)
		Year = float64(dates.Year()) + float64(dates.YearDay())/365.0
	} else {
		Year = Date
	}
	if Year < 2010 {
		Result = dt_cal(Year)
		return
	}
	if Year < 2100 && Year >= 2010 {
		var t = (Year - 2000.0)
		Result = 62.92 + 0.32217*t + 0.005589*t*t
		return
	}
	if Year >= 2100 && Year <= 2150 {
		Result = -20 + 32*(((Year-1820.0)/100.0)*((Year-1820.0)/100.0)) - 0.5628*(2150-Year)
		return
	}
	if Year > 2150 {
		//tmp=(Year-1820)/100;
		//Result= -20 + 32 * tmp*tmp;
		Result = dt_cal(Year)
		return
	}
	return
}

func DefaultDeltaTv2(date float64, isJd bool) float64 { //传入年或儒略日，传出为秒
	if !isJd {
		date = JDECalc(int(date), int((date-math.Floor(date))*12)+1, (date-math.Floor(date))*365.25+1)
	}
	return DeltaTv2(date)
}

// 使用Stephenson等人(2016)和Morrison等人(2021)的拟合和外推公式计算Delta T
// http://astro.ukho.gov.uk/nao/lvm/
// 2010年后的系数已修改以包含2019年后的数据
// 返回Delta T，单位为秒
func DeltaTSplineY(y float64) float64 {
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
	DT := 42.184
	for i := 0; i < n; i++ {
		if jd > jdLeaps[i] {
			DT += float64(n - i - 1)
			break
		}
	}
	return DT
}

func TD2UT(JDE float64, UT2TD bool) float64 { // true 世界时转力学时CC，false 力学时转世界时VV

	Deltat := DeltaT(JDE, true)
	if UT2TD {
		return JDE + Deltat/3600/24
	} else {
		return JDE - Deltat/3600/24
	}
}

/*
 * @name: JDE转日期，输出为数组
 */
func JDE2Date(JD float64) time.Time {
	JD = JD + 0.5
	Z := float64(int(JD))
	F := JD - Z
	var A, B, Years, Months, Days float64
	if Z < 2299161.0 {
		A = Z
	} else {
		alpha := math.Floor((Z - 1867216.25) / 36524.25)
		A = Z + 1 + alpha - math.Floor(alpha/4)
	}
	B = A + 1524
	C := math.Floor((B - 122.1) / 365.25)
	D := math.Floor(365.25 * C)
	E := math.Floor((B - D) / 30.6001)
	Days = B - D - math.Floor(30.6001*E) + F
	if E < 14 {
		Months = E - 1
	}
	if E == 14 || E == 15 {
		Months = E - 13
	}
	if Months > 2 {
		Years = C - 4716
	}
	if Months == 1 || Months == 2 {
		Years = C - 4715
	}
	tms := (Days - math.Floor(Days)) * 24 * 3600
	Days = math.Floor(Days)
	tz, _ := time.LoadLocation("Local")
	dates := time.Date(int(Years), time.Month(int(Months)), int(Days), 0, 0, 0, 0, tz)
	return time.Unix(dates.Unix()+int64(tms), int64((tms-math.Floor(tms))*1000000000))
}

// JDE2DateByZone JDE（儒略日）转日期
// JD: 儒略日
// tz: 目标时区
// byZone: (true: 传入的儒略日视为目标时区当地时间的儒略日，false: 传入的儒略日视为UTC时间的儒略日)
// 回参：转换后的日期，时区始终为目标时区
func JDE2DateByZone(JD float64, tz *time.Location, byZone bool) time.Time {
	JD = JD + 0.5
	Z := float64(int(JD))
	F := JD - Z
	var A, B, Years, Months, Days float64
	if Z < 2299161.0 {
		A = Z
	} else {
		alpha := math.Floor((Z - 1867216.25) / 36524.25)
		A = Z + 1 + alpha - math.Floor(alpha/4)
	}
	B = A + 1524
	C := math.Floor((B - 122.1) / 365.25)
	D := math.Floor(365.25 * C)
	E := math.Floor((B - D) / 30.6001)
	Days = B - D - math.Floor(30.6001*E) + F
	if E < 14 {
		Months = E - 1
	}
	if E == 14 || E == 15 {
		Months = E - 13
	}
	if Months > 2 {
		Years = C - 4716
	}
	if Months == 1 || Months == 2 {
		Years = C - 4715
	}
	tms := (Days - math.Floor(Days)) * 24 * 3600
	Days = math.Floor(Days)
	var transTz = tz
	if !byZone {
		transTz = time.UTC
	}
	return time.Date(int(Years), time.Month(int(Months)), int(Days), 0, 0, 0, 0, transTz).
		Add(time.Duration(int64(1000000000 * tms))).In(tz)
}

// Date2JDE 日期转儒略日
func Date2JDE(date time.Time) float64 {
	day := float64(date.Day()) + float64(date.Hour())/24.0 + float64(date.Minute())/24.0/60.0 + float64(date.Second())/24.0/3600.0 + float64(date.Nanosecond())/1000000000.0/3600.0/24.0
	return JDECalc(date.Year(), int(date.Month()), day)
}

func GetLunar(year, month, day int, tz float64) (lyear, lmonth, lday int, leap bool, result string) {
	julianDayEpoch := JDECalc(year, month, float64(day))
	// 确定农历年份
	lyear = year
	adjustedYear := year
	if month == 11 || month == 12 {
		winterSolsticeDay := GetJQTime(year, 270) + tz
		firstNewMoonDay := TD2UT(CalcMoonS(float64(year)+11.0/12.0+5.0/30.0/12.0, 0), true) + tz
		nextNewMoonDay := TD2UT(CalcMoonS(float64(year)+1.0, 0), true) + tz

		firstNewMoonDay = normalizeTimePoint(firstNewMoonDay)
		nextNewMoonDay = normalizeTimePoint(nextNewMoonDay)

		if winterSolsticeDay >= firstNewMoonDay && winterSolsticeDay < nextNewMoonDay && julianDayEpoch <= firstNewMoonDay {
			adjustedYear--
		}
		if winterSolsticeDay >= nextNewMoonDay && julianDayEpoch < nextNewMoonDay {
			adjustedYear--
		}
	} else {
		adjustedYear--
	}

	// 获取节气和朔望月数据
	solarTerms := GetJieqiLoops(adjustedYear, 25)
	newMoonDays := GetMoonLoops(float64(adjustedYear), 17)

	// 计算冬至日期
	winterSolsticeFirst := solarTerms[0] - 8.0/24 + tz
	winterSolsticeSecond := solarTerms[24] - 8.0/24 + tz

	// 规范化时间点
	normalizeTimeArray(newMoonDays, tz)
	normalizeTimeArray(solarTerms, tz)

	// 计算朔望月范围
	minMoonIndex, maxMoonIndex := 20, 0
	moonCount := 0
	for i := 0; i < 15; i++ {
		if newMoonDays[i] >= winterSolsticeFirst && newMoonDays[i] < winterSolsticeSecond {
			if i <= minMoonIndex {
				minMoonIndex = i
			}
			if i >= maxMoonIndex {
				maxMoonIndex = i
			}
			moonCount++
		}
	}

	// 确定闰月位置
	leapMonthPos := 20
	if moonCount == 13 {
		solarTermIndex, i := 2, 0
		for i = minMoonIndex; i <= maxMoonIndex; i++ {
			if !(newMoonDays[i] <= solarTerms[solarTermIndex] && newMoonDays[i+1] > solarTerms[solarTermIndex]) {
				break
			}
			solarTermIndex += 2
		}
		leapMonthPos = i - minMoonIndex + 1
	}

	// 找到当前月相索引
	currentMoonIndex := 0
	for currentMoonIndex = minMoonIndex - 1; currentMoonIndex <= maxMoonIndex; currentMoonIndex++ {
		if newMoonDays[currentMoonIndex] > julianDayEpoch {
			break
		}
	}

	// 计算农历月份
	lmonth = currentMoonIndex - minMoonIndex
	shouldAdjustLeap := false
	leap = false

	if lmonth >= leapMonthPos {
		shouldAdjustLeap = true
	}
	if lmonth == leapMonthPos {
		leap = true
	}
	if lmonth < 2 {
		lmonth += 11
	} else {
		lmonth--
	}
	if shouldAdjustLeap {
		lmonth--
	}
	if lmonth <= 0 {
		lmonth += 12
	}

	// 计算农历日期
	lday = int(julianDayEpoch-newMoonDays[currentMoonIndex-1]) + 1

	// 生成农历日期字符串
	result = formatLunarDateString(lmonth, lday, leap)
	if lmonth >= 10 && month < 3 {
		lyear--
	}
	return
}

func GetSolar(year, month, day int, leap bool, tz float64) float64 {
	adjustedYear := year
	if month < 11 {
		adjustedYear--
	}

	// 获取节气和朔望月数据
	solarTerms := GetJieqiLoops(adjustedYear, 25)
	newMoonDays := GetMoonLoops(float64(adjustedYear), 17)

	// 计算冬至日期
	winterSolsticeFirst := solarTerms[0] - 8.0/24 + tz
	winterSolsticeSecond := solarTerms[24] - 8.0/24 + tz

	// 规范化时间点
	normalizeTimeArray(newMoonDays, tz)
	normalizeTimeArray(solarTerms, tz)

	// 计算朔望月范围
	minMoonIndex, maxMoonIndex := 20, 0
	moonCount := 0
	for i := 0; i < 15; i++ {
		if newMoonDays[i] >= winterSolsticeFirst && newMoonDays[i] < winterSolsticeSecond {
			if i <= minMoonIndex {
				minMoonIndex = i
			}
			if i >= maxMoonIndex {
				maxMoonIndex = i
			}
			moonCount++
		}
	}

	// 确定闰月位置
	leapMonthPos := 20
	if moonCount == 13 {
		solarTermIndex, i := 2, 0
		for i = minMoonIndex; i <= maxMoonIndex; i++ {
			if !(newMoonDays[i] <= solarTerms[solarTermIndex] && newMoonDays[i+1] > solarTerms[solarTermIndex]) {
				break
			}
			solarTermIndex += 2
		}
		leapMonthPos = i - minMoonIndex + 1
	}

	// 计算实际月份索引
	actualMonth := month
	if leap {
		actualMonth++
	}
	if actualMonth > 10 {
		actualMonth -= 11
	} else {
		actualMonth++
	}
	if actualMonth >= leapMonthPos && !leap {
		actualMonth++
	}

	return newMoonDays[minMoonIndex-1+actualMonth] + float64(day) - 1
}

func normalizeTimeArray(timeArray []float64, tz float64) {
	for idx, timeValue := range timeArray {
		adjustedTime := timeValue
		if tz != 8.0/24 {
			adjustedTime = timeValue - 8.0/24 + tz
		}
		timeArray[idx] = normalizeTimePoint(adjustedTime)
	}
}

func normalizeTimePoint(timePoint float64) float64 {
	if timePoint-math.Floor(timePoint) > 0.5 {
		return math.Floor(timePoint) + 0.5
	}
	return math.Floor(timePoint) - 0.5
}

func formatLunarDateString(lunarMonth, lunarDay int, isLeap bool) string {
	monthNames := []string{"十", "一", "二", "三", "四", "五", "六", "七", "八", "九", "十", "冬", "腊"}
	dayPrefixes := []string{"初", "十", "廿", "三"}

	var dateString string

	if isLeap {
		dateString += "闰"
	}

	if lunarMonth == 1 {
		dateString += "正月"
	} else {
		dateString += monthNames[lunarMonth] + "月"
	}

	if lunarDay == 20 {
		dateString += "二十"
	} else if lunarDay == 10 {
		dateString += "初十"
	} else {
		dateString += dayPrefixes[lunarDay/10] + monthNames[lunarDay%10]
	}

	return dateString
}

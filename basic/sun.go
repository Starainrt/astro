package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

// SunLo 太阳几何黄经
func SunLo(jd float64) float64 {
	T := (jd - 2451545) / 365250
	SunLo := 280.4664567 + 360007.6982779*T + 0.03032028*T*T + T*T*T/49931 - T*T*T*T/15299 - T*T*T*T*T/1988000
	return Limit360(SunLo)
}

func SunM(JD float64) float64 {
	T := (JD - 2451545) / 36525
	sunM := 357.5291092 + 35999.0502909*T - 0.0001559*T*T - 0.00000048*T*T*T
	return Limit360(sunM)
}

/*
@name 地球偏心率
*/
func Earthe(JD float64) float64 { //'地球偏心率
	T := (JD - 2451545) / 36525
	Earthe := 0.016708617 - 0.000042037*T - 0.0000001236*T*T
	return Earthe
}

func EarthPI(JD float64) float64 { //近日點經度
	T := (JD - 2451545) / 36525
	return 102.93735 + 1.71953*T + 000046*T*T
}

func SunMidFun(JD float64) float64 { //'太阳中间方程
	T := (JD - 2451545) / 36525
	M := SunM(JD)
	SunMidFun := (1.9146-0.004817*T-0.000014*T*T)*Sin(M) + (0.019993-0.000101*T)*Sin(2*M) + 0.00029*Sin(3*M)
	return SunMidFun
}

func SunTrueLo(JD float64) float64 { // '太阳真黄经
	SunTrueLo := SunLo(JD) + SunMidFun(JD)
	return SunTrueLo
}

func SunApparentLo(JD float64) float64 { //'太阳视黄经

	T := (JD - 2451545) / 36525
	SunApparentLo := SunTrueLo(JD) - 0.00569 - 0.00478*Sin(125.04-1934.136*T)
	return SunApparentLo
}

func SunApparentRa(JD float64) float64 { // '太阳视赤经
	return LoToRa(JD, SunApparentLo(JD), 0)
}

func SunApparentRaDec(JD float64) (float64, float64) {
	return LoBoToRaDec(JD, SunApparentLo(JD), 0)
}

func SunTrueRa(JD float64) float64 { //'太阳真赤经
	sitas := Sita(JD)
	sunTrueRa := ArcTan(Cos(sitas) * Sin(SunTrueLo(JD)) / Cos(SunTrueLo(JD)))
	//Select Case SunTrueLo(JD)
	tmp := SunTrueLo(JD)
	if tmp >= 90 && tmp < 180 {
		sunTrueRa = 180 + sunTrueRa
	} else if tmp >= 180 && tmp < 270 {
		sunTrueRa = 180 + sunTrueRa
	} else if tmp >= 270 && tmp <= 360 {
		sunTrueRa = 360 + sunTrueRa
	}
	return sunTrueRa
}

func SunApparentDec(JD float64) float64 { // '太阳视赤纬
	T := (JD - 2451545) / 36525
	sitas := Sita(JD) + 0.00256*Cos(125.04-1934.136*T)
	SunApparentDec := ArcSin(Sin(sitas) * Sin(SunApparentLo(JD)))
	return SunApparentDec
}

func SunTrueDec(JD float64) float64 { // '太阳真赤纬
	sitas := Sita(JD)
	SunTrueDec := ArcSin(Sin(sitas) * Sin(SunTrueLo(JD)))
	return SunTrueDec
}

func SunTime(JD float64) float64 { //均时差
	tm := (SunLo(JD) - 0.0057183 - (HSunApparentRa(JD)) + (Nutation2000Bi(JD))*Cos(Sita(JD))) / 15
	if tm > 23 {
		tm = -24 + tm
	}
	return tm
}

func SunSC(Lo, JD float64) float64 { //黄道上的岁差，仅黄纬=0时
	t := (JD - 2451545) / 36525
	//n := 47.0029/3600*t - 0.03302/3600*t*t + 0.000060/3600*t*t*t
	//m := 174.876384/3600 - 869.8089/3600*t + 0.03536/3600*t*t
	pk := 5029.0966/3600.00*t + 1.11113/3600.00*t*t - 0.000006/3600.00*t*t*t
	return Lo + pk
}

// 高精度，使用VSOP87
func HSunTrueLo(JD float64) float64 {
	L := planet.WherePlanet(0, 0, JD)
	return L
}

func HSunTrueBo(JD float64) float64 {
	L := planet.WherePlanet(0, 1, JD)
	return L
}

func HSunApparentLo(JD float64) float64 {
	L := HSunTrueLo(JD)
	L = L + Nutation2000Bi(JD) + SunLoGXC(JD)
	return L
}

func SunLoGXC(JD float64) float64 {
	R := planet.WherePlanet(0, 2, JD)
	return -20.49552 / R / 3600
}

func EarthAway(JD float64) float64 {
	//t=(JD - 2451545) / 365250;
	//R=Earth_R5(t)+Earth_R4(t)+Earth_R3(t)+Earth_R2(t)+Earth_R1(t)+Earth_R0(t);
	return planet.WherePlanet(0, 2, JD)
}

func HSunApparentRaDec(JD float64) (float64, float64) {
	return LoBoToRaDec(JD, HSunApparentLo(JD), HSunTrueBo(JD))
}

func HSunApparentRa(JD float64) float64 { // '太阳视赤经
	return LoToRa(JD, HSunApparentLo(JD), HSunTrueBo(JD))
}

func HSunTrueRa(JD float64) float64 {
	tmp := HSunTrueLo(JD)
	sitas := Sita(JD)

	numerator := Cos(sitas) * Sin(tmp)
	denominator := Cos(tmp)

	return ArcTan2(numerator, denominator)
}

func HSunApparentDec(JD float64) float64 { // '太阳视赤纬
	return ArcSin(Sin(EclipticObliquity(JD, true)) * Sin(HSunApparentLo(JD)))
}

func HSunTrueDec(JD float64) float64 { // '太阳真赤纬
	return ArcSin(Sin(EclipticObliquity(JD, false)) * Sin(HSunTrueLo(JD)))
}

func Distance(jd float64) float64 { //ri di ju li
	f := SunMidFun(jd)
	m := SunM(jd)
	e := Earthe(jd)
	return (1.000001018 * (1 - e*e) / (1 + e*Cos(f+m)))
}

func GetMoonLoops(year float64, loop int) []float64 {
	var start float64
	var newMoon, tmp float64
	moon := make([]float64, loop)
	if year < 6000 {
		start = year + 11.00/12.00 + 5.00/30.00/12.00
	} else {
		start = year + 9.00/12.00 + 5.00/30.00/12.00
	}
	i := 1
	for j := 0; j < loop; j++ {
		if year > 3000 {
			newMoon = TD2UT(CalcMoonSH(start+float64(i-1)/12.5, 0)+8.0/24.0, false)
		} else {
			newMoon = TD2UT(CalcMoonS(start+float64(i-1)/12.5, 0)+8.0/24.0, false)
		}
		if i != 1 {
			if newMoon == tmp {
				j--
				i++
				continue
			}
		}
		moon[j] = newMoon
		tmp = moon[j]
		i++
	}
	return moon
}

func GetJieqiLoops(year, loop int) []float64 {
	start := 270
	jq := make([]float64, loop)
	for i := 1; i <= loop; i++ {
		angle := start + 15*(i-1)
		if angle > 360 {
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
	initialJD := JDECalc(year, int(initialMonth), initialDay)

	// Set target angle for iteration; if angle is 0, use 360
	targetAngle := float64(angle)
	if angle == 0 {
		targetAngle = 360.0
	}

	// Newton-Raphson iteration to find precise Julian date
	currentJD := initialJD
	for {
		previousJD := currentJD
		errorValue := JQLospec(previousJD, targetAngle) - targetAngle
		derivative := (JQLospec(previousJD+0.000005, targetAngle) - JQLospec(previousJD-0.000005, targetAngle)) / 0.00001
		currentJD = previousJD - errorValue/derivative

		// Check for convergence
		if math.Abs(currentJD-previousJD) <= 0.00001 {
			break
		}
	}

	// Convert to UT and return
	return TD2UT(currentJD, false)
}

func JQLospec(JD float64, target float64) float64 {
	t := HSunApparentLo(JD)
	if target >= 345 {
		if t <= 12 {
			t += 360
		}
	} else if target <= 15 {
		if t >= 350 {
			t -= 360
		}
	}
	return t
}

func GetXC(jd float64) string { //十二次
	tlo := HSunApparentLo(jd)
	if tlo >= 255 && tlo < 285 {
		return "星纪"
	} else if tlo >= 285 && tlo < 315 {
		return "玄枵"
	} else if tlo >= 315 && tlo < 345 {
		return "娵訾"
	} else if tlo >= 345 || tlo < 15 {
		return "降娄"
	} else if tlo >= 15 && tlo < 45 {
		return "大梁"
	} else if tlo >= 45 && tlo < 75 {
		return "实沈"
	} else if tlo >= 75 && tlo < 105 {
		return "鹑首"
	} else if tlo >= 105 && tlo < 135 {
		return "鹑火"
	} else if tlo >= 135 && tlo < 165 {
		return "鹑尾"
	} else if tlo >= 165 && tlo < 195 {
		return "寿星"
	} else if tlo >= 195 && tlo < 225 {
		return "大火"
	} else if tlo >= 225 && tlo < 255 {
		return "析木"
	}
	return ""
}

func GetWHTime(Year, Angle int) float64 {
	tmp := Angle
	var Day int
	var tp float64
	Angle = int(Angle/15) * 15
	if Angle%2 == 0 {
		Day = 18
	} else {
		Day = 3
	}
	if Angle%10 != 0 {
		tp = float64(Angle+15) / 30.0
	} else {
		tp = float64(Angle) / 30.0
	}
	Month := int(3 + tp)
	if Month > 12 {
		Month -= 12
	}
	JD1 := JDECalc(Year, Month, float64(Day))
	JD1 += float64(tmp - Angle)
	Angle = tmp
	if Angle <= 5 {
		Angle = 360 + Angle
	}
	for {
		JD0 := JD1
		stDegree := JQLospec(JD0, float64(Angle)) - float64(Angle)
		stDegreep := (JQLospec(JD0+0.000005, float64(Angle)) - JQLospec(JD0-0.000005, float64(Angle))) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) <= 0.00001 {
			break
		}
	}
	return TD2UT(JD1, false)
}

// 太阳中天时刻，通过均时差计算
func CulminationTime(JD, Lon, TZ float64) float64 { //实际中天时间
	JD = math.Floor(JD)
	tmp := (TZ*15 - Lon) * 4 / 60
	return JD + tmp/24.0 - SunTime(JD)/24.0
}

/*
 * 昏朦影传入 当天0时时刻
 */
func EveningTwilight(JD, Lon, Lat, TZ, An float64) float64 {
	JD = math.Floor(JD) + 1.5
	ntz := math.Round(Lon / 15)
	culminationTime := CulminationTime(JD, Lon, ntz)
	if SunHeight(culminationTime, Lon, Lat, ntz) < An {
		return -2 //极夜
	}
	if SunHeight(culminationTime+0.5, Lon, Lat, ntz) > An {
		return -1 //极昼
	}
	tmp := (Sin(An) - Sin(HSunApparentDec(culminationTime))*Sin(Lat)) / (Cos(HSunApparentDec(culminationTime)) * Cos(Lat))
	var sundown float64
	if math.Abs(tmp) <= 1 && Lat < 85 {
		rzsc := ArcCos(tmp) / 15
		sundown = culminationTime + rzsc/24.0 + 35.0/24.0/60.0
	} else {
		sundown = culminationTime
		i := 0
		for LowSunHeight(sundown, Lon, Lat, ntz) > An {
			i++
			sundown += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}
	JD1 := sundown - 5.00/24.00/60.00
	for {
		JD0 := JD1
		stDegree := SunHeight(JD0, Lon, Lat, ntz) - An
		stDegreep := (SunHeight(JD0+0.000005, Lon, Lat, ntz) - SunHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - stDegree/stDegreep
		if math.Abs(JD1-JD0) < 0.00001 {
			break
		}
	}
	return JD1 - ntz/24 + TZ/24
}

func MorningTwilight(JD, Lon, Lat, TZ, An float64) float64 {
	// 调整到中午12点
	JD = math.Floor(JD) + 1.5

	// 计算时区
	ntz := math.Round(Lon / 15)

	// 计算太阳上中天时间
	culminationTime := CulminationTime(JD, Lon, ntz)

	// 检查极夜和极昼条件
	if SunHeight(culminationTime, Lon, Lat, ntz) < An {
		return -2 // 极夜
	}
	if SunHeight(culminationTime-0.5, Lon, Lat, ntz) > An {
		return -1 // 极昼
	}

	// 计算日出时间
	sunDec := HSunApparentDec(culminationTime)
	tmp := (Sin(An) - Sin(sunDec)*Sin(Lat)) / (Cos(sunDec) * Cos(Lat))

	var sunrise float64
	if math.Abs(tmp) <= 1 && Lat < 85 {
		hourAngle := ArcCos(tmp) / 15
		sunrise = culminationTime - hourAngle/24 - 25.0/(24.0*60.0)
	} else {
		sunrise = culminationTime
		for i := 0; i < 48 && LowSunHeight(sunrise, Lon, Lat, ntz) > An; i++ {
			sunrise -= 15.0 / (60.0 * 24.0) // 每次减少15分钟
		}
	}

	JD1 := sunrise - 5.0/(24.0*60.0)
	for {
		JD0 := JD1
		heightDiff := SunHeight(JD0, Lon, Lat, ntz) - An
		heightDerivative := (SunHeight(JD0+0.000005, Lon, Lat, ntz) - SunHeight(JD0-0.000005, Lon, Lat, ntz)) / 0.00001
		JD1 = JD0 - heightDiff/heightDerivative

		if math.Abs(JD1-JD0) < 0.00001 {
			break
		}
	}

	return JD1 - ntz/24 + TZ/24
}

/*
 * 太阳时角
 */
func SunTimeAngle(JD, Lon, Lat, TZ float64) float64 {
	startime := Limit360(ApparentSiderealTime(JD-TZ/24)*15 + Lon)
	timeangle := startime - HSunApparentRa(TD2UT(JD-TZ/24, true))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

// GetSunRiseTime 精确计算日出时间，传入当日0时JDE
func GetSunRiseTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) float64 {
	return calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height, true)
}

// GetSunSetTime 精确计算日落时间，传入当日0时JDE
func GetSunSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) float64 {
	return calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height, false)
}

// calculateSunRiseSetTime 统一的日出日落计算函数
func calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64, isSunrise bool) float64 {
	var sunAngle float64
	julianDay = math.Floor(julianDay) + 1.5
	naturalTimeZone := math.Round(longitude / 15)

	// 计算太阳高度角
	if zenithShift != 0 {
		sunAngle = -0.8333
	}
	sunAngle = sunAngle - HeightDegreeByLat(height, latitude)

	// 获取太阳上中天时间
	solarNoonTime := CulminationTime(julianDay, longitude, naturalTimeZone)

	// 检查极夜极昼条件
	polarCondition := checkPolarConditions(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, isSunrise)
	if polarCondition != 0 {
		return polarCondition
	}

	// 计算初始估算时间
	initialTime := calculateInitialSunTime(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, isSunrise)

	// 牛顿-拉夫逊迭代求精确解
	return sunRiseSetNewtonRaphsonIteration(initialTime, longitude, latitude, naturalTimeZone, sunAngle, timeZone)
}

// checkPolarConditions 检查极夜极昼条件
func checkPolarConditions(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle float64, isSunrise bool) float64 {
	if SunHeight(solarNoonTime, longitude, latitude, naturalTimeZone) < sunAngle {
		return -2 // 极夜
	}

	checkTime := solarNoonTime + 0.5
	if isSunrise {
		checkTime = solarNoonTime - 0.5
	}

	if SunHeight(checkTime, longitude, latitude, naturalTimeZone) > sunAngle {
		return -1 // 极昼
	}

	return 0 // 正常条件
}

// calculateInitialSunTime 计算日出日落的初始估算时间
func calculateInitialSunTime(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle float64, isSunrise bool) float64 {
	// 使用球面三角法计算: (sin(ho)-sin(φ)*sin(δ))/(cos(φ)*cos(δ))
	apparentDeclination := HSunApparentDec(solarNoonTime)
	cosHourAngle := (Sin(sunAngle) - Sin(apparentDeclination)*Sin(latitude)) / (Cos(apparentDeclination) * Cos(latitude))

	if math.Abs(cosHourAngle) <= 1 && latitude < 85 {
		// 使用解析解
		hourAngle := ArcCos(cosHourAngle) / 15
		timeOffset := 25.0 / 24.0 / 60.0 // 日出偏移
		if !isSunrise {
			timeOffset = 35.0 / 24.0 / 60.0 // 日落偏移
		}

		if isSunrise {
			return solarNoonTime - hourAngle/24 - timeOffset
		} else {
			return solarNoonTime + hourAngle/24 + timeOffset
		}
	} else {
		// 使用迭代逼近法（极地条件）
		return iterativeApproach(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, isSunrise)
	}
}

// iterativeApproach 迭代逼近法计算（用于极地等特殊条件）
func iterativeApproach(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle float64, isSunrise bool) float64 {
	estimatedTime := solarNoonTime
	stepSize := 15.0 / 60.0 / 24.0 // 15分钟步长
	if isSunrise {
		stepSize = -stepSize
	}

	const maxIterations = 48
	for i := 0; i < maxIterations && LowSunHeight(estimatedTime, longitude, latitude, naturalTimeZone) > sunAngle; i++ {
		estimatedTime += stepSize
	}

	return estimatedTime
}

// sunRiseSetNewtonRaphsonIteration 牛顿-拉夫逊迭代法求精确解
func sunRiseSetNewtonRaphsonIteration(initialTime, longitude, latitude, naturalTimeZone, sunAngle, timeZone float64) float64 {
	const (
		convergenceThreshold = 0.00001
		derivativeStep       = 0.000005
	)

	currentTime := initialTime

	for {
		previousTime := currentTime

		// 计算函数值：f(t) = SunHeight(t) - targetAngle
		functionValue := SunHeight(previousTime, longitude, latitude, naturalTimeZone) - sunAngle

		// 计算导数：f'(t) ≈ (f(t+h) - f(t-h)) / (2h)
		derivative := (SunHeight(previousTime+derivativeStep, longitude, latitude, naturalTimeZone) -
			SunHeight(previousTime-derivativeStep, longitude, latitude, naturalTimeZone)) / (2 * derivativeStep)

		// 牛顿-拉夫逊公式：t_new = t_old - f(t) / f'(t)
		currentTime = previousTime - functionValue/derivative

		// 检查收敛
		if math.Abs(currentTime-previousTime) <= convergenceThreshold {
			break
		}
	}

	// 转换为指定时区
	return currentTime - naturalTimeZone/24 + timeZone/24
}

/*
 * 太阳高度角 世界时
 */
func SunHeight(JD, Lon, Lat, TZ float64) float64 {
	//tmp := (TZ*15 - Lon) * 4 / 60
	//truejd := JD - tmp/24
	calcjd := JD - TZ/24.0
	tjde := TD2UT(calcjd, true)
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	ra, dec := HSunApparentRaDec(tjde)
	H := Limit360(st - ra)
	tmp2 := Sin(Lat)*Sin(dec) + Cos(dec)*Cos(Lat)*Cos(H)
	return ArcSin(tmp2)
}
func LowSunHeight(JD, Lon, Lat, TZ float64) float64 {
	//tmp := (TZ*15 - Lon) * 4 / 60
	//truejd := JD - tmp/24
	calcjd := JD - TZ/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	H := Limit360(st - SunApparentRa(TD2UT(calcjd, true)))
	dec := SunApparentDec(TD2UT(calcjd, true))
	tmp2 := Sin(Lat)*Sin(dec) + Cos(dec)*Cos(Lat)*Cos(H)
	return ArcSin(tmp2)
}
func SunAngle(JD, Lon, Lat, TZ float64) float64 {
	//tmp := (TZ*15 - Lon) * 4 / 60
	//truejd := JD - tmp/24
	calcjd := JD - TZ/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + Lon)
	H := Limit360(st - HSunApparentRa(TD2UT(calcjd, true)))
	tmp2 := Sin(H) / (Cos(H)*Sin(Lat) - Tan(HSunApparentDec(TD2UT(calcjd, true)))*Cos(Lat))
	Angle := ArcTan(tmp2)
	if Angle < 0 {
		if H/15 < 12 {
			return Angle + 360
		}
		return Angle + 180
	}
	if H/15 < 12 {
		return Angle + 180
	}
	return Angle
}

/*
* 干支
 */
func GetGZ(year int) string {
	tiangan := []string{"庚", "辛", "壬", "癸", "甲", "乙", "丙", "丁", "戊", "己"}
	dizhi := []string{"申", "酉", "戌", "亥", "子", "丑", "寅", "卯", "辰", "巳", "午", "未"}
	t := year - (year / 10 * 10)
	if t < 0 {
		t += 10
	}
	d := year % 12
	if d < 0 {
		d += 12
	}
	return tiangan[t] + dizhi[d]
}

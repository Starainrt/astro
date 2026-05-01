package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// 太阳中天时刻，通过均时差计算
func CulminationTime(jd, lon, tz float64) float64 { //实际中天时间
	jd = math.Floor(jd)
	tmp := (tz*15 - lon) * 4 / 60
	return jd + tmp/24.0 - SunTime(jd)/24.0
}

func CulminationTimeN(jd, lon, tz float64, n int) float64 { //实际中天时间
	jd = math.Floor(jd)
	tmp := (tz*15 - lon) * 4 / 60
	return jd + tmp/24.0 - SunTimeN(jd, n)/24.0
}

/*
 * 昏朦影传入 当天0时时刻
 */
func EveningTwilight(jd, lon, lat, tz, targetAltitude float64) (float64, error) {
	jd = math.Floor(jd) + 1.5
	localTimeZone := math.Round(lon / 15)
	culminationTime := CulminationTime(jd, lon, localTimeZone)
	if SunHeight(culminationTime, lon, lat, localTimeZone) < targetAltitude {
		return 0, ErrNeverRise
	}
	if SunHeight(culminationTime+0.5, lon, lat, localTimeZone) > targetAltitude {
		return 0, ErrNeverSet
	}
	tmp := (Sin(targetAltitude) - Sin(HSunApparentDec(culminationTime))*Sin(lat)) / (Cos(HSunApparentDec(culminationTime)) * Cos(lat))
	var sundown float64
	if math.Abs(tmp) <= 1 && lat < 85 {
		hourOffset := ArcCos(tmp) / 15
		sundown = culminationTime + hourOffset/24.0 + 35.0/24.0/60.0
	} else {
		sundown = culminationTime
		i := 0
		for LowSunHeight(sundown, lon, lat, localTimeZone) > targetAltitude {
			i++
			sundown += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}
	estimateJD := sundown - 5.00/24.00/60.00
	for {
		prevJD := estimateJD
		stDegree := SunHeight(prevJD, lon, lat, localTimeZone) - targetAltitude
		stDegreep := (SunHeight(prevJD+0.000005, lon, lat, localTimeZone) - SunHeight(prevJD-0.000005, lon, lat, localTimeZone)) / 0.00001
		estimateJD = prevJD - stDegree/stDegreep
		if math.Abs(estimateJD-prevJD) < 0.00001 {
			break
		}
	}
	return estimateJD - localTimeZone/24 + tz/24, nil
}

func EveningTwilightN(jd, lon, lat, tz, targetAltitude float64, n int) (float64, error) {
	jd = math.Floor(jd) + 1.5
	localTimeZone := math.Round(lon / 15)
	culminationTime := CulminationTimeN(jd, lon, localTimeZone, n)
	if SunHeightN(culminationTime, lon, lat, localTimeZone, n) < targetAltitude {
		return 0, ErrNeverRise
	}
	if SunHeightN(culminationTime+0.5, lon, lat, localTimeZone, n) > targetAltitude {
		return 0, ErrNeverSet
	}
	tmp := (Sin(targetAltitude) - Sin(HSunApparentDecN(culminationTime, n))*Sin(lat)) / (Cos(HSunApparentDecN(culminationTime, n)) * Cos(lat))
	var sundown float64
	if math.Abs(tmp) <= 1 && lat < 85 {
		hourOffset := ArcCos(tmp) / 15
		sundown = culminationTime + hourOffset/24.0 + 35.0/24.0/60.0
	} else {
		sundown = culminationTime
		i := 0
		for lowSunHeightForN(sundown, lon, lat, localTimeZone, n) > targetAltitude {
			i++
			sundown += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}
	estimateJD := sundown - 5.00/24.00/60.00
	for {
		prevJD := estimateJD
		stDegree := SunHeightN(prevJD, lon, lat, localTimeZone, n) - targetAltitude
		stDegreep := (SunHeightN(prevJD+0.000005, lon, lat, localTimeZone, n) - SunHeightN(prevJD-0.000005, lon, lat, localTimeZone, n)) / 0.00001
		estimateJD = prevJD - stDegree/stDegreep
		if math.Abs(estimateJD-prevJD) < 0.00001 {
			break
		}
	}
	return estimateJD - localTimeZone/24 + tz/24, nil
}

func MorningTwilight(jd, lon, lat, tz, targetAltitude float64) (float64, error) {
	// 调整到中午12点
	jd = math.Floor(jd) + 1.5

	// 计算时区
	localTimeZone := math.Round(lon / 15)

	// 计算太阳上中天时间
	culminationTime := CulminationTime(jd, lon, localTimeZone)

	// 检查极夜和极昼条件
	if SunHeight(culminationTime, lon, lat, localTimeZone) < targetAltitude {
		return 0, ErrNeverRise
	}
	if SunHeight(culminationTime-0.5, lon, lat, localTimeZone) > targetAltitude {
		return 0, ErrNeverSet
	}

	// 计算日出时间
	sunDec := HSunApparentDec(culminationTime)
	tmp := (Sin(targetAltitude) - Sin(sunDec)*Sin(lat)) / (Cos(sunDec) * Cos(lat))

	var sunrise float64
	if math.Abs(tmp) <= 1 && lat < 85 {
		hourAngle := ArcCos(tmp) / 15
		sunrise = culminationTime - hourAngle/24 - 25.0/(24.0*60.0)
	} else {
		sunrise = culminationTime
		for i := 0; i < 48 && LowSunHeight(sunrise, lon, lat, localTimeZone) > targetAltitude; i++ {
			sunrise -= 15.0 / (60.0 * 24.0) // 每次减少15分钟
		}
	}

	estimateJD := sunrise - 5.0/(24.0*60.0)
	for {
		prevJD := estimateJD
		heightDiff := SunHeight(prevJD, lon, lat, localTimeZone) - targetAltitude
		heightDerivative := (SunHeight(prevJD+0.000005, lon, lat, localTimeZone) - SunHeight(prevJD-0.000005, lon, lat, localTimeZone)) / 0.00001
		estimateJD = prevJD - heightDiff/heightDerivative

		if math.Abs(estimateJD-prevJD) < 0.00001 {
			break
		}
	}

	return estimateJD - localTimeZone/24 + tz/24, nil
}

func MorningTwilightN(jd, lon, lat, tz, targetAltitude float64, n int) (float64, error) {
	jd = math.Floor(jd) + 1.5
	localTimeZone := math.Round(lon / 15)
	culminationTime := CulminationTimeN(jd, lon, localTimeZone, n)
	if SunHeightN(culminationTime, lon, lat, localTimeZone, n) < targetAltitude {
		return 0, ErrNeverRise
	}
	if SunHeightN(culminationTime-0.5, lon, lat, localTimeZone, n) > targetAltitude {
		return 0, ErrNeverSet
	}

	sunDec := HSunApparentDecN(culminationTime, n)
	tmp := (Sin(targetAltitude) - Sin(sunDec)*Sin(lat)) / (Cos(sunDec) * Cos(lat))

	var sunrise float64
	if math.Abs(tmp) <= 1 && lat < 85 {
		hourAngle := ArcCos(tmp) / 15
		sunrise = culminationTime - hourAngle/24 - 25.0/(24.0*60.0)
	} else {
		sunrise = culminationTime
		for i := 0; i < 48 && lowSunHeightForN(sunrise, lon, lat, localTimeZone, n) > targetAltitude; i++ {
			sunrise -= 15.0 / (60.0 * 24.0)
		}
	}

	estimateJD := sunrise - 5.0/(24.0*60.0)
	for {
		prevJD := estimateJD
		heightDiff := SunHeightN(prevJD, lon, lat, localTimeZone, n) - targetAltitude
		heightDerivative := (SunHeightN(prevJD+0.000005, lon, lat, localTimeZone, n) - SunHeightN(prevJD-0.000005, lon, lat, localTimeZone, n)) / 0.00001
		estimateJD = prevJD - heightDiff/heightDerivative

		if math.Abs(estimateJD-prevJD) < 0.00001 {
			break
		}
	}

	return estimateJD - localTimeZone/24 + tz/24, nil
}

/*
 * 太阳时角
 */
func SunTimeAngle(jd, lon, lat, tz float64) float64 {
	startime := Limit360(ApparentSiderealTime(jd-tz/24)*15 + lon)
	timeangle := startime - HSunApparentRa(TD2UT(jd-tz/24, true))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

func SunTimeAngleN(jd, lon, lat, tz float64, n int) float64 {
	startime := Limit360(ApparentSiderealTime(jd-tz/24)*15 + lon)
	timeangle := startime - HSunApparentRaN(TD2UT(jd-tz/24, true), n)
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

// GetSunRiseTime 精确计算日出时间，传入当日0时JDE
func GetSunRiseTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	return calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height, true)
}

func GetSunRiseTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height float64, n int) (float64, error) {
	return calculateSunRiseSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height, true, n)
}

// GetSunSetTime 精确计算日落时间，传入当日0时JDE
func GetSunSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	return calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height, false)
}

func GetSunSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height float64, n int) (float64, error) {
	return calculateSunRiseSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height, false, n)
}

// calculateSunRiseSetTime 统一的日出日落计算函数
func calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64, isSunrise bool) (float64, error) {
	julianDay = math.Floor(julianDay) + 1.5
	naturalTimeZone := math.Round(longitude / 15)
	sunAngle := StandardAltitudeSun(zenithShift, height, latitude)

	// 获取太阳上中天时间
	solarNoonTime := CulminationTime(julianDay, longitude, naturalTimeZone)

	// 检查极夜极昼条件
	if err := checkPolarConditions(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, isSunrise); err != nil {
		return 0, err
	}

	// 计算初始估算时间
	initialTime := calculateInitialSunTime(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, isSunrise)

	// 牛顿-拉夫逊迭代求精确解
	return sunRiseSetNewtonRaphsonIteration(initialTime, longitude, latitude, naturalTimeZone, sunAngle, timeZone), nil
}

func calculateSunRiseSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height float64, isSunrise bool, n int) (float64, error) {
	julianDay = math.Floor(julianDay) + 1.5
	naturalTimeZone := math.Round(longitude / 15)
	sunAngle := StandardAltitudeSun(zenithShift, height, latitude)

	solarNoonTime := CulminationTimeN(julianDay, longitude, naturalTimeZone, n)
	if err := checkPolarConditionsN(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, isSunrise, n); err != nil {
		return 0, err
	}

	initialTime := calculateInitialSunTimeN(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, isSunrise, n)
	return sunRiseSetNewtonRaphsonIterationN(initialTime, longitude, latitude, naturalTimeZone, sunAngle, timeZone, n), nil
}

// checkPolarConditions 检查极夜极昼条件
func checkPolarConditions(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle float64, isSunrise bool) error {
	if SunHeight(solarNoonTime, longitude, latitude, naturalTimeZone) < sunAngle {
		return ErrNeverRise
	}

	checkTime := solarNoonTime + 0.5
	if isSunrise {
		checkTime = solarNoonTime - 0.5
	}

	if SunHeight(checkTime, longitude, latitude, naturalTimeZone) > sunAngle {
		return ErrNeverSet
	}

	return nil
}

func checkPolarConditionsN(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle float64, isSunrise bool, n int) error {
	if SunHeightN(solarNoonTime, longitude, latitude, naturalTimeZone, n) < sunAngle {
		return ErrNeverRise
	}

	checkTime := solarNoonTime + 0.5
	if isSunrise {
		checkTime = solarNoonTime - 0.5
	}

	if SunHeightN(checkTime, longitude, latitude, naturalTimeZone, n) > sunAngle {
		return ErrNeverSet
	}

	return nil
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

func calculateInitialSunTimeN(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle float64, isSunrise bool, n int) float64 {
	apparentDeclination := HSunApparentDecN(solarNoonTime, n)
	cosHourAngle := (Sin(sunAngle) - Sin(apparentDeclination)*Sin(latitude)) / (Cos(apparentDeclination) * Cos(latitude))

	if math.Abs(cosHourAngle) <= 1 && latitude < 85 {
		hourAngle := ArcCos(cosHourAngle) / 15
		timeOffset := 25.0 / 24.0 / 60.0
		if !isSunrise {
			timeOffset = 35.0 / 24.0 / 60.0
		}

		if isSunrise {
			return solarNoonTime - hourAngle/24 - timeOffset
		}
		return solarNoonTime + hourAngle/24 + timeOffset
	}

	return iterativeApproachN(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, isSunrise, n)
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

func iterativeApproachN(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle float64, isSunrise bool, n int) float64 {
	estimatedTime := solarNoonTime
	stepSize := 15.0 / 60.0 / 24.0
	if isSunrise {
		stepSize = -stepSize
	}

	const maxIterations = 48
	for i := 0; i < maxIterations && lowSunHeightForN(estimatedTime, longitude, latitude, naturalTimeZone, n) > sunAngle; i++ {
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

func sunRiseSetNewtonRaphsonIterationN(initialTime, longitude, latitude, naturalTimeZone, sunAngle, timeZone float64, n int) float64 {
	const (
		convergenceThreshold = 0.00001
		derivativeStep       = 0.000005
	)

	currentTime := initialTime

	for {
		previousTime := currentTime
		functionValue := SunHeightN(previousTime, longitude, latitude, naturalTimeZone, n) - sunAngle
		derivative := (SunHeightN(previousTime+derivativeStep, longitude, latitude, naturalTimeZone, n) -
			SunHeightN(previousTime-derivativeStep, longitude, latitude, naturalTimeZone, n)) / (2 * derivativeStep)
		currentTime = previousTime - functionValue/derivative
		if math.Abs(currentTime-previousTime) <= convergenceThreshold {
			break
		}
	}

	return currentTime - naturalTimeZone/24 + timeZone/24
}

/*
 * 太阳高度角 世界时
 */
func SunHeight(jd, lon, lat, tz float64) float64 {
	//tmp := (tz*15 - lon) * 4 / 60
	//truejd := jd - tmp/24
	calcjd := jd - tz/24.0
	tjde := TD2UT(calcjd, true)
	st := Limit360(ApparentSiderealTime(calcjd)*15 + lon)
	ra, dec := HSunApparentRaDec(tjde)
	hourAngle := Limit360(st - ra)
	tmp2 := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(tmp2)
}

func SunHeightN(jd, lon, lat, tz float64, n int) float64 {
	calcjd := jd - tz/24.0
	tjde := TD2UT(calcjd, true)
	st := Limit360(ApparentSiderealTime(calcjd)*15 + lon)
	ra, dec := HSunApparentRaDecN(tjde, n)
	hourAngle := Limit360(st - ra)
	tmp2 := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(tmp2)
}

func LowSunHeight(jd, lon, lat, tz float64) float64 {
	//tmp := (tz*15 - lon) * 4 / 60
	//truejd := jd - tmp/24
	calcjd := jd - tz/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + lon)
	hourAngle := Limit360(st - SunApparentRa(TD2UT(calcjd, true)))
	dec := SunApparentDec(TD2UT(calcjd, true))
	tmp2 := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(tmp2)
}

func lowSunHeightForN(jd, lon, lat, tz float64, n int) float64 {
	if n < 0 {
		return LowSunHeight(jd, lon, lat, tz)
	}
	return SunHeightN(jd, lon, lat, tz, n)
}

func SunAzimuth(jd, lon, lat, tz float64) float64 {
	//tmp := (tz*15 - lon) * 4 / 60
	//truejd := jd - tmp/24
	calcjd := jd - tz/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + lon)
	hourAngle := Limit360(st - HSunApparentRa(TD2UT(calcjd, true)))
	tmp2 := Sin(hourAngle) / (Cos(hourAngle)*Sin(lat) - Tan(HSunApparentDec(TD2UT(calcjd, true)))*Cos(lat))
	azimuth := ArcTan(tmp2)
	if azimuth < 0 {
		if hourAngle/15 < 12 {
			return azimuth + 360
		}
		return azimuth + 180
	}
	if hourAngle/15 < 12 {
		return azimuth + 180
	}
	return azimuth
}

func SunAzimuthN(jd, lon, lat, tz float64, n int) float64 {
	calcjd := jd - tz/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + lon)
	hourAngle := Limit360(st - HSunApparentRaN(TD2UT(calcjd, true), n))
	tmp2 := Sin(hourAngle) / (Cos(hourAngle)*Sin(lat) - Tan(HSunApparentDecN(TD2UT(calcjd, true), n))*Cos(lat))
	azimuth := ArcTan(tmp2)
	if azimuth < 0 {
		if hourAngle/15 < 12 {
			return azimuth + 360
		}
		return azimuth + 180
	}
	if hourAngle/15 < 12 {
		return azimuth + 180
	}
	return azimuth
}

package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// CulminationTime 太阳中天时刻（按均时差计算）/ solar culmination time from the equation of time.
//
// localJD 是本地民用日锚点，只取 floor(localJD)（JD 整数 = 12:00 的正午锚点），既不是午夜也不是力学时；
// 调用方要传“本地 0 时对应 JD + 0.5”
// 才能落在同一本地日（sun/sun.go 的 CulminationTime 就是这么补的）。地方时相对世界时的偏移按角度归化到
// ±180°：超过 ±12 小时（如 UTC+14 配西经）时不归化会把中天推到相邻的一天。
func CulminationTime(localJD, lon, tz float64) float64 { //实际中天时间
	localJD = math.Floor(localJD)
	tmp := longitudeOffsetDegrees(tz*15-lon) * 4 / 60
	return localJD + tmp/24.0 - SunTime(localJD)/24.0
}

// CulminationTimeN 截断项太阳中天时刻 / truncated solar culmination time.
func CulminationTimeN(localJD, lon, tz float64, n int) float64 { //实际中天时间
	localJD = math.Floor(localJD)
	tmp := longitudeOffsetDegrees(tz*15-lon) * 4 / 60
	return localJD + tmp/24.0 - SunTimeN(localJD, n)/24.0
}

func longitudeOffsetDegrees(offset float64) float64 {
	if offset > 180 {
		offset -= 360
	}
	if offset < -180 {
		offset += 360
	}
	return offset
}

// EveningTwilight 昏朦影；localJD 是本地民用日锚点（当地 0 时），只取整数日。
func EveningTwilight(localJD, lon, lat, tz, targetAltitude float64) (float64, error) {
	localJD = math.Floor(localJD) + 1.5
	localTimeZone := math.Round(lon / 15)
	culminationTime := CulminationTime(localJD, lon, localTimeZone)
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
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		stDegree := SunHeight(prevJD, lon, lat, localTimeZone) - targetAltitude
		stDegreep := (SunHeight(prevJD+0.000005, lon, lat, localTimeZone) - SunHeight(prevJD-0.000005, lon, lat, localTimeZone)) / 0.00001
		return stDegree / stDegreep
	})
	if !ok {
		return 0, ErrInvalidObservationInput
	}
	return estimateJD - localTimeZone/24 + tz/24, nil
}

func EveningTwilightN(localJD, lon, lat, tz, targetAltitude float64, n int) (float64, error) {
	localJD = math.Floor(localJD) + 1.5
	localTimeZone := math.Round(lon / 15)
	culminationTime := CulminationTimeN(localJD, lon, localTimeZone, n)
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
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		stDegree := SunHeightN(prevJD, lon, lat, localTimeZone, n) - targetAltitude
		stDegreep := (SunHeightN(prevJD+0.000005, lon, lat, localTimeZone, n) - SunHeightN(prevJD-0.000005, lon, lat, localTimeZone, n)) / 0.00001
		return stDegree / stDegreep
	})
	if !ok {
		return 0, ErrInvalidObservationInput
	}
	return estimateJD - localTimeZone/24 + tz/24, nil
}

func MorningTwilight(localJD, lon, lat, tz, targetAltitude float64) (float64, error) {
	// 调整到中午12点
	localJD = math.Floor(localJD) + 1.5

	// 计算时区
	localTimeZone := math.Round(lon / 15)

	// 计算太阳上中天时间
	culminationTime := CulminationTime(localJD, lon, localTimeZone)

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
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		heightDiff := SunHeight(prevJD, lon, lat, localTimeZone) - targetAltitude
		heightDerivative := (SunHeight(prevJD+0.000005, lon, lat, localTimeZone) - SunHeight(prevJD-0.000005, lon, lat, localTimeZone)) / 0.00001
		return heightDiff / heightDerivative
	})
	if !ok {
		return 0, ErrInvalidObservationInput
	}

	return estimateJD - localTimeZone/24 + tz/24, nil
}

func MorningTwilightN(localJD, lon, lat, tz, targetAltitude float64, n int) (float64, error) {
	localJD = math.Floor(localJD) + 1.5
	localTimeZone := math.Round(lon / 15)
	culminationTime := CulminationTimeN(localJD, lon, localTimeZone, n)
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
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		heightDiff := SunHeightN(prevJD, lon, lat, localTimeZone, n) - targetAltitude
		heightDerivative := (SunHeightN(prevJD+0.000005, lon, lat, localTimeZone, n) - SunHeightN(prevJD-0.000005, lon, lat, localTimeZone, n)) / 0.00001
		return heightDiff / heightDerivative
	})
	if !ok {
		return 0, ErrInvalidObservationInput
	}

	return estimateJD - localTimeZone/24 + tz/24, nil
}

/*
 * 太阳时角
 */
func SunTimeAngle(jd, lon, lat, tz float64) float64 {
	startime := Limit360(ApparentSiderealTime(UTC2UT1(jd-tz/24))*15 + lon)
	timeangle := startime - HSunApparentRa(UTC2TT(jd-tz/24))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

func SunTimeAngleN(jd, lon, lat, tz float64, n int) float64 {
	startime := Limit360(ApparentSiderealTime(UTC2UT1(jd-tz/24))*15 + lon)
	timeangle := startime - HSunApparentRaN(UTC2TT(jd-tz/24), n)
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

type sunObservationState struct {
	altitude   float64
	distanceAU float64
}

func sunObservationStateN(jd, lon, lat, tz float64, n int) sunObservationState {
	calculationJD := jd - tz/24.0
	tt := UTC2TT(calculationJD)
	siderealTime := Limit360(ApparentSiderealTime(UTC2UT1(calculationJD))*15 + lon)
	ra, dec, distanceAU := hSunApparentRaDecDistanceN(tt, n)
	hourAngle := Limit360(siderealTime - ra)
	altitudeSine := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
	return sunObservationState{
		altitude:   ArcSin(altitudeSine),
		distanceAU: distanceAU,
	}
}

func sunRiseSetResidual(jd, longitude, latitude, timeZone, zenithShift, height float64, n int) float64 {
	state := sunObservationStateN(jd, longitude, latitude, timeZone, n)
	// 相对观测者下沉地平线的视上缘高度角 / Apparent upper-limb altitude relative to the observer's depressed horizon.
	residual := state.altitude + HeightDegreeByLat(height, latitude)
	if zenithShift != 0 {
		residual += RefractionFromTrueAltitude(state.altitude, refractionStandardPressureHPa, refractionStandardTemperatureC)
		residual += angularSemidiameterFromAU(sunEquatorialRadiusKM, state.distanceAU) / 3600
	}
	return residual
}

func sunRiseSetOnCivilDay(candidate, slope, civilDayStart, longitude, latitude, requestedTimeZone,
	localTimeZone, zenithShift, height float64, isSunrise bool, n int, fallbackErr error) (float64, error) {
	if eventRiseSetCandidateValid(candidate, civilDayStart, slope, isSunrise) {
		return candidate, nil
	}
	return eventDirectionalRiseSetSearch(civilDayStart, isSunrise, fallbackErr, func(outputJD float64) float64 {
		localJD := outputJD + localTimeZone/24 - requestedTimeZone/24
		return sunRiseSetResidual(localJD, longitude, latitude, localTimeZone, zenithShift, height, n)
	})
}

// GetSunRiseTime 精确计算日出时间，传入本地民用日 0 时锚点
func GetSunRiseTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	return calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height, true)
}

func GetSunRiseTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height float64, n int) (float64, error) {
	return calculateSunRiseSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height, true, n)
}

// GetSunSetTime 精确计算日落时间，传入本地民用日 0 时锚点
func GetSunSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	return calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height, false)
}

func GetSunSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height float64, n int) (float64, error) {
	return calculateSunRiseSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height, false, n)
}

// calculateSunRiseSetTime 统一的日出日落计算函数
func calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64, isSunrise bool) (float64, error) {
	if !isFiniteFloat(julianDay) || !isFiniteFloat(longitude) || !isFiniteFloat(latitude) || !isFiniteFloat(timeZone) || !isFiniteFloat(zenithShift) || !isFiniteFloat(height) {
		return 0, ErrInvalidObservationInput
	}
	civilDayStart := math.Floor(julianDay) + 0.5
	julianDay = math.Floor(julianDay) + 1.5
	naturalTimeZone := math.Round(longitude / 15)
	sunAngle := StandardAltitudeSun(zenithShift, height, latitude)

	// 获取太阳上中天时间
	solarNoonTime := CulminationTime(julianDay, longitude, naturalTimeZone)

	// 检查极夜极昼条件
	if err := checkPolarConditions(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, isSunrise); err != nil {
		return sunRiseSetOnCivilDay(math.NaN(), math.NaN(), civilDayStart, longitude, latitude, timeZone,
			naturalTimeZone, zenithShift, height, isSunrise, -1, err)
	}

	// 计算初始估算时间
	initialTime := calculateInitialSunTime(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, zenithShift, height, isSunrise)

	// 牛顿-拉夫逊迭代求精确解
	result, slope := sunRiseSetNewtonRaphsonIteration(initialTime, longitude, latitude, naturalTimeZone, zenithShift, height, timeZone)
	return sunRiseSetOnCivilDay(result, slope, civilDayStart, longitude, latitude, timeZone,
		naturalTimeZone, zenithShift, height, isSunrise, -1, nil)
}

func calculateSunRiseSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height float64, isSunrise bool, n int) (float64, error) {
	if !isFiniteFloat(julianDay) || !isFiniteFloat(longitude) || !isFiniteFloat(latitude) || !isFiniteFloat(timeZone) || !isFiniteFloat(zenithShift) || !isFiniteFloat(height) {
		return 0, ErrInvalidObservationInput
	}
	civilDayStart := math.Floor(julianDay) + 0.5
	julianDay = math.Floor(julianDay) + 1.5
	naturalTimeZone := math.Round(longitude / 15)
	sunAngle := StandardAltitudeSun(zenithShift, height, latitude)

	solarNoonTime := CulminationTimeN(julianDay, longitude, naturalTimeZone, n)
	if err := checkPolarConditionsN(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, isSunrise, n); err != nil {
		return sunRiseSetOnCivilDay(math.NaN(), math.NaN(), civilDayStart, longitude, latitude, timeZone,
			naturalTimeZone, zenithShift, height, isSunrise, n, err)
	}

	initialTime := calculateInitialSunTimeN(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, zenithShift, height, isSunrise, n)
	result, slope := sunRiseSetNewtonRaphsonIterationN(initialTime, longitude, latitude, naturalTimeZone, zenithShift, height, timeZone, n)
	return sunRiseSetOnCivilDay(result, slope, civilDayStart, longitude, latitude, timeZone,
		naturalTimeZone, zenithShift, height, isSunrise, n, nil)
}

// checkPolarConditions 检查极夜极昼条件
func checkPolarConditions(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height float64, isSunrise bool) error {
	if sunRiseSetResidual(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, -1) < 0 {
		return ErrNeverRise
	}

	checkTime := solarNoonTime + 0.5
	if isSunrise {
		checkTime = solarNoonTime - 0.5
	}

	if sunRiseSetResidual(checkTime, longitude, latitude, naturalTimeZone, zenithShift, height, -1) > 0 {
		return ErrNeverSet
	}

	return nil
}

func checkPolarConditionsN(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height float64, isSunrise bool, n int) error {
	if sunRiseSetResidual(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, n) < 0 {
		return ErrNeverRise
	}

	checkTime := solarNoonTime + 0.5
	if isSunrise {
		checkTime = solarNoonTime - 0.5
	}

	if sunRiseSetResidual(checkTime, longitude, latitude, naturalTimeZone, zenithShift, height, n) > 0 {
		return ErrNeverSet
	}

	return nil
}

// calculateInitialSunTime 计算日出日落的初始估算时间
func calculateInitialSunTime(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, zenithShift, height float64, isSunrise bool) float64 {
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
		return iterativeApproach(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, isSunrise)
	}
}

func calculateInitialSunTimeN(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, zenithShift, height float64, isSunrise bool, n int) float64 {
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

	return iterativeApproachN(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, isSunrise, n)
}

// iterativeApproach 迭代逼近法计算（用于极地等特殊条件）
func iterativeApproach(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height float64, isSunrise bool) float64 {
	estimatedTime := solarNoonTime
	stepSize := 15.0 / 60.0 / 24.0 // 15分钟步长
	if isSunrise {
		stepSize = -stepSize
	}

	const maxIterations = 48
	for i := 0; i < maxIterations && sunRiseSetResidual(estimatedTime, longitude, latitude, naturalTimeZone, zenithShift, height, -1) > 0; i++ {
		estimatedTime += stepSize
	}

	return estimatedTime
}

func iterativeApproachN(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height float64, isSunrise bool, n int) float64 {
	estimatedTime := solarNoonTime
	stepSize := 15.0 / 60.0 / 24.0
	if isSunrise {
		stepSize = -stepSize
	}

	const maxIterations = 48
	for i := 0; i < maxIterations && sunRiseSetResidual(estimatedTime, longitude, latitude, naturalTimeZone, zenithShift, height, n) > 0; i++ {
		estimatedTime += stepSize
	}

	return estimatedTime
}

// sunRiseSetNewtonRaphsonIteration 牛顿-拉夫逊迭代法求精确解
func sunRiseSetNewtonRaphsonIteration(initialTime, longitude, latitude, naturalTimeZone, zenithShift, height, timeZone float64) (float64, float64) {
	const (
		convergenceThreshold = 0.00001
		derivativeStep       = 0.000005
	)

	currentTime := initialTime
	slope := math.NaN()
	var ok bool
	currentTime, ok = eventNewtonRefine(currentTime, convergenceThreshold, func(previousTime float64) float64 {
		functionValue := sunRiseSetResidual(previousTime, longitude, latitude, naturalTimeZone, zenithShift, height, -1)
		slope = (sunRiseSetResidual(previousTime+derivativeStep, longitude, latitude, naturalTimeZone, zenithShift, height, -1) -
			sunRiseSetResidual(previousTime-derivativeStep, longitude, latitude, naturalTimeZone, zenithShift, height, -1)) / (2 * derivativeStep)
		return functionValue / slope
	})
	if !ok {
		return math.NaN(), math.NaN()
	}

	// 转换为指定时区
	return currentTime - naturalTimeZone/24 + timeZone/24, slope
}

func sunRiseSetNewtonRaphsonIterationN(initialTime, longitude, latitude, naturalTimeZone, zenithShift, height, timeZone float64, n int) (float64, float64) {
	const (
		convergenceThreshold = 0.00001
		derivativeStep       = 0.000005
	)

	currentTime := initialTime
	slope := math.NaN()
	var ok bool
	currentTime, ok = eventNewtonRefine(currentTime, convergenceThreshold, func(previousTime float64) float64 {
		functionValue := sunRiseSetResidual(previousTime, longitude, latitude, naturalTimeZone, zenithShift, height, n)
		slope = (sunRiseSetResidual(previousTime+derivativeStep, longitude, latitude, naturalTimeZone, zenithShift, height, n) -
			sunRiseSetResidual(previousTime-derivativeStep, longitude, latitude, naturalTimeZone, zenithShift, height, n)) / (2 * derivativeStep)
		return functionValue / slope
	})
	if !ok {
		return math.NaN(), math.NaN()
	}

	return currentTime - naturalTimeZone/24 + timeZone/24, slope
}

/*
 * 太阳高度角 世界时
 */
func SunHeight(jd, lon, lat, tz float64) float64 {
	return SunHeightN(jd, lon, lat, tz, -1)
}

func SunHeightN(jd, lon, lat, tz float64, n int) float64 {
	return sunObservationStateN(jd, lon, lat, tz, n).altitude
}

func LowSunHeight(jd, lon, lat, tz float64) float64 {
	//tmp := (tz*15 - lon) * 4 / 60
	//truejd := jd - tmp/24
	calcjd := jd - tz/24
	st := Limit360(ApparentSiderealTime(UTC2UT1(calcjd))*15 + lon)
	hourAngle := Limit360(st - SunApparentRa(UTC2TT(calcjd)))
	dec := SunApparentDec(UTC2TT(calcjd))
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
	st := Limit360(ApparentSiderealTime(UTC2UT1(calcjd))*15 + lon)
	hourAngle := Limit360(st - HSunApparentRa(UTC2TT(calcjd)))
	tmp2 := Sin(hourAngle) / (Cos(hourAngle)*Sin(lat) - Tan(HSunApparentDec(UTC2TT(calcjd)))*Cos(lat))
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
	st := Limit360(ApparentSiderealTime(UTC2UT1(calcjd))*15 + lon)
	hourAngle := Limit360(st - HSunApparentRaN(UTC2TT(calcjd), n))
	tmp2 := Sin(hourAngle) / (Cos(hourAngle)*Sin(lat) - Tan(HSunApparentDecN(UTC2TT(calcjd), n))*Cos(lat))
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

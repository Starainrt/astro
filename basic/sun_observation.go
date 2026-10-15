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

// CulminationTimeN 太阳中天时刻（n 变体） / n-parameterised solar culmination time.
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

// EveningTwilightWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func EveningTwilightWithInstance(localJD, lon, lat, tz, targetAltitude float64, inst *Instance) (float64, error) {
	localJD = math.Floor(localJD) + 1.5
	localTimeZone := math.Round(lon / 15)
	culminationTime := CulminationTimeWithInstance(localJD, lon, localTimeZone, inst)
	if SunHeightWithInstance(culminationTime, lon, lat, localTimeZone, inst) < targetAltitude {
		return 0, ErrNeverRise
	}
	if SunHeightWithInstance(culminationTime+0.5, lon, lat, localTimeZone, inst) > targetAltitude {
		return 0, ErrNeverSet
	}
	tmp := (Sin(targetAltitude) - Sin(HSunApparentDecWithInstance(culminationTime, inst))*Sin(lat)) / (Cos(HSunApparentDecWithInstance(culminationTime, inst)) * Cos(lat))
	var sundown float64
	if math.Abs(tmp) <= 1 && lat < 85 {
		hourOffset := ArcCos(tmp) / 15
		sundown = culminationTime + hourOffset/24.0 + 35.0/24.0/60.0
	} else {
		sundown = culminationTime
		i := 0
		for LowSunHeightWithInstance(sundown, lon, lat, localTimeZone, inst) > targetAltitude {
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
		stDegree := SunHeightWithInstance(prevJD, lon, lat, localTimeZone, inst) - targetAltitude
		stDegreep := (SunHeightWithInstance(prevJD+0.000005, lon, lat, localTimeZone, inst) - SunHeightWithInstance(prevJD-0.000005, lon, lat, localTimeZone, inst)) / 0.00001
		return stDegree / stDegreep
	})
	if !ok {
		return 0, ErrInvalidObservationInput
	}
	return estimateJD - localTimeZone/24 + tz/24, nil
}

// EveningTwilightNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func EveningTwilightNWithInstance(localJD, lon, lat, tz, targetAltitude float64, n int, inst *Instance) (float64, error) {
	localJD = math.Floor(localJD) + 1.5
	localTimeZone := math.Round(lon / 15)
	culminationTime := CulminationTimeNWithInstance(localJD, lon, localTimeZone, n, inst)
	if SunHeightNWithInstance(culminationTime, lon, lat, localTimeZone, n, inst) < targetAltitude {
		return 0, ErrNeverRise
	}
	if SunHeightNWithInstance(culminationTime+0.5, lon, lat, localTimeZone, n, inst) > targetAltitude {
		return 0, ErrNeverSet
	}
	tmp := (Sin(targetAltitude) - Sin(HSunApparentDecNWithInstance(culminationTime, n, inst))*Sin(lat)) / (Cos(HSunApparentDecNWithInstance(culminationTime, n, inst)) * Cos(lat))
	var sundown float64
	if math.Abs(tmp) <= 1 && lat < 85 {
		hourOffset := ArcCos(tmp) / 15
		sundown = culminationTime + hourOffset/24.0 + 35.0/24.0/60.0
	} else {
		sundown = culminationTime
		i := 0
		for lowSunHeightForN(sundown, lon, lat, localTimeZone, n, inst) > targetAltitude {
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
		stDegree := SunHeightNWithInstance(prevJD, lon, lat, localTimeZone, n, inst) - targetAltitude
		stDegreep := (SunHeightNWithInstance(prevJD+0.000005, lon, lat, localTimeZone, n, inst) - SunHeightNWithInstance(prevJD-0.000005, lon, lat, localTimeZone, n, inst)) / 0.00001
		return stDegree / stDegreep
	})
	if !ok {
		return 0, ErrInvalidObservationInput
	}
	return estimateJD - localTimeZone/24 + tz/24, nil
}

// MorningTwilightWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MorningTwilightWithInstance(localJD, lon, lat, tz, targetAltitude float64, inst *Instance) (float64, error) {
	// 调整到中午12点
	localJD = math.Floor(localJD) + 1.5

	// 计算时区
	localTimeZone := math.Round(lon / 15)

	// 计算太阳上中天时间
	culminationTime := CulminationTimeWithInstance(localJD, lon, localTimeZone, inst)

	// 检查极夜和极昼条件
	if SunHeightWithInstance(culminationTime, lon, lat, localTimeZone, inst) < targetAltitude {
		return 0, ErrNeverRise
	}
	if SunHeightWithInstance(culminationTime-0.5, lon, lat, localTimeZone, inst) > targetAltitude {
		return 0, ErrNeverSet
	}

	// 计算日出时间
	sunDec := HSunApparentDecWithInstance(culminationTime, inst)
	tmp := (Sin(targetAltitude) - Sin(sunDec)*Sin(lat)) / (Cos(sunDec) * Cos(lat))

	var sunrise float64
	if math.Abs(tmp) <= 1 && lat < 85 {
		hourAngle := ArcCos(tmp) / 15
		sunrise = culminationTime - hourAngle/24 - 25.0/(24.0*60.0)
	} else {
		sunrise = culminationTime
		for i := 0; i < 48 && LowSunHeightWithInstance(sunrise, lon, lat, localTimeZone, inst) > targetAltitude; i++ {
			sunrise -= 15.0 / (60.0 * 24.0) // 每次减少15分钟
		}
	}

	estimateJD := sunrise - 5.0/(24.0*60.0)
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		heightDiff := SunHeightWithInstance(prevJD, lon, lat, localTimeZone, inst) - targetAltitude
		heightDerivative := (SunHeightWithInstance(prevJD+0.000005, lon, lat, localTimeZone, inst) - SunHeightWithInstance(prevJD-0.000005, lon, lat, localTimeZone, inst)) / 0.00001
		return heightDiff / heightDerivative
	})
	if !ok {
		return 0, ErrInvalidObservationInput
	}

	return estimateJD - localTimeZone/24 + tz/24, nil
}

// MorningTwilightNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MorningTwilightNWithInstance(localJD, lon, lat, tz, targetAltitude float64, n int, inst *Instance) (float64, error) {
	localJD = math.Floor(localJD) + 1.5
	localTimeZone := math.Round(lon / 15)
	culminationTime := CulminationTimeNWithInstance(localJD, lon, localTimeZone, n, inst)
	if SunHeightNWithInstance(culminationTime, lon, lat, localTimeZone, n, inst) < targetAltitude {
		return 0, ErrNeverRise
	}
	if SunHeightNWithInstance(culminationTime-0.5, lon, lat, localTimeZone, n, inst) > targetAltitude {
		return 0, ErrNeverSet
	}

	sunDec := HSunApparentDecNWithInstance(culminationTime, n, inst)
	tmp := (Sin(targetAltitude) - Sin(sunDec)*Sin(lat)) / (Cos(sunDec) * Cos(lat))

	var sunrise float64
	if math.Abs(tmp) <= 1 && lat < 85 {
		hourAngle := ArcCos(tmp) / 15
		sunrise = culminationTime - hourAngle/24 - 25.0/(24.0*60.0)
	} else {
		sunrise = culminationTime
		for i := 0; i < 48 && lowSunHeightForN(sunrise, lon, lat, localTimeZone, n, inst) > targetAltitude; i++ {
			sunrise -= 15.0 / (60.0 * 24.0)
		}
	}

	estimateJD := sunrise - 5.0/(24.0*60.0)
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		heightDiff := SunHeightNWithInstance(prevJD, lon, lat, localTimeZone, n, inst) - targetAltitude
		heightDerivative := (SunHeightNWithInstance(prevJD+0.000005, lon, lat, localTimeZone, n, inst) - SunHeightNWithInstance(prevJD-0.000005, lon, lat, localTimeZone, n, inst)) / 0.00001
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

func sunObservationStateN(jd, lon, lat, tz float64, n int, instances ...*Instance) sunObservationState {
	hourAngle, dec, distanceAU := sunTopocentricHourAngleDecN(jd, lon, lat, tz, n, instanceValue(instances))
	return sunObservationState{
		altitude:   horizontalAltitude(hourAngle, dec, lat),
		distanceAU: distanceAU,
	}
}

func sunRiseSetResidual(jd, longitude, latitude, timeZone, zenithShift, height float64, n int, instances ...*Instance) float64 {
	return sunRiseSetResidualWeather(jd, longitude, latitude, timeZone, zenithShift, height,
		refractionStandardPressureHPa, refractionStandardTemperatureC, n, instanceValue(instances))
}

func sunRiseSetResidualWeather(jd, longitude, latitude, timeZone, zenithShift, height, pressureHPa, temperatureC float64, n int, inst *Instance) float64 {
	state := sunObservationStateN(jd, longitude, latitude, timeZone, n, inst)
	// 相对观测者下沉地平线的视上缘高度角 / Apparent upper-limb altitude relative to the observer's depressed horizon.
	residual := state.altitude + HeightDegreeByLat(height, latitude)
	if zenithShift != 0 {
		residual += RefractionFromTrueAltitude(state.altitude, pressureHPa, temperatureC)
		residual += angularSemidiameterFromAU(sunEquatorialRadiusKM, state.distanceAU) / 3600
	}
	return residual
}

func sunRiseSetOnCivilDay(candidate, slope, civilDayStart, longitude, latitude, requestedTimeZone,
	localTimeZone, zenithShift, height float64, isSunrise bool, n int, fallbackErr error, instances ...*Instance) (float64, error) {
	inst := instanceValue(instances)

	if eventRiseSetCandidateValid(candidate, civilDayStart, slope, isSunrise) {
		return candidate, nil
	}
	return eventDirectionalRiseSetSearch(civilDayStart, isSunrise, fallbackErr, func(outputJD float64) float64 {
		localJD := outputJD + localTimeZone/24 - requestedTimeZone/24
		return sunRiseSetResidual(localJD, longitude, latitude, localTimeZone, zenithShift, height, n, inst)
	})
}

// GetSunRiseTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func GetSunRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, zenithShift, height float64, inst *Instance) (float64, error) {
	return calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height, true, inst)
}

// GetSunRiseTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func GetSunRiseTimeNWithInstance(julianDay, longitude, latitude, timeZone, zenithShift, height float64, n int, inst *Instance) (float64, error) {
	return calculateSunRiseSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height, true, n, inst)
}

// GetSunSetTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func GetSunSetTimeWithInstance(julianDay, longitude, latitude, timeZone, zenithShift, height float64, inst *Instance) (float64, error) {
	return calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height, false, inst)
}

// GetSunSetTimeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func GetSunSetTimeNWithInstance(julianDay, longitude, latitude, timeZone, zenithShift, height float64, n int, inst *Instance) (float64, error) {
	return calculateSunRiseSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height, false, n, inst)
}

// calculateSunRiseSetTime 统一的日出日落计算函数
func calculateSunRiseSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64, isSunrise bool, instances ...*Instance) (float64, error) {
	inst := instanceValue(instances)

	if !isFiniteFloat(julianDay) || !isFiniteFloat(longitude) || !isFiniteFloat(latitude) || !isFiniteFloat(timeZone) || !isFiniteFloat(zenithShift) || !isFiniteFloat(height) {
		return 0, ErrInvalidObservationInput
	}
	civilDayStart := math.Floor(julianDay) + 0.5
	julianDay = math.Floor(julianDay) + 1.5
	naturalTimeZone := math.Round(longitude / 15)
	sunAngle := StandardAltitudeSun(zenithShift, height, latitude)

	// 获取太阳上中天时间
	solarNoonTime := CulminationTimeWithInstance(julianDay, longitude, naturalTimeZone, inst)

	// 检查极夜极昼条件
	if err := checkPolarConditions(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, isSunrise, inst); err != nil {
		return sunRiseSetOnCivilDay(math.NaN(), math.NaN(), civilDayStart, longitude, latitude, timeZone,
			naturalTimeZone, zenithShift, height, isSunrise, -1, err, inst)
	}

	// 计算初始估算时间
	initialTime := calculateInitialSunTime(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, zenithShift, height, isSunrise, inst)

	// 牛顿-拉夫逊迭代求精确解
	result, slope := sunRiseSetNewtonRaphsonIteration(initialTime, longitude, latitude, naturalTimeZone, zenithShift, height, timeZone, inst)
	return sunRiseSetOnCivilDay(result, slope, civilDayStart, longitude, latitude, timeZone,
		naturalTimeZone, zenithShift, height, isSunrise, -1, nil, inst)
}

func calculateSunRiseSetTimeN(julianDay, longitude, latitude, timeZone, zenithShift, height float64, isSunrise bool, n int, instances ...*Instance) (float64, error) {
	inst := instanceValue(instances)

	if !isFiniteFloat(julianDay) || !isFiniteFloat(longitude) || !isFiniteFloat(latitude) || !isFiniteFloat(timeZone) || !isFiniteFloat(zenithShift) || !isFiniteFloat(height) {
		return 0, ErrInvalidObservationInput
	}
	civilDayStart := math.Floor(julianDay) + 0.5
	julianDay = math.Floor(julianDay) + 1.5
	naturalTimeZone := math.Round(longitude / 15)
	sunAngle := StandardAltitudeSun(zenithShift, height, latitude)

	solarNoonTime := CulminationTimeNWithInstance(julianDay, longitude, naturalTimeZone, n, inst)
	if err := checkPolarConditionsN(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, isSunrise, n, inst); err != nil {
		return sunRiseSetOnCivilDay(math.NaN(), math.NaN(), civilDayStart, longitude, latitude, timeZone,
			naturalTimeZone, zenithShift, height, isSunrise, n, err, inst)
	}

	initialTime := calculateInitialSunTimeN(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, zenithShift, height, isSunrise, n, inst)
	result, slope := sunRiseSetNewtonRaphsonIterationN(initialTime, longitude, latitude, naturalTimeZone, zenithShift, height, timeZone, n, inst)
	return sunRiseSetOnCivilDay(result, slope, civilDayStart, longitude, latitude, timeZone,
		naturalTimeZone, zenithShift, height, isSunrise, n, nil, inst)
}

// checkPolarConditions 检查极夜极昼条件
func checkPolarConditions(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height float64, isSunrise bool, instances ...*Instance) error {
	inst := instanceValue(instances)

	if sunRiseSetResidual(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, -1, inst) < 0 {
		return ErrNeverRise
	}

	checkTime := solarNoonTime + 0.5
	if isSunrise {
		checkTime = solarNoonTime - 0.5
	}

	if sunRiseSetResidual(checkTime, longitude, latitude, naturalTimeZone, zenithShift, height, -1, inst) > 0 {
		return ErrNeverSet
	}

	return nil
}

func checkPolarConditionsN(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height float64, isSunrise bool, n int, instances ...*Instance) error {
	inst := instanceValue(instances)

	if sunRiseSetResidual(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, n, inst) < 0 {
		return ErrNeverRise
	}

	checkTime := solarNoonTime + 0.5
	if isSunrise {
		checkTime = solarNoonTime - 0.5
	}

	if sunRiseSetResidual(checkTime, longitude, latitude, naturalTimeZone, zenithShift, height, n, inst) > 0 {
		return ErrNeverSet
	}

	return nil
}

// calculateInitialSunTime 计算日出日落的初始估算时间
func calculateInitialSunTime(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, zenithShift, height float64, isSunrise bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	// 使用球面三角法计算: (sin(ho)-sin(φ)*sin(δ))/(cos(φ)*cos(δ))
	apparentDeclination := HSunApparentDecWithInstance(solarNoonTime, inst)
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
		return iterativeApproach(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, isSunrise, inst)
	}
}

func calculateInitialSunTimeN(solarNoonTime, longitude, latitude, naturalTimeZone, sunAngle, zenithShift, height float64, isSunrise bool, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	apparentDeclination := HSunApparentDecNWithInstance(solarNoonTime, n, inst)
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

	return iterativeApproachN(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height, isSunrise, n, inst)
}

// iterativeApproach 迭代逼近法计算（用于极地等特殊条件）
func iterativeApproach(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height float64, isSunrise bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	estimatedTime := solarNoonTime
	stepSize := 15.0 / 60.0 / 24.0 // 15分钟步长
	if isSunrise {
		stepSize = -stepSize
	}

	const maxIterations = 48
	for i := 0; i < maxIterations && sunRiseSetResidual(estimatedTime, longitude, latitude, naturalTimeZone, zenithShift, height, -1, inst) > 0; i++ {
		estimatedTime += stepSize
	}

	return estimatedTime
}

func iterativeApproachN(solarNoonTime, longitude, latitude, naturalTimeZone, zenithShift, height float64, isSunrise bool, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	estimatedTime := solarNoonTime
	stepSize := 15.0 / 60.0 / 24.0
	if isSunrise {
		stepSize = -stepSize
	}

	const maxIterations = 48
	for i := 0; i < maxIterations && sunRiseSetResidual(estimatedTime, longitude, latitude, naturalTimeZone, zenithShift, height, n, inst) > 0; i++ {
		estimatedTime += stepSize
	}

	return estimatedTime
}

// sunRiseSetNewtonRaphsonIteration 牛顿-拉夫逊迭代法求精确解
func sunRiseSetNewtonRaphsonIteration(initialTime, longitude, latitude, naturalTimeZone, zenithShift, height, timeZone float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	const (
		convergenceThreshold = 0.00001
		derivativeStep       = 0.000005
	)

	currentTime := initialTime
	slope := math.NaN()
	var ok bool
	currentTime, ok = eventNewtonRefine(currentTime, convergenceThreshold, func(previousTime float64) float64 {
		functionValue := sunRiseSetResidual(previousTime, longitude, latitude, naturalTimeZone, zenithShift, height, -1, inst)
		slope = (sunRiseSetResidual(previousTime+derivativeStep, longitude, latitude, naturalTimeZone, zenithShift, height, -1, inst) -
			sunRiseSetResidual(previousTime-derivativeStep, longitude, latitude, naturalTimeZone, zenithShift, height, -1, inst)) / (2 * derivativeStep)
		return functionValue / slope
	})
	if !ok {
		return math.NaN(), math.NaN()
	}

	// 转换为指定时区
	return currentTime - naturalTimeZone/24 + timeZone/24, slope
}

func sunRiseSetNewtonRaphsonIterationN(initialTime, longitude, latitude, naturalTimeZone, zenithShift, height, timeZone float64, n int, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	const (
		convergenceThreshold = 0.00001
		derivativeStep       = 0.000005
	)

	currentTime := initialTime
	slope := math.NaN()
	var ok bool
	currentTime, ok = eventNewtonRefine(currentTime, convergenceThreshold, func(previousTime float64) float64 {
		functionValue := sunRiseSetResidual(previousTime, longitude, latitude, naturalTimeZone, zenithShift, height, n, inst)
		slope = (sunRiseSetResidual(previousTime+derivativeStep, longitude, latitude, naturalTimeZone, zenithShift, height, n, inst) -
			sunRiseSetResidual(previousTime-derivativeStep, longitude, latitude, naturalTimeZone, zenithShift, height, n, inst)) / (2 * derivativeStep)
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

func lowSunHeightForN(jd, lon, lat, tz float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if n < 0 {
		return LowSunHeightWithInstance(jd, lon, lat, tz, inst)
	}
	return SunHeightNWithInstance(jd, lon, lat, tz, n, inst)
}

func SunAzimuth(jd, lon, lat, tz float64) float64 {
	return sunAzimuthN(jd, lon, lat, tz, -1, nil)
}

func SunAzimuthN(jd, lon, lat, tz float64, n int) float64 {
	return sunAzimuthN(jd, lon, lat, tz, n, nil)
}

package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

/*
 * 月球方位角
 */
func MoonAzimuth(jd, lon, lat, tz float64) float64 {

	//tmp := (tz*15 - lon) * 4 / 60
	calcjd := TD2UT(jd-tz/24, true)
	ra := MoonTrueRa(calcjd)
	dec := MoonTrueDec(calcjd)
	away := MoonAway(calcjd) / 149597870.7
	ndec := TopocentricDec(ra, dec, lat, lon, jd-tz/24, away, 0)
	nra := TopocentricRa(ra, dec, lat, lon, jd-tz/24, away, 0)
	calcjd = jd - tz/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + lon)
	hourAngle := Limit360(st - nra)
	tmp2 := Sin(hourAngle) / (Cos(hourAngle)*Sin(lat) - Tan(ndec)*Cos(lat))
	azimuth := ArcTan(tmp2)
	if azimuth < 0 {
		if hourAngle/15 < 12 {
			return azimuth + 360
		} else {
			return azimuth + 180
		}
	} else {
		if hourAngle/15 < 12 {
			return azimuth + 180
		} else {
			return azimuth
		}
	}

}

func MoonHeight(jd, lon, lat, tz float64) float64 {
	//	tmp := (tz*15 - lon) * 4 / 60
	//truejd=jd-tmp/24;
	calcjd := TD2UT(jd-tz/24, true)
	ra := MoonTrueRa(calcjd)
	dec := MoonTrueDec(calcjd)
	away := MoonAway(calcjd) / 149597870.7
	ndec := TopocentricDec(ra, dec, lat, lon, jd-tz/24, away, 0)
	nra := TopocentricRa(ra, dec, lat, lon, jd-tz/24, away, 0)
	calcjd = jd - tz/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + lon)
	hourAngle := Limit360(st - nra)
	tmp2 := Sin(lat)*Sin(ndec) + Cos(ndec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(tmp2)
}

func HMoonAzimuth(jd, lon, lat, tz float64) float64 {
	return HMoonAzimuthN(jd, lon, lat, tz, -1)
}

func HMoonAzimuthN(jd, lon, lat, tz float64, n int) float64 {
	calcjd := TD2UT(jd-tz/24, true)
	ra := HMoonTrueRaN(calcjd, n)
	dec := HMoonTrueDecN(calcjd, n)
	away := HMoonAwayN(calcjd, n) / 149597870.7
	ndec := TopocentricDec(ra, dec, lat, lon, jd-tz/24, away, 0)
	nra := TopocentricRa(ra, dec, lat, lon, jd-tz/24, away, 0)
	calcjd = jd - tz/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + lon)
	hourAngle := Limit360(st - nra)
	tmp2 := Sin(hourAngle) / (Cos(hourAngle)*Sin(lat) - Tan(ndec)*Cos(lat))
	azimuth := ArcTan(tmp2)
	if azimuth < 0 {
		if hourAngle/15 < 12 {
			return azimuth + 360
		} else {
			return azimuth + 180
		}
	} else {
		if hourAngle/15 < 12 {
			return azimuth + 180
		} else {
			return azimuth
		}
	}
}
func HMoonHeight(jd, lon, lat, tz float64) float64 {
	return HMoonHeightN(jd, lon, lat, tz, -1)
}

type moonObservationState struct {
	altitude   float64
	distanceKM float64
}

func hMoonObservationStateN(jd, lon, lat, tz, height float64, n int) moonObservationState {
	calculationJD := TD2UT(jd-tz/24, true)
	ra, dec := HMoonTrueRaDecN(calculationJD, n)
	distanceKM := HMoonAwayN(calculationJD, n)
	distanceAU := distanceKM / angularDiameterAstronomicalUnitKM
	topocentricRA, topocentricDec := TopocentricRaDec(ra, dec, lat, lon, jd-tz/24, distanceAU, height)
	siderealTime := Limit360(ApparentSiderealTime(jd-tz/24)*15 + lon)
	hourAngle := Limit360(siderealTime - topocentricRA)
	altitudeSine := Sin(lat)*Sin(topocentricDec) + Cos(topocentricDec)*Cos(lat)*Cos(hourAngle)
	return moonObservationState{
		altitude:   ArcSin(altitudeSine),
		distanceKM: distanceKM,
	}
}

func HMoonHeightN(jd, lon, lat, tz float64, n int) float64 {
	return hMoonObservationStateN(jd, lon, lat, tz, 0, n).altitude
}

func moonRiseSetResidual(jd, longitude, latitude, timeZone, zenithShift, height float64, n int) float64 {
	state := hMoonObservationStateN(jd, longitude, latitude, timeZone, height, n)
	// 相对观测者下沉地平线的视上缘高度角 / Apparent upper-limb altitude relative to the observer's depressed horizon.
	residual := state.altitude + HeightDegreeByLat(height, latitude)
	if zenithShift != 0 {
		residual += RefractionFromTrueAltitude(state.altitude, refractionStandardPressureHPa, refractionStandardTemperatureC)
		residual += angularSemidiameterArcsec(moonEquatorialRadiusKM, state.distanceKM) / 3600
	}
	return residual
}

// moonRiseSetOnCivilDay 在民用日内求升/落时刻；找不到过零时的错误口径与 rise_set.go 的 ErrNeverRise/ErrNeverSet 一致，
// fallbackErr 是调用方用中天/下中天残差预判的同一几何结论，命中时优先于扫描结果。
func moonRiseSetOnCivilDay(candidate, slope, civilDayStart, longitude, latitude, originalTimeZone,
	localTimeZone, zenithShift, height float64, isRise bool, fallbackErr error) (float64, error) {
	if eventRiseSetCandidateValid(candidate, civilDayStart, slope, isRise) {
		return candidate, nil
	}
	return eventDirectionalRiseSetSearch(civilDayStart, isRise, fallbackErr, func(outputJD float64) float64 {
		localJD := outputJD + localTimeZone/24 - originalTimeZone/24
		return moonRiseSetResidual(localJD, longitude, latitude, localTimeZone, zenithShift, height, -1)
	})
}

// 废弃
func GetMoonTZTime(jd, lon, lat, tz float64) float64 { //实际中天时间{
	jd = math.Floor(jd) + 0.5
	ttm := MoonTimeAngle(jd, lon, lat, tz)
	if ttm > 0 && ttm < 180 {
		jd += 0.5
	}
	estimateJD := jd
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		stDegree := MoonTimeAngle(prevJD, lon, lat, tz) - 359.599
		stDegreep := (MoonTimeAngle(prevJD+0.000005, lon, lat, tz) - MoonTimeAngle(prevJD-0.000005, lon, lat, tz)) / 0.00001
		return stDegree / stDegreep
	})
	if !ok {
		return math.NaN()
	}
	return estimateJD
}

func MoonCulminationTime(jde, lon, lat, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	estimateJD := jde + Limit360(360-MoonTimeAngle(jde, lon, lat, timezone))/15.0/24.0/0.9
	limitHA := func(jde, lon, timezone float64) float64 {
		ha := MoonTimeAngle(jde, lon, lat, timezone)
		if ha < 180 {
			ha += 360
		}
		return ha
	}
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		stDegree := limitHA(prevJD, lon, timezone) - 360
		stDegreep := (limitHA(prevJD+0.000005, lon, timezone) - limitHA(prevJD-0.000005, lon, timezone)) / 0.00001
		return stDegree / stDegreep
	})
	if !ok {
		return math.NaN()
	}
	return estimateJD
}

func MoonTimeAngle(jd, lon, lat, tz float64) float64 {
	startime := Limit360(ApparentSiderealTime(jd-tz/24)*15 + lon)
	timeangle := startime - HMoonApparentRa(jd, lon, lat, tz)
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

func GetMoonRiseTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	if !isFiniteFloat(julianDay) || !isFiniteFloat(longitude) || !isFiniteFloat(latitude) || !isFiniteFloat(timeZone) || !isFiniteFloat(zenithShift) || !isFiniteFloat(height) {
		return 0, ErrInvalidObservationInput
	}
	originalTimeZone := timeZone
	timeZone = longitude / 15
	var timeToMeridian float64
	civilDayStart := math.Floor(julianDay) + 0.5
	//julianDay = math.Floor(julianDay) + 0.5 - originalTimeZone/24 + timeZone/24 // 求0时JDE
	//fix:这里时间分界线应当以传入的时区为准，不应当使用当地时区，否则在0时的判断会出错
	julianDay = math.Floor(julianDay) + 0.5
	estimatedTime := julianDay
	moonResidual := moonRiseSetResidual(julianDay, longitude, latitude, originalTimeZone, zenithShift, height, -1)

	moonAngle := StandardAltitudeMoon(zenithShift, height, latitude)

	moonAngleTime := MoonTimeAngle(julianDay, longitude, latitude, originalTimeZone)

	if moonResidual > 0 { // 月亮在地平线上或在落下与下中天之间
		if moonAngleTime > 180 {
			timeToMeridian = (180 + 360 - moonAngleTime) / 15
		} else {
			timeToMeridian = (180 - moonAngleTime) / 15
		}
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24*12.0)/15.0/24.0)
	}

	if moonResidual < 0 && moonAngleTime > 180 {
		timeToMeridian = (180 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24*12.0)/15.0/24.0)
	} else if moonResidual < 0 && moonAngleTime < 180 {
		timeToMeridian = (180 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24*12.0)/15.0/24.0)
	}

	currentAngle := MoonTimeAngle(estimatedTime, longitude, latitude, timeZone)
	if math.Abs(currentAngle-180) > 0.5 {
		estimatedTime += (180 - currentAngle) * 4.0 / 60.0 / 24.0
	}

	currentResidual := moonRiseSetResidual(estimatedTime, longitude, latitude, timeZone, zenithShift, height, -1)
	if !(currentResidual < -10 && math.Abs(latitude) < 60) {
		if currentResidual > 0 {
			// 下中天仍在地平线上：当日无落下（也无可升起），口径见 moonRiseSetOnCivilDay。
			return moonRiseSetOnCivilDay(math.NaN(), math.NaN(), civilDayStart, longitude, latitude,
				originalTimeZone, timeZone, zenithShift, height, true, ErrNeverSet)
		}
		checkTime := estimatedTime + 12.0/24.0 + 6.0/15.0/24.0
		checkAngle := MoonTimeAngle(checkTime, longitude, latitude, timeZone)
		if checkAngle < 90 {
			checkAngle += 360
		}
		checkTime += (360 - checkAngle) * 4.0 / 60.0 / 24.0
		if moonRiseSetResidual(checkTime, longitude, latitude, timeZone, zenithShift, height, -1) < 0 {
			// 上中天仍在地平线下：当日无升起。
			return moonRiseSetOnCivilDay(math.NaN(), math.NaN(), civilDayStart, longitude, latitude,
				originalTimeZone, timeZone, zenithShift, height, true, ErrNeverRise)
		}
	}

	moonDeclination := MoonApparentDec(estimatedTime, longitude, latitude, timeZone)
	tmp := (Sin(moonAngle) - Sin(moonDeclination)*Sin(latitude)) / (Cos(moonDeclination) * Cos(latitude))

	if math.Abs(tmp) <= 1 && latitude < 85 {
		hourAngle := (180 - ArcCos(tmp)) / 15
		estimatedTime += hourAngle/24.00 + hourAngle/33.00/15.00
	} else {
		i := 0
		for moonRiseSetResidual(estimatedTime, longitude, latitude, timeZone, zenithShift, height, -1) < 0 {
			i++
			estimatedTime += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}

	// 使用牛顿迭代法求精确解
	estimatedTime, slope := moonRiseSetResidualIteration(estimatedTime, longitude, latitude, timeZone, zenithShift, height, 0.00002)
	estimatedTime = estimatedTime - timeZone/24 + originalTimeZone/24
	return moonRiseSetOnCivilDay(estimatedTime, slope, civilDayStart, longitude, latitude,
		originalTimeZone, timeZone, zenithShift, height, true, nil)
}

func GetMoonSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	if !isFiniteFloat(julianDay) || !isFiniteFloat(longitude) || !isFiniteFloat(latitude) || !isFiniteFloat(timeZone) || !isFiniteFloat(zenithShift) || !isFiniteFloat(height) {
		return 0, ErrInvalidObservationInput
	}
	originalTimeZone := timeZone
	timeZone = longitude / 15
	var timeToMeridian float64
	civilDayStart := math.Floor(julianDay) + 0.5
	//julianDay = math.Floor(julianDay) + 0.5 - originalTimeZone/24 + timeZone/24 // 求0时JDE
	//fix:这里时间分界线应当以传入的时区为准，不应当使用当地时区，否则在0时的判断会出错
	julianDay = math.Floor(julianDay) + 0.5
	estimatedTime := julianDay
	moonResidual := moonRiseSetResidual(julianDay, longitude, latitude, originalTimeZone, zenithShift, height, -1)

	moonAngle := StandardAltitudeMoon(zenithShift, height, latitude)

	moonAngleTime := MoonTimeAngle(julianDay, longitude, latitude, originalTimeZone)

	if moonResidual < 0 {
		timeToMeridian = (360 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24.0*12.0)/15.0/24.0)
	}

	// 月亮在地平线上或在落下与下中天之间
	if moonResidual > 0 && moonAngleTime < 180 {
		timeToMeridian = (-moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24.0 + (timeToMeridian/24.0*12.0)/15.0/24.0)
	} else if moonResidual > 0 {
		timeToMeridian = (360 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24.0 + (timeToMeridian/24.0*12.0)/15.0/24.0)
	}

	currentAngle := MoonTimeAngle(estimatedTime, longitude, latitude, timeZone)
	if currentAngle < 180 {
		currentAngle += 360
	}
	if math.Abs(currentAngle-360) > 0.5 {
		estimatedTime += (360 - currentAngle) * 4.0 / 60.0 / 24.0
	}

	// estimatedTime = 月球中天时间
	currentResidual := moonRiseSetResidual(estimatedTime, longitude, latitude, timeZone, zenithShift, height, -1)
	if !(currentResidual > 10 && math.Abs(latitude) < 60) {
		if currentResidual < 0 {
			// 上中天仍在地平线下：当日无升起，也就无落下。
			return moonRiseSetOnCivilDay(math.NaN(), math.NaN(), civilDayStart, longitude, latitude,
				originalTimeZone, timeZone, zenithShift, height, false, ErrNeverRise)
		}
		checkTime := estimatedTime + 12.0/24.0 + 6.0/15.0/24.0
		angleSubtraction := 180 - MoonTimeAngle(checkTime, longitude, latitude, timeZone)
		checkTime += angleSubtraction * 4.0 / 60.0 / 24.0
		if moonRiseSetResidual(checkTime, longitude, latitude, timeZone, zenithShift, height, -1) > 0 {
			// 下中天仍在地平线上：当日无落下。
			return moonRiseSetOnCivilDay(math.NaN(), math.NaN(), civilDayStart, longitude, latitude,
				originalTimeZone, timeZone, zenithShift, height, false, ErrNeverSet)
		}
	}

	moonDeclination := MoonApparentDec(estimatedTime, longitude, latitude, timeZone)
	tmp := (Sin(moonAngle) - Sin(moonDeclination)*Sin(latitude)) / (Cos(moonDeclination) * Cos(latitude))

	if math.Abs(tmp) <= 1 && latitude < 85 {
		hourAngle := (ArcCos(tmp)) / 15.0
		estimatedTime += hourAngle/24 + hourAngle/33.0/15.0
	} else {
		i := 0
		for moonRiseSetResidual(estimatedTime, longitude, latitude, timeZone, zenithShift, height, -1) > 0 {
			i++
			estimatedTime += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}

	// 使用牛顿迭代法求精确解
	estimatedTime, slope := moonRiseSetResidualIteration(estimatedTime, longitude, latitude, timeZone, zenithShift, height, 0.00002)
	estimatedTime = estimatedTime - timeZone/24 + originalTimeZone/24
	return moonRiseSetOnCivilDay(estimatedTime, slope, civilDayStart, longitude, latitude,
		originalTimeZone, timeZone, zenithShift, height, false, nil)
}

// heightFunction 高度函数类型定义，用于牛顿迭代法
type heightFunction func(time, longitude, latitude, timeZone float64) float64

// moonRiseSetNewtonRaphsonIteration 牛顿-拉夫逊迭代法求解天体高度方程
func moonRiseSetNewtonRaphsonIteration(initialTime, longitude, latitude, timeZone, targetAngle float64,
	heightFunc heightFunction, tolerance float64) float64 {
	const derivativeStep = 0.000005

	currentTime := initialTime

	var ok bool
	currentTime, ok = eventNewtonRefine(currentTime, tolerance, func(previousTime float64) float64 {
		functionValue := heightFunc(previousTime, longitude, latitude, timeZone) - targetAngle
		derivative := (heightFunc(previousTime+derivativeStep, longitude, latitude, timeZone) -
			heightFunc(previousTime-derivativeStep, longitude, latitude, timeZone)) / (2 * derivativeStep)
		return functionValue / derivative
	})
	if !ok {
		return math.NaN()
	}

	return currentTime
}

func moonRiseSetResidualIteration(initialTime, longitude, latitude, timeZone, zenithShift, height, tolerance float64) (float64, float64) {
	const derivativeStep = 0.000005

	slope := math.NaN()
	currentTime, ok := eventNewtonRefine(initialTime, tolerance, func(previousTime float64) float64 {
		functionValue := moonRiseSetResidual(previousTime, longitude, latitude, timeZone, zenithShift, height, -1)
		slope = (moonRiseSetResidual(previousTime+derivativeStep, longitude, latitude, timeZone, zenithShift, height, -1) -
			moonRiseSetResidual(previousTime-derivativeStep, longitude, latitude, timeZone, zenithShift, height, -1)) / (2 * derivativeStep)
		return functionValue / slope
	})
	if !ok {
		return math.NaN(), math.NaN()
	}
	return currentTime, slope
}

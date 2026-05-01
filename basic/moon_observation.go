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

func HMoonHeightN(jd, lon, lat, tz float64, n int) float64 {
	calcjd := TD2UT(jd-tz/24, true)
	ra, dec := HMoonTrueRaDecN(calcjd, n)
	away := HMoonAwayN(calcjd, n) / 149597870.7
	nra, ndec := TopocentricRaDec(ra, dec, lat, lon, calcjd, away, 0)
	calcjd = jd - tz/24
	st := Limit360(ApparentSiderealTime(calcjd)*15 + lon)
	hourAngle := Limit360(st - nra)
	tmp2 := Sin(lat)*Sin(ndec) + Cos(ndec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(tmp2)
}

// 废弃
func GetMoonTZTime(jd, lon, lat, tz float64) float64 { //实际中天时间{
	jd = math.Floor(jd) + 0.5
	ttm := MoonTimeAngle(jd, lon, lat, tz)
	if ttm > 0 && ttm < 180 {
		jd += 0.5
	}
	estimateJD := jd
	for {
		prevJD := estimateJD
		stDegree := MoonTimeAngle(prevJD, lon, lat, tz) - 359.599
		stDegreep := (MoonTimeAngle(prevJD+0.000005, lon, lat, tz) - MoonTimeAngle(prevJD-0.000005, lon, lat, tz)) / 0.00001
		estimateJD = prevJD - stDegree/stDegreep
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
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
	for {
		prevJD := estimateJD
		stDegree := limitHA(prevJD, lon, timezone) - 360
		stDegreep := (limitHA(prevJD+0.000005, lon, timezone) - limitHA(prevJD-0.000005, lon, timezone)) / 0.00001
		estimateJD = prevJD - stDegree/stDegreep
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
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
	originalTimeZone := timeZone
	timeZone = longitude / 15
	var timeToMeridian float64
	julianDayZero := math.Floor(julianDay) + 0.5
	//julianDay = math.Floor(julianDay) + 0.5 - originalTimeZone/24 + timeZone/24 // 求0时JDE
	//fix:这里时间分界线应当以传入的时区为准，不应当使用当地时区，否则在0时的判断会出错
	julianDay = math.Floor(julianDay) + 0.5
	estimatedTime := julianDay
	moonHeight := MoonHeight(julianDay, longitude, latitude, originalTimeZone) // 求此时月亮高度

	moonAngle := StandardAltitudeMoon(zenithShift, height, latitude)

	moonAngleTime := MoonTimeAngle(julianDay, longitude, latitude, originalTimeZone)

	if moonHeight-moonAngle > 0 { // 月亮在地平线上或在落下与下中天之间
		if moonAngleTime > 180 {
			timeToMeridian = (180 + 360 - moonAngleTime) / 15
		} else {
			timeToMeridian = (180 - moonAngleTime) / 15
		}
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24*12.0)/15.0/24.0)
	}

	if moonHeight-moonAngle < 0 && moonAngleTime > 180 {
		timeToMeridian = (180 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24*12.0)/15.0/24.0)
	} else if moonHeight-moonAngle < 0 && moonAngleTime < 180 {
		timeToMeridian = (180 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24*12.0)/15.0/24.0)
	}

	currentAngle := MoonTimeAngle(estimatedTime, longitude, latitude, timeZone)
	if math.Abs(currentAngle-180) > 0.5 {
		estimatedTime += (180 - currentAngle) * 4.0 / 60.0 / 24.0
	}

	currentHeight := HMoonHeight(estimatedTime, longitude, latitude, timeZone)
	if !(currentHeight < -10 && math.Abs(latitude) < 60) {
		if currentHeight > moonAngle {
			return 0, ErrNeverSet
		}
		checkTime := estimatedTime + 12.0/24.0 + 6.0/15.0/24.0
		checkAngle := MoonTimeAngle(checkTime, longitude, latitude, timeZone)
		if checkAngle < 90 {
			checkAngle += 360
		}
		checkTime += (360 - checkAngle) * 4.0 / 60.0 / 24.0
		if HMoonHeight(checkTime, longitude, latitude, timeZone) < moonAngle {
			return 0, ErrNeverRise
		}
	}

	moonDeclination := MoonApparentDec(estimatedTime, longitude, latitude, timeZone)
	tmp := (Sin(moonAngle) - Sin(moonDeclination)*Sin(latitude)) / (Cos(moonDeclination) * Cos(latitude))

	if math.Abs(tmp) <= 1 && latitude < 85 {
		hourAngle := (180 - ArcCos(tmp)) / 15
		estimatedTime += hourAngle/24.00 + hourAngle/33.00/15.00
	} else {
		i := 0
		for MoonHeight(estimatedTime, longitude, latitude, timeZone) < moonAngle {
			i++
			estimatedTime += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}

	// 使用牛顿迭代法求精确解
	estimatedTime = moonRiseSetNewtonRaphsonIteration(estimatedTime, longitude, latitude, timeZone, moonAngle, HMoonHeight, 0.00002)

	estimatedTime = estimatedTime - timeZone/24 + originalTimeZone/24

	if estimatedTime > julianDayZero+1 || estimatedTime < julianDayZero {
		return 0, ErrNotOnThisDate
	}
	return estimatedTime, nil
}

func GetMoonSetTime(julianDay, longitude, latitude, timeZone, zenithShift, height float64) (float64, error) {
	originalTimeZone := timeZone
	timeZone = longitude / 15
	var timeToMeridian float64
	julianDayZero := math.Floor(julianDay) + 0.5
	//julianDay = math.Floor(julianDay) + 0.5 - originalTimeZone/24 + timeZone/24 // 求0时JDE
	//fix:这里时间分界线应当以传入的时区为准，不应当使用当地时区，否则在0时的判断会出错
	julianDay = math.Floor(julianDay) + 0.5
	estimatedTime := julianDay
	moonHeight := MoonHeight(julianDay, longitude, latitude, originalTimeZone) // 求此时月亮高度

	moonAngle := StandardAltitudeMoon(zenithShift, height, latitude)

	moonAngleTime := MoonTimeAngle(julianDay, longitude, latitude, originalTimeZone)

	if moonHeight-moonAngle < 0 {
		timeToMeridian = (360 - moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24 + (timeToMeridian/24.0*12.0)/15.0/24.0)
	}

	// 月亮在地平线上或在落下与下中天之间
	if moonHeight-moonAngle > 0 && moonAngleTime < 180 {
		timeToMeridian = (-moonAngleTime) / 15
		estimatedTime += (timeToMeridian/24.0 + (timeToMeridian/24.0*12.0)/15.0/24.0)
	} else if moonHeight-moonAngle > 0 {
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
	currentHeight := HMoonHeight(estimatedTime, longitude, latitude, timeZone)
	if !(currentHeight > 10 && math.Abs(latitude) < 60) {
		if currentHeight < moonAngle {
			return 0, ErrNeverRise
		}
		checkTime := estimatedTime + 12.0/24.0 + 6.0/15.0/24.0
		angleSubtraction := 180 - MoonTimeAngle(checkTime, longitude, latitude, timeZone)
		checkTime += angleSubtraction * 4.0 / 60.0 / 24.0
		if HMoonHeight(checkTime, longitude, latitude, timeZone) > moonAngle {
			return 0, ErrNeverSet
		}
	}

	moonDeclination := MoonApparentDec(estimatedTime, longitude, latitude, timeZone)
	tmp := (Sin(moonAngle) - Sin(moonDeclination)*Sin(latitude)) / (Cos(moonDeclination) * Cos(latitude))

	if math.Abs(tmp) <= 1 && latitude < 85 {
		hourAngle := (ArcCos(tmp)) / 15.0
		estimatedTime += hourAngle/24 + hourAngle/33.0/15.0
	} else {
		i := 0
		for MoonHeight(estimatedTime, longitude, latitude, timeZone) > moonAngle {
			i++
			estimatedTime += 15.0 / 60.0 / 24.0
			if i > 48 {
				break
			}
		}
	}

	// 使用牛顿迭代法求精确解
	estimatedTime = moonRiseSetNewtonRaphsonIteration(estimatedTime, longitude, latitude, timeZone, moonAngle, HMoonHeight, 0.00002)

	estimatedTime = estimatedTime - timeZone/24 + originalTimeZone/24

	if estimatedTime > julianDayZero+1 || estimatedTime < julianDayZero {
		return 0, ErrNotOnThisDate
	}
	return estimatedTime, nil
}

// heightFunction 高度函数类型定义，用于牛顿迭代法
type heightFunction func(time, longitude, latitude, timeZone float64) float64

// moonRiseSetNewtonRaphsonIteration 牛顿-拉夫逊迭代法求解天体高度方程
func moonRiseSetNewtonRaphsonIteration(initialTime, longitude, latitude, timeZone, targetAngle float64,
	heightFunc heightFunction, tolerance float64) float64 {
	const derivativeStep = 0.000005

	currentTime := initialTime

	for {
		previousTime := currentTime

		// 计算函数值：f(t) = height(t) - targetAngle
		functionValue := heightFunc(previousTime, longitude, latitude, timeZone) - targetAngle

		// 计算导数：f'(t) ≈ (f(t+h) - f(t-h)) / (2h)
		derivative := (heightFunc(previousTime+derivativeStep, longitude, latitude, timeZone) -
			heightFunc(previousTime-derivativeStep, longitude, latitude, timeZone)) / (2 * derivativeStep)

		// 牛顿-拉夫逊公式：t_new = t_old - f(t) / f'(t)
		currentTime = previousTime - functionValue/derivative

		// 检查收敛
		if math.Abs(currentTime-previousTime) <= tolerance {
			break
		}
	}

	return currentTime
}

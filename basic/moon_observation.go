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
	jde := UTC2TT(jd - tz/24)
	ra := MoonTrueRa(jde)
	dec := MoonTrueDec(jde)
	away := MoonAway(jde) / 149597870.7
	ndec := TopocentricDec(ra, dec, lat, lon, jd-tz/24, away, 0)
	nra := TopocentricRa(ra, dec, lat, lon, jd-tz/24, away, 0)
	jdUT := jd - tz/24
	st := Limit360(ApparentSiderealTime(UTC2UT1(jdUT))*15 + lon)
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
	jde := UTC2TT(jd - tz/24)
	ra := MoonTrueRa(jde)
	dec := MoonTrueDec(jde)
	away := MoonAway(jde) / 149597870.7
	ndec := TopocentricDec(ra, dec, lat, lon, jd-tz/24, away, 0)
	nra := TopocentricRa(ra, dec, lat, lon, jd-tz/24, away, 0)
	jdUT := jd - tz/24
	st := Limit360(ApparentSiderealTime(UTC2UT1(jdUT))*15 + lon)
	hourAngle := Limit360(st - nra)
	tmp2 := Sin(lat)*Sin(ndec) + Cos(ndec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(tmp2)
}

func HMoonAzimuth(jd, lon, lat, tz float64) float64 {
	return HMoonAzimuthN(jd, lon, lat, tz, -1)
}

func HMoonAzimuthN(jd, lon, lat, tz float64, n int) float64 {
	jde := UTC2TT(jd - tz/24)
	ra := HMoonTrueRaN(jde, n)
	dec := HMoonTrueDecN(jde, n)
	away := HMoonAwayN(jde, n) / 149597870.7
	ndec := TopocentricDec(ra, dec, lat, lon, jd-tz/24, away, 0)
	nra := TopocentricRa(ra, dec, lat, lon, jd-tz/24, away, 0)
	jdUT := jd - tz/24
	st := Limit360(ApparentSiderealTime(UTC2UT1(jdUT))*15 + lon)
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

// HMoonHeight 当地民用时儒略日下的月心几何高度角（度，不含折射）/ geometric Moon-centre altitude in degrees for a local civil Julian day.
//
// jd 是该时区的当地民用时（墙上时刻）儒略日，tz 是时区偏移小时数，库内按 jd−tz/24 换成 UTC。
// 只有 tz 给 0 时 jd 才是 UTC 儒略日；不要拿 UTC 数值再配非零 tz，那会多减一次时区。
// jd is that zone's local civil (wall-clock) Julian day and tz is the zone offset in hours,
// converted internally as jd-tz/24. Only tz 0 makes jd a UTC Julian day: pairing a UTC value with a
// non-zero tz subtracts the offset twice.
func HMoonHeight(jd, lon, lat, tz float64) float64 {
	return HMoonHeightN(jd, lon, lat, tz, -1)
}

type moonObservationState struct {
	altitude   float64
	distanceKM float64
}

func hMoonObservationStateN(jd, lon, lat, tz, height float64, n int) moonObservationState {
	calculationJDE := UTC2TT(jd - tz/24)
	ra, dec := HMoonTrueRaDecN(calculationJDE, n)
	distanceKM := HMoonAwayN(calculationJDE, n)
	distanceAU := distanceKM / angularDiameterAstronomicalUnitKM
	topocentricRA, topocentricDec := TopocentricRaDec(ra, dec, lat, lon, jd-tz/24, distanceAU, height)
	siderealTime := Limit360(ApparentSiderealTime(UTC2UT1(jd-tz/24))*15 + lon)
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

// MoonState 同一瞬间可对任意观测点复用的月球位置与恒星时 / one instant's lunar position and sidereal time, reusable across observers.
type MoonState struct {
	rightAscension float64
	declination    float64
	distanceAU     float64
	siderealTime   float64
}

// MoonStateAt 由 UTC 儒略日构造该瞬间的可复用月球状态 / builds the reusable state for one UTC Julian day.
func MoonStateAt(utcJD float64) MoonState {
	jde := UTC2TT(utcJD)
	rightAscension, declination := HMoonTrueRaDec(jde)
	return MoonState{
		rightAscension: rightAscension,
		declination:    declination,
		distanceAU:     HMoonAway(jde) / angularDiameterAstronomicalUnitKM,
		siderealTime:   ApparentSiderealTime(UTC2UT1(utcJD)) * 15,
	}
}

func (state MoonState) finite() bool {
	return finite(state.rightAscension) && finite(state.declination) &&
		finite(state.distanceAU) && finite(state.siderealTime)
}

// HMoonHeight 给定观测者经度、纬度（度，椭球高 0）的月心几何高度角，等于 HMoonHeight(构造本状态时的 UTC 儒略日, 经, 纬, 0)。
// HMoonHeight returns the geometric Moon-centre altitude for one observer, equal to HMoonHeight(the UTC Julian day given to MoonStateAt, lon, lat, 0).
func (state MoonState) HMoonHeight(longitude, latitude float64) float64 {
	// 本状态固定是 UTC 瞬间、椭球高 0，因此只对应包级 tz=0、height=0 的用法。
	// 恒星时已在状态里算好，这里不再走会重算恒星时与时标换算的 TopocentricRaDec。
	topocentricRA, topocentricDec := topocentricRaDecWithSidereal(
		state.rightAscension, state.declination, latitude, longitude, state.siderealTime, state.distanceAU, 0,
	)
	hourAngle := Limit360(Limit360(state.siderealTime+longitude) - topocentricRA)
	return ArcSin(Sin(latitude)*Sin(topocentricDec) + Cos(topocentricDec)*Cos(latitude)*Cos(hourAngle))
}

// MoonHorizon 用本状态生成海平面几何月心地平圈，口径同包级 MoonHorizon / sea-level geometric Moon-centre horizon ring from this state.
func (state MoonState) MoonHorizon(samples int) [][2]float64 {
	if !state.finite() {
		return nil
	}
	if samples <= 0 {
		samples = 360
	}
	if samples < 12 {
		samples = 12
	} else if samples > 1440 {
		samples = 1440
	}
	parallax := math.Sin(0.0024427777777*rad) / state.distanceAU
	longitude := (state.rightAscension - state.siderealTime) * rad
	latitude := state.declination * rad
	if !finite(parallax) || parallax <= 0 || parallax >= 1 || !finite(longitude) || !finite(latitude) {
		return nil
	}
	center := [3]float64{math.Cos(latitude) * math.Cos(longitude), math.Cos(latitude) * math.Sin(longitude), math.Sin(latitude)}
	north := [3]float64{-math.Sin(latitude) * math.Cos(longitude), -math.Sin(latitude) * math.Sin(longitude), math.Cos(latitude)}
	east := [3]float64{-math.Sin(longitude), math.Cos(longitude), 0}
	points := make([][2]float64, samples)
	for index := range points {
		bearing := 2 * math.Pi * float64(index) / float64(samples)
		radius := math.Acos(parallax)
		var point [3]float64
		for iteration := 0; iteration < 8; iteration++ {
			for axis := range point {
				point[axis] = center[axis]*math.Cos(radius) +
					(north[axis]*math.Cos(bearing)+east[axis]*math.Sin(bearing))*math.Sin(radius)
			}
			lat := math.Asin(math.Max(-1, math.Min(1, point[2]))) / rad
			// The topocentric direction is horizontal when its dot product
			// with the geodetic zenith vanishes: cos(radius)=observer/range.
			next := math.Acos(parallax * (pcosi(lat, 0)*math.Cos(lat*rad) + psini(lat, 0)*math.Sin(lat*rad)))
			if math.Abs(next-radius) < 1e-14 {
				break
			}
			radius = next
		}
		points[index] = [2]float64{math.Atan2(point[1], point[0]) / rad, math.Asin(math.Max(-1, math.Min(1, point[2]))) / rad}
	}
	return points
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

func MoonCulminationTime(localJD, lon, lat, timezone float64) float64 {
	// localJD 是本地民用日锚点（当地 0 时），不是力学时；ra/dec 为瞬时天球坐标，非 J2000 等固定历元。
	localJD = math.Floor(localJD) + 0.5
	estimateJD := localJD + Limit360(360-MoonTimeAngle(localJD, lon, lat, timezone))/15.0/24.0/0.9
	limitHA := func(localJD, lon, timezone float64) float64 {
		ha := MoonTimeAngle(localJD, lon, lat, timezone)
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
	startime := Limit360(ApparentSiderealTime(UTC2UT1(jd-tz/24))*15 + lon)
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
	// 时间分界线以传入的时区为准，不用当地时区，否则 0 时的判断会出错。
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
	// 时间分界线以传入的时区为准，不用当地时区，否则 0 时的判断会出错。
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

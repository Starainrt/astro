package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

func sunTimeWithInstance(jde float64, n int, inst *Instance) float64 {
	if inst == nil {
		if n < 0 {
			return SunTime(jde)
		}
		return SunTimeN(jde, n)
	}
	// 均时差必须用平黄经 SunLo，真黄经会丢掉中心差。
	rightAscension := HSunApparentRaNWithInstance(jde, n, inst)
	time := (SunLo(jde) - 0.0057183 - rightAscension + Nutation2000Bi(jde)*Cos(TrueObliquity(jde))) / 15
	if time > 23 {
		time -= 24
	}
	return time
}

func sunHeightWithInstance(jd, lon, lat, tz float64, n int, inst *Instance) float64 {
	if inst == nil {
		if n < 0 {
			return SunHeight(jd, lon, lat, tz)
		}
		return SunHeightN(jd, lon, lat, tz, n)
	}
	return sunObservationStateN(jd, lon, lat, tz, n, inst).altitude
}

// SunTimeWithInstance 太阳均时差 / equation of time with an explicit ephemeris instance.
func SunTimeWithInstance(jde float64, inst *Instance) float64 {
	return sunTimeWithInstance(jde, -1, inst)
}

// SunTimeNWithInstance 太阳均时差（n 变体） / n-parameterised equation of time with an explicit ephemeris instance.
func SunTimeNWithInstance(jde float64, n int, inst *Instance) float64 {
	return sunTimeWithInstance(jde, n, inst)
}

// CulminationTimeWithInstance 太阳中天时刻 / solar culmination time with an explicit ephemeris instance.
func CulminationTimeWithInstance(localJD, lon, tz float64, inst *Instance) float64 {
	if inst == nil {
		return CulminationTime(localJD, lon, tz)
	}
	localJD = math.Floor(localJD)
	tmp := longitudeOffsetDegrees(tz*15-lon) * 4 / 60
	return localJD + tmp/24 - SunTimeWithInstance(localJD, inst)/24
}

// CulminationTimeNWithInstance 太阳中天时刻（n 变体） / n-parameterised solar culmination time with an explicit ephemeris instance.
func CulminationTimeNWithInstance(localJD, lon, tz float64, n int, inst *Instance) float64 {
	if inst == nil {
		return CulminationTimeN(localJD, lon, tz, n)
	}
	localJD = math.Floor(localJD)
	tmp := longitudeOffsetDegrees(tz*15-lon) * 4 / 60
	return localJD + tmp/24 - SunTimeNWithInstance(localJD, n, inst)/24
}

// SunTimeAngleWithInstance 太阳时角 / solar hour angle with an explicit ephemeris instance.
func SunTimeAngleWithInstance(jd, lon, lat, tz float64, inst *Instance) float64 {
	if inst == nil {
		return SunTimeAngle(jd, lon, lat, tz)
	}
	calculationJD := jd - tz/24
	siderealTime := Limit360(ApparentSiderealTime(UTC2UT1(calculationJD))*15 + lon)
	hourAngle := siderealTime - HSunApparentRaWithInstance(UTC2TT(calculationJD), inst)
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

// SunTimeAngleNWithInstance 太阳时角（n 变体） / n-parameterised solar hour angle with an explicit ephemeris instance.
func SunTimeAngleNWithInstance(jd, lon, lat, tz float64, n int, inst *Instance) float64 {
	if inst == nil {
		return SunTimeAngleN(jd, lon, lat, tz, n)
	}
	calculationJD := jd - tz/24
	siderealTime := Limit360(ApparentSiderealTime(UTC2UT1(calculationJD))*15 + lon)
	hourAngle := siderealTime - HSunApparentRaNWithInstance(UTC2TT(calculationJD), n, inst)
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

// SunHeightWithInstance 太阳高度角 / solar altitude with an explicit ephemeris instance.
func SunHeightWithInstance(jd, lon, lat, tz float64, inst *Instance) float64 {
	return sunHeightWithInstance(jd, lon, lat, tz, -1, inst)
}

// SunHeightNWithInstance 太阳高度角（n 变体） / n-parameterised solar altitude with an explicit ephemeris instance.
func SunHeightNWithInstance(jd, lon, lat, tz float64, n int, inst *Instance) float64 {
	return sunHeightWithInstance(jd, lon, lat, tz, n, inst)
}

// LowSunHeightWithInstance 低精度太阳高度角 / low-precision solar altitude with an explicit ephemeris instance.
func LowSunHeightWithInstance(jd, lon, lat, tz float64, inst *Instance) float64 {
	if inst == nil {
		return LowSunHeight(jd, lon, lat, tz)
	}
	calculationJD := jd - tz/24
	siderealTime := Limit360(ApparentSiderealTime(UTC2UT1(calculationJD))*15 + lon)
	ra, dec := HSunApparentRaDecWithInstance(UTC2TT(calculationJD), inst)
	hourAngle := Limit360(siderealTime - ra)
	return ArcSin(Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle))
}

// SunAzimuthWithInstance 太阳方位角 / solar azimuth with an explicit ephemeris instance.
func SunAzimuthWithInstance(jd, lon, lat, tz float64, inst *Instance) float64 {
	return sunAzimuthWithInstance(jd, lon, lat, tz, -1, inst)
}

// SunAzimuthNWithInstance 太阳方位角（n 变体） / n-parameterised solar azimuth with an explicit ephemeris instance.
func SunAzimuthNWithInstance(jd, lon, lat, tz float64, n int, inst *Instance) float64 {
	return sunAzimuthWithInstance(jd, lon, lat, tz, n, inst)
}

func sunAzimuthWithInstance(jd, lon, lat, tz float64, n int, inst *Instance) float64 {
	if inst == nil {
		if n < 0 {
			return SunAzimuth(jd, lon, lat, tz)
		}
		return SunAzimuthN(jd, lon, lat, tz, n)
	}
	return sunAzimuthN(jd, lon, lat, tz, n, inst)
}

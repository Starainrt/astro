package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func NeptuneL(jde float64) float64 {
	return planet.WherePlanet(7, 0, jde)
}

func NeptuneB(jde float64) float64 {
	return planet.WherePlanet(7, 1, jde)
}
func NeptuneR(jde float64) float64 {
	return planet.WherePlanet(7, 2, jde)
}
func ANeptuneX(jde float64) float64 {
	l := NeptuneL(jde)
	b := NeptuneB(jde)
	r := NeptuneR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func ANeptuneY(jde float64) float64 {

	l := NeptuneL(jde)
	b := NeptuneB(jde)
	r := NeptuneR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func ANeptuneZ(jde float64) float64 {
	//l := NeptuneL(jde)
	b := NeptuneB(jde)
	r := NeptuneR(jde)
	//	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func ANeptuneXYZ(jde float64) (float64, float64, float64) {
	l := NeptuneL(jde)
	b := NeptuneB(jde)
	r := NeptuneR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func NeptuneApparentRa(jde float64) float64 {
	lo, bo := NeptuneApparentLoBo(jde)
	eps := TrueObliquity(jde)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func NeptuneApparentDec(jde float64) float64 {
	lo, bo := NeptuneApparentLoBo(jde)
	eps := TrueObliquity(jde)
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return dec
}

func NeptuneApparentRaDec(jde float64) (float64, float64) {
	lo, bo := NeptuneApparentLoBo(jde)
	eps := TrueObliquity(jde)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return Limit360(ra), dec
}

func EarthNeptuneAway(jd float64) float64 {
	return planetEarthAwayExplicitN(7, jd, -1)
}

func NeptuneApparentLo(jd float64) float64 {
	geo, _ := planetApparentGeocentricPositionN(7, jd, -1)
	return geo.lo
}

func NeptuneApparentBo(jd float64) float64 {
	geo, _ := planetApparentGeocentricPositionN(7, jd, -1)
	return geo.bo
}

func NeptuneApparentLoBo(jd float64) (float64, float64) {
	geo, _ := planetApparentGeocentricPositionN(7, jd, -1)
	return geo.lo, geo.bo
}

func NeptuneMag(jde float64) float64 {
	sunDistance := NeptuneR(jde)
	earthDistance := EarthNeptuneAway(jde)
	earthSunDistance := planet.WherePlanet(-1, 2, jde)
	i := (sunDistance*sunDistance + earthDistance*earthDistance - earthSunDistance*earthSunDistance) / (2 * sunDistance * earthDistance)
	i = ArcCos(i)
	mag := -6.87 + 5*math.Log10(sunDistance*earthDistance)
	return FloatRound(mag, 2)
}

func NeptuneHeight(localJD, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJD := localJD - timezone/24.0
	// 计算视恒星时
	ra, dec := NeptuneApparentRaDec(UTC2TT(utcJD))
	st := Limit360(ApparentSiderealTime(UTC2UT1(utcJD))*15 + lon)
	// 计算时角
	hourAngle := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(hourAngle)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(sinHeight)
}

func NeptuneAzimuth(localJD, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJD := localJD - timezone/24.0
	// 计算视恒星时
	ra, dec := NeptuneApparentRaDec(UTC2TT(utcJD))
	st := Limit360(ApparentSiderealTime(UTC2UT1(utcJD))*15 + lon)
	// 计算时角
	hourAngle := Limit360(st - ra)
	// 三角转换公式
	tanAzimuth := Sin(hourAngle) / (Cos(hourAngle)*Sin(lat) - Tan(dec)*Cos(lat))
	azimuth := ArcTan(tanAzimuth)
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

func NeptuneHourAngle(jd, lon, timezone float64) float64 {
	siderealLongitude := Limit360(ApparentSiderealTime(UTC2UT1(jd-timezone/24))*15 + lon)
	hourAngle := siderealLongitude - NeptuneApparentRa(UTC2TT(jd-timezone/24.0))
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

func NeptuneCulminationTime(localJD, lon, timezone float64) float64 {
	// localJD 是本地民用日锚点（当地 0 时），不是力学时。
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	localJD = math.Floor(localJD) + 0.5
	estimateJD := localJD + Limit360(360-NeptuneHourAngle(localJD, lon, timezone))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(localJD, lon, timezone float64) float64 {
		currentHourAngle := NeptuneHourAngle(localJD, lon, timezone)
		if currentHourAngle < 180 {
			currentHourAngle += 360
		}
		return currentHourAngle
	}
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		hourAngleDelta := normalizedHourAngle(prevJD, lon, timezone) - 360
		hourAngleSlope := (normalizedHourAngle(prevJD+0.000005, lon, timezone) - normalizedHourAngle(prevJD-0.000005, lon, timezone)) / 0.00001
		return hourAngleDelta / hourAngleSlope
	})
	if !ok {
		return math.NaN()
	}
	return estimateJD
}

func NeptuneRiseTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return neptuneRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, true)
}

func NeptuneSetTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return neptuneRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, false)
}

func neptuneRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, isRise bool) (float64, error) {
	return planetRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, isRise, NeptuneCulminationTime, NeptuneHeight, NeptuneApparentDec)
}

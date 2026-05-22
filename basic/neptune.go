package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func NeptuneL(jd float64) float64 {
	return planet.WherePlanet(7, 0, jd)
}

func NeptuneB(jd float64) float64 {
	return planet.WherePlanet(7, 1, jd)
}
func NeptuneR(jd float64) float64 {
	return planet.WherePlanet(7, 2, jd)
}
func ANeptuneX(jd float64) float64 {
	l := NeptuneL(jd)
	b := NeptuneB(jd)
	r := NeptuneR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func ANeptuneY(jd float64) float64 {

	l := NeptuneL(jd)
	b := NeptuneB(jd)
	r := NeptuneR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func ANeptuneZ(jd float64) float64 {
	//l := NeptuneL(jd)
	b := NeptuneB(jd)
	r := NeptuneR(jd)
	//	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func ANeptuneXYZ(jd float64) (float64, float64, float64) {
	l := NeptuneL(jd)
	b := NeptuneB(jd)
	r := NeptuneR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func NeptuneApparentRa(jd float64) float64 {
	lo, bo := NeptuneApparentLoBo(jd)
	eps := TrueObliquity(jd)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func NeptuneApparentDec(jd float64) float64 {
	lo, bo := NeptuneApparentLoBo(jd)
	eps := TrueObliquity(jd)
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return dec
}

func NeptuneApparentRaDec(jd float64) (float64, float64) {
	lo, bo := NeptuneApparentLoBo(jd)
	eps := TrueObliquity(jd)
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

func NeptuneMag(jd float64) float64 {
	sunDistance := NeptuneR(jd)
	earthDistance := EarthNeptuneAway(jd)
	earthSunDistance := planet.WherePlanet(-1, 2, jd)
	i := (sunDistance*sunDistance + earthDistance*earthDistance - earthSunDistance*earthSunDistance) / (2 * sunDistance * earthDistance)
	i = ArcCos(i)
	mag := -6.87 + 5*math.Log10(sunDistance*earthDistance)
	return FloatRound(mag, 2)
}

func NeptuneHeight(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := NeptuneApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	hourAngle := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(hourAngle)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(sinHeight)
}

func NeptuneAzimuth(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := NeptuneApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
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
	siderealLongitude := Limit360(ApparentSiderealTime(jd-timezone/24)*15 + lon)
	hourAngle := siderealLongitude - NeptuneApparentRa(TD2UT(jd-timezone/24.0, true))
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

func NeptuneCulminationTime(jde, lon, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	estimateJD := jde + Limit360(360-NeptuneHourAngle(jde, lon, timezone))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(jde, lon, timezone float64) float64 {
		currentHourAngle := NeptuneHourAngle(jde, lon, timezone)
		if currentHourAngle < 180 {
			currentHourAngle += 360
		}
		return currentHourAngle
	}
	for {
		prevJD := estimateJD
		hourAngleDelta := normalizedHourAngle(prevJD, lon, timezone) - 360
		hourAngleSlope := (normalizedHourAngle(prevJD+0.000005, lon, timezone) - normalizedHourAngle(prevJD-0.000005, lon, timezone)) / 0.00001
		estimateJD = prevJD - hourAngleDelta/hourAngleSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
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

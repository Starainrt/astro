package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func VenusL(jde float64) float64 {
	return planet.WherePlanet(2, 0, jde)
}

func VenusB(jde float64) float64 {
	return planet.WherePlanet(2, 1, jde)
}
func VenusR(jde float64) float64 {
	return planet.WherePlanet(2, 2, jde)
}
func AVenusX(jde float64) float64 {
	l := VenusL(jde)
	b := VenusB(jde)
	r := VenusR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func AVenusY(jde float64) float64 {

	l := VenusL(jde)
	b := VenusB(jde)
	r := VenusR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func AVenusZ(jde float64) float64 {
	//l := VenusL(jde)
	b := VenusB(jde)
	r := VenusR(jde)
	//	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func AVenusXYZ(jde float64) (float64, float64, float64) {
	l := VenusL(jde)
	b := VenusB(jde)
	r := VenusR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func VenusApparentRa(jde float64) float64 {
	lo, bo := VenusApparentLoBo(jde)
	eps := TrueObliquity(jde)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func VenusApparentDec(jde float64) float64 {
	lo, bo := VenusApparentLoBo(jde)
	eps := TrueObliquity(jde)
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return dec
}

func VenusApparentRaDec(jde float64) (float64, float64) {
	lo, bo := VenusApparentLoBo(jde)
	eps := TrueObliquity(jde)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return Limit360(ra), dec
}

func EarthVenusAway(jd float64) float64 {
	return planetEarthAwayExplicitN(2, jd, -1)
}

func VenusApparentLo(jd float64) float64 {
	geo, _ := planetApparentGeocentricPositionN(2, jd, -1)
	return geo.lo
}

func VenusApparentBo(jd float64) float64 {
	geo, _ := planetApparentGeocentricPositionN(2, jd, -1)
	return geo.bo
}

func VenusApparentLoBo(jd float64) (float64, float64) {
	geo, _ := planetApparentGeocentricPositionN(2, jd, -1)
	return geo.lo, geo.bo
}

func VenusMag(jde float64) float64 {
	sunDistance := VenusR(jde)
	earthDistance := EarthVenusAway(jde)
	earthSunDistance := planet.WherePlanet(-1, 2, jde)
	i := (sunDistance*sunDistance + earthDistance*earthDistance - earthSunDistance*earthSunDistance) / (2 * sunDistance * earthDistance)
	i = ArcCos(i)
	mag := -4.40 + 5*math.Log10(sunDistance*earthDistance) + 0.0009*i + 0.000239*i*i - 0.00000065*i*i*i
	return FloatRound(mag, 2)
}

func VenusHeight(localJD, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJD := localJD - timezone/24.0
	// 计算视恒星时
	ra, dec := VenusApparentRaDec(UTC2TT(utcJD))
	st := Limit360(ApparentSiderealTime(UTC2UT1(utcJD))*15 + lon)
	// 计算时角
	hourAngle := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(hourAngle)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(sinHeight)
}

func VenusAzimuth(localJD, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJD := localJD - timezone/24.0
	// 计算视恒星时
	ra, dec := VenusApparentRaDec(UTC2TT(utcJD))
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

func VenusHourAngle(jd, lon, tz float64) float64 {
	startime := Limit360(ApparentSiderealTime(UTC2UT1(jd-tz/24))*15 + lon)
	timeangle := startime - VenusApparentRa(UTC2TT(jd-tz/24.0))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

func VenusCulminationTime(localJD, lon, timezone float64) float64 {
	// localJD 是本地民用日锚点（当地 0 时），不是力学时。
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	localJD = math.Floor(localJD) + 0.5
	estimateJD := localJD + Limit360(360-VenusHourAngle(localJD, lon, timezone))/15.0/24.0*0.99726851851851851851
	limitHA := func(localJD, lon, timezone float64) float64 {
		ha := VenusHourAngle(localJD, lon, timezone)
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

func VenusRiseTime(jd, lon, lat, tz, aeroCorrection, observerHeight float64) (float64, error) {
	return venusRiseDown(jd, lon, lat, tz, aeroCorrection, observerHeight, true)
}

func VenusSetTime(jd, lon, lat, tz, aeroCorrection, observerHeight float64) (float64, error) {
	return venusRiseDown(jd, lon, lat, tz, aeroCorrection, observerHeight, false)
}

func venusRiseDown(jd, lon, lat, tz, aeroCorrection, observerHeight float64, isRise bool) (float64, error) {
	return planetRiseDown(jd, lon, lat, tz, aeroCorrection, observerHeight, isRise, VenusCulminationTime, VenusHeight, VenusApparentDec)
}

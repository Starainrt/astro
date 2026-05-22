package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func VenusL(jd float64) float64 {
	return planet.WherePlanet(2, 0, jd)
}

func VenusB(jd float64) float64 {
	return planet.WherePlanet(2, 1, jd)
}
func VenusR(jd float64) float64 {
	return planet.WherePlanet(2, 2, jd)
}
func AVenusX(jd float64) float64 {
	l := VenusL(jd)
	b := VenusB(jd)
	r := VenusR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func AVenusY(jd float64) float64 {

	l := VenusL(jd)
	b := VenusB(jd)
	r := VenusR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func AVenusZ(jd float64) float64 {
	//l := VenusL(jd)
	b := VenusB(jd)
	r := VenusR(jd)
	//	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func AVenusXYZ(jd float64) (float64, float64, float64) {
	l := VenusL(jd)
	b := VenusB(jd)
	r := VenusR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func VenusApparentRa(jd float64) float64 {
	lo, bo := VenusApparentLoBo(jd)
	eps := TrueObliquity(jd)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func VenusApparentDec(jd float64) float64 {
	lo, bo := VenusApparentLoBo(jd)
	eps := TrueObliquity(jd)
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return dec
}

func VenusApparentRaDec(jd float64) (float64, float64) {
	lo, bo := VenusApparentLoBo(jd)
	eps := TrueObliquity(jd)
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

func VenusMag(jd float64) float64 {
	sunDistance := VenusR(jd)
	earthDistance := EarthVenusAway(jd)
	earthSunDistance := planet.WherePlanet(-1, 2, jd)
	i := (sunDistance*sunDistance + earthDistance*earthDistance - earthSunDistance*earthSunDistance) / (2 * sunDistance * earthDistance)
	i = ArcCos(i)
	mag := -4.40 + 5*math.Log10(sunDistance*earthDistance) + 0.0009*i + 0.000239*i*i - 0.00000065*i*i*i
	return FloatRound(mag, 2)
}

func VenusHeight(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := VenusApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	hourAngle := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(hourAngle)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(sinHeight)
}

func VenusAzimuth(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := VenusApparentRaDec(TD2UT(utcJde, true))
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

func VenusHourAngle(jd, lon, tz float64) float64 {
	startime := Limit360(ApparentSiderealTime(jd-tz/24)*15 + lon)
	timeangle := startime - VenusApparentRa(TD2UT(jd-tz/24.0, true))
	if timeangle < 0 {
		timeangle += 360
	}
	return timeangle
}

func VenusCulminationTime(jde, lon, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	estimateJD := jde + Limit360(360-VenusHourAngle(jde, lon, timezone))/15.0/24.0*0.99726851851851851851
	limitHA := func(jde, lon, timezone float64) float64 {
		ha := VenusHourAngle(jde, lon, timezone)
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

func VenusRiseTime(jd, lon, lat, tz, aeroCorrection, observerHeight float64) (float64, error) {
	return venusRiseDown(jd, lon, lat, tz, aeroCorrection, observerHeight, true)
}

func VenusSetTime(jd, lon, lat, tz, aeroCorrection, observerHeight float64) (float64, error) {
	return venusRiseDown(jd, lon, lat, tz, aeroCorrection, observerHeight, false)
}

func venusRiseDown(jd, lon, lat, tz, aeroCorrection, observerHeight float64, isRise bool) (float64, error) {
	return planetRiseDown(jd, lon, lat, tz, aeroCorrection, observerHeight, isRise, VenusCulminationTime, VenusHeight, VenusApparentDec)
}

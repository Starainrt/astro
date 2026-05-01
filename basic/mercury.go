package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func MercuryL(jd float64) float64 {
	return planet.WherePlanet(1, 0, jd)
}

func MercuryB(jd float64) float64 {
	return planet.WherePlanet(1, 1, jd)
}
func MercuryR(jd float64) float64 {
	return planet.WherePlanet(1, 2, jd)
}
func AMercuryX(jd float64) float64 {
	l := MercuryL(jd)
	b := MercuryB(jd)
	r := MercuryR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func AMercuryY(jd float64) float64 {

	l := MercuryL(jd)
	b := MercuryB(jd)
	r := MercuryR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func AMercuryZ(jd float64) float64 {
	//l := MercuryL(jd)
	b := MercuryB(jd)
	r := MercuryR(jd)
	//	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func AMercuryXYZ(jd float64) (float64, float64, float64) {
	l := MercuryL(jd)
	b := MercuryB(jd)
	r := MercuryR(jd)
	el := planet.WherePlanet(-1, 0, jd)
	eb := planet.WherePlanet(-1, 1, jd)
	er := planet.WherePlanet(-1, 2, jd)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func MercuryApparentRa(jd float64) float64 {
	lo, bo := MercuryApparentLoBo(jd)
	return LoToRa(jd, lo, bo)
}
func MercuryApparentDec(jd float64) float64 {
	lo, bo := MercuryApparentLoBo(jd)
	eps := TrueObliquity(jd)
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return dec
}

func MercuryApparentRaDec(jd float64) (float64, float64) {
	lo, bo := MercuryApparentLoBo(jd)
	return LoBoToRaDec(jd, lo, bo)
}

func EarthMercuryAway(jd float64) float64 {
	x, y, z := AMercuryXYZ(jd)
	to := math.Sqrt(x*x + y*y + z*z)
	return to
}

func MercuryApparentLo(jd float64) float64 {
	x, y, z := AMercuryXYZ(jd)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMercuryXYZ(jd - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo)
	//lo-=GXCLo(lo,bo,jd)/3600;
	//bo+=GXCBo(lo,bo,jd);
	lo += Nutation2000Bi(jd)
	return lo
}

func MercuryApparentBo(jd float64) float64 {
	x, y, z := AMercuryXYZ(jd)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMercuryXYZ(jd - to)
	//lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	//lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	//lo+=GXCLo(lo,bo,jd);
	//bo+=GXCBo(lo,bo,jd)/3600;
	//lo+=Nutation2000Bi(jd);
	return bo
}

func MercuryApparentLoBo(jd float64) (float64, float64) {
	x, y, z := AMercuryXYZ(jd)
	to := 0.0057755183 * math.Sqrt(x*x+y*y+z*z)
	x, y, z = AMercuryXYZ(jd - to)
	lo := math.Atan2(y, x)
	bo := math.Atan2(z, math.Sqrt(x*x+y*y))
	lo = lo * 180 / math.Pi
	bo = bo * 180 / math.Pi
	lo = Limit360(lo) + Nutation2000Bi(jd)
	//lo-=GXCLo(lo,bo,jd)/3600;
	//bo+=GXCBo(lo,bo,jd);
	return lo, bo
}

func MercuryMag(jd float64) float64 {
	sunDistance := MercuryR(jd)
	earthDistance := EarthMercuryAway(jd)
	earthSunDistance := planet.WherePlanet(-1, 2, jd)
	i := (sunDistance*sunDistance + earthDistance*earthDistance - earthSunDistance*earthSunDistance) / (2 * sunDistance * earthDistance)
	i = ArcCos(i)
	mag := -0.42 + 5*math.Log10(sunDistance*earthDistance) + 0.0380*i - 0.000273*i*i + 0.000002*i*i*i
	return FloatRound(mag, 2)
}

func MercuryHeight(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := MercuryApparentRaDec(TD2UT(utcJde, true))
	st := Limit360(ApparentSiderealTime(utcJde)*15 + lon)
	// 计算时角
	hourAngle := Limit360(st - ra)
	// 高度角、时角与天球座标三角转换公式
	// sin(h)=sin(lat)*sin(dec)+cos(dec)*cos(lat)*cos(hourAngle)
	sinHeight := Sin(lat)*Sin(dec) + Cos(dec)*Cos(lat)*Cos(hourAngle)
	return ArcSin(sinHeight)
}

func MercuryAzimuth(jde, lon, lat, timezone float64) float64 {
	// 转换为世界时
	utcJde := jde - timezone/24.0
	// 计算视恒星时
	ra, dec := MercuryApparentRaDec(TD2UT(utcJde, true))
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

func MercuryHourAngle(jd, lon, timezone float64) float64 {
	siderealLongitude := Limit360(ApparentSiderealTime(jd-timezone/24)*15 + lon)
	hourAngle := siderealLongitude - MercuryApparentRa(TD2UT(jd-timezone/24.0, true))
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

func MercuryCulminationTime(jde, lon, timezone float64) float64 {
	//jde 世界时，非力学时，当地时区 0时，无需转换力学时
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	jde = math.Floor(jde) + 0.5
	estimateJD := jde + Limit360(360-MercuryHourAngle(jde, lon, timezone))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(jde, lon, timezone float64) float64 {
		currentHourAngle := MercuryHourAngle(jde, lon, timezone)
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

func MercuryRiseTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return mercuryRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, true)
}

func MercurySetTime(jd, lon, lat, timezone, aeroCorrection, observerHeight float64) (float64, error) {
	return mercuryRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, false)
}

func mercuryRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, isRise bool) (float64, error) {
	return planetRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, isRise, MercuryCulminationTime, MercuryHeight, MercuryApparentDec)
}
